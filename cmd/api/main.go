package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"lex-router/internal/database"
	"lex-router/internal/handlers"
	authMiddleware "lex-router/internal/middleware"
)

func main() {
	db := database.Connect()
	env := &handlers.Env{DB: db}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API is running"))
	})
	r.Post("/login", handlers.Login)
	
	r.Get("/requests", env.GetServeRequests)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireJWT)

		r.Get("/requests", env.GetServeRequests)
	})

	log.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", r)
}