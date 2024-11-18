package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort           string
	JWTSecretKey       string
	UserGRPCPort       string
	ContentGRPCPort    string
	AdminGRPCPort      string
	NotificationGRPCPort string
}

func LoadConfig() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return Config{
		HTTPPort:           os.Getenv("HTTP_PORT"),
		JWTSecretKey:       os.Getenv("JWT_SECRET"),
		UserGRPCPort:       os.Getenv("USER_GRPC_PORT"),
		ContentGRPCPort:    os.Getenv("CONTENT_GRPC_PORT"),
		AdminGRPCPort:      os.Getenv("ADMIN_GRPC_PORT"),
		NotificationGRPCPort: os.Getenv("NOTIFICATION_GRPC_PORT"),
	}
}