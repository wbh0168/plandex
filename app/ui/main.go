package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"plandex/shared" // Assuming shared types are accessible

	"github.com/gorilla/mux"
)

const mainServerURL = "http://localhost:8080" // Adjust if your main server runs elsewhere

// Helper function to proxy requests
func proxyRequest(method, path string, body io.Reader, w http.ResponseWriter, r *http.Request) {
	// Construct the target URL
	targetURL := mainServerURL + path // mainServerURL is global

	// Create a new request to the main server
	proxyReq, err := http.NewRequest(method, targetURL, body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating proxy request: %v", err), http.StatusInternalServerError)
		return
	}

	// Forward essential headers, including Authorization
	// The frontend is expected to construct and send the full "Authorization: Bearer <base64_encoded_json_auth_header>"
	// The UI backend will forward this exact header to the main Plandex server.
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		proxyReq.Header.Set("Authorization", authHeader)
	}

	// Forward Content-Type if present, especially for POST/PUT with a body
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		proxyReq.Header.Set("Content-Type", contentType)
	}
	
	// Add X-Forwarded-For
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)


	// Send the request using the default client
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error sending proxy request: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers from the proxy response to our response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the body from the proxy response to our response
	if _, err := io.Copy(w, resp.Body); err != nil {
		// Log error, but the header and status are already sent
		fmt.Printf("Error copying response body: %v\n", err)
	}
}

// --- Custom Models Handlers ---

