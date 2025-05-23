---
sidebar_label: 'Terminal Input'
sidebar_position: 5
---

# Terminal Input

The Terminal Input feature in the Plandex UI provides a command-line interface to execute commands directly within the context of your selected project. This can be useful for various tasks like running scripts, listing directory contents, or performing simple operations.

## Accessing the Terminal

Navigate to the "Terminal" section from the main UI navigation bar.

[Screenshot: Main navigation bar highlighting "Terminal"]

## Interface Overview

The terminal interface aims to replicate a standard command-line experience:

*   **Display Area:** A text area where commands are entered and their output (both standard output and standard error) is displayed.
*   **Prompt:** A command prompt is shown, typically indicating the current project context (e.g., `test-project-123 $ `).
*   **Command History:** You can navigate through previously entered commands using the Up and Down arrow keys.
*   **Input Line:** Where you type your commands.

[Screenshot: Terminal view showing the prompt, an example command, and its output]

## Features

### Executing Commands

1.  **Type Command:** Enter the command you wish to execute at the prompt.
    *   Example: `ls -la` or `cat myfile.txt`.
2.  **Press Enter:** After typing your command, press the Enter key.
3.  **Output Display:**
    *   The command's standard output (`stdout`) will be printed directly to the terminal.
    *   Any error messages (`stderr`) will also be printed, often highlighted in a different color (e.g., red).
    *   An exit code for the command will be displayed after the output (e.g., `Exit Code: 0` for success, non-zero for errors).
    *   A new prompt will appear, ready for your next command.

[Screenshot: Terminal showing a command `ls -la` and its successful output, followed by a new prompt]
[Screenshot: Terminal showing a command `cat non_existent_file.txt` and its error output (stderr) and non-zero exit code]

### Command Input

*   **Backspace:** Use the Backspace key to delete characters from the current command line.
*   **Enter:** Submits the current command for execution. If the command line is empty, it simply displays a new prompt.

### Command History

*   **Up Arrow:** Press the Up arrow key to cycle backwards through previously executed commands.
*   **Down Arrow:** Press the Down arrow key to cycle forwards through the command history, or to clear the current line if at the end of the history.

## Important Notes & Security Considerations

*   **Project Context:** Commands are executed within a specific directory on the Plandex server associated with the currently selected project (e.g., `PLANDEX_BASE_DIR/orgs/{orgId}/projects/{projectId}/managed_files/`). Be aware of the current working directory when running commands.
*   **No Real-time Streaming (Current Limitation):** The current implementation executes commands synchronously. This means the UI waits for the command to complete before displaying its full output. For long-running commands, the UI may appear unresponsive until the command finishes. Real-time streaming of output is not yet supported.
*   **Security:**
    *   **Command Execution Scope:** The ability to execute commands is a powerful feature. Ensure that only authorized users have access to this functionality.
    *   **Input Sanitization:** While some basic command splitting is done by the backend, the range of allowed commands and the robustness of input sanitization are critical for security. Avoid executing arbitrary or untrusted command strings.
    *   **Allowed Commands:** In a production environment, it is strongly recommended to restrict command execution to a predefined allow-list of safe commands or to implement more sophisticated parsing and validation to prevent malicious activities like command injection or unauthorized access. The current implementation may not have a strict allow-list.
    *   **Environment Access:** Be mindful that executed commands will run with the permissions of the Plandex server process and within the defined project directory.
*   **Hardcoded Project ID:** In the current version of the UI, the `projectId` is hardcoded (e.g., `test-project-123`). Future versions will allow dynamic project selection.

Always exercise caution when using the terminal input, especially in shared or production environments.
