---
sidebar_label: 'Real-time Preview'
sidebar_position: 4
---

# Real-time HTML/CSS/JS Preview

The Real-time Preview feature provides an interactive sandbox environment within the Plandex UI for frontend code experimentation. It consists of three code editors (HTML, CSS, JavaScript) and a live preview pane that updates as you type.

## Accessing the Real-time Preview

Navigate to the "Real-time Preview" section from the main UI navigation bar.

[Screenshot: Main navigation bar highlighting "Real-time Preview"]

## Layout

The Real-time Preview interface is divided into two main panes:

*   **Editors Pane (Top):** Contains three side-by-side code editors for HTML, CSS, and JavaScript.
    [Screenshot: Editors pane showing HTML, CSS, and JS editors side-by-side]
*   **Preview Pane (Bottom):** An iframe that renders the combined output of the HTML, CSS, and JavaScript code from the editors above.
    [Screenshot: Preview pane showing the rendered output]

## Features

### Code Editors

*   **Syntax Highlighting:** Each editor (HTML, CSS, JavaScript) provides syntax highlighting to improve code readability, powered by `simple-code-editor` and `highlight.js`.
*   **Real-time Updates:** As you type in any of the code editors, the preview pane updates automatically.
*   **Labels:** Each editor is clearly labeled "HTML", "CSS", or "JavaScript".

### Live Preview Pane

*   **Immediate Rendering:** The iframe in the preview pane renders the HTML structure, applies the CSS styles, and executes the JavaScript code.
*   **Debounced Updates:** To optimize performance, the preview update is debounced. This means it waits for a short pause in your typing (e.g., 250 milliseconds) before refreshing the iframe. This prevents the preview from updating too frequently during rapid typing.
*   **Sandboxed Environment:** The iframe is sandboxed (`sandbox="allow-scripts allow-same-origin"`) for security. This restricts some capabilities of the code running within the iframe, such as accessing parent window resources directly, though `srcdoc` content is generally treated as a unique opaque origin. JavaScript execution is enabled via `allow-scripts`.

## How to Use

1.  **Enter HTML:** Type your HTML structure into the "HTML" editor.
    *   Example: `<h1>My Title</h1><div id="content">Hello!</div>`
2.  **Style with CSS:** Add CSS rules to the "CSS" editor.
    *   Example: `body { font-family: Arial, sans-serif; } h1 { color: navy; } #content { padding: 10px; background-color: #f0f0f0; }`
3.  **Add Interactivity with JavaScript:** Write JavaScript code in the "JavaScript" editor.
    *   Example: `const contentDiv = document.getElementById('content'); if (contentDiv) { contentDiv.textContent = 'Hello from JavaScript!'; } console.log('Preview script executed.');`
4.  **Observe Preview:** The preview pane will render your code. Any `console.log` statements from your JavaScript will appear in your browser's developer console, not directly in the preview pane.

### Example Initial Code

When you first open the Real-time Preview, it might come with some default placeholder code:

*   **HTML:**
    ```html
    <h1>Hello, World!</h1>
    <p>Type your HTML here.</p>
    ```
*   **CSS:** (Note: `_orange_` is sanitized to `orange` by the component on load)
    ```css
    body {
      font-family: sans-serif;
      color: #333;
    }
    h1 {
      color: orange; /* Was initially _orange_ */
    }
    ```
*   **JavaScript:** (Note: `querySelector_` is sanitized to `querySelector` by the component on load)
    ```javascript
    console.log("JavaScript loaded!");
    const h1 = document.querySelector("h1"); // Was initially querySelector_
    if (h1) h1.textContent += " (modified by JS)";
    ```

### JavaScript Error Handling

The JavaScript code you write is wrapped in a `try...catch` block within the iframe. If your script throws an error, it will be caught, and a message will be logged to the browser's developer console (e.g., "Error in user script: ..."). This prevents script errors from completely breaking the preview iframe in many cases.

## Tips

*   Use your browser's developer tools (inspect element, console) to debug the content running inside the iframe preview.
*   The preview is for quick experimentation. For complex projects, use a dedicated local development environment.
*   Remember the sandbox restrictions if certain browser APIs or interactions don't behave as expected.
