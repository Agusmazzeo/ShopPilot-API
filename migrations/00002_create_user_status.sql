-- +goose Up
-- +goose StatementBegin
-- Create user_status lookup table
CREATE TABLE user_status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for performance
CREATE INDEX idx_user_status_name ON user_status(name);

-- Insert default user statuses
INSERT INTO user_status (name, description, is_active) VALUES
    ('ACTIVE', 'User is active and can access the system', true),
    ('INACTIVE', 'User is inactive and cannot access the system', true),
    ('INVITED', 'User has been invited but has not yet activated their account', true),
    ('SUSPENDED', 'User account has been suspended', true);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS user_status CASCADE;
-- +goose StatementEnd
