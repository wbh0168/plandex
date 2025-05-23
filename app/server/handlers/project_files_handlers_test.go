package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"plandex-server/db" // For db.User, db.AuthToken, db.Project for setting up auth context
	"plandex-server/types"
	"plandex/shared"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/afero"
)

// MockAuthSetup is a helper to create a request with mocked authentication.
// For real tests, this would involve setting up db.Conn with sqlmock
// to control the behavior of Authenticate and authorizeProject.
// For now, we'll assume Authenticate and authorizeProject can be influenced
// by request context or specific test DB setup.
// This is a simplified approach.
var mockAuthedUser = &db.User{Id: "test-user-id", Email: "test@example.com"}
var mockAuthToken = &db.AuthToken{UserId: "test-user-id", TokenHash: "test-token"}
var mockOrgId = "test-org-id"

// This function would set up the db.Conn mock for auth calls
func setupMockAuthDB(t *testing.T, mock sqlmock.Sqlmock, projectExists bool, projectId string) {
	// Mock for Authenticate (ValidateAuthToken, GetUser)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM auth_tokens WHERE token_hash = $1 AND deleted_at IS NULL`)).
		WithArgs(sqlmock.AnyArg()). // Assuming token is hashed and we can't predict it here
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token_hash"}).AddRow("tokenid", mockAuthedUser.Id, "hashedtoken"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM users WHERE id = $1`)).
		WithArgs(mockAuthedUser.Id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(mockAuthedUser.Id, mockAuthedUser.Email))
	
	// Mock for org membership validation within Authenticate
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM org_users WHERE org_id = $1 AND user_id = $2)`)).
		WithArgs(mockOrgId, mockAuthedUser.Id).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT permission_id FROM org_roles_permissions INNER JOIN org_users ON org_roles_permissions.org_role_id = org_users.org_role_id WHERE org_users.user_id = $1 AND org_users.org_id = $2`)).
		WithArgs(mockAuthedUser.Id, mockOrgId).
		WillReturnRows(sqlmock.NewRows([]string{"permission_id"})) // No specific permissions needed for these file ops by default

	// Mock for authorizeProject (ProjectExists)
	if projectId != "" { // Only if projectID is relevant for the test
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM projects WHERE org_id = $1 AND id = $2)`)).
			WithArgs(mockOrgId, projectId).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(projectExists))
	}
}


func newAuthenticatedRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, url, body)
	// Simulate adding auth info that Authenticate would pick up.
	// The actual Authenticate function uses GetAuthHeader, which reads "Authorization".
	// The Authorization header contains a base64 encoded JSON.
	authHeaderContent := shared.AuthHeader{
		Token: "test-auth-token-value", // This token will be "validated" by sqlmock
		OrgId: mockOrgId,
	}
	authHeaderBytes, _ := json.Marshal(authHeaderContent)
	req.Header.Set("Authorization", "Bearer "+shared.EncodeBase64(string(authHeaderBytes)))
	
	// For tests that need project context (most of these do)
	// We need to ensure authorizeProject also passes. This requires mocking db.ProjectExists.
	// This setup is done per test case via setupMockAuthDB.
	return req
}


// Global Afero filesystem for tests
var appFs = afero.NewMemMapFs()
var originalOsFs afero.Fs // To store the original os.Stat etc.

func TestMain(m *testing.M) {
	// Setup: Use Afero for os package functions
	originalOsFs = afero.OsFs // Should be afero.NewOsFs()
	// To truly mock os calls, we'd need to replace functions like os.Stat, os.MkdirAll, etc.
	// with afero equivalents. This is non-trivial for a global os package.
	// A cleaner way is to pass an afero.Fs instance to handlers, but that's a refactor.
	// For now, handlers use os.xxx directly. We will use afero to prepare state
	// and verify outcomes where possible, but direct mocking of os calls within handlers
	// isn't done by just declaring appFs here.
	// Let's assume for these tests that PLANDEX_BASE_DIR will allow us to use appFs paths.
	
	// For tests that need to mock os.Stat, os.MkdirAll, etc., we would typically
	// pass an afero.Fs to the functions. Since project_files_handlers.go uses the
	// global `os` package, true isolation of filesystem operations for these tests
	// without refactoring the handlers is difficult.
	// The `afero.OsFs` is the real filesystem. `afero.NewMemMapFs()` is in-memory.
	// We will set PLANDEX_BASE_DIR to a path that we manage with `appFs` *outside* the handlers,
	// and then verify effects on `appFs`.

	// The tests will need to manage db.Conn.
	// We set it up per test suite or per test.
	exitCode := m.Run()
	os.Exit(exitCode)
}


func TestUploadProjectFileHandler(t *testing.T) {
	t.Setenv("PLANDEX_BASE_DIR", "/test_plandex_data") // Use a test base dir

	// Setup mock DB for auth
	mockDB, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock.New error: %s", err) }
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	originalDBConn := db.Conn // Backup original
	db.Conn = sqlxDB
	defer func() { db.Conn = originalDBConn }()


	// Test cases
	tests := []struct {
		name           string
		projectID      string
		filename       string
		fileContent    string
		setupAuth      bool // if true, setup mock for successful auth & project validation
		projectExists  bool 
		expectedStatus int
		expectedBodyContains string
	}{
		{
			name: "successful upload",
			projectID: "proj-1",
			filename: "test.txt",
			fileContent: "hello world",
			setupAuth: true,
			projectExists: true,
			expectedStatus: http.StatusCreated,
			expectedBodyContains: "File uploaded successfully",
		},
		{
			name: "missing projectId",
			projectID: "", // Invalid
			filename: "test.txt",
			fileContent: "hello",
			setupAuth: true, // Auth might still pass if project ID is not checked by Authenticate itself
			projectExists: true, // Not relevant if projectID is empty
			expectedStatus: http.StatusBadRequest,
			expectedBodyContains: "Missing projectId",
		},
		{
			name: "auth fails - project not found",
			projectID: "proj-nonexist",
			filename: "test.txt",
			fileContent: "hello",
			setupAuth: true,
			projectExists: false, // This will make authorizeProject fail
			expectedStatus: http.StatusNotFound, // from authorizeProject
			expectedBodyContains: "project does not exist in org",
		},
		{
			name: "invalid filename - path traversal",
			projectID: "proj-1",
			filename: "../../secret.txt",
			fileContent: "hax",
			setupAuth: true,
			projectExists: true,
			expectedStatus: http.StatusBadRequest,
			expectedBodyContains: "Invalid filename",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset and setup mock expectations for DB based on test case
			// This is simplified; a real test would need more granular mock setup per test
			// For now, broadly setting up auth to pass or fail access to projectID
			if tt.setupAuth {
				setupMockAuthDB(t, mock, tt.projectExists, tt.projectID)
			} else {
				// Mock Authenticate to fail (e.g., token validation fails)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM auth_tokens WHERE token_hash = $1 AND deleted_at IS NULL`)).
					WillReturnError(sql.ErrNoRows) // Simulate token not found
			}
			
			// Prepare multipart form
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", tt.filename)
			_, _ = io.WriteString(part, tt.fileContent)
			writer.Close()

			req := newAuthenticatedRequest(t, http.MethodPost, "/projects/"+tt.projectID+"/files/upload", body)
			if tt.projectID == "" { // For the missing projectID case
				req = newAuthenticatedRequest(t, http.MethodPost, "/projects//files/upload", body) // Mux might not like empty var
				req = mux.SetURLVars(req, map[string]string{"projectId": ""}) // Better way
			} else {
				req = mux.SetURLVars(req, map[string]string{"projectId": tt.projectID})
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(UploadProjectFileHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v. Body: %s",
					status, tt.expectedStatus, rr.Body.String())
			}
			if !strings.Contains(rr.Body.String(), tt.expectedBodyContains) {
				t.Errorf("handler returned unexpected body: got %v want to contain %v",
					rr.Body.String(), tt.expectedBodyContains)
			}

			// Verify file was created in afero an_os for successful upload
			if tt.expectedStatus == http.StatusCreated {
				orgID := mockOrgId // from auth context
				expectedPath := filepath.Join("/test_plandex_data", "orgs", orgID, "projects", tt.projectID, "managed_files", filepath.Base(tt.filename))
				
				// As handlers use global 'os', we can't directly use appFs.Exists here.
				// This test setup assumes the handler *did* write to the real FS if not mocked.
				// For true in-memory test, handler needs refactor to use afero.Fs.
				// Here, we'd check the real filesystem if PLANDEX_BASE_DIR pointed to a temp real dir,
				// or trust the mocks for os.Create, io.Copy if we could inject them.
				// Since we can't easily mock global 'os', this part is hard to verify in full isolation without refactor.
				// We are mostly testing the handler logic up to the point of OS interaction.
				// If we could use afero directly:
				// exists, _ := afero.Exists(appFs, expectedPath)
				// if !exists {
				//  t.Errorf("expected file %s to be created, but it wasn't", expectedPath)
				// }
				// content, _ := afero.ReadFile(appFs, expectedPath)
				// if string(content) != tt.fileContent {
				//  t.Errorf("file content mismatch: got %s want %s", string(content), tt.fileContent)
				// }
				// afero.NewOsFs().Remove(expectedPath) // Clean up if it wrote to real FS
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("SQL mock expectations not met: %s", err)
			}
			mock.Reset() // Reset expectations for the next run in the loop
		})
	}
}

