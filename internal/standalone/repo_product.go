package standalone

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/natthan/api/internal/modules/product"
)

type sqliteProductRepo struct{ db *sql.DB }

func NewProductRepository(db *sql.DB) product.Repository {
	return &sqliteProductRepo{db: db}
}

const productCols = `id, product_code, name, description, purchase_price, sale_price, created_at, updated_at`

func (r *sqliteProductRepo) Create(ctx context.Context, p product.Product) (*product.Product, error) {
	now := nowStr()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO products (id, product_code, name, description, purchase_price, sale_price, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		p.ID.String(), p.ProductCode, p.Name, nullStr(p.Description), p.PurchasePrice, p.SalePrice, now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, p.ID)
}

func (r *sqliteProductRepo) GetByID(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+productCols+` FROM products WHERE id = ?`, id.String())
	return scanProduct(row)
}

func (r *sqliteProductRepo) List(ctx context.Context, name string, limit, offset int) ([]product.Product, int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+productCols+`, COUNT(*) OVER () AS total
		FROM products
		WHERE active = 1
		  AND (? = '' OR LOWER(name) LIKE LOWER('%' || ? || '%'))
		ORDER BY name
		LIMIT ? OFFSET ?
	`, name, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []product.Product
	var total int64
	for rows.Next() {
		var p product.Product
		var id string
		var desc sql.NullString
		var cat, uat string
		var t int64
		if err := rows.Scan(&id, &p.ProductCode, &p.Name, &desc, &p.PurchasePrice, &p.SalePrice, &cat, &uat, &t); err != nil {
			return nil, 0, err
		}
		p.ID, _ = uuid.Parse(id)
		p.Description = desc.String
		p.CreatedAt = parseTime(cat)
		p.UpdatedAt = parseTime(uat)
		total = t
		products = append(products, p)
	}
	return products, total, rows.Err()
}

func (r *sqliteProductRepo) Update(ctx context.Context, p product.Product) (*product.Product, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE products
		SET product_code = ?, name = ?, description = ?, purchase_price = ?, sale_price = ?, updated_at = ?
		WHERE id = ?`,
		p.ProductCode, p.Name, nullStr(p.Description), p.PurchasePrice, p.SalePrice, nowStr(), p.ID.String())
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, product.ErrNotFound
	}
	return r.GetByID(ctx, p.ID)
}

func (r *sqliteProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products SET active = 0, updated_at = ? WHERE id = ? AND active = 1`,
		nowStr(), id.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return product.ErrNotFound
	}
	return nil
}

func scanProduct(s rowScanner) (*product.Product, error) {
	var p product.Product
	var id string
	var desc sql.NullString
	var cat, uat string
	err := s.Scan(&id, &p.ProductCode, &p.Name, &desc, &p.PurchasePrice, &p.SalePrice, &cat, &uat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product.ErrNotFound
		}
		return nil, err
	}
	p.ID, _ = uuid.Parse(id)
	p.Description = desc.String
	p.CreatedAt = parseTime(cat)
	p.UpdatedAt = parseTime(uat)
	return &p, nil
}
