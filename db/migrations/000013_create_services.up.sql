CREATE TABLE services (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    service_code VARCHAR(100)   NOT NULL UNIQUE,
    name         VARCHAR(255)   NOT NULL,
    description  TEXT,
    price        NUMERIC(10, 2) NOT NULL DEFAULT 0,
    active       BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_services_price_non_negative CHECK (price >= 0)
);

CREATE INDEX idx_services_active ON services (active);
CREATE INDEX idx_services_name   ON services USING gin (name gin_trgm_ops);
