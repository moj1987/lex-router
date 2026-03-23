package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
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