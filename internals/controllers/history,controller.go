package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetHistory godoc
// @Summary Get user transaction/history list
// @Description Retrieves a history list of the currently logged in user with filters and pagination.
// @Tags History
// @Param month query int false "Filter bulan (1-12)"
// @Param status query int false "Filter status history"
// @Param page query int false "Page number (default: 1)"
// @Success 200 {object} models.ResponseSucces
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /history [get]
// @Security BearerAuth
func GetHistory(ctx *gin.Context, db *pgxpool.Pool) {

	// --- GET USER IN CONTEXT ---
	userIDInterface, exists := ctx.Get(middlewares.UserIDKey)
	if !exists {
		ctx.JSON(401, models.Response{
			Success: false,
			Message: "Unauthorized: user not logged in",
		})
		return
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		ctx.JSON(401, models.Response{
			Success: false,
			Message: "Invalid user ID type in context",
		})
		return
	}

	// --- FILTER MONTH ---
	monthStr := ctx.Query("month")
	month, _ := strconv.Atoi(monthStr)

	// --- FILTER STATUS ---
	statusStr := ctx.Query("status")
	status, _ := strconv.Atoi(statusStr)

	// --- GET QUERY PARAMS ---
	pageStr := ctx.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 5
	offset := (page - 1) * limit

	// --- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- GET TOTAL HISTORY ---
	total, err := models.GetCountHistory(ctxTimeout, db, userID, month, status)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to get total history count",
		})
		return
	}

	histories, err := models.GetHistory(ctxTimeout, db, userID, month, status, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data histories",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	var prevURL *string
	var nextURL *string

	// --- TOTAL PAGES ---
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// --- BUILD QUERY PARAMS FOR URL ---
	q := url.Values{}
	if month > 0 {
		q.Add("month", strconv.Itoa(month))
	}
	if status > 0 {
		q.Add("status", strconv.Itoa(status))
	}

	baseURL := "/history"

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
	if len(histories) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list history",
		})
		return
	}
	ctx.JSON(200, models.PaginatedResponse[models.History]{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		PrevURL:    prevURL,
		NextURL:    nextURL,
		Result:     histories,
	})
}

// DetailHistory godoc
// @Summary Get detail history
// @Description Retrieve history details by ID (user must be logged in)
// @Tags History
// @Param id path int true "History ID"
// @Success 200 {object} models.ResponseSucces "Success"
// @Failure 401 {object} models.Response "Unauthorized"
// @Failure 404 {object} models.Response "Not Found"
// @Failure 500 {object} models.Response "Internal Server Error"
// @Router /history/{id} [get]
// @Security BearerAuth
func DetailHistory(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET HISTORY ID ---
	historyIDStr := ctx.Param("id")
	historyID, err := strconv.Atoi(historyIDStr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Invalid order ID",
		})
		return
	}

	// --- GET USER IN CONTEXT ---
	userIDInterface, exists := ctx.Get(middlewares.UserIDKey)
	if !exists {
		ctx.JSON(401, models.Response{
			Success: false,
			Message: "Unauthorized: user not logged in",
		})
		return
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		ctx.JSON(401, models.Response{
			Success: false,
			Message: "Invalid user ID type in context",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME --
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- GET DETAIL  HISTORY---
	history, err := models.DetailHistory(ctxTimeout, db, userID, historyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "history not found",
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
		Message: "history detail retrieved successfully",
		Result:  history,
	})
}
