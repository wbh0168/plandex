<script>
  import { onMount, onDestroy } from 'svelte';
  import * as monaco from 'monaco-editor';

  // Hardcoded for now
  const currentProjectId = 'test-project-123'; 

  let editorEl;
  let monacoEditor;
  let customModels = [];
  let selectedModelId = '';
  let userPrompt = '';
  let llmResponse = '';
  let isLoading = false;
  let currentCodeLanguage = 'javascript'; // Default language

  // Language options for Monaco Editor
  const languages = [
    { id: 'javascript', name: 'JavaScript' },
    { id: 'python', name: 'Python' },
    { id: 'html', name: 'HTML' },
    { id: 'css', name: 'CSS' },
    { id: 'json', name: 'JSON' },
    { id: 'markdown', name: 'Markdown' },
    { id: 'typescript', name: 'TypeScript' },
    { id: 'java', name: 'Java' },
    { id: 'csharp', name: 'C#' },
    { id: 'cpp', name: 'C++' },
    { id: 'go', name: 'Go' },
    { id: 'php', name: 'PHP' },
    { id: 'ruby', name: 'Ruby' },
    { id: 'rust', name: 'Rust' },
    { id: 'sql', name: 'SQL' },
    { id: 'xml', name: 'XML' },
    { id: 'yaml', name: 'YAML' },
  ];

  async function fetchCustomModels() {
    try {
      const response = await fetch('/api/ui/custom-models'); // Proxied by Vite dev server
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      customModels = await response.json();
      if (customModels && customModels.length > 0) {
        // Default to the first model's user-defined modelId (e.g., "gpt-4-turbo")
        selectedModelId = customModels[0].modelId; 
      }
    } catch (error) {
      console.error("Failed to fetch custom models:", error);
      llmResponse = `Error fetching models: ${error.message}`;
      customModels = [];
    }
  }

  onMount(() => {
    if (editorEl) {
      monacoEditor = monaco.editor.create(editorEl, {
        value: `// Welcome to the Code Interaction View!\n// Paste your code here or type new code.\n\nfunction greet() {\n  console.log("Hello, world!");\n}\n`,
        language: currentCodeLanguage,
        theme: 'vs-dark', // Standard themes: 'vs', 'vs-dark', 'hc-black'
        automaticLayout: true, // Adjusts editor layout on container resize
        minimap: { enabled: true },
      });
    }
    fetchCustomModels();

    return () => {
      if (monacoEditor) {
        monacoEditor.dispose();
      }
    };
  });

  async function handleCodeInteraction(action) {
    if (!selectedModelId) {
      llmResponse = 'Please select a model.';
      return;
    }
    if (!userPrompt) {
      llmResponse = 'Please enter a prompt.';
      return;
    }

    isLoading = true;
    llmResponse = 'Processing...';
    const codeContext = monacoEditor ? monacoEditor.getValue() : '';

    try {
      const response = await fetch(`/api/ui/projects/${currentProjectId}/code/interact`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model_id: selectedModelId,
          prompt: userPrompt,
          code_context: codeContext,
          action: action,
        }),
      });

      const result = await response.json();
      isLoading = false;

      if (!response.ok) {
        throw new Error(result.error || result.message || `API Error: ${response.status}`);
      }
      
      llmResponse = result.content;

      if (action === 'generate_code' && result.content) {
        // Example: Replace current editor content with generated code.
        // More sophisticated logic could be: insert at cursor, replace selection, etc.
        const currentModel = monacoEditor.getModel();
        if (currentModel) {
            // A common way to replace all text is to execute an edit operation
            const fullRange = currentModel.getFullModelRange();
            monacoEditor.executeEdits("llm-interaction", [{
                range: fullRange,
                text: result.content
            }]);
             llmResponse = "Code generated and placed in editor. Original response also below:\n\n" + result.content;
        } else {
             llmResponse = "Code generated (see below), but failed to update editor content directly.\n\n" + result.content;
        }
      }

    } catch (error) {
      isLoading = false;
      console.error('Code interaction error:', error);
      llmResponse = `Error: ${error.message}`;
    }
  }
  
  // Update Monaco editor language when selection changes
  $: if (monacoEditor && monaco.editor.getModels().length > 0) {
      monaco.editor.setModelLanguage(monacoEditor.getModel(), currentCodeLanguage);
  }

</script>

