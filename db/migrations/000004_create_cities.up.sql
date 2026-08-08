CREATE TABLE cities (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    state_id   UUID         NOT NULL REFERENCES states (id),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cities_state_id ON cities (state_id);
CREATE INDEX idx_cities_name     ON cities (name);
