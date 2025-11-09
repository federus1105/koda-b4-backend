package controllers

import (
	"context"
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

func CreateUser(ctx *gin.Context, db *pgxpool.Pool) {
	var body models.UserBody

	// --- VALIDATION ---
	if err := ctx.ShouldBind(&body); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var msgs []string
			for _, fe := range ve {
				msgs = append(msgs, utils.ErrorUserMsg(fe))
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
				Message: "failed to upload image",
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
				msgs = append(msgs, utils.ErrorUserMsg(fe))
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
				Message: "failed to upload image",
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

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	users, err := models.EditUser(ctxTimeout, db, body, userID)
	if err != nil {
		log.Println("ERROR : ", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "An error occurred while saving data",
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
		Message: "Create User Succesfully",
		Result:  response,
	})

}