<style>
  .code-interaction-view {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 100px); /* Adjust based on nav/header height */
    padding: 15px;
    gap: 15px;
    background-color: #282c34; /* Dark background for the view */
    color: #abb2bf; /* Light text color */
  }

  .controls-pane {
    display: flex;
    flex-direction: column; /* Stack controls vertically */
    gap: 10px;
    padding: 10px;
    background-color: #323842; /* Slightly lighter than main background */
    border-radius: 6px;
    box-shadow: 0 2px 5px rgba(0,0,0,0.2);
  }
  
  .control-row {
    display: flex;
    gap: 15px;
    align-items: center; /* Vertically align items in a row */
  }

  .control-row label {
    margin-right: 5px;
    font-weight: bold;
    color: #9da5b4;
  }

  .control-row select,
  .control-row textarea,
  .control-row button {
    padding: 8px 12px;
    border-radius: 4px;
    border: 1px solid #4f5865;
    background-color: #21252b; /* Darker input fields */
    color: #abb2bf;
    font-family: inherit;
  }
  
  .control-row select {
      min-width: 200px; /* Give select some base width */
  }

  .control-row textarea {
    flex-grow: 1; /* Prompt textarea takes available space */
    min-height: 60px; /* Start with a decent height */
    resize: vertical;
  }
  
  .control-row button {
    background-color: #61afef; /* A distinct button color */
    color: #282c34; /* Dark text on button */
    cursor: pointer;
    transition: background-color 0.2s;
  }
  .control-row button:hover {
    background-color: #5295cf;
  }
  .control-row button:disabled {
    background-color: #4f5865;
    cursor: not-allowed;
  }
  
  .editor-output-split {
    display: flex;
    flex-direction: row; /* Editor and output side-by-side */
    flex-grow: 1; /* Take remaining vertical space */
    gap: 15px;
    min-height: 0; /* Important for flex children to shrink correctly */
  }

  .editor-container {
    flex: 1; /* Editor takes half the space */
    border: 1px solid #4f5865;
    border-radius: 6px;
    overflow: hidden; /* Monaco handles its own scrolling */
    min-height: 300px; /* Ensure editor has some minimum height */
  }

  .output-pane {
    flex: 1; /* Output takes half the space */
    display: flex; /* To allow pre to grow */
    flex-direction: column;
    border: 1px solid #4f5865;
    border-radius: 6px;
    background-color: #21252b; /* Dark background for output */
    padding: 10px;
    min-height: 300px;
  }
  
  .output-pane h3 {
      margin-top: 0;
      color: #9da5b4;
      border-bottom: 1px solid #4f5865;
      padding-bottom: 8px;
  }

  .output-pane pre {
    white-space: pre-wrap; /* Wrap long lines */
    word-wrap: break-word; /* Break words if necessary */
    flex-grow: 1; /* Make pre take available space */
    overflow-y: auto; /* Scroll if content overflows */
    margin: 0; /* Remove default pre margin */
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
    font-size: 14px;
  }
</style>

<div class="code-interaction-view">
  <div class="controls-pane">
    <div class="control-row">
      <label for="language-select">Editor Language:</label>
      <select id="language-select" bind:value={currentCodeLanguage}>
        {#each languages as lang}
          <option value={lang.id}>{lang.name}</option>
        {/each}
      </select>

      <label for="model-select">Select Model:</label>
      <select id="model-select" bind:value={selectedModelId} disabled={customModels.length === 0}>
        {#if customModels.length === 0}
          <option value="">Loading models...</option>
        {/if}
        {#each customModels as model}
          <option value={model.modelId}>{model.modelName} ({model.modelId})</option>
        {/each}
      </select>
    </div>
    <div class="control-row">
      <label for="prompt-input">Your Prompt:</label>
      <textarea id="prompt-input" rows="3" bind:value={userPrompt} placeholder="e.g., 'Convert this function to be asynchronous' or 'Add error handling'"></textarea>
    </div>
    <div class="control-row">
      <button on:click={() => handleCodeInteraction('generate_code')} disabled={isLoading}>Generate Code</button>
      <button on:click={() => handleCodeInteraction('explain_code')} disabled={isLoading}>Explain Code</button>
      <!-- Add more action buttons here e.g., refactor_code -->
    </div>
  </div>

  <div class="editor-output-split">
    <div class="editor-container" bind:this={editorEl}>
      <!-- Monaco Editor will be mounted here -->
    </div>

    <div class="output-pane">
      <h3>LLM Response:</h3>
      {#if isLoading}
        <p>Thinking...</p>
      {:else}
        <pre>{llmResponse}</pre>
      {/if}
    </div>
  </div>
</div>
