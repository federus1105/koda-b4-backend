package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/middlewares"
	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProfileUpdate godoc
// @Summary      Update user profile
// @Description  Update user profile including fullname, phone, address, email, and photo
// @Tags         Users
// @Param        fullname  formData  string  false  "Full name of the user"
// @Param        phone     formData  string  false  "Phone number of the user"
// @Param        address   formData  string  false  "Address of the user"
// @Param        email     formData  string  false  "Email of the user"
// @Param        photos    formData  file    false  "Profile photo file"
// @Success      200  {object}  models.ResponseSucces
// @Router       /profile [patch]
// @Security BearerAuth
func ProfileUpdate(ctx *gin.Context, db *pgxpool.Pool) {
	var input models.ProfileUpdate
	// --- GET USER IN CONTEXT ---
	userIDInterface, exists := ctx.Get(middlewares.UserIDKey)
	if !exists {
		ctx.JSON(401, models.Response{
			Success: false,
			Message: "Unauthorized: user not logged in",
		})
		return
	}

	var userID int
	switch v := userIDInterface.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		ctx.JSON(401, models.Response{
			Success: false,
			Message: "Invalid user ID type in context",
		})
		return
	}

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
			Message: "invalid FORM format",
		})
		return
	}

	// --- CHECKING CLAIMS TOKEN ---
	claims, exists := ctx.Get("claims")
	if !exists {
		fmt.Println("ERROR :", !exists)
		ctx.AbortWithStatusJSON(403, models.Response{
			Success: false,
			Message: "Please log in again",
		})
		return
	}
	user, ok := claims.(libs.Claims)
	if !ok {
		fmt.Println("ERROR", !ok)
		ctx.AbortWithStatusJSON(500, models.Response{
			Success: false,
			Message: "An error occurred!, please try again.",
		})
		return
	}

	// --- UPLOAD PHOTO ---
	if input.Photos != nil {
		savePath, generatedFilename, err := utils.UploadImageFile(ctx, input.Photos, "public", fmt.Sprintf("user_%d", user.ID))
		if err != nil {
			log.Println("Upload image failed:", err)
			ctx.JSON(400, models.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		if err := ctx.SaveUploadedFile(input.Photos, savePath); err != nil {
			log.Println("Save file failed : ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to save image",
			})
			return
		}
		input.PhotosStr = &generatedFilename
	}

	// --- CHECKING ROWS UPDATE ---
	if libs.IsStructEmptyExcept(input, "Id") && (input.PhotosStr == nil || *input.PhotosStr == "") {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "No data to update",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	users, err := models.UpdateProfile(ctxTimeout, db, input, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "user not found",
			})
			return
		}
		fmt.Println("error :", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to update user",
		})
		return
	}

	// ---- ASIGN RESPONSE ---
	response := gin.H{
		"id":       users.Id,
		"fullname": users.Fullname,
		"email":    users.Email,
		"photos":   users.PhotosStr,
		"address":  users.Address,
		"phone":    users.Phone,
	}
	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Update Profile Succesfully",
		Result:  response,
	})

}
