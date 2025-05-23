<script>
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from 'xterm';
  // Note: xterm-addon-fit is deprecated, suggests @xterm/addon-fit
  // For this task, using the version installed.
  import { FitAddon } from 'xterm-addon-fit';
  import 'xterm/css/xterm.css';

  // Hardcoded for now, until project selection is implemented
  const currentProjectId = 'test-project-123'; 
  const prompt = `\x1b[1;32m${currentProjectId}\x1b[0m $ `;

  let terminalEl; // Div element to host the terminal
  let term;       // xterm.js Terminal instance
  let fitAddon;   // xterm-addon-fit instance
  let currentCommand = '';
  let commandHistory = [];
  let historyIndex = -1;

  function initializeTerminal() {
    if (!terminalEl) return;

    term = new Terminal({
      cursorBlink: true,
      convertEol: true, // Convert \n to \r\n for proper line endings in terminal
      rows: 20, // Default rows
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#d4d4d4',
        selectionBackground: '#555555',
      }
    });

    fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    term.open(terminalEl);
    fitAddon.fit(); // Fit the terminal to the container size

    term.onData(handleTermData); // Handle user input
    
    // Display initial prompt
    term.write(prompt);

    // Resize observer for the terminal container
    const resizeObserver = new ResizeObserver(() => {
      fitAddon.fit();
    });
    resizeObserver.observe(terminalEl);

    // Focus on the terminal
    term.focus();
    
    return () => { // Cleanup function
      resizeObserver.disconnect();
      term.dispose();
    };
  }

  onMount(() => {
    const cleanup = initializeTerminal();
    return cleanup;
  });

  onDestroy(() => {
    // Any additional cleanup if needed, though initializeTerminal's return handles xterm disposal
  });

  function handleTermData(data) {
    const code = data.charCodeAt(0);

    if (code === 13) { // Enter key
      if (currentCommand.trim().length > 0) {
        term.writeln(''); // New line after command
        sendCommand(currentCommand);
        if (currentCommand !== commandHistory[commandHistory.length -1]) {
          commandHistory.push(currentCommand);
        }
        historyIndex = commandHistory.length; // Reset history index
      } else {
        term.writeln('');
        term.write(prompt); // Show prompt again if empty command
      }
      currentCommand = '';
    } else if (code === 127 || code === 8) { // Backspace
      if (currentCommand.length > 0) {
        term.write('\b \b'); // Move cursor back, erase char, move back again
        currentCommand = currentCommand.slice(0, -1);
      }
    } else if (code === 27) { // Escape sequences (e.g., arrow keys)
        // Check for arrow keys (common escape sequences)
        // \x1b[A (Up), \x1b[B (Down), \x1b[C (Right), \x1b[D (Left)
        if (data === '\x1b[A') { // Up arrow
            if (historyIndex > 0) {
                historyIndex--;
                clearCurrentTerminalLine();
                currentCommand = commandHistory[historyIndex];
                term.write(currentCommand);
            }
        } else if (data === '\x1b[B') { // Down arrow
            if (historyIndex < commandHistory.length - 1) {
                historyIndex++;
                clearCurrentTerminalLine();
                currentCommand = commandHistory[historyIndex];
                term.write(currentCommand);
            } else if (historyIndex === commandHistory.length -1) {
                 historyIndex++;
                 clearCurrentTerminalLine();
                 currentCommand = "";
            }
        }
        // Ignore Right/Left arrows for now or implement cursor movement if needed
    } else if (code >= 32) { // Printable characters
      term.write(data);
      currentCommand += data;
    }
  }
  
  function clearCurrentTerminalLine() {
    const commandLength = currentCommand.length;
    for (let i = 0; i < commandLength; i++) {
        term.write('\b \b');
    }
  }

  async function sendCommand(command) {
    term.write('\r\n'); // Start output on a new line
    try {
      const response = await fetch(`/api/ui/projects/${currentProjectId}/commands`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ command: command }),
      });

      const result = await response.json();

      if (!response.ok) {
        // Handle HTTP errors (e.g., 400, 403, 500)
        // The main server might return JSON with an error message
        term.writeln(`\x1b[1;31mError: ${result.message || `HTTP ${response.status}`}\x1b[0m`);
        if (result.stderr) {
            result.stderr.split('\n').forEach(line => term.writeln(line));
        }
      } else {
        // Command executed, result contains stdout, stderr, exitCode
        if (result.stdout) {
          result.stdout.split('\n').forEach(line => term.writeln(line));
        }
        if (result.stderr) {
          // Display stderr in a different color, e.g., red
          result.stderr.split('\n').forEach(line => term.writeln(`\x1b[31m${line}\x1b[0m`));
        }
        term.writeln(`Exit Code: ${result.exitCode}`);
      }
    } catch (error) {
      // Handle network errors or if response is not JSON
      term.writeln(`\x1b[1;31mClient-side error sending command: ${error.message}\x1b[0m`);
    }
    term.write(prompt); // Show prompt for next command
    term.focus(); // Keep terminal focused
  }

</script>

<style>
  .terminal-container {
    width: 100%;
    height: 70vh; /* Adjust as needed, or make it dynamic */
    padding: 10px;
    background-color: #1e1e1e; /* Match terminal theme if needed */
    box-sizing: border-box;
  }
  /* xterm.css handles internal styling, this is for the container div */
</style>

<div class="terminal-container" bind:this={terminalEl}>
  <!-- xterm.js will attach here -->
</div>
