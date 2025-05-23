import { authStore, logout, selectOrg as updateAuthStoreOrg } from './authStore';
import { get } from 'svelte/store';

const BASE_URL = '/api/ui'; // Proxied by Vite dev server to UI backend

// Helper to construct the Authorization header value
function getAuthorizationHeader() {
  const { authHeader } = get(authStore);
  return authHeader;
}

async function request(method, path, data, options = {}) {
  const headers = options.headers || {};
  const authHeaderValue = getAuthorizationHeader();

  if (authHeaderValue) {
    headers['Authorization'] = authHeaderValue;
  }

  if (!(data instanceof FormData)) { // Don't set Content-Type for FormData
    headers['Content-Type'] = 'application/json';
  }

  const config = {
    method: method,
    headers: headers,
  };

  if (data) {
    config.body = (data instanceof FormData) ? data : JSON.stringify(data);
  }

  try {
    const response = await fetch(`${BASE_URL}${path}`, config);

    if (response.status === 401 && !options.skipAuthRedirect) {
      // Unauthorized: Token might be expired or invalid
      logout(); // Clear auth state
      window.location.href = '/login'; // Redirect to login
      // Throw an error to prevent further processing in the calling function
      throw new Error('Unauthorized - Redirecting to login');
    }

    if (options.isBlob) {
        if (!response.ok) {
            // Try to parse error as JSON, then fall back to text
            let errorPayload;
            try {
                errorPayload = await response.json();
            } catch (e) {
                errorPayload = { message: await response.text() };
            }
            throw { status: response.status, ...errorPayload };
        }
        return response.blob();
    }
    
    const responseData = await response.json().catch(() => {
        // Handle cases where response might not be JSON (e.g., empty body on 204)
        if (response.ok && response.status === 204) return null; 
        return { message: 'Received non-JSON response from server' };
    });


    if (!response.ok) {
      // Throw an error object that includes status and the parsed JSON body (which might contain error details)
      throw { status: response.status, ...responseData };
    }

    return responseData;
  } catch (error) {
    console.error(`API ${method} request to ${path} failed:`, error);
    throw error; // Re-throw to be caught by the calling function
  }
}

export const api = {
  get: (path, options) => request('GET', path, null, options),
  post: (path, data, options) => request('POST', path, data, options),
  put: (path, data, options) => request('PUT', path, data, options),
  delete: (path, options) => request('DELETE', path, null, options),
  // Special login function that doesn't require auth headers
  login: (credentials) => {
    // The login handler itself in auth_handlers.go doesn't expect an Authorization header
    // It expects credentials and returns a session which includes the token for future auth.
    return request('POST', '/auth/login', credentials, { skipAuthRedirect: true });
  },
  logout: () => {
    // Logout requires sending the current Authorization header to invalidate the session on the server
    return request('POST', '/auth/logout', null, { skipAuthRedirect: true }); 
  },
  getSession: () => {
    // getSession also requires sending the current Authorization header
    return request('GET', '/auth/session', null, { skipAuthRedirect: true });
  },
  // Utility to update the orgId in the authStore and thus in the Authorization header for subsequent requests
  selectOrg: (newOrgId) => {
    updateAuthStoreOrg(newOrgId);
    // No API call needed here, just updating local store which affects future headers.
    return Promise.resolve();
  }
};

export default api;
