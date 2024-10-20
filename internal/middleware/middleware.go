package middleware

import (
	"context"
	"github.com/jmoiron/sqlx"
	"net/http"
	"restaurant-management-backend/internal/database"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/types"
	"strings"
)

var s = database.New()

func JWTMiddleware(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var accessToken string
			cookie, err := r.Cookie("accessToken")
			if err == nil {
				accessToken = strings.Replace(cookie.Value, "accessToken=", "", 1)
			}

			bearer := r.Header.Get("Authorization")
			if bearer != "" {
				splitBearer := strings.Split(bearer, " ")
				if len(splitBearer) == 2 {
					accessToken = splitBearer[1]
				}
			}

			if accessToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := helpers.ParseJWT(accessToken)
			if err != nil {
				helpers.HandleError(w, http.StatusUnauthorized, "Invalid access token")
				return
			}

			var user types.User
			if err = db.Get(&user, "SELECT * FROM users WHERE id = $1", userID); err != nil {
				helpers.HandleError(w, http.StatusUnauthorized, "User not found")
				return
			}

			if err := s.GetRoles(&user); err != nil {
				helpers.HandleError(w, http.StatusInternalServerError, "Unable to fetch user roles")
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleMiddleware(allowedRoles ...int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			user, ok := r.Context().Value("user").(types.User)
			if !ok {
				helpers.HandleError(w, http.StatusUnauthorized, "Unauthorized: User information is missing")
				return
			}

			roleMap := make(map[int]bool)
			for _, role := range user.Roles {
				roleMap[role] = true
			}

			for _, allowedRole := range allowedRoles {
				if roleMap[allowedRole] {
					next.ServeHTTP(w, r)
					return
				}
			}

			helpers.HandleError(w, http.StatusForbidden, "Forbidden: You do not have the required role to access this resource")
		})
	}
}
