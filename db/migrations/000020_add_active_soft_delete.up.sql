ALTER TABLE persons  ADD COLUMN active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE products ADD COLUMN active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX idx_persons_active  ON persons  (active);
CREATE INDEX idx_products_active ON products (active);
