import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import monacoEditorPlugin from 'vite-plugin-monaco-editor';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte(),
    // Default options are fine for most cases, customize if needed
    monacoEditorPlugin({}) 
  ],
  server: {
    // Configure the dev server proxy if your UI backend (Go) runs on a different port
    // and you want to avoid CORS issues during development.
    // Example: Proxy API requests to a Go backend running on :8081
    proxy: {
      '/api': {
        target: 'http://localhost:8081', // Your Go UI backend address
        changeOrigin: true, // Recommended for virtual hosted sites
        // secure: false, // Uncomment if your backend uses self-signed SSL certs
      }
    }
  }
})
