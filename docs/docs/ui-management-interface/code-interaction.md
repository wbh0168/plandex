---
sidebar_label: 'Code Interaction View'
sidebar_position: 6
---

# Code Interaction View

The Code Interaction View in the Plandex UI provides an advanced environment for leveraging Large Language Models (LLMs) to work with code. It features a sophisticated code editor (Monaco Editor), model selection, prompt input, and areas to display LLM-generated responses or explanations.

## Accessing the Code Interaction View

Navigate to "Code Interaction" from the main UI navigation bar.

[Screenshot: Main navigation bar highlighting "Code Interaction"]

## Interface Overview

The view is typically structured with the following components:

*   **Controls Pane:** Located at the top, this area contains:
    *   **Editor Language Selector:** A dropdown to choose the programming language for the code editor (e.g., JavaScript, Python, HTML). This affects syntax highlighting and potentially how the LLM interprets the code.
    *   **Model Selector:** A dropdown to select which configured Custom Model (from LLM Management) should be used for the interaction.
    *   **Prompt Input:** A textarea where you type your instructions or questions for the LLM regarding the code.
    *   **Action Buttons:** Buttons like "Generate Code" and "Explain Code" to trigger specific LLM interactions.
    [Screenshot: Controls Pane showing Language Selector, Model Selector, Prompt Input, and Action Buttons]

*   **Editor & Output Split Pane:**
    *   **Code Editor (Left/Top):** A full-featured Monaco Editor instance where you can write, paste, or view code. LLM-generated code may also appear here.
        [Screenshot: Monaco Editor with example code]
    *   **LLM Response Pane (Right/Bottom):** A display area (often a `<pre>` block) where the text-based output from the LLM (generated code, explanations, errors) is shown.
        [Screenshot: LLM Response Pane showing an example explanation or generated code snippet]

## Features

### Code Editor (Monaco)

*   **Rich Editing Features:** Provides syntax highlighting for numerous languages, code completion suggestions (depending on language services), minimap, and other standard features of the Monaco Editor.
*   **Language Selection:** You can change the active language of the editor using the "Editor Language" dropdown in the controls pane. This will update syntax highlighting.

### Model Selection

*   Before interacting with an LLM, you must select a configured **Custom Model** from the "Select Model" dropdown. This list is populated from the models you've set up in the "LLM Management" -> "Custom Models" section.
*   The chosen model will be used to process your prompt and code context.

### Prompt Input

*   Type your instructions for the LLM in the "Your Prompt" textarea. Be clear and specific.
    *   Example for "Generate Code": `"Create a Python function that takes a list of numbers and returns their sum."`
    *   Example for "Explain Code": `"Explain what this JavaScript function does and point out any potential issues."` (assuming the function is in the code editor).

### Action Buttons

*   **Generate Code:**
    1.  Optionally, provide existing code in the editor as context.
    2.  Enter a prompt describing the code you want the LLM to generate.
    3.  Select the desired model.
    4.  Click "Generate Code".
    5.  The LLM will process your request. The generated code will typically replace the content of the Monaco Editor, and the full response (which might include the code and other text) will appear in the "LLM Response" pane.
        [Screenshot: Code Editor View after "Generate Code" - editor updated, response pane shows generation details]

*   **Explain Code:**
    1.  Place the code you want explained into the Monaco Editor.
    2.  Enter a prompt asking for an explanation (e.g., "Explain this code," "What does this function do?").
    3.  Select the model.
    4.  Click "Explain Code".
    5.  The LLM's explanation will appear in the "LLM Response" pane. The code in the editor will not be changed by this action.
        [Screenshot: Code Editor View after "Explain Code" - response pane shows explanation]

*   **Other Actions (Potential):** Additional buttons for actions like "Refactor Code," "Find Bugs," etc., could be available depending on the Plandex version and configuration.

### LLM Response Display

*   **Loading State:** While the LLM is processing your request, a "Thinking..." or similar loading indicator will appear in the response pane.
*   **Content Display:** The LLM's textual response is displayed. This might be generated code, an explanation, or an error message from the LLM.
*   **Error Handling:** If the interaction with the LLM fails (either an API error or an issue within the Plandex backend), an error message will be shown in the response pane (e.g., "Error: LLM API failed," "Model not found").

## Important Notes

*   **Project Context:** While this view is for general code interaction, actions might eventually be tied to a specific project context (`test-project-123` is currently hardcoded).
*   **LLM Capabilities:** The quality and nature of the LLM's output depend heavily on the selected model's capabilities and the clarity of your prompt.
*   **No File Saving (Directly):** This view is primarily for interacting with code and LLMs. Code written or generated in the editor is not automatically saved to the project's file system unless explicitly stated or integrated with the File Management features. You would typically copy and paste code from the editor or response pane to a local file or another system if needed.
*   **Token Limits:** Be mindful of the token limits (context window size, output token limits) of the selected LLM. Very large code contexts or prompts might exceed these limits.

This Code Interaction View is a powerful tool for developers to leverage AI assistance directly within their Plandex workflow.
