package middlewares

import (
	"slices"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/gin-gonic/gin"
)

func Access(roles ...string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		// ambil data claim
		claims, isExist := ctx.Get("claims")
		if !isExist {
			ctx.AbortWithStatusJSON(403, gin.H{
				"success": false,
				"error":   "Please log in again",
			})
			return
		}
		user, ok := claims.(libs.Claims)
		if !ok {
			ctx.AbortWithStatusJSON(500, models.Response{
				Success: false,
				Message: "Internal server error",
			})
			return
		}
		if !slices.Contains(roles, user.Role) {
			ctx.AbortWithStatusJSON(403, gin.H{
				"success": false,
				"error":   "You do not have access rights to this resource.",
			})
			return
		}
		ctx.Next()
	}
}
