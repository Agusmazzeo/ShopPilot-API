-- +goose Up
-- +goose StatementBegin

-- Create shops table with multi-tenant support
CREATE TABLE shops (
    id SERIAL PRIMARY KEY,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    logo_url TEXT,
    theme VARCHAR(50) DEFAULT 'default',
    custom_domain VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(client_id, slug)
);

-- Create indexes
CREATE INDEX idx_shops_client_id ON shops(client_id);
CREATE INDEX idx_shops_user_id ON shops(user_id);
CREATE INDEX idx_shops_client_slug ON shops(client_id, slug);
CREATE INDEX idx_shops_is_active ON shops(is_active);

-- Create trigger to update updated_at
CREATE TRIGGER update_shops_updated_at BEFORE UPDATE ON shops
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_shops_updated_at ON shops;
DROP TABLE IF EXISTS shops CASCADE;
-- +goose StatementEnd
