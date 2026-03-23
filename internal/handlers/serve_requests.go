package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"lex-router/internal/models"
)

type Env struct {
	DB *sqlx.DB
}

func (env *Env) GetServeRequests(w http.ResponseWriter, r *http.Request) {
	var requests []models.ServeRequest
	
	err := env.DB.Select(&requests, "SELECT * FROM serve_requests")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}