CREATE TABLE service_order_products (
    id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    service_order_id UUID           NOT NULL REFERENCES service_orders (id) ON DELETE CASCADE,
    product_id       UUID           NOT NULL REFERENCES products (id),
    quantity         INTEGER        NOT NULL,
    unit_price       NUMERIC(10, 2) NOT NULL,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_sop_quantity_positive       CHECK (quantity > 0),
    CONSTRAINT chk_sop_unit_price_non_negative CHECK (unit_price >= 0)
);

CREATE INDEX idx_sop_service_order_id ON service_order_products (service_order_id);
CREATE INDEX idx_sop_product_id       ON service_order_products (product_id);
