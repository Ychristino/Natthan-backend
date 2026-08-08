CREATE TABLE stock (
    id         UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID    NOT NULL UNIQUE REFERENCES products (id),
    quantity   INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_stock_quantity_non_negative CHECK (quantity >= 0)
);
