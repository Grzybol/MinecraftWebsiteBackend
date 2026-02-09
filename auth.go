package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type RevokedToken struct {
	Token     string             `bson:"token"`
	ExpiresAt primitive.DateTime `bson:"expires_at"`
}

var jwtSecret = []byte("your-secret-key")
var backendJwtSecret = []byte("your-backend-secret-key")

// Hashowanie hasła
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Sprawdzenie poprawności hasła
func CheckPassword(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// Generowanie tokena JWT
func GenerateJWT(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 10).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// Odnawianie tokena JWT
func RenewJWT(tokenString string) (string, error) {
	// Parsujemy istniejący token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	// Sprawdzamy czy token jest ważny
	if !token.Valid {
		return "", errors.New("token is not valid")
	}

	// Wyciągamy claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	// Sprawdzamy czy token nie wygasł
	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", errors.New("invalid expiration claim")
	}

	// Sprawdzamy czy token nie wygasł (z małym marginesem - 1 minuta)
	if time.Unix(int64(exp), 0).Before(time.Now().Add(-time.Minute)) {
		return "", errors.New("token has expired")
	}

	// Wyciągamy user_id
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("invalid user_id claim")
	}

	// Generujemy nowy token z tym samym user_id
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 10).Unix(),
	})

	return newToken.SignedString(jwtSecret)
}

// Weryfikacja tokena JWT
func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["user_id"].(string), nil
	}
	return "", errors.New("invalid token")
}

/*
	func AuthMiddleware() gin.HandlerFunc {
		return func(c *gin.Context) {
			tokenString := c.GetHeader("Authorization")

			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
				c.Abort()
				return
			}

			log.Printf("?? Otrzymano token: %s", tokenString)

			// Usun "Bearer " z poczatku tokena
			parts := strings.Split(tokenString, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Println("? Zly format tokena")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
				c.Abort()
				return
			}
			tokenString = parts[1]

			userID, err := ValidateJWT(tokenString)
			if err != nil {
				log.Printf("? Blad walidacji JWT: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			log.Printf("? Token poprawny dla user_id: %s", userID)

			c.Set("user_id", userID)
			c.Next()
		}

}
*/
func AuthMiddleware(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		parts := strings.Split(tokenString, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}
		tokenString = parts[1]

		// 🔒 Sprawdź czy token jest zablokowany
		isRevoked, err := IsTokenRevoked(tokenString, db)
		if err != nil {
			log.Printf("❌ Błąd podczas sprawdzania blackliste: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			c.Abort()
			return
		}
		if isRevoked {
			log.Printf("❌ Token jest unieważniony (blacklista)")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
			c.Abort()
			return
		}

		userID, err := ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Szukamy użytkownika po UUID (czyli user_id z tokena)
		var user User
		err = db.Collection("users").FindOne(context.TODO(), bson.M{"playerName": userID}).Decode(&user)
		if err != nil {
			log.Println("❌ Nie znaleziono użytkownika o playerName:", userID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("playerName", user.PlayerName) // teraz działa
		c.Set("token_string", tokenString)
		c.Next()

	}
}

func GenerateBackendJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized_backend": true,
		"exp":                time.Now().Add(time.Minute * 10).Unix(), // Krótszy czas życia dla tego typu tokenów
	})
	return token.SignedString(backendJwtSecret)
}

func ValidateBackendJWT(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return backendJwtSecret, nil
	})
	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if authorized, ok := claims["authorized_backend"].(bool); ok && authorized {
			return true, nil
		}
	}
	return false, errors.New("invalid backend token")
}
func BackendAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No backend token provided"})
			c.Abort()
			return
		}

		// Usun "Bearer " z poczatku tokena, jeśli jest
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		if _, err := ValidateBackendJWT(tokenString); err != nil {
			log.Printf("Backend token validation error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid backend token"})
			c.Abort()
			return
		}

		log.Println("Backend token validated successfully")
		c.Next()
	}
}
func IsTokenRevoked(tokenString string, db *mongo.Database) (bool, error) {
	collection := db.Collection("revoked_tokens")

	filter := bson.M{"token": tokenString}
	var result RevokedToken
	err := collection.FindOne(context.TODO(), filter).Decode(&result)

	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
