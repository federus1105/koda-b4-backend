package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetListCategories godoc
// @Summary      Get list of categories
// @Description  Retrieve a paginated list of categories, optionally filtered by name.
// @Tags         Categories
// @Param        page  query     int     false  "Page number for pagination (default: 1)"
// @Param        name  query     string  false  "Filter categories by name"
// @Success      200   {object}  models.ResponseSucces  "Get data successfully"
// @Router       /admin/categories [get]
// @Security BearerAuth
func GetListCategories(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET QUERY PARAMS ---
	pageStr := ctx.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit
	name := ctx.Query("name")

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- GET TOTAL CATEGORIES ---
	total, err := models.GetCountCategories(ctxTimeout, db)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to get total categories count",
		})
		return
	}

	categories, err := models.GetListCategories(ctxTimeout, db, name, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data categories",
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
	if len(categories) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found favorite product",
		})
		return
	}
	ctx.JSON(200, models.PaginatedResponse[models.Categories]{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		PrevURL:    prevURL,
		NextURL:    nextURL,
		Result:     categories,
	})
}

// CreateCategory godoc
// @Summary      Create a new category
// @Description  Create a new category with the provided data.
// @Tags         Categories
// @Param        category  body utils.CategoriesRequest  true  "Category data"
// @Success      200  {object}  models.ResponseSucces  "Create Categories Successfully"
// @Router       /admin/categories [post]
// @Security BearerAuth
func CreateCategory(ctx *gin.Context, db *pgxpool.Pool) {
	var input models.Categories

	// --- VALIDATION ---
	if err := ctx.ShouldBind(&input); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorMessage(fe))
			}
			ctx.JSON(400, models.Response{
				Success: false,
				Message: strings.Join(msgs, ", "),
			})
			return
		}

		ctx.JSON(400, models.Response{
			Success: false,
			Message: "invalid Form format",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	newCategory, err := models.CreateCategories(ctxTimeout, db, input)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Create Categories Successfully",
		Result:  newCategory,
	})
}

// UpdateCategories godoc
// @Summary      Update category by ID
// @Description  Update an existing category using its ID.
// @Tags         Categories
// @Param        id        path      int                     true   "Category ID"
// @Param        category   body      utils.CategoriesRequest  true   "Updated category data"
// @Success      200        {object}  models.ResponseSucces    "Update Categories Successfully"
// @Router       /admin/categories/{id} [put]
// @Security BearerAuth
func UpdateCategories(ctx *gin.Context, db *pgxpool.Pool) {
	var body models.Categories
	// --- GET CATEGORIES ID ---
	categoryIDstr := ctx.Param("id")
	categoriesID, err := strconv.Atoi(categoryIDstr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Invalid categories id",
		})
		return
	}

	// --- VALIDATION ---
	if err := ctx.ShouldBind(&body); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorMessage(fe))
			}
			ctx.JSON(400, models.Response{
				Success: false,
				Message: strings.Join(msgs, ", "),
			})
			return
		}

		ctx.JSON(400, models.Response{
			Success: false,
			Message: "invalid Form format",
		})
		return
	}
	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	categories, err := models.UpdateCategories(ctxTimeout, db, body, categoriesID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "categories not found",
			})
			return
		}
		fmt.Println("error :", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to update categories",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Update Categories Succesfully",
		Result:  categories,
	})

}

// DeleteCategories godoc
// @Summary      Delete category by ID
// @Description  Delete a category based on its ID. Returns 404 if the category is not found.
// @Tags         Categories
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  models.ResponseSucces  "Delete categories successfully"
// @Router       /admin/categories/{id} [delete]
// @Security BearerAuth
func DeleteCategories(ctx *gin.Context, db *pgxpool.Pool) {
	categoryIDstr := ctx.Param("id")
	categoryID, err := strconv.Atoi(categoryIDstr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "categories id not found",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = models.DeleteCategories(ctxTimeout, db, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "categories not found",
			})
			return
		}
		fmt.Println("error :", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to delete categories",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Delete categories successfully",
		Result:  fmt.Sprintf("categories id %d", categoryID),
	})
}
