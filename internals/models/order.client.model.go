package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CartItemRequest struct {
	ProductID int `json:"product_id" binding:"gt=0"`
	SizeID    int `json:"size" binding:"gt=0,lte=3"`
	VariantID int `json:"variant" binding:"gt=0,lte=2"`
	Quantity  int `json:"quantity" binding:"gt=0"`
}

type CartItemResponse struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	ProductID int       `json:"product_id"`
	SizeID    int       `json:"size_id"`
	VariantID int       `json:"variant_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Card struct {
	Id            int     `json:"id"`
	Image         string  `json:"images"`
	Name          string  `json:"name"`
	Quantity      int     `json:"qty"`
	Size          string  `json:"size"`
	Variant       string  `json:"variant"`
	Price         float64 `json:"price"`
	PriceDiscount float64 `json:"discount"`
	FlashSale     bool    `json:"flash_sale"`
	Subtotal      float64 `json:"subtotal"`
}

func CreateCartProduct(ctx context.Context, db *pgxpool.Pool, accountID int, input CartItemRequest) (*CartItemResponse, error) {
	sql := `INSERT INTO cart (account_id, product_id, size_id, variant_id, quantity)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (account_id, product_id, size_id, variant_id)
		DO UPDATE SET 
			quantity = cart.quantity + EXCLUDED.quantity,
			updated_at = NOW()
		RETURNING id, account_id, product_id, size_id, variant_id, quantity, created_at, updated_at;`

	var item CartItemResponse

	err := db.QueryRow(ctx, sql,
		accountID,
		input.ProductID,
		input.SizeID,
		input.VariantID,
		input.Quantity,
	).Scan(
		&item.ID,
		&item.AccountID,
		&item.ProductID,
		&item.SizeID,
		&item.VariantID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return &CartItemResponse{}, fmt.Errorf("failed to insert or update cart item: %w", err)
	}

	return &item, nil
}

func GetCartProduct(ctx context.Context, db *pgxpool.Pool, UserID int) ([]Card, error) {
	sql := `SELECT 
    c.id, 
    p.name, 
    p.priceoriginal,
    p.pricediscount,
    p.flash_sale,
    pi.photos_one, 
    c.quantity, 
    s.name AS size, 
    v.name AS variant,
    (p.priceoriginal * c.quantity) AS subtotal
FROM cart c
JOIN sizes s ON s.id = c.size_id
JOIN variants v ON v.id = c.variant_id
JOIN product p ON p.id = c.product_id
JOIN product_images pi ON p.id_product_images = pi.id
WHERE c.account_id = $1;`

	rows, err := db.Query(ctx, sql, UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var carts []Card
	for rows.Next() {
		var c Card
		if err := rows.Scan(
			&c.Id,
			&c.Name,
			&c.Price,
			&c.PriceDiscount,
			&c.FlashSale,
			&c.Image,
			&c.Quantity,
			&c.Size,
			&c.Variant,
			&c.Subtotal); err != nil {
			return nil, err
		}
		carts = append(carts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return carts, nil
}
