package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourorg/shoppilot/internal/models"
)

// JWTClaims represents the claims stored in the JWT token
type JWTClaims struct {
	UserID      uuid.UUID         `json:"userId"`
	Email       string            `json:"email"`
	Username    string            `json:"username"`
	RoleID      int               `json:"roleId"`
	RoleName    string            `json:"roleName"`
	Permissions []PermissionClaim `json:"permissions"`
	jwt.RegisteredClaims
}

// PermissionClaim represents a permission in the JWT token
type PermissionClaim struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// GenerateToken generates a JWT token for a user with their roles and permissions
func (j *JWTManager) GenerateToken(user *models.AuthUser) (string, error) {
	// Convert permissions to claims
	var permissionClaims []PermissionClaim
	for _, permission := range user.Permissions {
		permissionClaims = append(permissionClaims, PermissionClaim{
			ID:       permission.ID,
			Name:     permission.Name,
			Resource: permission.Resource,
			Action:   permission.Action,
		})
	}

	// Create claims
	claims := JWTClaims{
		UserID:      user.ID,
		Email:       user.Email,
		Username:    user.Username,
		RoleID:      user.RoleID,
		RoleName:    user.RoleName,
		Permissions: permissionClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "shoppilot",
			Subject:   user.Email,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token and extract claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new token from an existing valid token
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	// Validate the existing token
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Create new claims with updated expiration
	newClaims := JWTClaims{
		UserID:      claims.UserID,
		Email:       claims.Email,
		Username:    claims.Username,
		RoleID:      claims.RoleID,
		RoleName:    claims.RoleName,
		Permissions: claims.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "shoppilot",
			Subject:   claims.Email,
		},
	}

	// Create and sign new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString(j.secretKey)
}

// HasPermission checks if the token claims contain a specific permission
func (claims *JWTClaims) HasPermission(resource, action string) bool {
	for _, permission := range claims.Permissions {
		if permission.Resource == resource && permission.Action == action {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the token claims contain any of the specified permissions
func (claims *JWTClaims) HasAnyPermission(permissions []string) bool {
	for _, requiredPerm := range permissions {
		for _, userPerm := range claims.Permissions {
			if userPerm.Name == requiredPerm {
				return true
			}
		}
	}
	return false
}
