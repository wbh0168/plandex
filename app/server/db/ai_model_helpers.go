package db

import (
	"database/sql"
	"fmt"
	"plandex/shared" // For shared.ModelId type
)

// GetAvailableModelByModelID retrieves a specific custom model by its user-defined model_id string (e.g., "gpt-4-turbo")
// for a given organization.
func GetAvailableModelByModelID(orgID string, modelID shared.ModelId) (*AvailableModel, error) {
	if orgID == "" {
		return nil, fmt.Errorf("orgID cannot be empty")
	}
	if modelID == "" {
		return nil, fmt.Errorf("modelID cannot be empty")
	}

	var model AvailableModel
	query := `SELECT * FROM custom_models WHERE org_id = $1 AND model_id = $2`

	err := Conn.Get(&model, query, orgID, modelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if not found, so handler can return 404
		}
		return nil, fmt.Errorf("error fetching custom model by model_id '%s' for org_id '%s': %w", modelID, orgID, err)
	}

	return &model, nil
}
