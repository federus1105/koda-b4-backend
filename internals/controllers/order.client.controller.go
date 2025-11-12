package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateCartProduct(ctx *gin.Context, db *pgxpool.Pool) {
	var input models.CartItemRequest

	// --- VALIDATION ---
	if err := ctx.ShouldBindJSON(&input); err != nil {
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
			Message: "invalid JSON format",
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
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Internal server error",
		})
		log.Println(err.Error())
		return
	}

	// --- SUCCESS RESPONSE ---
	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Product added to cart successfully",
		Result:  newCartItem,
	})
}

func GetCartProduct(ctx *gin.Context, db *pgxpool.Pool) {
	userIDRaw, exists := ctx.Get(middlewares.UserIDKey)
	// --- CHECKING IN CONTEXT ---
	if !exists {
		ctx.JSON(401, models.Response{
			Success: false,
			Message: "User ID not found in context",
		})
		return
	}

	userID, ok := userIDRaw.(int)
	if !ok {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "User ID in context is invalid",
		})
		return
	}

	// --- LIMIT EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	carts, err := models.GetCartProduct(ctxTimeout, db, userID)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed get data carts",
		})
		fmt.Println(err.Error())
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Cart data retrieved successfully",
		Result:  carts,
	})
}
