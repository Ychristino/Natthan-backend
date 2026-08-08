package standalone

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	personName string
	username   string
	email      string
	password   string
	role       string
}

var demoUsers = []seedUser{
	{personName: "Administrador", username: "admin",    email: "admin@admin.com", password: "123456", role: "admin"},
	{personName: "Employee NZ",   username: "employee", email: "employee@nz.com", password: "123456", role: "employee"},
	{personName: "Natthan NZ",    username: "natthan",  email: "natthan@nz.com",  password: "123456", role: "visitor"},
}

// SeedDemoUsers inserts the demo accounts if they don't exist yet.
// Safe to call on every startup — skips any email that already exists.
func SeedDemoUsers(db *sql.DB) {
	ctx := context.Background()
	for _, u := range demoUsers {
		if userExists(ctx, db, u.email) {
			continue
		}
		if err := insertDemoUser(ctx, db, u); err != nil {
			log.Printf("seed: erro ao criar %s: %v", u.email, err)
		} else {
			log.Printf("seed: usuario criado — %s (%s)", u.email, u.role)
		}
	}
}

func userExists(ctx context.Context, db *sql.DB, email string) bool {
	var n int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE email = ?`, email).Scan(&n)
	return n > 0
}

func insertDemoUser(ctx context.Context, db *sql.DB, u seedUser) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	personID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	now := nowStr()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO persons (id, full_name, active, created_at, updated_at)
		VALUES (?, ?, 1, ?, ?)`,
		personID.String(), u.personName, now, now); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO users (id, person_id, username, email, password_hash, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?)`,
		userID.String(), personID.String(), u.username, u.email, string(hash), now, now); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO user_roles (id, user_id, role, active, created_at, updated_at)
		VALUES (?, ?, ?, 1, ?, ?)`,
		roleID.String(), userID.String(), u.role, now, now); err != nil {
		return err
	}

	return tx.Commit()
}
