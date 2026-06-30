-- +goose Up
-- +goose StatementBegin

-- Inventory movement type enum
CREATE TYPE inventory_movement_type AS ENUM (
    'purchase',           -- Receiving from supplier
    'sale',              -- Sale to customer
    'adjustment',        -- Manual adjustment (count, correction)
    'return_from_customer', -- Customer return
    'return_to_supplier',  -- Return to supplier
    'damaged',           -- Damaged/lost inventory
    'transfer'           -- Transfer between locations
);

-- Inventory movements table (audit trail for all inventory changes)
CREATE TABLE inventory_movements (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    variant_id UUID NOT NULL,
    shop_id UUID NOT NULL,
    movement_type inventory_movement_type NOT NULL,
    quantity INTEGER NOT NULL,
    previous_quantity INTEGER NOT NULL,
    new_quantity INTEGER NOT NULL,
    reference_type VARCHAR(50),  -- 'purchase_order', 'sales_order', 'manual', etc.
    reference_id UUID,            -- ID of source document
    notes TEXT,
    performed_by UUID,            -- client_user_id or platform_user_id
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    FOREIGN KEY (client_id, variant_id) REFERENCES product_variants(client_id, id) ON DELETE CASCADE,
    FOREIGN KEY (client_id, shop_id) REFERENCES shops(client_id, id) ON DELETE CASCADE
) PARTITION BY HASH (client_id);

CREATE TABLE inventory_movements_p0 PARTITION OF inventory_movements FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE inventory_movements_p1 PARTITION OF inventory_movements FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE inventory_movements_p2 PARTITION OF inventory_movements FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE inventory_movements_p3 PARTITION OF inventory_movements FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE inventory_movements_p4 PARTITION OF inventory_movements FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE inventory_movements_p5 PARTITION OF inventory_movements FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE inventory_movements_p6 PARTITION OF inventory_movements FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE inventory_movements_p7 PARTITION OF inventory_movements FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_inv_movements_variant ON inventory_movements(client_id, variant_id);
CREATE INDEX idx_inv_movements_shop ON inventory_movements(client_id, shop_id);
CREATE INDEX idx_inv_movements_type ON inventory_movements(movement_type);
CREATE INDEX idx_inv_movements_reference ON inventory_movements(reference_type, reference_id);
CREATE INDEX idx_inv_movements_created ON inventory_movements(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory_movements CASCADE;
DROP TYPE IF EXISTS inventory_movement_type;
-- +goose StatementEnd
