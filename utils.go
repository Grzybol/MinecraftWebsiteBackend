package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func SendCommandToMinecraft(command string) error {
	const authToken = "your_default_token_here"

	conn, err := net.Dial("unix", "/tmp/plugin_minecraft.sock")
	if err != nil {
		return err
	}
	defer conn.Close()

	full := fmt.Sprintf("%s:%s", authToken, command)
	_, err = conn.Write([]byte(full + "\n"))
	return err
}

// Wysyła komendę do konkretnego socketa, używając odpowiedniego tokena
func SendCommandToMinecraftToServer(server string, command string) error {
	token, ok := ServerTokens[server]
	if !ok {
		return fmt.Errorf("brak tokena dla serwera %s", server)
	}

	socketPath := fmt.Sprintf("/tmp/plugin_%s.sock", strings.ToLower(server)) // np. /tmp/plugin_boxpvp.sock
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("błąd połączenia z %s: %w", socketPath, err)
	}
	defer conn.Close()

	fullCommand := fmt.Sprintf("%s:%s", token, command)
	_, err = conn.Write([]byte(fullCommand + "\n"))
	return err
}

// Sprawdza, czy gracz jest online, pytając serwer przez socket statusy
func CheckPlayerStatus(server string, playerName string) (bool, error) {
	server = strings.ToLower(server) // Upewniamy się, że serwer jest w małych literach
	token, ok := ServerTokens[server]
	if !ok {
		return false, fmt.Errorf("brak tokena dla serwera %s", server)
	}

	socketPath := fmt.Sprintf("/tmp/plugin_%s_status.sock", strings.ToLower(server)) // np. /tmp/plugin_boxpvp_status.sock
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return false, fmt.Errorf("błąd połączenia z socketem statusowym %s: %w", socketPath, err)
	}
	defer conn.Close()

	// Wysyłamy zapytanie w formacie: token:isOnline:nazwaGracza
	query := fmt.Sprintf("%s:isOnline:%s\n", token, playerName)
	_, err = conn.Write([]byte(query))
	if err != nil {
		return false, fmt.Errorf("błąd wysyłania zapytania: %w", err)
	}

	// Odczytujemy odpowiedź
	buffer := make([]byte, 128)
	n, err := conn.Read(buffer)
	if err != nil {
		return false, fmt.Errorf("błąd odczytu odpowiedzi: %w", err)
	}

	response := strings.TrimSpace(string(buffer[:n]))
	switch response {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "unauthorized":
		return false, fmt.Errorf("nieautoryzowane zapytanie")
	default:
		return false, fmt.Errorf("nieznana odpowiedź: %s", response)
	}
}

const maxLoginAttempts = 10
const lockDuration = 60 * time.Second

func isLockedOut(identifier string) bool {
	loginLock.Lock()
	defer loginLock.Unlock()

	attempt, exists := loginAttempts[identifier]
	if !exists {
		return false
	}

	if time.Since(attempt.LockedAt) < lockDuration {
		return true
	}

	// Resetuj po odblokowaniu
	delete(loginAttempts, identifier)
	return false
}

func recordFailedAttempt(identifier string) {
	loginLock.Lock()
	defer loginLock.Unlock()

	attempt, exists := loginAttempts[identifier]
	if !exists {
		loginAttempts[identifier] = &LoginAttempt{Count: 1}
		return
	}

	attempt.Count++
	if attempt.Count >= maxLoginAttempts {
		attempt.LockedAt = time.Now()
	}
}

func resetLoginAttempts(identifier string) {
	loginLock.Lock()
	defer loginLock.Unlock()
	delete(loginAttempts, identifier)
}
