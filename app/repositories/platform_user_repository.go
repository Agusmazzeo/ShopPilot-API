package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/shoppilot/internal/models"
)

// PlatformUserRepository defines the interface for platform user data operations
type PlatformUserRepository interface {
	// User CRUD
	Create(ctx context.Context, user *models.PlatformUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.PlatformUser, error)
	GetByEmail(ctx context.Context, email string) (*models.PlatformUser, error)
	GetByUsername(ctx context.Context, username string) (*models.PlatformUser, error)
	Update(ctx context.Context, user *models.PlatformUser) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.PlatformUser, error)

	// Role assignments
	AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.PlatformRole, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.PlatformPermission, error)
}

// platformUserRepository is the concrete implementation of PlatformUserRepository
type platformUserRepository struct {
	*BaseRepository
}

// NewPlatformUserRepository creates a new instance of platform user repository
func NewPlatformUserRepository(pool *pgxpool.Pool) PlatformUserRepository {
	return &platformUserRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new platform user into the database
func (r *platformUserRepository) Create(ctx context.Context, user *models.PlatformUser) error {
	query := `
		INSERT INTO platform_users (
			id, email, username, password, first_name, last_name,
			phone, avatar_url, user_status_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Generate UUID if not provided
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set timestamps
	now := GetCurrentTimestamp()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Default status if not set (1 = active)
	if user.UserStatusID == 0 {
		user.UserStatusID = 1
	}

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Phone,
		NullableString(user.AvatarURL),
		user.UserStatusID,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetByID retrieves a platform user by their UUID
func (r *platformUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PlatformUser, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name,
		       phone, avatar_url, user_status_id, last_login_at,
		       created_at, updated_at
		FROM platform_users
		WHERE id = $1
	`

	var user models.PlatformUser
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.AvatarURL,
		&user.UserStatusID,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a platform user by their email address
func (r *platformUserRepository) GetByEmail(ctx context.Context, email string) (*models.PlatformUser, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name,
		       phone, avatar_url, user_status_id, last_login_at,
		       created_at, updated_at
		FROM platform_users
		WHERE email = $1
	`

	var user models.PlatformUser
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.AvatarURL,
		&user.UserStatusID,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByUsername retrieves a platform user by their username
func (r *platformUserRepository) GetByUsername(ctx context.Context, username string) (*models.PlatformUser, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name,
		       phone, avatar_url, user_status_id, last_login_at,
		       created_at, updated_at
		FROM platform_users
		WHERE username = $1
	`

	var user models.PlatformUser
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.AvatarURL,
		&user.UserStatusID,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update modifies an existing platform user
func (r *platformUserRepository) Update(ctx context.Context, user *models.PlatformUser) error {
	query := `
		UPDATE platform_users
		SET email = $2,
		    username = $3,
		    password = $4,
		    first_name = $5,
		    last_name = $6,
		    phone = $7,
		    avatar_url = $8,
		    user_status_id = $9,
		    last_login_at = $10,
		    updated_at = $11
		WHERE id = $1
	`

	user.UpdatedAt = GetCurrentTimestamp()

	result, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Phone,
		NullableString(user.AvatarURL),
		user.UserStatusID,
		user.LastLoginAt,
		user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("platform user with id %s not found", user.ID)
	}

	return nil
}

// Delete removes a platform user from the database
func (r *platformUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM platform_users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("platform user with id %s not found", id)
	}

	return nil
}

// List retrieves a paginated list of platform users
func (r *platformUserRepository) List(ctx context.Context, limit, offset int) ([]*models.PlatformUser, error) {
	query := `
		SELECT id, email, username, password, first_name, last_name,
		       phone, avatar_url, user_status_id, last_login_at,
		       created_at, updated_at
		FROM platform_users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.PlatformUser
	for rows.Next() {
		var user models.PlatformUser
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.Phone,
			&user.AvatarURL,
			&user.UserStatusID,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// AssignRole assigns a role to a platform user
func (r *platformUserRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	query := `
		INSERT INTO platform_user_roles (user_id, role_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	now := GetCurrentTimestamp()

	_, err := r.pool.Exec(ctx, query, userID, roleID, now, now)
	return err
}

// RemoveRole removes a role from a platform user
func (r *platformUserRepository) RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	query := `DELETE FROM platform_user_roles WHERE user_id = $1 AND role_id = $2`

	result, err := r.pool.Exec(ctx, query, userID, roleID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role assignment not found for user %s and role %d", userID, roleID)
	}

	return nil
}

// GetUserRoles retrieves all roles assigned to a platform user, including their permissions.
func (r *platformUserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.PlatformRole, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_system_role, r.created_at, r.updated_at,
		       p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM platform_roles r
		INNER JOIN platform_user_roles ur ON r.id = ur.role_id
		LEFT JOIN platform_role_permissions rp ON r.id = rp.role_id
		LEFT JOIN platform_permissions p ON rp.permission_id = p.id
		WHERE ur.user_id = $1
		ORDER BY r.name, p.resource, p.action
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roleMap := make(map[int]*models.PlatformRole)
	var roleOrder []int

	for rows.Next() {
		var role models.PlatformRole
		var pID *int
		var pName, pDesc, pResource, pAction *string
		var pCreatedAt *time.Time

		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &role.IsSystemRole, &role.CreatedAt, &role.UpdatedAt,
			&pID, &pName, &pDesc, &pResource, &pAction, &pCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := roleMap[role.ID]; !exists {
			r := role
			r.Permissions = []models.PlatformPermission{}
			roleMap[role.ID] = &r
			roleOrder = append(roleOrder, role.ID)
		}

		if pID != nil {
			roleMap[role.ID].Permissions = append(roleMap[role.ID].Permissions, models.PlatformPermission{
				ID:          *pID,
				Name:        *pName,
				Description: *pDesc,
				Resource:    *pResource,
				Action:      *pAction,
				CreatedAt:   *pCreatedAt,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	roles := make([]*models.PlatformRole, 0, len(roleOrder))
	for _, id := range roleOrder {
		roles = append(roles, roleMap[id])
	}

	return roles, nil
}

// GetUserPermissions retrieves all permissions for a platform user (resolved from roles)
func (r *platformUserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.PlatformPermission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM platform_permissions p
		INNER JOIN platform_role_permissions rp ON p.id = rp.permission_id
		INNER JOIN platform_user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.resource, p.action
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*models.PlatformPermission
	for rows.Next() {
		var permission models.PlatformPermission
		err := rows.Scan(
			&permission.ID,
			&permission.Name,
			&permission.Description,
			&permission.Resource,
			&permission.Action,
			&permission.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
