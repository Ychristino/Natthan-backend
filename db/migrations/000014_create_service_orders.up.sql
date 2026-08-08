CREATE TABLE service_orders (
    id          UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id   UUID                 NOT NULL REFERENCES persons (id),
    employee_id UUID                 NOT NULL REFERENCES persons (id),
    status      service_order_status NOT NULL DEFAULT 'open',
    entry_date  TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
    exit_date   TIMESTAMPTZ,
    notes       TEXT,
    created_at  TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ          NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_service_orders_client_id        ON service_orders (client_id);
CREATE INDEX idx_service_orders_employee_id      ON service_orders (employee_id);
CREATE INDEX idx_service_orders_status_entry_date ON service_orders (status, entry_date DESC);
