package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(ctx *gin.Context) {
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	origin := ctx.GetHeader("Origin")

	if origin == allowedOrigin {
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Credentials", "true")
	}

	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
	ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.Next()
}
