// Assuming CodeEditorView.svelte is in src/ or src/components/
// Path: app/ui/frontend/my-svelte-app/src/CodeEditorView.test.js

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { render, fireEvent, screen, waitFor, act } from '@testing-library/svelte';
import CodeEditorView from '../CodeEditorView.svelte'; // Adjust path
import api from './api'; // To mock API calls for custom models and code interaction

// Mock the api module
vi.mock('./api', () => ({
  default: {
    get: vi.fn(),  // For fetching custom models
    post: vi.fn(), // For code interaction
  }
}));

// Mock Monaco Editor
// Similar to xterm.js, full Monaco rendering in jsdom is complex.
// We mock the parts of the Monaco API that our component interacts with.
let mockEditorInstance;
const mockMonacoEditor = {
  create: vi.fn((_element, options) => {
    mockEditorInstance = {
      getValue: vi.fn(() => options.value || ''),
      setValue: vi.fn((newValue) => { // Allow tests to simulate editor content change if needed
        mockEditorInstance._currentValue = newValue;
      }),
      _currentValue: options.value || '', // Internal mock tracking of value
      executeEdits: vi.fn((_source, edits) => {
        // Simulate applying edits to our internal mock value
        if (edits && edits.length > 0 && edits[0].text) {
          mockEditorInstance._currentValue = edits[0].text;
        }
      }),
      getModel: vi.fn(() => ({
        getFullModelRange: vi.fn(() => ({ startLineNumber: 1, startColumn: 1, endLineNumber: 10, endColumn: 10})), // Dummy range
      })),
      dispose: vi.fn(),
      // Add other methods if your component uses them
    };
    return mockEditorInstance;
  }),
  editor: { // Nested structure if component uses monaco.editor.create etc.
    setModelLanguage: vi.fn(),
    // Mock other static editor methods if needed
    // create: (...) // if used as monaco.editor.create
  }
};
// If the component imports * as monaco from 'monaco-editor'
vi.mock('monaco-editor', () => ({
  editor: mockMonacoEditor, // Assuming component uses monaco.editor.create
  // Add other exports from 'monaco-editor' if component uses them (KeyMod, KeyCode, etc.)
}));


