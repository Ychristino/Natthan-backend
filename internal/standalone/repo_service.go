package standalone

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/service"
)

type sqliteServiceRepo struct{ db *sql.DB }

func NewServiceRepository(db *sql.DB) service.Repository {
	return &sqliteServiceRepo{db: db}
}

const serviceCols = `id, service_code, name, description, price, active, created_at, updated_at`

func (r *sqliteServiceRepo) Create(ctx context.Context, s service.Service) (*service.Service, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO services (id, service_code, name, description, price, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 1, ?, ?)`,
		s.ID.String(), s.ServiceCode, s.Name, nullStr(s.Description), s.Price, now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, s.ID)
}

func (r *sqliteServiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*service.Service, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+serviceCols+` FROM services WHERE id = ? AND active = 1`, id.String())
	return scanService(row)
}

func (r *sqliteServiceRepo) List(ctx context.Context, name string, limit, offset int) ([]service.Service, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+serviceCols+`, COUNT(*) OVER () AS total
		FROM services
		WHERE active = 1
		  AND (? = '' OR LOWER(name) LIKE LOWER('%' || ? || '%'))
		ORDER BY name
		LIMIT ? OFFSET ?
	`, name, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var services []service.Service
	var total int64
	for rows.Next() {
		var svc service.Service
		var id string
		var desc sql.NullString
		var activeVal int
		var cat, uat string
		var t int64
		if err := rows.Scan(&id, &svc.ServiceCode, &svc.Name, &desc, &svc.Price, &activeVal, &cat, &uat, &t); err != nil {
			return nil, 0, err
		}
		svc.ID, _ = uuid.Parse(id)
		svc.Description = desc.String
		svc.Active = activeVal == 1
		svc.CreatedAt = parseTime(cat)
		svc.UpdatedAt = parseTime(uat)
		total = t
		services = append(services, svc)
	}
	return services, total, rows.Err()
}

func (r *sqliteServiceRepo) Update(ctx context.Context, s service.Service) (*service.Service, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE services
		SET service_code = ?, name = ?, description = ?, price = ?, updated_at = ?
		WHERE id = ?`,
		s.ServiceCode, s.Name, nullStr(s.Description), s.Price, nowStr(), s.ID.String())
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, service.ErrNotFound
	}
	return r.GetByID(ctx, s.ID)
}

func (r *sqliteServiceRepo) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE services SET active = 0, updated_at = ? WHERE id = ?`, nowStr(), id.String())
	return err
}

func scanService(s rowScanner) (*service.Service, error) {
	var svc service.Service
	var id string
	var desc sql.NullString
	var activeVal int
	var cat, uat string
	err := s.Scan(&id, &svc.ServiceCode, &svc.Name, &desc, &svc.Price, &activeVal, &cat, &uat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrNotFound
		}
		return nil, err
	}
	svc.ID, _ = uuid.Parse(id)
	svc.Description = desc.String
	svc.Active = activeVal == 1
	svc.CreatedAt = parseTime(cat)
	svc.UpdatedAt = parseTime(uat)
	return &svc, nil
}
