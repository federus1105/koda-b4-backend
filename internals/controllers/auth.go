package controllers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Register godoc
// @Summary 	Register user
// @Description Register user and Hash Password
// @Tags 		Auth
// @Param 		Register body 	utils.RegisterRequest  true 	"Register Info"
// @Success 	200 {object} 	models.ResponseSucces
// @Router 		/auth/register [post]
func Register(ctx *gin.Context, db *pgxpool.Pool) {
	var req models.AuthRegister
	// --- VALIDATION ---
	if err := ctx.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorMessage(fe))
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

// Login godoc
// @Summary 	Login user
// @Description Login user and get JWT token
// @Tags 		Auth
// @Param 		login body 		utils.LoginRequest  true 	"Login Info"
// @Success 	200 {object} 	models.ResponseSucces
// @Router 		/auth/login [post]
func Login(ctx *gin.Context, db *pgxpool.Pool) {
	var req models.AuthLogin
	// --- VALIDATION ---
	if err := ctx.ShouldBindJSON(&req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorMessage(fe))
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
	claims := libs.NewJWTClaims(user.Id, user.Role)
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

// ForgotPassword godoc
// @Summary Send password reset link to user email
// @Description Receive user email, create password reset token, store it in Redis, and send email containing password reset link
// @Tags Auth
// @Param input body models.ReqForgot true "Email user"
// @Success 200 {object} models.Response "Reset link successfully sent"
// @Router /auth/forgot-password [post]
func ForgotPassword(ctx *gin.Context, db *pgxpool.Pool, rdb *redis.Client) {
	var input models.ReqForgot

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "invalid email",
		})
		return
	}

	// --- GET USER BY EMAIL ---
	user, err := models.GetUserByEmail(ctx, db, input.Email)
	if err != nil {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "user not found",
		})
		return
	}

	// --- GENEATE RANDOM TOKEN ---
	token, _ := utils.GenerateRandomToken(32)
	key := "reset:pwd:" + token

	// --- SAVER RANDOM TOKEN IN REDIS --
	if err := models.SaveResetToken(ctx.Request.Context(), rdb, key, fmt.Sprintf("%d", user.Id), 15*time.Minute); err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "failed to save token",
		})
		fmt.Println(err)
		return
	}

	// --- URL ---
	frontendURL := os.Getenv("FRONTEND_RESET_URL")
	if frontendURL == "" {
		frontendURL = "localhost:8011/auth/reset-password"
	}

	resetLink := fmt.Sprintf("%s?token=%s", frontendURL, token)

	// --- MESSAGE EMAIL ---
	emailBody := fmt.Sprintf(`
    <div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
        <h2 style="color: #4CAF50;">Reset Password</h2>
        <p>Halo,</p>
        <p>Anda atau seseorang telah meminta reset password untuk akun Anda. Gunakan token berikut untuk melakukan reset password:</p>
        <p style="font-size: 20px; font-weight: bold; color: #000;">%s</p>
        <p>Atau klik tombol berikut untuk langsung ke halaman reset password:</p>
        <p>
            <a href="%s" style="display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: #fff; text-decoration: none; border-radius: 5px;">
                Reset Password
            </a>
        </p>
        <p>Jika Anda tidak meminta reset password, abaikan email ini.</p>
        <p>Salam,<br/>Tim Senja Kopi kiri</p>
    </div>
`, token, resetLink)

	err = utils.Send(utils.SendOptions{
		To:         []string{input.Email},
		Subject:    "Reset Password",
		Body:       emailBody,
		BodyIsHTML: true,
	})

	// --- ERROR HANDLING ---
	if err != nil {
		fmt.Println("SMTP ERROR:", err)
	} else {
		fmt.Println("Email sent to:", input.Email)
	}

	ctx.JSON(200, gin.H{
		"success":    true,
		"message":    "Reset link sent to email",
		"reset_link": resetLink,
	})
}

// ResetPassword godoc
// @Summary Reset user password
// @Description Resets the password for a user using a valid token. The token must have been issued during the forgot password process.
// @Tags Auth
// @Param input body models.ReqResetPassword true "Reset password request"
// @Success 200 {object} models.ResponseSucces "Password updated successfully"
// @Router /auth/reset-password [post]
func ResetPassword(ctx *gin.Context, db *pgxpool.Pool, rdb *redis.Client) {
	var input models.ReqResetPassword

	// --- VALIDATION ---
	if err := ctx.ShouldBind(&input); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorMessage(fe))
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

	// --- GET USER ID FROM TOKEN ---
	key := "reset:pwd:" + input.Token
	userIDStr, err := models.GetUserIDFromToken(ctx, rdb, key)
	if err != nil {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "invalid or expired token",
		})
		return
	}

	userID, _ := strconv.Atoi(userIDStr)

	// --- HASH NEW PASSWORD ---
	hashed, err := libs.HashPassword(input.NewPassword)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "failed to hash password",
		})
		return
	}

	// --- UPDATE PASSWORD BY USER ID ---
	if err := models.UpdatePasswordByID(ctx, db, userID, hashed); err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "failed to update password",
		})
		return
	}

	// --- DELETE TOKEN AFTER USED ---
	rdb.Del(ctx, key)

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "password updated successfully",
		Result: gin.H{
			"iduser": userID,
		},
	})
}
