CREATE TABLE user_roles (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role            user_role   NOT NULL,
    expiration_date TIMESTAMPTZ,
    active          BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, role)
);

CREATE INDEX idx_user_roles_user_id ON user_roles (user_id);
CREATE INDEX idx_user_roles_role    ON user_roles (role);
CREATE INDEX idx_user_roles_active  ON user_roles (active);
