package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FavoriteProduct struct {
	Id          int    `json:"id"`
	Image       string `json:"image"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	Description string `json:"description"`
}

func GetListFavoriteProduct(ctx context.Context, db *pgxpool.Pool, limit, offset int) ([]FavoriteProduct, error) {
	sql := `SELECT pi.photos_one as image,
	p.id,
	p.name, 
	p.priceoriginal as price,
	p.description
 	FROM product p
	JOIN product_images pi ON pi.id = p.id_product_images
	WHERE is_deleted = false AND is_favorite = true
	LIMIT $1 OFFSET $2`

	// --- EXECUTE QUERY ---
	rows, err := db.Query(ctx, sql, limit, offset)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var products []FavoriteProduct
	for rows.Next() {
		var fp FavoriteProduct
		if err := rows.Scan(&fp.Image, &fp.Id, &fp.Name, &fp.Price, &fp.Description); err != nil {
			return nil, err
		}
		products = append(products, fp)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return products, nil

}

func GetListProductFilter(ctx context.Context, db *pgxpool.Pool,
	name string,
	categoryIDs []int,
	minPrice, maxPrice float64,
	sortBy string,
	limit, offset int) ([]FavoriteProduct, error) {

	sql := `
SELECT DISTINCT p.id,
       p.name,
       p.priceoriginal AS price,
       p.description,
       pi.photos_one AS image
FROM product p
JOIN product_images pi ON pi.id = p.id_product_images
JOIN product_categories pc ON p.id = pc.id_product
WHERE p.is_deleted = false
`
	args := []interface{}{}
	argIdx := 1

	// --- SEARCH BY NAME ---
	if strings.TrimSpace(name) != "" {
		sql += fmt.Sprintf(" AND p.name ILIKE $%d", argIdx)
		args = append(args, "%"+name+"%")
		argIdx++
	}

	// --- FILTER CATEGORY ---
	if len(categoryIDs) > 0 {
		placeholders := make([]string, len(categoryIDs))
		for i, id := range categoryIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}

		sql += fmt.Sprintf(`
      AND p.id IN (
          SELECT id_product
          FROM product_categories
          WHERE id_categories IN (%s)
          GROUP BY id_product
          HAVING COUNT(DISTINCT id_categories) = %d
      )
    `, strings.Join(placeholders, ","), len(categoryIDs))
	}

	if minPrice > 0 {
		sql += fmt.Sprintf(" AND p.priceOriginal >= $%d", argIdx)
		args = append(args, minPrice)
		argIdx++
	}

	if maxPrice > 0 {
		sql += fmt.Sprintf(" AND p.priceOriginal <= $%d", argIdx)
		args = append(args, maxPrice)
		argIdx++
	}

	// --- SORT ---
	if sortBy != "priceOriginal" {
		sortBy = "name"
	}
	sql += fmt.Sprintf(" ORDER BY p.%s ASC", sortBy)

	// --- LIMIT & OFFSET ---
	sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// --- EXECUTE QUERY ---
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []FavoriteProduct
	for rows.Next() {
		var p FavoriteProduct
		if err := rows.Scan(&p.Id, &p.Name, &p.Price, &p.Description, &p.Image); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return products, nil
}
