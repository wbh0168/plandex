package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"context"
	"plandex-server/db"
	"plandex-server/model" // For model interaction logic
	"plandex-server/types" // For types.ExtendedChatCompletionRequest
	"plandex/shared"       // For shared types like ApiError, ModelProvider etc.
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai" // For openai.ChatMessageRoleSystem etc.
)

// CodeInteractionRequest struct for incoming requests
type CodeInteractionRequest struct {
	ModelID     shared.ModelId `json:"model_id"` // Using ModelId type from shared
	Prompt      string         `json:"prompt"`
	CodeContext string         `json:"code_context,omitempty"`
	Action      string         `json:"action"` // e.g., "generate_code", "explain_code"
}

// CodeInteractionResponse struct for sending back LLM's response
type CodeInteractionResponse struct {
	Content string `json:"content"` // Generated code, explanation, etc.
	Error   string `json:"error,omitempty"`
}

// CodeInteractionHandler handles LLM-powered code interactions
func CodeInteractionHandler(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Add specific permission check if available, e.g., auth.HasPermission(shared.PermissionCodeInteract)

	var req CodeInteractionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Invalid request body: %v", err), Status: http.StatusBadRequest})
		return
	}

	if req.Prompt == "" || req.ModelID == "" || req.Action == "" {
		WriteJSONError(w, shared.ApiError{Msg: "Missing required fields: model_id, prompt, or action", Status: http.StatusBadRequest})
		return
	}

	// --- 1. Retrieve Model Configuration ---
	dbModel, err := db.GetAvailableModelByModelID(auth.OrgId, req.ModelID)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to retrieve model configuration for %s: %v", req.ModelID, err), Status: http.StatusInternalServerError})
		return
	}
	if dbModel == nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Model configuration not found for %s in org %s", req.ModelID, auth.OrgId), Status: http.StatusNotFound})
		return
	}

	// Convert db.AvailableModel to shared.ModelRoleConfig (or directly use shared.BaseModelConfig)
	// For now, we'll construct a temporary ModelRoleConfig as model.CreateChatCompletionStream expects it.
	// In a real scenario, this might come from a ModelPack or be constructed more robustly.
	tempModelRoleConfig := &shared.ModelRoleConfig{
		BaseModelConfig: shared.BaseModelConfig{
			Provider:                   dbModel.Provider,
			CustomProvider:             dbModel.CustomProvider,
			BaseUrl:                    dbModel.BaseUrl,
			ModelName:                  dbModel.ModelName,
			ModelId:                    dbModel.ModelId,
			MaxTokens:                  dbModel.MaxTokens,
			MaxOutputTokens:            dbModel.MaxOutputTokens,
			ReservedOutputTokens:       dbModel.ReservedOutputTokens,
			ApiKeyEnvVar:               dbModel.ApiKeyEnvVar,
			PreferredModelOutputFormat: dbModel.PreferredOutputFormat,
			// SystemPromptDisabled, RoleParamsDisabled, etc., would need to be set if they exist on dbModel
			// For simplicity, they are omitted here but should be mapped from dbModel if present.
			ModelCompatibility: shared.ModelCompatibility{
				HasImageSupport: dbModel.HasImageSupport,
			},
		},
		// Temperature, TopP, etc., for ModelRoleConfig could be default values or configured elsewhere.
		Temperature: 0.7, // Example
	}

	// --- 2. Construct LLM Prompt ---
	// This is highly dependent on the `action` and the desired LLM behavior.
	// Example:
	var fullPrompt strings.Builder
	fullPrompt.WriteString(fmt.Sprintf("User Request (%s):\n%s\n\n", req.Action, req.Prompt))
	if req.CodeContext != "" {
		fullPrompt.WriteString(fmt.Sprintf("Current Code Context:\n```\n%s\n```\n\n", req.CodeContext))
	}

	switch req.Action {
	case "generate_code":
		fullPrompt.WriteString("Please generate the code based on the request and context above.")
	case "explain_code":
		if req.CodeContext == "" {
			WriteJSONError(w, shared.ApiError{Msg: "Code context is required for 'explain_code' action", Status: http.StatusBadRequest})
			return
		}
		fullPrompt.WriteString("Please explain the provided code context based on the user request.")
	default:
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Unsupported action: %s", req.Action), Status: http.StatusBadRequest})
		return
	}

	// --- 3. LLM Interaction Logic ---
	// This requires using Plandex's existing LLM client.
	// model.CreateChatCompletionStream is the target.

	// Construct types.ExtendedChatCompletionRequest
	// The exact structure of messages (system prompt, user prompt) might need adjustment
	// based on how Plandex typically structures them for different actions.
	chatMessages := []types.ExtendedChatMessage{
		{Role: openai.ChatMessageRoleSystem, Content: []types.ExtendedChatMessagePart{{Type: "text", Text: "You are a helpful AI programming assistant."}}},
		{Role: openai.ChatMessageRoleUser, Content: []types.ExtendedChatMessagePart{{Type: "text", Text: fullPrompt.String()}}},
	}

	// TODO: Determine how API keys and clients are managed and passed to CreateChatCompletionStream.
	// model.InitClients seems to initialize a map of clients. This map needs to be accessible here.
	// This is a major dependency. For now, we'll assume it's globally available or passed via context.
	// This part is a placeholder and needs to be replaced with actual client map retrieval.
	clientsMap := model.GetClientsMap() // Placeholder for actual client map retrieval
	if clientsMap == nil {
		// Fallback: try to initialize clients with environment variables if GetClientsMap() is a placeholder
		// This is a temporary measure for this subtask and might not reflect the actual Plandex setup.
		apiKey := os.Getenv(tempModelRoleConfig.BaseModelConfig.ApiKeyEnvVar)
		if apiKey == "" {
			WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("API key not found for env var: %s", tempModelRoleConfig.BaseModelConfig.ApiKeyEnvVar), Status: http.StatusInternalServerError})
			return
		}
		clientsMap = make(map[string]model.ClientInfo)
		// Assuming OpenAI for now if not specified, or use tempModelRoleConfig.BaseModelConfig.BaseUrl
		var endpoint string
		if tempModelRoleConfig.BaseModelConfig.Provider == shared.ModelProviderOpenAI {
			endpoint = openai.DefaultConfig("").BaseURL // Use default OpenAI if no specific BaseUrl
		}
		if tempModelRoleConfig.BaseModelConfig.BaseUrl != "" {
			endpoint = tempModelRoleConfig.BaseModelConfig.BaseUrl
		}
		clientsMap[tempModelRoleConfig.BaseModelConfig.ApiKeyEnvVar] = model.NewClient(apiKey, endpoint, "") // Assuming no specific OpenAI org ID for this client
		// model.SetClientsMap(clientsMap) // If there's a setter for a global map
	}


	// Prepare the request for the LLM stream
	llmRequest := types.ExtendedChatCompletionRequest{
		Model:       string(tempModelRoleConfig.BaseModelConfig.ModelId), // ModelId is like "gpt-4-turbo"
		Messages:    chatMessages,
		Temperature: float32(tempModelRoleConfig.Temperature),
		MaxTokens:   tempModelRoleConfig.BaseModelConfig.MaxOutputTokens,
		Stream:      true, // Important: model.CreateChatCompletionStream expects to create a stream
	}

	// Timeout for the entire LLM interaction, including stream reading
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	stream, err := model.CreateChatCompletionStream(clientsMap, tempModelRoleConfig, ctx, llmRequest)
	if err != nil {
		WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Failed to create LLM stream: %v", err), Status: http.StatusInternalServerError})
		return
	}
	defer stream.Close()

	var llmResponseContent strings.Builder
	for {
		responseChunk, err := stream.Recv()
		if err != nil {
			// io.EOF means stream ended successfully
			if err.Error() == "EOF" || err.Error() == "context deadline exceeded" { // Handle timeout as EOF for content
				break
			}
			WriteJSONError(w, shared.ApiError{Msg: fmt.Sprintf("Error receiving from LLM stream: %v", err), Status: http.StatusInternalServerError})
			return
		}
		if len(responseChunk.Choices) > 0 && len(responseChunk.Choices[0].Delta.Content) > 0 {
			llmResponseContent.WriteString(responseChunk.Choices[0].Delta.Content[0].Text)
		}
		// Check for usage or other non-content parts if necessary from responseChunk
	}
	
	if ctx.Err() == context.DeadlineExceeded {
        WriteJSONError(w, shared.ApiError{Msg: "LLM request timed out while reading stream", Status: http.StatusGatewayTimeout})
        return
    }

	response := CodeInteractionResponse{Content: llmResponseContent.String()}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding code interaction response: %v\n", err)
	}
}

// Helper function to get the first N characters of a string (for logging/dummy response)
// func firstN(s string, n int) string {
// 	if len(s) <= n {
// 		return s
// 	}
// 	return s[:n]
// }

// Placeholder for model.GetClientsMap() - this needs to be implemented based on Plandex's actual client management.
// It might involve a global map initialized at startup, or passed through request context.
// For now, this is a stub.
// var globalClientsMap map[string]model.ClientInfo
// func InitGlobalClients(apiKeys map[string]string, endpointsByApiKeyEnvVar map[string]string, openAIEndpoint, orgId string) {
// 	globalClientsMap = model.InitClients(apiKeys, endpointsByApiKeyEnvVar, openAIEndpoint, orgId)
// }
// func GetClientsMap() map[string]model.ClientInfo {
// 	return globalClientsMap
// }
