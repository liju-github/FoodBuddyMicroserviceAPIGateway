package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"io"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	config "github.com/liju-github/FoodBuddyAPIGateway/configs"
	"github.com/liju-github/FoodBuddyAPIGateway/middleware"
	"github.com/liju-github/FoodBuddyAPIGateway/model"
	"github.com/sirupsen/logrus"
)

type RestaurantController struct {
	restaurantClient restaurantPb.RestaurantServiceClient
	validator        *validator.Validate
	logger           *logrus.Logger
	jwtSecret        []byte
}

// Custom validation rules
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	passwordRegex = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*]{8,}$`)
	nameRegex     = regexp.MustCompile(`^[a-zA-Z\s]{2,50}$`)
	phoneRegex    = regexp.MustCompile(`^\d{10}$`)
	pincodeRegex  = regexp.MustCompile(`^\d{6}$`)
)

// Validation functions
func (rc *RestaurantController) validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func (rc *RestaurantController) validatePassword(password string) bool {
	return passwordRegex.MatchString(password)
}

func (rc *RestaurantController) validateName(name string) bool {
	return nameRegex.MatchString(name)
}

func (rc *RestaurantController) validatePhone(phone uint64) bool {
	return phoneRegex.MatchString(fmt.Sprint(phone))
}

func (rc *RestaurantController) validatePincode(pincode string) bool {
	return pincodeRegex.MatchString(pincode)
}

func (rc *RestaurantController) validateAddress(address model.Address) error {
	if strings.TrimSpace(address.StreetName) == "" {
		return fmt.Errorf("street name cannot be empty")
	}
	if strings.TrimSpace(address.Locality) == "" {
		return fmt.Errorf("locality cannot be empty")
	}
	if strings.TrimSpace(address.State) == "" {
		return fmt.Errorf("state cannot be empty")
	}
	if !rc.validatePincode(address.Pincode) {
		return fmt.Errorf("invalid pincode format")
	}
	return nil
}

func (rc *RestaurantController) validateRestaurantInput(request model.RestaurantSignupRequest) error {
	if !rc.validateEmail(request.OwnerEmail) {
		return fmt.Errorf("invalid email format")
	}

	if !rc.validatePassword(request.Password) {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if !rc.validateName(request.RestaurantName) {
		return fmt.Errorf("invalid restaurant name format")
	}

	if !rc.validatePhone(request.PhoneNumber) {
		return fmt.Errorf("invalid phone number format")
	}

	if err := rc.validateAddress(request.Address); err != nil {
		return fmt.Errorf("invalid address: %v", err)
	}

	return nil
}

func NewRestaurantController(restaurantClient restaurantPb.RestaurantServiceClient) *RestaurantController {
	validate := validator.New()
	logger := logrus.New()

	// Configure JSON formatter with custom fields
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
		PrettyPrint: false,
	})

	// Set log level
	logger.SetLevel(logrus.InfoLevel)

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
	}

	// Open log file with date in filename
	currentTime := time.Now()
	logFileName := fmt.Sprintf("logs/api_%s.log", currentTime.Format("2006-01-02"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	} else {
		// Use both file and stdout for logging
		logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	// Add default fields to all log entries
	logger = logger.WithFields(logrus.Fields{
		"service": "api_gateway",
		"version": "1.0",
		"env":     config.LoadConfig().Environment,
	}).Logger

	// Get JWT secret from environment variable
	jwtSecret := []byte(config.LoadConfig().JWTSecretKey)

	return &RestaurantController{
		restaurantClient: restaurantClient,
		validator:        validate,
		logger:           logger,
		jwtSecret:        jwtSecret,
	}
}

func (rc *RestaurantController) generateToken(ID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      ID,
		"role":    middleware.RoleRestaurant,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"created": time.Now().Unix(),
	})

	return token.SignedString(rc.jwtSecret)
}

// RestaurantSignup handles restaurant registration
func (rc *RestaurantController) RestaurantSignup(ctx *gin.Context) {
	var request model.RestaurantSignupRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		rc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/auth/restaurant/signup",
		}).Error("Failed to bind signup request")
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidRequestFormat, err))
		return
	}

	// Log sanitized request (excluding password)
	rc.logger.WithFields(logrus.Fields{
		"restaurantName": request.RestaurantName,
		"ownerEmail":     request.OwnerEmail,
		"path":           "/auth/restaurant/signup",
	}).Info("Processing signup request")

	// Validate input
	if err := rc.validateRestaurantInput(request); err != nil {
		rc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/auth/restaurant/signup",
		}).Warn("Validation failed")
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error(), nil))
		return
	}

	// Convert to protobuf request
	pbRequest := &restaurantPb.RestaurantSignupRequest{
		RestaurantName: request.RestaurantName,
		OwnerEmail:     request.OwnerEmail,
		Password:       request.Password,
		PhoneNumber:    request.PhoneNumber,
		Address: &restaurantPb.Address{
			StreetName: request.Address.StreetName,
			Locality:   request.Address.Locality,
			State:      request.Address.State,
			Pincode:    request.Address.Pincode,
		},
	}

	response, err := rc.restaurantClient.RestaurantSignup(context.Background(), pbRequest)
	if err != nil {
		rc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/auth/restaurant/signup",
		}).Error("Signup failed")
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrSignupFailed, err))
		return
	}

	// Generate JWT token
	token, err := rc.generateToken(response.RestaurantId)
	if err != nil {
		rc.logger.WithFields(logrus.Fields{
			"restaurantId": response.RestaurantId,
			"error":        err.Error(),
		}).Error(model.ErrFailedGenerateToken)
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedGenerateToken, err))
		return
	}

	response.Token = token

	rc.logger.WithFields(logrus.Fields{
		"restaurantId":   response.RestaurantId,
		"restaurantName": request.RestaurantName,
	}).Info("Signup successful")

	ctx.JSON(http.StatusOK, model.SuccessResponse("Restaurant registered successfully", response))
}

// RestaurantLogin handles restaurant authentication
func (rc *RestaurantController) RestaurantLogin(ctx *gin.Context) {
	var request model.RestaurantLoginRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		rc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/auth/restaurant/login",
		}).Error("Failed to bind login request")
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidRequestFormat, err))
		return
	}

	// Log sanitized request (excluding password)
	rc.logger.WithFields(logrus.Fields{
		"ownerEmail": request.OwnerEmail,
		"path":       "/auth/restaurant/login",
	}).Info("Processing login request")

	// Validate input
	if !rc.validateEmail(request.OwnerEmail) {
		rc.logger.WithFields(logrus.Fields{
			"email": request.OwnerEmail,
			"path":  "/auth/restaurant/login",
		}).Warn("Invalid email format")
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidEmailFormat, nil))
		return
	}

	if !rc.validatePassword(request.Password) {
		rc.logger.WithField("email", request.OwnerEmail).Warn("Invalid password format")
		ctx.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrPasswordTooShort, nil))
		return
	}

	// Convert to protobuf request
	pbRequest := &restaurantPb.RestaurantLoginRequest{
		OwnerEmail: request.OwnerEmail,
		Password:   request.Password,
	}

	response, err := rc.restaurantClient.RestaurantLogin(context.Background(), pbRequest)
	if err != nil {
		rc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/auth/restaurant/login",
		}).Error("Login failed")
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrLoginFailed, err))
		return
	}

	// Generate JWT token
	token, err := rc.generateToken(response.RestaurantId)
	if err != nil {
		rc.logger.WithFields(logrus.Fields{
			"restaurantId": response.RestaurantId,
			"error":        err.Error(),
		}).Error(model.ErrFailedGenerateToken)
		ctx.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedGenerateToken, err))
		return
	}

	response.Token = token

	rc.logger.WithFields(logrus.Fields{
		"restaurantId": response.RestaurantId,
		"ownerEmail":   request.OwnerEmail,
	}).Info("Login successful")

	ctx.JSON(http.StatusOK, model.SuccessResponse("Login successful", response))
}

func (rc *RestaurantController) EditRestaurant(c *gin.Context) {
	// Get restaurant ID from JWT token
	restaurantID, exists := middleware.GetEntityID(c)
	if !exists {
		rc.logger.Error("Restaurant ID not found in token")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse("Restaurant ID not found in token", nil))
		return
	}

	var request restaurantPb.EditRestaurantRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		rc.logger.WithError(err).Error("Failed to bind edit restaurant request")
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid request format", err))
		return
	}

	// Set the restaurant ID from token
	request.RestaurantId = restaurantID

	// Validate input
	if !rc.validateName(request.RestaurantName) {
		rc.logger.Error("Invalid restaurant name format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid restaurant name format", nil))
		return
	}

	if !rc.validatePhone(request.PhoneNumber) {
		rc.logger.Error("Invalid phone number format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid phone number format", nil))
		return
	}

	if err := rc.validateAddress(model.Address{
		StreetName: request.Address.StreetName,
		Locality:   request.Address.Locality,
		State:      request.Address.State,
		Pincode:    request.Address.Pincode,
	}); err != nil {
		rc.logger.WithError(err).Error("Invalid address")
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid address", err))
		return
	}
	response, err := rc.restaurantClient.EditRestaurant(context.Background(), &request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to edit restaurant")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to edit restaurant", err))
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse("Restaurant updated successfully", response))
}

func (rc *RestaurantController) GetRestaurantProductsByID(c *gin.Context) {
	restaurantID := c.Query("restaurantId")
	request := &restaurantPb.GetRestaurantProductsByIDRequest{
		RestaurantId: restaurantID,
	}

	response, err := rc.restaurantClient.GetRestaurantProductsByID(context.Background(), request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to get restaurant products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) GetAllRestaurantWithProducts(c *gin.Context) {
	request := &restaurantPb.GetAllRestaurantAndProductsRequest{}

	response, err := rc.restaurantClient.GetAllRestaurantWithProducts(context.Background(), request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to get all restaurants with products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) AddProduct(c *gin.Context) {
	var request restaurantPb.AddProductRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		rc.logger.WithError(err).Error("Failed to bind add product request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get restaurant ID from token
	restaurantID, exists := middleware.GetEntityID(c)
	if !exists {
		rc.logger.Error("Restaurant ID not found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	request.RestaurantId = restaurantID

	response, err := rc.restaurantClient.AddProduct(context.Background(), &request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to add product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) EditProduct(c *gin.Context) {
	var request restaurantPb.EditProductRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		rc.logger.WithError(err).Error("Failed to bind edit product request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get restaurant ID and role from token
	role, exists := middleware.GetEntityRole(c)
	if !exists {
		rc.logger.Error("Role not found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// If not admin, verify restaurant ownership
	if role != middleware.RoleAdmin {
		restaurantID, exists := middleware.GetEntityID(c)
		if !exists {
			rc.logger.Error("Restaurant ID not found in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get restaurant ID for the product
		productRestaurantResp, err := rc.restaurantClient.GetRestaurantIDviaProductID(context.Background(), &restaurantPb.GetRestaurantIDviaProductIDRequest{
			ProductId: request.ProductId,
		})
		if err != nil {
			rc.logger.WithError(err).Error("Failed to get restaurant ID for product")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership
		if productRestaurantResp.RestaurantId != restaurantID {
			rc.logger.Error("Restaurant not authorized to edit this product")
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to edit this product"})
			return
		}

		request.RestaurantId = restaurantID
	}

	// Validate product details
	if strings.TrimSpace(request.ProductId) == "" {
		rc.logger.Error("Product ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	if strings.TrimSpace(request.Name) == "" {
		rc.logger.Error("Product name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product name is required"})
		return
	}

	if request.Price <= 0 {
		rc.logger.Error("Invalid product price")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than 0"})
		return
	}

	if request.Stock < 0 {
		rc.logger.Error("Invalid product stock")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stock cannot be negative"})
		return
	}

	response, err := rc.restaurantClient.EditProduct(context.Background(), &request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to edit product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) DeleteProductByID(c *gin.Context) {
	var request restaurantPb.DeleteProductByIDRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		rc.logger.WithError(err).Error("Failed to bind delete product request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get role from token
	role, exists := middleware.GetEntityRole(c)
	if !exists {
		rc.logger.Error("Role not found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// If not admin, verify restaurant ownership
	if role != middleware.RoleAdmin {
		restaurantID, exists := middleware.GetEntityID(c)
		if !exists {
			rc.logger.Error("Restaurant ID not found in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get restaurant ID for the product
		productRestaurantResp, err := rc.restaurantClient.GetRestaurantIDviaProductID(context.Background(), &restaurantPb.GetRestaurantIDviaProductIDRequest{
			ProductId: request.ProductId,
		})
		if err != nil {
			rc.logger.WithError(err).Error("Failed to get restaurant ID for product")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership
		if productRestaurantResp.RestaurantId != restaurantID {
			rc.logger.Error("Restaurant not authorized to delete this product")
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this product"})
			return
		}

		request.RestaurantId = restaurantID
	}

	if strings.TrimSpace(request.ProductId) == "" {
		rc.logger.Error("Product ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	response, err := rc.restaurantClient.DeleteProductByID(context.Background(), &request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to delete product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) GetProductByID(c *gin.Context) {
	productID := c.Query("productId")
	request := &restaurantPb.GetProductByIDRequest{
		ProductId: productID,
	}

	response, err := rc.restaurantClient.GetProductByID(context.Background(), request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to get product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) IncrementProductStock(c *gin.Context) {
	var request restaurantPb.IncremenentProductStockByValueRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		rc.logger.WithError(err).Error("Failed to bind increment stock request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get role from token
	role, exists := middleware.GetEntityRole(c)
	if !exists {
		rc.logger.Error("Role not found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// If not admin, verify restaurant ownership
	if role != middleware.RoleAdmin {
		restaurantID, exists := middleware.GetEntityID(c)
		if !exists {
			rc.logger.Error("Restaurant ID not found in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get restaurant ID for the product
		productRestaurantResp, err := rc.restaurantClient.GetRestaurantIDviaProductID(context.Background(), &restaurantPb.GetRestaurantIDviaProductIDRequest{
			ProductId: request.ProductId,
		})
		if err != nil {
			rc.logger.WithError(err).Error("Failed to get restaurant ID for product")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership
		if productRestaurantResp.RestaurantId != restaurantID {
			rc.logger.Error("Restaurant not authorized to modify this product's stock")
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this product's stock"})
			return
		}
	}

	if strings.TrimSpace(request.ProductId) == "" {
		rc.logger.Error("Product ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	if request.Value <= 0 {
		rc.logger.Error("Invalid increment value")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Increment value must be greater than 0"})
		return
	}

	response, err := rc.restaurantClient.IncremenentProductStockByValue(context.Background(), &request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to increment stock")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) DecrementProductStock(c *gin.Context) {
	var request restaurantPb.DecrementProductStockByValueByValueRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		rc.logger.WithError(err).Error("Failed to bind decrement stock request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get role from token
	role, exists := middleware.GetEntityRole(c)
	if !exists {
		rc.logger.Error("Role not found in token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// If not admin, verify restaurant ownership
	if role != middleware.RoleAdmin {
		restaurantID, exists := middleware.GetEntityID(c)
		if !exists {
			rc.logger.Error("Restaurant ID not found in token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get restaurant ID for the product
		productRestaurantResp, err := rc.restaurantClient.GetRestaurantIDviaProductID(context.Background(), &restaurantPb.GetRestaurantIDviaProductIDRequest{
			ProductId: request.ProductId,
		})
		if err != nil {
			rc.logger.WithError(err).Error("Failed to get restaurant ID for product")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership
		if productRestaurantResp.RestaurantId != restaurantID {
			rc.logger.Error("Restaurant not authorized to modify this product's stock")
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this product's stock"})
			return
		}
	}

	if strings.TrimSpace(request.ProductId) == "" {
		rc.logger.Error("Product ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	if request.Value <= 0 {
		rc.logger.Error("Invalid decrement value")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Decrement value must be greater than 0"})
		return
	}

	response, err := rc.restaurantClient.DecrementProductStockByValue(context.Background(), &request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to decrement stock")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) BanRestaurant(c *gin.Context) {
	var request restaurantPb.BanRestaurantRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		rc.logger.WithError(err).Error("Failed to bind ban restaurant request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := rc.restaurantClient.BanRestaurant(context.Background(), &request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to ban restaurant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) UnbanRestaurant(c *gin.Context) {
	var request restaurantPb.UnbanRestaurantRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		rc.logger.WithError(err).Error("Failed to bind unban restaurant request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := rc.restaurantClient.UnbanRestaurant(context.Background(), &request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to unban restaurant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) GetRestaurantIDviaProductID(c *gin.Context) {
	productID := c.Query("productId")
	request := &restaurantPb.GetRestaurantIDviaProductIDRequest{
		ProductId: productID,
	}

	response, err := rc.restaurantClient.GetRestaurantIDviaProductID(context.Background(), request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to get restaurant ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (rc *RestaurantController) GetStockByProductID(c *gin.Context) {
	productID := c.Query("productId")
	request := &restaurantPb.GetStockByProductIDRequest{
		ProductId: productID,
	}

	response, err := rc.restaurantClient.GetStockByProductID(context.Background(), request)
	if err != nil {
		rc.logger.WithError(err).Error("Failed to get stock")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
