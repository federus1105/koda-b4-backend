package models

import (
	"context"
	"encoding/json"
	"errors"
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

type Items struct {
	ID         int    `json:"id"`
	Image      string `json:"image"`
	Flash_sale bool   `json:"flash_Sale"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	Delivery   string `json:"delivery"`
	Size       string `json:"size"`
	Variant    string `json:"variant"`
}

type DetailHistories struct {
	Id          int     `json:"id"`
	OrderNumber string     `json:"order_number"`
	Fullname    string  `json:"fullname"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	Addres      string  `json:"address"`
	Payment     string  `json:"payment"`
	Delivery    string  `json:"delivery"`
	Status      string  `json:"status"`
	Total       float64 `json:"total"`
	Items       []Items `json:"items"`
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

func DetailHistory(ctx context.Context, db *pgxpool.Pool, idUser, idHistory int) (DetailHistories, error) {
	var history DetailHistories
	var productsJSON []byte

	sql := `SELECT o.id, o.order_number, 
    o.fullname, o.phonenumber, 
    o.email, o.address, pm.name as payment, 
    d.name as delivery,
    s.name as status,
    o.total,
    json_agg(
        json_build_object(
            'id', p.id,
            'image', pi.photos_one,
            'flash_sale', p.flash_sale,
            'name',  p.name,
            'quantity', po.quantity,
            'delivery', d.name,
            'size' ,po.size,
            'variant',  po.variant
        )
    ) AS items
    FROM orders o
    JOIN product_orders po ON o.id = po.id_order
    JOIN payment_method pm ON pm.id = o.id_paymentmethod
    JOIN delivery d ON d.id = o.id_delivery
    JOIN status s ON s.id = o.id_status
    JOIN product p ON p.id = po.id_product
    JOIN product_images pi ON pi.id = p.id_product_images
    WHERE o.id = $1 AND o.id_account = $2
    GROUP BY 
    o.id, o.order_number, o.fullname, o.phonenumber, o.email,
    o.address, pm.name, d.name, s.name, o.total`

	err := db.QueryRow(ctx, sql, idHistory, idUser).Scan(
		&history.Id,
		&history.OrderNumber,
		&history.Fullname,
		&history.Phone,
		&history.Email,
		&history.Addres,
		&history.Payment,
		&history.Delivery,
		&history.Status,
		&history.Total,
		&productsJSON,
	)
	if err != nil {
		return history, err
	}
	if err := json.Unmarshal(productsJSON, &history.Items); err != nil {
		return history, errors.New("failed to parse products JSON: " + err.Error())
	}

	return history, nil
}
