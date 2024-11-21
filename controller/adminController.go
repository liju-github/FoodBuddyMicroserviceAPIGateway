package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	adminPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Admin"
	config "github.com/liju-github/FoodBuddyAPIGateway/configs"
	"github.com/liju-github/FoodBuddyAPIGateway/middleware"
)

type AdminController struct {
	adminClient adminPb.AdminServiceClient
	jwtSecret   []byte
}

func NewAdminController(adminClient adminPb.AdminServiceClient) *AdminController {
	return &AdminController{
		adminClient: adminClient,
		jwtSecret:   []byte(config.LoadConfig().JWTSecretKey),
	}
}

func (ac *AdminController) AdminLogin(ctx *gin.Context) {
	var AdminLoginRequest adminPb.AdminLoginRequest
	if err := ctx.ShouldBindJSON(&AdminLoginRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := ac.adminClient.AdminLogin(context.Background(), &AdminLoginRequest)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response.Token, _ = ac.generateToken("admin")

	ctx.JSON(http.StatusOK, response)
}

func (ac *AdminController) generateToken(ID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      ID,
		"role":    middleware.RoleAdmin,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"created": time.Now().Unix(),
	})

	return token.SignedString(ac.jwtSecret)
}
