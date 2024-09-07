package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"restaurant-management-backend/internal/types"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", s.healthHandler)
	r.Get("/users", s.GetUsersHandler)
	//r.Get("/user/{id}", s.IndexUserHandler)
	r.Post("/user", s.StoreUserHandler)
	r.Put("/user/{id}", s.UpdateUserHandler)
	r.Delete("/user/{id}", s.DeleteUserHandler)

	return r
}

func (s *Server) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.ListUsers()
	if err != nil {
		fmt.Errorf("InternalServerError %w", err)
	}
	WriteJSONResponse(w, 200, users)
}

//
//func (s *Server) IndexUserHandler(w http.ResponseWriter, r *http.Request) {
//	id := r.PathValue("id")
//	resp := make(map[string]string)
//	resp["message"] = "Yo yo user here wass good"
//	resp["userID"] = id
//
//	WriteJSONResponse(w, 200, resp)
//	//writeJSONResponse(w, 200, map[string]string{"message": "Yo yo user here wass good"})
//
//}

func (s *Server) StoreUserHandler(w http.ResponseWriter, r *http.Request) {

	user := types.User{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Phone:    r.FormValue("phone"),
		Password: r.FormValue("password"),
	}

	if err := s.db.CreateUser(user); err != nil {
		fmt.Errorf("InternalServerError %w", err)
	}
	resp := make(map[string]string)
	resp["message"] = "User stored successfully!"
	WriteJSONResponse(w, http.StatusCreated, user)

}

func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp := make(map[string]string)
	resp["message"] = "Updated user successfully!"
	resp["userID"] = id
	WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	//user := User{
	//	ID:   id,
	//	Name: "Jiji",
	//}
	resp := make(map[string]string)
	resp["message"] = "Deleted user successfully!"
	resp["userID"] = id
	WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSONResponse(w, 200, s.db.Health())

}
