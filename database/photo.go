package database

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Configuration for image serving
const (
    // Change this to match your actual image storage directory
    ImageStorageDir = `D:\Coding File 2\frozen_food_v2\images`
    // Allowed image extensions to prevent serving arbitrary files
    AllowedExtensions = ".jpg,.jpeg,.png,.gif,.webp,.bmp"
)

// @Summary Serve product image
// @Description Serve a product image from the server's storage
// @Tags Images
// @Accept json
// @Produce octet-stream
// @Param filename path string true "Image filename"
// @Success 200 {file} binary "Image file"
// @Failure 404 {string} string "Image not found"
// @Failure 400 {string} string "Invalid filename or unsupported file type"
// @Failure 500 {string} string "Error reading image file"
// @Router /images/{filename} [get]

// @Summary Upload product image
// @Description Upload a new product image to the server
// @Tags Images
// @Accept multipart/form-data
// @Produce json
// @Param token header string true "Authentication token"
// @Param image formData file true "Image file to upload"
// @Success 200 {object} map[string]string "filename: uploaded_filename.ext"
// @Failure 400 {object} map[string]string "Error: Invalid file or missing parameters"
// @Failure 401 {object} map[string]string "Error: Invalid token"
// @Failure 500 {object} map[string]string "Error: Failed to upload image"
// @Router /uploadimage [post]


// CopyImagesToUploadsDir copies images from a source directory to the uploads directory
func CopyImagesToUploadsDir() {
    sourceDir := "/images/"
    // Check if source directory exists
    if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
        log.Printf("Source directory does not exist: %s", sourceDir)
        return
    }
    files, err := os.ReadDir(sourceDir)
    if err != nil {
        log.Printf("Error reading source directory: %v", err)
        return
    }
    
    // Create destination directory if it doesn't exist
    if _, err := os.Stat(ImageStorageDir); os.IsNotExist(err) {
        err = os.MkdirAll(ImageStorageDir, 0755)
        if err != nil {
            log.Printf("Error creating directory: %v", err)
            return
        }
    }
    
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        
        ext := filepath.Ext(file.Name())
        if ext == "" || !strings.Contains(AllowedExtensions, strings.ToLower(ext)) {
            continue // Skip non-image files
        }
        
        sourcePath := filepath.Join(sourceDir, file.Name())
        destPath := filepath.Join(ImageStorageDir, file.Name())
        
        // Skip if file already exists in destination
        if _, err := os.Stat(destPath); err == nil {
            continue
        }
        
        // Copy file
        sourceFile, err := os.Open(sourcePath)
        if err != nil {
            log.Printf("Error opening source file %s: %v", file.Name(), err)
            continue
        }
        defer sourceFile.Close()
        
        destFile, err := os.Create(destPath)
        if err != nil {
            log.Printf("Error creating destination file %s: %v", file.Name(), err)
            continue
        }
        defer destFile.Close()
        
        _, err = io.Copy(destFile, sourceFile)
        if err != nil {
            log.Printf("Error copying file %s: %v", file.Name(), err)
        }
    }
    
    log.Printf("Image migration complete")
}