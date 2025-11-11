package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Images struct {
	ImagesOne   *string `form:"imagesOne"`
	ImagesTwo   *string `form:"imagesTwo"`
	ImagesThree *string `form:"imagesThree"`
	ImagesFour  *string `form:"imagesFour"`
}

func GetListImageById(ctx context.Context, db *pgxpool.Pool, idProduct int) ([]Images, error) {
	sql := `SELECT pi.photos_one, pi.photos_two, pi.photos_three, pi.photos_four
		FROM product_images pi
		JOIN product p ON p.id_product_images =  pi.id
		WHERE p.id = $1;`

	// --- EXECUTE QUERY ---
	rows, err := db.Query(ctx, sql, idProduct)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var images []Images

	for rows.Next() {
		var p Images
		if err := rows.Scan(&p.ImagesOne, &p.ImagesTwo, &p.ImagesThree, &p.ImagesFour); err != nil {
			return nil, err
		}
		images = append(images, p)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return images, nil
}
