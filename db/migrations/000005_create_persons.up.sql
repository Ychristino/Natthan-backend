CREATE TABLE persons (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name   VARCHAR(255) NOT NULL,
    birth_date  DATE,
    nationality VARCHAR(100),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_persons_full_name ON persons USING gin (full_name gin_trgm_ops);
