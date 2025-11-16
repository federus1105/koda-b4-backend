package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetListProductFavorite godoc
// @Summary Get list products Favorite
// @Description Get paginated list of products with pagination
// @Tags Products
// @Param page query int false "Page number" default(1)
// @Success 200 {object} models.ResponseSucces
// @Router /favorite-product [get]
// @Security BearerAuth
func GetListFavoriteProduct(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET QUERY PARAMS ---
	pageStr := ctx.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 3
	offset := (page - 1) * limit

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- GET TOTAL COUNT ---
	total, err := models.GetCountFavoriteProduct(ctxTimeout, db)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to get total product count",
		})
		return
	}

	products, err := models.GetListFavoriteProduct(ctxTimeout, db, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data products",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	var prevURL *string
	var nextURL *string

	// --- TOTAL PAGES ---
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// --- QUERY PARAMS ---

	baseURL := "/favorite-product"
	// --- PREV ---
	if page > 1 {
		url := fmt.Sprintf("%s?page=%d", baseURL, page-1)
		prevURL = &url
	}

	// --- NEXT ---
	if page < totalPages {
		url := fmt.Sprintf("%s?page=%d", baseURL, page+1)
		nextURL = &url
	}

	// --- VALIDATION FOR LIST PRODUCT ---
	if len(products) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found favorite product",
		})
		return
	}
	ctx.JSON(200, models.PaginatedResponse[models.FavoriteProduct]{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		PrevURL:    prevURL,
		NextURL:    nextURL,
		Result:     products,
	})
}

// GetListProductFilter godoc
// @Summary Get filtered product list
// @Description Retrieves a list of products using filters such as name, category, price range, sorting, and pagination.
// @Tags Products
// @Param page query int false "Page number (default: 1)"
// @Param name query string false "Search by product name"
// @Param category query []int false "Filter by category IDs (can be multiple)"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param sort_by query string false "Sort by criteria (price_asc, price_desc, latest, oldest)"
// @Success 200 {object} models.ResponseSucces "Successful response with product list"
// @Failure 500 {object} models.Response "Failed to retrieve product list"
// @Router /product [get]
// @Security BearerAuth
func GetListProductFilter(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET QUERY PARAMS ---
	pageStr := ctx.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit
	name := ctx.Query("name")

	// --- CATEGORY FILTER ----
	categoryIDsStr := ctx.QueryArray("category")
	var categoryIDs []int
	for _, idStr := range categoryIDsStr {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			categoryIDs = append(categoryIDs, id)
		}
	}

	// --- PRICE RANGE FILTER ---
	minPriceStr := ctx.Query("min_price")
	maxPriceStr := ctx.Query("max_price")
	var minPrice, maxPrice float64
	if minPriceStr != "" {
		minPrice, _ = strconv.ParseFloat(minPriceStr, 64)
	}
	if maxPriceStr != "" {
		maxPrice, _ = strconv.ParseFloat(maxPriceStr, 64)
	}

	// --- SORTING ---
	sortBy := ctx.Query("sort_by")

	// --- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- GET TOTAL PRODUCT ---
	total, err := models.GetCountProductFilter(ctxTimeout, db, name, categoryIDs, minPrice, maxPrice)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to get total product count",
		})
		return
	}

	products, err := models.GetListProductFilter(ctxTimeout, db, name, categoryIDs, minPrice, maxPrice, sortBy, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data products",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- TOTAL PAGES ---
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// --- BUILD QUERY PARAMS FOR URL ---
	q := url.Values{}
	if name != "" {
		q.Add("name", name)
	}
	for _, id := range categoryIDs {
		q.Add("category", strconv.Itoa(id))
	}
	if minPriceStr != "" {
		q.Add("min_price", minPriceStr)
	}
	if maxPriceStr != "" {
		q.Add("max_price", maxPriceStr)
	}
	if sortBy != "" {
		q.Add("sort_by", sortBy)
	}

	baseURL := "/product"
	var prevURL *string
	var nextURL *string

	// --- PREV URL ---
	if page > 1 {
		p := page - 1
		url := fmt.Sprintf("%s?%s&page=%d", baseURL, q.Encode(), p)
		prevURL = &url
	}

	// --- NEXT URL ---
	if page < totalPages {
		n := page + 1
		url := fmt.Sprintf("%s?%s&page=%d", baseURL, q.Encode(), n)
		nextURL = &url
	}

	// --- VALIDATION FOR LIST PRODUCT ---
	if len(products) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list product",
		})
		return
	}
	ctx.JSON(200, models.PaginatedResponse[models.FavoriteProduct]{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		PrevURL:    prevURL,
		NextURL:    nextURL,
		Result:     products,
	})
}

// GetProductById godoc
// @Summary Get product by ID
// @Description Retrieves detailed information about a product using its ID.
// @Tags Products
// @Param id path int true "Product ID"
// @Success 200 {object} models.ResponseSucces "Product retrieved successfully"
// @Failure 404 {object} models.Response "Product not found or invalid product ID"
// @Failure 500 {object} models.Response "Failed to retrieve product"
// @Router /product/{id} [get]
// @Security BearerAuth
func GetProductById(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET PORDUCT ID ---
	productIDstr := ctx.Param("id")
	productID, err := strconv.Atoi(productIDstr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Invalid product id",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	product, err := models.GetProductById(ctxTimeout, db, productID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "product not found",
			})
			return
		}
		log.Println("Failed to get product:", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "failed to get product",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get Data succesfully",
		Result:  product,
	})
}
