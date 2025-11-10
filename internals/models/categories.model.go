package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Categories struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func GetListCategories(ctx context.Context, db *pgxpool.Pool, name string, limit, offset int) ([]Categories, error) {
	sql := `SELECT id, name FROM categories`

	args := []interface{}{}
	argIdx := 1

	// --- SEARCH ---
	if strings.TrimSpace(name) != "" {
		sql += fmt.Sprintf(" WHERE name ILIKE $%d", argIdx)
		args = append(args, "%"+name+"%")
		argIdx++
	}

	// --- ORDER LIMIT, OFFSET ---
	sql += fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// --- EXECUTE QUERY ---
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var categories []Categories
	for rows.Next() {
		var c Categories
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return categories, nil

}
