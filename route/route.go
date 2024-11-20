package router

import (
	"github.com/gin-gonic/gin"
	user "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/User"
	"github.com/liju-github/FoodBuddyAPIGateway/clients"
	"github.com/liju-github/FoodBuddyAPIGateway/controller"
	"github.com/liju-github/FoodBuddyAPIGateway/middleware"
)

// InitializeServiceRoutes initializes gRPC clients for each service, creates controllers,
// and configures routes for each service.
func InitializeServiceRoutes(router *gin.Engine, Client *clients.ClientConnections) {
	// User Client setup
	userClient := user.NewUserServiceClient(Client.ConnUser)
	userController := controller.NewUserController(userClient)
	SetupUserRoutes(router, userController)
}

// SetupUserRoutes configures routes for User-related operations
func SetupUserRoutes(router *gin.Engine, userController *controller.UserController) {
	// Public routes (no authentication required)
	auth := router.Group("/auth/user")
	{
		auth.POST("/signup", userController.Signup)
		auth.POST("/login", userController.Login)
		auth.POST("/verify-email", userController.VerifyEmail)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/users")
	protected.Use(middleware.JWTAuthMiddleware(), middleware.UserAuthMiddleware())
	{
		// Profile management
		protected.GET("/profile/:userId", userController.GetProfile)
		protected.PUT("/profile", userController.UpdateProfile)

		// Address management
		address := protected.Group("/address")
		{
			address.POST("/", userController.AddAddress)
			address.GET("/", userController.GetAddresses)
			address.PUT("/", userController.EditAddress)
			address.DELETE("/", userController.DeleteAddress)
		}
	}

	// Admin routes
	admin := router.Group("/admin/users")
	// admin.Use(middleware.JWTAuthMiddleware(), middleware.AdminAuthMiddleware())
	{
		admin.GET("", userController.GetAllUsers)                // Get all users
		admin.POST("/ban", userController.BanUser)               // Ban a user
		admin.POST("/unban", userController.UnBanUser)           // Unban a user
		admin.GET("/check-ban", userController.CheckBan) // Check if user is banned
	}
}
