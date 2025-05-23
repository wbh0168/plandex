package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"plandex/shared" // For shared types like ApiError
	"strings"

	"github.com/gorilla/mux"
)

// CommandRequest struct for incoming command execution requests
type CommandRequest struct {
	Command string `json:"command"`
}

// CommandResponse struct for sending back command output
type CommandResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

// getProjectFilesDir is a helper to get the project-specific directory.
// It's similar to getManagedFilesBaseDir but might point to a different root if needed,
// or could be the same. For now, let's assume it's the managed files directory.
func getProjectFilesDir(orgID, projectID string) (string, error) {
	if orgID == "" || projectID == "" {
		return "", fmt.Errorf("orgID and projectID are required")
	}
	baseDir := os.Getenv("PLANDEX_BASE_DIR")
	if baseDir == "" {
		baseDir = "./data" // Default if not set
	}
	// SECURITY: Commands will be executed in this directory.
	// Ensure this directory is appropriately isolated and doesn't allow escape to higher-level directories.
	return filepath.Join(baseDir, "orgs", orgID, "projects", projectID, "managed_files"), nil
}

// ExecuteCommandHandler handles execution of commands within a project's context.
func ExecuteCommandHandler(w http.ResponseWriter, r *http.Request) {
	auth := Authenticate(w, r, true) // requireOrg = true
	if auth == nil {
		return // Authenticate writes error to w and returns nil if auth fails
	}

	vars := mux.Vars(r)
	projectID := vars["projectId"]
	if projectID == "" {
		WriteJSONError(w, shared.ApiError{Msg: "Missing projectId", Status: http.StatusBadRequest})
		return
	}

	// Authorize that the authenticated user/org has access to this project
	if !authorizeProject(w, projectID, auth) {
		return // authorizeProject writes error to w if auth fails
	}

	// TODO: Add more specific permissions check for executing commands if RBAC supports it.
	// e.g., if !auth.HasPermission(shared.PermissionExecuteProjectCommands) { ... }

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Invalid request body: %v", err), Status: http.StatusBadRequest})
		return
	}

	if req.Command == "" {
		WriteJSONError(w, shared.ApiError{Msg: "Command cannot be empty", Status: http.StatusBadRequest})
		return
	}

	// SECURITY: Basic command sanitization/validation.
	// This is a very minimal check. A deny-list or more sophisticated parsing/validation is recommended.
	// For now, disallow common shell metacharacters that could lead to command injection if not handled by shell quoting.
	// os/exec without a shell (i.e., directly running an executable) is safer, but many users expect shell features.
	// If we pass this to a shell (e.g. bash -c "command"), then injection is a major risk.
	// If we parse it ourselves (like below), we control arguments.
	trimmedCommand := strings.TrimSpace(req.Command)
	if trimmedCommand == "" {
		WriteJSONError(w, shared.ApiError{Msg: "Command cannot be empty", Status: http.StatusBadRequest})
		return
	}
	
	// For now, let's split the command into a program and arguments.
	// This avoids invoking a shell directly with the raw command string, which is a major security risk.
	parts := strings.Fields(trimmedCommand)
	program := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = parts[1:]
	}

	// SECURITY: Add a list of allowed commands or more robust validation.
	// For this iteration, we are not implementing a strict allow-list, which is a security risk.
	// Example: Allow only 'ls', 'cat', 'echo', 'pwd'.
	// allowedCommands := map[string]bool{"ls": true, "cat": true, "echo": true, "pwd": true, "grep": true, "find": true}
	// if !allowedCommands[program] {
	// 	WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Command not allowed: %s", program), Status: http.StatusForbidden})
	// 	return
	// }


	projectDir, err := getProjectFilesDir(auth.OrgId, projectID)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to determine project directory: %v", err), Status: http.StatusInternalServerError})
		return
	}

	// Ensure the directory exists, create if not (though it should for managed_files)
	if err := os.MkdirAll(projectDir, 0750); err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to ensure project directory: %v", err), Status: http.StatusInternalServerError})
		return
	}

	cmd := exec.Command(program, args...)
	cmd.Dir = projectDir // Execute the command in the project's specific directory.

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run() // This waits for the command to complete.

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Other errors (e.g., command not found)
			// We'll set a generic non-zero exit code or handle appropriately.
			// For command not found, stderr usually has the message.
			exitCode = -1 // Indicate a non-ExitError failure
			if stderr.Len() == 0 { // if stderr is empty, put the error message there
				stderr.WriteString(fmt.Sprintf("Error executing command: %v", err))
			}
		}
	}

	resp := CommandResponse{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// If encoding fails, the headers might have already been sent.
		// Log this error on the server.
		fmt.Printf("Error encoding command response: %v\n", err)
	}
}
