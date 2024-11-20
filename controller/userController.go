package controller

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	User "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/User"
	config "github.com/liju-github/FoodBuddyAPIGateway/configs"
	"github.com/liju-github/FoodBuddyAPIGateway/middleware"
	"github.com/liju-github/FoodBuddyAPIGateway/model"
	"github.com/sirupsen/logrus"
)

type UserController struct {
	userClient User.UserServiceClient
	validator  *validator.Validate
	logger     *logrus.Logger
	jwtSecret  []byte
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
func (uc *UserController) validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func (uc *UserController) validatePassword(password string) bool {
	return passwordRegex.MatchString(password)
}

func (uc *UserController) validateName(name string) bool {
	return nameRegex.MatchString(name)
}

func (uc *UserController) validatePhone(phone uint64) bool {
	return phoneRegex.MatchString(fmt.Sprint(phone))
}

func (uc *UserController) validatePincode(pincode string) bool {
	return pincodeRegex.MatchString(pincode)
}

func (uc *UserController) validateAddress(address model.Address) error {
	if strings.TrimSpace(address.StreetName) == "" {
		return fmt.Errorf("street name cannot be empty")
	}
	if strings.TrimSpace(address.Locality) == "" {
		return fmt.Errorf("locality cannot be empty")
	}
	if strings.TrimSpace(address.State) == "" {
		return fmt.Errorf("state cannot be empty")
	}
	if !uc.validatePincode(address.Pincode) {
		return fmt.Errorf("invalid pincode format")
	}
	return nil
}

func NewUserController(userClient User.UserServiceClient) *UserController {
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

	// Get JWT secret from environment variable or use a default for development
	jwtSecret := []byte(config.LoadConfig().JWTSecretKey)

	return &UserController{
		userClient: userClient,
		validator:  validate,
		logger:     logger,
		jwtSecret:  jwtSecret,
	}
}

func (uc *UserController) generateToken(ID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      ID,
		"role":    middleware.RoleUser,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"created": time.Now().Unix(),
	})

	return token.SignedString(uc.jwtSecret)
}

// Login handles user authentication
func (uc *UserController) Login(c *gin.Context) {
	var request model.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/auth/user/login",
		}).Error("Failed to bind login request")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidRequestFormat, err))
		return
	}

	// Log sanitized request (excluding password)
	uc.logger.WithFields(logrus.Fields{
		"email": request.Email,
		"path":  "/auth/user/login",
	}).Info("Processing login request")

	// Additional validation
	if !uc.validateEmail(request.Email) {
		uc.logger.WithFields(logrus.Fields{
			"email": request.Email,
			"path":  "/auth/user/login",
		}).Warn("Invalid email format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidEmailFormat, nil))
		return
	}

	if !uc.validatePassword(request.Password) {
		uc.logger.WithField("email", request.Email).Warn("Invalid password format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrPasswordTooShort, nil))
		return
	}

	resp, err := uc.userClient.UserLogin(context.Background(), &User.UserLoginRequest{
		Email:    request.Email,
		Password: request.Password,
	})

	resp.Token, err = uc.generateToken(resp.UserId)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"email": request.Email,
			"error": err.Error(),
		}).Error(model.ErrFailedGenerateToken)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedGenerateToken, err))
		return
	}

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"email": request.Email,
			"error": err.Error(),
		}).Error("Login failed")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrLoginFailed, err))
		return
	}

	uc.logger.WithFields(logrus.Fields{
		"email":  request.Email,
		"userId": resp.UserId,
	}).Info("Login successful")

	c.JSON(http.StatusOK, model.SuccessResponse("Login successful", resp))
}

