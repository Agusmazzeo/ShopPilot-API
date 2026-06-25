package models

import (
	"time"

	"github.com/google/uuid"
)

type PlatformRole struct {
	ID           int       `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Description  string    `db:"description" json:"description"`
	IsSystemRole bool      `db:"is_system_role" json:"isSystemRole"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

type PlatformPermission struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Resource    string    `db:"resource" json:"resource"`
	Action      string    `db:"action" json:"action"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

type PlatformUserRole struct {
	ID        int       `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"userId"`
	RoleID    int       `db:"role_id" json:"roleId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
