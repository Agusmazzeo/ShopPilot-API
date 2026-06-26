-- +goose Up
-- +goose StatementBegin

-- Products table (partitioned by client_id hash for scalability)
CREATE TABLE products (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    shop_id UUID NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    metadata JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    FOREIGN KEY (client_id, shop_id) REFERENCES shops(client_id, id) ON DELETE CASCADE,
    UNIQUE(client_id, code)
) PARTITION BY HASH (client_id);

-- Create 8 partitions for products
CREATE TABLE products_p0 PARTITION OF products FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE products_p1 PARTITION OF products FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE products_p2 PARTITION OF products FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE products_p3 PARTITION OF products FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE products_p4 PARTITION OF products FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE products_p5 PARTITION OF products FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE products_p6 PARTITION OF products FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE products_p7 PARTITION OF products FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_products_shop ON products(shop_id);
CREATE INDEX idx_products_client ON products(client_id);
CREATE INDEX idx_products_code ON products(client_id, code);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_metadata ON products USING GIN (metadata);

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Product variants table (SKU and inventory management)
-- Each product must have at least one variant
CREATE TABLE product_variants (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    shop_id UUID NOT NULL,
    product_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    compare_at_price DECIMAL(10, 2),
    cost DECIMAL(10, 2),
    quantity INTEGER NOT NULL DEFAULT 0,
    weight DECIMAL(10, 3),
    weight_unit VARCHAR(10) DEFAULT 'kg',
    requires_shipping BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    attributes JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    FOREIGN KEY (client_id, product_id) REFERENCES products(client_id, id) ON DELETE CASCADE,
    FOREIGN KEY (client_id, shop_id) REFERENCES shops(client_id, id) ON DELETE CASCADE,
    UNIQUE(client_id, sku)
) PARTITION BY HASH (client_id);

-- Create 8 partitions for product_variants
CREATE TABLE product_variants_p0 PARTITION OF product_variants FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE product_variants_p1 PARTITION OF product_variants FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE product_variants_p2 PARTITION OF product_variants FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE product_variants_p3 PARTITION OF product_variants FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE product_variants_p4 PARTITION OF product_variants FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE product_variants_p5 PARTITION OF product_variants FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE product_variants_p6 PARTITION OF product_variants FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE product_variants_p7 PARTITION OF product_variants FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_product_variants_shop ON product_variants(shop_id);
CREATE INDEX idx_product_variants_product ON product_variants(client_id, product_id);
CREATE INDEX idx_product_variants_sku ON product_variants(client_id, sku);
CREATE INDEX idx_product_variants_is_active ON product_variants(is_active);
CREATE INDEX idx_product_variants_is_default ON product_variants(is_default);
CREATE INDEX idx_product_variants_attributes ON product_variants USING GIN (attributes);

CREATE TRIGGER update_product_variants_updated_at BEFORE UPDATE ON product_variants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS products CASCADE;
-- +goose StatementEnd
