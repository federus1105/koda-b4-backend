package models

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Product struct {
	Id          int    `json:"id"`
	Image       string `json:"image"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	Description string `json:"description"`
	Stock       string `json:"stock"`
}
type FavoriteProduct struct {
	Id          int    `json:"id"`
	Image       string `json:"image"`
	Name        string `json:"name"`
	Price       string `json:"price"`
	Description string `json:"description"`
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

type UpdateProducts struct {
	Id             int                   `form:"id"`
	Name           *string               `form:"name"`
	Image_one      *multipart.FileHeader `form:"image_one"`
	Image_two      *multipart.FileHeader `form:"image_two"`
	Image_three    *multipart.FileHeader `form:"image_three"`
	Image_four     *multipart.FileHeader `form:"image_four"`
	Image_oneStr   *string               `form:"image_oneStr"`
	Image_twoStr   *string               `form:"image_twoStr,omitempty"`
	Image_threeStr *string               `form:"image_threeStr,omitempty"`
	Image_fourStr  *string               `form:"image_fourStr,omitempty"`
	Price          *float64              `form:"price" binding:"omitempty,gte=5000"`
	Rating         *float64              `form:"rating" binding:"omitempty,gte=1,lte=10"`
	Description    *string               `form:"description"`
	Stock          *int                  `form:"stock" binding:"omitempty,gte=0"`
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

func GetListProduct(ctx context.Context, db *pgxpool.Pool, rd *redis.Client, name string, limit, offset int) ([]Product, error) {
	redisKey := "list-product"

	// --- GET CACHE ---
	if cached, err := libs.GetFromCache[[]Product](ctx, rd, redisKey); err != nil {
		log.Println("Redis Error:", err)
	} else if cached != nil && len(*cached) > 0 {
		log.Printf("Key %s found in cache Served in using Redis 👌", redisKey)
		return *cached, nil
	}

	sql := `SELECT pi.photos_one as image,
	fp.id,
	p.name, 
	p.priceoriginal as price,
	p.description, 
	p.stock FROM product p
	JOIN product_images pi ON pi.id = p.id_product_images
	WHERE is_deleted = false`

	args := []interface{}{}
	argIdx := 1

	// --- SEARCH ---
	if strings.TrimSpace(name) != "" {
		sql += fmt.Sprintf(" AND p.name ILIKE $%d", argIdx)
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

	// --- SAVING TO CACHE ---
	if offset == 0 {
		if err := libs.SetToCache(ctx, rd, redisKey, products, 1*time.Minute); err != nil {
			log.Println("Redis Error:", err)
		}
	}
	return products, nil
}

func CreateProduct(ctx context.Context, db *pgxpool.Pool, rd *redis.Client, body CreateProducts) (CreateProducts, error) {
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

	// --- INVALIDATE ---
	if err := libs.InvalidateCacheByPattern(ctx, rd, "list-product"); err != nil {
		log.Println("Failed to invalidate product cache:", err)
	}

	return newProduct, nil
}

func EditProduct(ctx context.Context, db *pgxpool.Pool, body UpdateProducts, images map[string]*string) (CreateProducts, error) {
	// --- START QUERY TRANSACTION ---
	tx, err := db.Begin(ctx)
	if err != nil {
		return CreateProducts{}, err
	}
	defer tx.Rollback(ctx)

	// --- DYNAMIC SET PRODUCT ---
	setClauses := []string{}
	args := []any{}
	idx := 1

	if body.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name=$%d", idx))
		args = append(args, *body.Name)
		idx++
	}
	if body.Price != nil {
		setClauses = append(setClauses, fmt.Sprintf("priceoriginal=$%d", idx))
		args = append(args, *body.Price)
		idx++
	}
	if body.Rating != nil {
		setClauses = append(setClauses, fmt.Sprintf("rating=$%d", idx))
		args = append(args, *body.Rating)
		idx++
	}
	if body.Stock != nil {
		setClauses = append(setClauses, fmt.Sprintf("stock=$%d", idx))
		args = append(args, *body.Stock)
		idx++
	}
	if body.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description=$%d", idx))
		args = append(args, *body.Description)
		idx++
	}

	if len(setClauses) > 0 {
		query := fmt.Sprintf("UPDATE product SET %s WHERE id=$%d", strings.Join(setClauses, ","), idx)
		args = append(args, body.Id)
		_, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return CreateProducts{}, err
		}
	}

	// --- TAKE THE ID FIRST ---
	var imageId int
	err = tx.QueryRow(ctx, "SELECT id_product_images FROM product WHERE id=$1", body.Id).Scan(&imageId)
	if err != nil {
		return CreateProducts{}, err
	}

	// --- HANDLE IMAGES ---
	for key, val := range images {
		var column string
		switch key {
		case "photos_one":
			column = "photos_one"
		case "photos_two":
			column = "photos_two"
		case "photos_three":
			column = "photos_three"
		case "photos_four":
			column = "photos_four"
		default:
			continue
		}

		if val != nil {
			imgQuery := fmt.Sprintf("UPDATE product_images SET %s=$1 WHERE id=$2", column)
			_, err := tx.Exec(ctx, imgQuery, *val, imageId)
			if err != nil {
				return CreateProducts{}, err
			}
		}
	}

	// ----- GET DATA  ----
	var product CreateProducts
	err = tx.QueryRow(ctx, `
        SELECT p.id, name, p.description, p.rating, p.priceoriginal, p.stock, p.id_product_images,
               pi.photos_one, pi.photos_two, pi.photos_three, pi.photos_four
        FROM product p
        JOIN product_images pi ON p.id_product_images = pi.id
        WHERE p.id=$1
    `, body.Id).Scan(
		&product.Id,
		&product.Name,
		&product.Description,
		&product.Rating,
		&product.Price,
		&product.Stock,
		&product.ImageId,
		&product.Image_oneStr,
		&product.Image_twoStr,
		&product.Image_threeStr,
		&product.Image_fourStr,
	)
	if err != nil {
		return CreateProducts{}, err
	}

	// --- ASIGN IMAGE STR TO RETURN STRUCT RESPONSE---
	if body.Image_oneStr != nil {
		product.Image_oneStr = *body.Image_oneStr
	}
	if body.Image_twoStr != nil {
		product.Image_twoStr = *body.Image_twoStr
	}

	if body.Image_threeStr != nil {
		product.Image_threeStr = *body.Image_threeStr
	}
	if body.Image_fourStr != nil {
		product.Image_fourStr = *body.Image_fourStr
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Failed to commit transaction:", err)
		return CreateProducts{}, err
	}

	return product, nil

}

func DeleteProduct(ctx context.Context, db *pgxpool.Pool, id int) error {
	sql := `UPDATE product SET is_deleted = TRUE
	WHERE id = $1`

	result, err := db.Exec(ctx, sql, id)
	if err != nil {
		log.Printf("failed to execute delete query: %v", err)
		if ctxErr := ctx.Err(); ctxErr != nil {
			log.Printf("context error: %v", ctxErr)
		}
		return err
	}

	rows := result.RowsAffected()
	log.Printf("Rows affected: %d", rows)

	if rows == 0 {
		return fmt.Errorf("product with id %d not found", id)
	}

	log.Printf("product with id %d successfully deleted", id)
	return nil
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
