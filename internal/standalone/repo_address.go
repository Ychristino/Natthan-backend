package standalone

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/address"
)

type sqliteAddressRepo struct{ db *sql.DB }

func NewAddressRepository(db *sql.DB) address.Repository {
	return &sqliteAddressRepo{db: db}
}

const addressDetailSelect = `
	SELECT
		a.id, a.person_id, a.address_type, a.street, a.number, a.complement,
		a.neighborhood, a.zip_code, a.created_at, a.updated_at,
		ci.id, ci.name,
		s.id, s.name, s.code,
		co.id, co.name, co.code
	FROM addresses a
	JOIN cities   ci ON ci.id = a.city_id
	JOIN states    s ON s.id  = ci.state_id
	JOIN countries co ON co.id = s.country_id`

func (r *sqliteAddressRepo) Create(ctx context.Context, a address.Address) (*address.AddressDetail, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO addresses (id, person_id, address_type, street, number, complement, neighborhood, zip_code, city_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID.String(), a.PersonID.String(), a.AddressType, a.Street, a.Number,
		nullStr(a.Complement), a.Neighborhood, a.ZipCode, a.CityID.String(), now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, a.ID)
}

func (r *sqliteAddressRepo) GetByID(ctx context.Context, id uuid.UUID) (*address.AddressDetail, error) {
	row := r.db.QueryRowContext(ctx, addressDetailSelect+` WHERE a.id = ?`, id.String())
	return scanAddressDetail(row)
}

func (r *sqliteAddressRepo) GetByPersonID(ctx context.Context, personID uuid.UUID) ([]address.AddressDetail, error) {
	rows, err := r.db.QueryContext(ctx, addressDetailSelect+` WHERE a.person_id = ? ORDER BY a.created_at`, personID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []address.AddressDetail
	for rows.Next() {
		d, err := scanAddressDetail(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *d)
	}
	return result, rows.Err()
}

func (r *sqliteAddressRepo) List(ctx context.Context, personID *uuid.UUID, limit, offset int) ([]address.AddressDetail, int64, error) {
	pid := ""
	if personID != nil {
		pid = personID.String()
	}

	var total int64
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM addresses WHERE (? = '' OR person_id = ?)
	`, pid, pid).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, addressDetailSelect+`
		WHERE (? = '' OR a.person_id = ?)
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`, pid, pid, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []address.AddressDetail
	for rows.Next() {
		d, err := scanAddressDetail(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *d)
	}
	return result, total, rows.Err()
}

func (r *sqliteAddressRepo) Update(ctx context.Context, a address.Address) (*address.AddressDetail, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE addresses
		SET address_type = ?, street = ?, number = ?, complement = ?,
		    neighborhood = ?, zip_code = ?, city_id = ?, updated_at = ?
		WHERE id = ?`,
		a.AddressType, a.Street, a.Number, nullStr(a.Complement),
		a.Neighborhood, a.ZipCode, a.CityID.String(), nowStr(), a.ID.String())
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, address.ErrNotFound
	}
	return r.GetByID(ctx, a.ID)
}

func (r *sqliteAddressRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM addresses WHERE id = ?`, id.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return address.ErrNotFound
	}
	return nil
}

func scanAddressDetail(s rowScanner) (*address.AddressDetail, error) {
	var d address.AddressDetail
	var id, personID, cityID, stateID, countryID string
	var complement sql.NullString
	var cat, uat string
	err := s.Scan(
		&id, &personID, &d.AddressType, &d.Street, &d.Number, &complement,
		&d.Neighborhood, &d.ZipCode, &cat, &uat,
		&cityID, &d.CityName,
		&stateID, &d.StateName, &d.StateCode,
		&countryID, &d.CountryName, &d.CountryCode,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, address.ErrNotFound
		}
		return nil, err
	}
	d.ID, _ = uuid.Parse(id)
	d.PersonID, _ = uuid.Parse(personID)
	d.CityID, _ = uuid.Parse(cityID)
	d.StateID, _ = uuid.Parse(stateID)
	d.CountryID, _ = uuid.Parse(countryID)
	d.Complement = complement.String
	d.CreatedAt = parseTime(cat)
	d.UpdatedAt = parseTime(uat)
	return &d, nil
}
