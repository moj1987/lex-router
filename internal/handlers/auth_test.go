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