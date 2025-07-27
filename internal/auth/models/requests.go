package models

// Estruturas para requests
type CreateUserRequest struct {
	Username      string `json:"username" binding:"required,alphanum"`
	Password      string `json:"password" binding:"required,min=6"`
	FullName      string `json:"full_name" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
	IsWhitelisted bool   `json:"is_whitelisted,omitempty"`
}

type LoginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type RenewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
} 