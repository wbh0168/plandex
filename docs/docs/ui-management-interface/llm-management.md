---
sidebar_label: 'LLM Management'
sidebar_position: 2
---

# LLM Management

The UI Management Interface allows you to manage your Large Language Models (LLMs) effectively. This includes managing Custom Models and Model Packs, which are used to define how Plandex interacts with different language models for various tasks.

## Accessing LLM Management

From the main navigation bar in the UI, you can access the LLM management features. Typically, these are found under "Custom Models" and "Model Packs" menu items.

[Screenshot: Main navigation bar highlighting "Custom Models" and "Model Packs"]

## Custom Models

Custom models are individual language model configurations. You define how Plandex should connect to and use a specific LLM API, whether it's from OpenAI, Anthropic, Google, or a custom provider (like OpenRouter or a self-hosted model).

### Listing Custom Models

Upon navigating to the "Custom Models" section, you will see a table listing all currently configured custom models.

*   **Columns:**
    *   **Model ID (User Defined):** The unique identifier you assigned to the model (e.g., `gpt-4-turbo`, `claude-3-opus`). This is used when referencing the model in Model Packs or other configurations.
    *   **Display Name:** A user-friendly name for the model (e.g., `GPT-4 Turbo`, `Claude Opus`).
    *   **Provider:** The source of the model (e.g., OpenAI, Anthropic, Google, Custom).
    *   **Description:** A brief description of the model.
    *   **Actions:** Buttons to "Edit" or "Delete" the model configuration.

[Screenshot: Custom Models list view showing table with Model ID, Display Name, Provider, Description, and Action buttons]

### Adding a New Custom Model

1.  Click the "Add New Custom Model" button.
2.  A modal dialog will appear with a form to enter the model's details.
    [Screenshot: "Add Custom Model" modal form]
3.  **Form Fields:**
    *   **Model ID (e.g., gpt-4-turbo, claude-3-opus-20240229):** A unique machine-readable identifier for the model. This ID cannot be changed after creation.
    *   **Display Name (e.g., GPT-4 Turbo, Claude Opus):** A human-readable name for easy identification.
    *   **Description:** Optional text to describe the model or its intended use.
    *   **Provider:** Select from a dropdown (OpenAI, Anthropic, Google, Custom).
    *   **Base URL (optional, for custom providers):** If using a "Custom" provider, enter the base URL of the LLM API endpoint. For standard providers, this is often pre-filled or not required.
    *   **API Key Environment Variable (e.g., OPENAI_API_KEY):** The name of the environment variable on your Plandex server that holds the API key for this model.
    *   **Max Tokens (Context Window):** The maximum number of tokens the model can process in a single request (input + output).
    *   **Max Output Tokens:** The maximum number of tokens the model is configured to generate in a response.
    *   **Default Max Conversation Tokens (for Planner):** A specific token limit setting often used by the "Planner" role within Model Packs.
    *   **Preferred Output Format:** Select the model's preferred output format (e.g., XML, Tool Call JSON). This influences how Plandex structures prompts for the model.
4.  Click "Create Model" to save the new custom model.

### Editing a Custom Model

1.  In the Custom Models list, find the model you wish to edit.
2.  Click the "Edit" button in its row.
3.  The "Edit Custom Model" modal will appear, pre-filled with the model's current details.
    [Screenshot: "Edit Custom Model" modal with fields pre-filled]
4.  Modify the necessary fields. Note that the "Model ID" cannot be changed after creation.
5.  Click "Save Changes".

### Deleting a Custom Model

1.  In the Custom Models list, find the model you wish to delete.
2.  Click the "Delete" button in its row.
3.  A confirmation dialog will appear.
    [Screenshot: Confirmation dialog for deleting a custom model]
4.  Click "OK" or "Confirm" to permanently delete the custom model configuration.
    *Note: Deleting a custom model may affect Model Packs that reference it. Ensure it's not in active use or update relevant Model Packs accordingly.*

## Model Packs

Model Packs group multiple configured Custom Models and assign them to specific roles within Plandex's operation (e.g., Planner, Coder, Builder). This allows you to use different models for different tasks based on their strengths.

### Listing Model Packs

Navigate to the "Model Packs" section. A table will display all existing model packs.

*   **Columns:**
    *   **Pack ID (User Defined):** The unique identifier you assigned to the pack.
    *   **Name:** A user-friendly name for the model pack.
    *   **Description:** A brief description of the pack's purpose or configuration.
    *   **Planner Model:** Displays the Model ID (and often Display Name) of the model assigned to the "Planner" role.
    *   **Actions:** Buttons to "Edit" or "Delete" the model pack.

[Screenshot: Model Packs list view with columns for Pack ID, Name, Description, Planner Model, and Actions]

### Adding a New Model Pack

1.  Click the "Add New Model Pack" button.
2.  A modal dialog will appear with a form.
    [Screenshot: "Add Model Pack" modal form with fields for Pack ID, Name, Description, and model selection dropdowns for each role]
3.  **Form Fields:**
    *   **Pack ID:** A unique machine-readable identifier for this pack. Cannot be changed after creation.
    *   **Pack Name:** A human-readable name.
    *   **Description:** Optional text.
    *   **Role Model Assignments:** For each defined role (Planner, Coder, Plan Summary, Builder, Whole File Builder, Namer, Commit Message, Execution Status, Architect/Context Loader), select a configured Custom Model from a dropdown list.
        *   The dropdowns will be populated with the "Display Name (Model ID)" of your available Custom Models.
        *   Some roles are core/required, while others might be optional (allowing you to leave them blank to use defaults or disable the role if applicable).
4.  Click "Create Pack" to save.

### Editing a Model Pack

1.  In the Model Packs list, click "Edit" for the desired pack.
2.  The "Edit Model Pack" modal will appear, pre-filled with current assignments.
    [Screenshot: "Edit Model Pack" modal with fields and dropdowns pre-filled]
3.  Adjust the Pack Name, Description, or reassign models to different roles using the dropdowns. The "Pack ID" cannot be changed.
4.  Click "Save Changes".

### Deleting a Model Pack

1.  In the Model Packs list, click "Delete" for the pack.
2.  Confirm the deletion in the dialog that appears.
    [Screenshot: Confirmation dialog for deleting a model pack]
    *Note: Deleting a model pack means plans configured to use it might need to be updated to a different pack.*

This LLM Management interface provides the flexibility to tailor Plandex's AI capabilities to your specific needs and available models.
