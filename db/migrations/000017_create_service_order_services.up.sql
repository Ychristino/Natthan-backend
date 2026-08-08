CREATE TABLE service_order_services (
    id               UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    service_order_id UUID           NOT NULL REFERENCES service_orders (id) ON DELETE CASCADE,
    service_id       UUID           NOT NULL REFERENCES services (id),
    price_applied    NUMERIC(10, 2) NOT NULL,
    notes            TEXT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_sos_price_non_negative CHECK (price_applied >= 0)
);

CREATE INDEX idx_sos_service_order_id ON service_order_services (service_order_id);
CREATE INDEX idx_sos_service_id       ON service_order_services (service_id);
