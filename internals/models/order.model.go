package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemList struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type OrderList struct {
	Id          int        `json:"id"`
	OrderNumber string     `json:"orderNumber"`
	Date        time.Time  `json:"date"`
	Status      string     `json:"status"`
	Total       float64    `json:"total"`
	Order       []ItemList `json:"order"`
}

type ProductItem struct {
	Quantity      int     `json:"quantity"`
	Size          string  `json:"size"`
	Variant       string  `json:"variant"`
	ProductName   string  `json:"product_name"`
	PriceOriginal float64 `json:"price_original"`
	PriceDiscount float64 `json:"price_discount"`
}

type OrderDetail struct {
	OrderNumber   string        `json:"orderNumber"`
	Fullname      string        `json:"fullname"`
	Address       string        `json:"address"`
	PhoneNumber   string        `json:"phonenumber"`
	PaymentMethod string        `json:"payment_method"`
	Delivery      string        `json:"delivery"`
	Status        string        `json:"status"`
	Total         float64       `json:"total"`
	Products      []ProductItem `json:"products"`
}

type UpdateStatusRequest struct {
	Status int `json:"status"`
}

func GetListOrder(ctx context.Context, db *pgxpool.Pool, OrderNumber string, status, limit, offset int) ([]OrderList, error) {
	var p OrderList
	var productsJSON []byte

	sql := `
SELECT
	o.id,
    o.order_number,
    o.createdAt,
    s.name,
    o.total,
    JSON_AGG(
        JSON_BUILD_OBJECT(
            'name', p.name,
            'quantity', po.quantity
        )
    ) AS order_items
FROM orders o
JOIN product_orders po ON po.id_order = o.id
JOIN product p ON p.id = po.id_product
JOIN status s ON s.id = o.id_status
`

	// --- SQL DINAMIS ---
	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	// --- SEARCH ---
	if strings.TrimSpace(OrderNumber) != "" {
		if id, err := strconv.Atoi(OrderNumber); err == nil {
			formatted := fmt.Sprintf("#ORD-%03d", id)
			whereClauses = append(whereClauses, fmt.Sprintf("o.order_number = $%d", argIdx))
			args = append(args, formatted)
			argIdx++
		}
	}

	// --- FILTER STATUS ---
	if status != 0 {
		sql += fmt.Sprintf(" AND o.id_status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	// --- APPEND ALL WHERE ---
	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// --- GROUP BY ---
	sql += ` GROUP BY o.id, s.name, o.createdAt, o.id_status, o.total`

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
		if err := rows.Scan(&p.Id, &p.OrderNumber, &p.Date, &p.Status, &p.Total, &productsJSON); err != nil {
			return nil, err
		}

		// --- PARSE JSON TO SLICE OF ITEMSLIST ---
		if err := json.Unmarshal(productsJSON, &p.Order); err != nil {
			return nil, fmt.Errorf("failed to parse order items JSON: %w", err)
		}
		order = append(order, p)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return order, nil

}

func GetDetailOrder(ctx context.Context, db *pgxpool.Pool, OrderID int) (OrderDetail, error) {
	var order OrderDetail
	var productsJSON []byte

	sql := `SELECT 
	o.order_number,
    o.fullname,
    o.address,
    o.phonenumber,
	pm.name AS payment_method,
    d.name,
    s.name,
    o.total,
    JSON_AGG(
        JSON_BUILD_OBJECT(
            'quantity', po.quantity,
            'size', po.size,
            'variant', po.variant,
            'product_name', p.name,
            'price_original', p.priceoriginal,
            'price_discount', p.pricediscount
        )
    ) AS products
	FROM orders o
	JOIN payment_method pm ON pm.id = o.id_paymentmethod
	JOIN product_orders po ON po.id_order = o.id
	JOIN product p ON p.id = po.id_product
    JOIN delivery d ON d.id = o.id_delivery
    JOIN status s ON s.id = o.id_status
	WHERE o.id = $1
	GROUP BY o.id, o.fullname, o.address, o.phonenumber, pm.name, d.name, s.name, o.total`

	err := db.QueryRow(ctx, sql, OrderID).Scan(
		&order.OrderNumber,
		&order.Fullname,
		&order.Address,
		&order.PhoneNumber,
		&order.PaymentMethod,
		&order.Delivery,
		&order.Status,
		&order.Total,
		&productsJSON,
	)
	if err != nil {
		return order, err
	}

	if err := json.Unmarshal(productsJSON, &order.Products); err != nil {
		return order, errors.New("failed to parse products JSON: " + err.Error())
	}

	return order, nil
}

func UpdateOrderStatus(ctx context.Context, db *pgxpool.Pool, orderID, status int) error {
	sql := `UPDATE orders SET id_status = $1 WHERE id = $2`
	_, err := db.Exec(ctx, sql, status, orderID)
	return err
}

func GetCountOrder(ctx context.Context, db *pgxpool.Pool, OrderNumber string, status int) (int64, error) {
	sql := `SELECT COUNT(DISTINCT o.id)
	FROM orders o
	JOIN product_orders po ON po.id_order = o.id
	JOIN product p ON p.id = po.id_product
	JOIN status s ON s.id = o.id_status
	`

	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	// --- Filter OrderNumber ---
	if strings.TrimSpace(OrderNumber) != "" {
		if id, err := strconv.Atoi(OrderNumber); err == nil {
			formatted := fmt.Sprintf("#ORD-%03d", id)
			whereClauses = append(whereClauses, fmt.Sprintf("o.order_number = $%d", argIdx))
			args = append(args, formatted)
			argIdx++
		}
	}

	// --- Filter Status ---
	if status != 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("o.id_status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	// --- Append WHERE clause ---
	if len(whereClauses) > 0 {
		sql += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var total int64
	err := db.QueryRow(ctx, sql, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
