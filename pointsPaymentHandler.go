package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CheckPlayerBalanceHandler(c *gin.Context) {
	log.Println("🔍 Sprawdzanie balansu gracza")
	playerName := c.Query("playerName")
	if playerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing playerName"})
		log.Println("❌ Błąd: Missing playerName")
		return
	}
	log.Printf("Sprawdzanie balansu gracza: %s", playerName)

	db := c.MustGet("db").(*mongo.Database)
	users := db.Collection("users")

	var user User
	err := users.FindOne(context.Background(), bson.M{"playerName": playerName}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		log.Printf("❌ Błąd: User not found: %s", playerName)
		return
	}
	log.Printf("Znaleziono użytkownika: %s, balans: %d", user.PlayerName, user.WebsiteBalance)
	c.JSON(http.StatusOK, gin.H{
		"playerName": playerName,
		"balance":    user.WebsiteBalance,
	})
	log.Printf("✅ Zwrócono balans gracza: %s, balans: %d", playerName, user.WebsiteBalance)
}

func ProcessCodePaymentHandler(c *gin.Context) {
	log.Println("🔍 Przetwarzanie płatności kodem")
	type RequestBody struct {
		Code string `json:"code"`
	}

	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil || len(strings.TrimSpace(req.Code)) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code"})
		log.Println("❌ Błąd: Invalid code")
		return
	}

	codeStr := strings.ToUpper(strings.TrimSpace(req.Code))

	db := c.MustGet("db").(*mongo.Database)
	codes := db.Collection("paymentCodes")
	users := db.Collection("users")

	// Sprawdź czy kod istnieje i nie jest użyty
	var code PaymentCode
	err := codes.FindOne(context.Background(), bson.M{
		"code": codeStr,
		"used": false,
	}).Decode(&code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or already used code"})
		log.Println("❌ Błąd: Invalid or already used code")
		return
	}
	log.Printf("Znaleziono kod: %s, wartość: %d", code.Code, code.Value)

	// Pobierz playerName z tokena (AuthMiddleware)
	playerName := c.MustGet("playerName").(string)

	// Zaktualizuj balans użytkownika
	filter := bson.M{"playerName": playerName}
	update := bson.M{"$inc": bson.M{"website_balance": code.Value}}
	_, err = users.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user balance"})
		log.Println("❌ Błąd: Failed to update user balance")
		return
	}
	log.Printf("Zaktualizowano balans użytkownika: %s, nowy balans: %d", playerName, code.Value)

	// Oznacz kod jako użyty
	_, err = codes.UpdateOne(context.Background(), bson.M{"_id": code.ID}, bson.M{
		"$set": bson.M{
			"used":   true,
			"usedBy": playerName,
			"usedAt": primitive.NewDateTimeFromTime(time.Now()),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark code as used"})
		log.Println("❌ Błąd: Failed to mark code as used")
		return
	}
	log.Printf("Oznaczono kod jako użyty: %s, przez użytkownika: %s", code.Code, playerName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"value":   code.Value,
	})
	log.Printf("✅ Płatność kodem zakończona sukcesem: %s, wartość: %d", code.Code, code.Value)
	log.Printf("✅ Zwrócono odpowiedź: %s", code.Code)
}

func ProcessPointsPaymentHandler(c *gin.Context) {
	log.Println("🔍 Przetwarzanie płatności punktami")
	type RequestBody struct {
		ProductID string `json:"productId"`
	}

	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing productId"})
		log.Println("❌ Błąd: Missing productId")
		return
	}

	// Pobierz playerName z AuthMiddleware
	playerName := c.MustGet("playerName").(string)

	// Pobierz produkt po ID
	product, err := GetProductByID(req.ProductID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		log.Printf("❌ Błąd: Product not found: %s", req.ProductID)
		return
	}

	// Cena w punktach
	requiredPoints := int(product.Price * 10) // 1zł = 10 punktów

	db := c.MustGet("db").(*mongo.Database)
	users := db.Collection("users")
	log.Printf("Sprawdzanie balansu gracza: %s, wymagane punkty: %d", playerName, requiredPoints)
	// Pobierz użytkownika
	var user User
	err = users.FindOne(context.TODO(), bson.M{"playerName": playerName}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		log.Printf("❌ Błąd: User not found: %s", playerName)
		return
	}

	if user.WebsiteBalance < requiredPoints {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough points", "required": requiredPoints, "current": user.WebsiteBalance})
		log.Printf("❌ Błąd: Not enough points for user: %s, required: %d, current: %d", playerName, requiredPoints, user.WebsiteBalance)
		return
	}

	_, err = users.UpdateOne(context.TODO(), bson.M{"playerName": playerName}, bson.M{
		"$inc": bson.M{"website_balance": -requiredPoints},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		log.Printf("❌ Błąd: Failed to update balance for user: %s", playerName)
		return
	}
	log.Printf("Zaktualizowano balans użytkownika: %s, nowy balans: %d", playerName, user.WebsiteBalance-requiredPoints)
	// Wyślij komendę do serwera
	command := strings.ReplaceAll(product.Command, "%player%", playerName)
	server := strings.ToLower(product.Server)
	log.Printf("Wysyłanie komendy: %s na serwer: %s", command, server)
	go func() {
		err := SendCommandToMinecraftToServer(server, command)
		if err != nil {
			log.Printf("❌ Błąd przy wysyłaniu komendy: %v", err)
		} else {
			log.Printf("✅ Komenda '%s' wysłana na serwer '%s'", command, server)
		}

		// Komunikat say
		sayCmd := fmt.Sprintf(`say "Gracz %s właśnie zakupił %s za kwotę %.2f zł"`, playerName, product.Name, product.Price)
		err = SendCommandToMinecraftToServer(server, sayCmd)
		if err != nil {
			log.Printf("❌ Błąd przy wysyłaniu komendy say: %v", err)
		} else {
			log.Printf("✅ Komenda 'say' wysłana na serwer '%s'", server)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"usedPoints":   requiredPoints,
		"remaining":    user.WebsiteBalance - requiredPoints,
		"executedCmd":  command,
		"targetServer": server,
		"productName":  product.Name,
	})
}
