ALTER TABLE persons
    DROP COLUMN nationality,
    ADD COLUMN nationality_id UUID REFERENCES countries (id) ON DELETE SET NULL;

CREATE INDEX idx_persons_nationality_id ON persons (nationality_id);
