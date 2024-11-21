package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	OrderCart "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/OrderCart"
	Restaurant "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	User "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/User"
	"github.com/liju-github/FoodBuddyAPIGateway/middleware"
	"github.com/sirupsen/logrus"
)

type OrderCartController struct {
	orderCartClient  OrderCart.OrderCartServiceClient
	userClient       User.UserServiceClient
	restaurantClient Restaurant.RestaurantServiceClient
	validator        *validator.Validate
	logger           *logrus.Logger
}

func NewOrderCartController(orderCartClient OrderCart.OrderCartServiceClient, userClient User.UserServiceClient, restaurantClient Restaurant.RestaurantServiceClient) *OrderCartController {
	return &OrderCartController{
		orderCartClient:  orderCartClient,
		userClient:       userClient,
		restaurantClient: restaurantClient,
		validator:        validator.New(),
		logger:           logrus.New(),
	}
}

// Cart Operations

func (oc *OrderCartController) AddProductToCart(c *gin.Context) {
	var req OrderCart.AddProductToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if req.UserId == "" || req.ProductId == "" || req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.AddProductToCart(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) GetCartItems(c *gin.Context) {
	var req OrderCart.GetCartItemsRequest
	req.UserId = c.Query("userId")
	req.RestaurantId = c.Query("restaurantId")

	if req.UserId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.GetCartItems(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) GetAllCarts(c *gin.Context) {
	userId := c.Query("userId")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.GetAllCarts(ctx, &OrderCart.GetAllCartsRequest{UserId: userId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) IncrementProductQuantity(c *gin.Context) {
	var req OrderCart.IncrementProductQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserId == "" || req.ProductId == "" || req.RestaurantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.IncrementProductQuantity(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) DecrementProductQuantity(c *gin.Context) {
	var req OrderCart.DecrementProductQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserId == "" || req.ProductId == "" || req.RestaurantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.DecrementProductQuantity(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) RemoveProductFromCart(c *gin.Context) {
	var req OrderCart.RemoveProductFromCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserId == "" || req.ProductId == "" || req.RestaurantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.RemoveProductFromCart(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) ClearCart(c *gin.Context) {
	var req OrderCart.ClearCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserId == "" || req.RestaurantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.ClearCart(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Order Operations

func (oc *OrderCartController) PlaceOrderByRestID(c *gin.Context) {
	// 1. Parse and validate request
	var req OrderCart.PlaceOrderByRestIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Validate required fields
	if req.UserId == "" || req.RestaurantId == "" || req.DeliveryAddressId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields including delivery address are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 3. Validate user's address
	addrResp, err := oc.userClient.ValidateUserAddress(ctx, &User.ValidateUserAddressRequest{
		UserId:    req.UserId,
		AddressId: req.DeliveryAddressId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate delivery address: " + err.Error()})
		return
	}
	if !addrResp.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery address"})
		return
	}

	// 4. Check restaurant status
	restResp, err := oc.restaurantClient.GetRestaurantByID(ctx, &Restaurant.GetRestaurantByIDRequest{
		RestaurantId: req.RestaurantId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get restaurant details: " + err.Error()})
		return
	}
	if restResp.IsBanned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Restaurant is currently unavailable"})
		return
	}

	// 5. Place the order
	response, err := oc.orderCartClient.PlaceOrderByRestID(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 6. Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": response.Success,
		"orderId": response.OrderId,
		"message": response.Message,
		"order":   response.Order,
	})
}

func (oc *OrderCartController) GetOrderDetailsAll(c *gin.Context) {
	var req OrderCart.GetOrderDetailsAllRequest
	req.UserId = c.Query("userId")
	req.Status = c.Query("status")

	if req.UserId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.GetOrderDetailsAll(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) GetOrderDetailsByID(c *gin.Context) {
	var req OrderCart.GetOrderDetailsByIDRequest
	req.OrderId = c.Query("orderId")
	req.UserId,_ = middleware.GetEntityID(c)

	if req.OrderId == "" || req.UserId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId and userId are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.GetOrderDetailsByID(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) CancelOrder(c *gin.Context) {
	var req OrderCart.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OrderId == "" || req.UserId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId and userId are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.CancelOrder(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) UpdateOrderStatus(c *gin.Context) {
	var req OrderCart.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OrderId == "" || req.RestaurantId == "" || req.NewStatus == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId, restaurantId, and newStatus are required"})
		return
	}

	// Validate order status
	validStatuses := map[string]bool{
		"ACCEPTED":  true,
		"PREPARING": true,
		"READY":     true,
		"DELIVERED": true,
	}

	if !validStatuses[req.NewStatus] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order status"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.UpdateOrderStatus(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) GetRestaurantOrders(c *gin.Context) {
	var req OrderCart.GetRestaurantOrdersRequest
	req.RestaurantId = c.Query("restaurantId")
	req.Status = c.Query("status")

	if req.RestaurantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "restaurantId is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.GetRestaurantOrders(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (oc *OrderCartController) ConfirmOrder(c *gin.Context) {
	var req OrderCart.ConfirmOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.OrderId == "" || req.RestaurantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId and restaurantId are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := oc.orderCartClient.ConfirmOrder(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
