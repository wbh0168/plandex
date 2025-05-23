---
sidebar_label: 'File Management'
sidebar_position: 3
---

# File Management

The File Management feature in the Plandex UI allows you to interact with files stored within a specific project's context. You can list, upload, download, and delete files, making it easier to manage project assets, data files, or any other resources your project might need.

## Accessing File Management

1.  Navigate to the "File Manager" section from the main UI navigation bar.
    [Screenshot: Main navigation bar highlighting "File Manager"]
2.  **Project Context:** File management is typically scoped to a project. While the current UI uses a hardcoded project ID (`test-project-123`) for demonstration, in a full setup, you would select or be working within a specific project context. The displayed project ID will be shown at the top of the File Manager view.

## Features

### Listing Files

Upon entering the File Manager, a list of files and folders currently stored for the selected project is displayed in a table.

*   **Columns:**
    *   **Name:** The name of the file or folder.
    *   **Size:** The size of the file (e.g., in KB, MB).
    *   **Last Modified:** The date and time the file was last modified.
    *   **Actions:** Buttons to "Download" or "Delete" the file.

[Screenshot: File Manager view showing a table of files with Name, Size, Last Modified, and Actions columns. Example files listed.]

If no files are present, a message "No files found for this project." will be displayed.

### Uploading Files

1.  **Choose File:** Click the "Choose File" or "Select File" button (the exact label may vary by browser). This will open your system's file dialog.
    [Screenshot: Upload section showing "Choose File" button and "Upload File" button]
2.  Select the file you wish to upload from your local computer.
3.  **Upload:** Once a file is selected, its name might appear next to the button. Click the "Upload File" button.
    *   The button will be enabled only after a file is selected.
4.  **Status:**
    *   During upload, a message like "Uploading..." may appear.
    *   Upon completion, a success message (e.g., "Successfully uploaded `filename.ext`!") or an error message will be displayed.
    *   The file list will automatically refresh to include the newly uploaded file.

[Screenshot: File Manager view showing a success message after an upload, and the new file appearing in the list]

### Downloading Files

1.  In the file list, locate the file you wish to download.
2.  Click the "Download" button in the "Actions" column for that file.
3.  Your browser will start downloading the file, typically saving it to your default downloads folder.

### Deleting Files

1.  In the file list, locate the file you wish to delete.
2.  Click the "Delete" button in the "Actions" column for that file.
3.  A confirmation dialog will appear, asking: "Are you sure you want to delete `filename.ext`?"
    [Screenshot: Confirmation dialog for deleting a file]
4.  **Confirm:** Click "OK" or "Confirm" to proceed with the deletion.
    **Cancel:** Click "Cancel" to abort the deletion.
5.  If confirmed, the file will be permanently removed from the project's storage.
6.  The file list will refresh, and the deleted file will no longer appear.
    *   If an error occurs during deletion, an error message will be displayed.

## Notes

*   **File Storage:** Files are stored in a specific directory on the Plandex server associated with the organization and project (e.g., `PLANDEX_BASE_DIR/orgs/{orgId}/projects/{projectId}/managed_files/`).
*   **File Overwrites:** The current implementation may overwrite files if a file with the same name is uploaded again. Be cautious when uploading files with common names.
*   **Error Handling:** If an operation (upload, list, delete) fails, an error message will typically be displayed at the top of the File Manager section or near the relevant control.
