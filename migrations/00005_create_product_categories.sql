-- +goose Up
-- +goose StatementBegin
-- Create product_categories table
CREATE TABLE product_categories (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id INTEGER REFERENCES product_categories(id) ON DELETE SET NULL,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(store_id, slug)
);

-- Create indexes
CREATE INDEX idx_product_categories_store_id ON product_categories(store_id);
CREATE INDEX idx_product_categories_parent_id ON product_categories(parent_id);
CREATE INDEX idx_product_categories_is_active ON product_categories(is_active);

-- Create trigger to update updated_at
CREATE TRIGGER update_product_categories_updated_at BEFORE UPDATE ON product_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_product_categories_updated_at ON product_categories;
DROP TABLE IF EXISTS product_categories CASCADE;
-- +goose StatementEnd
