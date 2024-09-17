package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"restaurant-management-backend/internal/helpers"
	"restaurant-management-backend/internal/service"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/health", s.healthHandler)
	r.Get("/users", s.indexUsersHandler)
	r.Get("/user/{id}", s.getUserHandler)
	r.Post("/user", s.createUserHandler)
	r.Put("/user/{id}", s.updateUserHandler)
	r.Delete("/user/{id}", s.deleteUserHandler)

	return r
}

func (s *Server) indexUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.ListUsers()
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	helpers.WriteJSONResponse(w, 200, users)
}

func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		helpers.HandleError(w, http.StatusBadRequest, "id is required")
		return
	}
	user, err := s.db.GetUserByID(id)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteJSONResponse(w, 200, user)

}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {

	user, err := service.SignUpHandler(r)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err = s.db.CreateUser(*user)
	if err != nil {
		helpers.HandleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make(map[string]interface{})
	resp["message"] = "User successfully signed up"
	resp["user"] = user
	helpers.WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp := make(map[string]string)
	resp["message"] = "Updated user successfully!"
	resp["userID"] = id
	helpers.WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	//user := User{
	//	ID:   id,
	//	Name: "Jiji",
	//}
	resp := make(map[string]string)
	resp["message"] = "Deleted user successfully!"
	resp["userID"] = id
	helpers.WriteJSONResponse(w, http.StatusCreated, resp)

}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	helpers.WriteJSONResponse(w, 200, s.db.Health())

}
