package utils

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetJWTClaim extracts the email and role claims from the JWT token in the Authorization cookie.
func GetJWTClaim(c *gin.Context) (email string, role string, err error) {
	// Retrieve the JWT token from the cookie
	JWTToken, err := c.Cookie("Authorization")
	if err != nil || JWTToken == "" {
		return "", "", errors.New("authorization token not found")
	}

	// Load the secret key
	hmacSecretString := LoadSecrets().JWTSecret
	hmacSecret := []byte(hmacSecretString)

	// Parse and validate the token
	token, err := jwt.Parse(JWTToken, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return hmacSecret, nil
	})
	if err != nil {
		return "", "", errors.New("invalid or malformed token")
	}

	// Extract and validate claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid token claims")
	}

	// Validate expiration
	expirationTime, ok := claims["exp"].(float64)
	if !ok {
		return "", "", errors.New("missing or invalid expiration claim")
	}
	if time.Now().Unix() > int64(expirationTime) {
		return "", "", errors.New("token has expired")
	}

	// Extract email and role claims
	email, ok = claims["email"].(string)
	if !ok || email == "" {
		return "", "", errors.New("missing or invalid email claim")
	}
	role, ok = claims["role"].(string)
	if !ok || role == "" {
		return "", "", errors.New("missing or invalid role claim")
	}

	return email, role, nil
}

