package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipRequestInfo struct {
	LastRequest time.Time
}

var ipRequests = make(map[string]*ipRequestInfo)
var ipRequestsLock = &sync.Mutex{}

const rateLimitWindow = time.Second // 1 sekunda

func RateLimitMiddleware(cooldown time.Duration) gin.HandlerFunc {
	ipRequests := make(map[string]time.Time)
	var mu sync.Mutex

	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.FullPath() // np. "/api/login"
		method := c.Request.Method
		key := fmt.Sprintf("%s:%s", ip, endpoint) // unikalny klucz: IP + endpoint

		now := time.Now()

		mu.Lock()
		last, exists := ipRequests[key]
		if exists && now.Sub(last) < cooldown {
			mu.Unlock()
			log.Printf("⛔ Rate limit - %s %s od IP %s (odstęp %.2fs < %.2fs)",
				method, endpoint, ip,
				now.Sub(last).Seconds(), cooldown.Seconds(),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": fmt.Sprintf("Zbyt wiele zapytań do %s – spróbuj ponownie za %.1f sekund.", endpoint, cooldown.Seconds()),
			})
			c.Abort()
			return
		}
		ipRequests[key] = now
		mu.Unlock()

		log.Printf("✅ Request do %s %s od IP %s", method, endpoint, ip)
		c.Next()
	}
}
