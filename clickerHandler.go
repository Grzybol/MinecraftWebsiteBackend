package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// calculateMaxPossibleIncome oblicza maksymalny możliwy przyrost zasobów
func calculateMaxPossibleIncome(verification ProgressVerification, deltaTime float64) (float64, float64, float64) {
	// Maksymalne kliknięcia w danym czasie
	maxClicks := float64(verification.MaxClicksPerSec) * deltaTime

	// Maksymalny przyrost z kliknięć
	maxClickIncome := maxClicks * verification.ClickDamage * verification.GoldBonus
	maxClickExp := maxClicks * verification.ClickExp * verification.ExpBonus

	// Maksymalny przyrost z DPS
	maxDPSIncome := verification.DPS * deltaTime * verification.GoldBonus
	maxDPSExp := verification.DPSExp * deltaTime * verification.ExpBonus

	// Maksymalny możliwy przyrost
	maxPossibleCoins := maxClickIncome + maxDPSIncome
	maxPossibleExp := maxClickExp + maxDPSExp

	// Maksymalny przyrost diamentów (z szansy na drop)
	maxPossibleDiamonds := maxClicks * verification.DiamondDropRate

	return maxPossibleCoins, maxPossibleExp, maxPossibleDiamonds
}

// verifyProgress weryfikuje czy postęp gracza jest możliwy
func verifyProgress(req SyncProgressRequest, existingProgress PlayerProgress) VerificationResult {
	// Jeśli brak danych o poprzednim stanie, używamy danych z bazy
	previousCoins := req.PreviousCoins
	previousExperience := req.PreviousExperience
	previousDiamonds := req.PreviousDiamonds
	previousLevel := req.PreviousLevel
	previousTimestamp := req.PreviousTimestamp

	// Jeśli nie podano poprzedniego stanu, używamy danych z bazy
	if previousCoins == 0 {
		previousCoins = existingProgress.Coins
		previousExperience = existingProgress.Experience
		previousDiamonds = existingProgress.Diamonds
		previousLevel = existingProgress.Level
		previousTimestamp = existingProgress.LastSync
	}

	// Oblicz czas między synchronizacjami
	deltaTime := req.Timestamp.Sub(previousTimestamp).Seconds()

	// Sprawdź czy delta time jest rozsądna (nie więcej niż 1 godzina)
	if deltaTime > 3600 {
		return VerificationResult{
			IsValid: false,
			Reason:  "Delta time too large (> 1 hour)",
		}
	}

	// Sprawdź czy nie ma rollbacku
	if req.Coins < previousCoins {
		return VerificationResult{
			IsValid: false,
			Reason:  "Coins rollback detected",
		}
	}

	if req.Experience < previousExperience {
		return VerificationResult{
			IsValid: false,
			Reason:  "Experience rollback detected",
		}
	}

	// Oblicz przyrost zasobów
	coinsGain := req.Coins - previousCoins
	expGain := req.Experience - previousExperience
	diamondsGain := req.Diamonds - previousDiamonds

	// Oblicz parametry weryfikacji na podstawie levelu i ulepszeń
	verification := calculateVerificationParams(req.Level, req.Upgrades)

	// Oblicz maksymalny możliwy przyrost
	maxPossibleCoins, maxPossibleExp, maxPossibleDiamonds := calculateMaxPossibleIncome(verification, deltaTime)

	// Dodaj margines bezpieczeństwa
	margin := verification.Margin
	maxPossibleCoinsWithMargin := maxPossibleCoins * (1 + margin)
	maxPossibleExpWithMargin := maxPossibleExp * (1 + margin)
	maxPossibleDiamondsWithMargin := maxPossibleDiamonds * (1 + margin)

	// Sprawdź czy przyrost nie przekracza maksimum
	if coinsGain > maxPossibleCoinsWithMargin {
		return VerificationResult{
			IsValid:          false,
			Reason:           fmt.Sprintf("Coins gain too high: %.2f > %.2f", coinsGain, maxPossibleCoinsWithMargin),
			MaxPossibleCoins: maxPossibleCoinsWithMargin,
			ActualCoinsGain:  coinsGain,
			DeltaTime:        deltaTime,
		}
	}

	if expGain > maxPossibleExpWithMargin {
		return VerificationResult{
			IsValid:        false,
			Reason:         fmt.Sprintf("Experience gain too high: %.2f > %.2f", expGain, maxPossibleExpWithMargin),
			MaxPossibleExp: maxPossibleExpWithMargin,
			ActualExpGain:  expGain,
			DeltaTime:      deltaTime,
		}
	}

	if diamondsGain > maxPossibleDiamondsWithMargin {
		return VerificationResult{
			IsValid:             false,
			Reason:              fmt.Sprintf("Diamonds gain too high: %.2f > %.2f", diamondsGain, maxPossibleDiamondsWithMargin),
			MaxPossibleDiamonds: maxPossibleDiamondsWithMargin,
			ActualDiamondsGain:  diamondsGain,
			DeltaTime:           deltaTime,
		}
	}

	// Sprawdź czy level nie zmienił się zbyt drastycznie
	levelGain := req.Level - previousLevel
	if levelGain > 5 && deltaTime < 60 { // Więcej niż 5 leveli w mniej niż minutę
		return VerificationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Level gain too high: %d levels in %.2f seconds", levelGain, deltaTime),
		}
	}

	return VerificationResult{
		IsValid:             true,
		MaxPossibleCoins:    maxPossibleCoinsWithMargin,
		MaxPossibleExp:      maxPossibleExpWithMargin,
		MaxPossibleDiamonds: maxPossibleDiamondsWithMargin,
		ActualCoinsGain:     coinsGain,
		ActualExpGain:       expGain,
		ActualDiamondsGain:  diamondsGain,
		DeltaTime:           deltaTime,
	}
}

