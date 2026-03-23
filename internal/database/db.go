package database

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)


func Connect() *sqlx.DB {
    // Connects to the local native Postgres server using the default Mac setup
	dsn := "postgres://localhost:5432/lex_router_db?sslmode=disable"
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