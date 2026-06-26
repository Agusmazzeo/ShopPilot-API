package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/shoppilot/internal/models"
)

// ClientUserRepository defines the interface for client user database operations
type ClientUserRepository interface {
	// User CRUD
	Create(ctx context.Context, user *models.ClientUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.ClientUser, error)
	GetByEmail(ctx context.Context, clientID uuid.UUID, email string) (*models.ClientUser, error)
	GetByUsername(ctx context.Context, clientID uuid.UUID, username string) (*models.ClientUser, error)
	Update(ctx context.Context, user *models.ClientUser) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.ClientUser, error)

	// Role assignments
	AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.ClientRole, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.ClientPermission, error)
}

// clientUserRepository is the concrete implementation
type clientUserRepository struct {
	*BaseRepository
}

// NewClientUserRepository creates a new client user repository
func NewClientUserRepository(pool *pgxpool.Pool) ClientUserRepository {
	return &clientUserRepository{
		BaseRepository: NewBaseRepository(pool),
	}
}

// Create inserts a new client user into the database
func (r *clientUserRepository) Create(ctx context.Context, user *models.ClientUser) error {
	query := `
		INSERT INTO client_users (
			client_id, email, username, password, first_name, last_name,
			phone, avatar_url, user_status_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.ClientID,
		user.Email,
		user.Username,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.AvatarURL,
		user.UserStatusID,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create client user: %w", err)
	}

	return nil
}

// GetByID retrieves a client user by ID
func (r *clientUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ClientUser, error) {
	query := `
		SELECT id, client_id, email, username, password, first_name, last_name,
		       phone, avatar_url, user_status_id, last_login_at, created_at, updated_at
		FROM client_users
		WHERE id = $1
	`

	var user models.ClientUser
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.ClientID,
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
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("client user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get client user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a client user by email within a specific client
func (r *clientUserRepository) GetByEmail(ctx context.Context, clientID uuid.UUID, email string) (*models.ClientUser, error) {
	query := `
		SELECT id, client_id, email, username, password, first_name, last_name,
		       phone, avatar_url, user_status_id, last_login_at, created_at, updated_at
		FROM client_users
		WHERE client_id = $1 AND email = $2
	`

	var user models.ClientUser
	err := r.pool.QueryRow(ctx, query, clientID, email).Scan(
		&user.ID,
		&user.ClientID,
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
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("client user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get client user by email: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a client user by username within a specific client
func (r *clientUserRepository) GetByUsername(ctx context.Context, clientID uuid.UUID, username string) (*models.ClientUser, error) {
	query := `
		SELECT id, client_id, email, username, password, first_name, last_name,
		       phone, avatar_url, user_status_id, last_login_at, created_at, updated_at
		FROM client_users
		WHERE client_id = $1 AND username = $2
	`

	var user models.ClientUser
	err := r.pool.QueryRow(ctx, query, clientID, username).Scan(
		&user.ID,
		&user.ClientID,
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
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("client user not found with username: %s", username)
		}
		return nil, fmt.Errorf("failed to get client user by username: %w", err)
	}

	return &user, nil
}

// Update updates an existing client user
func (r *clientUserRepository) Update(ctx context.Context, user *models.ClientUser) error {
	query := `
		UPDATE client_users
		SET email = $1, username = $2, password = $3, first_name = $4,
		    last_name = $5, phone = $6, avatar_url = $7, user_status_id = $8,
		    last_login_at = $9, updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.Email,
		user.Username,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.AvatarURL,
		user.UserStatusID,
		user.LastLoginAt,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("client user not found: %s", user.ID)
		}
		return fmt.Errorf("failed to update client user: %w", err)
	}

	return nil
}

// Delete removes a client user by ID
func (r *clientUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM client_users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete client user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client user not found: %s", id)
	}

	return nil
}

// ListByClient retrieves all users for a specific client with pagination
func (r *clientUserRepository) ListByClient(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*models.ClientUser, error) {
	query := `
		SELECT id, client_id, email, username, password, first_name, last_name,
		       phone, avatar_url, user_status_id, last_login_at, created_at, updated_at
		FROM client_users
		WHERE client_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list client users: %w", err)
	}
	defer rows.Close()

	var users []*models.ClientUser
	for rows.Next() {
		var user models.ClientUser
		err := rows.Scan(
			&user.ID,
			&user.ClientID,
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
			return nil, fmt.Errorf("failed to scan client user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client users: %w", err)
	}

	return users, nil
}

// AssignRole assigns a role to a client user
func (r *clientUserRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	query := `
		INSERT INTO client_user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to assign role to client user: %w", err)
	}

	return nil
}

// RemoveRole removes a role from a client user
func (r *clientUserRepository) RemoveRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	query := `DELETE FROM client_user_roles WHERE user_id = $1 AND role_id = $2`

	result, err := r.pool.Exec(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from client user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("role assignment not found for user: %s, role: %d", userID, roleID)
	}

	return nil
}

// GetUserRoles retrieves all roles assigned to a client user
func (r *clientUserRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.ClientRole, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_system_role, r.created_at, r.updated_at
		FROM client_roles r
		INNER JOIN client_user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.ClientRole
	for rows.Next() {
		var role models.ClientRole
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsSystemRole,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client role: %w", err)
		}
		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client roles: %w", err)
	}

	return roles, nil
}

// GetUserPermissions retrieves all permissions granted to a client user through their roles
func (r *clientUserRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.ClientPermission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM client_permissions p
		INNER JOIN client_role_permissions rp ON p.id = rp.permission_id
		INNER JOIN client_user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.resource, p.action
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.ClientPermission
	for rows.Next() {
		var permission models.ClientPermission
		err := rows.Scan(
			&permission.ID,
			&permission.Name,
			&permission.Description,
			&permission.Resource,
			&permission.Action,
			&permission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client permission: %w", err)
		}
		permissions = append(permissions, &permission)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client permissions: %w", err)
	}

	return permissions, nil
}
