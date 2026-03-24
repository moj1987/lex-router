package middleware


import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"lex-router/internal/handlers"

	"github.com/golang-jwt/jwt/v5"
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