// calculateVerificationParams oblicza parametry weryfikacji na podstawie levelu i ulepszeń
func calculateVerificationParams(level int, upgrades []Upgrade) ProgressVerification {
	// Podstawowe wartości
	baseClickDamage := 10.0
	baseDPS := 5.0
	baseClickExp := 1.0
	baseDPSExp := 0.5
	baseDiamondDropRate := 0.001 // 0.1% szansy na diament

	// Bonusy z levelu
	levelBonus := float64(level) * 0.1

	// Bonusy z ulepszeń
	upgradeBonus := 0.0
	for _, upgrade := range upgrades {
		switch upgrade.ID {
		case "wooden_pickaxe":
			upgradeBonus += float64(upgrade.Level) * 2.0
		case "stone_pickaxe":
			upgradeBonus += float64(upgrade.Level) * 5.0
		case "iron_pickaxe":
			upgradeBonus += float64(upgrade.Level) * 10.0
		case "diamond_pickaxe":
			upgradeBonus += float64(upgrade.Level) * 25.0
		case "gold_bonus":
			upgradeBonus += float64(upgrade.Level) * 0.1
		case "exp_bonus":
			upgradeBonus += float64(upgrade.Level) * 0.05
		}
	}

	return ProgressVerification{
		ClickDamage:     baseClickDamage + levelBonus + upgradeBonus,
		DPS:             baseDPS + (levelBonus * 0.5) + (upgradeBonus * 0.3),
		MaxClicksPerSec: 25, // Ustalony limit
		ClickExp:        baseClickExp + (levelBonus * 0.1),
		DPSExp:          baseDPSExp + (levelBonus * 0.05),
		DiamondDropRate: baseDiamondDropRate + (float64(level) * 0.0001),
		GoldBonus:       1.0 + (upgradeBonus * 0.1),
		ExpBonus:        1.0 + (upgradeBonus * 0.05),
		Margin:          0.05, // 5% margines bezpieczeństwa
	}
}

