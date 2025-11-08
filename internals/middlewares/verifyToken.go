package middlewares

import (
	"fmt"
	"log"
	"strings"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func VerifyToken(ctx *gin.Context) {
	// --- GET TOKEN FROM HEADER ---
	bearerToken := ctx.GetHeader("Authorization")
	// --- BEARER TOKEN
	token := strings.Split(bearerToken, " ")[1]
	if token == "" {
		ctx.AbortWithStatusJSON(401, models.Response{
			Success: false,
			Message: "Please log in first",
		})
		return
	}

	var claims libs.Claims
	if err := claims.VerifyToken(token); err != nil {
		if strings.Contains(err.Error(), jwt.ErrTokenInvalidIssuer.Error()) {
			log.Println("JWT Error.\nCause: ", err.Error())
			ctx.AbortWithStatusJSON(401, models.Response{
				Success: false,
				Message: "Please log in again",
			})
			return
		}
		// --- TOKEN EXPIRE
		if strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) {
			log.Println("JWT Error.\nCause: ", err.Error())
			ctx.AbortWithStatusJSON(401, models.Response{
				Success: false,
				Message: "Token expired, please log in again!",
			})
			return
		}
		fmt.Println(jwt.ErrTokenExpired)
		log.Println("Internal Server Error.\nCause: ", err.Error())
		ctx.AbortWithStatusJSON(500, models.Response{
			Success: false,
			Message: "Internal Server Error",
		})
		return
	}
	ctx.Set("claims", claims)
	ctx.Next()
}
