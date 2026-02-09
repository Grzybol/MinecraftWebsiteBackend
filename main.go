package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Zezwalamy tylko zaufanym domenom
		if origin == "https://bestservers.fun" || origin == "https://boxpvp.top" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

var ElasticWriter *ElasticLogWriter

func main() {
	// Connect to MongoDB

	db, err := ConnectMongoDB()
	if err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}
	SeedProducts(db)
	defer db.Client().Disconnect(nil)

	// Connect to ELK
	elasticCfg := LoadElasticConfig()
	elasticSender := NewElasticSender(elasticCfg)
	ElasticWriter = NewElasticLogWriter(elasticSender) // ⬅️ bufujący writer
	//log.SetFlags(0) // opcjonalnie: usuwa timestampy z logów, bo Elasticsearch i tak ma je w JSON
	//log.SetOutput(&ElasticLogWriter{elastic: elasticSender})
	logger := NewElasticLogWriter(elasticSender) // ⬅️ bufujący writer
	log.SetOutput(logger)                        // ⬅️ przechwytujemy logi
	defer logger.Stop()                          // ⬅️ flush przy zamykaniu

	// Create router
	r := gin.Default()
	r.Use(CORSMiddleware())
	//r.Use(RateLimitMiddleware(10 * time.Millisecond))
	// 🔧 Dodaj middleware z przekazaniem db
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("elastic", elasticSender)
		c.Next()
	})

	// API Endpoints
	//r.POST("/api/register", RateLimitMiddleware(1*time.Millisecond), RegisterHandler)
	r.POST("/api/login", RateLimitMiddleware(1*time.Millisecond), LoginHandler)
	r.GET("/api/user", AuthMiddleware(db), UserHandler)
	r.GET("/api/usergroup", RateLimitMiddleware(1*time.Millisecond), AuthMiddleware(db), GetUserGroupHandler)
	r.POST("/api/renew", RateLimitMiddleware(5*time.Millisecond), AuthMiddleware(db), RenewTokenHandler)
	//r.POST("/api/processCommand", BackendAuthMiddleware(), processCommandHandler) // New endpoint for processing commands
	r.POST("/api/createTransaction", RateLimitMiddleware(10*time.Millisecond), AuthMiddleware(db), CreateTransactionHandler)
	r.POST("/api/paymentCallback", PaymentCallbackHandler)
	r.POST("/api/checkPlayerOnline", RateLimitMiddleware(10*time.Millisecond), AuthMiddleware(db), CheckPlayerOnlineHandler)
	r.POST("/api/logout", AuthMiddleware(db), LogoutHandler)

	r.GET("/stats/worlds", RateLimitMiddleware(5*time.Millisecond), GetWorldsHandler)
	r.GET("/stats/:world", RateLimitMiddleware(5*time.Millisecond), GetWorldStatsHandler)

	// New endpoints for player equipment
	r.GET("/api/getBackpack", RateLimitMiddleware(5*time.Millisecond), AuthMiddleware(db), GetBackpackHandler)
	r.GET("/api/getArmor", RateLimitMiddleware(5*time.Millisecond), AuthMiddleware(db), GetArmorHandler)
	r.GET("/api/getHotbar", RateLimitMiddleware(5*time.Millisecond), AuthMiddleware(db), GetHotbarHandler)

	// New endpoint for eco balance
	r.GET("/api/getEcoBalance", RateLimitMiddleware(5*time.Millisecond), AuthMiddleware(db), GetEcoBalanceHandler)

	// New endpoint for checking player balance
	r.GET("/api/checkPlayerBalance", RateLimitMiddleware(5*time.Millisecond), AuthMiddleware(db), CheckPlayerBalanceHandler)
	//endpoint do procesowania platnosci kodem
	r.POST("/api/processCodePayment", RateLimitMiddleware(10000*time.Millisecond), AuthMiddleware(db), ProcessCodePaymentHandler)
	//endpoint dla placenia punktami
	r.POST("/api/processPointsPayment", RateLimitMiddleware(10*time.Millisecond), AuthMiddleware(db), ProcessPointsPaymentHandler)

	//Elo endpoints
	r.GET("/api/getEloRanking", RateLimitMiddleware(10*time.Millisecond), AuthMiddleware(db), GetEloRankingHandler)
	r.GET("/api/getPlayerElo", RateLimitMiddleware(10*time.Millisecond), AuthMiddleware(db), GetPlayerEloHandler)
	r.GET("/api/getPlayerByRank", RateLimitMiddleware(10*time.Millisecond), AuthMiddleware(db), GetPlayerByRankHandler)

	//Mobile app API endpoints
	r.GET("/api/barcodeinfo", RateLimitMiddleware(100*time.Millisecond), AuthMiddleware(db), BarcodeInfoHandler)

	// Clicker game endpoints
	r.POST("/api/progress/sync", RateLimitMiddleware(10*time.Millisecond), AuthMiddleware(db), SyncProgressHandler)
	r.GET("/api/progress", RateLimitMiddleware(5*time.Millisecond), AuthMiddleware(db), GetProgressHandler)
	r.GET("/api/leaderboard", RateLimitMiddleware(10*time.Millisecond), AuthMiddleware(db), GetLeaderboardHandler)

	// Start the server on HTTPS

	log.Println("✅ Server running on HTTPS port 8443")
	err = r.RunTLS(":8443", "/home/wwwbackend/fullchain.pem", "/home/wwwbackend/privkey.pem")
	//log.Println("✅ Server running on HTTP port 8080")
	//err = r.Run(":8080")
	if err != nil {
		log.Fatal("❌ Failed to run server:", err)
	}
}
