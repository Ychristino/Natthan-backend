CREATE TABLE products (
    id             UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    product_code   VARCHAR(100)   NOT NULL UNIQUE,
    name           VARCHAR(255)   NOT NULL,
    description    TEXT,
    purchase_price NUMERIC(10, 2) NOT NULL DEFAULT 0,
    sale_price     NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_products_purchase_price_non_negative CHECK (purchase_price >= 0),
    CONSTRAINT chk_products_sale_price_non_negative     CHECK (sale_price >= 0)
);

CREATE INDEX idx_products_name ON products USING gin (name gin_trgm_ops);
