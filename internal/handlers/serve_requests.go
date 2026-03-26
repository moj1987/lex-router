package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"lex-router/internal/models"

	"github.com/jmoiron/sqlx"
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

func (env *Env) UpdateRequestStatus (ctx context.Context, id int, status string) error {
	_, err := env.DB.ExecContext (ctx, "UPDATE serve_requests SET status = $1 WHERE id = $2", status, id)
	return err
}