// SyncProgressHandler synchronizuje postęp gracza
func SyncProgressHandler(c *gin.Context) {
	ip := c.ClientIP()
	playerName := c.MustGet("playerName").(string)

	ElasticWriter.WriteWithIP(fmt.Sprintf("🔄 Sync Progress Request → player: %s", playerName), ip)

	var req SyncProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd parsowania JSON: %v", err), ip)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Walidacja danych
	if req.Level < 0 || req.Experience < 0 || req.Coins < 0 || req.Diamonds < 0 {
		ElasticWriter.WriteWithIP("❌ Nieprawidłowe wartości w request (ujemne liczby)", ip)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid values in request"})
		return
	}

	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("player_progress")

	// Sprawdź czy gracz już istnieje
	var existingProgress PlayerProgress
	filter := bson.M{"playerName": playerName}
	err := collection.FindOne(context.TODO(), filter).Decode(&existingProgress)

	if err == mongo.ErrNoDocuments {
		// Nowy gracz - utwórz dokument bez weryfikacji
		ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Nowy gracz: %s - pomijam weryfikację", playerName), ip)
	} else if err != nil {
		// Błąd bazy danych
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd bazy danych: %v", err), ip)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// Weryfikuj postęp dla istniejącego gracza
		verificationResult := verifyProgress(req, existingProgress)

		if !verificationResult.IsValid {
			// Loguj podejrzany postęp
			ElasticWriter.WriteWithIP(fmt.Sprintf("🚨 Podejrzany postęp dla gracza %s: %s", playerName, verificationResult.Reason), ip)

			// Możesz tutaj dodać dodatkową logikę:
			// 1. Odrzuć synchronizację
			// 2. Przyciąć do maksimum
			// 3. Zaloguj do bazy danych

			// Na razie odrzucamy synchronizację
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Progress verification failed",
				"details": verificationResult,
			})
			return
		}

		// Loguj informacje o weryfikacji
		ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Weryfikacja OK dla gracza %s: coins +%.2f, exp +%.2f, diamonds +%.2f w %.2fs",
			playerName, verificationResult.ActualCoinsGain, verificationResult.ActualExpGain,
			verificationResult.ActualDiamondsGain, verificationResult.DeltaTime), ip)
	}

	now := time.Now()

	// Przygotuj dane do zapisu
	progressData := PlayerProgress{
		PlayerName:            playerName,
		Level:                 req.Level,
		Experience:            req.Experience,
		ExperienceToNextLevel: req.ExperienceToNextLevel,
		Coins:                 req.Coins,
		Diamonds:              req.Diamonds,
		PremiumBalance:        req.PremiumBalance,
		BetterCoinBalance:     req.BetterCoinBalance,
		TotalCoinsEarned:      req.TotalCoinsEarned,
		TotalDiamondsEarned:   req.TotalDiamondsEarned,
		TotalClicks:           req.TotalClicks,
		TotalUpgrades:         req.TotalUpgrades,
		Upgrades:              req.Upgrades,
		LastSync:              now,
		LastDevice:            c.GetHeader("User-Agent"),
		LastIP:                ip,
		UpdatedAt:             now,
	}

	if err == mongo.ErrNoDocuments {
		// Nowy gracz - utwórz dokument
		progressData.CreatedAt = now
		_, err = collection.InsertOne(context.TODO(), progressData)
		if err != nil {
			ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd tworzenia postępu: %v", err), ip)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create progress"})
			return
		}
		ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Utworzono nowy postęp dla gracza: %s", playerName), ip)
	} else {
		// Aktualizuj istniejący postęp
		update := bson.M{
			"$set": bson.M{
				"level":                 req.Level,
				"experience":            req.Experience,
				"experienceToNextLevel": req.ExperienceToNextLevel,
				"coins":                 req.Coins,
				"diamonds":              req.Diamonds,
				"premiumBalance":        req.PremiumBalance,
				"betterCoinBalance":     req.BetterCoinBalance,
				"totalCoinsEarned":      req.TotalCoinsEarned,
				"totalDiamondsEarned":   req.TotalDiamondsEarned,
				"totalClicks":           req.TotalClicks,
				"totalUpgrades":         req.TotalUpgrades,
				"upgrades":              req.Upgrades,
				"lastSync":              now,
				"lastDevice":            c.GetHeader("User-Agent"),
				"lastIp":                ip,
				"updatedAt":             now,
			},
		}

		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd aktualizacji postępu: %v", err), ip)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update progress"})
			return
		}
		ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Zaktualizowano postęp dla gracza: %s", playerName), ip)
	}

	// Przygotuj odpowiedź
	response := SyncProgressResponse{
		Success: true,
		Message: "Progress synced successfully",
	}

	// Opcjonalnie zwróć aktualny stan serwera
	if c.Query("includeServerState") == "true" {
		var serverProgress PlayerProgress
		err = collection.FindOne(context.TODO(), filter).Decode(&serverProgress)
		if err == nil {
			response.ServerState = &serverProgress
		}
	}

	ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Sync zakończony sukcesem dla gracza: %s", playerName), ip)
	c.JSON(http.StatusOK, response)
}

