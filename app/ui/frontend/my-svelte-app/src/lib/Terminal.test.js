// Assuming Terminal.svelte is in src/ or src/components/
// Adjust path: app/ui/frontend/my-svelte-app/src/Terminal.test.js

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/svelte';
import TerminalComponent from '../Terminal.svelte'; // Adjust path to your Terminal.svelte
import api from './api';

// Mock the api module
vi.mock('./api', () => ({
  default: {
    post: vi.fn(), // For command execution
  }
}));

// Mock xterm.js and xterm-addon-fit
// This is crucial because xterm.js relies on a real terminal environment
// and DOM APIs that might not be fully available or performant in jsdom.
let mockTerminalInstance;
const mockFitAddonInstance = { fit: vi.fn() };

vi.mock('xterm', () => {
  // Mock the Terminal class constructor and its methods
  const Terminal = vi.fn(() => {
    mockTerminalInstance = {
      open: vi.fn(),
      loadAddon: vi.fn(),
      onData: vi.fn((callback) => {
        // Store the callback to simulate data input later
        mockTerminalInstance._onDataCallback = callback;
      }),
      write: vi.fn(),
      writeln: vi.fn(),
      focus: vi.fn(),
      dispose: vi.fn(),
      // Add any other methods your component calls
      _onDataCallback: null, // To store the callback passed to onData
      clear: vi.fn(), // Common method
      scrollToBottom: vi.fn(),
      // Simulate key events if your onData handler expects specific event structures not just strings
      // For this component, onData receives string data, so direct callback call is enough.
    };
    return mockTerminalInstance;
  });
  return { Terminal };
});

vi.mock('xterm-addon-fit', () => {
  const FitAddon = vi.fn(() => mockFitAddonInstance);
  return { FitAddon };
});

// Mock ResizeObserver (used by Terminal.svelte for fitAddon)
global.ResizeObserver = vi.fn(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));