describe('CodeEditorView.svelte', () => {
  const projectId = 'test-project-123'; // Matches hardcoded projectId
  const mockCustomModels = [
    { modelId: 'gpt-4-test', modelName: 'GPT-4 Test', description: 'Test model A' },
    { modelId: 'claude-opus-test', modelName: 'Claude Opus Test', description: 'Test model B' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    // Reset editor instance value if needed between tests, though create should do this
    if (mockEditorInstance) {
        mockEditorInstance._currentValue = `// Welcome to the Code Interaction View!\n// Paste your code here or type new code.\n\nfunction greet() {\n  console.log("Hello, world!");\n}\n`;
    }
    // Mock api.get to return custom models by default for most tests
    api.get.mockResolvedValue(mockCustomModels); 
  });

  it('renders and initializes Monaco editor, fetches custom models', async () => {
    render(CodeEditorView);

    // Check for Monaco editor initialization (via mock)
    expect(mockMonacoEditor.create).toHaveBeenCalledOnce();
    expect(screen.getByText('Editor Language:')).toBeInTheDocument(); // Part of UI

    // Check for model fetching and selection
    expect(screen.getByLabelText('Select Model:')).toBeInTheDocument();
    await waitFor(() => {
      expect(api.get).toHaveBeenCalledWith('/custom-models');
      // First model should be selected by default if models are fetched
      expect(screen.getByRole('option', { name: /GPT-4 Test/i }).selected).toBe(true);
    });
    
    expect(screen.getByLabelText('Your Prompt:')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Generate Code' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Explain Code' })).toBeInTheDocument();
    expect(screen.getByText('LLM Response:')).toBeInTheDocument();
  });
  
  it('handles model selection change', async () => {
    render(CodeEditorView);
    await waitFor(() => expect(api.get).toHaveBeenCalled()); // Wait for models to load

    const modelSelect = screen.getByLabelText('Select Model:');
    // mockCustomModels[1].modelId is 'claude-opus-test'
    await fireEvent.change(modelSelect, { target: { value: mockCustomModels[1].modelId } });
    
    // Check if the selected value in the component's state (if exposed) or via selected option
    expect(screen.getByRole('option', { name: /Claude Opus Test/i }).selected).toBe(true);
  });

  it('handles language selection change for Monaco editor', async () => {
    render(CodeEditorView);
    await waitFor(() => expect(mockMonacoEditor.create).toHaveBeenCalled());

    const languageSelect = screen.getByLabelText('Editor Language:');
    await fireEvent.change(languageSelect, { target: { value: 'python' } });
    
    // Check if monaco.editor.setModelLanguage was called with the new language
    // This requires getModel() to return a mock model that setModelLanguage can be called on
    expect(mockMonacoEditor.editor.setModelLanguage).toHaveBeenCalledWith(expect.anything(), 'python');
  });


  it('calls API for "Generate Code" action and updates editor on success', async () => {
    const generatedCode = 'const newGeneratedCode = () => "hello";';
    api.post.mockResolvedValue({ content: generatedCode });
    
    render(CodeEditorView);
    await waitFor(() => expect(api.get).toHaveBeenCalled()); // Models loaded

    // Set prompt and model
    const promptInput = screen.getByLabelText('Your Prompt:');
    await fireEvent.input(promptInput, { target: { value: 'generate a function' } });
    // Model is already selected by default (gpt-4-test)

    // Mock editor getValue before action
    const initialEditorContent = '// initial content';
    mockEditorInstance.getValue.mockReturnValue(initialEditorContent);

    await fireEvent.click(screen.getByRole('button', { name: 'Generate Code' }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(
        `/projects/${projectId}/code/interact`,
        {
          model_id: mockCustomModels[0].modelId, // default selected
          prompt: 'generate a function',
          code_context: initialEditorContent,
          action: 'generate_code',
        }
      );
    });
    
    // Check if editor content was updated by executeEdits
    expect(mockEditorInstance.executeEdits).toHaveBeenCalledWith(
        "llm-interaction", 
        [{ range: expect.anything(), text: generatedCode }]
    );
    
    // Check if LLM response display is updated (component prepends a message)
    await screen.findByText(/Code generated and placed in editor./);
    expect(screen.getByText(new RegExp(generatedCode.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))).toBeInTheDocument();
  });

  it('calls API for "Explain Code" action and displays response', async () => {
    const explanation = 'This code does amazing things.';
    api.post.mockResolvedValue({ content: explanation });

    render(CodeEditorView);
    await waitFor(() => expect(api.get).toHaveBeenCalled());

    const promptInput = screen.getByLabelText('Your Prompt:');
    await fireEvent.input(promptInput, { target: { value: 'explain this code' } });
    
    const editorContentForExplain = 'function complex() { /* ... */ }';
    mockEditorInstance.getValue.mockReturnValue(editorContentForExplain);

    await fireEvent.click(screen.getByRole('button', { name: 'Explain Code' }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(
        `/projects/${projectId}/code/interact`,
        {
          model_id: mockCustomModels[0].modelId,
          prompt: 'explain this code',
          code_context: editorContentForExplain,
          action: 'explain_code',
        }
      );
    });
    
    // Editor content should NOT be updated for 'explain_code'
    expect(mockEditorInstance.executeEdits).not.toHaveBeenCalled();
    // Check LLM response display
    await screen.findByText(explanation);
  });
  
  it('shows loading state during API call', async () => {
    api.post.mockReturnValue(new Promise(resolve => setTimeout(() => resolve({ content: 'done' }), 100))); // Delayed response
    render(CodeEditorView);
    await waitFor(() => expect(api.get).toHaveBeenCalled());

    await fireEvent.input(screen.getByLabelText('Your Prompt:'), { target: { value: 'long task' } });
    fireEvent.click(screen.getByRole('button', { name: 'Generate Code' })); // No await here

    // Check for loading indicator immediately after click
    // (or as soon as Svelte updates the DOM, use await tick if needed or findBy)
    expect(await screen.findByText('Thinking...')).toBeInTheDocument();
    
    // Wait for API call to resolve and loading to disappear
    await waitFor(() => {
      expect(screen.queryByText('Thinking...')).not.toBeInTheDocument();
      expect(screen.getByText(/done/)).toBeInTheDocument(); // Response content
    });
  });

  it('displays error message if API call fails', async () => {
    const errorMsg = 'LLM API failed';
    api.post.mockRejectedValue({ message: errorMsg });
    render(CodeEditorView);
    await waitFor(() => expect(api.get).toHaveBeenCalled());

    await fireEvent.input(screen.getByLabelText('Your Prompt:'), { target: { value: 'trigger error' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Generate Code' }));

    await waitFor(() => {
      expect(screen.getByText(`Error: ${errorMsg}`)).toBeInTheDocument();
    });
  });
  
  it('shows error if no model is selected', async () => {
    api.get.mockResolvedValue([]); // No models loaded
    render(CodeEditorView);
    await waitFor(() => expect(api.get).toHaveBeenCalled()); // Wait for initial model fetch attempt
    
    // Clear default selection if any, or ensure no models are there
    // For this test, api.get returning [] means select will be empty.

    await fireEvent.input(screen.getByLabelText('Your Prompt:'), { target: { value: 'test' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Generate Code' }));

    await waitFor(() => {
        expect(screen.getByText('Please select a model.')).toBeInTheDocument();
    });
    expect(api.post).not.toHaveBeenCalled();
  });

  it('shows error if no prompt is entered', async () => {
    render(CodeEditorView);
    await waitFor(() => expect(api.get).toHaveBeenCalled()); // Models loaded
    
    // No prompt entered
    await fireEvent.click(screen.getByRole('button', { name: 'Generate Code' }));

    await waitFor(() => {
        expect(screen.getByText('Please enter a prompt.')).toBeInTheDocument();
    });
    expect(api.post).not.toHaveBeenCalled();
  });

});
