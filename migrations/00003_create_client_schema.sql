-- +goose Up
-- +goose StatementBegin

-- Clients table
CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    contact_email VARCHAR(255) DEFAULT '',
    contact_phone VARCHAR(20) DEFAULT '',
    website_url TEXT DEFAULT '',
    logo_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_clients_slug ON clients(slug);
CREATE INDEX idx_clients_is_active ON clients(is_active);

CREATE TRIGGER update_clients_updated_at BEFORE UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Client users (scoped to a specific client)
CREATE TABLE client_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) DEFAULT '',
    last_name VARCHAR(100) DEFAULT '',
    phone VARCHAR(20) DEFAULT '',
    avatar_url TEXT,
    user_status_id INTEGER NOT NULL REFERENCES user_status(id),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(client_id, email),
    UNIQUE(client_id, username)
);

CREATE INDEX idx_client_users_client ON client_users(client_id);
CREATE INDEX idx_client_users_email ON client_users(email);
CREATE INDEX idx_client_users_username ON client_users(username);
CREATE INDEX idx_client_users_status ON client_users(user_status_id);

CREATE TRIGGER update_client_users_updated_at BEFORE UPDATE ON client_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Client permissions (scoped to client operations)
CREATE TABLE client_permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_client_permissions_resource ON client_permissions(resource);
CREATE INDEX idx_client_permissions_action ON client_permissions(action);

INSERT INTO client_permissions (name, description, resource, action) VALUES
    ('create_shop', 'Create new shops', 'shop', 'create'),
    ('read_shop', 'View shop details', 'shop', 'read'),
    ('update_shop', 'Update shop information', 'shop', 'update'),
    ('delete_shop', 'Delete shops', 'shop', 'delete'),
    ('manage_shop_users', 'Manage users within a shop', 'shop_user', 'manage'),
    ('create_product', 'Create new products', 'product', 'create'),
    ('read_product', 'View product details', 'product', 'read'),
    ('update_product', 'Update product information', 'product', 'update'),
    ('delete_product', 'Delete products', 'product', 'delete'),
    ('manage_inventory', 'Manage product inventory', 'inventory', 'manage');

-- Client roles
CREATE TABLE client_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    is_system_role BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_client_roles_updated_at BEFORE UPDATE ON client_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

INSERT INTO client_roles (name, description, is_system_role) VALUES
    ('client_admin', 'Full client access with all permissions', true),
    ('shop_manager', 'Can create and manage shops', true),
    ('inventory_manager', 'Can manage products and inventory', true),
    ('viewer', 'Can view shops and products', true);

-- Client role permissions junction
CREATE TABLE client_role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES client_roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES client_permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_client_role_permissions_role ON client_role_permissions(role_id);
CREATE INDEX idx_client_role_permissions_permission ON client_role_permissions(permission_id);

-- Assign all permissions to client_admin
INSERT INTO client_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM client_roles r
CROSS JOIN client_permissions p
WHERE r.name = 'client_admin';

-- Assign permissions to shop_manager
INSERT INTO client_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM client_roles r
CROSS JOIN client_permissions p
WHERE r.name = 'shop_manager'
  AND p.name IN ('create_shop', 'read_shop', 'update_shop', 'manage_shop_users');

-- Assign permissions to inventory_manager
INSERT INTO client_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM client_roles r
CROSS JOIN client_permissions p
WHERE r.name = 'inventory_manager'
  AND p.name IN ('read_shop', 'create_product', 'read_product', 'update_product', 'manage_inventory');

-- Assign permissions to viewer
INSERT INTO client_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM client_roles r
CROSS JOIN client_permissions p
WHERE r.name = 'viewer'
  AND p.name IN ('read_shop', 'read_product');

-- Client user roles junction
CREATE TABLE client_user_roles (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES client_users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES client_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_client_user_roles_user ON client_user_roles(user_id);
CREATE INDEX idx_client_user_roles_role ON client_user_roles(role_id);

CREATE TRIGGER update_client_user_roles_updated_at BEFORE UPDATE ON client_user_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS client_user_roles CASCADE;
DROP TABLE IF EXISTS client_role_permissions CASCADE;
DROP TABLE IF EXISTS client_roles CASCADE;
DROP TABLE IF EXISTS client_permissions CASCADE;
DROP TABLE IF EXISTS client_users CASCADE;
DROP TABLE IF EXISTS clients CASCADE;
-- +goose StatementEnd
