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

func GetListOrder(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET QUERY PARAMS ---
	pageStr := ctx.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// --- ARGUMEN ---
	limit := 10
	offset := (page - 1) * limit
	orderNumber := ctx.Query("ordernumber")
	status := ctx.Query("status")

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	order, err := models.GetListOrder(ctxTimeout, db, orderNumber, status, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data orders",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- VALIDATION FOR LIST ORDER ---
	if len(order) == 0 {
		ctx.JSON(200, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list order",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get data succesfully",
		Result:  order,
	})
}

func GetDetailOrder(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET ORDER ID ---
	orderIDStr := ctx.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}
	// ---- LIMITS QUERY EXECUTION TIME --
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- GET DETAIL ORDER ---
	order, err := models.GetDetailOrder(ctxTimeout, db, orderID)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to fetch order details",
		})
		fmt.Println("Error:", err)
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Order detail retrieved successfully",
		Result:  order,
	})
}

func UpdateOrderStatus(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET ORDER ID ---
	orderIDStr := ctx.Param("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}

	var body models.UpdateStatusRequest
	if err := ctx.BindJSON(&body); err != nil {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "Invalid JSON body",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := models.UpdateOrderStatus(ctxTimeout, db, orderID, body.Status); err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to update order status",
		})
		fmt.Println("Error:", err)
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Order status updated successfully",
		Result: gin.H{
			"productID": orderID,
			"newStatus": body.Status},
	})
}
