package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
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
// @Tags         Profile
// @Param        fullname  formData  string  false  "Full name of the user"
// @Param        phone     formData  string  false  "Phone number of the user"
// @Param        address   formData  string  false  "Address of the user"
// @Param        email     formData  string  false  "Email of the user"
// @Param        photos    formData  file    false  "Profile photo file"
// @Success      200  {object}  models.ResponseSucces
// @Router       /profile [patch]
// @Security BearerAuth
func ProfileUpdate(ctx *gin.Context, db *pgxpool.Pool, cld *cloudinary.Cloudinary) {
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

	useCloudinary := os.Getenv("CLOUDINARY_URL") != ""
	// --- UPLOAD PHOTO ---
	if input.Photos != nil {
		fileID := fmt.Sprintf("user_%d", user.ID)
		if !useCloudinary {

			// === LOCAL UPLOAD ===
			savePath, generatedFilename, err := utils.UploadImageFile(
				ctx,
				input.Photos,
				"public",
				fileID,
			)
			if err != nil {
				ctx.JSON(400, models.Response{
					Success: false,
					Message: err.Error(),
				})
				return
			}

			if err := ctx.SaveUploadedFile(input.Photos, savePath); err != nil {
				ctx.JSON(500, gin.H{
					"success": false,
					"error":   "failed to save image",
				})
				return
			}

			input.PhotosStr = &generatedFilename

		} else {
			// === CLOUDINARY UPLOAD ===
			overwrite := true
			uploadResp, err := cld.Upload.Upload(ctx, input.Photos, uploader.UploadParams{
				Folder:    "assets/user",
				PublicID:  fileID,
				Overwrite: &overwrite,
			})

			if err != nil {
				ctx.JSON(500, gin.H{
					"success": false,
					"message": "failed to upload image to cloudinary",
				})
				return
			}

			url := uploadResp.SecureURL
			input.PhotosStr = &url
		}
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

// UpdatePasswordHandler godoc
// @Summary      Update user password
// @Description  Update password for the logged-in user
// @Tags         Profile
// @Param        body  body      models.ReqUpdatePassword  true  "Password update data"
// @Success      200   {object}  models.ResponseSucces
// @Router       /profile [put]
// @Security BearerAuth
func UpdatePassword(ctx *gin.Context, db *pgxpool.Pool) {
	var input models.ReqUpdatePassword

	//  --- VALIDATION ---
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

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := models.UpdatePassword(ctxTimeout, db, userID, input.OldPassword, input.ConfirmPassword); err != nil {
		if errors.Is(err, utils.ErrValidation) {
			ctx.JSON(400, models.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Internal server error",
		})
		fmt.Println(err)
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "password update successfully",
		Result: gin.H{
			"ID_user": userID,
		},
	})

}

// Profile godoc
// @Summary Get user profile
// @Description Get user profile data based on the currently logged in user.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ResponseSucces "get Profile successfully"
// @Failure 401 {object} models.Response "Unauthorized: user not logged in"
// @Failure 500 {object} models.Response "Internal server error"
// @Router /profile [get]
func Profile(ctx *gin.Context, db *pgxpool.Pool) {
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

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	profile, err := models.Profile(ctxTimeout, db, userID)
	if err != nil {
		fmt.Println("error :", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "internal server error",
		})
		return
	}
	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "get Profile Succesfully",
		Result:  profile,
	})
}
