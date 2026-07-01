package services

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/yourorg/shoppilot/app/repositories"
	"github.com/yourorg/shoppilot/internal/models"
	"github.com/yourorg/shoppilot/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotActive = errors.New("user account is not active")
	ErrInvalidToken  = errors.New("invalid token")
)

// LoginResult represents the result of a successful login
type LoginResult struct {
	Token string            `json:"token"`
	User  *models.AuthUser `json:"user"`
}

// AuthServiceI defines the interface for authentication business logic
type AuthServiceI interface {
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	ValidateToken(ctx context.Context, token string) (*models.AuthUser, error)
	RefreshToken(ctx context.Context, token string) (string, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error
}

// AuthService implements the AuthServiceI interface
type AuthService struct {
	platformUserRepo repositories.PlatformUserRepository
	jwtManager       *utils.JWTManager
}

// NewAuthService creates a new auth service
// For now, we only support PlatformUser authentication
// TODO: Add ClientUser authentication support
func NewAuthService(
	platformUserRepo repositories.PlatformUserRepository,
	jwtSecretKey string,
	jwtExpirationHours int,
) AuthServiceI {
	jwtManager := utils.NewJWTManager(jwtSecretKey, time.Duration(jwtExpirationHours)*time.Hour)

	return &AuthService{
		platformUserRepo: platformUserRepo,
		jwtManager:       jwtManager,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	// Try to find PlatformUser
	platformUser, err := s.platformUserRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// TODO: Try ClientUser if PlatformUser not found
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := s.VerifyPassword(platformUser.Password, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if platformUser.UserStatusID != models.UserStatusActive {
		return nil, ErrUserNotActive
	}

	// TODO: Fetch role and permissions
	// For now, create a basic AuthUser
	authUser := &models.AuthUser{
		ID:           platformUser.ID,
		Email:        platformUser.Email,
		Username:     platformUser.Username,
		FirstName:    platformUser.FirstName,
		LastName:     platformUser.LastName,
		Phone:        platformUser.Phone,
		AvatarURL:    platformUser.AvatarURL,
		UserStatusID: platformUser.UserStatusID,
		LastLoginAt:  platformUser.LastLoginAt,
		CreatedAt:    platformUser.CreatedAt,
		UpdatedAt:    platformUser.UpdatedAt,
		ClientID:     nil, // PlatformUser
		RoleID:       1,   // TODO: Get from database
		RoleName:     "Admin",
		Permissions: []models.Permission{
			// TODO: Load from database
		},
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(authUser)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &LoginResult{
		Token: token,
		User:  authUser,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.AuthUser, error) {
	// Validate JWT token
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Get fresh user data from database to ensure it's still active
	platformUser, err := s.platformUserRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check if user is still active
	if platformUser.UserStatusID != models.UserStatusActive {
		return nil, ErrUserNotActive
	}

	// TODO: Fetch role and permissions
	authUser := &models.AuthUser{
		ID:           platformUser.ID,
		Email:        platformUser.Email,
		Username:     platformUser.Username,
		FirstName:    platformUser.FirstName,
		LastName:     platformUser.LastName,
		Phone:        platformUser.Phone,
		AvatarURL:    platformUser.AvatarURL,
		UserStatusID: platformUser.UserStatusID,
		LastLoginAt:  platformUser.LastLoginAt,
		CreatedAt:    platformUser.CreatedAt,
		UpdatedAt:    platformUser.UpdatedAt,
		RoleID:       claims.RoleID,
		RoleName:     claims.RoleName,
		Permissions:  convertPermissionClaims(claims.Permissions),
	}

	return authUser, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	// Validate the current token and get fresh user data
	user, err := s.ValidateToken(ctx, token)
	if err != nil {
		return "", err
	}

	// Generate a new token with fresh user data
	newToken, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return "", errors.New("failed to generate new token")
	}

	return newToken, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func convertPermissionClaims(claims []utils.PermissionClaim) []models.Permission {
	permissions := make([]models.Permission, len(claims))
	for i, claim := range claims {
		permissions[i] = models.Permission{
			ID:       claim.ID,
			Name:     claim.Name,
			Resource: claim.Resource,
			Action:   claim.Action,
		}
	}
	return permissions
}
