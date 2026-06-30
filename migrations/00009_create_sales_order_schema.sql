-- +goose Up
-- +goose StatementBegin

-- Sales order status enum
CREATE TYPE sales_order_status AS ENUM (
    'pending',
    'confirmed',
    'processing',
    'partially_fulfilled',
    'fulfilled',
    'cancelled'
);

-- Sales orders table (partitioned by client_id)
CREATE TABLE sales_orders (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL,
    shop_id UUID NOT NULL,
    order_number VARCHAR(100) NOT NULL,
    status sales_order_status DEFAULT 'pending',
    order_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    shipping_date TIMESTAMP WITH TIME ZONE,
    delivery_date TIMESTAMP WITH TIME ZONE,
    subtotal DECIMAL(15, 2) DEFAULT 0.00,
    tax_amount DECIMAL(15, 2) DEFAULT 0.00,
    shipping_amount DECIMAL(15, 2) DEFAULT 0.00,
    total_amount DECIMAL(15, 2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    shipping_address TEXT,
    billing_address TEXT,
    notes TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    FOREIGN KEY (client_id, customer_id) REFERENCES customers(client_id, id) ON DELETE RESTRICT,
    FOREIGN KEY (client_id, shop_id) REFERENCES shops(client_id, id) ON DELETE CASCADE,
    UNIQUE(client_id, order_number)
) PARTITION BY HASH (client_id);

CREATE TABLE sales_orders_p0 PARTITION OF sales_orders FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE sales_orders_p1 PARTITION OF sales_orders FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE sales_orders_p2 PARTITION OF sales_orders FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE sales_orders_p3 PARTITION OF sales_orders FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE sales_orders_p4 PARTITION OF sales_orders FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE sales_orders_p5 PARTITION OF sales_orders FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE sales_orders_p6 PARTITION OF sales_orders FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE sales_orders_p7 PARTITION OF sales_orders FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_so_client ON sales_orders(client_id);
CREATE INDEX idx_so_customer ON sales_orders(client_id, customer_id);
CREATE INDEX idx_so_shop ON sales_orders(client_id, shop_id);
CREATE INDEX idx_so_status ON sales_orders(status);
CREATE INDEX idx_so_order_date ON sales_orders(order_date);

CREATE TRIGGER update_sales_orders_updated_at BEFORE UPDATE ON sales_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sales order line items
CREATE TABLE sales_order_items (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    sales_order_id UUID NOT NULL,
    variant_id UUID NOT NULL,
    quantity_ordered INTEGER NOT NULL,
    quantity_fulfilled INTEGER DEFAULT 0,
    unit_price DECIMAL(10, 2) NOT NULL,
    tax_rate DECIMAL(5, 4) DEFAULT 0.0000,
    discount_amount DECIMAL(10, 2) DEFAULT 0.00,
    total_price DECIMAL(15, 2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    FOREIGN KEY (client_id, sales_order_id) REFERENCES sales_orders(client_id, id) ON DELETE CASCADE,
    FOREIGN KEY (client_id, variant_id) REFERENCES product_variants(client_id, id) ON DELETE RESTRICT,
    CONSTRAINT quantity_check CHECK (quantity_fulfilled <= quantity_ordered)
) PARTITION BY HASH (client_id);

CREATE TABLE sales_order_items_p0 PARTITION OF sales_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE sales_order_items_p1 PARTITION OF sales_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE sales_order_items_p2 PARTITION OF sales_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE sales_order_items_p3 PARTITION OF sales_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE sales_order_items_p4 PARTITION OF sales_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE sales_order_items_p5 PARTITION OF sales_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE sales_order_items_p6 PARTITION OF sales_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE sales_order_items_p7 PARTITION OF sales_order_items FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_so_items_so ON sales_order_items(client_id, sales_order_id);
CREATE INDEX idx_so_items_variant ON sales_order_items(client_id, variant_id);

CREATE TRIGGER update_sales_order_items_updated_at BEFORE UPDATE ON sales_order_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sales_order_items CASCADE;
DROP TABLE IF EXISTS sales_orders CASCADE;
DROP TYPE IF EXISTS sales_order_status;
-- +goose StatementEnd
