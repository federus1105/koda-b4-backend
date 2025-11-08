package models

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
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

type CreateProducts struct {
	Id             int                   `form:"id"`
	Name           string                `form:"name" binding:"required"`
	ImageId        int                   `form:"imageId"`
	Image_one      *multipart.FileHeader `form:"image_one" binding:"required"`
	Image_two      *multipart.FileHeader `form:"image_two"`
	Image_three    *multipart.FileHeader `form:"image_three"`
	Image_four     *multipart.FileHeader `form:"image_four"`
	Image_oneStr   string                `form:"image_oneStr"`
	Image_twoStr   string                `form:"image_twoStr,omitempty"`
	Image_threeStr string                `form:"image_threeStr,omitempty"`
	Image_fourStr  string                `form:"image_fourStr,omitempty"`
	Price          float64               `form:"price" binding:"required,gte=5000"`
	Rating         float64               `form:"rating" binding:"required,gte=1,lte=10"`
	Description    string                `form:"description" binding:"required"`
	Stock          int                   `form:"stock" binding:"gte=0"`
}

type ProductResponse struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	ImageID     int               `json:"idImage"`
	Images      map[string]string `json:"images,omitempty"`
	Price       float64           `json:"price"`
	Rating      float64           `json:"rating"`
	Description string            `json:"description"`
	Stock       int               `json:"stock"`
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

func CreateProduct(ctx context.Context, db *pgxpool.Pool, body CreateProducts) (CreateProducts, error) {
	// --- START QUERY TRANSACTION ---
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return CreateProducts{}, err
	}
	defer tx.Rollback(ctx)

	// --- INSERT PRODUCT IMAGES ---
	imageSQL := `
        INSERT INTO product_images (photos_one, photos_two, photos_three, photos_four)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	var imageId int
	if err := tx.QueryRow(ctx, imageSQL,
		body.Image_oneStr,
		body.Image_twoStr,
		body.Image_threeStr,
		body.Image_fourStr,
	).Scan(&imageId); err != nil {
		log.Println("Failed to insert product_images:", err)
		return CreateProducts{}, err
	}

	// --- INSERT PRODUCT
	productSQL := `INSERT INTO product (name, description, rating, priceoriginal, stock, id_product_images) 
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, name, description, rating, priceoriginal, stock, id_product_images`
	values := []any{body.Name, body.Description, body.Rating, body.Price, body.Stock, imageId}
	var newProduct CreateProducts
	if err := tx.QueryRow(ctx, productSQL, values...).Scan(
		&newProduct.Id,
		&newProduct.Name,
		&newProduct.Description,
		&newProduct.Rating,
		&newProduct.Price,
		&newProduct.Stock,
		&newProduct.ImageId,
	); err != nil {
		log.Println("Failed to insert product:", err)
		return CreateProducts{}, err
	}

	// --- ASIGN IMAGE STR TO RETURN STRUCT RESPONSE---
	newProduct.Image_oneStr = body.Image_oneStr
	newProduct.Image_twoStr = body.Image_twoStr
	newProduct.Image_threeStr = body.Image_threeStr
	newProduct.Image_fourStr = body.Image_fourStr

	// --- COMMIT TRANSACTION ---
	if err := tx.Commit(ctx); err != nil {
		log.Println("Failed to commit transaction:", err)
		return CreateProducts{}, err
	}

	return newProduct, nil
}
