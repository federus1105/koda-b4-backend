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
	OrderNumber string     `json:"orderNumber"`
	Date        time.Time  `json:"date"`
	Status      bool       `json:"status"`
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
	Status        bool          `json:"status"`
	Total         float64       `json:"total"`
	Products      []ProductItem `json:"products"`
}

type UpdateStatusRequest struct {
	Status bool `json:"status"`
}

func GetListOrder(ctx context.Context, db *pgxpool.Pool, OrderNumber, Status string, limit, offset int) ([]OrderList, error) {
	var p OrderList
	var productsJSON []byte

	sql := `SELECT 
	'#ORD-' || LPAD(o.id::text, 3, '0') AS order_number,
    o.createdAt,
    o.status,
    o.total,
	JSON_AGG(
			JSON_BUILD_OBJECT(
				'name', p.name,
				'quantity', o.quantity
			)
		) AS order_items
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
		if err := rows.Scan(&p.OrderNumber, &p.Date, &p.Status, &p.Total, &productsJSON); err != nil {
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
    '#ORD-' || LPAD(o.id::text, 3, '0') AS order_code,
    o.fullname,
    o.address,
    o.phonenumber,
	pm.name AS payment_method,
    o.delivery,
    o.status,
    o.total,
    JSON_AGG(
        JSON_BUILD_OBJECT(
            'quantity', o.quantity,
            'size', o.size,
            'variant', o.variant,
            'product_name', p.name,
            'price_original', p.priceoriginal,
            'price_discount', p.pricediscount
        )
    ) AS products
	FROM orders o
	JOIN payment_method pm ON pm.id = o.id_paymentmethod
	JOIN product_orders po ON po.id_order = o.id
	JOIN product p ON p.id = po.id_product
	WHERE o.id = $1
	GROUP BY o.id, o.fullname, o.address, o.phonenumber, pm.name, o.delivery, o.status, o.total`

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

func UpdateOrderStatus(ctx context.Context, db *pgxpool.Pool, orderID int, status bool) error {
	sql := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := db.Exec(ctx, sql, status, orderID)
	return err
}
