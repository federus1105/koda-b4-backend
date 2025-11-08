package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitProductRouter(router *gin.Engine, db *pgxpool.Pool) {
	productRouter := router.Group("/admin/product")

	productRouter.GET("/list", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.GetListProduct(ctx, db)
	})

	productRouter.POST("/create", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.CreateProduct(ctx, db)
	})

	productRouter.PATCH("/update/:id", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.EditProduct(ctx, db)
	})

	productRouter.POST("/delete/:id", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.DeleteProduct(ctx, db)
	})
}
