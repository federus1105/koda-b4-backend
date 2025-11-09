package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitOrderRouter(router *gin.Engine, db *pgxpool.Pool) {
	orderRouter := router.Group("/admin/order")

	orderRouter.POST("/list", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.GetListOrder(ctx, db)
	})

	orderRouter.GET("/:id", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.GetDetailOrder(ctx, db)
	})
}
