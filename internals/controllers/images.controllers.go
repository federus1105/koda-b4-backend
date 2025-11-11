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

func GetListImageById(ctx *gin.Context, db *pgxpool.Pool) {
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
	images, err := models.GetListImageById(ctxTimeout, db, productID)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data images",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- VALIDATION FOR LIST PRODUCT ---
	if len(images) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list images",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get data succesfully",
		Result:  images,
	})
}
