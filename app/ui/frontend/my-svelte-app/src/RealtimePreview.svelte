<script>
  import { onMount, tick } from 'svelte';
  // Attempt to use svelte-simple-code-editor
  // If this causes issues during npm install or build, will revert to textareas
  import CodeEditor from 'simple-code-editor/CodeEditor.svelte'; 
  import hljs from 'highlight.js/lib/core';
  import xml from 'highlight.js/lib/languages/xml'; // For HTML
  import css from 'highlight.js/lib/languages/css';
  import javascript from 'highlight.js/lib/languages/javascript';
  
  hljs.registerLanguage('xml', xml);
  hljs.registerLanguage('css', css);
  hljs.registerLanguage('javascript', javascript);

  let htmlCode = `<h1>Hello, World!</h1>\n<p>Type your HTML here.</p>`;
  let cssCode = `body {\n  font-family: sans-serif;\n  color: #333;\n}\nh1 {\n  color:_orange_;\n}`; // Intentionally using _orange_ to test if it breaks or renders as is
  let jsCode = `console.log("JavaScript loaded!");\nconst h1 = document.querySelector_("h1");\nif (h1) h1.textContent += " (modified by JS)";`; // Intentionally using querySelector_

  let iframeElement;
  let lastUpdateTime = 0;
  const debounceTime = 250; // milliseconds

  function updatePreview() {
    const now = Date.now();
    if (now - lastUpdateTime < debounceTime && !(htmlCode === '' && cssCode === '' && jsCode === '')) {
      // If not enough time has passed since last update, schedule another update
      setTimeout(updatePreview, debounceTime - (now - lastUpdateTime));
      return;
    }
    lastUpdateTime = now;

    const previewDoc = `
      <!DOCTYPE html>
      <html>
      <head>
        <style>${cssCode}</style>
      </head>
      <body>
        ${htmlCode}
        <script>
          try {
            ${jsCode}
          } catch (e) {
            console.error("Error in user script:", e);
          }
        <\/script> 
      </body>
      </html>
    `;
    // Using srcdoc for the iframe content
    if (iframeElement) {
      iframeElement.srcdoc = previewDoc;
    }
  }
  
  // Trigger initial update and then update on code changes
  onMount(() => {
    // Replace invalid characters for CSS color (example: _orange_)
    // This is a simplistic way; a more robust CSS parser/sanitizer might be needed for complex cases.
    cssCode = cssCode.replace(/_([a-zA-Z]+)_/g, '$1'); 
    // Replace invalid JS function name (example: querySelector_)
    jsCode = jsCode.replace(/querySelector_/g, 'querySelector');

    updatePreview();
  });

  // $: indicates that this block of code should re-run whenever htmlCode, cssCode, or jsCode changes
  $: if (typeof htmlCode === 'string' || typeof cssCode === 'string' || typeof jsCode === 'string') {
    // Debounce the update
    const handler = setTimeout(() => {
      updatePreview();
    }, debounceTime);
    
    // Cleanup function for the timeout if the component is destroyed or variables change again quickly
    // Svelte doesn't have a direct cleanup for reactive statements like this,
    // but for debouncing, this pattern is generally okay.
    // More complex scenarios might need manual clearTimeout in onDestroy.
  }

</script>

<style>
  .preview-container {
    display: flex;
    flex-direction: column;
    height: 90vh; /* Make it take most of the viewport height */
    padding: 10px;
    gap: 10px; /* Spacing between elements */
  }

  .editors-pane {
    display: flex;
    flex-direction: row; /* Editors side-by-side */
    gap: 10px;
    height: 50%; /* Editors take half the vertical space */
  }

  .editor-wrapper {
    flex: 1; /* Each editor takes equal width */
    display: flex;
    flex-direction: column;
    border: 1px solid #ccc;
    border-radius: 4px;
    overflow: hidden; /* Important for editor component's own scrollbars */
  }

  .editor-wrapper label {
    background-color: #f0f0f0;
    padding: 8px 12px;
    font-weight: bold;
    border-bottom: 1px solid #ccc;
  }

  /* Style for svelte-simple-code-editor or fallback textarea */
  :global(.code-editor), /* Class used by svelte-simple-code-editor */
  textarea {
    flex-grow: 1; /* Take remaining space in wrapper */
    border: none;
    outline: none;
    padding: 10px;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
    font-size: 14px;
    line-height: 1.5;
    resize: none; /* Disable textarea resize handle if it appears */
    width: 100% !important; /* Ensure it fills wrapper */
    box-sizing: border-box; /* Ensure padding doesn't break layout */
  }

  .preview-pane {
    flex: 1; /* Takes the remaining vertical space */
    border: 1px solid #ccc;
    border-radius: 4px;
    overflow: auto; /* Scrollbars if content overflows */
  }

  iframe {
    width: 100%;
    height: 100%;
    border: none;
  }
</style>

<div class="preview-container">
  <div class="editors-pane">
    <div class="editor-wrapper">
      <label for="html-editor">HTML</label>
      <CodeEditor bind:value={htmlCode} language={hljs.getLanguage('xml')} highlightJs={hljs} id="html-editor" />
      <!-- Fallback to textarea if CodeEditor fails: -->
      <!-- <textarea bind:value={htmlCode} on:input={updatePreview} id="html-editor"></textarea> -->
    </div>

    <div class="editor-wrapper">
      <label for="css-editor">CSS</label>
      <CodeEditor bind:value={cssCode} language={hljs.getLanguage('css')} highlightJs={hljs} id="css-editor" />
      <!-- <textarea bind:value={cssCode} on:input={updatePreview} id="css-editor"></textarea> -->
    </div>

    <div class="editor-wrapper">
      <label for="js-editor">JavaScript</label>
      <CodeEditor bind:value={jsCode} language={hljs.getLanguage('javascript')} highlightJs={hljs} id="js-editor" />
      <!-- <textarea bind:value={jsCode} on:input={updatePreview} id="js-editor"></textarea> -->
    </div>
  </div>

  <div class="preview-pane">
    <iframe title="Realtime Preview" bind:this={iframeElement} sandbox="allow-scripts allow-same-origin"></iframe>
    <!-- allow-same-origin is needed for some JS interactions like modifying parent if they were from same origin,
         but srcdoc is treated as unique opaque origin, so it's more about what scripts *within* the iframe can do.
         allow-scripts is necessary for the JS code to run.
         Consider further restrictions if needed: allow-forms, allow-popups, etc.
    -->
  </div>
</div>
