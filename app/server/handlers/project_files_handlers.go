package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"plandex-server/types" // For ServerAuth
	"plandex/shared"       // For shared types like ApiError
	"strings"

	"github.com/gorilla/mux"
)

// getManagedFilesBaseDir constructs the base directory path for a project's managed files.
// PLANDEX_BASE_DIR should be an environment variable or a globally configured path.
func getManagedFilesBaseDir(orgID, projectID string) (string, error) {
	if orgID == "" {
		return "", fmt.Errorf("organization ID is required")
	}
	if projectID == "" {
		return "", fmt.Errorf("project ID is required")
	}

	baseDir := os.Getenv("PLANDEX_BASE_DIR")
	if baseDir == "" {
		baseDir = "./data" // Default if not set
		// It's better to log this or handle it as a startup error elsewhere
		// fmt.Println("PLANDEX_BASE_DIR not set, using default ./data")
	}
	return filepath.Join(baseDir, "orgs", orgID, "projects", projectID, "managed_files"), nil
}

// UploadProjectFileHandler handles file uploads for a project.
func UploadProjectFileHandler(w http.ResponseWriter, r *http.Request) {
	auth := Authenticate(w, r, true) // requireOrg = true
	if auth == nil {
		// Authenticate writes error to w and returns nil if auth fails
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectId"]
	if projectID == "" {
		WriteJSONError(w, shared.ApiError{Msg: "Missing projectId", Status: http.StatusBadRequest})
		return
	}

	// Authorize that the authenticated user/org has access to this project
	if !authorizeProject(w, projectID, auth) {
		// authorizeProject writes error to w if auth fails
		return
	}

	baseDir, err := getManagedFilesBaseDir(auth.OrgId, projectID)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to determine storage directory: %v", err), Status: http.StatusInternalServerError})
		return
	}

	if err := os.MkdirAll(baseDir, 0750); err != nil { // Changed permissions to be more restrictive
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to create storage directory: %v", err), Status: http.StatusInternalServerError})
		return
	}

	// Parse the multipart form (max 32MB, adjust as needed)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to parse multipart form: %v", err), Status: http.StatusBadRequest})
		return
	}

	file, handler, err := r.FormFile("file") // "file" is the name of the form field
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Error retrieving file from form: %v", err), Status: http.StatusBadRequest})
		return
	}
	defer file.Close()

	// Sanitize filename to prevent path traversal issues
	// Also ensure filename is not empty or just "."
	filename := filepath.Clean(filepath.Base(handler.Filename)) // Use filepath.Base to remove directory parts
	if filename == "." || filename == "" || strings.Contains(filename, "..") {
		WriteJSONError(w, shared.ApiError{Msg: "Invalid filename", Status: http.StatusBadRequest})
		return
	}

	filePath := filepath.Join(baseDir, filename)

	// Check if file already exists; decide on overwrite policy (currently overwrites)
	// if _, err := os.Stat(filePath); err == nil {
	// 	WriteJSONError(w, shared.ApiError{Msg: "File already exists", Status: http.StatusConflict})
	// 	return
	// }

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to create file on server: %v", err), Status: http.StatusInternalServerError})
		return
	}
	defer dst.Close()

	// Copy the uploaded file data
	if _, err := io.Copy(dst, file); err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to save uploaded file: %v", err), Status: http.StatusInternalServerError})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "File uploaded successfully", "filename": filename})
}

