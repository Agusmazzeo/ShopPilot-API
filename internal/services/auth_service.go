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

	// Fetch roles and permissions in a single query
	roles, err := s.platformUserRepo.GetUserRoles(ctx, platformUser.ID)
	if err != nil {
		return nil, err
	}

	roleID, roleName := primaryRole(roles)

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
		ClientID:     nil,
		RoleID:       roleID,
		RoleName:     roleName,
		Permissions:  permissionsFromRoles(roles),
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

	// Fetch fresh roles and permissions from DB on every request
	roles, err := s.platformUserRepo.GetUserRoles(ctx, platformUser.ID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	roleID, roleName := primaryRole(roles)

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
		RoleID:       roleID,
		RoleName:     roleName,
		Permissions:  permissionsFromRoles(roles),
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

// primaryRole returns the ID and name of the first assigned role, or zero values if none.
func primaryRole(roles []*models.PlatformRole) (int, string) {
	if len(roles) == 0 {
		return 0, ""
	}
	return roles[0].ID, roles[0].Name
}

// permissionsFromRoles aggregates all permissions across roles into a deduplicated slice.
func permissionsFromRoles(roles []*models.PlatformRole) []models.Permission {
	seen := make(map[int]struct{})
	var out []models.Permission
	for _, role := range roles {
		for _, p := range role.Permissions {
			if _, ok := seen[p.ID]; ok {
				continue
			}
			seen[p.ID] = struct{}{}
			out = append(out, models.Permission{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Resource:    p.Resource,
				Action:      p.Action,
				CreatedAt:   p.CreatedAt,
			})
		}
	}
	return out
}
