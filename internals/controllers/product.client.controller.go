package controllers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetListProductFavorite godoc
// @Summary Get list products Favorite
// @Description Get paginated list of products with pagination
// @Tags Products
// @Param page query int false "Page number" default(1)
// @Success 200 {object} models.ResponseSucces
// @Router /favorite-product [get]
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
	products, err := models.GetListFavoriteProduct(ctxTimeout, db, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data products",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- VALIDATION FOR LIST PRODUCT ---
	if len(products) == 0 {
		ctx.JSON(200, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list product favorite",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get data succesfully",
		Result:  products,
	})
}

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

	products, err := models.GetListProductFilter(ctxTimeout, db, name, categoryIDs, minPrice, maxPrice, sortBy, limit, offset)
	if err != nil {
		ctx.JSON(500, gin.H{
			"success": false,
			"message": "Failed Get list data products",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- VALIDATION FOR LIST PRODUCT ---
	if len(products) == 0 {
		ctx.JSON(200, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list product",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
		"message": "Get data successfully",
		"result":  products,
	})
}
