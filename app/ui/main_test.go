package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	// "github.com/gorilla/mux" // Not strictly needed if testing proxyRequest directly or simple handlers
)

// TestProxyRequest_HeaderForwarding tests if headers (Auth, Content-Type) are forwarded.
func TestProxyRequest_HeaderForwarding(t *testing.T) {
	var capturedRequestToMainServer *http.Request

	mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestToMainServer = r // Capture the request
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "main server ok"}`))
	}))
	defer mainServer.Close()

	originalMainServerURL := mainServerURL
	mainServerURL = mainServer.URL
	defer func() { mainServerURL = originalMainServerURL }()

	// Create a dummy handler in the UI backend that uses proxyRequest
	uiBackendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path here is relative to mainServerURL, e.g., "/test/proxy"
		proxyRequest(r.Method, r.URL.Path, r.Body, w, r)
	})

	frontendReqBody := `{"key": "value"}`
	req := httptest.NewRequest(http.MethodPost, "/test/proxy", strings.NewReader(frontendReqBody))
	req.Header.Set("Authorization", "Bearer frontend-auth-token")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom-Header", "custom-value") // proxyRequest doesn't explicitly forward this one

	rr := httptest.NewRecorder()
	uiBackendHandler.ServeHTTP(rr, req)

	if capturedRequestToMainServer == nil {
		t.Fatal("Main server was not called by proxyRequest")
	}

	// Check Authorization header
	if gotAuth := capturedRequestToMainServer.Header.Get("Authorization"); gotAuth != "Bearer frontend-auth-token" {
		t.Errorf("Expected Authorization header 'Bearer frontend-auth-token', got '%s'", gotAuth)
	}

	// Check Content-Type header
	if gotContentType := capturedRequestToMainServer.Header.Get("Content-Type"); gotContentType != "application/json" {
		t.Errorf("Expected Content-Type header 'application/json', got '%s'", gotContentType)
	}
	
	// Check X-Forwarded-For (proxyRequest should add this)
	if capturedRequestToMainServer.Header.Get("X-Forwarded-For") == "" {
		t.Error("Expected X-Forwarded-For header to be set, but it was empty")
	}
	
	// Check that X-Custom-Header was NOT forwarded by the current proxyRequest logic
    // (Current proxyRequest only explicitly forwards Auth and Content-Type, and adds X-Forwarded-For)
    if gotCustomHeader := capturedRequestToMainServer.Header.Get("X-Custom-Header"); gotCustomHeader != "" {
        t.Errorf("X-Custom-Header should not have been forwarded by default, but got '%s'", gotCustomHeader)
    }


	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "main server ok") {
		t.Errorf("Expected body from main server, got %s", rr.Body.String())
	}
}

// TestProxyRequest_BodyAndResponseProxying tests if body and response are proxied.
func TestProxyRequest_BodyAndResponseProxying(t *testing.T) {
	var capturedRequestBody string
	mainServerResponsePayload := map[string]string{"reply": "data from main server"}

	mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		capturedRequestBody = string(bodyBytes)
		r.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-MainServer-Header", "main-server-custom-value")
		w.WriteHeader(http.StatusCreated) // Use a non-200 status to check status proxying
		json.NewEncoder(w).Encode(mainServerResponsePayload)
	}))
	defer mainServer.Close()

	originalMainServerURL := mainServerURL
	mainServerURL = mainServer.URL
	defer func() { mainServerURL = originalMainServerURL }()

	uiBackendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(r.Method, "/test/bodyproxy", r.Body, w, r)
	})

	frontendReqPayload := map[string]string{"action": "create", "data": "test"}
	frontendReqBodyBytes, _ := json.Marshal(frontendReqPayload)

	req := httptest.NewRequest(http.MethodPost, "/test/bodyproxy", bytes.NewBuffer(frontendReqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	uiBackendHandler.ServeHTTP(rr, req)

	// Check request body received by main server
	var expectedReqBodyMap map[string]string
	json.Unmarshal(frontendReqBodyBytes, &expectedReqBodyMap)
	var capturedReqBodyMap map[string]string
	json.Unmarshal([]byte(capturedRequestBody), &capturedReqBodyMap)

	if capturedReqBodyMap["action"] != expectedReqBodyMap["action"] || capturedReqBodyMap["data"] != expectedReqBodyMap["data"] {
		t.Errorf("Expected main server to receive body '%s', got '%s'", string(frontendReqBodyBytes), capturedRequestBody)
	}

	// Check status code relayed from main server
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	// Check response body relayed from main server
	var responsePayload map[string]string
	json.Unmarshal(rr.Body.Bytes(), &responsePayload)
	if responsePayload["reply"] != mainServerResponsePayload["reply"] {
		t.Errorf("Expected response body '%v', got '%v'", mainServerResponsePayload, responsePayload)
	}
	
	// Check headers relayed from main server
	if rr.Header().Get("X-MainServer-Header") != "main-server-custom-value" {
		t.Errorf("Expected main server header 'X-MainServer-Header' to be proxied, got '%s'", rr.Header().Get("X-MainServer-Header"))
	}
	if !strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json") { // Main server sets application/json
		t.Errorf("Expected proxied Content-Type to be 'application/json', got '%s'", rr.Header().Get("Content-Type"))
	}
}

// TestProxyRequest_DifferentMethods tests proxying with GET and DELETE
func TestProxyRequest_DifferentMethods(t *testing.T) {
	var lastMethodUsedOnMainServer string

	mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethodUsedOnMainServer = r.Method
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			w.Write([]byte("GET processed"))
		} else if r.Method == http.MethodDelete {
			w.Write([]byte("DELETE processed")) // Or often 204 No Content
		}
	}))
	defer mainServer.Close()

	originalMainServerURL := mainServerURL
	mainServerURL = mainServer.URL
	defer func() { mainServerURL = originalMainServerURL }()

	uiBackendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(r.Method, "/test/methods", r.Body, w, r)
	})

	// Test GET
	reqGet := httptest.NewRequest(http.MethodGet, "/test/methods", nil)
	rrGet := httptest.NewRecorder()
	uiBackendHandler.ServeHTTP(rrGet, reqGet)

	if lastMethodUsedOnMainServer != http.MethodGet {
		t.Errorf("Expected GET method on main server, got %s", lastMethodUsedOnMainServer)
	}
	if rrGet.Code != http.StatusOK {
		t.Errorf("GET request: Expected status 200, got %d", rrGet.Code)
	}
	if !strings.Contains(rrGet.Body.String(), "GET processed") {
		t.Errorf("GET request: Expected body 'GET processed', got '%s'", rrGet.Body.String())
	}

	// Test DELETE
	reqDelete := httptest.NewRequest(http.MethodDelete, "/test/methods", nil)
	rrDelete := httptest.NewRecorder()
	uiBackendHandler.ServeHTTP(rrDelete, reqDelete)

	if lastMethodUsedOnMainServer != http.MethodDelete {
		t.Errorf("Expected DELETE method on main server, got %s", lastMethodUsedOnMainServer)
	}
	if rrDelete.Code != http.StatusOK { // Or 204 if main server sent that
		t.Errorf("DELETE request: Expected status 200, got %d", rrDelete.Code)
	}
	if !strings.Contains(rrDelete.Body.String(), "DELETE processed") {
		t.Errorf("DELETE request: Expected body 'DELETE processed', got '%s'", rrDelete.Body.String())
	}
}

// TestProxyRequest_MainServerError tests how proxy handles main server errors (e.g., 500)
func TestProxyRequest_MainServerError(t *testing.T) {
	mainServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error from main"})
	}))
	defer mainServer.Close()

	originalMainServerURL := mainServerURL
	mainServerURL = mainServer.URL
	defer func() { mainServerURL = originalMainServerURL }()

	uiBackendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(r.Method, "/test/error", r.Body, w, r)
	})

	req := httptest.NewRequest(http.MethodGet, "/test/error", nil)
	rr := httptest.NewRecorder()
	uiBackendHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	var respBody map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &respBody); err != nil {
		t.Fatalf("Could not unmarshal error response body: %v", err)
	}
	if respBody["error"] != "internal server error from main" {
		t.Errorf("Expected error message 'internal server error from main', got '%s'", respBody["error"])
	}
	// Check that Content-Type from main server's error response is also proxied
    if !strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json") {
        t.Errorf("Expected error response Content-Type 'application/json', got '%s'", rr.Header().Get("Content-Type"))
    }
}
