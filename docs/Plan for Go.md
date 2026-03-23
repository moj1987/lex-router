Create a new db 

```bash
createdb lex_router_db
```
*(Alternatively, you can open DataGrip, connect to your local `localhost:5432` server, right-click, and create a new database named `lex_router_db`).*


# Lex Router - Phase 1 Setup & Development Guide

**Goal:** Build a "Service of Process" API using Go, `chi`, local PostgreSQL, and `sqlx`.

## Prerequisites
Ensure Go is installed (`brew install go`) and your local PostgreSQL server is running.

---

## Step 1: Initialize the Project

Create your workspace and initialize the Go module:
```bash
mkdir lex-router
cd lex-router
go mod init lex-router

# Install dependencies
go get -u github.com/go-chi/chi/v5
go get -u github.com/jmoiron/sqlx
go get -u github.com/lib/pq
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

## Step 3: Domain Models & Database Connection

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

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect() *sqlx.DB {
    // Connects to the local native Postgres server using the default Mac setup
	dsn := "postgres://localhost:5432/proof_db?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}
	
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

## Step 4: The Router & Handlers

**1. Create `internal/handlers/serve_requests.go`:**
```go
package handlers

import (
	"encoding/json"
	"net/http"

	"[github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)"
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
```

**2. Create `cmd/api/main.go`:**
```go
package main

import (
	"log"
	"net/http"

	"[github.com/go-chi/chi/v5](https://github.com/go-chi/chi/v5)"
	"[github.com/go-chi/chi/v5/middleware](https://github.com/go-chi/chi/v5/middleware)"
	"lex-router/internal/database"
	"lex-router/internal/handlers"
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
