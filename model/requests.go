package model

// Address represents the address structure
type Address struct {
	StreetName string `json:"streetName" binding:"required"`
	Locality   string `json:"locality" binding:"required"`
	State      string `json:"state" binding:"required"`
	Pincode    string `json:"pincode" binding:"required"`
}

// LoginRequest represents the request structure for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// SignupRequest represents the request structure for user signup
type SignupRequest struct {
	Email       string  `json:"email" binding:"required,email"`
	Password    string  `json:"password" binding:"required,min=8"`
	FirstName   string  `json:"firstName" binding:"required"`
	LastName    string  `json:"lastName" binding:"required"`
	PhoneNumber uint64  `json:"phoneNumber" binding:"required"`
	Address     Address `json:"address" binding:"required"`
}

// UpdateProfileRequest represents the request structure for profile updates
type UpdateProfileRequest struct {
	Name        string `json:"name" `
	PhoneNumber uint64 `json:"phoneNumber" `
}

// VerifyEmailRequest represents the request structure for email verification
type VerifyEmailRequest struct {
	VerificationCode string `json:"verificationCode" binding:"required"`
}

// AddAddressRequest represents the request structure for adding an address
type AddAddressRequest struct {
	Address Address `json:"address" binding:"required"`
}

// GetAddressesRequest represents the request structure for getting addresses
type GetAddressesRequest struct {
}

// EditAddressRequest represents the request structure for editing an address
type EditAddressRequest struct {
	Address Address `json:"address" binding:"required"`
}

// DeleteAddressRequest represents the request structure for deleting an address
type DeleteAddressRequest struct {
}

// BanUserRequest represents the request structure for banning a user
type BanUserRequest struct {
}

// UnbanUserRequest represents the request structure for unbanning a user
type UnbanUserRequest struct {
}

// CheckBanRequest represents the request structure for checking user ban status
type CheckBanRequest struct {
}

// GetUserByTokenRequest represents the request structure for getting user by token
type GetUserByTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// GetAllUsersRequest represents an empty request for getting all users
type GetAllUsersRequest struct{}

// RestaurantLoginRequest represents the request structure for restaurant login
type RestaurantLoginRequest struct {
	OwnerEmail string `json:"ownerEmail" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
}

// RestaurantSignupRequest represents the request structure for restaurant signup
type RestaurantSignupRequest struct {
	RestaurantName string  `json:"restaurantName" binding:"required"`
	OwnerEmail     string  `json:"ownerEmail" binding:"required,email"`
	Password       string  `json:"password" binding:"required,min=8"`
	PhoneNumber    uint64  `json:"phoneNumber" binding:"required"`
	Address        Address `json:"address" binding:"required"`
}
