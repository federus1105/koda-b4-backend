package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CartItemRequest struct {
	ProductID int `json:"product_id"`
	SizeID    int `json:"size"`
	VariantID int `json:"variant"`
	Quantity  int `json:"quantity"`
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