package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitProductRouter(router *gin.Engine, db *pgxpool.Pool) {
	productRouter := router.Group("/admin/product")

	productRouter.GET("", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.GetListProduct(ctx, db)
	})

	productRouter.POST("", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.CreateProduct(ctx, db)
	})

	productRouter.PATCH("/:id", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.EditProduct(ctx, db)
	})

	productRouter.POST("/delete/:id", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.DeleteProduct(ctx, db)
	})
}
