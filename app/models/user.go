package models

import "time"

// User represents a user in the system
type User struct {
	ID           int       `json:"id" db:"id"`
	ClientID     int       `json:"client_id" db:"client_id"`
	Email        string    `json:"email" db:"email"`
	Username     string    `json:"username" db:"username"`
	Password     string    `json:"-" db:"password"` // Never send password in JSON
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Phone        string    `json:"phone" db:"phone"`
	AvatarURL    *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	UserStatusID int       `json:"user_status_id" db:"user_status_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	ClientID  int     `json:"client_id" validate:"required,gt=0"`
	Email     string  `json:"email" validate:"required,email"`
	Username  string  `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password  string  `json:"password" validate:"required,min=8"`
	FirstName string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string  `json:"last_name" validate:"required,min=1,max=100"`
	Phone     *string `json:"phone,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Phone     *string `json:"phone,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UserWithStatus extends User with status information
type UserWithStatus struct {
	User
	StatusName string `json:"status_name" db:"status_name"`
}
