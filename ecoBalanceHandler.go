package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetEcoBalanceHandler(c *gin.Context) {
	ip := c.ClientIP() // pobieramy IP klienta na początku

	server := c.Query("server")
	if server == "" {
		ElasticWriter.WriteWithIP("❌ Brak nazwy serwera w query param.", ip)
		c.JSON(400, gin.H{"error": "missing server name"})
		return
	}

	playerName := c.Query("player")
	if playerName == "" {
		ElasticWriter.WriteWithIP("❌ Brak nazwy gracza w query param.", ip)
		c.JSON(400, gin.H{"error": "missing player name"})
		return
	}

	serverKey := strings.ToLower(server) // np. survival → survival
	socketPath := "/tmp/plugin_minecraft_ecobalance.sock"

	token, ok := ServerTokens[serverKey]
	if !ok {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Brak tokena dla serwera '%s' (klucz: %s)", server, serverKey), ip)
		c.JSON(403, gin.H{"error": "unauthorized or unknown server"})
		return
	}

	ElasticWriter.WriteWithIP(fmt.Sprintf("💰 Eco Balance Request → server: %s | player: %s | socket: %s | token: %s", serverKey, playerName, socketPath, token), ip)

	jsonResponse, err := getEcoBalanceSocketResponse(socketPath, token, playerName, ip)

	if err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd od socketu (%s): %v", socketPath, err), ip)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Odpowiedź z socketa (eco balance): %s", truncateForLog(jsonResponse, 100)), ip)

	c.Data(200, "application/json", []byte(jsonResponse))
}

func getEcoBalanceSocketResponse(socketPath, token, playerName, ip string) (string, error) {
	ElasticWriter.WriteWithIP(fmt.Sprintf("📨 Próba połączenia z socketem: %s", socketPath), ip)

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd połączenia z socketem: %s | %v", socketPath, err), ip)
		return "", fmt.Errorf("failed to connect to socket: %w", err)
	}
	defer func() {
		ElasticWriter.WriteWithIP("🔌 Zamykam połączenie z socketem.", ip)
		conn.Close()
	}()

	message := fmt.Sprintf("%s:%s\n", token, playerName)
	ElasticWriter.WriteWithIP(fmt.Sprintf("📤 Wysyłam do socketu: %s", message), ip)

	_, err = fmt.Fprintf(conn, message)
	if err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd przy wysyłaniu do socketu: %v", err), ip)
		return "", fmt.Errorf("failed to send data to socket: %w", err)
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd przy odczycie z socketu: %v", err), ip)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	ElasticWriter.WriteWithIP(fmt.Sprintf("📥 Odczytano odpowiedź z socketu: %s", response), ip)
	return strings.TrimSpace(response), nil
}
