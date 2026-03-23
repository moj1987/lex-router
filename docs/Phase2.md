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

# Authentication Tests

**Goal:** Write table-driven tests for the JSON body parsing in the Login handler, and test the HTTP header extraction in our JWT middleware.

---

## Step 1: Testing the Login Handler

Create a new file: `internal/handlers/auth_test.go`. Because this endpoint expects a JSON body, we will use `strings.NewReader` to simulate the incoming request payloads.

```go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		payload        LoginRequest
		expectedStatus int
	}{
		{
			name:           "Valid credentials",
			payload:        LoginRequest{Username: "admin", Password: "password123"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid password",
			payload:        LoginRequest{Username: "admin", Password: "wrong"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert our struct payload into JSON bytes
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(Login)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
```

---

## Step 2: Testing the Middleware

Create a new file: `internal/middleware/auth_test.go`. To test middleware, we create a dummy "next" handler that just returns a 200 OK. If the middleware fails, it will block the request and return a 401. If it succeeds, it passes through to our dummy handler.

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"[github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5)"
	"lex-router/internal/handlers"
)

// Helper function to generate a real token for our success test
func generateValidToken() string {
	claims := jwt.MapClaims{
		"username": "admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(handlers.SecretKey)
	return tokenString
}

func TestRequireJWT(t *testing.T) {
	// A dummy handler to simulate our protected endpoint
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap our dummy handler in the middleware
	handlerToTest := RequireJWT(nextHandler)

	tests := []struct {
		name           string
		setupAuth      func(*http.Request)
		expectedStatus int
	}{
		{
			name:           "Missing Authorization header",
			setupAuth:      func(r *http.Request) {}, // Do nothing
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid token",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer fake-bad-token")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Valid token",
			setupAuth: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+generateValidToken())
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			tt.setupAuth(req) // Apply the specific headers for this test case

			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
```
