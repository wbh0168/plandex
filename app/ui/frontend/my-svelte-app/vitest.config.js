import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [
    svelte({ hot: !process.env.VITEST }), // Disable HMR in test environment
  ],
  test: {
    globals: true, // Use global APIs like describe, it, expect
    environment: 'jsdom', // Simulate browser environment
    setupFiles: ['./vitest-setup.js'], // Optional setup file (e.g., for global mocks, jest-dom extensions)
    include: ['src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'], // Test file patterns
    deps: {
      // Ensure svelte internal dependencies are processed correctly by Vitest
      // This might be needed if you encounter issues with Svelte component resolution in tests.
      // inline: [/svelte/], 
    },
    // For @testing-library/jest-dom matchers
    // extends: {
    //   expect: ['@testing-library/jest-dom/matchers'],
    // },
    css: false, // Disable CSS processing if not needed for tests, can speed up
    coverage: {
      provider: 'v8', // or 'istanbul'
      reporter: ['text', 'json', 'html'],
    },
  },
});
