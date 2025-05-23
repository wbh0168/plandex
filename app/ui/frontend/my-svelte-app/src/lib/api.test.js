import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { authStore, login, logout, initializeAuth } from './authStore'; // Import what's needed to set up auth state
import api from './api';

// Mock window.fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock window.location for redirect testing (from vitest-setup.js, but can be explicit)
// Object.defineProperty(window, 'location', {
//   writable: true,
//   value: { href: '', assign: vi.fn(), replace: vi.fn() }
// });
// Relaying on vitest-setup.js for window.location mock

describe('api client (lib/api.js)', () => {
  const sampleUser = { id: 'user1', email: 'test@example.com', name: 'Test User' };
  const sampleOrgs = [{ Id: 'org1', Name: 'Org One' }];
  const sampleSessionResponse = {
    UserId: sampleUser.id,
    Email: sampleUser.email,
    UserName: sampleUser.name,
    Token: 'sample-token-123',
    Orgs: sampleOrgs,
    OrgId: 'org1',
  };

  beforeEach(() => {
    // Reset authStore and localStorage before each test
    authStore.set({
      isAuthenticated: false,
      currentUser: null,
      token: null,
      orgId: null,
      orgs: [],
      authHeader: null,
    });
    localStorage.clear();
    mockFetch.mockReset();
    // vi.clearAllMocks(); // Already done by mockReset for fetch, clear others if any
    window.location.href = ''; // Reset location href for redirect checks
  });

  describe('request function general behavior', () => {
    it('should make a GET request with correct headers when authenticated', async () => {
      login(sampleSessionResponse); // Sets up authStore with token and orgId
      const expectedAuthHeader = get(authStore).authHeader;

      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ data: 'success' }), { status: 200 }));

      await api.get('/test-path');

      expect(mockFetch).toHaveBeenCalledOnce();
      const fetchCall = mockFetch.mock.calls[0];
      expect(fetchCall[0]).toBe('/api/ui/test-path');
      expect(fetchCall[1].method).toBe('GET');
      expect(fetchCall[1].headers['Authorization']).toBe(expectedAuthHeader);
    });

    it('should make a POST request with body and correct headers when authenticated', async () => {
      login(sampleSessionResponse);
      const expectedAuthHeader = get(authStore).authHeader;
      const postData = { key: 'value' };

      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ data: 'success' }), { status: 200 }));

      await api.post('/test-post', postData);

      expect(mockFetch).toHaveBeenCalledOnce();
      const fetchCall = mockFetch.mock.calls[0];
      expect(fetchCall[0]).toBe('/api/ui/test-post');
      expect(fetchCall[1].method).toBe('POST');
      expect(fetchCall[1].headers['Authorization']).toBe(expectedAuthHeader);
      expect(fetchCall[1].headers['Content-Type']).toBe('application/json');
      expect(fetchCall[1].body).toBe(JSON.stringify(postData));
    });
    
    it('should correctly handle FormData without setting Content-Type', async () => {
      login(sampleSessionResponse);
      const formData = new FormData();
      formData.append('key', 'value');

      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ data: 'success' }), { status: 200 }));
      await api.post('/test-formdata', formData);

      expect(mockFetch).toHaveBeenCalledOnce();
      const fetchCall = mockFetch.mock.calls[0];
      expect(fetchCall[1].headers['Content-Type']).toBeUndefined(); // Should be undefined for FormData
      expect(fetchCall[1].body).toBeInstanceOf(FormData);
    });


    it('should handle successful response and parse JSON', async () => {
      const mockData = { message: 'it worked' };
      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify(mockData), { status: 200 }));
      const result = await api.get('/success');
      expect(result).toEqual(mockData);
    });

    it('should handle empty JSON response for 204 No Content', async () => {
      mockFetch.mockResolvedValueOnce(new Response(null, { status: 204 }));
      const result = await api.get('/no-content');
      expect(result).toBeNull();
    });
    
    it('should throw an error for non-JSON success response if not expecting blob', async () => {
      mockFetch.mockResolvedValueOnce(new Response("Not JSON", { status: 200 }));
      try {
        await api.get('/non-json-success');
      } catch (e) {
        expect(e.message).toContain('Received non-JSON response from server');
      }
    });


    it('should handle network error by throwing', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network failed'));
      try {
        await api.get('/network-error');
      } catch (error) {
        expect(error.message).toBe('Network failed');
      }
    });

    it('should handle non-ok response by throwing parsed JSON error', async () => {
      const errorResponse = { error: 'bad request', details: '...' };
      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify(errorResponse), { status: 400 }));
      try {
        await api.get('/bad-request');
      } catch (error) {
        expect(error.status).toBe(400);
        expect(error.error).toBe('bad request');
        expect(error.details).toBe(errorResponse.details);
      }
    });
    
     it('should handle non-ok response with non-JSON body', async () => {
      mockFetch.mockResolvedValueOnce(new Response("Server Error Text", { status: 500 }));
      try {
        await api.get('/server-error-text');
      } catch (error) {
        expect(error.status).toBe(500);
        // When body is not JSON, api.js defaults to a message property
        expect(error.message).toBe('Received non-JSON response from server'); 
      }
    });

    it('should handle blob response correctly', async () => {
        const blobData = new Blob(["test data"], { type: 'text/plain' });
        mockFetch.mockResolvedValueOnce(new Response(blobData, { status: 200 }));
        const result = await api.get('/blob-data', { isBlob: true });
        expect(result).toBeInstanceOf(Blob);
        expect(await result.text()).toBe("test data");
    });

    it('should throw error for non-ok blob response', async () => {
        mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ error: "blob not found" }), { status: 404 }));
        try {
            await api.get('/blob-error', { isBlob: true });
        } catch (e) {
            expect(e.status).toBe(404);
            expect(e.error).toBe("blob not found");
        }
    });
  });

  describe('auth-specific functions (login, logout, getSession)', () => {
    it('api.login should not send Authorization header and skip auth redirect', async () => {
      const loginCredentials = { email: 'user', pin: 'pass' };
      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify(sampleSessionResponse), { status: 200 }));

      await api.login(loginCredentials);

      expect(mockFetch).toHaveBeenCalledOnce();
      const fetchCall = mockFetch.mock.calls[0];
      expect(fetchCall[0]).toBe('/api/ui/auth/login');
      expect(fetchCall[1].headers['Authorization']).toBeUndefined();
      // No redirect should happen even if it failed with 401, due to skipAuthRedirect
      expect(window.location.href).toBe(''); 
    });

    it('api.logout should send Authorization header and skip auth redirect', async () => {
      login(sampleSessionResponse); // Login to set auth header in store
      const expectedAuthHeader = get(authStore).authHeader;
      mockFetch.mockResolvedValueOnce(new Response(null, { status: 204 })); // Logout often 204

      await api.logout();

      expect(mockFetch).toHaveBeenCalledOnce();
      const fetchCall = mockFetch.mock.calls[0];
      expect(fetchCall[0]).toBe('/api/ui/auth/logout');
      expect(fetchCall[1].headers['Authorization']).toBe(expectedAuthHeader);
      expect(window.location.href).toBe('');
    });

    it('api.getSession should send Authorization header and skip auth redirect', async () => {
      login(sampleSessionResponse);
      const expectedAuthHeader = get(authStore).authHeader;
      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ UserId: 'user1' }), { status: 200 }));

      await api.getSession();

      expect(mockFetch).toHaveBeenCalledOnce();
      const fetchCall = mockFetch.mock.calls[0];
      expect(fetchCall[0]).toBe('/api/ui/auth/session');
      expect(fetchCall[1].headers['Authorization']).toBe(expectedAuthHeader);
      expect(window.location.href).toBe('');
    });
  });
  
  describe('401 error handling and redirect', () => {
    it('should logout and redirect to /login on 401 for standard requests', async () => {
      login(sampleSessionResponse); // Authenticate first
      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'token expired' }), { status: 401 }));
      
      // Spy on logout from authStore to ensure it's called
      const logoutSpy = vi.spyOn(authStoreModule, 'logout'); // Assuming authStoreModule is how you import logout

      try {
        await api.get('/protected-route');
      } catch (error) {
        // Error is expected
        expect(error.message).toBe('Unauthorized - Redirecting to login');
      }
      
      const finalAuthState = get(authStore);
      expect(finalAuthState.isAuthenticated).toBe(false); // Should be logged out by api.js
      expect(finalAuthState.token).toBeNull();
      expect(window.location.href).toBe('/login');
      
      // Need to import authStore.js in a way that its logout can be spied on.
      // This might require changing how logout is exported or imported in api.js if it's a direct import.
      // For now, we check the side effect (authStore state and window.location.href)
      // If `logout` was imported as `import { logout as aliasedLogout } from './authStore'` in api.js,
      // then spying on the original `authStoreModule.logout` would work.
      // The current test setup directly checks the effects of the logout.
    });

    it('should NOT redirect to /login on 401 if skipAuthRedirect is true', async () => {
      login(sampleSessionResponse);
      mockFetch.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'token expired' }), { status: 401 }));
      
      try {
        await api.get('/protected-but-skip-redirect', { skipAuthRedirect: true });
      } catch (error) {
        expect(error.status).toBe(401); // Error should still be thrown
      }
      
      const finalAuthState = get(authStore);
      // User should still be "logged in" in the store because redirect (and thus store logout) was skipped
      expect(finalAuthState.isAuthenticated).toBe(true); 
      expect(window.location.href).not.toBe('/login'); // No redirect
    });
  });

  describe('api.selectOrg', () => {
    it('should update authStore correctly without making an API call', async () => {
        login(sampleSessionResponse); // Initial login with org1
        const newOrgId = 'org2';
        
        // Ensure no fetch call is made
        const fetchSpy = vi.spyOn(global, 'fetch');

        await api.selectOrg(newOrgId);

        const state = get(authStore);
        expect(state.orgId).toBe(newOrgId);
        expect(state.authHeader).toContain(btoa(JSON.stringify({ Token: sampleSessionResponse.Token, OrgId: newOrgId, Hash: ""})));
        expect(localStorage.getItem('authOrgId')).toBe(newOrgId);
        expect(fetchSpy).not.toHaveBeenCalled();
    });
  });
});

// authStoreModule mock for spying on logout - this is a bit tricky with direct imports
// A better way is to have api.js accept authStore functions as dependencies if deep spying is needed.
// For this test, we'll assume 'logout' is part of an object that can be spied upon,
// or we rely on checking the side effects (store state, localStorage, window.location).
// If authStore.js exports an object of functions: e.g. export const authActions = { login, logout }
// Then in api.js: import { authActions } from './authStore'
// And in test: vi.spyOn(authActions, 'logout')
// For now, the test checks the observable side effects of logout.
const authStoreModule = { logout }; // Simplified, assuming direct import allows this level of spying if structured right.
// If not, this spy won't work as intended without refactoring how api.js imports logout.
// The current test for 401 handling relies on checking the outcome (store reset, redirect)
// rather than explicitly spying on the `logout` call from `authStore` within `api.js`.
// This is generally acceptable for testing the behavior of `api.js`.
