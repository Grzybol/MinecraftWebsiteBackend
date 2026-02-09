package main

import (
	"os"
	"strconv"
)

const (
	DefaultMongoURI     = "mongodb://localhost:27017"
	DefaultMariaBase    = "groupmanager:xxxxx@tcp(127.0.0.1:3306)/"
	DatabaseName        = "velocity_login"
	DefaultBackpackSock = "/tmp/equipment_backpack.sock"
	DefaultArmorSock    = "/tmp/equipment_armor.sock"
	DefaultHotbarSock   = "/tmp/equipment_hotbar.sock"
	BackendAuthToken    = "your_default_token_here"
)

// Mapa tokenów autoryzacyjnych dla poszczególnych serwerów
var ServerTokens = map[string]string{
	"boxpvp":   "xxxxx",
	"survival": "xxxxx",
	"skygen":   "xxxxx",
	"global":   "global_token_xyz",
	"questmc":  "xxxxx",
}

func GetMongoURI() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}
	return DefaultMongoURI
}

func GetMariaDBBaseURI() string {
	if uri := os.Getenv("MARIADB_BASE_URI"); uri != "" {
		return uri
	}
	return DefaultMariaBase
}

func GetBackpackSocket() string {
	return DefaultBackpackSock
}

func GetArmorSocket() string {
	return DefaultArmorSock
}

func GetHotbarSocket() string {
	return DefaultHotbarSock
}

func GetBackendAuthToken() string {
	return BackendAuthToken
}

func GetBarcodeCacheTTLHours() int {
	if ttlStr := os.Getenv("BARCODE_CACHE_TTL_HOURS"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil {
			return ttl
		}
	}
	return 24 // domyślnie 24 godziny
}
