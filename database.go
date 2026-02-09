package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	usersCollection        *mongo.Collection
	transactionsCollection *mongo.Collection
	paymentCodesCollection *mongo.Collection // 🔥 NOWE
	barcodeCacheCollection *mongo.Collection // Kolekcja do cache'owania kodów kreskowych
)

// Connects to MongoDB
func ConnectMongoDB() (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(GetMongoURI())
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	log.Println("? Connected to MongoDB")
	db := client.Database(DatabaseName)
	usersCollection = db.Collection("users")
	transactionsCollection = db.Collection("transactions") // 🔥 NOWE
	paymentCodesCollection = db.Collection("paymentCodes")
	barcodeCacheCollection = db.Collection("barcode_cache")
	log.Println("✅ MongoDB barcode_cache initialized")
	log.Println("? MongoDB collections initialized")

	return db, nil
}

// Function to dynamically connect to a MariaDB database based on serverName
func ConnectMariaDB(serverName string) (*sql.DB, error) {
	baseURI := GetMariaDBBaseURI()
	fullURI := fmt.Sprintf("%s%s", baseURI, serverName)

	db, err := sql.Open("mysql", fullURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open MariaDB connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close() // Ensure cleanup on failure
		return nil, fmt.Errorf("failed to ping MariaDB: %w", err)
	}

	log.Printf("? Connected to MariaDB - Database: %s", serverName)
	return db, nil
}
