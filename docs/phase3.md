# Lex Router - Phase 3: Background Jobs & E2E Testing

**Goal:** Build a route that triggers an asynchronous Goroutine using `context`, and write an End-to-End integration test that hits the database and cleans up after itself.

## Step 1: Database Update Method

Open `internal/handlers/serve_requests.go` and add a helper method to our `Env` struct so we can easily update the status of a request.

```go
func (env *Env) UpdateRequestStatus(ctx context.Context, id int, status string) error {
	// ExecContext allows us to pass our timeout context directly to the database driver
	_, err := env.DB.ExecContext(ctx, "UPDATE serve_requests SET status = $1 WHERE id = $2", status, id)
	return err
}
```

---

## Step 2: The Background Processing Handler

Still in `internal/handlers/serve_requests.go`, add the new endpoint logic. Notice how we create a completely fresh background context.

```go
import (
    "context"
    "log"
    "strconv"
    "time"
    "[github.com/go-chi/chi/v5](https://github.com/go-chi/chi/v5)"
)

func (env *Env) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL (e.g., /requests/5/process)
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	// 1. Create a detached context with a 10-second timeout for the background worker
	bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// 2. Spin up the Goroutine
	go func(jobCtx context.Context, reqID int) {
		defer cancel() // Ensures resources are freed when the job finishes

		log.Printf("Starting background job for request %d...", reqID)
		
		// Simulate heavy PDF generation
		time.Sleep(5 * time.Second) 
		
		// Simulate a random failure (fails on even seconds) just to see our error logging
		if time.Now().Unix()%2 == 0 {
			env.UpdateRequestStatus(jobCtx, reqID, "failed")
			log.Printf("ERROR: Background job for request %d failed during generation", reqID)
			return
		}

		err := env.UpdateRequestStatus(jobCtx, reqID, "completed")
		if err != nil {
			log.Printf("ERROR: Could not update DB for request %d: %v", reqID, err)
			env.UpdateRequestStatus(context.Background(), reqID, "failed") // Fallback
			return
		}

		log.Printf("Background job for request %d completed successfully!", reqID)
	}(bgCtx, id) // We pass bgCtx and id IN as arguments to avoid closure bugs

	// 3. Instantly return success to the user so their app doesn't freeze
	w.WriteHeader(http.StatusAccepted) // 202 Accepted
	w.Write([]byte("Processing started in the background"))
}
```

*Don't forget to register this new route inside `cmd/api/main.go` inside your protected group:*
`r.Post("/requests/{id}/process", env.ProcessRequest)`

---

## Step 3: End-to-End Integration Testing

Create a new file: `cmd/api/integration_test.go`. This test will actually connect to your local database, insert a test row, hit the API, wait for the Goroutine to finish, check the result, and finally delete the test row.

```go
package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"[github.com/go-chi/chi/v5](https://github.com/go-chi/chi/v5)"
	"lex-router/internal/database"
	"lex-router/internal/handlers"
)

func TestBackgroundProcessingFlow(t *testing.T) {
	// 1. Setup: Connect to the real DB
	db := database.Connect()
	env := &handlers.Env{DB: db}
	defer db.Close()

	// 2. Setup: Insert a dummy Law Firm and Request
	var firmID, reqID int
	db.QueryRow("INSERT INTO law_firms (name) VALUES ('Test Firm') RETURNING id").Scan(&firmID)
	db.QueryRow("INSERT INTO serve_requests (law_firm_id, defendant, status) VALUES ($1, 'John Doe', 'pending') RETURNING id", firmID).Scan(&reqID)

	// 3. Teardown: Ensure we delete this test data when the test finishes, even if it fails
	defer func() {
		db.Exec("DELETE FROM serve_requests WHERE id = $1", reqID)
		db.Exec("DELETE FROM law_firms WHERE id = $1", firmID)
	}()

	// 4. Execute: Create a test router and hit the endpoint
	r := chi.NewRouter()
	r.Post("/requests/{id}/process", env.ProcessRequest)

	req, _ := http.NewRequest("POST", "/requests/"+strconv.Itoa(reqID)+"/process", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("Expected 202 Accepted, got %v", rr.Code)
	}

	// 5. Assert: Wait 6 seconds for the Goroutine to finish its 5-second sleep
	time.Sleep(6 * time.Second)

	var finalStatus string
	err := db.Get(&finalStatus, "SELECT status FROM serve_requests WHERE id = $1", reqID)
	if err != nil {
		t.Fatalf("Failed to query final status: %v", err)
	}

	// It should be either 'completed' or 'failed' based on our random logic, but never 'pending'
	if finalStatus == "pending" {
		t.Errorf("Expected status to change, but it remained 'pending'")
	}
}
```
