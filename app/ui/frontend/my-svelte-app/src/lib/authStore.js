import { writable, get } from 'svelte/store';

// Define the shape of the auth state
const initialAuthState = {
  isAuthenticated: false,
  currentUser: null, // Will store user details from SessionResponse (e.g., email, name, id)
  token: null,       // Raw token string from SessionResponse
  orgId: null,       // Selected Organization ID
  orgs: [],          // Array of orgs from SessionResponse
  authHeader: null,  // Stores the full "Bearer <base64_encoded_json_auth_header>"
};

// Create a writable store
export const authStore = writable(initialAuthState);

// Helper to construct the Authorization header value (the "Bearer <encoded_json>" string)
function constructAuthHeader(token, orgId) {
  if (!token || !orgId) return null;
  const authHeaderData = {
    Token: token,
    OrgId: orgId,
    Hash: "", // Client-side hash omitted as per plan
  };
  const jsonAuthHeader = JSON.stringify(authHeaderData);
  const base64AuthHeader = btoa(jsonAuthHeader); // btoa for base64 encoding in browsers
  return `Bearer ${base64AuthHeader}`;
}

// Function to initialize the store from localStorage
export function initializeAuth() {
  if (typeof localStorage !== 'undefined') {
    const storedToken = localStorage.getItem('authToken');
    const storedOrgId = localStorage.getItem('authOrgId');
    const storedUser = localStorage.getItem('authUser');
    const storedOrgs = localStorage.getItem('authOrgs');

    if (storedToken && storedOrgId) {
      authStore.set({
        isAuthenticated: true, // Will be validated by a session check
        currentUser: storedUser ? JSON.parse(storedUser) : null,
        token: storedToken,
        orgId: storedOrgId,
        orgs: storedOrgs ? JSON.parse(storedOrgs) : [],
        authHeader: constructAuthHeader(storedToken, storedOrgId),
      });
    }
  }
}

// Function to handle successful login
export function login(sessionResponse) {
  if (!sessionResponse || !sessionResponse.Token || !sessionResponse.UserId) {
    console.error("Login failed: Invalid session response", sessionResponse);
    return;
  }

  const user = {
    id: sessionResponse.UserId,
    email: sessionResponse.Email,
    name: sessionResponse.UserName,
    // Add other relevant fields from SessionResponse.User if needed
  };
  
  let selectedOrgId = sessionResponse.OrgId; // OrgId might be pre-selected if only one org or from sign-in code
  if (!selectedOrgId && sessionResponse.Orgs && sessionResponse.Orgs.length > 0) {
    selectedOrgId = sessionResponse.Orgs[0].Id; // Default to the first org if multiple and none pre-selected
  } else if (!selectedOrgId && (!sessionResponse.Orgs || sessionResponse.Orgs.length === 0)) {
     console.warn("No organizations found for user, or no default orgId provided in session.");
     // Proceeding without an orgId might cause issues for org-specific API calls.
     // Depending on app logic, might want to force org selection or handle this state.
  }


  const newAuthState = {
    isAuthenticated: true,
    currentUser: user,
    token: sessionResponse.Token,
    orgId: selectedOrgId,
    orgs: sessionResponse.Orgs || [],
    authHeader: constructAuthHeader(sessionResponse.Token, selectedOrgId),
  };

  authStore.set(newAuthState);

  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('authToken', newAuthState.token);
    localStorage.setItem('authOrgId', newAuthState.orgId);
    localStorage.setItem('authUser', JSON.stringify(newAuthState.currentUser));
    localStorage.setItem('authOrgs', JSON.stringify(newAuthState.orgs));
  }
}

// Function to handle logout
export function logout() {
  authStore.set(initialAuthState);
  if (typeof localStorage !== 'undefined') {
    localStorage.removeItem('authToken');
    localStorage.removeItem('authOrgId');
    localStorage.removeItem('authUser');
    localStorage.removeItem('authOrgs');
  }
  // Optionally, redirect to login page or homepage
  // This might be better handled in the component calling logout
}

// Function to update selected organization
export function selectOrg(newOrgId) {
    const currentAuth = get(authStore);
    if (currentAuth.isAuthenticated && currentAuth.token) {
        const newAuthState = {
            ...currentAuth,
            orgId: newOrgId,
            authHeader: constructAuthHeader(currentAuth.token, newOrgId),
        };
        authStore.set(newAuthState);
        if (typeof localStorage !== 'undefined') {
            localStorage.setItem('authOrgId', newOrgId);
        }
    } else {
        console.error("Cannot select org: user not authenticated or token missing.");
    }
}
