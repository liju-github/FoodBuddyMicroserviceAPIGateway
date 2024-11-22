package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	User "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/User"
	config "github.com/liju-github/FoodBuddyAPIGateway/configs"
)

// Custom claims structure
type Claims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Context keys
const (
	EntityID = "id"
	RoleKey  = "role"
)

// Role constants
const (
	RoleAdmin      = "admin"
	RoleUser       = "user"
	RoleRestaurant = "restaurant"
)

// JWTAuthMiddleware handles JWT authentication and role verification
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Remove Bearer prefix
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token format",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			config := config.LoadConfig()
			return []byte(config.JWTSecretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Verify token expiration
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token has expired",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set(EntityID, claims.ID)
		c.Set(RoleKey, claims.Role)

		// Log the values that were set
		entityID, _ := c.Get(EntityID)
		role, _ := c.Get(RoleKey)
		log.Printf("Context values set - EntityID: %v, Role: %v", entityID, role)

		c.Next()
	}
}

// AdminAuthMiddleware verifies if the user has admin role
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(RoleKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Role information not found",
			})
			c.Abort()
			return
		}

		if role != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RestaurantAuthMiddleware verifies if the user has restaurant role
func RestaurantAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(RoleKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Role information not found",
			})
			c.Abort()
			return
		}

		if role != RoleRestaurant {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Restaurant access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserAuthMiddleware verifies if the user has user role
func UserAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(RoleKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Role information not found",
			})
			c.Abort()
			return
		}

		if role != RoleUser {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "User access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserBanCheckMiddleware checks if a user is banned before allowing access
func UserBanCheckMiddleware(userClient User.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTAuthMiddleware)
		userId, exists := GetEntityID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Check if user is banned
		response, err := userClient.CheckBan(ctx, &User.CheckBanRequest{
			UserId: userId,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check ban status"})
			c.Abort()
			return
		}

		if response.BanStatus {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is banned"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID retrieves the user ID from the context
func GetEntityID(c *gin.Context) (string, bool) {
	ID, exists := c.Get(EntityID)
	if !exists {
		return "", false
	}
	return ID.(string), true
}

// GetUserRole retrieves the user role from the context
func GetEntityRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(RoleKey)
	if !exists {
		return "", false
	}
	return role.(string), true
}
