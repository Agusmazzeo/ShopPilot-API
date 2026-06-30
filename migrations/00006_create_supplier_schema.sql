-- +goose Up
-- +goose StatementBegin

-- Suppliers table (partitioned by client_id)
CREATE TABLE suppliers (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    tax_id VARCHAR(100),
    payment_terms VARCHAR(100),
    currency VARCHAR(3) DEFAULT 'USD',
    notes TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    UNIQUE(client_id, code)
) PARTITION BY HASH (client_id);

CREATE TABLE suppliers_p0 PARTITION OF suppliers FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE suppliers_p1 PARTITION OF suppliers FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE suppliers_p2 PARTITION OF suppliers FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE suppliers_p3 PARTITION OF suppliers FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE suppliers_p4 PARTITION OF suppliers FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE suppliers_p5 PARTITION OF suppliers FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE suppliers_p6 PARTITION OF suppliers FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE suppliers_p7 PARTITION OF suppliers FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_suppliers_client ON suppliers(client_id);
CREATE INDEX idx_suppliers_code ON suppliers(client_id, code);
CREATE INDEX idx_suppliers_name ON suppliers(name);
CREATE INDEX idx_suppliers_is_active ON suppliers(is_active);

CREATE TRIGGER update_suppliers_updated_at BEFORE UPDATE ON suppliers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS suppliers CASCADE;
-- +goose StatementEnd
