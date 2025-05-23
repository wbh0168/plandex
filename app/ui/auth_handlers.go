package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"plandex/shared" // For shared.SessionResponse, shared.SignInRequest
)

// loginHandler handles requests to /api/ui/auth/login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var signInReq shared.SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&signInReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Construct request to the main server's /accounts/sign_in endpoint
	mainServerLoginURL := mainServerURL + "/accounts/sign_in" // mainServerURL is defined in main.go

	reqBodyBytes, err := json.Marshal(signInReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshalling login request: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a new request to the main server
	proxyReq, err := http.NewRequest(http.MethodPost, mainServerLoginURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating proxy request for login: %v", err), http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	// Copy X-Forwarded-For and other relevant headers if needed for main server's IP logging/checks
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)


	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error sending login request to main server: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers from the main server's response (e.g., Set-Cookie for authToken)
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	
	w.WriteHeader(resp.StatusCode) // Send status code from main server

	// Copy the body (shared.SessionResponse or error) from the main server's response
	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("Error copying login response body from main server: %v\n", err)
		// Don't write http.Error here as headers/status might already be sent
	}
}

// logoutHandler handles requests to /api/ui/auth/logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the Authorization header received from the frontend
	frontendAuthHeader := r.Header.Get("Authorization")
	if frontendAuthHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Construct request to the main server's /accounts/sign_out endpoint
	mainServerLogoutURL := mainServerURL + "/accounts/sign_out"

	proxyReq, err := http.NewRequest(http.MethodPost, mainServerLogoutURL, nil) // No body for logout
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating proxy request for logout: %v", err), http.StatusInternalServerError)
		return
	}

	// Forward the exact Authorization header from the frontend to the main server
	proxyReq.Header.Set("Authorization", frontendAuthHeader)
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)


	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error sending logout request to main server: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// Copy headers from the main server's response (e.g., Set-Cookie to clear authToken)
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode) // Send status code from main server
	// Copy body if any (usually logout is empty or a simple message)
	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("Error copying logout response body: %v\n", err)
	}
}

// sessionHandler handles requests to /api/ui/auth/session
func sessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the Authorization header received from the frontend
	frontendAuthHeader := r.Header.Get("Authorization")
	if frontendAuthHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Construct request to the main server's /orgs/session endpoint
	mainServerSessionURL := mainServerURL + "/orgs/session"

	proxyReq, err := http.NewRequest(http.MethodGet, mainServerSessionURL, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating proxy request for session check: %v", err), http.StatusInternalServerError)
		return
	}

	// Forward the exact Authorization header from the frontend to the main server
	proxyReq.Header.Set("Authorization", frontendAuthHeader)
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error sending session check request to main server: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy all headers from the main server's response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	
	w.WriteHeader(resp.StatusCode) // Send status code from main server

	// Copy the body (session data or error) from the main server's response
	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("Error copying session response body: %v\n", err)
	}
}
