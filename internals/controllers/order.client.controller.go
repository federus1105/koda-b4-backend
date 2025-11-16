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

// CreateCartProduct godoc
// @Summary Add product to cart
// @Description Adding products to the cart of the logged in user
// @Tags Cart
// @Param request body models.CartItemRequest true "Request Body"
// @Success 200 {object} models.ResponseSucces "Product added to cart successfully"
// @Failure 400 {object} models.Response "Invalid JSON, validation error, out of stock, or insufficient stock"
// @Failure 401 {object} models.Response "Unauthorized: user not logged in"
// @Failure 404 {object} models.Response "Product not found"
// @Failure 500 {object} models.Response "Internal server error"
// @Router /cart [post]
// @Security BearerAuth
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
		errMsg := err.Error()

		// -- MESSAGES ERROR --
		switch {
		case strings.Contains(errMsg, "product is out of stock"):
			ctx.JSON(400, models.Response{
				Success: false,
				Message: "product is out of stock",
			})
		case strings.Contains(errMsg, "insufficient stock"):
			ctx.JSON(400, models.Response{
				Success: false,
				Message: errMsg,
			})
		case strings.Contains(errMsg, "product Not found"):
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "product not found",
			})
		default:
			// Jika bukan error validasi user, berarti error sistem
			ctx.JSON(500, models.Response{
				Success: false,
				Message: "internal server error",
			})
			log.Println(errMsg)
		}
		return
	}

	// --- SUCCESS RESPONSE ---
	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Product added to cart successfully",
		Result:  newCartItem,
	})
}

// GetCartProduct godoc
// @Summary Get cart products
// @Description Gets a list of products in the cart of the logged in user.
// @Tags Cart
// @Success 200 {object} models.ResponseSucces "Cart data retrieved successfully"
// @Failure 401 {object} models.Response "User ID not found in context"
// @Failure 500 {object} models.Response "Failed get data carts"
// @Router /cart [get]
// @Security BearerAuth
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

// Transactions godoc
// @Summary Process a transaction
// @Description Performs a transaction for the authenticated user. Includes validation and business logic checks.
// @Tags Transactions
// @Param request body utils.RequestTransactions true "Transaction Request Body"
// @Success 200 {object} models.ResponseSucces "Transaction completed successfully"
// @Failure 400 {object} models.Response "Validation error or invalid JSON format"
// @Failure 401 {object} models.Response "Unauthorized: user not logged in"
// @Failure 500 {object} models.Response "Internal server error"
// @Router /transactions [post]
// @Security BearerAuth
func Transactions(ctx *gin.Context, db *pgxpool.Pool) {
	var input models.TransactionsInput

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
	result, err := models.Transactions(ctxTimeout, db, input, userID)
	if err != nil {
		var ve utils.ValidationError
		if errors.As(err, &ve) {
			ctx.JSON(400, models.Response{
				Success: false,
				Message: ve.Error(),
			})
			return
		}
		ctx.JSON(500, models.Response{
			Success: false,
			Message: err.Error(),
		})
		log.Println(err.Error())
		return
	}
	// --- SUCCESS RESPONSE ---
	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Transaction completed successfully",
		Result:  result,
	})
}
