package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitUserRoute(router *gin.Engine, db *pgxpool.Pool) {
	userRouter := router.Group("/admin/user")

	userRouter.GET("/list", middlewares.VerifyToken, func(ctx *gin.Context) {
		controllers.GetListUser(ctx, db)
	})
}
