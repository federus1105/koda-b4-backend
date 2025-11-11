package controllers

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetListImageById(ctx *gin.Context, db *pgxpool.Pool) {
	// --- GET PORDUCT ID ---
	productIDstr := ctx.Param("id")
	productID, err := strconv.Atoi(productIDstr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Invalid images id",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	images, err := models.GetListImageById(ctxTimeout, db, productID)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data images",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	// --- VALIDATION FOR LIST PRODUCT ---
	if len(images) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list images",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Get data succesfully",
		Result:  images,
	})
}

func CreateImagesbyId(ctx *gin.Context, db *pgxpool.Pool) {
	var body models.ImagesBody
	// --- GET PORDUCT ID ---
	productIDstr := ctx.Param("id")
	productID, err := strconv.Atoi(productIDstr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "Invalid product id",
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

	// --- UPLOAD IMAGES ---
	imageFiles := map[string]*multipart.FileHeader{
		"photos_one":   body.ImagesOne,
		"photos_two":   body.ImagesTwo,
		"photos_three": body.ImagesThree,
		"photos_four":  body.ImagesFour,
	}

	imageStrs := map[string]string{}

	// --- MULTIPLE UPLOAD IMAGES ---
	for key, file := range imageFiles {
		if file == nil {
			continue
		}
		savePath, generatedFilename, err := utils.UploadImageFile(ctx, file, "public", fmt.Sprintf("%s_%d", key, user.ID))
		if err != nil {
			ctx.JSON(400, models.Response{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		if err := ctx.SaveUploadedFile(file, savePath); err != nil {
			ctx.JSON(500, models.Response{
				Success: false,
				Message: fmt.Sprintf("Failed to save %s", key),
			})
			return
		}

		imageStrs[key] = generatedFilename
	}

	// --- ASSIGN TO BODY ---
	body.ImagesOneStr = imageStrs["photos_one"]
	body.ImagesTwoStr = imageStrs["photos_two"]
	body.ImagesThreeStr = imageStrs["photos_three"]
	body.ImagesFourStr = imageStrs["photos_four"]

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	images, err := models.CreateImagesbyId(ctxTimeout, db, body, productID)
	if err != nil {
		log.Println("ERROR : ", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "An error occurred while saving data",
		})
		return
	}

	//  ---- ASIGN LIST IMAGE ---
	image := make(map[string]string)
	if images.ImagesOneStr != "" {
		image["image_one"] = images.ImagesOneStr
	}
	if images.ImagesTwoStr != "" {
		image["image_two"] = images.ImagesTwoStr
	}
	if images.ImagesThreeStr != "" {
		image["image_three"] = images.ImagesThreeStr
	}
	if images.ImagesFourStr != "" {
		image["image_four"] = images.ImagesFourStr
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Created Images Succesfully",
		Result:  images,
	})
}
