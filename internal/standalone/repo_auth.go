package standalone

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/auth"
)

type sqliteAuthRepo struct{ db *sql.DB }

func NewAuthRepository(db *sql.DB) auth.Repository {
	return &sqliteAuthRepo{db: db}
}

const userCols = `id, person_id, username, email, password_hash, active, created_at, updated_at`

func (r *sqliteAuthRepo) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE email = ? AND active = 1`, email)
	u, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	u.Roles, _ = r.GetActiveRoles(ctx, u.ID)
	return u, nil
}

func (r *sqliteAuthRepo) GetByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE id = ? AND active = 1`, id.String())
	u, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	u.Roles, _ = r.GetActiveRoles(ctx, u.ID)
	return u, nil
}

func (r *sqliteAuthRepo) Create(ctx context.Context, u auth.User) (*auth.User, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, person_id, username, email, password_hash, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?)`,
		u.ID.String(), u.PersonID.String(), u.Username, u.Email, u.PasswordHash, now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, u.ID)
}

func (r *sqliteAuthRepo) List(ctx context.Context, name, role string, activeOnly bool, limit, offset int) ([]auth.User, int64, error) {
	activeInt := 0
	if activeOnly {
		activeInt = 1
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.person_id, u.username, u.email, u.active, u.created_at, u.updated_at,
		       COUNT(*) OVER () AS total
		FROM users u
		WHERE u.active = ?
		  AND (? = '' OR LOWER(u.username) LIKE LOWER('%' || ? || '%') OR LOWER(u.email) LIKE LOWER('%' || ? || '%'))
		  AND (? = '' OR EXISTS(
		      SELECT 1 FROM user_roles ur
		      WHERE ur.user_id = u.id AND ur.role = ?
		        AND ur.active = 1
		        AND (ur.expiration_date IS NULL OR ur.expiration_date > ?)
		  ))
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`, activeInt, name, name, name, role, role, nowStr(), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []auth.User
	var total int64
	for rows.Next() {
		var u auth.User
		var id, personID string
		var activeVal int
		var cat, uat string
		var t int64
		if err := rows.Scan(&id, &personID, &u.Username, &u.Email, &activeVal, &cat, &uat, &t); err != nil {
			return nil, 0, err
		}
		u.ID, _ = uuid.Parse(id)
		u.PersonID, _ = uuid.Parse(personID)
		u.Active = activeVal == 1
		u.CreatedAt = parseTime(cat)
		u.UpdatedAt = parseTime(uat)
		total = t
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *sqliteAuthRepo) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET active = 0, updated_at = ? WHERE id = ?`, nowStr(), id.String())
	return err
}

func (r *sqliteAuthRepo) Reactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET active = 1, updated_at = ? WHERE id = ?`, nowStr(), id.String())
	return err
}

// ─── user_roles ───────────────────────────────────────────────────────────────

func (r *sqliteAuthRepo) AddRole(ctx context.Context, ur auth.UserRole) (*auth.UserRole, error) {
	now := nowStr()
	var expDate sql.NullString
	if ur.ExpirationDate != nil {
		expDate = sql.NullString{String: ur.ExpirationDate.UTC().Format("2006-01-02T15:04:05Z"), Valid: true}
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_roles (id, user_id, role, expiration_date, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, ?, ?)
		ON CONFLICT (user_id, role) DO UPDATE
		  SET active = 1, expiration_date = excluded.expiration_date, updated_at = excluded.updated_at`,
		ur.ID.String(), ur.UserID.String(), string(ur.Role), expDate, now, now)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, role, expiration_date, active, created_at, updated_at FROM user_roles WHERE user_id = ? AND role = ?`,
		ur.UserID.String(), string(ur.Role))
	return scanUserRole(row)
}

func (r *sqliteAuthRepo) RemoveRole(ctx context.Context, userID uuid.UUID, role auth.Role) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_roles SET active = 0, updated_at = ?
		WHERE user_id = ? AND role = ?`,
		nowStr(), userID.String(), string(role))
	return err
}

func (r *sqliteAuthRepo) GetActiveRoles(ctx context.Context, userID uuid.UUID) ([]auth.UserRole, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, role, expiration_date, active, created_at, updated_at
		FROM user_roles
		WHERE user_id = ? AND active = 1
		  AND (expiration_date IS NULL OR expiration_date > ?)
		ORDER BY role
	`, userID.String(), nowStr())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRoles(rows)
}

func (r *sqliteAuthRepo) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]auth.UserRole, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, role, expiration_date, active, created_at, updated_at
		FROM user_roles WHERE user_id = ? ORDER BY role
	`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserRoles(rows)
}

// ─── scanners ─────────────────────────────────────────────────────────────────

func scanUser(s rowScanner) (*auth.User, error) {
	var u auth.User
	var id, personID, cat, uat string
	var activeVal int
	err := s.Scan(&id, &personID, &u.Username, &u.Email, &u.PasswordHash, &activeVal, &cat, &uat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrNotFound
		}
		return nil, err
	}
	u.ID, _ = uuid.Parse(id)
	u.PersonID, _ = uuid.Parse(personID)
	u.Active = activeVal == 1
	u.CreatedAt = parseTime(cat)
	u.UpdatedAt = parseTime(uat)
	return &u, nil
}

func scanUserRole(s rowScanner) (*auth.UserRole, error) {
	var ur auth.UserRole
	var id, userID, roleStr, cat, uat string
	var expDate sql.NullString
	var activeVal int
	err := s.Scan(&id, &userID, &roleStr, &expDate, &activeVal, &cat, &uat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrNotFound
		}
		return nil, err
	}
	ur.ID, _ = uuid.Parse(id)
	ur.UserID, _ = uuid.Parse(userID)
	ur.Role = auth.Role(roleStr)
	ur.Active = activeVal == 1
	ur.CreatedAt = parseTime(cat)
	ur.UpdatedAt = parseTime(uat)
	if expDate.Valid && expDate.String != "" {
		t := parseTime(expDate.String)
		ur.ExpirationDate = &t
	}
	return &ur, nil
}

func scanUserRoles(rows *sql.Rows) ([]auth.UserRole, error) {
	var result []auth.UserRole
	for rows.Next() {
		ur, err := scanUserRole(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *ur)
	}
	return result, rows.Err()
}
