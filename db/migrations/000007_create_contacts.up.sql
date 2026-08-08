CREATE TABLE contacts (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id  UUID         NOT NULL REFERENCES persons (id) ON DELETE CASCADE,
    type       contact_type NOT NULL,
    value      VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contacts_person_id ON contacts (person_id);
CREATE INDEX idx_contacts_value     ON contacts (value);
