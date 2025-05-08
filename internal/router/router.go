package router

import (
	chi "github.com/go-chi/chi/v5"

	"UserManagement/internal/handler"
)

func NewRouter(uh *handler.UserHandler) *chi.Mux {
	r := chi.NewRouter()

	// User management routes
	r.Get("/users", uh.GetUsers)
	r.Get("/users/{id}", uh.GetUserById)
	r.Post("/users", uh.CreateUser)
	r.Delete("/users/{id}", uh.DeleteUser)
	r.Patch("/users/{id}", uh.UpdateUser)

	/*
		🔸 This is not a REST API route
		🔸 This is used by a WebSocket client (like JavaScript in browser)
		🔸 This route upgrades the connection from HTTP → WebSocket (bi-directional connection)
	*/
	// WebSocket route
	//r.Get("/ws", wh.HandleWebSocket)

	return r
}
