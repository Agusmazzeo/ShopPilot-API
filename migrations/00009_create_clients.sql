-- +goose Up
-- +goose StatementBegin

-- Create clients table (tenant root)
CREATE TABLE clients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(20),
    subscription_tier VARCHAR(50) DEFAULT 'free',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_clients_slug ON clients(slug);
CREATE INDEX idx_clients_is_active ON clients(is_active);

-- Create trigger to update updated_at
CREATE TRIGGER update_clients_updated_at BEFORE UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default client for existing data
INSERT INTO clients (name, slug, contact_email, subscription_tier, is_active)
VALUES ('Default Client', 'default', 'admin@shoppilot.com', 'enterprise', true);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;
DROP TABLE IF EXISTS clients CASCADE;
-- +goose StatementEnd
