package models

import (
	"time"

	"github.com/google/uuid"
)

type ClientUser struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	ClientID     uuid.UUID  `db:"client_id" json:"clientId"`
	Email        string     `db:"email" json:"email"`
	Username     string     `db:"username" json:"username"`
	Password     string     `db:"password" json:"-"` // Never expose in JSON
	FirstName    string     `db:"first_name" json:"firstName"`
	LastName     string     `db:"last_name" json:"lastName"`
	Phone        string     `db:"phone" json:"phone"`
	AvatarURL    *string    `db:"avatar_url" json:"avatarUrl"`
	UserStatusID int        `db:"user_status_id" json:"userStatusId"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"lastLoginAt"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
}

type ClientRole struct {
	ID           int       `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Description  string    `db:"description" json:"description"`
	IsSystemRole bool      `db:"is_system_role" json:"isSystemRole"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

type ClientPermission struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Resource    string    `db:"resource" json:"resource"`
	Action      string    `db:"action" json:"action"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

type ClientUserRole struct {
	ID        int       `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"userId"`
	RoleID    int       `db:"role_id" json:"roleId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