// Signup handles user registration
func (uc *UserController) Signup(c *gin.Context) {
	var request model.SignupRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/auth/user/signup",
		}).Error("Failed to bind signup request")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidRequestFormat, err))
		return
	}

	// Log sanitized request (excluding password)
	uc.logger.WithFields(logrus.Fields{
		"email":       request.Email,
		"firstName":   request.FirstName,
		"lastName":    request.LastName,
		"phoneNumber": request.PhoneNumber,
		"address": logrus.Fields{
			"streetName": request.Address.StreetName,
			"locality":   request.Address.Locality,
			"state":      request.Address.State,
			"pincode":    request.Address.Pincode,
		},
		"path": "/auth/user/signup",
	}).Info("Processing signup request")

	// Validate all fields
	if !uc.validateEmail(request.Email) {
		uc.logger.WithField("email", request.Email).Warn("Invalid email format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidEmailFormat, nil))
		return
	}

	if !uc.validatePassword(request.Password) {
		uc.logger.WithField("email", request.Email).Warn("Invalid password format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrPasswordTooShort, nil))
		return
	}

	if !uc.validateName(request.FirstName) || !uc.validateName(request.LastName) {
		uc.logger.WithFields(logrus.Fields{
			"email":     request.Email,
			"firstName": request.FirstName,
			"lastName":  request.LastName,
		}).Warn("Invalid name format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid name format", nil))
		return
	}

	if !uc.validatePhone(request.PhoneNumber) {
		uc.logger.WithFields(logrus.Fields{
			"email":       request.Email,
			"phoneNumber": request.PhoneNumber,
		}).Warn("Invalid phone number format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid phone number format", nil))
		return
	}

	if err := uc.validateAddress(request.Address); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"email":   request.Email,
			"address": request.Address,
			"error":   err.Error(),
		}).Warn("Invalid address")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error(), nil))
		return
	}

	// Create gRPC request
	grpcRequest := &User.UserSignupRequest{
		Email:       request.Email,
		Password:    request.Password,
		Username:    request.FirstName + " " + request.LastName,
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		PhoneNumber: request.PhoneNumber,
		Address: &User.Address{
			StreetName: request.Address.StreetName,
			Locality:   request.Address.Locality,
			State:      request.Address.State,
			Pincode:    request.Address.Pincode,
		},
	}

	resp, err := uc.userClient.UserSignup(context.Background(), grpcRequest)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"email": request.Email,
			"error": err.Error(),
		}).Error("Signup failed")
		c.JSON(http.StatusConflict, model.ErrorResponse(model.ErrSignupFailed, err))
		return
	}

	log.Println("response", resp)

	// Generate JWT token
	resp.Token, err = uc.generateToken(resp.UserId)
	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": resp.UserId,
			"error":  err.Error(),
		}).Error("Failed to generate token")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Failed to generate token", err))
		return
	}

	uc.logger.WithFields(logrus.Fields{
		"email":  request.Email,
		"userId": resp.UserId,
	}).Info("Signup successful")

	c.JSON(http.StatusOK, model.SuccessResponse("Signup successful", resp))
}

// GetProfile retrieves user profile
func (uc *UserController) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetEntityID(c)
	if !exists {
		uc.logger.WithField("path", "/user/profile").Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.ErrUserIDNotFound, nil))
		return
	}

	resp, err := uc.userClient.GetProfile(context.Background(), &User.GetProfileRequest{
		UserId: userID,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": userID,
			"error":  err.Error(),
		}).Error("Failed to retrieve profile")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedRetrieveProfile, err))
		return
	}

	uc.logger.WithField("userId", userID).Info("Profile retrieved successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Profile retrieved successfully", resp))
}

// UpdateProfile handles profile updates
func (uc *UserController) UpdateProfile(c *gin.Context) {
	var request model.UpdateProfileRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/user/profile/update",
		}).Error("Failed to bind update profile request")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidRequestFormat, err))
		return
	}

	userID, exists := middleware.GetEntityID(c)
	if !exists {
		uc.logger.WithField("path", "/user/profile/update").Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.ErrUserIDNotFound, nil))
		return
	}

	if !uc.validateName(request.Name) {
		uc.logger.WithFields(logrus.Fields{
			"userId": userID,
			"name":   request.Name,
		}).Warn("Invalid name format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid name format", nil))
		return
	}

	if !uc.validatePhone(request.PhoneNumber) {
		uc.logger.WithFields(logrus.Fields{
			"userId":      userID,
			"phoneNumber": request.PhoneNumber,
		}).Warn("Invalid phone number format")
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Invalid phone number format", nil))
		return
	}

	resp, err := uc.userClient.UpdateProfile(context.Background(), &User.UpdateProfileRequest{
		UserId:      userID,
		Name:        request.Name,
		PhoneNumber: request.PhoneNumber,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": userID,
			"error":  err.Error(),
		}).Error("Failed to update profile")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedUpdateProfile, err))
		return
	}

	uc.logger.WithField("userId", userID).Info("Profile updated successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Profile updated successfully", resp))
}

