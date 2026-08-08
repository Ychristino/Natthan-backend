package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("usuário não encontrado")

type Repository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, u User) (*User, error)
	List(ctx context.Context, name, role string, activeOnly bool, limit, offset int) ([]User, int64, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
	Reactivate(ctx context.Context, id uuid.UUID) error

	AddRole(ctx context.Context, ur UserRole) (*UserRole, error)
	RemoveRole(ctx context.Context, userID uuid.UUID, role Role) error
	GetActiveRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
	ListUserRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const userColumns = `id, person_id, username, email, password_hash, active, created_at, updated_at`
const userColumnsPublic = `id, person_id, username, email, active, created_at, updated_at`

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1 AND active = true`, email)
	u, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	u.Roles, _ = r.GetActiveRoles(ctx, u.ID)
	return u, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1 AND active = true`,
		pgtype.UUID{Bytes: id, Valid: true})
	u, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	u.Roles, _ = r.GetActiveRoles(ctx, u.ID)
	return u, nil
}

func (r *repository) Create(ctx context.Context, u User) (*User, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO users (id, person_id, username, email, password_hash, active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING `+userColumns,
		pgtype.UUID{Bytes: u.ID, Valid: true},
		pgtype.UUID{Bytes: u.PersonID, Valid: true},
		u.Username, u.Email, u.PasswordHash,
	)
	return scanUser(row)
}

func (r *repository) List(ctx context.Context, name, role string, activeOnly bool, limit, offset int) ([]User, int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+userColumnsPublic+`, COUNT(*) OVER () AS total
		FROM users u
		WHERE u.active = $1
		  AND ($2 = '' OR u.username ILIKE '%' || $2 || '%' OR u.email ILIKE '%' || $2 || '%')
		  AND ($3 = '' OR EXISTS(
		      SELECT 1 FROM user_roles ur
		      WHERE ur.user_id = u.id AND ur.role::text = $3
		        AND ur.active = true
		        AND (ur.expiration_date IS NULL OR ur.expiration_date > NOW())
		  ))
		ORDER BY u.created_at DESC
		LIMIT $4 OFFSET $5
	`, activeOnly, name, role, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []User
	var total int64
	for rows.Next() {
		var u User
		var id, personID pgtype.UUID
		var t int64
		err := rows.Scan(&id, &personID, &u.Username, &u.Email, &u.Active, &u.CreatedAt, &u.UpdatedAt, &t)
		if err != nil {
			return nil, 0, err
		}
		u.ID = uuid.UUID(id.Bytes)
		u.PersonID = uuid.UUID(personID.Bytes)
		total = t
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET active = false, updated_at = NOW() WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	return err
}

func (r *repository) Reactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET active = true, updated_at = NOW() WHERE id = $1`,
		pgtype.UUID{Bytes: id, Valid: true})
	return err
}

// ─── user_roles ───────────────────────────────────────────────────────────────

func (r *repository) AddRole(ctx context.Context, ur UserRole) (*UserRole, error) {
	var expDate *pgtype.Timestamptz
	if ur.ExpirationDate != nil {
		v := pgtype.Timestamptz{Time: *ur.ExpirationDate, Valid: true}
		expDate = &v
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO user_roles (id, user_id, role, expiration_date, active)
		VALUES ($1, $2, $3, $4, true)
		ON CONFLICT (user_id, role) DO UPDATE
		  SET active = true, expiration_date = EXCLUDED.expiration_date, updated_at = NOW()
		RETURNING id, user_id, role, expiration_date, active, created_at, updated_at
	`,
		pgtype.UUID{Bytes: ur.ID, Valid: true},
		pgtype.UUID{Bytes: ur.UserID, Valid: true},
		ur.Role, expDate,
	)
	return scanUserRole(row)
}

func (r *repository) RemoveRole(ctx context.Context, userID uuid.UUID, role Role) error {
	_, err := r.db.Exec(ctx, `
		UPDATE user_roles SET active = false, updated_at = NOW()
		WHERE user_id = $1 AND role = $2
	`, pgtype.UUID{Bytes: userID, Valid: true}, role)
	return err
}

func (r *repository) GetActiveRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, role, expiration_date, active, created_at, updated_at
		FROM user_roles
		WHERE user_id = $1 AND active = true
		  AND (expiration_date IS NULL OR expiration_date > NOW())
		ORDER BY role
	`, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRoles(rows)
}

func (r *repository) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, role, expiration_date, active, created_at, updated_at
		FROM user_roles WHERE user_id = $1 ORDER BY role
	`, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRoles(rows)
}

// ─── scanners ─────────────────────────────────────────────────────────────────

func scanUser(row pgx.Row) (*User, error) {
	var u User
	var id, personID pgtype.UUID
	err := row.Scan(&id, &personID, &u.Username, &u.Email, &u.PasswordHash, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	u.ID = uuid.UUID(id.Bytes)
	u.PersonID = uuid.UUID(personID.Bytes)
	return &u, nil
}

func scanUserRole(row pgx.Row) (*UserRole, error) {
	var ur UserRole
	var id, userID pgtype.UUID
	var expDate pgtype.Timestamptz
	var roleStr string
	err := row.Scan(&id, &userID, &roleStr, &expDate, &ur.Active, &ur.CreatedAt, &ur.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	ur.ID = uuid.UUID(id.Bytes)
	ur.UserID = uuid.UUID(userID.Bytes)
	ur.Role = Role(roleStr)
	if expDate.Valid {
		t := expDate.Time
		ur.ExpirationDate = &t
	}
	return &ur, nil
}

func scanUserRoles(rows pgx.Rows) ([]UserRole, error) {
	var result []UserRole
	for rows.Next() {
		var ur UserRole
		var id, userID pgtype.UUID
		var expDate pgtype.Timestamptz
		var roleStr string
		err := rows.Scan(&id, &userID, &roleStr, &expDate, &ur.Active, &ur.CreatedAt, &ur.UpdatedAt)
		if err != nil {
			return nil, err
		}
		ur.ID = uuid.UUID(id.Bytes)
		ur.UserID = uuid.UUID(userID.Bytes)
		ur.Role = Role(roleStr)
		if expDate.Valid {
			t := expDate.Time
			ur.ExpirationDate = &t
		}
		result = append(result, ur)
	}
	return result, rows.Err()
}
