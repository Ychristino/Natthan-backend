package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("produto não encontrado")

type Repository interface {
	Create(ctx context.Context, p Product) (*Product, error)
	List(ctx context.Context, name string, limit, offset int) ([]Product, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	Update(ctx context.Context, p Product) (*Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

const productColumns = `id, product_code, name, description, purchase_price, sale_price, created_at, updated_at`

func (r *repository) Create(ctx context.Context, p Product) (*Product, error) {
	var desc pgtype.Text
	if p.Description != "" {
		desc = pgtype.Text{String: p.Description, Valid: true}
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO products (id, product_code, name, description, purchase_price, sale_price)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING `+productColumns,
		pgtype.UUID{Bytes: p.ID, Valid: true},
		p.ProductCode, p.Name, desc, p.PurchasePrice, p.SalePrice,
	)
	return scanProduct(row)
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+productColumns+` FROM products WHERE id = $1
	`, pgtype.UUID{Bytes: id, Valid: true})
	return scanProduct(row)
}

func (r *repository) List(ctx context.Context, name string, limit, offset int) ([]Product, int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+productColumns+`, COUNT(*) OVER () AS total
		FROM products
		WHERE active = true
		  AND ($1 = '' OR name ILIKE '%' || $1 || '%')
		ORDER BY name
		LIMIT $2 OFFSET $3
	`, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []Product
	var total int64
	for rows.Next() {
		p, t, err := scanProductWithTotal(rows)
		if err != nil {
			return nil, 0, err
		}
		total = t
		products = append(products, *p)
	}
	return products, total, rows.Err()
}

func (r *repository) Update(ctx context.Context, p Product) (*Product, error) {
	var desc pgtype.Text
	if p.Description != "" {
		desc = pgtype.Text{String: p.Description, Valid: true}
	}
	row := r.db.QueryRow(ctx, `
		UPDATE products
		SET product_code = $2, name = $3, description = $4,
		    purchase_price = $5, sale_price = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING `+productColumns,
		pgtype.UUID{Bytes: p.ID, Valid: true},
		p.ProductCode, p.Name, desc, p.PurchasePrice, p.SalePrice,
	)
	return scanProduct(row)
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET active = false, updated_at = NOW() WHERE id = $1 AND active = true`,
		pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── scanners ─────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(s scanner) (*Product, error) {
	var p Product
	var id pgtype.UUID
	var desc pgtype.Text
	err := s.Scan(&id, &p.ProductCode, &p.Name, &desc, &p.PurchasePrice, &p.SalePrice, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	p.ID = uuid.UUID(id.Bytes)
	p.Description = desc.String
	return &p, nil
}

func scanProductWithTotal(s scanner) (*Product, int64, error) {
	var p Product
	var id pgtype.UUID
	var desc pgtype.Text
	var total int64
	err := s.Scan(&id, &p.ProductCode, &p.Name, &desc, &p.PurchasePrice, &p.SalePrice, &p.CreatedAt, &p.UpdatedAt, &total)
	if err != nil {
		return nil, 0, err
	}
	p.ID = uuid.UUID(id.Bytes)
	p.Description = desc.String
	return &p, total, nil
}
