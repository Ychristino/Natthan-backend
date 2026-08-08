CREATE TABLE addresses (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id    UUID         NOT NULL REFERENCES persons (id) ON DELETE CASCADE,
    address_type address_type NOT NULL,
    street       VARCHAR(255) NOT NULL,
    number       VARCHAR(20)  NOT NULL,
    complement   VARCHAR(100),
    neighborhood VARCHAR(100) NOT NULL,
    zip_code     VARCHAR(10)  NOT NULL,
    city_id      UUID         NOT NULL REFERENCES cities (id),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_addresses_person_id ON addresses (person_id);
CREATE INDEX idx_addresses_city_id   ON addresses (city_id);
CREATE INDEX idx_addresses_zip_code  ON addresses (zip_code);
