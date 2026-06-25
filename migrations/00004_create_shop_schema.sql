-- +goose Up
-- +goose StatementBegin

-- Shops table (partitioned by client_id hash for scalability)
CREATE TABLE shops (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    webpage_url TEXT DEFAULT '',
    address TEXT DEFAULT '',
    city VARCHAR(100) DEFAULT '',
    state VARCHAR(100) DEFAULT '',
    country VARCHAR(100) DEFAULT '',
    postal_code VARCHAR(20) DEFAULT '',
    phone VARCHAR(20) DEFAULT '',
    email VARCHAR(255) DEFAULT '',
    logo_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    UNIQUE(client_id, slug)
) PARTITION BY HASH (client_id);

-- Create 8 partitions for shops
CREATE TABLE shops_p0 PARTITION OF shops FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE shops_p1 PARTITION OF shops FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE shops_p2 PARTITION OF shops FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE shops_p3 PARTITION OF shops FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE shops_p4 PARTITION OF shops FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE shops_p5 PARTITION OF shops FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE shops_p6 PARTITION OF shops FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE shops_p7 PARTITION OF shops FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_shops_client ON shops(client_id);
CREATE INDEX idx_shops_slug ON shops(client_id, slug);
CREATE INDEX idx_shops_is_active ON shops(is_active);

CREATE TRIGGER update_shops_updated_at BEFORE UPDATE ON shops
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Shop users (manages specific shops)
-- Users must have a client_user_role to be assigned to shops
CREATE TABLE shop_users (
    id SERIAL PRIMARY KEY,
    client_id UUID NOT NULL,
    shop_id UUID NOT NULL,
    client_user_role_id INTEGER NOT NULL REFERENCES client_user_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (client_id, shop_id) REFERENCES shops(client_id, id) ON DELETE CASCADE,
    UNIQUE(shop_id, client_user_role_id)
);

CREATE INDEX idx_shop_users_shop ON shop_users(shop_id);
CREATE INDEX idx_shop_users_client_user_role ON shop_users(client_user_role_id);

CREATE TRIGGER update_shop_users_updated_at BEFORE UPDATE ON shop_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shop_users CASCADE;
DROP TABLE IF EXISTS shops CASCADE;
-- +goose StatementEnd
