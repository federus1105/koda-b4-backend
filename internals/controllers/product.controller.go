package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/federus1105/koda-b4-backend/internals/models"
	"github.com/federus1105/koda-b4-backend/internals/pkg/libs"
	"github.com/federus1105/koda-b4-backend/internals/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// GetListProduct godoc
// @Summary Get list products
// @Description Get paginated list of products with optional name filter
// @Tags Products
// @Param page query int false "Page number" default(1)
// @Param name query string false "Filter by product name"
// @Success 200 {object} models.ResponseSucces
// @Router /admin/product [get]
// @Security BearerAuth
func GetListProduct(ctx *gin.Context, db *pgxpool.Pool, rd *redis.Client) {
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

	// --- GET TOTAL COUNT ---
	total, err := models.GetCountProduct(ctxTimeout, db, name)
	if err != nil {
		ctx.JSON(500, gin.H{
			"success": false,
			"message": "Failed to get total product count",
		})
		return
	}

	products, err := models.GetListProduct(ctxTimeout, db, rd, name, limit, offset)
	if err != nil {
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed Get list data products",
		})
		fmt.Println("Error : ", err.Error())
		return
	}

	var prevURL *string
	var nextURL *string

	// --- TOTAL PAGES ---
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// --- QUERY PARAMS ---
	queryPrefix := "?"
	if name != "" {
		queryPrefix = "?name=" + url.QueryEscape(name)
	}

	baseURL := "/admin/product"
	// --- PREV ---
	if page > 1 {
		p := page - 1
		sep := "&"
		if queryPrefix == "" {
			sep = "?"
		}
		url := fmt.Sprintf("%s%s%spage=%d", baseURL, queryPrefix, sep, p)
		prevURL = &url
	}

	// --- NEXT ---
	if page < totalPages {
		n := page + 1
		sep := "&"
		if queryPrefix == "" {
			sep = "?"
		}
		url := fmt.Sprintf("%s%s%spage=%d", baseURL, queryPrefix, sep, n)
		nextURL = &url
	}

	// --- VALIDATION FOR LIST PRODUCT ---
	if len(products) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []string{},
			"message": "Not found list product",
		})
		return
	}
	ctx.JSON(200, models.PaginatedResponse[models.Product]{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		PrevURL:    prevURL,
		NextURL:    nextURL,
		Result:     products,
	})

}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with multiple images
// @Tags Products
// @Accept multipart/form-data
// @Param name formData string true "Product name"
// @Param description formData string true "Product description"
// @Param rating formData number true "Product rating"
// @Param price formData number true "Product price"
// @Param stock formData int true "Product stock"
// @Param image_one formData file true "Primary product image"
// @Param image_two formData file false "Secondary image"
// @Param image_three formData file false "Third image"
// @Param image_four formData file false "Fourth image"
// @Success 200 {object} models.ResponseSucces
// @Router /admin/product [post]
// @Security BearerAuth
func CreateProduct(ctx *gin.Context, db *pgxpool.Pool, rd *redis.Client, cld *cloudinary.Cloudinary) {
	var body models.CreateProducts
	godotenv.Load()

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
			Message: "invalid Form format",
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
		"photos_one":   body.Image_one,
		"photos_two":   body.Image_two,
		"photos_three": body.Image_three,
		"photos_four":  body.Image_four,
	}

	imageStrs := map[string]string{}
	useCloudinary := os.Getenv("CLOUDINARY_URL") != ""

	// --- MULTIPLE UPLOAD IMAGES ---
	for key, file := range imageFiles {
		if file == nil {
			continue
		}
		if !useCloudinary {

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
		} else {
			// --- UPLOAD CLOUDINARY ---
			overwrite := true
			uploadResp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
				Folder:    "assets/product",
				PublicID:  fmt.Sprintf("%s_%d", key, user.ID),
				Overwrite: &overwrite,
			})

			if err != nil {
				ctx.JSON(500, gin.H{
					"success": false,
					"message": fmt.Sprintf("Failed to upload %s to cloudinary", key),
				})
				return
			}
			imageStrs[key] = uploadResp.SecureURL
		}
	}

	// --- ASSIGN TO BODY ---
	body.Image_oneStr = imageStrs["photos_one"]
	body.Image_twoStr = imageStrs["photos_two"]
	body.Image_threeStr = imageStrs["photos_three"]
	body.Image_fourStr = imageStrs["photos_four"]

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	product, err := models.CreateProduct(ctxTimeout, db, rd, body)
	if err != nil {
		log.Println("ERROR : ", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "An error occurred while saving data",
		})
		return
	}

	//  ---- ASIGN LIST IMAGE ---
	images := make(map[string]string)
	if product.Image_oneStr != "" {
		images["image_one"] = product.Image_oneStr
	}
	if product.Image_twoStr != "" {
		images["image_two"] = product.Image_twoStr
	}
	if product.Image_threeStr != "" {
		images["image_three"] = product.Image_threeStr
	}
	if product.Image_fourStr != "" {
		images["image_four"] = product.Image_fourStr
	}

	// --- ASSIGN TO STRUCT RESPONSE ---
	response := models.ProductResponse{
		ID:          product.Id,
		Name:        product.Name,
		ImageID:     product.ImageId,
		Images:      images,
		Price:       product.Price,
		Rating:      product.Rating,
		Description: product.Description,
		Stock:       product.Stock,
		Size:        product.Size,
		Variant:     product.Variant,
		Category:    product.Category,
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Created Product Succesfully",
		Result:  response,
	})

}