describe('Terminal.svelte', () => {
  const projectId = 'test-project-123'; // Matches hardcoded projectId

  beforeEach(() => {
    vi.clearAllMocks();
    // Reset terminal instance for each test if Terminal constructor is called per render
    // If Terminal is singleton-like in mock, reset its methods' call counts.
    // The current mock creates a new mockTerminalInstance each time Terminal is newed up.
  });
  
  // Helper function to simulate user typing into the mocked terminal
  function simulateTerminalInput(data) {
    if (mockTerminalInstance && mockTerminalInstance._onDataCallback) {
      mockTerminalInstance._onDataCallback(data);
    } else {
      throw new Error("Terminal's onData callback not registered or terminal not initialized.");
    }
  }

  it('renders and initializes xterm.js on mount', () => {
    render(TerminalComponent);
    
    expect(mockTerminalInstance).not.toBeNull();
    expect(mockTerminalInstance.open).toHaveBeenCalledOnce();
    expect(mockTerminalInstance.loadAddon).toHaveBeenCalledWith(mockFitAddonInstance);
    expect(mockFitAddonInstance.fit).toHaveBeenCalledOnce();
    expect(mockTerminalInstance.focus).toHaveBeenCalledOnce();
    // Check initial prompt
    expect(mockTerminalInstance.write).toHaveBeenCalledWith(expect.stringContaining(`${projectId} $ `));
  });

  it('handles basic command input and execution', async () => {
    const command = 'ls -la';
    const mockApiResponse = { stdout: 'file1\nfile2', stderr: '', exitCode: 0 };
    api.post.mockResolvedValue(mockApiResponse);

    render(TerminalComponent);

    // Simulate typing "ls -la"
    command.split('').forEach(char => simulateTerminalInput(char));
    // Simulate pressing Enter
    simulateTerminalInput('\r'); 

    // Check if api.post was called correctly
    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/projects/${projectId}/commands`, { command });
    });

    // Check if output was written to terminal
    await waitFor(() => {
      // For prompt, writeln for newline after command, then stdout lines, then exit code, then new prompt
      expect(mockTerminalInstance.writeln).toHaveBeenCalledWith(expect.stringContaining('file1'));
      expect(mockTerminalInstance.writeln).toHaveBeenCalledWith(expect.stringContaining('file2'));
      expect(mockTerminalInstance.writeln).toHaveBeenCalledWith(expect.stringContaining('Exit Code: 0'));
    });
    // Prompt is written again after command execution
    expect(mockTerminalInstance.write).toHaveBeenCalledTimes(2); // Initial prompt + prompt after command
    expect(mockTerminalInstance.write).toHaveBeenLastCalledWith(expect.stringContaining(`${projectId} $ `));
  });

  it('handles command execution with stderr output', async () => {
    const command = 'cat non_existent_file';
    const mockApiResponse = { stdout: '', stderr: 'cat: No such file or directory', exitCode: 1 };
    api.post.mockResolvedValue(mockApiResponse);

    render(TerminalComponent);
    command.split('').forEach(char => simulateTerminalInput(char));
    simulateTerminalInput('\r');

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/projects/${projectId}/commands`, { command });
    });
    
    await waitFor(() => {
      // Stderr is written with specific ANSI color codes by the component
      expect(mockTerminalInstance.writeln).toHaveBeenCalledWith(expect.stringContaining('\x1b[31mcat: No such file or directory\x1b[0m'));
      expect(mockTerminalInstance.writeln).toHaveBeenCalledWith(expect.stringContaining('Exit Code: 1'));
    });
  });
  
  it('handles API error during command execution', async () => {
    const command = 'error_command';
    const apiErrorMsg = 'Network failure';
    api.post.mockRejectedValue({ message: apiErrorMsg });

    render(TerminalComponent);
    command.split('').forEach(char => simulateTerminalInput(char));
    simulateTerminalInput('\r');

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/projects/${projectId}/commands`, { command });
    });

    await waitFor(() => {
      expect(mockTerminalInstance.writeln).toHaveBeenCalledWith(expect.stringContaining(`Client-side error sending command: ${apiErrorMsg}`));
    });
  });
  
  it('handles non-OK HTTP response from API (e.g. 400, 500)', async () => {
    const command = 'server_error_command';
    // api.js wraps errors, so the component receives the error object from api.js
    const serverErrorPayload = { message: "Command execution failed on server", details: "Syntax error" };
    // This error structure matches what api.js throws for non-ok JSON responses
    api.post.mockRejectedValue({ status: 500, ...serverErrorPayload });


    render(TerminalComponent);
    command.split('').forEach(char => simulateTerminalInput(char));
    simulateTerminalInput('\r');

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(`/projects/${projectId}/commands`, { command });
    });
    
    // The component formats the error message. Check for the core message.
    await waitFor(() => {
      // The component's error handler for API calls might prepend "Error: " or similar.
      // It receives the error object thrown by api.js.
      // Terminal.svelte: term.writeln(`\x1b[1;31mError: ${result.message || `HTTP ${response.status}`}\x1b[0m`);
      // If API throws {status, message}, then it should be result.message
      expect(mockTerminalInstance.writeln).toHaveBeenCalledWith(expect.stringContaining(`\x1b[1;31mError: ${serverErrorPayload.message}\x1b[0m`));
    });
  });


  it('handles backspace correctly in command input', () => {
    render(TerminalComponent);
    simulateTerminalInput('l');
    simulateTerminalInput('s');
    simulateTerminalInput('a'); // currentCommand = "lsa"
    
    // Check that write was called for each char
    expect(mockTerminalInstance.write).toHaveBeenCalledWith('l');
    expect(mockTerminalInstance.write).toHaveBeenCalledWith('s');
    expect(mockTerminalInstance.write).toHaveBeenCalledWith('a');

    simulateTerminalInput('\b'); // Backspace (ASCII 8) or '\x7f' (DEL)

    // Expect terminal to erase the character
    expect(mockTerminalInstance.write).toHaveBeenCalledWith('\b \b');
    
    // currentCommand in component should be "ls"
    // Submit command to verify
    api.post.mockResolvedValue({ stdout: 'ok', stderr: '', exitCode: 0 });
    simulateTerminalInput('\r'); 
    
    expect(api.post).toHaveBeenCalledWith(`/projects/${projectId}/commands`, { command: 'ls' });
  });

  it('handles empty command submission (just shows new prompt)', () => {
    render(TerminalComponent);
    const initialWriteCount = mockTerminalInstance.write.mock.calls.length;
    
    simulateTerminalInput('\r'); // Enter

    // Should write a newline, then a new prompt.
    expect(mockTerminalInstance.writeln).toHaveBeenCalledWith(''); // For the newline
    // Check that prompt was written again
    expect(mockTerminalInstance.write.mock.calls.length).toBe(initialWriteCount + 1); // One more prompt
    expect(mockTerminalInstance.write).toHaveBeenLastCalledWith(expect.stringContaining(`${projectId} $ `));
    expect(api.post).not.toHaveBeenCalled(); // No API call for empty command
  });

  // Test command history (Up/Down arrows)
  // Simulating arrow keys is tricky as onData receives processed key sequences.
  // Arrow keys are typically multi-character escape sequences (e.g., '\x1b[A' for Up).
  it('cycles through command history with Up/Down arrow keys', async () => {
    render(TerminalComponent);

    // Command 1
    api.post.mockResolvedValueOnce({ stdout: 'cmd1 out', exitCode: 0 });
    "cmd1".split('').forEach(char => simulateTerminalInput(char));
    simulateTerminalInput('\r');
    await waitFor(() => expect(api.post).toHaveBeenCalledWith(expect.anything(), {command: 'cmd1'}));
    
    // Command 2
    api.post.mockResolvedValueOnce({ stdout: 'cmd2 out', exitCode: 0 });
    "cmd2".split('').forEach(char => simulateTerminalInput(char));
    simulateTerminalInput('\r');
    await waitFor(() => expect(api.post).toHaveBeenCalledWith(expect.anything(), {command: 'cmd2'}));

    // At this point, command history should be ["cmd1", "cmd2"]
    // Current command line is empty.

    // Press Up Arrow (should show "cmd2")
    // The mockTerminalInstance.write calls for 'c', 'm', 'd', '2' should occur
    // And currentCommand in component state becomes "cmd2"
    vi.clearAllMocks(); // Clear mocks for api.post and terminal.write/writeln before arrow keys
    simulateTerminalInput('\x1b[A'); // Up Arrow
    expect(mockTerminalInstance.write).toHaveBeenCalledWith('cmd2');
    
    // Press Up Arrow again (should show "cmd1")
    // First, existing "cmd2" is cleared (4 * '\b \b'), then "cmd1" is written.
    vi.clearAllMocks();
    simulateTerminalInput('\x1b[A'); // Up Arrow
    expect(mockTerminalInstance.write).toHaveBeenCalledTimes(1 + 4); // "cmd1" + 4 clear operations
    expect(mockTerminalInstance.write).toHaveBeenCalledWith('cmd1');


    // Press Down Arrow (should show "cmd2")
    vi.clearAllMocks();
    simulateTerminalInput('\x1b[B'); // Down Arrow
    expect(mockTerminalInstance.write).toHaveBeenCalledWith('cmd2');

    // Press Down Arrow again (should clear to empty command line)
    vi.clearAllMocks();
    simulateTerminalInput('\x1b[B'); // Down Arrow
    // Expect clear operations for "cmd2"
    expect(mockTerminalInstance.write).toHaveBeenCalledTimes(4); // 4 * '\b \b'
    expect(mockTerminalInstance.write).toHaveBeenCalledWith('\b \b');


    // Submit this (empty) command to verify currentCommand is ""
    vi.clearAllMocks(); // Clear api.post mocks
    simulateTerminalInput('\r');
    expect(api.post).not.toHaveBeenCalled(); // No command sent
  });

  afterEach(() => {
    // Ensure component is destroyed if necessary, or Vitest handles it with `render`
  });
});
