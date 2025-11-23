package models

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type FavoriteProduct struct {
	Id          int     `json:"id"`
	Image       string  `json:"image"`
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
	Discount    float64 `json:"discount"`
	Flash_sale  bool    `json:"flash_sale"`
	Description string  `json:"description"`
}

type Option struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type ProductClient struct {
	ImageOne      *string  `json:"-"`
	ImageTwo      *string  `json:"-"`
	ImageThree    *string  `json:"-"`
	ImageFour     *string  `json:"-"`
	Images        []string `json:"images"`
	Name          string   `json:"name"`
	Price         float64  `json:"price"`
	PriceDiscount float64  `json:"priceDiscount"`
	Rating        string   `json:"rating"`
	Description   string   `json:"desc"`
	Stock         int      `json:"stock"`
	Size          []Option `json:"sizes"`
	Variant       []Option `json:"variant"`
	Flash_sale    bool     `json:"flash_sale"`
}

func GetListFavoriteProduct(ctx context.Context, db *pgxpool.Pool, limit, offset int) ([]FavoriteProduct, error) {
	sql := `SELECT pi.photos_one as image,
	p.id,
	p.name, 
	p.flash_sale,
	p.priceoriginal as price,
    p.pricediscount as discount,
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
		if err := rows.Scan(&fp.Image, &fp.Id, &fp.Name, &fp.Flash_sale, &fp.Price, &fp.Discount, &fp.Description); err != nil {
			return nil, err
		}
		products = append(products, fp)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return products, nil

}

func GetListProductFilter(ctx context.Context, db *pgxpool.Pool, rd *redis.Client,
	name string,
	categoryIDs []int,
	minPrice, maxPrice float64,
	sortBy string,
	limit, offset int) ([]FavoriteProduct, error) {

	// --- REDIS ----
	catStr := ""
	if len(categoryIDs) > 0 {
		strIDs := make([]string, len(categoryIDs))
		for i, id := range categoryIDs {
			strIDs[i] = fmt.Sprintf("%d", id)
		}
		catStr = strings.Join(strIDs, ",")
	}
	redisKey := fmt.Sprintf("product_filter:name=%s&category=%s&min=%.2f&max=%.2f&sort=%s&limit=%d&offset=%d",
		name, catStr, minPrice, maxPrice, sortBy, limit, offset)

	// --- CHECK CACHE ---
	if cached, err := libs.GetFromCache[[]FavoriteProduct](ctx, rd, redisKey); err != nil {
		log.Println("Redis Error:", err)
	} else if cached != nil && len(*cached) > 0 {
		log.Printf("Key %s found in cache Served in using Redis 👌", redisKey)
		return *cached, nil
	}

	sql := `
SELECT DISTINCT p.id,
       p.name,
       p.flash_sale,
	   p.priceoriginal as price,
       p.pricediscount as discount,
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
		if err := rows.Scan(&p.Id, &p.Name, &p.Flash_sale, &p.Price, &p.Discount, &p.Description, &p.Image); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	// --- Save to Cache ---
	if err := libs.SetToCache(ctx, rd, redisKey, products, 5*time.Minute); err != nil {
		log.Println("Redis Error:", err)
	}

	return products, nil
}

func GetProductById(ctx context.Context, db *pgxpool.Pool, productId int) (ProductClient, error) {
	var product ProductClient
	var priceDiscount *float64
	// --- QUERY ----
	err := db.QueryRow(ctx, `
        SELECT p.name, p.flash_sale, p.priceoriginal, p.priceDiscount, p.rating, p.description, p.stock,
               pi.photos_one, pi.photos_two, pi.photos_three, pi.photos_four
        FROM product p
        JOIN product_images pi ON p.id_product_images = pi.id
        WHERE p.id=$1
    `, productId).Scan(
		&product.Name,
		&product.Flash_sale,
		&product.Price,
		&priceDiscount,
		&product.Rating,
		&product.Description,
		&product.Stock,
		&product.ImageOne,
		&product.ImageTwo,
		&product.ImageThree,
		&product.ImageFour,
	)
	if err != nil {
		return ProductClient{}, err
	}

	// -- ASSIGN IF NULL IN DB MAKE IT O ---
	if priceDiscount == nil {
		product.PriceDiscount = 0.0
	} else {
		product.PriceDiscount = *priceDiscount
	}

	// --- APPEND IMAGES TO SLICE IMAGES ---
	product.Images = []string{
		libs.StringOrEmpty(product.ImageOne),
		libs.StringOrEmpty(product.ImageTwo),
		libs.StringOrEmpty(product.ImageThree),
		libs.StringOrEmpty(product.ImageFour),
	}

	// --- GET SIZE ---
	rows, err := db.Query(ctx, `
    SELECT s.id, s.name
    FROM size_product sp
    JOIN sizes s ON sp.id_size = s.id
    WHERE sp.id_product=$1
`, productId)
	if err != nil {
		return ProductClient{}, err
	}
	defer rows.Close()

	product.Size = []Option{}
	for rows.Next() {
		var size Option
		if err := rows.Scan(&size.Id, &size.Name); err != nil {
			return ProductClient{}, err
		}
		product.Size = append(product.Size, size)
	}
	// --- GET VARIANT ----
	rowsVariant, err := db.Query(ctx, `
    SELECT v.id, v.name
    FROM variant_product vp
    JOIN variants v ON vp.id_variant = v.id
    WHERE vp.id_product = $1
`, productId)
	if err != nil {
		return ProductClient{}, err
	}
	defer rowsVariant.Close()

	product.Variant = []Option{}
	for rowsVariant.Next() {
		var variant Option
		if err := rowsVariant.Scan(&variant.Id, &variant.Name); err != nil {
			return ProductClient{}, err
		}
		product.Variant = append(product.Variant, variant)
	}

	return product, nil
}

// --- GET COUNT FOR PRODUCT FAVORITE ---
func GetCountFavoriteProduct(ctx context.Context, db *pgxpool.Pool) (int64, error) {
	var total int64

	query := `
        SELECT COUNT(*)
        FROM product p
        WHERE p.is_favorite = true
        AND p.is_deleted = false
    `

	err := db.QueryRow(ctx, query).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// --- GET COUNT FOR PRODUCT FILTER ---
func GetCountProductFilter(
	ctx context.Context,
	db *pgxpool.Pool,
	name string,
	categoryIDs []int,
	minPrice, maxPrice float64,
) (int64, error) {

	var total int64

	sql := `
        SELECT COUNT(DISTINCT p.id)
        FROM product p
        JOIN product_categories pc ON p.id = pc.id_product
        WHERE p.is_deleted = false
    `
	args := []interface{}{}
	argIdx := 1

	// --- FILTER NAME ---
	if strings.TrimSpace(name) != "" {
		sql += fmt.Sprintf(" AND p.name ILIKE $%d", argIdx)
		args = append(args, "%"+name+"%")
		argIdx++
	}

	// --- FILTER CATEGORY MULTIPLE AND (match all categories) ---
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

	// --- FILTER MIN PRICE ---
	if minPrice > 0 {
		sql += fmt.Sprintf(" AND p.priceOriginal >= $%d", argIdx)
		args = append(args, minPrice)
		argIdx++
	}

	// --- FILTER MAX PRICE ---
	if maxPrice > 0 {
		sql += fmt.Sprintf(" AND p.priceOriginal <= $%d", argIdx)
		args = append(args, maxPrice)
		argIdx++
	}

	// --- EXEC ---
	if err := db.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}
