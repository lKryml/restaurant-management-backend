package helpers

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	jsonResp, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
		return
	}

	_, _ = w.Write(jsonResp)
}

func HandleFileUpload(r *http.Request) (*string, error) {
	file, fileHeader, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		return nil, fmt.Errorf("Error retrieving file: %w", err)
	}
	if file == nil {
		return nil, nil
	}
	defer file.Close()

	if !CheckValidImageType(fileHeader.Filename) {
		return nil, fmt.Errorf("Invalid image type. Only PNG, JPG, or GIF allowed")
	}

	safeFilename := SanitizeFilename(fileHeader.Filename)
	if safeFilename == "" {
		return nil, fmt.Errorf("Invalid file name")
	}

	uploadDir := "./uploads/"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("Unable to create upload directory: %w", err)
	}

	filePath := filepath.Join(uploadDir, GenerateUniqueFilename(safeFilename))

	out, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to save the file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		return nil, fmt.Errorf("Error saving file: %w", err)
	}

	return &filePath, nil
}

func HandleError(w http.ResponseWriter, status int, message string) {
	WriteJSONResponse(w, status, map[string]string{"error": message})

}

func CheckValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func CheckValidPhone(phone string) bool {
	// Simple regex for phone number validation; adjust as needed
	re := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return re.MatchString(phone)
}

func CheckValidImageType(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return false
}

func SanitizeFilename(filename string) string {
	return filepath.Base(filename)
}

func GenerateUniqueFilename(filename string) string {
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", name, timestamp, ext)
}

func GenerateHashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Error hashing password: %w", err)
	}
	return string(hashedPassword), nil
}
