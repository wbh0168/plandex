// Assuming FileManager.svelte is in src/ or src/components/
// Adjust path if needed: app/ui/frontend/my-svelte-app/src/FileManager.test.js

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, fireEvent, screen, waitFor, act } from '@testing-library/svelte';
import FileManager from '../FileManager.svelte'; // Adjust path to your FileManager.svelte
import api from './api'; // To mock API calls

// Mock the api module
vi.mock('./api', () => ({
  default: {
    get: vi.fn(),    // For fetching files
    post: vi.fn(),   // For uploading files
    delete: vi.fn(), // For deleting files
    // Add other methods if FileManager uses them
  }
}));

// Mock window.confirm (already in vitest-setup.js, but can be explicit for clarity)
// global.confirm = vi.fn(() => true); // Default to 'OK' clicked

describe('FileManager.svelte', () => {
  const mockFiles = [
    { name: 'file1.txt', size: 1024, lastModified: Math.floor(Date.now() / 1000) - 3600 },
    { name: 'image.png', size: 204800, lastModified: Math.floor(Date.now() / 1000) - 7200 },
  ];
  const projectId = 'test-project-123'; // Matches hardcoded projectId in component

  beforeEach(() => {
    vi.clearAllMocks();
    // Reset window.confirm mock before each test if needed for specific scenarios
    global.confirm = vi.fn(() => true); 
  });

  it('renders and fetches files on mount, then displays them', async () => {
    api.get.mockResolvedValue(mockFiles); // Mock successful file list fetch

    render(FileManager);

    expect(screen.getByText(`File Manager (Project: ${projectId})`)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Upload File' })).toBeInTheDocument();
    
    await waitFor(() => {
      expect(api.get).toHaveBeenCalledWith(`/projects/${projectId}/files`);
    });

    // Check if files are rendered
    for (const file of mockFiles) {
      expect(screen.getByText(file.name)).toBeInTheDocument();
      // Check for formatted size and date (can be more specific if needed)
      expect(screen.getByText(new RegExp(file.size === 1024 ? "1 KB" : "200 KB"))).toBeInTheDocument();
    }
  });

  it('displays "No files found" message when no files are fetched', async () => {
    api.get.mockResolvedValue([]); // Mock empty file list
    render(FileManager);
    await waitFor(() => {
      expect(screen.getByText('No files found for this project.')).toBeInTheDocument();
    });
  });
  
  it('displays error message if fetching files fails', async () => {
    const errorMsg = "Network Error Fetching Files";
    api.get.mockRejectedValue({ message: errorMsg });
    render(FileManager);
    await waitFor(() => {
        // Error message might be part of a larger string if component adds context
        expect(screen.getByText(new RegExp(errorMsg))).toBeInTheDocument();
    });
  });

  it('handles file selection and enables upload button', async () => {
    api.get.mockResolvedValue([]); // Start with no files
    render(FileManager);

    const uploadButton = screen.getByRole('button', { name: 'Upload File' });
    expect(uploadButton).toBeDisabled();

    const fileInput = screen.getByLabelText(/Choose File/i, { selector: 'input[type="file"]' }) || screen.getByRole('button', { name: /Choose File/i })?.previousElementSibling;
    // Testing file inputs can be tricky. The above selector might need adjustment
    // based on how your file input is structured or if it's hidden and styled.
    // A common pattern is to have an id on the input: <input type="file" id="fileInput">
    // Then use: screen.getByTestId('file-upload-input') or similar if you add data-testid
    // For now, let's assume a simple input or one found by its implicit role/label.
    // If the input is visually hidden and controlled by a label, target the label then the input.
    // We'll use a more robust way if the default selectors fail.
    // The component uses `id="fileInput"`, so:
    const actualFileInput = document.getElementById('fileInput'); // Not ideal in RTL, but works for jsdom

    const testFile = new File(['content'], 'testfile.txt', { type: 'text/plain' });
    
    // Attach file to input
    // Need to ensure the file input is found correctly.
    // If `screen.getByLabelText` doesn't work due to complex styling:
    const inputElement = screen.getByRole('button', { name: 'Upload File' }).previousElementSibling; // Assuming it's the one before the button
    
    await fireEvent.change(inputElement, { target: { files: [testFile] } });
    
    expect(uploadButton).not.toBeDisabled();
  });


  it('uploads a selected file successfully', async () => {
    api.get.mockResolvedValue([]); // Initial file list
    api.post.mockResolvedValue({ message: 'File testfile.txt uploaded.', filename: 'testfile.txt' });

    render(FileManager);
    
    const fileInput = screen.getByRole('button', { name: 'Upload File' }).previousElementSibling;
    const testFile = new File(['(⌐□_□)'], 'testfile.txt', { type: 'text/plain' });
    await fireEvent.change(fileInput, { target: { files: [testFile] } });

    const uploadButton = screen.getByRole('button', { name: 'Upload File' });
    await fireEvent.click(uploadButton);

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledTimes(1);
      // Check FormData content (can be complex, often easier to check if specific keys exist)
      const formData = api.post.mock.calls[0][1]; // Second argument to api.post
      expect(formData.get('file')).toEqual(testFile);
      expect(api.post).toHaveBeenCalledWith(`/projects/${projectId}/files/upload`, expect.any(FormData));
    });
    
    // Check for success message
    await waitFor(() => {
        expect(screen.getByText(/Successfully uploaded testfile.txt!/)).toBeInTheDocument();
    });

    // Verify file list is refreshed (api.get called again)
    // api.get was called once on mount, so it should be called again after upload
    // To make this distinct, reset mock before upload or check call count.
    api.get.mockResolvedValueOnce([ { name: 'testfile.txt', size: 7, lastModified: Date.now()/1000 } ]); // Mock new file list
    // The component already called fetchFiles after successful upload.
    // We need to ensure this second call happens.
    // The first call is onMount. The second should be after successful upload.
    await waitFor(() => expect(api.get).toHaveBeenCalledTimes(2), { timeout: 2000 });
    await waitFor(() => expect(screen.getByText('testfile.txt')).toBeInTheDocument());

  });

  it('shows an error message if file upload fails', async () => {
    api.get.mockResolvedValue([]);
    const uploadErrorMsg = 'Upload failed on server';
    api.post.mockRejectedValue({ message: uploadErrorMsg });

    render(FileManager);
    const fileInput = screen.getByRole('button', { name: 'Upload File' }).previousElementSibling;
    const testFile = new File(['content'], 'badupload.txt', { type: 'text/plain' });
    await fireEvent.change(fileInput, { target: { files: [testFile] } });
    
    await fireEvent.click(screen.getByRole('button', { name: 'Upload File' }));

    await waitFor(() => {
      expect(screen.getByText(new RegExp(uploadErrorMsg))).toBeInTheDocument();
    });
  });

  it('deletes a file successfully after confirmation', async () => {
    api.get.mockResolvedValue(mockFiles); // Initial files
    api.delete.mockResolvedValue({ message: 'File deleted' }); // Mock successful deletion
    global.confirm = vi.fn(() => true); // Ensure confirm is 'OK'

    render(FileManager);
    await waitFor(() => expect(screen.getByText('file1.txt')).toBeInTheDocument());

    // Find delete button for file1.txt. This assumes a certain DOM structure.
    // A more robust way is data-testid attributes.
    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    // Assuming order in mockFiles matches rendered order.
    await fireEvent.click(deleteButtons[0]); 

    expect(global.confirm).toHaveBeenCalledWith('Are you sure you want to delete file1.txt?');
    
    await waitFor(() => {
      expect(api.delete).toHaveBeenCalledWith(`/projects/${projectId}/files/file1.txt`);
    });

    // Verify file list is refreshed (api.get called again)
    api.get.mockResolvedValueOnce([mockFiles[1]]); // Return list without deleted file
    await waitFor(() => expect(api.get).toHaveBeenCalledTimes(2)); // Mount + after delete
    await waitFor(() => {
        expect(screen.queryByText('file1.txt')).not.toBeInTheDocument();
        expect(screen.getByText('image.png')).toBeInTheDocument();
    });
  });

  it('does not delete a file if confirmation is cancelled', async () => {
    api.get.mockResolvedValue(mockFiles);
    global.confirm = vi.fn(() => false); // Simulate 'Cancel' clicked

    render(FileManager);
    await waitFor(() => expect(screen.getByText('file1.txt')).toBeInTheDocument());

    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    await fireEvent.click(deleteButtons[0]);

    expect(global.confirm).toHaveBeenCalled();
    expect(api.delete).not.toHaveBeenCalled(); // Delete API should not be called
    expect(screen.getByText('file1.txt')).toBeInTheDocument(); // File should still be there
  });
  
  it('shows an error if deleting a file fails', async () => {
    api.get.mockResolvedValue(mockFiles);
    const deleteErrorMsg = "Failed to delete on server";
    api.delete.mockRejectedValue({ message: deleteErrorMsg });
    global.confirm = vi.fn(() => true);

    render(FileManager);
    await waitFor(() => expect(screen.getByText('file1.txt')).toBeInTheDocument());
    
    const deleteButtons = screen.getAllByRole('button', { name: 'Delete' });
    await fireEvent.click(deleteButtons[0]);

    await waitFor(() => {
        expect(screen.getByText(new RegExp(deleteErrorMsg))).toBeInTheDocument();
    });
    expect(screen.getByText('file1.txt')).toBeInTheDocument(); // File still present
  });

  // Test download link construction (not actual download)
  it('renders correct download links for files', async () => {
    api.get.mockResolvedValue(mockFiles);
    render(FileManager);
    await waitFor(() => expect(screen.getByText('file1.txt')).toBeInTheDocument());

    const downloadLinkFile1 = screen.getByRole('link', { name: /Download/i, href: `/api/ui/projects/${projectId}/files/download/file1.txt` });
    // The name might be just "Download". If multiple, need more specific selector.
    // Let's assume for file1.txt, its download link is distinguishable or first.
    // A better way would be to find the row for "file1.txt" then the download link within it.
    
    // This finds all download links and checks the first one.
    const firstDownloadLink = screen.getAllByRole('link', { name: 'Download' })[0];
    expect(firstDownloadLink).toHaveAttribute('href', `/api/ui/projects/${projectId}/files/download/${mockFiles[0].name}`);
    expect(firstDownloadLink).toHaveAttribute('download', mockFiles[0].name);

    const secondDownloadLink = screen.getAllByRole('link', { name: 'Download' })[1];
    expect(secondDownloadLink).toHaveAttribute('href', `/api/ui/projects/${projectId}/files/download/${mockFiles[1].name}`);
    expect(secondDownloadLink).toHaveAttribute('download', mockFiles[1].name);

  });

});
