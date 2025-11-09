package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderList struct {
	OrderNumber string    `json:"orderNumber"`
	Date        time.Time `json:"date"`
	Order       string    `json:"order"`
	Status      bool      `json:"status"`
	Total       float64   `json:"total"`
}

func GetListOrder(ctx context.Context, db *pgxpool.Pool, OrderNumber, Status string, limit, offset int) ([]OrderList, error) {
	sql := `SELECT 
	'#ORD-' || LPAD(o.id::text, 3, '0') AS order_number,
    o.createdAt,
    STRING_AGG(p.name || ' ' || o.quantity || 'x', ', ') AS products,
    o.status,
    o.total
	FROM orders o
	JOIN product_orders po ON po.id_order = o.id
	JOIN product p ON p.id = po.id_product`

	// --- SQL DINAMIS ---
	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	// --- SEARCH ---
	if strings.TrimSpace(OrderNumber) != "" {
		if id, err := strconv.Atoi(OrderNumber); err == nil {
			whereClauses = append(whereClauses, fmt.Sprintf("o.id = $%d", argIdx))
			args = append(args, id)
			argIdx++
		}
	}

	// --- FILTER STATUS ---
	switch Status {
	case "true":
		whereClauses = append(whereClauses, "o.status = true")
	case "false":
		whereClauses = append(whereClauses, "o.status = false")
	}

	// --- APPEND ALL WHERE ---
	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// --- GROUP BY ---
	sql += ` GROUP BY o.id, o.createdAt, o.status, o.total`

	// --- ORDER LIMIT, OFFSET ---
	sql += fmt.Sprintf(" ORDER BY o.createdAt DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// --- EXECUTE QUERY ---
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var order []OrderList
	for rows.Next() {
		var p OrderList
		if err := rows.Scan(&p.OrderNumber, &p.Date, &p.Order, &p.Status, &p.Total); err != nil {
			return nil, err
		}

		order = append(order, p)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return order, nil

}
