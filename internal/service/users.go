package service

import (
	"fmt"
	"net/http"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/types"
)

func UserValidator(r *http.Request) (*types.User, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, fmt.Errorf("unable to parse form file size too large: %w", err)
	}

	user := &types.User{
		Name:     r.FormValue("name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := UserVerifyRequired(*user); err != nil {
		return nil, err
	}

	filePath, err := helpers.HandleFileUpload(r, "users")
	if err != nil {
		return nil, fmt.Errorf("failed to handle file upload: %w", err)
	}
	if filePath != nil {
		user.Img = filePath
	}
	return user, nil

}
func SignUpHandler(r *http.Request) (*types.User, error) {
	user, err := UserValidator(r)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := helpers.GenerateHashedPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hashed password: %w", err)
	}
	user.Password = hashedPassword

	return user, nil
}

func UserVerifyRequired(user types.User) error {
	if user.Name == "" || user.Email == "" || user.Password == "" {
		return fmt.Errorf("name, email, and password are required")
	}

	if !helpers.CheckValidEmail(user.Email) {
		return fmt.Errorf("invalid email format: %s", user.Email)
	}

	if user.Phone != "" && !helpers.CheckValidPhone(user.Phone) {
		return fmt.Errorf("invalid phone number format: %s", user.Phone)
	}

	return nil
}
