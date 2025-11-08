package controllers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(ctx *gin.Context, db *pgxpool.Pool) {
	var req models.AuthRegister
	// --- VALIDATION ---
	if err := ctx.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorRegisterMsg(fe))
			}
			ctx.JSON(400, models.Response{
				Success: false,
				Message: strings.Join(msgs, ", "),
			})
			return
		}

		ctx.JSON(400, models.Response{
			Success: false,
			Message: "invalid JSON format",
		})
		return
	}

	// --- HASHING ---
	hashed, err := libs.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "failed to hash password",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	newUser, err := models.Register(ctxTimeout, db, hashed, req)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "internal server error",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Register Succesfully",
		Result: gin.H{
			"id":       newUser.Id,
			"fullname": newUser.Fullname,
			"email":    newUser.Email,
		},
	})
}

func Login(ctx *gin.Context, db *pgxpool.Pool) {
	var req models.AuthLogin
	// --- VALIDATION ---
	if err := ctx.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorLoginMsg(fe))
			}
			ctx.JSON(400, models.Response{
				Success: false,
				Message: strings.Join(msgs, ", "),
			})
			return
		}

		ctx.JSON(400, models.Response{
			Success: false,
			Message: "invalid JSON format",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	user, err := models.Login(ctxTimeout, db, req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			ctx.JSON(401, models.Response{
				Success: false,
				Message: "Nama atau Password salah",
			})
			return
		}
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "internal server errorr",
		})
		return
	}

	// --- VERIFICATION HASH PASSWORD
	ok, err := libs.VerifyPassword(req.Password, user.Password)
	if err != nil || !ok {
		ctx.JSON(401, models.Response{
			Success: false,
			Message: "invalid email or password",
		})
		return
	}

	// --- GENERATE JWT TOKEN
	claims := libs.NewJWTClaims(req.Id)
	jwtToken, err := claims.GenToken()
	if err != nil {
		fmt.Println("Internal Server Error.\nCause: ", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "internal server errorrr",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "login successful",
		Result: gin.H{
			"token": jwtToken},
	})

}
