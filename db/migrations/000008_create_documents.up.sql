CREATE TABLE documents (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id  UUID          NOT NULL REFERENCES persons (id) ON DELETE CASCADE,
    type       document_type NOT NULL,
    value      VARCHAR(50)   NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    UNIQUE (type, value)
);

CREATE INDEX idx_documents_person_id ON documents (person_id);
CREATE INDEX idx_documents_value     ON documents (value);
