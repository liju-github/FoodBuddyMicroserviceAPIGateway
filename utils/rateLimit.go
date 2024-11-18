package utils

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware creates a rate limiter with a max of 3 requests per IP per minute.
func RateLimitMiddleware() gin.HandlerFunc {
	const apiRate = 3
	const resetInterval = time.Minute
	const ttl = 3 * time.Minute // IPs inactive for longer than ttl are removed

	type Visitor struct {
		requests int
		lastSeen time.Time
	}

	var (
		mutex    sync.Mutex
		visitors = make(map[string]*Visitor)
	)

	// Background cleanup for stale visitors
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mutex.Lock()
			for ip, visitor := range visitors {
				if time.Since(visitor.lastSeen) > ttl {
					delete(visitors, ip)
				}
			}
			mutex.Unlock()
		}
	}()

	return func(c *gin.Context) {
		visitorIP := c.ClientIP()

		// Check and update visitor data
		mutex.Lock()
		visitorData, exists := visitors[visitorIP]
		if !exists {
			visitorData = &Visitor{
				requests: 1,
				lastSeen: time.Now(),
			}
			visitors[visitorIP] = visitorData
		} else {
			visitorData.requests++
			visitorData.lastSeen = time.Now()
		}
		requests := visitorData.requests
		mutex.Unlock()

		// If rate limit exceeded, return 429 response
		if requests > apiRate {
			message := fmt.Sprintf("rate limit exceeded for IP: %v", visitorIP)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":     false,
				"message":    message,
				"error_code": http.StatusTooManyRequests,
			})
			return
		}

		c.Next()

		// Reset visitor requests every minute
		go func() {
			time.Sleep(resetInterval)
			mutex.Lock()
			if visitor, ok := visitors[visitorIP]; ok && visitor.requests > 0 {
				visitor.requests = 0
			}
			mutex.Unlock()
		}()
	}
}
