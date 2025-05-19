package database

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Add this function near the top of your photo.go file

// Configuration for image serving
const (
    // Use relative path that works both locally and on VPS
    ImageStorageDir = ".database/images"
    // Allowed image extensions to prevent serving arbitrary files
    AllowedExtensions = ".jpg,.jpeg,.png,.gif,.webp,.bmp"
)
func isValidTokens(token string) bool {
    // For image endpoints, we'll allow access without a token
    return true
}

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
func ServeProductImage(w http.ResponseWriter, r *http.Request) {
    fmt.Println("============= IMAGE REQUEST =============")
    fmt.Println("Path:", r.URL.Path)
    fmt.Println("Working Directory:", getWorkingDir())
    fmt.Println("=========================================")
    
    log.Printf("Image request received for: %s", r.URL.Path)
    enableCors(&w)

    // Extract filename from path
    // Expected path format: /images/filename.ext
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        log.Printf("Invalid image path: %s", r.URL.Path)
        http.Error(w, "Invalid image path", http.StatusBadRequest)
        return
    }
    
    filename := parts[len(parts)-1]
    log.Printf("Requested filename: %s", filename)
    
    // Basic security check: validate filename
    if filename == "" || strings.Contains(filename, "..") {
        http.Error(w, "Invalid filename", http.StatusBadRequest)
        return
    }
    
    // Check file extension
    ext := filepath.Ext(filename)
    if ext == "" || !strings.Contains(AllowedExtensions, strings.ToLower(ext)) {
        http.Error(w, "Unsupported file type", http.StatusBadRequest)
        return
    }
    
    // Construct the full path to the image file
    imagePath := filepath.Join(ImageStorageDir, filename)
    log.Printf("Looking for image at: %s", imagePath)
      // Attempt to resolve absolute path for logging
    absPath, _ := filepath.Abs(imagePath)
    
    // Check if the file exists
    if _, err := os.Stat(imagePath); os.IsNotExist(err) {
        log.Printf("Image not found at path: %s (abs: %s)", imagePath, absPath)
        
        // Ensure the image directory exists
        if _, err := os.Stat(ImageStorageDir); os.IsNotExist(err) {
            log.Printf("Warning: Image directory does not exist: %s", ImageStorageDir)
            if err := os.MkdirAll(ImageStorageDir, 0755); err != nil {
                log.Printf("Failed to create image directory: %v", err)
            } else {
                log.Printf("Created missing image directory: %s", ImageStorageDir)
            }
        }
        
        // Serve placeholder instead
        placeholderPath := filepath.Join(ImageStorageDir, "placeholder.png")
        if _, err := os.Stat(placeholderPath); os.IsNotExist(err) {
            log.Printf("Placeholder image not found at: %s", placeholderPath)
            // If placeholder doesn't exist, serve a 404 error
            http.Error(w, "Image not found", http.StatusNotFound)
            return
        }
        log.Printf("Serving placeholder image from: %s", placeholderPath)
        http.ServeFile(w, r, placeholderPath)
        return
    } else {
        log.Printf("Image found at path: %s (abs: %s)", imagePath, absPath)
    }
    
    // Determine content type based on file extension
    var contentType string
    switch strings.ToLower(ext) {
    case ".jpg", ".jpeg":
        contentType = "image/jpeg"
    case ".png":
        contentType = "image/png"
    case ".gif":
        contentType = "image/gif"
    case ".webp":
        contentType = "image/webp"
    case ".bmp":
        contentType = "image/bmp"
    default:
        contentType = "application/octet-stream"
    }
    
    // Set appropriate headers
    w.Header().Set("Content-Type", contentType)
    w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
    
    // Serve the file
    http.ServeFile(w, r, imagePath)
}

// Add this helper function
func getWorkingDir() string {
    dir, err := os.Getwd()
    if err != nil {
        return "Error getting working directory"
    }
    return dir
}

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
func UploadProductImage(w http.ResponseWriter, r *http.Request) {
    enableCors(&w)
    
    // Check token for authentication
    token := r.Header.Get("token")
    if !isValidToken(token) {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    
    // Check if request method is POST
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Parse the multipart form
    err := r.ParseMultipartForm(10 << 20) // Max 10 MB
    if err != nil {
        http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
        return
    }
    
    // Get the file from form data
    file, handler, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // Check file extension
    ext := filepath.Ext(handler.Filename)
    if ext == "" || !strings.Contains(AllowedExtensions, strings.ToLower(ext)) {
        http.Error(w, "Unsupported file type. Allowed: "+AllowedExtensions, http.StatusBadRequest)
        return
    }
      // Create uploads directory if it doesn't exist
    if _, err := os.Stat(ImageStorageDir); os.IsNotExist(err) {
        err = os.MkdirAll(ImageStorageDir, 0755)
        if err != nil {
            log.Printf("Error creating directory: %v", err)
            http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
            return
        }
        log.Printf("Created image directory at: %s", ImageStorageDir)
    }
    
    // Log the absolute path for debugging
    absPath, err := filepath.Abs(ImageStorageDir)
    if err == nil {
        log.Printf("Using image directory (absolute path): %s", absPath)
    }
    
    // Generate unique filename to prevent overwriting existing files
    // You could use UUID or timestamp-based filenames for production
    filename := fmt.Sprintf("%d_%s", timeNow().Unix(), handler.Filename)
    filepath := filepath.Join(ImageStorageDir, filename)
    
    // Create the file
    dst, err := os.Create(filepath)
    if err != nil {
        log.Printf("Error creating file: %v", err)
        http.Error(w, "Failed to save image: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    
    // Copy the uploaded file data to the newly created file
    _, err = io.Copy(dst, file)
    if err != nil {
        log.Printf("Error copying file data: %v", err)
        http.Error(w, "Failed to save image data: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return success response with the filename
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "filename": filename,
        "message": "Image uploaded successfully",
    })
}

// timeNow is a function that returns the current time
// It's defined as a variable so it can be mocked in tests
var timeNow = func() time.Time {
    return time.Now()
}