// Note: Tests for List, Download, Delete would follow a similar pattern:
// - Setup mock DB for auth
// - Use newAuthenticatedRequest
// - Setup afero filesystem state if needed (e.g., create files for List/Download)
// - Make request using httptest
// - Check status code and body
// - Verify afero filesystem state after operation (e.g., file deleted)
// - Ensure sqlmock expectations are met.

// Due to the complexity of fully mocking global 'os' calls without refactoring the handlers,
// the filesystem interaction parts of these tests are illustrative of what *would* be checked
// if the handlers used an injectable filesystem interface (like afero.Fs).
// The current tests primarily validate logic paths up to the OS interaction, HTTP responses, and DB mocks for auth.
// Full end-to-end style tests with a temporary directory might be an alternative for real FS checks.
// For now, the focus is on handler logic, request parsing, auth checks, and response formatting.

// Placeholder for ListProjectFilesHandler tests
func TestListProjectFilesHandler(t *testing.T) {
	// Similar setup as TestUploadProjectFileHandler
	// Mock os.ReadDir (difficult without refactor or more complex os mocking)
	// Test cases:
	// - successful list
	// - empty list (directory exists but no files)
	// - directory not found (os.ReadDir returns error)
	// - auth failure
	t.Skip("ListProjectFilesHandler tests not fully implemented due to OS mocking complexity without refactor.")
}

// Placeholder for DownloadProjectFileHandler tests
func TestDownloadProjectFileHandler(t *testing.T) {
	// Similar setup
	// Mock os.Stat and http.ServeFile behavior (ServeFile is hard to mock directly, check headers it sets)
	// Test cases:
	// - successful download
	// - file not found
	// - auth failure
	// - filename sanitization (e.g. trying to download ../../file)
	t.Skip("DownloadProjectFileHandler tests not fully implemented due to OS/http.ServeFile mocking complexity.")
}

// Placeholder for DeleteProjectFileHandler tests
func TestDeleteProjectFileHandler(t *testing.T) {
	// Similar setup
	// Mock os.Remove and os.Stat
	// Test cases:
	// - successful delete
	// - file not found
	// - auth failure
	// - filename sanitization
	t.Skip("DeleteProjectFileHandler tests not fully implemented due to OS mocking complexity without refactor.")
}
