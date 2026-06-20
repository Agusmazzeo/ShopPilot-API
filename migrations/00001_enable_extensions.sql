-- +goose Up
-- +goose StatementBegin
-- Enable PostgreSQL extensions needed for e-commerce

-- UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- pg_trgm: Trigram similarity for fuzzy text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- unaccent: Remove accents for better search
CREATE EXTENSION IF NOT EXISTS unaccent;

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
-- Note: We don't drop extensions as other databases might use them
-- +goose StatementEnd
