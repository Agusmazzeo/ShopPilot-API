-- +goose Up
-- +goose StatementBegin

-- Inventory alerts table (low stock and reorder points)
CREATE TABLE inventory_alerts (
    id UUID DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    variant_id UUID NOT NULL,
    shop_id UUID NOT NULL,
    reorder_point INTEGER NOT NULL DEFAULT 10,
    reorder_quantity INTEGER NOT NULL DEFAULT 50,
    low_stock_threshold INTEGER NOT NULL DEFAULT 20,
    is_enabled BOOLEAN DEFAULT true,
    last_alert_sent_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (client_id, id),
    FOREIGN KEY (client_id, variant_id) REFERENCES product_variants(client_id, id) ON DELETE CASCADE,
    FOREIGN KEY (client_id, shop_id) REFERENCES shops(client_id, id) ON DELETE CASCADE,
    UNIQUE(client_id, variant_id, shop_id)
) PARTITION BY HASH (client_id);

CREATE TABLE inventory_alerts_p0 PARTITION OF inventory_alerts FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE inventory_alerts_p1 PARTITION OF inventory_alerts FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE inventory_alerts_p2 PARTITION OF inventory_alerts FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE inventory_alerts_p3 PARTITION OF inventory_alerts FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE inventory_alerts_p4 PARTITION OF inventory_alerts FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE inventory_alerts_p5 PARTITION OF inventory_alerts FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE inventory_alerts_p6 PARTITION OF inventory_alerts FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE inventory_alerts_p7 PARTITION OF inventory_alerts FOR VALUES WITH (MODULUS 8, REMAINDER 7);

CREATE INDEX idx_inv_alerts_variant ON inventory_alerts(client_id, variant_id);
CREATE INDEX idx_inv_alerts_shop ON inventory_alerts(client_id, shop_id);
CREATE INDEX idx_inv_alerts_enabled ON inventory_alerts(is_enabled);

CREATE TRIGGER update_inventory_alerts_updated_at BEFORE UPDATE ON inventory_alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory_alerts CASCADE;
-- +goose StatementEnd
