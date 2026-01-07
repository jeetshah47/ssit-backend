package dto

// UserResponse represents a user in API responses
type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Phone         *string `json:"phone,omitempty"`
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	IsMfCustomer  bool    `json:"isMfCustomer"`
	EmailVerified bool    `json:"emailVerified"`
	CreatedAt     string  `json:"createdAt"`
}

// GetUserParams represents path parameters for get user
type GetUserParams struct {
	ID string `uri:"id" binding:"required,uuid"`
}
