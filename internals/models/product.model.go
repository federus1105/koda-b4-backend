package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Product struct {
	Id          int    `json:"id"`
	Image       string `json:"image"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	Description string `json:"description"`
	Stock       string `json:"stock"`
}

func GetListProduct(ctx context.Context, db *pgxpool.Pool, name string, limit, offset int) ([]Product, error) {
	sql := `SELECT pi.photos_one as image,
	p.id,
	p.name, 
	p.priceoriginal as price,
	p.description, 
	p.stock FROM product p
	JOIN product_images pi ON pi.id = p.id_product_images`

	args := []interface{}{}
	argIdx := 1

	// --- SEARCH ---
	if strings.TrimSpace(name) != "" {
		sql += fmt.Sprintf(" WHERE p.name ILIKE $%d", argIdx)
		args = append(args, "%"+name+"%")
		argIdx++
	}

	// --- ORDER LIMIT, OFFSET ---
	sql += fmt.Sprintf(" ORDER BY p.name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// --- EXECUTE QUERY ---
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.Image, &p.Id, &p.Name, &p.Price, &p.Description, &p.Stock); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return products, nil
}
