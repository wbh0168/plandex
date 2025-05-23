<script>
  import { onMount } from 'svelte';

  // Hardcoded for now, until project selection is implemented
  const currentProjectId = 'test-project-123'; 

  let files = [];
  let selectedFile = null;
  let uploadStatus = '';
  let errorStatus = '';

  async function fetchFiles() {
    uploadStatus = '';
    errorStatus = '';
    try {
      const response = await fetch(`/api/ui/projects/${currentProjectId}/files`);
      if (!response.ok) {
        const errText = await response.text();
        throw new Error(`Failed to fetch files: ${response.status} ${errText}`);
      }
      files = await response.json();
      if (!files) files = [];
    } catch (error) {
      console.error('Fetch files error:', error);
      errorStatus = error.message;
      files = [];
    }
  }

  onMount(fetchFiles);

  function handleFileSelect(event) {
    selectedFile = event.target.files[0];
    uploadStatus = '';
    errorStatus = '';
  }

  async function handleFileUpload() {
    if (!selectedFile) {
      errorStatus = 'Please select a file to upload.';
      return;
    }
    uploadStatus = 'Uploading...';
    errorStatus = '';

    const formData = new FormData();
    formData.append('file', selectedFile);

    try {
      const response = await fetch(`/api/ui/projects/${currentProjectId}/files/upload`, {
        method: 'POST',
        body: formData,
        // Content-Type is automatically set by the browser for FormData
      });

      const result = await response.json();
      if (!response.ok) {
        throw new Error(result.message || `Upload failed: ${response.status}`);
      }
      
      uploadStatus = `Successfully uploaded ${selectedFile.name}! Server message: ${result.message}`;
      selectedFile = null; // Reset file input
      document.getElementById('fileInput').value = ''; // Clear the file input display
      await fetchFiles(); // Refresh file list
    } catch (error) {
      console.error('Upload error:', error);
      errorStatus = error.message;
      uploadStatus = '';
    }
  }

  async function deleteFile(filename) {
    if (!confirm(`Are you sure you want to delete ${filename}?`)) {
      return;
    }
    errorStatus = '';
    try {
      const response = await fetch(`/api/ui/projects/${currentProjectId}/files/${filename}`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        const result = await response.json().catch(() => ({ message: `Failed to delete ${filename}: ${response.status}` }));
        throw new Error(result.message);
      }
      await fetchFiles(); // Refresh file list
    } catch (error) {
      console.error('Delete error:', error);
      errorStatus = error.message;
    }
  }

  function formatBytes(bytes, decimals = 2) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
  }

  function formatDate(timestamp) {
    return new Date(timestamp * 1000).toLocaleString();
  }
</script>

<style>
  .file-manager-container {
    padding: 20px;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  }
  h2 {
    color: #333;
    border-bottom: 2px solid #eee;
    padding-bottom: 10px;
    margin-bottom: 20px;
  }
  .upload-section {
    margin-bottom: 30px;
    padding: 15px;
    background-color: #f9f9f9;
    border: 1px dashed #ddd;
    border-radius: 4px;
  }
  .upload-section input[type="file"] {
    padding: 8px;
    border: 1px solid #ccc;
    border-radius: 4px;
    margin-right: 10px;
  }
  .upload-section button {
    padding: 9px 15px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  .upload-section button:hover {
    background-color: #0056b3;
  }
  .status-message {
    margin-top: 10px;
    padding: 10px;
    border-radius: 4px;
  }
  .status-message.success {
    background-color: #e6ffed;
    color: #28a745;
    border: 1px solid #c3e6cb;
  }
  .status-message.error {
    background-color: #ffe6e6;
    color: #dc3545;
    border: 1px solid #f5c6cb;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 20px;
  }
  th, td {
    border: 1px solid #ddd;
    padding: 12px;
    text-align: left;
    vertical-align: middle;
  }
  th {
    background-color: #f2f2f2;
    font-weight: bold;
  }
  tr:nth-child(even) {
    background-color: #f9f9f9;
  }
  .actions button, .actions a {
    margin-right: 8px;
    padding: 6px 10px;
    border-radius: 4px;
    text-decoration: none;
    cursor: pointer;
    font-size: 14px;
  }
  .actions .download-btn {
    background-color: #28a745;
    color: white;
    border: none;
  }
  .actions .download-btn:hover {
    background-color: #218838;
  }
  .actions .delete-btn {
    background-color: #dc3545;
    color: white;
    border: none;
  }
  .actions .delete-btn:hover {
    background-color: #c82333;
  }
  .no-files {
    text-align: center;
    padding: 20px;
    color: #777;
  }
</style>

<div class="file-manager-container">
  <h2>File Manager (Project: {currentProjectId})</h2>

  <div class="upload-section">
    <input type="file" id="fileInput" on:change={handleFileSelect} />
    <button on:click={handleFileUpload} disabled={!selectedFile}>Upload File</button>
    {#if uploadStatus}
      <div class="status-message success">{uploadStatus}</div>
    {/if}
    {#if errorStatus && !uploadStatus} <!-- Show general errors if no upload status -->
      <div class="status-message error">{errorStatus}</div>
    {/if}
  </div>

  <h3>Uploaded Files</h3>
  {#if files && files.length > 0}
    <table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Size</th>
          <th>Last Modified</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each files as file}
          <tr>
            <td>{file.name}</td>
            <td>{formatBytes(file.size)}</td>
            <td>{formatDate(file.lastModified)}</td>
            <td class="actions">
              <a 
                href={`/api/ui/projects/${currentProjectId}/files/download/${encodeURIComponent(file.name)}`} 
                class="download-btn"
                download={file.name}  {# download attribute hints browser to download #}
              >
                Download
              </a>
              <button class="delete-btn" on:click={() => deleteFile(file.name)}>Delete</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {:else if !errorStatus}
    <p class="no-files">No files found for this project.</p>
  {/if}
  
  {#if errorStatus && files.length === 0} <!-- Show error if fetch failed and no files shown -->
    <div class="status-message error">{errorStatus}</div>
  {/if}
</div>
