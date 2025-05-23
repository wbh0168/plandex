// Path: app/ui/frontend/my-svelte-app/src/RealtimePreview.test.js

import { describe, it, expect, beforeEach, vi, afterEach, vitest } from 'vitest';
import { render, fireEvent, screen, waitFor, act } from '@testing-library/svelte';
import RealtimePreview from '../RealtimePreview.svelte'; // Adjust path

// Mock simple-code-editor and highlight.js as they are complex and not core to srcdoc logic
// We'll simulate them with basic textareas for these tests.
vi.mock('simple-code-editor/CodeEditor.svelte', () => {
  // Mock the CodeEditor component to behave like a textarea for testing purposes
  // It needs to accept 'value' and dispatch an 'input' or 'change' event with { detail: { value: ... } }
  // or allow direct binding. For simplicity, we'll assume direct value binding works if possible,
  // or simulate input events that update a bound value.
  // Svelte Testing Library's `fireEvent.input` works well with native inputs.
  // For custom components, they need to forward events or allow programmatic changes.
  //
  // A simple mock: Render a textarea, and ensure `bind:value` works as expected.
  // The component `RealtimePreview.svelte` uses `bind:value={htmlCode}` etc.
  // So, the mock needs to support this binding.
  //
  // Let's create a mock Svelte component for CodeEditor.
  const CodeEditorMock = ({ value, id }) => {
    // This is a conceptual mock. In reality, you'd import a Svelte component mock.
    // For Vitest, you can create a __mocks__ folder or use vi.mock to provide a fake component.
    // For now, we'll assume RealtimePreview.svelte can handle if CodeEditor is not present,
    // OR we test the logic that would be passed to CodeEditor.
    //
    // The `RealtimePreview.svelte` has fallback textareas commented out.
    // If we can't easily mock `CodeEditor` to act like a textarea for `bind:value`,
    // we might need to test a version of RealtimePreview that uses plain textareas.
    //
    // However, `simple-code-editor` might just work in jsdom if it doesn't do heavy canvas/DOM stuff.
    // Let's assume it does, or test around it by focusing on the state changes.
    // For this test, we will assume the `bind:value` works and we can fire input events
    // on elements that will update the bound variables (htmlCode, cssCode, jsCode).
    // The component uses `id="html-editor"`, etc. We can try to find these.
    //
    // To truly mock it as a textarea for testing:
    // 1. Create a `__mocks__/simple-code-editor/CodeEditor.svelte` file.
    // 2. Inside, have `<script>export let value;</script><textarea bind:value />`
    // Vitest should pick this up if configured.
    //
    // For this test file, we'll assume the component renders *something* for the editors
    // that we can target and simulate input on, and that this input updates the bound variables.
    // We will target by the label a CodeEditor has.
    // The actual component `simple-code-editor` renders a `div` with `contenteditable="true"`.
    // We can try to interact with that.
    return {
        // This mock structure is more for jest. For vitest with Svelte,
        // it's often easier to test the component that *uses* the child,
        // and if child is complex, mock its outputs or side effects.
        // We will try to find the contenteditable divs.
    };
});

vi.mock('highlight.js/lib/core', () => ({
  default: {
    registerLanguage: vi.fn(),
    getLanguage: vi.fn(lang => ({ name: lang })), // Return a dummy language object
    highlight: vi.fn((code, options) => ({ value: code })), // Return raw code
  }
}));
vi.mock('highlight.js/lib/languages/xml', () => ({ default: {} }));
vi.mock('highlight.js/lib/languages/css', () => ({ default: {} }));
vi.mock('highlight.js/lib/languages/javascript', () => ({ default: {} }));


describe('RealtimePreview.svelte', () => {
  beforeEach(() => {
    vi.useFakeTimers(); // Use fake timers for debounce testing
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  function findEditorByLabelText(labelText) {
    // simple-code-editor renders a div with contenteditable="true"
    // The label is a sibling to the editor's wrapper div.
    // This is a bit fragile. data-testid attributes would be better.
    const label = screen.getByText(labelText);
    // Assuming the editor div is the next sibling or within a specific structure
    // This will depend on the actual DOM structure rendered by CodeEditor.svelte
    // For `simple-code-editor`, it's usually a div with class `code-editor` or `inputarea`
    // Let's assume our mock or the real component makes it findable.
    // The `CodeEditor` component itself has an id passed to it.
    // `simple-code-editor` creates a `div.codejar-wrap > div.editor-container > textarea.inputarea`
    // and `div.codejar-wrap > div.editor-container > div.codejar.editor`
    // We need to target the `textarea.inputarea` for input events if it's used for binding,
    // or the contenteditable div for direct manipulation if that's how it works.
    //
    // Given the mock of simple-code-editor is tricky, we'll assume we can find an input field
    // associated with the label. If `simple-code-editor` was replaced by a textarea, this would be easier.
    // Let's assume the component structure allows finding the input area via its ID.
    // The component passes id="html-editor", etc. to CodeEditor.
    // `simple-code-editor` might not directly use that ID on an input element.
    //
    // For this test, we'll try to get by role if possible, or a test-id if we could add it.
    // Since `simple-code-editor` uses contenteditable divs, we might need to query for those.
    // The component structure is: <div class="editor-wrapper"><label>HTML</label><CodeEditor id="html-editor" .../></div>
    // Let's get the parent, then find the contenteditable div inside the CodeEditor's output.
    // This is still very dependent on simple-code-editor's internal structure.
    //
    // A simpler approach for testing RealtimePreview's logic:
    // Assume the `bind:value` on `<CodeEditor>` works.
    // We can then programmatically change the props passed to RealtimePreview if we could,
    // or trigger an event on a mock that updates the value.
    //
    // Let's try to find the contenteditable div that `simple-code-editor` likely creates.
    // It usually has a class like `editor`.
    // The label is for "html-editor", "css-editor", "js-editor".
    // The CodeEditor component is given these IDs.
    // `simple-code-editor` creates a `textarea` with class `inputarea` and a `div` with class `editor`.
    // We should be able to find the textarea.
    return screen.getByRole('textbox', { name: labelText }); // This should find the textarea if it's labeled by the <label>
  }


  it('renders editors and an iframe', async () => {
    render(RealtimePreview);

    expect(screen.getByText('HTML')).toBeInTheDocument();
    expect(screen.getByText('CSS')).toBeInTheDocument();
    expect(screen.getByText('JavaScript')).toBeInTheDocument();
    
    // Find by role textbox with accessible name from label
    // This relies on simple-code-editor making its textarea accessible via the label.
    // If not, we need a more specific selector based on its DOM structure.
    // For now, assuming the labels correctly associate with the input areas.
    // `simple-code-editor` uses a textarea internally.
    expect(screen.getByRole('textbox', { name: 'HTML' })).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: 'CSS' })).toBeInTheDocument();
    expect(screen.getByRole('textbox', { name: 'JavaScript' })).toBeInTheDocument();

    expect(screen.getByTitle('Realtime Preview')).toBeInTheDocument(); // iframe
  });

  it('updates iframe srcdoc when HTML code changes', async () => {
    render(RealtimePreview);
    const iframe = screen.getByTitle('Realtime Preview');
    const htmlEditor = screen.getByRole('textbox', { name: 'HTML' });

    const newHtml = '<p>New HTML content</p>';
    // Simulate user typing into the HTML editor
    // The `simple-code-editor` component uses a textarea internally for input.
    await fireEvent.input(htmlEditor, { target: { value: newHtml } });
    
    act(() => vi.runAllTimers()); // Trigger debounce

    await waitFor(() => {
      expect(iframe.srcdoc).toContain(newHtml);
      // Default CSS and JS should also be there
      expect(iframe.srcdoc).toContain('body {'); // Default CSS
      expect(iframe.srcdoc).toContain('console.log("JavaScript loaded!")'); // Default JS (after sanitization)
    });
  });

  it('updates iframe srcdoc when CSS code changes', async () => {
    render(RealtimePreview);
    const iframe = screen.getByTitle('Realtime Preview');
    const cssEditor = screen.getByRole('textbox', { name: 'CSS' });
    
    const newCss = 'h1 { color: blue; }';
    await fireEvent.input(cssEditor, { target: { value: newCss } });
    
    act(() => vi.runAllTimers());

    await waitFor(() => {
      expect(iframe.srcdoc).toContain(newCss);
      expect(iframe.srcdoc).toContain('<h1>Hello, World!</h1>'); // Default HTML
    });
  });

  it('updates iframe srcdoc when JavaScript code changes', async () => {
    render(RealtimePreview);
    const iframe = screen.getByTitle('Realtime Preview');
    const jsEditor = screen.getByRole('textbox', { name: 'JavaScript' });

    const newJs = 'document.body.style.backgroundColor = "yellow";';
    await fireEvent.input(jsEditor, { target: { value: newJs } });

    act(() => vi.runAllTimers());

    await waitFor(() => {
      expect(iframe.srcdoc).toContain(newJs);
      // Check if the script tag is correctly formed
      expect(iframe.srcdoc).toMatch(/<script>[\s\S]*try\s*{[\s\S]*document\.body\.style\.backgroundColor = "yellow";[\s\S]*} catch \(e\) {[\s\S]*<\/script>/);
    });
  });

  it('debounces iframe updates', async () => {
    render(RealtimePreview);
    const iframe = screen.getByTitle('Realtime Preview');
    const htmlEditor = screen.getByRole('textbox', { name: 'HTML' });

    // Initial content check (after onMount sanitization)
    act(() => vi.runAllTimers()); // Ensure onMount update happens
    await waitFor(() => {
        expect(iframe.srcdoc).toContain('<h1>Hello, World!</h1>');
        expect(iframe.srcdoc).toContain('color: orange'); // _orange_ -> orange
        expect(iframe.srcdoc).toContain('document.querySelector("h1")'); // querySelector_ -> querySelector
    });


    // Change 1
    await fireEvent.input(htmlEditor, { target: { value: '<p>Change 1</p>' } });
    // srcdoc should not update immediately due to debounce
    expect(iframe.srcdoc).not.toContain('<p>Change 1</p>'); 
    
    act(() => vi.advanceTimersByTime(100)); // Advance time by less than debounceTime (250ms)
    expect(iframe.srcdoc).not.toContain('<p>Change 1</p>');

    // Change 2 (before debounce timeout for Change 1)
    await fireEvent.input(htmlEditor, { target: { value: '<p>Change 2</p>' } });
    expect(iframe.srcdoc).not.toContain('<p>Change 1</p>');
    expect(iframe.srcdoc).not.toContain('<p>Change 2</p>');

    act(() => vi.runAllTimers()); // Run all pending timers

    await waitFor(() => {
      expect(iframe.srcdoc).not.toContain('<p>Change 1</p>'); // Should be overwritten by Change 2
      expect(iframe.srcdoc).toContain('<p>Change 2</p>');
    });
  });
  
  it('includes error handling for user script in iframe', async () => {
    render(RealtimePreview);
    const iframe = screen.getByTitle('Realtime Preview');
    const jsEditor = screen.getByRole('textbox', { name: 'JavaScript' });

    const errorJs = 'throw new Error("Test script error");';
    await fireEvent.input(jsEditor, { target: { value: errorJs } });
    act(() => vi.runAllTimers());

    await waitFor(() => {
      expect(iframe.srcdoc).toContain('try {');
      expect(iframe.srcdoc).toContain('throw new Error("Test script error");');
      expect(iframe.srcdoc).toContain('catch (e) {');
      expect(iframe.srcdoc).toContain('console.error("Error in user script:", e);');
    });
  });
  
  it('sanitizes initial CSS and JS code onMount', async () => {
    // Default values in component: cssCode has `_orange_`, jsCode has `querySelector_`
    render(RealtimePreview);
    const iframe = screen.getByTitle('Realtime Preview');

    act(() => vi.runAllTimers()); // Ensure onMount and subsequent updatePreview run

    await waitFor(() => {
      // Check that _orange_ was replaced with orange
      expect(iframe.srcdoc).toContain('color: orange;');
      expect(iframe.srcdoc).not.toContain('color:_orange_;');
      
      // Check that querySelector_ was replaced with querySelector
      expect(iframe.srcdoc).toContain('document.querySelector("h1")');
      expect(iframe.srcdoc).not.toContain('document.querySelector_("h1")');
    });
  });

});
