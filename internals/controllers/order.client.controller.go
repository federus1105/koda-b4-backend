package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateCartProduct(ctx *gin.Context, db *pgxpool.Pool) {
	var input models.CartItemRequest

	// --- VALIDATION ---
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "invalid Input Json",
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

	// --- LIMIT EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- CALL MODEL FUNCTION ---
	newCartItem, err := models.CreateCartProduct(ctxTimeout, db, userID, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Internal server error",
		})
		log.Println(err.Error())
		return
	}

	// --- SUCCESS RESPONSE ---
	ctx.JSON(http.StatusOK, models.ResponseSucces{
		Success: true,
		Message: "Product added to cart successfully",
		Result:  newCartItem,
	})
}
