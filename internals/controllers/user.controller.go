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

func GetListUser(ctx *gin.Context, db *pgxpool.Pool) {
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
	users, err := models.GetListUser(ctxTimeout, db, name, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data users",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- VALIDATION FOR LIST USERS ---
	if len(users) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list users",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get data succesfully",
		Result:  users,
	})
}
