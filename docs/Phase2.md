# Lex Router - Phase 2: JWT Authentication

**Goal:** Implement a dummy login endpoint to generate a JSON Web Token (JWT), and create selective middleware to protect the `/requests` endpoint.

## Step 1: Install the JWT Package

Stop your server (`Ctrl+C`) and run this command in your terminal to install the industry-standard Go JWT library:
`go get -u github.com/golang-jwt/jwt/v5`

---

## Step 2: The Login Handler (Token Generation)

Create a new file: `internal/handlers/auth.go`. This handler checks a hardcoded dummy username/password and mints a token.

```go
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"[github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5)"
)

// In a real app, this would be a secure environment variable
var SecretKey = []byte("my-super-secret-key")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds LoginRequest
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Dummy user check (we will replace this with bcrypt and DB checks later)
	if creds.Username != "admin" || creds.Password != "password123" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create the JWT claims (the payload data)
	claims := jwt.MapClaims{
		"username": creds.Username,
		"role":     "law_firm",
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	}

	// Sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(SecretKey)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
```

---

## Step 3: The Interceptor (Middleware)

Create a new file: `internal/middleware/auth.go`. This is the security desk.

```go
package middleware

import (
	"net/http"
	"strings"

	"[github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5)"
	"lex-router/internal/handlers"
)

func RequireJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Look for the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// 2. Extract the token (Format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// 3. Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return handlers.SecretKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// 4. Token is valid! Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}
```

---

## Step 4: Wire It Up in Main

Update your `cmd/api/main.go` to register the new endpoint and apply the selective middleware using Chi's `.With()` grouping.

```go
package main

import (
	"log"
	"net/http"

	"[github.com/go-chi/chi/v5](https://github.com/go-chi/chi/v5)"
	"[github.com/go-chi/chi/v5/middleware](https://github.com/go-chi/chi/v5/middleware)"
	"lex-router/internal/database"
	"lex-router/internal/handlers"
	authMiddleware "lex-router/internal/middleware" // Renamed on import to avoid package clash
)

func main() {
	db := database.Connect()
	env := &handlers.Env{DB: db}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Public Endpoints (No security desk)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API is running"))
	})
	r.Post("/login", handlers.Login)

	// Protected Endpoints (Requires passing the security desk)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireJWT) // Applies middleware only to routes inside this group
		
		r.Get("/requests", env.GetServeRequests)
	})

	log.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", r)
}
```
