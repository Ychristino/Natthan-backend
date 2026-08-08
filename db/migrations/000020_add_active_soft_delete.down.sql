DROP INDEX IF EXISTS idx_persons_active;
ALTER TABLE persons  DROP COLUMN IF EXISTS active;

DROP INDEX IF EXISTS idx_products_active;
ALTER TABLE products DROP COLUMN IF EXISTS active;
