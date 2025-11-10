package models

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Categories struct {
	Id        int       `json:"id"`
	Name      string    `json:"name" binding:"required,max=20"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetListCategories(ctx context.Context, db *pgxpool.Pool, name string, limit, offset int) ([]Categories, error) {
	sql := `SELECT id, name, created_at, updated_at FROM categories`

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
		if err := rows.Scan(&c.Id, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return categories, nil

}

func CreateCategory(ctx context.Context, db *pgxpool.Pool, body Categories) (Categories, error) {
	sql := `INSERT INTO categories (name) VALUES ($1) RETURNING id, name, created_at`
	values := []any{body.Name}
	var newCategory Categories
	if err := db.QueryRow(ctx, sql, values...).Scan(&newCategory.Id, &newCategory.Name, &newCategory.CreatedAt); err != nil {
		log.Println("Failed to insert Categories, Error :", err)
		return Categories{}, err
	}
	return newCategory, nil
}

func UpdateCategories(ctx context.Context, db *pgxpool.Pool, body Categories, id int) (Categories, error) {
	sql := `UPDATE categories
		SET name = $1
		WHERE id = $2
		RETURNING id, name, updated_at, created_at`

	var updated Categories
	err := db.QueryRow(ctx, sql, body.Name, id).Scan(
		&updated.Id, &updated.Name, &updated.UpdatedAt, &updated.CreatedAt)
	if err != nil {
		log.Println("Failed to update category:", err)
		return Categories{}, fmt.Errorf("category update failed: %w", err)
	}

	return updated, nil
}
