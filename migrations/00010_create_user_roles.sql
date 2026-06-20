-- +goose Up
-- +goose StatementBegin

-- Create user_roles lookup table
CREATE TABLE user_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    permissions JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default roles
INSERT INTO user_roles (name, description, permissions) VALUES
    ('owner', 'Shop owner with full access',
     '{"products": ["create", "read", "update", "delete"], "categories": ["create", "read", "update", "delete"], "users": ["create", "read", "update", "delete"], "shop": ["read", "update", "delete"]}'::jsonb),
    ('admin', 'Administrator with most permissions',
     '{"products": ["create", "read", "update", "delete"], "categories": ["create", "read", "update", "delete"], "users": ["read"], "shop": ["read"]}'::jsonb),
    ('editor', 'Can create and edit products and categories',
     '{"products": ["create", "read", "update"], "categories": ["create", "read", "update"], "users": [], "shop": ["read"]}'::jsonb),
    ('viewer', 'Read-only access',
     '{"products": ["read"], "categories": ["read"], "users": [], "shop": ["read"]}'::jsonb);

-- Create shop_users junction table (user-shop-role relationship)
CREATE TABLE shop_users (
    id SERIAL PRIMARY KEY,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    shop_id INTEGER NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES user_roles(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(shop_id, user_id)
);

-- Create indexes
CREATE INDEX idx_shop_users_client_id ON shop_users(client_id);
CREATE INDEX idx_shop_users_shop_id ON shop_users(shop_id);
CREATE INDEX idx_shop_users_user_id ON shop_users(user_id);
CREATE INDEX idx_shop_users_role_id ON shop_users(role_id);

-- Create trigger to update updated_at
CREATE TRIGGER update_shop_users_updated_at BEFORE UPDATE ON shop_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_shop_users_updated_at ON shop_users;
DROP TABLE IF EXISTS shop_users CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
-- +goose StatementEnd
