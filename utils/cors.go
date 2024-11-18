package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsMiddleware sets the necessary headers to support Cross-Origin Resource Sharing (CORS).
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Define allowed origins and headers
		allowedOrigins := "*"
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		allowedHeaders := []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}

		// Set headers
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204) // No Content status code
			return
		}

		// Continue to the next middleware
		c.Next()
	}
}
