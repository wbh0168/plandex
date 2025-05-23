package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"plandex/shared" // For shared types
	"strings"
	"testing"
	// "github.com/gorilla/mux" // Only if routes are tested directly with mux dispatcher
)

// TestLoginHandler tests the /api/ui/auth/login endpoint proxying
func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name                 string
		mainServerResponse   interface{}
		mainServerStatus     int
		requestBody          interface{}
		expectedStatusCode   int
		expectInBody         string
		expectSessionResponse *shared.SessionResponse // For successful login check
	}{
		{
			name: "successful login",
			mainServerResponse: &shared.SessionResponse{
				UserId: "user-123", Token: "test-token", Email: "test@example.com",
				Orgs: []*shared.Org{{Id: "org-1", Name: "Test Org"}},
			},
			mainServerStatus:   http.StatusOK,
			requestBody:        shared.SignInRequest{Email: "test@example.com", Pin: "1234"},
			expectedStatusCode: http.StatusOK,
			expectSessionResponse: &shared.SessionResponse{UserId: "user-123", Token: "test-token"},
		},
		{
			name:               "failed login - main server returns 401",
			mainServerResponse: map[string]string{"error": "Invalid credentials"},
			mainServerStatus:   http.StatusUnauthorized,
			requestBody:        shared.SignInRequest{Email: "test@example.com", Pin: "wrongpin"},
			expectedStatusCode: http.StatusUnauthorized,
			expectInBody:       "Invalid credentials",
		},
		{
			name:               "invalid request body to UI backend",
			mainServerResponse: nil, // Main server won't be called
			mainServerStatus:   0,
			requestBody:        "not-json-string",
			expectedStatusCode: http.StatusBadRequest,
			expectInBody:       "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mainServer *httptest.Server
			if tt.mainServerResponse != nil || tt.mainServerStatus != 0 {
				mainServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/accounts/sign_in" {
						t.Fatalf("Main server mock expected path /accounts/sign_in, got %s", r.URL.Path)
					}
					w.WriteHeader(tt.mainServerStatus)
					if tt.mainServerResponse != nil {
						json.NewEncoder(w).Encode(tt.mainServerResponse)
					}
				}))
				defer mainServer.Close()
				originalMainServerURL := mainServerURL // backup
				mainServerURL = mainServer.URL          // point to mock server
				defer func() { mainServerURL = originalMainServerURL }()
			}

			var reqBodyBytes []byte
			var err error
			if strBody, ok := tt.requestBody.(string); ok {
				reqBodyBytes = []byte(strBody)
			} else {
				reqBodyBytes, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/ui/auth/login", bytes.NewBuffer(reqBodyBytes))
			if _, ok := tt.requestBody.(string); !ok { // If not plain string, set content type
				req.Header.Set("Content-Type", "application/json")
			}
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(loginHandler) // Assuming loginHandler is accessible
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tt.expectedStatusCode, rr.Code, rr.Body.String())
			}

			if tt.expectInBody != "" && !strings.Contains(rr.Body.String(), tt.expectInBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectInBody, rr.Body.String())
			}

			if tt.expectSessionResponse != nil {
				var sessionResp shared.SessionResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &sessionResp); err != nil {
					t.Fatalf("Failed to unmarshal successful login response: %v. Body: %s", err, rr.Body.String())
				}
				if sessionResp.UserId != tt.expectSessionResponse.UserId || sessionResp.Token != tt.expectSessionResponse.Token {
					t.Errorf("Expected session response %+v, got %+v", *tt.expectSessionResponse, sessionResp)
				}
			}
		})
	}
}