// VerifyEmail handles email verification
func (uc *UserController) VerifyEmail(c *gin.Context) {
	var request model.VerifyEmailRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/user/email/verify",
		}).Error("Failed to bind verify email request")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidRequestFormat, err))
		return
	}

	userID, exists := middleware.GetEntityID(c)
	if !exists {
		uc.logger.WithField("path", "/user/email/verify").Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.ErrUserIDNotFound, nil))
		return
	}

	resp, err := uc.userClient.VerifyEmail(context.Background(), &User.EmailVerificationRequest{
		UserId:           userID,
		VerificationCode: request.VerificationCode,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": userID,
			"error":  err.Error(),
		}).Error("Failed to verify email")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrEmailVerificationFailed, err))
		return
	}

	uc.logger.WithField("userId", userID).Info("Email verified successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Email verified successfully", resp))
}

// GetUserByToken retrieves user information using token
func (uc *UserController) GetUserByToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		uc.logger.WithField("path", "/user/token").Warn("Authorization token is missing")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.ErrAuthorizationTokenRequired, nil))
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")

	resp, err := uc.userClient.GetUserByToken(context.Background(), &User.GetUserByTokenRequest{
		Token: token,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"token": token,
			"error": err.Error(),
		}).Error("Failed to retrieve user by token")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedRetrieveUser, err))
		return
	}

	uc.logger.WithField("token", token).Info("User retrieved successfully by token")
	c.JSON(http.StatusOK, model.SuccessResponse("User retrieved successfully", resp))
}

// Address Management

func (uc *UserController) AddAddress(c *gin.Context) {
	var request model.AddAddressRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/user/address/add",
		}).Error("Failed to bind add address request")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidRequestFormat, err))
		return
	}

	userID, exists := middleware.GetEntityID(c)
	if !exists {
		uc.logger.WithField("path", "/user/address/add").Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.ErrUserIDNotFound, nil))
		return
	}

	if err := uc.validateAddress(request.Address); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId":  userID,
			"address": request.Address,
			"error":   err.Error(),
		}).Warn("Invalid address")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error(), nil))
		return
	}

	resp, err := uc.userClient.AddAddress(context.Background(), &User.AddAddressRequest{
		UserId: userID,
		Address: &User.Address{
			StreetName: request.Address.StreetName,
			Locality:   request.Address.Locality,
			State:      request.Address.State,
			Pincode:    request.Address.Pincode,
		},
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": userID,
			"error":  err.Error(),
		}).Error("Failed to add address")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedAddAddress, err))
		return
	}

	uc.logger.WithField("userId", userID).Info("Address added successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Address added successfully", resp))
}

func (uc *UserController) GetAddresses(c *gin.Context) {
	userID, exists := middleware.GetEntityID(c)
	if !exists {
		uc.logger.WithField("path", "/user/addresses").Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.ErrUserIDNotFound, nil))
		return
	}

	resp, err := uc.userClient.GetAddresses(context.Background(), &User.GetAddressesRequest{
		UserId: userID,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": userID,
			"error":  err.Error(),
		}).Error("Failed to retrieve addresses")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedRetrieveAddresses, err))
		return
	}

	uc.logger.WithField("userId", userID).Info("Addresses retrieved successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Addresses retrieved successfully", resp))
}

func (uc *UserController) EditAddress(c *gin.Context) {
	addressID := c.Query("addressId")
	if addressID == "" {
		uc.logger.WithField("path", "/user/address/edit").Warn("Address ID is missing")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrAddressIDRequired, nil))
		return
	}

	var request model.EditAddressRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  "/user/address/edit",
		}).Error("Failed to bind edit address request")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrInvalidRequestFormat, err))
		return
	}

	userID, exists := middleware.GetEntityID(c)
	if !exists {
		uc.logger.WithField("path", "/user/address/edit").Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.ErrUserIDNotFound, nil))
		return
	}

	if err := uc.validateAddress(request.Address); err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId":  userID,
			"address": request.Address,
			"error":   err.Error(),
		}).Warn("Invalid address")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(err.Error(), nil))
		return
	}

	resp, err := uc.userClient.EditAddress(context.Background(), &User.EditAddressRequest{
		UserId:    userID,
		AddressId: addressID,
		Address: &User.Address{
			StreetName: request.Address.StreetName,
			Locality:   request.Address.Locality,
			State:      request.Address.State,
			Pincode:    request.Address.Pincode,
		},
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": userID,
			"error":  err.Error(),
		}).Error("Failed to edit address")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedUpdateAddress, err))
		return
	}

	uc.logger.WithField("userId", userID).Info("Address updated successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Address updated successfully", resp))
}

