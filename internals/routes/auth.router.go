package routes

import (
	"github.com/federus1105/koda-b4-backend/internals/controllers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func InitAuthRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	authRouter := router.Group("/auth")

	authRouter.POST("/register", func(ctx *gin.Context) {
		controllers.Register(ctx, db)
	})
	authRouter.POST("/login", func(ctx *gin.Context) {
		controllers.Login(ctx, db)
	})
	authRouter.POST("/forgot-password", func(ctx *gin.Context) {
		controllers.ForgotPassword(ctx, db, rdb)
	})

	authRouter.POST("/reset-password", func(ctx *gin.Context) {
		controllers.ResetPassword(ctx, db, rdb)
	})

}