func TestLogoutHandler(t *testing.T) {
	tests := []struct {
		name                string
		sendAuthHeader      bool   // Whether frontend sends Authorization header
		mainServerStatus    int
		mainServerResponse  map[string]string
		expectedStatusCode  int
		expectInBody        string
	}{
		{
			name: "successful logout",
			sendAuthHeader: true,
			mainServerStatus: http.StatusOK,
			mainServerResponse: map[string]string{"message": "Logged out"},
			expectedStatusCode: http.StatusOK,
			expectInBody: "Logged out",
		},
		{
			name: "logout without frontend Auth header",
			sendAuthHeader: false,
			mainServerStatus: 0, // Main server won't be called by UI backend if it requires header
			expectedStatusCode: http.StatusUnauthorized, // UI backend should reject
			expectInBody: "Missing Authorization header",
		},
		{
			name: "main server returns error on logout",
			sendAuthHeader: true,
			mainServerStatus: http.StatusInternalServerError,
			mainServerResponse: map[string]string{"error": "Server error during logout"},
			expectedStatusCode: http.StatusInternalServerError,
			expectInBody: "Server error during logout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/accounts/sign_out" {
					t.Fatalf("Main server mock expected path /accounts/sign_out, got %s", r.URL.Path)
				}
				// Check if main server received the Authorization header (only if frontend sent it)
				if tt.sendAuthHeader && r.Header.Get("Authorization") == "" {
					t.Error("Main server expected Authorization header, but got none")
				}
				w.WriteHeader(tt.mainServerStatus)
				if tt.mainServerResponse != nil {
					json.NewEncoder(w).Encode(tt.mainServerResponse)
				}
			}))
			defer mainServer.Close()
			originalMainServerURL := mainServerURL
			mainServerURL = mainServer.URL
			defer func() { mainServerURL = originalMainServerURL }()

			req := httptest.NewRequest(http.MethodPost, "/api/ui/auth/logout", nil)
			if tt.sendAuthHeader {
				req.Header.Set("Authorization", "Bearer test-token-from-frontend")
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(logoutHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tt.expectedStatusCode, rr.Code, rr.Body.String())
			}
			if tt.expectInBody != "" && !strings.Contains(rr.Body.String(), tt.expectInBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectInBody, rr.Body.String())
			}
		})
	}
}

func TestSessionHandler(t *testing.T) {
	// Similar structure to TestLogoutHandler
	// Mock main server's /orgs/session
	// Test cases:
	// 1. Successful session retrieval (main server returns 200 + session data)
	// 2. Session invalid/expired (main server returns 401)
	// 3. Frontend doesn't send Authorization header (UI backend returns 401)

	mockSessionData := shared.SessionResponse{ // Using SessionResponse as it contains user and orgs
		UserId: "user-test-id", Email:"test@example.com", UserName: "Test User",
		Orgs: []*shared.Org{{Id: "org-active", Name: "Active Org"}},
		OrgId: "org-active", // Assuming current org is part of session data
	}

	tests := []struct {
		name string
		sendAuthHeader bool
		mainServerStatus int
		mainServerResponse interface{}
		expectedStatusCode int
		expectInBodyPart string // Check for a part of the user's email or ID
		expectErrorContains string
	}{
		{
			name: "successful session check",
			sendAuthHeader: true,
			mainServerStatus: http.StatusOK,
			mainServerResponse: mockSessionData,
			expectedStatusCode: http.StatusOK,
			expectInBodyPart: "user-test-id",
		},
		{
			name: "session check - main server returns 401 (invalid token)",
			sendAuthHeader: true,
			mainServerStatus: http.StatusUnauthorized,
			mainServerResponse: map[string]string{"error": "Token expired"},
			expectedStatusCode: http.StatusUnauthorized,
			expectErrorContains: "Token expired",
		},
		{
			name: "session check without frontend Auth header",
			sendAuthHeader: false,
			mainServerStatus: 0, // Main server not called by UI backend
			expectedStatusCode: http.StatusUnauthorized,
			expectErrorContains: "Missing Authorization header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/orgs/session" {
					t.Fatalf("Main server mock expected path /orgs/session, got %s", r.URL.Path)
				}
				if tt.sendAuthHeader && r.Header.Get("Authorization") == "" {
					t.Error("Main server expected Authorization header, but got none")
				}
				w.WriteHeader(tt.mainServerStatus)
				if tt.mainServerResponse != nil {
					json.NewEncoder(w).Encode(tt.mainServerResponse)
				}
			}))
			defer mainServer.Close()
			originalMainServerURL := mainServerURL
			mainServerURL = mainServer.URL
			defer func() { mainServerURL = originalMainServerURL }()

			req := httptest.NewRequest(http.MethodGet, "/api/ui/auth/session", nil)
			if tt.sendAuthHeader {
				req.Header.Set("Authorization", "Bearer valid-frontend-token")
			}
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(sessionHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatusCode, rr.Code, rr.Body.String())
			}

			bodyStr := rr.Body.String()
			if tt.expectInBodyPart != "" && !strings.Contains(bodyStr, tt.expectInBodyPart) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectInBodyPart, bodyStr)
			}
			if tt.expectErrorContains != "" && !strings.Contains(bodyStr, tt.expectErrorContains) {
				t.Errorf("Expected body to contain error '%s', got '%s'", tt.expectErrorContains, bodyStr)
			}
		})
	}
}
