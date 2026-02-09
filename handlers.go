package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var loginAttempts = make(map[string]*LoginAttempt)
var loginLock = &sync.Mutex{}

const authToken = "your_default_token_here"

// Logout handler
func LogoutHandler(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	tokenString := c.MustGet("token_string").(string)

	// Parsujemy token, żeby dostać czas wygaśnięcia
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	exp := int64(claims["exp"].(float64))

	revoked := RevokedToken{
		Token:     tokenString,
		ExpiresAt: primitive.NewDateTimeFromTime(time.Unix(exp, 0)),
	}

	_, err = db.Collection("revoked_tokens").InsertOne(context.TODO(), revoked)
	if err != nil {
		log.Printf("❌ Błąd przy zapisie do black listy: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	log.Println("🔒 Token dodany do black listy (logout)")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Token autoryzacyjny dla połączenia z socketem, zdefiniowany globalnie
func generateCallbackToken() (string, error) {
	b := make([]byte, 16) // 128 bit = 32 hex chars
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GetWorldStatsHandler(c *gin.Context) {
	world := strings.TrimSpace(c.Param("world"))
	if world == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "world parameter is required"})
		return
	}

	db, err := ConnectMariaDB("playerstats")
	if err != nil {
		log.Println("❌ Failed to connect to MariaDB:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT player_uuid, player_name, world, statistic, value FROM questmc_playerstats_world_stats WHERE world = ?", world)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, []WorldStat{})
			return
		}
		log.Println("❌ Failed to query world stats:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query world stats"})
		return
	}
	defer rows.Close()

	stats := make([]WorldStat, 0)
	for rows.Next() {
		var stat WorldStat
		if err := rows.Scan(&stat.PlayerUUID, &stat.PlayerName, &stat.World, &stat.Statistic, &stat.Value); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusOK, []WorldStat{})
				return
			}
			log.Println("❌ Failed to scan world stat row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read world stats"})
			return
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, []WorldStat{})
			return
		}
		log.Println("❌ Rows iteration error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read world stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func GetWorldsHandler(c *gin.Context) {
	db, err := ConnectMariaDB("playerstats")
	if err != nil {
		log.Println("❌ Failed to connect to MariaDB:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT DISTINCT world FROM questmc_playerstats_world_stats ORDER BY world")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, []string{})
			return
		}
		log.Println("❌ Failed to query worlds:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query worlds"})
		return
	}
	defer rows.Close()

	worlds := make([]string, 0)
	for rows.Next() {
		var world string
		if err := rows.Scan(&world); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusOK, []string{})
				return
			}
			log.Println("❌ Failed to scan world row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read worlds"})
			return
		}
		worlds = append(worlds, world)
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, []string{})
			return
		}
		log.Println("❌ Rows iteration error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read worlds"})
		return
	}

	c.JSON(http.StatusOK, worlds)
}

