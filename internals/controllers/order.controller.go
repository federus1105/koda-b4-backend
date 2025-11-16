package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetListOrder godoc
// @Summary 		Get list orders
// @Description 	Get paginated list of orders with optional filters
// @Tags 		Orders
// @Param 		page 		query 	int 	false 	"Page number" 	default(1)
// @Param 		ordernumber query 	string 	false 	"Filter by order number"
// @Param 		status 		query 	string 	false 	"Filter by status"
// @Success 200 {object} models.ResponseSucces
// @Router /admin/order [get]
// @Security BearerAuth
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

	// --- FILTER STATUS ---
	statusStr := ctx.Query("status")
	status, _ := strconv.Atoi(statusStr)

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- TOTAL ORDER ---
	total, err := models.GetCountOrder(ctxTimeout, db, orderNumber, status)
	if err != nil {
		ctx.JSON(500, gin.H{"success": false, "message": "Failed to get total orders"})
		return
	}
	order, err := models.GetListOrder(ctxTimeout, db, orderNumber, status, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data orders",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- TOTAL ---
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	var prevURL, nextURL *string

	baseURL := "/orders"
	q := url.Values{}
	if orderNumber != "" {
		q.Set("order_number", orderNumber)
	}
	if status != 0 {
		q.Set("status", strconv.Itoa(status))
	}

	// --- PREV ---
	if page > 1 {
		p := page - 1
		u := fmt.Sprintf("%s?%s&page=%d", baseURL, q.Encode(), p)
		prevURL = &u
	}

	// --- NEXT ---
	if page < totalPages {
		n := page + 1
		u := fmt.Sprintf("%s?%s&page=%d", baseURL, q.Encode(), n)
		nextURL = &u
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

	ctx.JSON(200, models.PaginatedResponse[models.OrderList]{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		PrevURL:    prevURL,
		NextURL:    nextURL,
		Result:     order,
	})
}

// GetDetailOrder godoc
// @Summary 		Get detail orders
// @Description 	Get paginated detail of orders
// @Tags 		Orders
// @Param 		id 		path 	int 	true 	"order ID"
// @Success 200 {object} models.ResponseSucces
// @Router /admin/order/{id} [get]
// @Security BearerAuth
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
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "order not found",
			})
			return
		}
		fmt.Println("error :", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to get detail product",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Order detail retrieved successfully",
		Result:  order,
	})
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order by ID
// @Tags Orders
// @Param id path int true "Order ID"
// @Param body body models.UpdateStatusRequest true "Status update info"
// @Success 200 {object} models.ResponseSucces
// @Router /admin/order/{id} [put]
// @Security BearerAuth
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
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "Invalid JSON body",
		})
		fmt.Println(err)
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := models.UpdateOrderStatus(ctxTimeout, db, orderID, body.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "order not found",
			})
			return
		}
		fmt.Println("error :", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to update status",
		})
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
