import { expect } from 'vitest';
import * =>testing-library/jest-dom/matchers';

// Extend Vitest's expect interface with jest-dom matchers
expect.extend(matchers);

// Mock localStorage for tests
const localStorageMock = (function() {
  let store = {};
  return {
    getItem(key) {
      return store[key] || null;
    },
    setItem(key, value) {
      store[key] = value.toString();
    },
    removeItem(key) {
      delete store[key];
    },
    clear() {
      store = {};
    }
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

// Mock window.confirm
Object.defineProperty(window, 'confirm', {
  writable: true,
  value: () => true, // Default to true, can be spied on and changed per test
});

// Mock window.alert
Object.defineProperty(window, 'alert', {
  writable: true,
  value: () => {}, // Default to no-op, can be spied on
});

// Mock window.scrollTo (often called by UI libraries on navigation or content change)
Object.defineProperty(window, 'scrollTo', {
  writable: true,
  value: () => {},
});

// Mock for navigation in Svelte components if not using svelte-routing's navigate
// or if testing components that might call location changes directly.
// If using svelte-routing, its navigate function can be mocked where used.
// Object.defineProperty(window, 'location', {
//   writable: true,
//   value: {
//     href: '',
//     pathname: '',
//     assign: vi.fn(),
//     replace: vi.fn(),
//   }
// });

// Clean up mocks after each test (optional, but good practice)
// import { afterEach } from 'vitest';
// afterEach(() => {
//   localStorageMock.clear();
//   vi.restoreAllMocks(); // If using vi.spyOn or vi.mock
// });
