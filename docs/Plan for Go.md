# Lex Router API - Phase 1 Setup & Development Guide

**Goal:** Build a "Service of Process" API using Go, `chi`, PostgreSQL (via Docker), and `sqlx`.

## Prerequisites

Open your Zsh terminal and install Go using Homebrew:
```bash
brew install go
```
Open VS Code and install the official **Go** extension (by the Go Team at Google).

---

## Step 1: Initialize the Project

Create your workspace and initialize the Go module:
```bash
mkdir lex-router
cd lex-router
go mod init lex-router

# Install dependencies
go get -u [github.com/go-chi/chi/v5](https://github.com/go-chi/chi/v5)
go get -u [github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)
go get -u [github.com/lib/pq](https://github.com/lib/pq)
```

---

## Step 2: Project Architecture

Create the standard Go directory layout:
```bash
mkdir -p cmd/api
mkdir -p internal/models
mkdir -p internal/handlers
mkdir -p internal/database
```

---

## Step 3: Database Setup (Docker)

Create a `docker-compose.yml` file in your project root to run PostgreSQL:

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: lex_router_db
    ports:
      - "5432:5432"
```

Start the database in the background:
```bash
docker compose up -d
```

---

## Step 4: Domain Models & Database Connection

**1. Create `internal/models/models.go`:**
```go
package models

import "time"

type LawFirm struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ServeRequest struct {
	ID         int       `json:"id" db:"id"`
	LawFirmID  int       `json:"law_firm_id" db:"law_firm_id"`
	Defendant  string    `json:"defendant" db:"defendant"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
```

**2. Create `internal/database/db.go`:**
```go
package database

import (
	"log"

	"[github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)"
	_ "[github.com/lib/pq](https://github.com/lib/pq)"
)

func Connect() *sqlx.DB {
	dsn := "postgres://user:password@localhost:5432/proof_db?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}
	
	// Create tables if they don't exist
	schema := `
	CREATE TABLE IF NOT EXISTS law_firms (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS serve_requests (
		id SERIAL PRIMARY KEY,
		law_firm_id INT REFERENCES law_firms(id),
		defendant TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	db.MustExec(schema)
	return db
}
```

---

## Step 5: The Router & Handlers

**1. Create `internal/handlers/serve_requests.go`:**
```go
package handlers

import (
	"encoding/json"
	"net/http"

	"[github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)"
	"proof-api/internal/models"
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
```

**2. Create `cmd/api/main.go`:**
```go
package main

import (
	"log"
	"net/http"

	"[github.com/go-chi/chi/v5](https://github.com/go-chi/chi/v5)"
	"[github.com/go-chi/chi/v5/middleware](https://github.com/go-chi/chi/v5/middleware)"
	"proof-api/internal/database"
	"proof-api/internal/handlers"
)

func main() {
	db := database.Connect()
	env := &handlers.Env{DB: db}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API is running"))
	})
	
	r.Get("/requests", env.GetServeRequests)

	log.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", r)
}
```

---

## Step 6: Testing

**Create `cmd/api/main_test.go`:**
```go
package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API is running"))
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
	}
}
```

---

## Running the Project

1. Ensure Docker is running: `docker compose up -d`
2. Run the tests: `go test ./cmd/api`
3. Start the server: `go run cmd/api/main.go`
4. Test the API in another terminal: `curl http://localhost:8080/health`
```

Do you want to run through the Go and VS Code installation right now, or will you be starting this at a later time?