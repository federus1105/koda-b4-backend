package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitCategoriesRouter(router *gin.Engine, db *pgxpool.Pool) {
	categoriesRouter := router.Group("/admin/categories")

	categoriesRouter.GET("", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.GetListCategories(ctx, db)
	})

	categoriesRouter.POST("", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.CreateCategory(ctx, db)
	})
}
