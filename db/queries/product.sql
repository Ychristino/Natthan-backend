-- name: CreateProduct :one
INSERT INTO products (id, product_code, name, description, purchase_price, sale_price)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProductByID :one
SELECT id, product_code, name, description, purchase_price, sale_price, created_at, updated_at
FROM products
WHERE id = $1;

-- name: ListProducts :many
SELECT id, product_code, name, description, purchase_price, sale_price, created_at, updated_at
FROM products
WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%')
ORDER BY name
LIMIT $2 OFFSET $3;

-- name: CountProducts :one
SELECT COUNT(*) FROM products
WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%');

-- name: UpdateProduct :one
UPDATE products
SET product_code   = $2,
    name           = $3,
    description    = $4,
    purchase_price = $5,
    sale_price     = $6,
    updated_at     = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;
