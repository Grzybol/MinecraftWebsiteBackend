package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	PlayerName    string      `bson:"playerName"`
	Products      []TxProduct `bson:"products"`
	Price         float64     `bson:"price"`
	Currency      string      `bson:"currency"`
	Status        string      `bson:"status"`
	TransactionID string      `bson:"transactionId"`
	CallbackToken string      `bson:"callbackToken"`
	CreatedAt     time.Time   `bson:"createdAt"`
}

type TxProduct struct {
	ID       string   `bson:"id"`
	Name     string   `bson:"name"`
	Price    float64  `bson:"price"`
	Currency string   `bson:"currency"`
	Servers  []string `bson:"servers"`
}

// PaymentCode reprezentuje strukturę kodów płatności w MongoDB
type PaymentCode struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Code   string             `bson:"code"`
	Value  int                `bson:"value"`
	Used   bool               `bson:"used"`
	UsedBy string             `bson:"usedBy,omitempty"`
	UsedAt primitive.DateTime `bson:"usedAt,omitempty"`
}

// Struktura uzytkownika - usunieto RegisteredIP
type User struct {
	PlayerName     string `json:"playerName"`
	Password       string `json:"password"`
	WebsiteBalance int    `json:"website_balance,omitempty" bson:"website_balance,omitempty"`
}

// Struktura danych dla odpowiedzi JSON
type UserGroup struct {
	UUID         string `json:"uuid"`
	LastName     string `json:"lastName"`
	PrimaryGroup string `json:"primaryGroup"`
	SubGroups    string `json:"subGroups"`
	Permissions  string `json:"permissions"`
	Info         string `json:"info"`
	TimeStamp    int64  `json:"timestamp"`
}

// ochronka przed brute-force
type LoginAttempt struct {
	Count    int
	LockedAt time.Time
}

// Upgrade reprezentuje pojedynczy upgrade w clickerze
type Upgrade struct {
	ID    string `json:"id" bson:"id"`
	Level int    `json:"level" bson:"level"`
}

// PlayerProgress reprezentuje postęp gracza w clickerze
type PlayerProgress struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty"`
	PlayerName            string             `json:"playerName" bson:"playerName"`
	Level                 int                `json:"level" bson:"level"`
	Experience            float64            `json:"experience" bson:"experience"`
	ExperienceToNextLevel float64            `json:"experienceToNextLevel" bson:"experienceToNextLevel"`
	Coins                 float64            `json:"coins" bson:"coins"`
	Diamonds              float64            `json:"diamonds" bson:"diamonds"`
	PremiumBalance        float64            `json:"premiumBalance" bson:"premiumBalance"`
	BetterCoinBalance     float64            `json:"betterCoinBalance" bson:"betterCoinBalance"`
	TotalCoinsEarned      float64            `json:"totalCoinsEarned" bson:"totalCoinsEarned"`
	TotalDiamondsEarned   float64            `json:"totalDiamondsEarned" bson:"totalDiamondsEarned"`
	TotalClicks           int                `json:"totalClicks" bson:"totalClicks"`
	TotalUpgrades         int                `json:"totalUpgrades" bson:"totalUpgrades"`
	Upgrades              []Upgrade          `json:"upgrades" bson:"upgrades"`
	LastSync              time.Time          `json:"lastSync" bson:"lastSync"`
	LastDevice            string             `json:"lastDevice" bson:"lastDevice"`
	LastIP                string             `json:"lastIp" bson:"lastIp"`
	CreatedAt             time.Time          `bson:"createdAt"`
	UpdatedAt             time.Time          `bson:"updatedAt"`
}

