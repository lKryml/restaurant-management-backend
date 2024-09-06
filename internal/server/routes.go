package server

import (
	"net/http"
	"restaurant-management-backend/internal/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", s.healthHandler)

	r.Get("/user/{id}", s.IndexUserHandler)
	r.Post("/user", s.StoreUserHandler)
	r.Put("/user/{id}", s.UpdateUserHandler)
	r.Delete("/user/{id}", s.DeleteUserHandler)

	return r
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	//Email      string `json:"email"`
	//Phone      string `json:"phone"`
	//Password   string `json:"password"`
	//Created_at string `json:"created_at"`
	//Updated_at string `json:"updated_at"`
}

func (s *Server) IndexUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	resp := make(map[string]string)
	resp["message"] = "Yo yo user here wass good"
	resp["userID"] = id
	utils.WriteJSONResponse(w, 200, resp)
	//writeJSONResponse(w, 200, map[string]string{"message": "Yo yo user here wass good"})

}

func (s *Server) StoreUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	var err error
	id := r.FormValue("id")
	name := r.FormValue("name")
	user.ID, err = strconv.Atoi(id)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, err.Error())
		return
	}
	user.Name = name
	resp := make(map[string]string)
	resp["message"] = "User stored successfully!"

	utils.WriteJSONResponse(w, http.StatusCreated, user)

}

func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp := make(map[string]string)
	resp["message"] = "Updated user successfully!"
	resp["userID"] = id
	utils.WriteJSONResponse(w, http.StatusCreated, resp)

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
	utils.WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSONResponse(w, 200, s.db.Health())

}