func handleCustomModels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		proxyRequest(http.MethodGet, "/custom_models", nil, w, r)
	case http.MethodPost:
		var model shared.AvailableModel
		if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
			http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
			return
		}
		bodyBytes, _ := json.Marshal(model)
		proxyRequest(http.MethodPost, "/custom_models", bytes.NewBuffer(bodyBytes), w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCustomModelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID := vars["modelId"]
	if modelID == "" {
		http.Error(w, "Missing modelId", http.StatusBadRequest)
		return
	}
	path := "/custom_models/" + modelID

	switch r.Method {
	case http.MethodPut:
		var model shared.AvailableModel
		if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
			http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
			return
		}
		bodyBytes, _ := json.Marshal(model)
		proxyRequest(http.MethodPut, path, bytes.NewBuffer(bodyBytes), w, r)
	case http.MethodDelete:
		proxyRequest(http.MethodDelete, path, nil, w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Model Packs Handlers ---

func handleModelPacks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// The main server endpoint is /model_sets
		proxyRequest(http.MethodGet, "/model_sets", nil, w, r)
	case http.MethodPost:
		var modelPack shared.ModelPack
		if err := json.NewDecoder(r.Body).Decode(&modelPack); err != nil {
			http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
			return
		}
		bodyBytes, _ := json.Marshal(modelPack)
		// The main server endpoint is /model_sets
		proxyRequest(http.MethodPost, "/model_sets", bytes.NewBuffer(bodyBytes), w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleModelPackByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packID := vars["packId"] // Corrected from modelId to packId
	if packID == "" {
		http.Error(w, "Missing packId", http.StatusBadRequest)
		return
	}
	// The main server endpoint is /model_sets/{setId}
	path := "/model_sets/" + packID

	switch r.Method {
	case http.MethodPut:
		var modelPack shared.ModelPack
		if err := json.NewDecoder(r.Body).Decode(&modelPack); err != nil {
			http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
			return
		}
		bodyBytes, _ := json.Marshal(modelPack)
		proxyRequest(http.MethodPut, path, bytes.NewBuffer(bodyBytes), w, r)
	case http.MethodDelete:
		proxyRequest(http.MethodDelete, path, nil, w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	r := mux.NewRouter()

	// API routes for UI backend
	apiRouter := r.PathPrefix("/api/ui").Subrouter()

	// Custom Models routes
	apiRouter.HandleFunc("/custom-models", handleCustomModels).Methods(http.MethodGet, http.MethodPost)
	apiRouter.HandleFunc("/custom-models/{modelId}", handleCustomModelByID).Methods(http.MethodPut, http.MethodDelete)

	// Model Packs routes
	apiRouter.HandleFunc("/model-packs", handleModelPacks).Methods(http.MethodGet, http.MethodPost)
	apiRouter.HandleFunc("/model-packs/{packId}", handleModelPackByID).Methods(http.MethodPut, http.MethodDelete)

	// Project File Management routes
	// Note: gorilla/mux processes routes in order. More specific routes should come before less specific ones
	// if there's a chance of overlap, though here it's distinct by path structure.
	apiRouter.HandleFunc("/projects/{projectId}/files/upload", handleProjectFileUpload).Methods(http.MethodPost)
	apiRouter.HandleFunc("/projects/{projectId}/files/download/{filename}", handleProjectFileDownload).Methods(http.MethodGet) // Order matters: specific download before general files list if paths could clash
	apiRouter.HandleFunc("/projects/{projectId}/files/{filename}", handleProjectFileDelete).Methods(http.MethodDelete)
	apiRouter.HandleFunc("/projects/{projectId}/files", handleProjectFilesList).Methods(http.MethodGet)

	// Project Command Execution route
	apiRouter.HandleFunc("/projects/{projectId}/commands", handleProjectCommandExecute).Methods(http.MethodPost)

	// Code Interaction route
	apiRouter.HandleFunc("/projects/{projectId}/code/interact", handleCodeInteraction).Methods(http.MethodPost)

	// Auth routes
	authRouter := r.PathPrefix("/api/ui/auth").Subrouter()
	authRouter.HandleFunc("/login", loginHandler).Methods(http.MethodPost)
	authRouter.HandleFunc("/logout", logoutHandler).Methods(http.MethodPost)
	authRouter.HandleFunc("/session", sessionHandler).Methods(http.MethodGet)

	// Original simple handler (can be removed or kept for basic check)
	// r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello from UI backend")
	// })

	// Serve Svelte app (static files)
	// This setup assumes your Svelte app is built into the 'public' or 'dist' directory
	// and files are served from 'app/ui/frontend/my-svelte-app/public' or 'app/ui/frontend/my-svelte-app/dist'
	// For development with Vite, Vite's dev server handles serving. For production, you build static assets.
	staticDir := http.Dir("../frontend/my-svelte-app/dist") // Adjust if your build output is different

	// Fallback for Svelte Router (Single Page Application)
	// Serve static files, and for any path not matched by API and not a static file, serve index.html
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for an API endpoint
		if strings.HasPrefix(r.URL.Path, "/api") {
			// This case should ideally not be reached if mux routes API calls correctly.
			// If it is, it means no API route matched.
			http.NotFound(w, r)
			return
		}

		// Attempt to serve static file
		staticPath := r.URL.Path
		// If path is "/", serve "index.html"
		if staticPath == "/" {
			staticPath = "index.html"
		}
		
		// Open the file. staticDir is already an http.Dir, so it handles path joining.
		file, err := staticDir.Open(staticPath)
		if err != nil {
			if os.IsNotExist(err) {
				// File does not exist, serve index.html for SPA routing
				http.ServeFile(w, r, filepath.Join(string(staticDir), "index.html"))
				return
			}
			// Any other error (e.g., permission issues)
			http.Error(w, fmt.Sprintf("Could not open static file: %v", err), http.StatusInternalServerError)
			return
		}
		file.Close() // Close after check, http.FileServer will reopen.

		// If file exists, serve it.
		http.FileServer(staticDir).ServeHTTP(w, r)
	})

	fmt.Println("UI backend server starting on :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

// --- Project File Management Handlers ---

// handleProjectFileUpload proxies file uploads to the main server
func handleProjectFileUpload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectId"]
	if projectID == "" {
		http.Error(w, "Missing projectId in UI backend", http.StatusBadRequest)
		return
	}

	// The path to the main server endpoint
	targetPath := fmt.Sprintf("/projects/%s/files/upload", projectID)

	// We need to pass through the multipart form data.
	// http.Request.Body is an io.ReadCloser.
	// We also need to pass through the Content-Type header, which includes the boundary.
	proxyRequest(r.Method, targetPath, r.Body, w, r)
}

// handleProjectFilesList proxies list requests to the main server
func handleProjectFilesList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectId"]
	if projectID == "" {
		http.Error(w, "Missing projectId in UI backend", http.StatusBadRequest)
		return
	}
	targetPath := fmt.Sprintf("/projects/%s/files", projectID)
	proxyRequest(http.MethodGet, targetPath, nil, w, r)
}

// handleProjectFileDownload proxies download requests to the main server
func handleProjectFileDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectId"]
	filename := vars["filename"]
	if projectID == "" || filename == "" {
		http.Error(w, "Missing projectId or filename in UI backend", http.StatusBadRequest)
		return
	}
	targetPath := fmt.Sprintf("/projects/%s/files/download/%s", projectID, filename)
	// Use the generic proxyRequest. It will copy headers like Content-Disposition.
	proxyRequest(http.MethodGet, targetPath, nil, w, r)
}

// handleProjectFileDelete proxies delete requests to the main server
func handleProjectFileDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectId"]
	filename := vars["filename"]
	if projectID == "" || filename == "" {
		http.Error(w, "Missing projectId or filename in UI backend", http.StatusBadRequest)
		return
	}
	targetPath := fmt.Sprintf("/projects/%s/files/%s", projectID, filename)
	proxyRequest(http.MethodDelete, targetPath, nil, w, r)
}

// --- Project Command Execution Handler ---

func handleProjectCommandExecute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectId"]
	if projectID == "" {
		http.Error(w, "Missing projectId in UI backend for command execution", http.StatusBadRequest)
		return
	}

	// The path to the main server endpoint for command execution
	targetPath := fmt.Sprintf("/projects/%s/commands", projectID)

	// We need to pass through the JSON body.
	// proxyRequest handles copying the body and necessary headers.
	proxyRequest(r.Method, targetPath, r.Body, w, r)
}

// --- Code Interaction Handler (UI Backend) ---
func handleCodeInteraction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectId"]
	if projectID == "" {
		http.Error(w, "Missing projectId in UI backend for code interaction", http.StatusBadRequest)
		return
	}

	targetPath := fmt.Sprintf("/projects/%s/code/interact", projectID)
	proxyRequest(r.Method, targetPath, r.Body, w, r)
}
