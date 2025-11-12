package models

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Images struct {
	ImagesOne   *string `form:"imagesOne"`
	ImagesTwo   *string `form:"imagesTwo"`
	ImagesThree *string `form:"imagesThree"`
	ImagesFour  *string `form:"imagesFour"`
}

type ImagesBody struct {
	ImagesOne      *multipart.FileHeader `form:"imagesOne"`
	ImagesTwo      *multipart.FileHeader `form:"imagesTwo"`
	ImagesThree    *multipart.FileHeader `form:"imagesThree"`
	ImagesFour     *multipart.FileHeader `form:"imagesFour"`
	ImagesOneStr   string                `form:"imagesOneStr"`
	ImagesTwoStr   string                `form:"imagesTwoStr"`
	ImagesThreeStr string                `form:"imagesThreeStr"`
	ImagesFourStr  string                `form:"imagesFourStr"`
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

func CreateImagesbyId(ctx context.Context, db *pgxpool.Pool, body ImagesBody, idProduct int) (ImagesBody, error) {
	sql := `UPDATE product_images (photos_one, photos_two, photos_three, photos_four)
        VALUES ($1, $2, $3, $4)
        RETURNING photos_one, photos_two, photos_three, photos_four`

	values := []any{
		body.ImagesOneStr,
		body.ImagesTwoStr,
		body.ImagesThreeStr,
		body.ImagesFourStr,
	}
	var newImages ImagesBody
	if err := db.QueryRow(ctx, sql, values...).Scan(
		&newImages.ImagesOneStr,
		&newImages.ImagesTwoStr,
		&newImages.ImagesThreeStr,
		&newImages.ImagesFourStr); err != nil {
		log.Println("Failed to insert product_images:", err)
		return ImagesBody{}, err
	}
	return newImages, nil

}
