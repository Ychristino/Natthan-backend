CREATE TABLE stock_movements (
    id               UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id       UUID                NOT NULL REFERENCES products (id),
    type             stock_movement_type NOT NULL,
    quantity         INTEGER             NOT NULL,
    service_order_id UUID                REFERENCES service_orders (id),
    created_by       UUID                NOT NULL REFERENCES users (id),
    created_at       TIMESTAMPTZ         NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_stock_movements_quantity_positive CHECK (quantity > 0)
);

CREATE INDEX idx_stock_movements_product_id       ON stock_movements (product_id);
CREATE INDEX idx_stock_movements_service_order_id ON stock_movements (service_order_id);
CREATE INDEX idx_stock_movements_created_by       ON stock_movements (created_by);
CREATE INDEX idx_stock_movements_created_at       ON stock_movements (created_at DESC);
