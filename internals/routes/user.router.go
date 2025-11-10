package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitUserRoute(router *gin.Engine, db *pgxpool.Pool) {
	userRouter := router.Group("/admin/user")

	userRouter.GET("", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.GetListUser(ctx, db)
	})

	userRouter.POST("", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.CreateUser(ctx, db)
	})

	userRouter.PATCH("/:id", middlewares.VerifyToken, middlewares.Access("admin"), func(ctx *gin.Context) {
		controllers.EditUser(ctx, db)
	})
}
