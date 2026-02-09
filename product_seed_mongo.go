package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var productsCollection *mongo.Collection

type Product struct {
	ID       string  `bson:"id"`
	Name     string  `bson:"name"`
	Price    float64 `bson:"price"`
	Currency string  `bson:"currency"`
	Type     string  `bson:"type"`
	Server   string  `bson:"server"`
	Command  string  `bson:"command"`
}

func SeedProducts(db *mongo.Database) {
	productsCollection = db.Collection("products")

	_, err := productsCollection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		log.Println("❌ Błąd czyszczenia kolekcji produktów:", err)
		return
	}

	products := []interface{}{
		Product{"premium_boost", "Boost Premium", 9.99, "PLN", "boost", "global", "give %player% boost"},
		Product{"vip_boxpvp", "Ranga VIP (30 dni)", 15.00, "PLN", "rank", "BoxPvP", "br add %player% VIP 30 d"},
		Product{"mvp_boxpvp", "Ranga MVP (30 dni)", 20.00, "PLN", "rank", "BoxPvP", "br add %player% MVP 30 d"},
		Product{"pro_boxpvp", "Ranga PRO (30 dni)", 25.00, "PLN", "rank", "BoxPvP", "br add %player% PRO 30 d"},
		Product{"god_boxpvp", "Ranga GOD (30 dni)", 30.00, "PLN", "rank", "BoxPvP", "br add %player% GOD 30 d"},
		Product{"death_key_boxpvp", "Death Key x2", 10.00, "PLN", "key", "BoxPvP", "excellentcrates key give %player% death 2"},
		///excellentcrates key give grzybol death 2
		Product{"vip_survival", "Ranga VIP (30 dni)", 15.00, "PLN", "rank", "Survival", "br add %player% vip 30d"},
		Product{"vip_skygen", "Ranga VIP (30 dni)", 15.00, "PLN", "rank", "SkyGen", "br add %player% vip 30 d"},
		Product{"vip_questmc", "Ranga VIP (30 dni)", 15.00, "PLN", "rank", "QuestMC", "br add %player% vip 30 d"},
	}

	_, err = productsCollection.InsertMany(context.TODO(), products)
	if err != nil {
		log.Println("❌ Błąd dodawania produktów:", err)
	} else {
		log.Println("✅ Produkty zostały zapisane w MongoDB")
	}
}

func GetProductByID(id string) (*Product, error) {
	var product Product
	err := productsCollection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&product)
	if err != nil {
		return nil, fmt.Errorf("❌ Produkt o ID '%s' nie został znaleziony: %v", id, err)
	}
	return &product, nil
}
