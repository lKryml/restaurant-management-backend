package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"restaurant-management-backend/internal/types"
	"strconv"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", s.healthHandler)
	r.Get("/users", s.GetUsers)
	r.Get("/user/{id}", s.IndexUserHandler)
	r.Post("/user", s.StoreUserHandler)
	r.Put("/user/{id}", s.UpdateUserHandler)
	r.Delete("/user/{id}", s.DeleteUserHandler)

	return r
}

type User struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}

func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.ListUsers()
	if err != nil {
		fmt.Println(err.Error())
	}
	types.WriteJSONResponse(w, 200, users)
}

func (s *Server) IndexUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp := make(map[string]string)
	resp["message"] = "Yo yo user here wass good"
	resp["userID"] = id

	types.WriteJSONResponse(w, 200, resp)
	//writeJSONResponse(w, 200, map[string]string{"message": "Yo yo user here wass good"})

}

func (s *Server) StoreUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	var err error
	id := r.FormValue("id")
	name := r.FormValue("name")
	user.ID, err = strconv.Atoi(id)
	if err != nil {
		types.HandleError(w, http.StatusBadRequest, err.Error())
		return
	}
	user.Name = name
	resp := make(map[string]string)
	resp["message"] = "User stored successfully!"

	types.WriteJSONResponse(w, http.StatusCreated, user)

}

func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp := make(map[string]string)
	resp["message"] = "Updated user successfully!"
	resp["userID"] = id
	types.WriteJSONResponse(w, http.StatusCreated, resp)

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
	types.WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	types.WriteJSONResponse(w, 200, s.db.Health())

}
