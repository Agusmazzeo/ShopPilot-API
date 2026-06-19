-- +goose Up
-- +goose StatementBegin
-- Create stores table
CREATE TABLE stores (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    logo_url TEXT,
    theme VARCHAR(50) DEFAULT 'default',
    custom_domain VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_stores_user_id ON stores(user_id);
CREATE INDEX idx_stores_slug ON stores(slug);
CREATE INDEX idx_stores_is_active ON stores(is_active);

-- Create trigger to update updated_at
CREATE TRIGGER update_stores_updated_at BEFORE UPDATE ON stores
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_stores_updated_at ON stores;
DROP TABLE IF EXISTS stores CASCADE;
-- +goose StatementEnd
