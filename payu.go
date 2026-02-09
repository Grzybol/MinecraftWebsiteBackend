package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	payuClientID     = "489437"
	payuClientSecret = "50940338572e48472982d52e9d0a8117"

	payuAuthURL  = "https://secure.snd.payu.com/pl/standard/user/oauth/authorize"
	payuOrderURL = "https://secure.snd.payu.com/api/v2_1/orders"
)

var cachedPayuToken string
var cachedTokenExpiry time.Time

type payuAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func getPayuAccessToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", payuClientID)
	data.Set("client_secret", payuClientSecret)

	req, _ := http.NewRequest("POST", payuAuthURL, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Println("❌ Błąd podczas pobierania tokena PayU:", string(body))
		return "", errors.New("invalid response from PayU auth")
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", errors.New("no access_token in response")
	}

	return accessToken, nil
}
