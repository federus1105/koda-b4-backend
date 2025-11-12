package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitOrderClientRoutes(router *gin.Engine, db *pgxpool.Pool) {
	InitOrderClientRoutes := router.Group("")

	InitOrderClientRoutes.POST("/cart", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.CreateCartProduct(ctx, db)
	})

	InitOrderClientRoutes.GET("/cart", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.GetCartProduct(ctx, db)
	})

	InitOrderClientRoutes.POST("/transactions", middlewares.VerifyToken, middlewares.AuthMiddleware(), func(ctx *gin.Context) {
		controllers.Transactions(ctx, db)
	})
}
