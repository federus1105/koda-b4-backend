package utils

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

// --- MAXIMAL FILE SIZE ---
const MaxFileSize = 500 * 1024

// ALLOWED FILE TYPES
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func UploadImageFile(ctx context.Context, file *multipart.FileHeader, uploadDir string, prefix string) (string, string, error) {
	if file == nil {
		return "", "", errors.New("file not found")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return "", "", errors.New("unsupported file formats (only jpg, jpeg, png, webp)")
	}

	if file.Size > MaxFileSize {
		return "", "", errors.New("maximum file size 500 kb")
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), prefix, ext)
	savePath := filepath.Join(uploadDir, filename)

	return savePath, filename, nil
}
