package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type ElasticLogWriter struct {
	elastic *ElasticSender
	buffer  []LogEntry
	mu      sync.Mutex
	ticker  *time.Ticker
	stopCh  chan struct{}
}

// NewElasticLogWriter zainicjuje bufor + scheduler
func NewElasticLogWriter(elastic *ElasticSender) *ElasticLogWriter {
	writer := &ElasticLogWriter{
		elastic: elastic,
		buffer:  make([]LogEntry, 0),
		ticker:  time.NewTicker(60 * time.Second),
		stopCh:  make(chan struct{}),
	}

	go writer.startFlushScheduler()
	return writer
}

func (w *ElasticLogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))

	entry := LogEntry{
		Timestamp:     time.Now().Format(time.RFC3339Nano), // pełna precyzja
		Plugin:        "BackendLogger",
		TransactionID: "auto",
		Level:         "INFO",
		Message:       msg,
		PlayerName:    "",
		UUID:          "",
		ServerName:    "backend",
		KeyValue:      0,
		IP:            "", // IP można dodać później, jeśli będzie potrzebne
	}

	w.mu.Lock()
	w.buffer = append(w.buffer, entry)
	w.mu.Unlock()
	// wypisujemy też lokalnie
	return os.Stderr.Write(p)
}
func (w *ElasticLogWriter) WriteWithIP(msg string, ip string) {
	entry := LogEntry{
		Timestamp:     time.Now().Format(time.RFC3339Nano),
		Plugin:        "BackendLogger",
		TransactionID: "auto",
		Level:         "INFO",
		Message:       msg,
		PlayerName:    "",
		UUID:          "",
		ServerName:    "backend",
		KeyValue:      0,
		IP:            ip,
	}

	w.mu.Lock()
	w.buffer = append(w.buffer, entry)
	w.mu.Unlock()

	// log lokalnie (opcjonalnie z IP)
	fmt.Fprintf(os.Stderr, "[%s] %s\n", ip, msg)
}

func (w *ElasticLogWriter) startFlushScheduler() {
	for {
		select {
		case <-w.ticker.C:
			w.flush()
		case <-w.stopCh:
			w.flush()
			return
		}
	}
}

// Flush force – można wywołać np. przy wyłączaniu serwera
func (w *ElasticLogWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buffer) == 0 {
		return
	}

	toSend := make([]LogEntry, len(w.buffer))
	copy(toSend, w.buffer)
	w.buffer = nil // czyścimy bufor

	go func(entries []LogEntry) {
		err := w.elastic.SendLogs(entries)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ElasticLogWriter] ❌ Failed to send logs: %v\n", err)
		}
	}(toSend)
}

func (w *ElasticLogWriter) Stop() {
	close(w.stopCh)
}