func (uc *UserController) DeleteAddress(c *gin.Context) {
	addressID := c.Param("addressId")
	if addressID == "" {
		uc.logger.WithField("path", "/user/address/delete").Warn("Address ID is missing")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrAddressIDRequired, nil))
		return
	}

	userID, exists := middleware.GetEntityID(c)
	if !exists {
		uc.logger.WithField("path", "/user/address/delete").Warn("User ID not found in context")
		c.JSON(http.StatusUnauthorized, model.ErrorResponse(model.ErrUserIDNotFound, nil))
		return
	}

	resp, err := uc.userClient.DeleteAddress(context.Background(), &User.DeleteAddressRequest{
		UserId:    userID,
		AddressId: addressID,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": userID,
			"error":  err.Error(),
		}).Error("Failed to delete address")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedDeleteAddress, err))
		return
	}

	uc.logger.WithField("userId", userID).Info("Address deleted successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Address deleted successfully", resp))
}

// Admin Functions

func (uc *UserController) BanUser(c *gin.Context) {
	targetUserID := c.Query("userId")
	if targetUserID == "" {
		uc.logger.WithField("path", "/admin/user/ban").Warn("Target user ID is missing")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrUserIDRequired, nil))
		return
	}

	resp, err := uc.userClient.BanUser(context.Background(), &User.BanUserRequest{
		UserId: targetUserID,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": targetUserID,
			"error":  err.Error(),
		}).Error("Failed to ban user")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedBanUser, err))
		return
	}

	uc.logger.WithFields(logrus.Fields{
		"userId": targetUserID,
	}).Info("User banned successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("User banned successfully", resp))
}

func (uc *UserController) UnBanUser(c *gin.Context) {
	targetUserID := c.Query("userId")
	if targetUserID == "" {
		uc.logger.WithField("path", "/admin/user/unban").Warn("Target user ID is missing")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrUserIDRequired, nil))
		return
	}

	resp, err := uc.userClient.UnBanUser(context.Background(), &User.UnBanUserRequest{
		UserId: targetUserID,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": targetUserID,
			"error":  err.Error(),
		}).Error("Failed to unban user")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedUnbanUser, err))
		return
	}

	uc.logger.WithFields(logrus.Fields{
		"userId": targetUserID,
	}).Info("User unbanned successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("User unbanned successfully", resp))
}

func (uc *UserController) CheckBan(c *gin.Context) {
	targetUserID := c.Query("userId")
	if targetUserID == "" {
		uc.logger.WithField("path", "/admin/user/checkban").Warn("Target user ID is missing")
		c.JSON(http.StatusBadRequest, model.ErrorResponse(model.ErrUserIDRequired, nil))
		return
	}

	resp, err := uc.userClient.CheckBan(context.Background(), &User.CheckBanRequest{
		UserId: targetUserID,
	})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"userId": targetUserID,
			"error":  err.Error(),
		}).Error("Failed to check ban status")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedCheckBan, err))
		return
	}

	uc.logger.WithFields(logrus.Fields{
		"userId": targetUserID,
	}).Info("Ban status checked successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Ban status checked successfully", resp))
}

func (uc *UserController) GetAllUsers(c *gin.Context) {
	resp, err := uc.userClient.GetAllUsers(context.Background(), &User.GetAllUsersRequest{})

	if err != nil {
		uc.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to retrieve all users")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse(model.ErrFailedRetrieveUsers, err))
		return
	}

	uc.logger.WithField("count", len(resp.Users)).Info("All users retrieved successfully")
	c.JSON(http.StatusOK, model.SuccessResponse("Users retrieved successfully", resp))
}
