package model

// GenericResponse represents a generic API response
type GenericResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UserProfile represents user profile data
type UserProfile struct {
	UserID      string    `json:"userId"`
	Email       string    `json:"email"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	PhoneNumber uint64    `json:"phoneNumber"`
	Addresses   []Address `json:"addresses,omitempty"`
	IsVerified  bool      `json:"isVerified"`
	IsBanned    bool      `json:"isBanned"`
}

// LoginResponse represents the response for login
type LoginResponse struct {
	Token       string      `json:"token"`
	UserProfile UserProfile `json:"userProfile"`
}

// AddressResponse represents the response for address operations
type AddressResponse struct {
	AddressID string  `json:"addressId"`
	Address   Address `json:"address"`
}

// ErrorResponse creates a new error response
func ErrorResponse(message string, err error) *GenericResponse {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return &GenericResponse{
		Success: false,
		Message: message,
		Error:   errMsg,
	}
}

// SuccessResponse creates a new success response
func SuccessResponse(message string, data interface{}) *GenericResponse {
	return &GenericResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}
