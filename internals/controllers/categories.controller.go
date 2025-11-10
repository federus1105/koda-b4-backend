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
