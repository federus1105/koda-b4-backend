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
	Id          int      `json:"id"`
	Image       string   `json:"image"`
	Name        string   `json:"name"`
	Price       string   `json:"price"`
	Description string   `json:"description"`
	Stock       string   `json:"stock"`
	Size        []string `json:"size"`
}
type CreateProducts struct {
	Id             int                   `form:"id"`
	Name           string                `form:"name" binding:"required"`
	ImageId        int                   `form:"imageId"`
	Image_one      *multipart.FileHeader `form:"image_one"`
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
	Size           []int                 `form:"size,omitempty" binding:"max=3,dive,gt=0,lte=3"`
	Variant        []int                 `form:"variant,omitempty" binding:"max=2,dive,gt=0,lte=2"`
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
	Size           []int                 `form:"size,omitempty" binding:"max=3,dive,gt=0,lte=3"`
	Variant        []int                 `form:"variant,omitempty" binding:"max=2,dive,gt=0,lte=2"`
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
	Size        []int             `json:"size,omitempty"`
	Variant     []int             `json:"variant,omitempty"`
}

func GetListProduct(ctx context.Context, db *pgxpool.Pool, rd *redis.Client, name string, limit, offset int) ([]Product, error) {
	// --- REDIS KEY ---
	redisKey := fmt.Sprintf(
		"list-product:name=%s:limit=%d:offset=%d",
		strings.ToLower(name),
		limit,
		offset,
	)

	// --- GET CACHE ---
	if cached, err := libs.GetFromCache[[]Product](ctx, rd, redisKey); err != nil {
		log.Println("Redis Error:", err)
	} else if cached != nil && len(*cached) > 0 {
		log.Printf("Key %s found in cache Served in using Redis 👌", redisKey)
		return *cached, nil
	}

	sql := `SELECT
        p.id,
        p.name,
        pi.photos_one AS image,
        p.priceOriginal AS price,
        p.description,
        p.stock,
        s.name AS size
    FROM product p
    JOIN product_images pi 
        ON pi.id = p.id_product_images
    JOIN size_product sp
        ON sp.id_product = p.id
    JOIN sizes s
        ON s.id = sp.id_size
    WHERE p.is_deleted = false`

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

	productMap := make(map[int]*Product)
	for rows.Next() {
		var id int
		var name, image, description string
		var price float64
		var stock int
		var size *string

		if err := rows.Scan(&id, &name, &image, &price, &description, &stock, &size); err != nil {
			return nil, err
		}

		// -- PARSE TO STRING ---
		priceStr := fmt.Sprintf("%.0f", price)
		stockStr := fmt.Sprintf("%d", stock)

		if p, exists := productMap[id]; exists {
			if size != nil {
				p.Size = append(p.Size, *size)
			}
		} else {
			newProduct := Product{
				Id:          id,
				Name:        name,
				Image:       image,
				Price:       priceStr,
				Description: description,
				Stock:       stockStr,
				Size:        []string{},
			}
			if size != nil {
				newProduct.Size = []string{*size}
			}
			productMap[id] = &newProduct
		}
	}

	products := make([]Product, 0, len(productMap))
	for _, p := range productMap {
		products = append(products, *p)
	}

	// --- SAVING TO CACHE ---
	if err := libs.SetToCache(ctx, rd, redisKey, products, 5*time.Minute); err != nil {
		log.Println("Redis Error:", err)
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

	// --- INSERT SIZE PRODUCT ---
	if len(body.Size) > 0 {
		sizeSQL := `INSERT INTO size_product (id_product, id_size) VALUES ($1, $2)`
		for _, sizeID := range body.Size {
			_, err := tx.Exec(ctx, sizeSQL, newProduct.Id, sizeID)
			if err != nil {
				log.Println("Failed to insert size product:", err)
				return CreateProducts{}, err
			}
		}
	}

	// --- INSERT VARIANT PRODUCT ---
	if len(body.Variant) > 0 {
		VariantSQL := `INSERT INTO variant_product (id_product, id_variant) VALUES ($1, $2)`
		for _, VariantID := range body.Variant {
			_, err := tx.Exec(ctx, VariantSQL, newProduct.Id, VariantID)
			if err != nil {
				log.Println("Failed to insert variant product:", err)
				return CreateProducts{}, err
			}
		}
	}

	// --- ASIGN IMAGE STR TO RETURN STRUCT RESPONSE---
	newProduct.Image_oneStr = body.Image_oneStr
	newProduct.Image_twoStr = body.Image_twoStr
	newProduct.Image_threeStr = body.Image_threeStr
	newProduct.Image_fourStr = body.Image_fourStr
	newProduct.Size = body.Size
	newProduct.Variant = body.Variant

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

	// --- HANDLE SIZE ---
	if len(body.Size) > 0 {
		_, err := tx.Exec(ctx, "DELETE FROM size_product WHERE id_product=$1", body.Id)
		if err != nil {
			return CreateProducts{}, err
		}

		// Insert size baru
		for _, sizeID := range body.Size {
			if sizeID > 0 && sizeID <= 3 {
				_, err := tx.Exec(ctx, "INSERT INTO size_product (id_product, id_size) VALUES ($1, $2)", body.Id, sizeID)
				if err != nil {
					return CreateProducts{}, err
				}
			}
		}
	}

	// --- HANDLE VARIANT ---
	if len(body.Variant) > 0 {
		// Hapus semua variant lama
		_, err := tx.Exec(ctx, "DELETE FROM variant_product WHERE id_product=$1", body.Id)
		if err != nil {
			return CreateProducts{}, err
		}

		// Insert variant baru
		for _, variantID := range body.Variant {
			if variantID > 0 && variantID <= 2 {
				_, err := tx.Exec(ctx, "INSERT INTO variant_product (id_product, id_variant) VALUES ($1, $2)", body.Id, variantID)
				if err != nil {
					return CreateProducts{}, err
				}
			}
		}
	}

	// ----- GET DATA  ----
	var product CreateProducts
	err = tx.QueryRow(ctx, `
    SELECT p.id, name, p.description, p.rating, p.priceoriginal, p.stock, p.id_product_images,
       COALESCE(pi.photos_one, '') AS photos_one,
       COALESCE(pi.photos_two, '') AS photos_two,
       COALESCE(pi.photos_three, '') AS photos_three,
       COALESCE(pi.photos_four, '') AS photos_four
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

	// --- GET SIZE LIST ---
	rowsSize, err := tx.Query(ctx, `
	SELECT id_size
	FROM size_product
	WHERE id_product = $1
`, body.Id)
	if err != nil {
		return CreateProducts{}, err
	}
	defer rowsSize.Close()

	var sizeIDs []int
	for rowsSize.Next() {
		var sizeID int
		if err := rowsSize.Scan(&sizeID); err != nil {
			return CreateProducts{}, err
		}
		sizeIDs = append(sizeIDs, sizeID)
	}
	product.Size = sizeIDs

	// --- GET VARIANT LIST ---
	rowsVariant, err := tx.Query(ctx, `
	SELECT id_variant
	FROM variant_product
	WHERE id_product = $1
`, body.Id)
	if err != nil {
		return CreateProducts{}, err
	}
	defer rowsVariant.Close()

	var variantIDs []int
	for rowsVariant.Next() {
		var variantID int
		if err := rowsVariant.Scan(&variantID); err != nil {
			return CreateProducts{}, err
		}
		variantIDs = append(variantIDs, variantID)
	}
	product.Variant = variantIDs

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

func GetCountProduct(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	var total int64

	query := "SELECT COUNT(*) FROM product WHERE is_deleted = false"
	args := []interface{}{}

	// --- EXECUTE QUERY ---
	err := db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
