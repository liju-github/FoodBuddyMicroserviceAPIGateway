package router

import (
	"github.com/gin-gonic/gin"
	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	adminPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Admin"
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

	// Restaurant Client setup
	restaurantClient := restaurantPb.NewRestaurantServiceClient(Client.ConnRestaurant)
	restaurantController := controller.NewRestaurantController(restaurantClient)
	SetupRestaurantRoutes(router, restaurantController)

	//Admin Client setup
	adminClient := adminPb.NewAdminServiceClient(Client.ConnAdmin)
	adminController := controller.NewAdminController(adminClient)
	SetUpAdminAuth(router, adminController)
}

func SetUpAdminAuth(router *gin.Engine, adminController *controller.AdminController) {
	router.POST("/admin/login", adminController.AdminLogin)
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
		admin.GET("", userController.GetAllUsers)        // Get all users
		admin.POST("/ban", userController.BanUser)       // Ban a user
		admin.POST("/unban", userController.UnBanUser)   // Unban a user
		admin.GET("/check-ban", userController.CheckBan) // Check if user is banned
	}
}

// SetupRestaurantRoutes configures routes for Restaurant-related operations
func SetupRestaurantRoutes(router *gin.Engine, restaurantController *controller.RestaurantController) {
	// Public routes (no authentication required)
	auth := router.Group("/auth/restaurant")
	{
		auth.POST("/signup", restaurantController.RestaurantSignup)
		auth.POST("/login", restaurantController.RestaurantLogin)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/restaurants")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		// Restaurant management (requires restaurant authentication)
		restaurant := protected.Group("")
		restaurant.Use(middleware.RestaurantAuthMiddleware())
		{
			restaurant.PUT("/profile", restaurantController.EditRestaurant)

			// Product management
			products := restaurant.Group("/products")
			{
				products.POST("", restaurantController.AddProduct)
				products.PUT("", restaurantController.EditProduct)
				products.DELETE("", restaurantController.DeleteProductByID)
				products.PUT("/stock/increment", restaurantController.IncrementProductStock)
				products.PUT("/stock/decrement", restaurantController.DecrementProductStock)
			}
		}

		// Admin routes (requires admin authentication)
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware())
		{
			admin.PUT("/products", restaurantController.EditProduct)
			admin.DELETE("/products", restaurantController.DeleteProductByID)
			admin.PUT("/products/stock/increment", restaurantController.IncrementProductStock)
			admin.PUT("/products/stock/decrement", restaurantController.DecrementProductStock)
			admin.POST("/ban", restaurantController.BanRestaurant)
			admin.POST("/unban", restaurantController.UnbanRestaurant)
		}
	}

	// Public restaurant and product routes (no authentication required)
	public := router.Group("/api/public/restaurants")
	{
		public.GET("", restaurantController.GetAllRestaurantWithProducts)
		public.GET("/products", restaurantController.GetRestaurantProductsByID) // Query param: restaurantId
		public.GET("/products/single", restaurantController.GetProductByID)     // Query param: productId
		public.GET("/products/stock", restaurantController.GetStockByProductID) // Query param: productId
		public.GET("/id", restaurantController.GetRestaurantIDviaProductID)     // Query param: productId
	}
}
