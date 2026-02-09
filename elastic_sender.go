package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type LogEntry struct {
	Timestamp     string  `json:"timestamp"`
	Plugin        string  `json:"plugin"`
	TransactionID string  `json:"transactionID"`
	Level         string  `json:"level"`
	Message       string  `json:"message"`
	PlayerName    string  `json:"playerName"`
	UUID          string  `json:"uuid"`
	ServerName    string  `json:"serverName"`
	KeyValue      float64 `json:"keyValue"`
	IP            string  `json:"IP"`
}

type ElasticSender struct {
	URL        string
	APIKey     string
	Index      string
	Client     *http.Client
	VerifyCert bool
}

func NewElasticSender(cfg ElasticConfig) *ElasticSender {
	transport := &http.Transport{}

	if !cfg.VerifyCert {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		// Możesz dodać obsługę CA cert tu, jak będziesz chciał
		caCert, err := os.ReadFile("your_ca_cert.pem")
		if err != nil {
			log.Fatal("❌ Error reading CA cert:", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		transport.TLSClientConfig = &tls.Config{RootCAs: caCertPool}
	}

	client := &http.Client{Transport: transport}

	return &ElasticSender{
		URL:        cfg.URL,
		APIKey:     cfg.APIKey,
		Index:      cfg.Index,
		Client:     client,
		VerifyCert: cfg.VerifyCert,
	}
}

func (s *ElasticSender) SendLogs(logs []LogEntry) error {
	var buf strings.Builder

	/*
		for _, logEntry := range logs {
			meta := `{"index":{}}` + "\n"
			doc := fmt.Sprintf(`{"timestamp":"%s","plugin":"ElasticBuffer","transactionID":"%s","level":"%s","message":"%s","playerName":"%s","uuid":"%s","serverName":"%s","keyValue":%.5f,"IP":"%s"}`,
				logEntry.Timestamp,
				logEntry.TransactionID,
				logEntry.Level,
				sanitize(logEntry.Message),
				logEntry.PlayerName,
				logEntry.UUID,
				logEntry.ServerName,
				logEntry.KeyValue,
				logEntry.IP,
			)
			buf.WriteString(meta)
			buf.WriteString(doc + "\n")
		}
	*/
	for _, logEntry := range logs {
		meta := `{"index":{}}` + "\n"

		jsonDoc, err := json.Marshal(logEntry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to marshal log entry: %v\n", err)
			continue
		}

		buf.WriteString(meta)
		buf.Write(jsonDoc)
		buf.WriteByte('\n')
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/_bulk", s.URL, s.Index), bytes.NewBufferString(buf.String()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-ndjson")
	if s.APIKey != "" {
		req.Header.Set("Authorization", "ApiKey "+s.APIKey)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// Używamy fmt zamiast log
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("[ElasticSender] Success (%d): %s\n", resp.StatusCode, string(bodyBytes))
		return nil
	} else {
		fmt.Printf("[ElasticSender] Failed (%d): %s\n", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("failed to send logs, status: %d", resp.StatusCode)
	}
}

func sanitize(input string) string {
	return strings.ReplaceAll(input, `"`, "")
}
