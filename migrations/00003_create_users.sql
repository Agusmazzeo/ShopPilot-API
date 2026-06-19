-- +goose Up
-- +goose StatementBegin
-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) DEFAULT '',
    last_name VARCHAR(100) DEFAULT '',
    phone VARCHAR(20) DEFAULT '',
    avatar_url TEXT,
    user_status_id INTEGER NOT NULL REFERENCES user_status(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_user_status_id ON users(user_status_id);

-- Create trigger to update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user
-- Password: 'admin123' (bcrypt hash with cost 10)
-- IMPORTANT: Change this password after first login!
INSERT INTO users (email, username, password, first_name, last_name, user_status_id) VALUES
    ('admin@shoppilot.com', 'admin', '$2a$10$xG2zJuLNETRmz/MvMGJ5ce7RkZVMRbNoCLeyUKGgnaVf3beNu0e5K', 'Admin', 'User', 1);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