// SyncProgressRequest reprezentuje request do synchronizacji postępu
type SyncProgressRequest struct {
	Level                 int       `json:"level"`
	Experience            float64   `json:"experience"`
	ExperienceToNextLevel float64   `json:"experienceToNextLevel"`
	Coins                 float64   `json:"coins"`
	Diamonds              float64   `json:"diamonds"`
	PremiumBalance        float64   `json:"premiumBalance"`
	BetterCoinBalance     float64   `json:"betterCoinBalance"`
	TotalCoinsEarned      float64   `json:"totalCoinsEarned"`
	TotalDiamondsEarned   float64   `json:"totalDiamondsEarned"`
	TotalClicks           int       `json:"totalClicks"`
	TotalUpgrades         int       `json:"totalUpgrades"`
	Upgrades              []Upgrade `json:"upgrades"`
	Timestamp             time.Time `json:"timestamp"`
	// Dodatkowe pola dla weryfikacji
	PreviousCoins      float64   `json:"previousCoins,omitempty"`      // Stan przed synchronizacją
	PreviousExperience float64   `json:"previousExperience,omitempty"` // Exp przed synchronizacją
	PreviousDiamonds   float64   `json:"previousDiamonds,omitempty"`   // Diamenty przed synchronizacją
	PreviousLevel      int       `json:"previousLevel,omitempty"`      // Level przed synchronizacją
	PreviousTimestamp  time.Time `json:"previousTimestamp,omitempty"`  // Timestamp przed synchronizacją
}

// SyncProgressResponse reprezentuje odpowiedź z synchronizacji
type SyncProgressResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message"`
	ServerState *PlayerProgress `json:"serverState,omitempty"`
}

// GetProgressResponse reprezentuje odpowiedź z pobierania postępu
type GetProgressResponse struct {
	Success  bool            `json:"success"`
	Progress *PlayerProgress `json:"progress"`
}

// ProgressVerification reprezentuje dane do weryfikacji postępu
type ProgressVerification struct {
	ClickDamage     float64 `json:"clickDamage"`     // Damage z jednego kliknięcia
	DPS             float64 `json:"dps"`             // Damage per second z auto-mining
	MaxClicksPerSec int     `json:"maxClicksPerSec"` // Maksymalne kliknięcia na sekundę (25)
	ClickExp        float64 `json:"clickExp"`        // Exp z jednego kliknięcia
	DPSExp          float64 `json:"dpsExp"`          // Exp per second z auto-mining
	DiamondDropRate float64 `json:"diamondDropRate"` // Szansa na diament z kliknięcia (0.0-1.0)
	GoldBonus       float64 `json:"goldBonus"`       // Bonus do złota (mnożnik)
	ExpBonus        float64 `json:"expBonus"`        // Bonus do exp (mnożnik)
	Margin          float64 `json:"margin"`          // Margines bezpieczeństwa (np. 0.05 = 5%)
}

// VerificationResult reprezentuje wynik weryfikacji
type VerificationResult struct {
	IsValid             bool    `json:"isValid"`
	Reason              string  `json:"reason,omitempty"`
	MaxPossibleCoins    float64 `json:"maxPossibleCoins"`
	MaxPossibleExp      float64 `json:"maxPossibleExp"`
	MaxPossibleDiamonds float64 `json:"maxPossibleDiamonds"`
	ActualCoinsGain     float64 `json:"actualCoinsGain"`
	ActualExpGain       float64 `json:"actualExpGain"`
	ActualDiamondsGain  float64 `json:"actualDiamondsGain"`
	DeltaTime           float64 `json:"deltaTime"`
}

type BarcodeInfo struct {
	Barcode   string    `bson:"barcode"`
	Name      string    `bson:"name"`
	Brand     string    `bson:"brand"`
	Country   string    `bson:"country"`
	IsFromUSA bool      `bson:"isFromUSA"`
	CreatedAt time.Time `bson:"createdAt"`
}

type WorldStat struct {
	PlayerUUID string `json:"player_uuid"`
	PlayerName string `json:"player_name"`
	World      string `json:"world"`
	Statistic  string `json:"statistic"`
	Value      int64  `json:"value"`
}
