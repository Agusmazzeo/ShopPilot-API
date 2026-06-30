-- +goose Up
-- +goose StatementBegin

-- Purchase order status enum
CREATE TYPE purchase_order_status AS ENUM (
    'draft',
    'submitted',
    'partially_received',
    'received',
    'cancelled'
);

-- Purchase orders table (partitioned by client_id)
CREATE TABLE purchase_orders (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL,
    shop_id UUID NOT NULL,
    po_number VARCHAR(100) NOT NULL,
    status purchase_order_status DEFAULT 'draft',
    order_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expected_delivery_date TIMESTAMP WITH TIME ZONE,
    received_date TIMESTAMP WITH TIME ZONE,
    total_amount DECIMAL(15, 2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    notes TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    FOREIGN KEY (client_id, supplier_id) REFERENCES suppliers(client_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (client_id, shop_id) REFERENCES shops(client_id, id) ON DELETE CASCADE,
    UNIQUE(client_id, po_number)
) PARTITION BY HASH (client_id);

CREATE TABLE purchase_orders_p0 PARTITION OF purchase_orders FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE purchase_orders_p1 PARTITION OF purchase_orders FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE purchase_orders_p2 PARTITION OF purchase_orders FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE purchase_orders_p3 PARTITION OF purchase_orders FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE purchase_orders_p4 PARTITION OF purchase_orders FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE purchase_orders_p5 PARTITION OF purchase_orders FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE purchase_orders_p6 PARTITION OF purchase_orders FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE purchase_orders_p7 PARTITION OF purchase_orders FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_po_client ON purchase_orders(client_id);
CREATE INDEX idx_po_supplier ON purchase_orders(client_id, supplier_id);
CREATE INDEX idx_po_shop ON purchase_orders(client_id, shop_id);
CREATE INDEX idx_po_status ON purchase_orders(status);
CREATE INDEX idx_po_order_date ON purchase_orders(order_date);

CREATE TRIGGER update_purchase_orders_updated_at BEFORE UPDATE ON purchase_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Purchase order line items
CREATE TABLE purchase_order_items (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    purchase_order_id UUID NOT NULL,
    variant_id UUID NOT NULL,
    quantity_ordered INTEGER NOT NULL,
    quantity_received INTEGER DEFAULT 0,
    unit_cost DECIMAL(10, 2) NOT NULL,
    total_cost DECIMAL(15, 2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    FOREIGN KEY (client_id, purchase_order_id) REFERENCES purchase_orders(client_id, id) ON DELETE CASCADE,
    FOREIGN KEY (client_id, variant_id) REFERENCES product_variants(client_id, id) ON DELETE RESTRICT,
    CONSTRAINT quantity_check CHECK (quantity_received <= quantity_ordered)
) PARTITION BY HASH (client_id);

CREATE TABLE purchase_order_items_p0 PARTITION OF purchase_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE purchase_order_items_p1 PARTITION OF purchase_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE purchase_order_items_p2 PARTITION OF purchase_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE purchase_order_items_p3 PARTITION OF purchase_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE purchase_order_items_p4 PARTITION OF purchase_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE purchase_order_items_p5 PARTITION OF purchase_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE purchase_order_items_p6 PARTITION OF purchase_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE purchase_order_items_p7 PARTITION OF purchase_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_po_items_po ON purchase_order_items(client_id, purchase_order_id);
CREATE INDEX idx_po_items_variant ON purchase_order_items(client_id, variant_id);

CREATE TRIGGER update_purchase_order_items_updated_at BEFORE UPDATE ON purchase_order_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS purchase_order_items CASCADE;
DROP TABLE IF EXISTS purchase_orders CASCADE;
DROP TYPE IF EXISTS purchase_order_status;
-- +goose StatementEnd
