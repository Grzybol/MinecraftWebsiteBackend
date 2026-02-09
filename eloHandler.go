package main

import (
	"context"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GET /api/getEloRanking
func GetEloRankingHandler(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("player_points_boxpvp")

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("❌ Błąd podczas pobierania rankingu:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.TODO())

	type Player struct {
		Name        string  `json:"name"`
		Points      float64 `json:"points"`
		RankingType string  `json:"rankingType"`
	}

	var players []Player

	for cursor.Next(context.TODO()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		name, _ := doc["playerName"].(string)
		points, _ := doc["points"].(float64)
		rankingType, _ := doc["rankingType"].(string)
		players = append(players, Player{
			Name:        name,
			Points:      points,
			RankingType: rankingType,
		})
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Points > players[j].Points
	})

	log.Println("✅ Zwrócono ranking Elo (posortowany)")
	c.JSON(http.StatusOK, players)
}

// GET /api/getPlayerElo?name=Grzybol
func GetPlayerEloHandler(c *gin.Context) {
	name := c.Query("name")
	rankingType := c.DefaultQuery("rankingType", "main") // domyślnie "main"

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing name param"})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("player_points_boxpvp")

	// Filtruj tylko dla danego rankingType
	cursor, err := collection.Find(context.TODO(), bson.M{"rankingType": rankingType})
	if err != nil {
		log.Println("❌ Błąd podczas pobierania danych gracza:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.TODO())

	type Player struct {
		Name        string  `json:"name"`
		Points      float64 `json:"points"`
		RankingType string  `json:"rankingType"`
	}
	var players []Player

	for cursor.Next(context.TODO()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		playerName, _ := doc["playerName"].(string)
		points, _ := doc["points"].(float64)

		players = append(players, Player{
			Name:        playerName,
			Points:      points,
			RankingType: rankingType,
		})
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Points > players[j].Points
	})

	for i, p := range players {
		if p.Name == name {
			log.Println("✅ Zwrócono dane gracza:", name)
			c.JSON(http.StatusOK, gin.H{
				"rank":        i + 1,
				"points":      p.Points,
				"name":        p.Name,
				"rankingType": p.RankingType,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Player not found in ranking"})
}

// GET /api/getPlayerByRank?rank=5
func GetPlayerByRankHandler(c *gin.Context) {
	rankStr := c.Query("rank")
	rank, err := strconv.Atoi(rankStr)
	if err != nil || rank < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rank"})
		return
	}
	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("player_points_boxpvp")

	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("❌ Błąd podczas pobierania danych rankingu:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer cursor.Close(context.TODO())

	type Player struct {
		Name   string  `json:"name"`
		Points float64 `json:"points"`
	}
	var players []Player

	for cursor.Next(context.TODO()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		name, _ := doc["playerName"].(string)
		points, _ := doc["points"].(float64)
		players = append(players, Player{Name: name, Points: points})
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Points > players[j].Points
	})

	if rank > len(players) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rank exceeds player count"})
		return
	}

	player := players[rank-1]
	log.Println("✅ Zwrócono gracza na pozycji", rank)
	c.JSON(http.StatusOK, gin.H{"name": player.Name, "points": player.Points})
}
