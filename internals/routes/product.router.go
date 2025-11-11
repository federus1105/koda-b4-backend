package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func InitProductRouter(router *gin.Engine, db *pgxpool.Pool, rd *redis.Client) {
	productRouter := router.Group("/admin/product")
	productRouterother := router.Group("/")
	productRouterFilter := router.Group("/product")

	productRouter.GET("", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.GetListProduct(ctx, db, rd)
	})

	productRouter.POST("", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.CreateProduct(ctx, db, rd)
	})

	productRouter.PATCH("/:id", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.EditProduct(ctx, db)
	})

	productRouter.POST("/delete/:id", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.DeleteProduct(ctx, db)
	})

	productRouter.GET("/:id/images", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.GetListImageById(ctx, db)
	})

	// ============ CLIENT ROUTER =========

	productRouterother.GET("favorite-product", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.GetListFavoriteProduct(ctx, db)
	})

	productRouterFilter.GET("", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.GetListProductFilter(ctx, db)
	})
}
