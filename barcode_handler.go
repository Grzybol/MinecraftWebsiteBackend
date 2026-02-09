package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func BarcodeInfoHandler(c *gin.Context) {
	ip := c.ClientIP()

	barcode := c.Query("barcode")
	if barcode == "" {
		ElasticWriter.WriteWithIP("❌ Brak parametru barcode w zapytaniu.", ip)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing barcode parameter"})
		return
	}

	ElasticWriter.WriteWithIP(fmt.Sprintf("🔍 Szukam produktu dla barcode: %s", barcode), ip)

	var cached BarcodeInfo
	err := barcodeCacheCollection.FindOne(context.TODO(), bson.M{"barcode": barcode}).Decode(&cached)
	cacheTTL := time.Duration(GetBarcodeCacheTTLHours()) * time.Hour

	useCache := false
	if err == nil {
		// Sprawdzamy czy CreatedAt + TTL > teraz
		if cached.CreatedAt.Add(cacheTTL).After(time.Now()) {
			useCache = true
			ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Produkt znaleziony w cache (świeży): %s (%s)", cached.Name, barcode), ip)
		} else {
			ElasticWriter.WriteWithIP(fmt.Sprintf("ℹ️ Produkt znaleziony w cache (przeterminowany), odświeżam... %s", barcode), ip)
		}
	}

	if useCache {
		c.JSON(http.StatusOK, cached)
		return
	}

	// Pobieramy z API
	product, err := fetchBarcodeFromAPI(barcode)
	if err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("❌ Błąd pobierania z API dla barcode %s: %v", barcode, err), ip)
		// Można też ewentualnie zwrócić stare dane jako fallback (opcja do rozważenia)
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Zapisujemy do cache
	_, err = barcodeCacheCollection.UpdateOne(
		context.TODO(),
		bson.M{"barcode": barcode},
		bson.M{"$set": product},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		ElasticWriter.WriteWithIP(fmt.Sprintf("⚠️ Nie udało się zapisać produktu do cache (barcode: %s): %v", barcode, err), ip)
	} else {
		ElasticWriter.WriteWithIP(fmt.Sprintf("✅ Produkt zapisany do cache: %s (%s)", product.Name, barcode), ip)
	}

	c.JSON(http.StatusOK, product)
}

func fetchBarcodeFromAPI(barcode string) (*BarcodeInfo, error) {
	url := fmt.Sprintf("https://world.openfoodfacts.org/api/v0/product/%s.json", barcode)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed")
	}
	defer resp.Body.Close()

	var apiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("Failed to parse API response")
	}

	status, ok := apiResp["status"].(float64)
	if !ok || int(status) != 1 {
		return nil, fmt.Errorf("Product not found in API")
	}

	product, _ := apiResp["product"].(map[string]interface{})

	name := safeString(product["product_name"])
	brand := safeString(product["brands"])
	country := safeString(product["countries"])
	isFromUSA := strings.Contains(strings.ToLower(country), "united states")

	return &BarcodeInfo{
		Barcode:   barcode,
		Name:      name,
		Brand:     brand,
		Country:   country,
		IsFromUSA: isFromUSA,
		CreatedAt: time.Now(),
	}, nil
}

func safeString(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}
