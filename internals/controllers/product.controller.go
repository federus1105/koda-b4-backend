package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetListProduct(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET QUERY PARAMS ---
	page, err := strconv.Atoi(ctx.Query("page"))
	if err != nil {
		fmt.Println(err)
	}
	limit := 10
	offset := (page - 1) * limit
	name := ctx.Query("name")

	//  --- VALIDATION PAGE  ---
	if page == 0 {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "field params page required",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	products, err := models.GetListProduct(ctxTimeout, db, name, limit, offset)
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
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list product",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get data succesfully",
		Result:  products,
	})
}
