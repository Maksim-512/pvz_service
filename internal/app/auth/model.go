package auth

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleEmployee  = "employee"
	RoleModerator = "moderator"
)

type User struct {
	ID           uuid.UUID `json:"id,omitempty"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role" `
}

type RegisterResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DummyLoginRequest struct {
	Role string `json:"role"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
