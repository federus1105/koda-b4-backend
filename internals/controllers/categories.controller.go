package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	categories, err := models.GetListCategories(ctxTimeout, db, name, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data categories",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- VALIDATION FOR LIST CATEGORIES ---
	if len(categories) == 0 {
		ctx.JSON(200, models.ResponseSucces{
			Success: true,
			Message: "Not found list categories",
			Result:  []string{},
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get data succesfully",
		Result:  categories,
	})
}

func CreateCategory(ctx *gin.Context, db *pgxpool.Pool) {
	var input models.Categories

	// --- VALIDATION ---
	if err := ctx.ShouldBind(&input); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorUserMsg(fe))
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
				msgs = append(msgs, utils.ErrorUserMsg(fe))
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
