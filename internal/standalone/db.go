package standalone

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenDB opens (or creates) the SQLite database at dataDir/natthan.db,
// enables WAL mode + foreign keys, and runs the embedded schema.
func OpenDB(dataDir string) (*sql.DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("criar diretório de dados: %w", err)
	}

	dbPath := filepath.Join(dataDir, "natthan.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite has one writer; WAL allows concurrent reads

	if err := applySchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("aplicar schema: %w", err)
	}

	return db, nil
}

func applySchema(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

const schema = `
CREATE TABLE IF NOT EXISTS countries (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    code       TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE TABLE IF NOT EXISTS states (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    code       TEXT NOT NULL,
    country_id TEXT NOT NULL REFERENCES countries(id),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    UNIQUE (code, country_id)
);
CREATE INDEX IF NOT EXISTS idx_states_country_id ON states (country_id);

CREATE TABLE IF NOT EXISTS cities (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    state_id   TEXT NOT NULL REFERENCES states(id),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_cities_state_id ON cities (state_id);
CREATE INDEX IF NOT EXISTS idx_cities_name     ON cities (name);

CREATE TABLE IF NOT EXISTS persons (
    id             TEXT PRIMARY KEY,
    full_name      TEXT NOT NULL,
    birth_date     TEXT,
    nationality_id TEXT REFERENCES countries(id),
    active         INTEGER NOT NULL DEFAULT 1,
    created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_persons_full_name ON persons (full_name);
CREATE INDEX IF NOT EXISTS idx_persons_active    ON persons (active);

CREATE TABLE IF NOT EXISTS addresses (
    id           TEXT PRIMARY KEY,
    person_id    TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    address_type TEXT NOT NULL,
    street       TEXT NOT NULL,
    number       TEXT NOT NULL,
    complement   TEXT,
    neighborhood TEXT NOT NULL,
    zip_code     TEXT NOT NULL,
    city_id      TEXT NOT NULL REFERENCES cities(id),
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_addresses_person_id ON addresses (person_id);

CREATE TABLE IF NOT EXISTS contacts (
    id         TEXT PRIMARY KEY,
    person_id  TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    value      TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_contacts_person_id ON contacts (person_id);

CREATE TABLE IF NOT EXISTS documents (
    id         TEXT PRIMARY KEY,
    person_id  TEXT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    value      TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    UNIQUE (type, value)
);
CREATE INDEX IF NOT EXISTS idx_documents_person_id ON documents (person_id);

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    person_id     TEXT NOT NULL UNIQUE REFERENCES persons(id),
    username      TEXT NOT NULL UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    active        INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at    TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_users_email  ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users (active);

CREATE TABLE IF NOT EXISTS user_roles (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES users(id),
    role            TEXT NOT NULL,
    expiration_date TEXT,
    active          INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    UNIQUE (user_id, role)
);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id);

CREATE TABLE IF NOT EXISTS products (
    id             TEXT PRIMARY KEY,
    product_code   TEXT NOT NULL UNIQUE,
    name           TEXT NOT NULL,
    description    TEXT,
    purchase_price REAL NOT NULL DEFAULT 0,
    sale_price     REAL NOT NULL DEFAULT 0,
    active         INTEGER NOT NULL DEFAULT 1,
    created_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_products_name   ON products (name);
CREATE INDEX IF NOT EXISTS idx_products_active ON products (active);

CREATE TABLE IF NOT EXISTS services (
    id           TEXT PRIMARY KEY,
    service_code TEXT NOT NULL UNIQUE,
    name         TEXT NOT NULL,
    description  TEXT,
    price        REAL NOT NULL DEFAULT 0,
    active       INTEGER NOT NULL DEFAULT 1,
    created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_services_name   ON services (name);
CREATE INDEX IF NOT EXISTS idx_services_active ON services (active);

CREATE TABLE IF NOT EXISTS stock (
    id         TEXT PRIMARY KEY,
    product_id TEXT NOT NULL UNIQUE REFERENCES products(id),
    quantity   INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

CREATE TABLE IF NOT EXISTS service_orders (
    id          TEXT PRIMARY KEY,
    client_id   TEXT NOT NULL REFERENCES persons(id),
    employee_id TEXT NOT NULL REFERENCES persons(id),
    status      TEXT NOT NULL DEFAULT 'open',
    entry_date  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    exit_date   TEXT,
    pay_date    TEXT,
    is_paid     INTEGER NOT NULL DEFAULT 0,
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_service_orders_client_id   ON service_orders (client_id);
CREATE INDEX IF NOT EXISTS idx_service_orders_employee_id ON service_orders (employee_id);
CREATE INDEX IF NOT EXISTS idx_service_orders_status      ON service_orders (status);
CREATE INDEX IF NOT EXISTS idx_service_orders_entry_date  ON service_orders (entry_date DESC);

CREATE TABLE IF NOT EXISTS service_order_products (
    id               TEXT PRIMARY KEY,
    service_order_id TEXT NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
    product_id       TEXT NOT NULL REFERENCES products(id),
    quantity         INTEGER NOT NULL,
    unit_price       REAL    NOT NULL DEFAULT 0,
    created_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_sop_service_order_id ON service_order_products (service_order_id);
CREATE INDEX IF NOT EXISTS idx_sop_product_id       ON service_order_products (product_id);

CREATE TABLE IF NOT EXISTS service_order_services (
    id               TEXT PRIMARY KEY,
    service_order_id TEXT NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
    service_id       TEXT NOT NULL REFERENCES services(id),
    price_applied    REAL NOT NULL DEFAULT 0,
    notes            TEXT NOT NULL DEFAULT '',
    created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_sos_service_order_id ON service_order_services (service_order_id);
CREATE INDEX IF NOT EXISTS idx_sos_service_id       ON service_order_services (service_id);

CREATE TABLE IF NOT EXISTS stock_movements (
    id               TEXT PRIMARY KEY,
    product_id       TEXT NOT NULL REFERENCES products(id),
    type             TEXT NOT NULL,
    quantity         INTEGER NOT NULL,
    service_order_id TEXT,
    created_by       TEXT NOT NULL REFERENCES users(id),
    created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_stock_movements_product_id ON stock_movements (product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements (created_at DESC);
`
