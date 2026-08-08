CREATE TABLE states (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    code       CHAR(2)      NOT NULL,
    country_id UUID         NOT NULL REFERENCES countries (id),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (code, country_id)
);

CREATE INDEX idx_states_country_id ON states (country_id);
