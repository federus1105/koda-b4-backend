package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetListUser godoc
// @Summary Get list users
// @Description Get paginated list of users with optional search by name
// @Tags Users
// @Param page query int false "Page number" default(1)
// @Param name query string false "Filter by name"
// @Success 200 {object} models.ResponseSucces
// @Router /admin/user [get]
// @Security BearerAuth
func GetListUser(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET QUERY PARAMS ---
	pageStr := ctx.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit
	name := ctx.Query("name")

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	users, err := models.GetListUser(ctxTimeout, db, name, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data users",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- VALIDATION FOR LIST USERS ---
	if len(users) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list users",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get data succesfully",
		Result:  users,
	})
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Create a new user with optional photo upload
// @Tags         Users
// @Accept       multipart/form-data
// @Param        fullname  formData  string  true   "Full name of the user"
// @Param        email     formData  string  true   "Email of the user"
// @Param        password  formData  string  true   "Password"
// @Param        role      formData  string  true   "Role (admin, user)"
// @Param        phone     formData  string  true  "Phone number"
// @Param        address   formData  string  true  "Address"
// @Param        photos    formData  file    true  "User photo"
// @Success      200 {object} models.ResponseSucces
// @Router       /admin/user [post]
// @Security     BearerAuth
func CreateUser(ctx *gin.Context, db *pgxpool.Pool) {
	var body models.UserBody

	// --- VALIDATION ---
	if err := ctx.ShouldBind(&body); err != nil {
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
	if body.Photos != nil {
		savePath, generatedFilename, err := utils.UploadImageFile(ctx, body.Photos, "public", fmt.Sprintf("user_%d", user.ID))
		if err != nil {
			log.Println("Upload image failed:", err)
			ctx.JSON(400, models.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		if err := ctx.SaveUploadedFile(body.Photos, savePath); err != nil {
			log.Println("Save file failed : ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to save image",
			})
			return
		}
		body.PhotosStr = generatedFilename
	}

	// --- HASHING ---
	hashed, err := libs.HashPassword(body.Password)
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
	newUser, err := models.CreateUser(ctxTimeout, db, hashed, body)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "internal server error",
		})
		return
	}

	// ---- ASIGN RESPONSE ---
	response := gin.H{
		"id":       newUser.Id,
		"fullname": newUser.Fullname,
		"email":    newUser.Email,
		"role":     newUser.Role,
		"photos":   newUser.PhotosStr,
		"address":  newUser.Address,
		"phone":    newUser.Phone,
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Create User Succesfully",
		Result:  response,
	})
}

// EditUser godoc
// @Summary      Edit an existing user
// @Description  Update user details with optional photo upload
// @Tags         Users
// @Accept       multipart/form-data
// @Param        id        path      int     true   "User ID"
// @Param        fullname  formData  string  false  "Full name of the user"
// @Param        email     formData  string  false  "Email of the user"
// @Param        role      formData  string  false  "Role (admin, user)"
// @Param        phone     formData  string  false  "Phone number"
// @Param        address   formData  string  false  "Address"
// @Param        photos    formData  file    false  "User photo"
// @Success      200 {object} models.ResponseSucces
// @Router       /admin/user/{id} [patch]
// @Security     BearerAuth
func EditUser(ctx *gin.Context, db *pgxpool.Pool) {
	var body models.UserUpdateBody

	// --- GET PORDUCT ID ---
	userIDSStr := ctx.Param("id")
	userID, err := strconv.Atoi(userIDSStr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Invalid user id",
		})
		return
	}

	// --- VALIDATION ---
	if err := ctx.ShouldBind(&body); err != nil {
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

	body.Id = userID

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
	if body.Photos != nil {
		savePath, generatedFilename, err := utils.UploadImageFile(ctx, body.Photos, "public", fmt.Sprintf("user_%d", user.ID))
		if err != nil {
			log.Println("Upload image failed:", err)
			ctx.JSON(400, models.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}
		if err := ctx.SaveUploadedFile(body.Photos, savePath); err != nil {
			log.Println("Save file failed : ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to save image",
			})
			return
		}
		body.PhotosStr = &generatedFilename
	}

	// --- CHECKING ROWS UPDATE ---
	if libs.IsStructEmptyExcept(body, "Id") && (body.PhotosStr == nil || *body.PhotosStr == "") {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "No data to update",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	users, err := models.EditUser(ctxTimeout, db, body, userID)
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
		"photos":   users.PhotosStr,
		"address":  users.Address,
		"phone":    users.Phone,
	}
	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Update User Succesfully",
		Result:  response,
	})

}
