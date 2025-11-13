package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type History struct {
	Id          int       `json:"id"`
	OrderNumber string    `json:"order_number"`
	Date        time.Time `json:"date"`
	Status      string    `json:"status"`
	Total       float64   `json:"total"`
	Image       string    `json:"image"`
}

func GetHistory(ctx context.Context, db *pgxpool.Pool, IdUser int, month, status, limit, offset int) ([]History, error) {

	sql := `SELECT o.id, o.order_number, o.createdat, s.name, o.total, pi.photos_one FROM orders o
	JOIN product_orders po ON po.id_order = o.id
	JOIN product p ON p.id = po.id_product
	JOIN product_images pi ON pi.id = p.id_product_images
	JOIN status s ON s.id = o.id
	WHERE o.id_account = $1`

	args := []interface{}{IdUser}
	argIdx := 2

	// --- FILTER MONTH---
	if month != 0 {
		sql += fmt.Sprintf(" AND EXTRACT(MONTH FROM o.createdat) = $%d", argIdx)
		args = append(args, month)
		argIdx++
	}

	// --- FILTER STATUS ---
	if status != 0 {
		sql += fmt.Sprintf(" AND o.id_status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	// --- SORT BY CREATED AT ---
	sql += " ORDER BY o.createdat DESC"

	// --- LIMIT & OFFSET ---
	sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// --- EXECUTE QUERY ---
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []History
	for rows.Next() {
		var h History
		if err := rows.Scan(&h.Id, &h.OrderNumber, &h.Date, &h.Status, &h.Total, &h.Image); err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return histories, nil
}