// ListProjectFilesHandler lists files for a project.
func ListProjectFilesHandler(w http.ResponseWriter, r *http.Request) {
	auth := Authenticate(w, r, true)
	if auth == nil {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectId"]
	if projectID == "" {
		WriteJSONError(w, shared.ApiError{Msg: "Missing projectId", Status: http.StatusBadRequest})
		return
	}

	if !authorizeProject(w, projectID, auth) {
		return
	}

	baseDir, err := getManagedFilesBaseDir(auth.OrgId, projectID)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to determine storage directory: %v", err), Status: http.StatusInternalServerError})
		return
	}

	dirEntries, err := os.ReadDir(baseDir) // Changed from files to dirEntries to match os.ReadDir return type
	if err != nil {
		if os.IsNotExist(err) { // If directory doesn't exist, means no files
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]interface{}{}) // Return empty list
			return
		}
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to read storage directory: %v", err), Status: http.StatusInternalServerError})
		return
	}

	type FileInfo struct {
		Name         string `json:"name"`
		Size         int64  `json:"size"`
		LastModified int64  `json:"lastModified"` // Unix timestamp
	}

	var fileInfos []FileInfo
	for _, entry := range dirEntries { // Changed from file to entry
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				// Log error but continue if possible
				// Consider how to handle this in production (e.g., structured logging)
				fmt.Printf("Error getting file info for %s: %v\n", entry.Name(), err)
				continue
			}
			fileInfos = append(fileInfos, FileInfo{
				Name:         entry.Name(),
				Size:         info.Size(),
				LastModified: info.ModTime().Unix(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfos)
}

// DownloadProjectFileHandler serves a specific file for download.
func DownloadProjectFileHandler(w http.ResponseWriter, r *http.Request) {
	auth := Authenticate(w, r, true)
	if auth == nil {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectId"]
	filename := vars["filename"]
	if projectID == "" || filename == "" {
		WriteJSONError(w, shared.ApiError{Msg: "Missing projectId or filename", Status: http.StatusBadRequest})
		return
	}

	if !authorizeProject(w, projectID, auth) {
		return
	}

	baseDir, err := getManagedFilesBaseDir(auth.OrgId, projectID)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to determine storage directory: %v", err), Status: http.StatusInternalServerError})
		return
	}

	// Sanitize filename
	cleanFilename := filepath.Clean(filepath.Base(filename))
	if cleanFilename == "." || cleanFilename == "" || strings.Contains(cleanFilename, "..") {
		WriteJSONError(w, shared.ApiError{Msg: "Invalid filename", Status: http.StatusBadRequest})
		return
	}
	filePath := filepath.Join(baseDir, cleanFilename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		WriteJSONError(w, shared.ApiError{Msg: "File not found", Status: http.StatusNotFound})
		return
	}

	// Set Content-Disposition for download.
	// Consider setting Content-Type based on file extension if necessary.
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", cleanFilename))
	http.ServeFile(w, r, filePath)
}

// DeleteProjectFileHandler deletes a specific file.
func DeleteProjectFileHandler(w http.ResponseWriter, r *http.Request) {
	auth := Authenticate(w, r, true)
	if auth == nil {
		return
	}

	vars := mux.Vars(r)
	projectID := vars["projectId"]
	filename := vars["filename"]
	if projectID == "" || filename == "" {
		WriteJSONError(w, shared.ApiError{Msg: "Missing projectId or filename", Status: http.StatusBadRequest})
		return
	}

	if !authorizeProject(w, projectID, auth) {
		return
	}

	baseDir, err := getManagedFilesBaseDir(auth.OrgId, projectID)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to determine storage directory: %v", err), Status: http.StatusInternalServerError})
		return
	}

	// Sanitize filename
	cleanFilename := filepath.Clean(filepath.Base(filename))
	if cleanFilename == "." || cleanFilename == "" || strings.Contains(cleanFilename, "..") {
		WriteJSONError(w, shared.ApiError{Msg: "Invalid filename", Status: http.StatusBadRequest})
		return
	}
	filePath := filepath.Join(baseDir, cleanFilename)

	// Check if file exists before attempting delete
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		WriteJSONError(w, shared.ApiError{Msg: "File not found", Status: http.StatusNotFound})
		return
	}

	err = os.Remove(filePath)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to delete file: %v", err), Status: http.StatusInternalServerError})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File deleted successfully"})
}
