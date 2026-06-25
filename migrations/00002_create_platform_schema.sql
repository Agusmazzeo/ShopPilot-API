-- +goose Up
-- +goose StatementBegin

-- User status lookup table (shared by platform and client users)
CREATE TABLE user_status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO user_status (name, description) VALUES
    ('ACTIVE', 'User is active and can access the system'),
    ('INACTIVE', 'User is inactive and cannot access the system'),
    ('INVITED', 'User has been invited but has not yet activated their account'),
    ('SUSPENDED', 'User account has been suspended');

CREATE INDEX idx_user_status_name ON user_status(name);

-- Platform users (no client_id - these are global admins)
CREATE TABLE platform_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) DEFAULT '',
    last_name VARCHAR(100) DEFAULT '',
    phone VARCHAR(20) DEFAULT '',
    avatar_url TEXT,
    user_status_id INTEGER NOT NULL REFERENCES user_status(id),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_platform_users_email ON platform_users(email);
CREATE INDEX idx_platform_users_username ON platform_users(username);
CREATE INDEX idx_platform_users_status ON platform_users(user_status_id);

CREATE TRIGGER update_platform_users_updated_at BEFORE UPDATE ON platform_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Platform permissions
CREATE TABLE platform_permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_platform_permissions_resource ON platform_permissions(resource);
CREATE INDEX idx_platform_permissions_action ON platform_permissions(action);

INSERT INTO platform_permissions (name, description, resource, action) VALUES
    ('create_client', 'Create new clients', 'client', 'create'),
    ('read_client', 'View client details', 'client', 'read'),
    ('update_client', 'Update client information', 'client', 'update'),
    ('delete_client', 'Delete clients', 'client', 'delete'),
    ('manage_client_users', 'Manage users within a client', 'client_user', 'manage'),
    ('view_platform_users', 'View platform admin users', 'platform_user', 'read'),
    ('create_platform_user', 'Create platform admin users', 'platform_user', 'create'),
    ('update_platform_user', 'Update platform admin users', 'platform_user', 'update'),
    ('delete_platform_user', 'Delete platform admin users', 'platform_user', 'delete'),
    ('manage_platform_roles', 'Manage platform roles and permissions', 'platform_role', 'manage');

-- Platform roles
CREATE TABLE platform_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    is_system_role BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_platform_roles_updated_at BEFORE UPDATE ON platform_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

INSERT INTO platform_roles (name, description, is_system_role) VALUES
    ('super_admin', 'Full platform access with all permissions', true),
    ('client_manager', 'Can create and manage clients', true),
    ('support', 'Can view clients and assist with support', true);

-- Platform role permissions junction
CREATE TABLE platform_role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES platform_roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES platform_permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_platform_role_permissions_role ON platform_role_permissions(role_id);
CREATE INDEX idx_platform_role_permissions_permission ON platform_role_permissions(permission_id);

-- Assign all permissions to super_admin
INSERT INTO platform_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM platform_roles r
CROSS JOIN platform_permissions p
WHERE r.name = 'super_admin';

-- Assign permissions to client_manager
INSERT INTO platform_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM platform_roles r
CROSS JOIN platform_permissions p
WHERE r.name = 'client_manager'
  AND p.name IN ('create_client', 'read_client', 'update_client', 'manage_client_users');

-- Assign permissions to support
INSERT INTO platform_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM platform_roles r
CROSS JOIN platform_permissions p
WHERE r.name = 'support'
  AND p.name IN ('read_client', 'manage_client_users');

-- Platform user roles junction
CREATE TABLE platform_user_roles (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES platform_users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES platform_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_platform_user_roles_user ON platform_user_roles(user_id);
CREATE INDEX idx_platform_user_roles_role ON platform_user_roles(role_id);

CREATE TRIGGER update_platform_user_roles_updated_at BEFORE UPDATE ON platform_user_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default super admin user
-- Password: 'admin123' (bcrypt hash with cost 10)
-- IMPORTANT: Change this password after first login!
INSERT INTO platform_users (email, username, password, first_name, last_name, user_status_id)
VALUES ('admin@shoppilot.com', 'superadmin', '$2a$10$xG2zJuLNETRmz/MvMGJ5ce7RkZVMRbNoCLeyUKGgnaVf3beNu0e5K', 'Super', 'Admin', 1);

-- Assign super_admin role to default user
INSERT INTO platform_user_roles (user_id, role_id)
SELECT u.id, r.id
FROM platform_users u
CROSS JOIN platform_roles r
WHERE u.username = 'superadmin' AND r.name = 'super_admin';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS platform_user_roles CASCADE;
DROP TABLE IF EXISTS platform_role_permissions CASCADE;
DROP TABLE IF EXISTS platform_roles CASCADE;
DROP TABLE IF EXISTS platform_permissions CASCADE;
DROP TABLE IF EXISTS platform_users CASCADE;
DROP TABLE IF EXISTS user_status CASCADE;
-- +goose StatementEnd
