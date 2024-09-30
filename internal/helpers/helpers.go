package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var (
	Domain      = os.Getenv("DOMAIN")
	ImageFormat = fmt.Sprintf("CASE WHEN NULLIF(img,'') IS NOT NULL THEN FORMAT ('%s/%%s',img) ELSE NULL END AS img ", Domain)
	secretKey   = []byte(os.Getenv("JWT_SECRET"))
)

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	jsonResp, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error handling JSON response. Err: %v", err)
		return
	}

	_, _ = w.Write(jsonResp)
}

func HandleFileUpload(r *http.Request, table string) (*string, error) {
	file, fileHeader, err := r.FormFile("img")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return nil, fmt.Errorf("error retrieving file: %w \n %w", err, http.ErrMissingFile)
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

	uploadDir := fmt.Sprintf("./uploads/%s", table)
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
	//saveFileMetadata(fileUUID, handler.Filename, filePath)

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
	name := uuid.New()
	fmt.Println(fmt.Sprintf("%s%s", name, ext))
	return fmt.Sprintf("%s%s", name, ext)
}

func GenerateHashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hashedPassword), nil
}

func DeleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error failed to delete file : %w", err)
	}
	return nil
}

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uuid.UUID) (TokenResponse, error) {
	claims := CustomClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	signedString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secretKey)
	if err != nil {
		return TokenResponse{}, err
	}

	tokenResponse := TokenResponse{
		Token:     signedString,
		ExpiresAt: fmt.Sprintf("%dh", 24),
	}

	return tokenResponse, nil
}

func ParseJWT(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}
