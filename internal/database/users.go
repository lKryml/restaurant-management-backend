package database

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"restaurant-management-backend/internal/types"
	"time"
)

var QB = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func (s *service) ListUsers() ([]types.User, error) {
	var users []types.User
	query, args, err := QB.Select("*").From("users").ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql query builder failed %w", err)
	}
	if err := s.db.Select(&users, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *service) GetUserByID(id int) (types.User, error) {
	var user types.User

	err := s.db.Get(&user, "SELECT id, name FROM users WHERE id = $1", id)
	return user, err
}

func (s *service) CreateUser(user types.User) error {
	var id string
	_, err := s.db.NamedQuery("INSERT INTO users (name,email,password,phone) VALUES (:name,:email,:password,:phone) RETURNING id", user)
	fmt.Println(id)
	return err
}

func (s *service) UpdateUser(user types.User) error {
	_, err := s.db.Exec("UPDATE users SET name = $1 WHERE id = $2", user.Name, user.ID)
	return err
}

func (s *service) DeleteUser(id int) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

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

	if file != nil {
		defer file.Close()

		if !isValidImageType(fileHeader.Filename) {
			http.Error(w, "Invalid image type. Only PNG, JPG, or GIF allowed", http.StatusBadRequest)
			return
		}

		safeFilename := sanitizeFilename(fileHeader.Filename)
		if safeFilename == "" {
			http.Error(w, "Invalid file name", http.StatusBadRequest)
			return
		}

		uploadDir := "./uploads/"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			http.Error(w, "Unable to create upload directory: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filePath := filepath.Join(uploadDir, uniqueFilename(safeFilename))

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

func uniqueFilename(filename string) string {
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", name, timestamp, ext)
}
