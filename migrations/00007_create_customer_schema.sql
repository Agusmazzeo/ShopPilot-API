-- +goose Up
-- +goose StatementBegin

-- Customers table (partitioned by client_id)
CREATE TABLE customers (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    code VARCHAR(100) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    shipping_address TEXT,
    shipping_city VARCHAR(100),
    shipping_state VARCHAR(100),
    shipping_postal_code VARCHAR(20),
    shipping_country VARCHAR(100),
    billing_address TEXT,
    billing_city VARCHAR(100),
    billing_state VARCHAR(100),
    billing_postal_code VARCHAR(20),
    billing_country VARCHAR(100),
    tax_id VARCHAR(100),
    notes TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    UNIQUE(client_id, code)
) PARTITION BY HASH (client_id);

CREATE TABLE customers_p0 PARTITION OF customers FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE customers_p1 PARTITION OF customers FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE customers_p2 PARTITION OF customers FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE customers_p3 PARTITION OF customers FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE customers_p4 PARTITION OF customers FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE customers_p5 PARTITION OF customers FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE customers_p6 PARTITION OF customers FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE customers_p7 PARTITION OF customers FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_customers_client ON customers(client_id);
CREATE INDEX idx_customers_code ON customers(client_id, code);
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_is_active ON customers(is_active);

CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customers CASCADE;
-- +goose StatementEnd