// EditProduct godoc
// @Summary Edit an existing product
// @Description Update product details and optionally update images
// @Tags Products
// @Accept multipart/form-data
// @Param id path int true "Product ID"
// @Param name formData string false "Product name"
// @Param description formData string false "Product description"
// @Param rating formData number false "Product rating"
// @Param price formData number false "Product price"
// @Param stock formData int false "Product stock"
// @Param image_one formData file false "Primary product image"
// @Param image_two formData file false "Secondary image"
// @Param image_three formData file false "Third image"
// @Param image_four formData file false "Fourth image"
// @Success 200 {object} models.ResponseSucces
// @Router /admin/product/{id} [patch]
// @Security BearerAuth
func EditProduct(ctx *gin.Context, db *pgxpool.Pool, rd *redis.Client, cld *cloudinary.Cloudinary) {
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
	var body models.UpdateProducts
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
			Message: "invalid Form format",
		})
		return
	}

	body.Id = productID

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

	// --- HANDLE IMAGE UPLOADS ---
	imageFiles := map[string]*multipart.FileHeader{
		"photos_one":   body.Image_one,
		"photos_two":   body.Image_two,
		"photos_three": body.Image_three,
		"photos_four":  body.Image_four,
	}

	imageStrs := map[string]*string{}
	useCloudinary := os.Getenv("CLOUDINARY_URL") != ""

	// --- MULTIPLE UPLOAD IMAGES ---
	for key, file := range imageFiles {
		if file == nil {
			continue
		}

		if !useCloudinary {
			// --- LOCAL UPLOAD ---
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

			imageStrs[key] = &generatedFilename

		} else {
			// --- CLOUDINARY UPLOAD ---
			overwrite := true
			uploadResp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
				Folder:    "assets/product",
				PublicID:  fmt.Sprintf("%s_%d", key, user.ID),
				Overwrite: &overwrite,
			})

			if err != nil {
				ctx.JSON(500, gin.H{
					"success": false,
					"message": fmt.Sprintf("Failed to upload %s to Cloudinary", key),
				})
				return
			}

			url := uploadResp.SecureURL
			imageStrs[key] = &url
		}
	}

	// --- CHECKING ROWS UPDATE ---
	if libs.IsStructEmptyExcept(body, "Id") && len(imageStrs) == 0 {
		ctx.JSON(400, models.Response{
			Success: false,
			Message: "No data to update",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	product, err := models.EditProduct(ctxTimeout, db, rd, body, imageStrs)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "Product not found",
			})
			return
		}
		fmt.Println("error :", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to update product",
		})
		return
	}

	// --- BUILD RESPONSE OBJECT ---
	response := map[string]any{
		"id":          product.Id,
		"name":        product.Name,
		"price":       product.Price,
		"rating":      product.Rating,
		"description": product.Description,
		"stock":       product.Stock,
		"images":      map[string]string{},
		"size":        product.Size,
		"variant":     product.Variant,
		"category":    product.Category,
	}

	images := map[string]string{}
	if product.Image_oneStr != "" {
		images["image_one"] = product.Image_oneStr
	}
	if product.Image_twoStr != "" {
		images["image_two"] = product.Image_twoStr
	}
	if product.Image_threeStr != "" {
		images["image_three"] = product.Image_threeStr
	}
	if product.Image_fourStr != "" {
		images["image_four"] = product.Image_fourStr
	}
	if len(images) > 0 {
		response["images"] = images
	}

	ctx.JSON(http.StatusOK, models.ResponseSucces{
		Success: true,
		Message: "Product updated successfully",
		Result:  response,
	})
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by its ID
// @Tags Products
// @Param id path int true "Product ID"
// @Success 200 {object} models.ResponseSucces
// @Router /admin/product/delete/{id} [post]
// @Security BearerAuth
func DeleteProduct(ctx *gin.Context, db *pgxpool.Pool) {
	productIDstr := ctx.Param("id")
	productId, err := strconv.Atoi(productIDstr)
	if err != nil {
		ctx.JSON(404, models.Response{
			Success: false,
			Message: "product id not found",
		})
		return
	}

	// ---- LIMITS QUERY EXECUTION TIME ---
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = models.DeleteProduct(ctxTimeout, db, productId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(404, models.Response{
				Success: false,
				Message: "Product not found",
			})
			return
		}
		fmt.Println("error :", err)
		ctx.JSON(500, models.Response{
			Success: false,
			Message: "Failed to delete product",
		})
		return
	}

	ctx.JSON(200, models.ResponseSucces{
		Success: true,
		Message: "Delete product successfully",
		Result:  fmt.Sprintf("product id %d", productId),
	})
}
