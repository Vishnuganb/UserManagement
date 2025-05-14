package router

import (
	chi "github.com/go-chi/chi/v5"
	"net/http"
)

type UserHandler interface {
	GetUsers(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	CreateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
}

func NewRouter(uh UserHandler) *chi.Mux {
	r := chi.NewRouter()

	// User management routes
	r.Get("/users", uh.GetUsers)
	r.Get("/users/{id}", uh.GetUserById)
	r.Post("/users", uh.CreateUser)
	r.Delete("/users/{id}", uh.DeleteUser)
	r.Patch("/users/{id}", uh.UpdateUser)

	return r
}
