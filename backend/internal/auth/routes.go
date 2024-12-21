package auth

import (
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, h *Handler) *mux.Router {
	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", h.Register).Methods("POST")
	auth.HandleFunc("/login", h.Login).Methods("POST")
	auth.HandleFunc("/refresh", h.RefreshToken).Methods("POST")

	// Protected routes example
	protected := r.PathPrefix("").Subrouter()
	protected.Use(h.AuthMiddleware)

	return protected
}