func CreateTransactionHandler(c *gin.Context) {
	type ProductInRequest struct {
		ID       string   `json:"id"`
		Name     string   `json:"name"`
		Price    float64  `json:"price"`
		Currency string   `json:"currency"`
		Servers  []string `json:"servers"`
	}

	type Req struct {
		PlayerName string             `json:"playerName"`
		ItemName   string             `json:"itemName"` // tylko do opisu PayU
		Price      float64            `json:"price"`
		Currency   string             `json:"currency"`
		Products   []ProductInRequest `json:"products"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("❌ Błąd parsowania JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 1️⃣ Token do callbacka
	callbackToken, err := generateCallbackToken()
	log.Println("✅ Wygenerowano callback token:", callbackToken)
	if err != nil {
		log.Println("❌ Błąd generowania callback tokena:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	log.Printf("✅ Otrzymano żądanie transakcji od %s z %d produktami", req.PlayerName, len(req.Products))
	for i, p := range req.Products {
		log.Printf("📦 [%d] %s | ID: %s | %.2f %s | Serwery: %v", i, p.Name, p.ID, p.Price, p.Currency, p.Servers)
	}

	// 2️⃣ Token PayU
	accessToken, err := getPayuAccessToken()
	if err != nil {
		log.Println("❌ Błąd pobierania tokena PayU:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get PayU token"})
		return
	}
	log.Println("✅ Pobrano token PayU")

	// 3️⃣ Tworzenie zamówienia
	productsForPayU := make([]map[string]interface{}, len(req.Products))
	for i, p := range req.Products {
		productsForPayU[i] = map[string]interface{}{
			"name":      p.Name,
			"unitPrice": fmt.Sprintf("%.0f", p.Price*100),
			"quantity":  1,
		}
	}

	order := map[string]interface{}{
		"notifyUrl":     fmt.Sprintf("https://boxpvp.top:8443/api/paymentCallback?token=%s", callbackToken),
		"continueUrl":   "https://bestservers.fun/payu-done.html",
		"customerIp":    c.ClientIP(),
		"merchantPosId": payuClientID,
		"description":   fmt.Sprintf("Zakup %d produktów przez %s", len(req.Products), req.PlayerName),
		"currencyCode":  req.Currency,
		"totalAmount":   fmt.Sprintf("%.0f", req.Price*100),
		"products":      productsForPayU,
		"buyer": map[string]interface{}{
			"email":    fmt.Sprintf("%s@boxpvp.fake", req.PlayerName),
			"language": "pl",
		},
	}

	orderBody, _ := json.Marshal(order)
	log.Println("✅ Utworzone zamówienie PayU:", string(orderBody))

	// 4️⃣ Wysłanie zamówienia do PayU
	reqPayu, _ := http.NewRequest("POST", payuOrderURL, bytes.NewBuffer(orderBody))
	reqPayu.Header.Set("Content-Type", "application/json")
	reqPayu.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(reqPayu)
	if err != nil {
		log.Println("❌ Błąd wysyłania zamówienia do PayU:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PayU request failed"})
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	body, _ := io.ReadAll(resp.Body)

	trimmedBody := string(body)
	if len(trimmedBody) > 300 {
		trimmedBody = trimmedBody[:300] + "... [obcięto]"
	}
	log.Println("📩 Odpowiedź PayU:", trimmedBody)

	var payuResp map[string]interface{}
	if err := json.Unmarshal(body, &payuResp); err != nil {
		log.Println("❌ Błąd parsowania JSON z odpowiedzi PayU:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse PayU response"})
		return
	}

	redirectURL, ok := payuResp["redirectUri"].(string)
	if !ok {
		log.Println("❌ Brak redirectUri w odpowiedzi PayU:", payuResp)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid PayU response"})
		return
	}

	if strings.Contains(contentType, "text/html") {
		log.Println("⚠️ PayU zwróciło HTML zamiast JSON-a")
		c.Redirect(http.StatusFound, string(body))
		return
	}

	statusMap, ok := payuResp["status"].(map[string]interface{})
	if !ok || statusMap["statusCode"] != "SUCCESS" {
		log.Println("❌ PayU zwróciło status != SUCCESS:", payuResp)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PayU status != SUCCESS", "details": payuResp})
		return
	}

	transactionID, ok := payuResp["orderId"].(string)
	if !ok {
		log.Println("❌ Brak orderId w odpowiedzi PayU")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid PayU response"})
		return
	}

	// 5️⃣ Zapis jednej transakcji z wieloma produktami
	productsToSave := make([]TxProduct, len(req.Products))
	for i, p := range req.Products {
		productsToSave[i] = TxProduct{
			ID:       p.ID,
			Name:     p.Name,
			Price:    p.Price,
			Currency: p.Currency,
			Servers:  p.Servers,
		}
	}

	transaction := Transaction{
		PlayerName:    req.PlayerName,
		Products:      productsToSave,
		Price:         req.Price,
		Currency:      req.Currency,
		Status:        "PENDING",
		TransactionID: transactionID,
		CallbackToken: callbackToken,
		CreatedAt:     time.Now(),
	}

	_, err = transactionsCollection.InsertOne(context.TODO(), transaction)
	if err != nil {
		log.Println("❌ Błąd zapisu transakcji w MongoDB:", err)
	} else {
		log.Println("✅ Zapisano transakcję z wieloma produktami:", transactionID)
	}

	c.JSON(http.StatusOK, gin.H{"redirectUrl": redirectURL})
}

func processCommandHandler(c *gin.Context) {

	command := c.PostForm("command")
	if command == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No command provided"})
		return
	}
	fullCommand := fmt.Sprintf("%s:%s", authToken, command)

	conn, err := net.Dial("unix", "/tmp/plugin_minecraft.sock")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to the Minecraft server"})
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(fullCommand + "\n"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send command to the Minecraft server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Command sent successfully", "command": fullCommand})
}

func PaymentCallbackHandler(c *gin.Context) {
	log.Println("🔔 Otrzymano callback z PayU")
	log.Println("🔗 Query:", c.Request.URL.Query())
	token := c.Query("token")
	if token == "" {
		log.Println("❌ Brak tokena w zapytaniu do paymentCallback")
		c.JSON(http.StatusForbidden, gin.H{"error": "Missing token"})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nie można odczytać danych"})
		return
	}

	log.Println("📩 Otrzymano callback z PayU:", string(body))

	var notification struct {
		Order struct {
			OrderId string `json:"orderId"`
			Status  string `json:"status"`
			Buyer   struct {
				Email string `json:"email"`
			} `json:"buyer"`
		} `json:"order"`
	}

	if err := json.Unmarshal(body, &notification); err != nil {
		log.Println("❌ Błąd dekodowania JSON:", err)
		c.Status(http.StatusBadRequest)
		return
	}

	// 🟢 Pobieramy transakcję na podstawie orderId i tokena
	var tx Transaction
	err = transactionsCollection.FindOne(context.TODO(), bson.M{
		"transactionId": notification.Order.OrderId,
		"callbackToken": token,
	}).Decode(&tx)
	if err != nil {
		log.Println("❌ Nie znaleziono transakcji lub token niepoprawny:", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid transaction or token"})
		return
	}

	// 🔄 Aktualizujemy status transakcji
	filter := bson.M{"transactionId": tx.TransactionID}
	update := bson.M{"$set": bson.M{"status": notification.Order.Status}}

	_, err = transactionsCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Println("❌ Błąd aktualizacji transakcji:", err)
	}

	// ✅ Jeśli zapłacono, wykonujemy przypisaną komendę
	if notification.Order.Status == "COMPLETED" {
		player := strings.Split(notification.Order.Buyer.Email, "@")[0]

		for _, p := range tx.Products {
			var product Product
			err := productsCollection.FindOne(context.TODO(), bson.M{"id": p.ID}).Decode(&product)
			if err != nil {
				log.Printf("❌ Nie znaleziono produktu o ID %s: %v", p.ID, err)
				continue
			}
			log.Printf("✅ Wykonanie komendy dla produktu %s: %s", p.Name, product.Command)
			for _, server := range p.Servers {
				finalCommand := strings.ReplaceAll(product.Command, "%player%", player)
				go func(srv string, cmd string) {
					srv = strings.ToLower(srv) // ⬅️ zmiana na lowercase tutaj
					log.Printf("🔌 Wysyłanie komendy '%s' na serwer %s", cmd, srv)
					err := SendCommandToMinecraftToServer(srv, cmd)
					if err != nil {
						log.Printf("❌ Błąd przy wysyłaniu komendy na serwer %s: %v", srv, err)
					} else {
						log.Printf("✅ Komenda wysłana na serwer %s: %s", srv, cmd)
					}
				}(server, finalCommand)
			}
		}
	}

	c.Status(http.StatusOK)
}

// Get user group based on serverName
func GetUserGroupHandler(c *gin.Context) {
	serverName := c.Query("server")
	if serverName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'server' parameter"})
		return
	}

	playerName := c.Query("name")
	if playerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' parameter"})
		return
	}

	// Connect to the appropriate MariaDB database
	db, err := ConnectMariaDB(serverName)
	if err != nil {
		log.Printf("⚠️ Database connection error for server: %s | Error: %v", serverName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	defer db.Close() // Ensure proper cleanup

	// Query to fetch user group details
	query := `SELECT UUID, LASTNAME, PRIMARYGROUP, COALESCE(SUBGROUPS, ''), COALESCE(PERMISSIONS, ''), COALESCE(INFO, ''), TIMESTAMP 
              FROM GROUPMANAGER_WORLD_USERS 
              WHERE LASTNAME = ? 
              LIMIT 1`

	var userGroup UserGroup
	err = db.QueryRow(query, playerName).Scan(
		&userGroup.UUID,
		&userGroup.LastName,
		&userGroup.PrimaryGroup,
		&userGroup.SubGroups,
		&userGroup.Permissions,
		&userGroup.Info,
		&userGroup.TimeStamp,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			log.Printf("⚠️ Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Return user group data in JSON
	c.JSON(http.StatusOK, userGroup)
}

// Rejestracja uzytkownika
func RegisterHandler(c *gin.Context) {
	var req User
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var existingUser bson.M
	err := usersCollection.FindOne(context.TODO(), bson.M{"playerName": req.PlayerName}).Decode(&existingUser)
	if err == nil {
		log.Printf("User already registered: %s", req.PlayerName)
		c.JSON(http.StatusConflict, gin.H{"error": "User already registered"})
		return
	}

	hashedPassword, _ := HashPassword(req.Password)
	newUser := bson.M{
		"playerName":   req.PlayerName,
		"password":     hashedPassword,
		"registeredAt": time.Now().Unix(),
		"lastLogin":    0,
		"lastIP":       "",
	}

	_, err = usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	log.Printf("User registered successfully: %s", req.PlayerName)
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Logowanie uzytkownika
// Logowanie uzytkownika
func LoginHandler(c *gin.Context) {

	// Reczne odczytanie ciala zadania
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("? Error reading request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	//log.Printf("?? Raw request body: %s", string(rawBody)) // Logowanie surowego JSON

	var req User
	err = json.Unmarshal(rawBody, &req) // Reczne dekodowanie JSON
	if err != nil {
		log.Printf("? Error decoding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	log.Printf("? Received login request for playerName: %s", req.PlayerName)
	identifier := c.ClientIP() + "_" + req.PlayerName // można też samo IP

	if isLockedOut(identifier) {
		log.Printf("❌ Login blocked for %s – too many attempts", identifier)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Zbyt wiele nieudanych prób. Spróbuj ponownie za minutę."})
		return
	}

	// Logowanie przed zapytaniem do bazy
	log.Printf("?? Searching for playerName: %s in MongoDB", req.PlayerName)

	var user bson.M
	err = usersCollection.FindOne(context.TODO(), bson.D{
		{"playerName", req.PlayerName},
	}).Decode(&user)

	if err != nil {
		log.Printf("? User not found in MongoDB for playerName: %s | Error: %v", req.PlayerName, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	log.Printf("? User found: %v", user)

	// Sprawdzenie poprawnosci hasla
	if !CheckPassword(req.Password, user["password"].(string)) {
		log.Printf("? Invalid password for user: %s", req.PlayerName)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		recordFailedAttempt(identifier)
		return
	}
	resetLoginAttempts(identifier)

	// Generowanie JWT
	token, err := GenerateJWT(req.PlayerName)
	if err != nil {
		log.Printf("? Error generating JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Printf("? Login successful for playerName: %s", req.PlayerName)

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Pobieranie danych uzytkownika
func UserHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("? Nie udalo sie pobrac `user_id` z kontekstu!")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	log.Printf("?? Pobieranie danych uzytkownika dla user_id: %s", userID)

	var user bson.M
	err := usersCollection.FindOne(context.TODO(), bson.M{"playerName": userID}).Decode(&user)
	if err != nil {
		log.Printf("? Uzytkownik nie znaleziony w bazie dla `playerName`: %s | Error: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	//log.Printf("? Uzytkownik znaleziony: %v", user)

	c.JSON(http.StatusOK, gin.H{
		"playerName": user["playerName"],
		"lastLogin":  user["lastLogin"],
		"lastIP":     user["lastIP"],
	})
}
func CheckPlayerOnlineHandler(c *gin.Context) {
	type Req struct {
		Server     string `json:"server"`
		PlayerName string `json:"playerName"`
	}
	// Odczytanie ciała zapytania
	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("❌ Błąd parsowania JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// Sprawdzenie, czy serwer jest znany
	log.Printf("🔍 Sprawdzanie statusu gracza %s na serwerze %s", req.PlayerName, req.Server)

	isOnline, err := CheckPlayerStatus(req.Server, req.PlayerName)
	if err != nil {
		log.Printf("❌ Błąd sprawdzania statusu: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Status gracza %s na %s: %t", req.PlayerName, req.Server, isOnline)
	c.JSON(http.StatusOK, gin.H{
		"player":   req.PlayerName,
		"server":   req.Server,
		"isOnline": isOnline,
	})
}

// Odnawianie tokena JWT
func RenewTokenHandler(c *gin.Context) {
	ip := c.ClientIP()

	// Pobieramy token z kontekstu (ustawiony przez AuthMiddleware)
	tokenString, exists := c.Get("token_string")
	if !exists {
		ElasticWriter.WriteWithIP("❌ Brak tokena w kontekście.", ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	ElasticWriter.WriteWithIP("🔄 Próba odnawiania tokena.", ip)

	// Odnawiamy token
	newToken, err := RenewJWT(tokenString.(string))
	if err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd odnawiania tokena: %v", err), ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ElasticWriter.WriteWithIP("✅ Token został pomyślnie odnowiony.", ip)

	c.JSON(http.StatusOK, gin.H{
		"token":   newToken,
		"message": "Token renewed successfully",
	})
}
