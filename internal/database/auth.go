package database

import (
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"restaurant-management-backend/internal/types"
)

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract user details from form
	user := types.User{
		Name:     r.FormValue("name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if user.Name == "" || user.Email == "" || user.Password == "" {
		http.Error(w, "Name, email, and password are required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(user.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if user.Phone != "" && !isValidPhone(user.Phone) {
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password: "+err.Error(), http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	file, fileHeader, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if file != nil {
		if !isValidImageType(fileHeader.Filename) {
			http.Error(w, "Invalid image type. Only PNG, JPG, or GIF allowed", http.StatusBadRequest)
			return
		}

		uploadDir := "./uploads/"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.Mkdir(uploadDir, os.ModePerm)
		}
		filePath := filepath.Join(uploadDir, sanitizeFilename(fileHeader.Filename))

		out, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Unable to save the file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		user.Img = &filePath
	}

	if err := New().CreateUser(user); err != nil {
		http.Error(w, "Error saving user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User successfully signed up"))
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func isValidPhone(phone string) bool {
	// Simple regex for phone number validation; adjust as needed
	re := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return re.MatchString(phone)
}

func isValidImageType(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return false
}

func sanitizeFilename(filename string) string {
	return filepath.Base(filename)
}
