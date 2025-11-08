package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitProductRouter(router *gin.Engine, db *pgxpool.Pool) {
	productRouter := router.Group("/admin/product")

	productRouter.GET("/list", func(ctx *gin.Context) {
		controllers.GetListProduct(ctx, db)
	})
}