// GetProgressHandler pobiera aktualny postęp gracza
func GetProgressHandler(c *gin.Context) {
	ip := c.ClientIP()
	playerName := c.MustGet("playerName").(string)

	ElasticWriter.WriteWithIP(fmt.Sprintf("📊 Get Progress Request → player: %s", playerName), ip)

	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("player_progress")

	// Pobierz postęp gracza
	var progress PlayerProgress
	filter := bson.M{"playerName": playerName}
	err := collection.FindOne(context.TODO(), filter).Decode(&progress)

	if err == mongo.ErrNoDocuments {
		// Gracz nie ma jeszcze postępu - zwróć domyślne wartości
		ElasticWriter.WriteWithIP(fmt.Sprintf("ℹ️ Brak postępu dla gracza: %s, zwracam domyślne wartości", playerName), ip)

		defaultProgress := PlayerProgress{
			PlayerName:            playerName,
			Level:                 1,
			Experience:            0,
			ExperienceToNextLevel: 100,
			Coins:                 0,
			Diamonds:              0,
			PremiumBalance:        0,
			BetterCoinBalance:     0,
			TotalCoinsEarned:      0,
			TotalDiamondsEarned:   0,
			TotalClicks:           0,
			TotalUpgrades:         0,
			Upgrades:              []Upgrade{},
			LastSync:              time.Now(),
			LastDevice:            c.GetHeader("User-Agent"),
			LastIP:                ip,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		response := GetProgressResponse{
			Success:  true,
			Progress: &defaultProgress,
		}

		c.JSON(http.StatusOK, response)
		return
	} else if err != nil {
		// Błąd bazy danych
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd bazy danych: %v", err), ip)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Zwróć znaleziony postęp
	response := GetProgressResponse{
		Success:  true,
		Progress: &progress,
	}

	ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Pobrano postęp dla gracza: %s (level: %d)", playerName, progress.Level), ip)
	c.JSON(http.StatusOK, response)
}

// GetLeaderboardHandler pobiera ranking graczy (opcjonalny endpoint)
func GetLeaderboardHandler(c *gin.Context) {
	ip := c.ClientIP()
	limit := 10 // domyślnie top 10

	ElasticWriter.WriteWithIP(fmt.Sprintf("🏆 Get Leaderboard Request → limit: %d", limit), ip)

	db := c.MustGet("db").(*mongo.Database)
	collection := db.Collection("player_progress")

	// Pobierz top graczy według poziomu, a następnie doświadczenia
	opts := options.Find().SetSort(bson.D{{"level", -1}, {"experience", -1}}).SetLimit(int64(limit))
	cursor, err := collection.Find(context.TODO(), bson.M{}, opts)
	if err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd pobierania rankingu: %v", err), ip)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard"})
		return
	}
	defer cursor.Close(context.TODO())

	var players []PlayerProgress
	if err = cursor.All(context.TODO(), &players); err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd parsowania rankingu: %v", err), ip)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse leaderboard"})
		return
	}

	ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Pobrano ranking: %d graczy", len(players)), ip)
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"leaderboard": players,
	})
}
