package db

import (
	"database/sql"
	"errors"
	"regexp" // Required for go-sqlmock query matching
	"testing"
	"time"

	"plandex/shared"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestGetAvailableModelByModelID(t *testing.T) {
	// Create mock DB and use it for the global Conn
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	// sqlx.NewDb wraps the standard sql.DB
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	
	// Store original Conn and replace with mock, then restore
	originalConn := Conn
	Conn = sqlxDB
	defer func() { Conn = originalConn }()

	// Define a sample model for testing
	sampleModel := AvailableModel{
		Id:                    "test-uuid",
		OrgId:                 "org-123",
		Provider:              shared.ModelProviderOpenAI,
		CustomProvider:        nil,
		BaseUrl:               "https://api.openai.com/v1",
		ModelId:               "gpt-4-turbo",
		ModelName:             "GPT-4 Turbo",
		Description:           "Test model",
		MaxTokens:             8000,
		ApiKeyEnvVar:          "OPENAI_API_KEY",
		DefaultMaxConvoTokens: 4000,
		MaxOutputTokens:       2000,
		ReservedOutputTokens:  100,
		HasImageSupport:       false,
		PreferredOutputFormat: shared.ModelOutputFormatXml,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	query := `SELECT \* FROM custom_models WHERE org_id = \$1 AND model_id = \$2`

	tests := []struct {
		name          string
		orgID         string
		modelID       shared.ModelId
		mockSetup     func()
		wantErr       bool
		expectedModel *AvailableModel
	}{
		{
			name:    "successful retrieval",
			orgID:   "org-123",
			modelID: "gpt-4-turbo",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "org_id", "provider", "custom_provider", "base_url", "model_id", "model_name", "description", "max_tokens", "api_key_env_var", "default_max_convo_tokens", "max_output_tokens", "reserved_output_tokens", "has_image_support", "preferred_output_format", "created_at", "updated_at"}).
					AddRow(sampleModel.Id, sampleModel.OrgId, sampleModel.Provider, sampleModel.CustomProvider, sampleModel.BaseUrl, sampleModel.ModelId, sampleModel.ModelName, sampleModel.Description, sampleModel.MaxTokens, sampleModel.ApiKeyEnvVar, sampleModel.DefaultMaxConvoTokens, sampleModel.MaxOutputTokens, sampleModel.ReservedOutputTokens, sampleModel.HasImageSupport, sampleModel.PreferredOutputFormat, sampleModel.CreatedAt, sampleModel.UpdatedAt)
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("org-123", "gpt-4-turbo").WillReturnRows(rows)
			},
			wantErr:       false,
			expectedModel: &sampleModel,
		},
		{
			name:    "model not found",
			orgID:   "org-123",
			modelID: "non-existent-model",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("org-123", "non-existent-model").WillReturnError(sql.ErrNoRows)
			},
			wantErr:       false, // Should return nil, nil for not found
			expectedModel: nil,
		},
		{
			name:    "database error",
			orgID:   "org-123",
			modelID: "gpt-4-turbo",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("org-123", "gpt-4-turbo").WillReturnError(errors.New("some db error"))
			},
			wantErr:       true,
			expectedModel: nil,
		},
		{
			name:          "empty orgID",
			orgID:         "",
			modelID:       "gpt-4-turbo",
			mockSetup:     func() {}, // No DB call expected
			wantErr:       true,
			expectedModel: nil,
		},
		{
			name:          "empty modelID",
			orgID:         "org-123",
			modelID:       "",
			mockSetup:     func() {}, // No DB call expected
			wantErr:       true,
			expectedModel: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			gotModel, err := GetAvailableModelByModelID(tt.orgID, tt.modelID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAvailableModelByModelID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare actual model with expected model
			// Note: Comparing structs directly can be tricky due to time.Time or pointer fields.
			// For this test, simple comparison or specific field checks might be enough.
			if tt.expectedModel == nil && gotModel != nil {
				t.Errorf("GetAvailableModelByModelID() got = %v, want nil", gotModel)
			}
			if tt.expectedModel != nil && gotModel == nil {
				t.Errorf("GetAvailableModelByModelID() got nil, want %v", tt.expectedModel)
			}
			if tt.expectedModel != nil && gotModel != nil {
				// Compare key fields
				if gotModel.Id != tt.expectedModel.Id || gotModel.ModelId != tt.expectedModel.ModelId || gotModel.OrgId != tt.expectedModel.OrgId {
					t.Errorf("GetAvailableModelByModelID() got = %+v, want %+v", gotModel, tt.expectedModel)
				}
			}
			
			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
