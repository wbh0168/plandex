import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { authStore, initializeAuth, login, logout, selectOrg } from './authStore';

// Mock localStorage (already in vitest-setup.js, but can be explicit here too if needed)
// For these tests, we'll rely on the global mock from vitest-setup.js

describe('authStore', () => {
  const initialAuthState = {
    isAuthenticated: false,
    currentUser: null,
    token: null,
    orgId: null,
    orgs: [],
    authHeader: null,
  };

  const sampleUser = { id: 'user1', email: 'test@example.com', name: 'Test User' };
  const sampleOrgs = [{ Id: 'org1', Name: 'Org One' }, { Id: 'org2', Name: 'Org Two' }];
  const sampleSessionResponse = {
    UserId: sampleUser.id,
    Email: sampleUser.email,
    UserName: sampleUser.name,
    Token: 'sample-token-123',
    Orgs: sampleOrgs,
    OrgId: 'org1', // Pre-selected orgId
  };

  beforeEach(() => {
    // Reset the store to its initial state and clear localStorage before each test
    authStore.set({ ...initialAuthState }); // Create a new object to avoid mutation issues
    localStorage.clear();
    vi.clearAllMocks(); // Clear any vi spies or mocks
  });

  it('should have correct initial state', () => {
    expect(get(authStore)).toEqual(initialAuthState);
  });

  describe('initializeAuth', () => {
    it('should initialize from localStorage if token and orgId exist', () => {
      localStorage.setItem('authToken', 'stored-token');
      localStorage.setItem('authOrgId', 'stored-org');
      localStorage.setItem('authUser', JSON.stringify(sampleUser));
      localStorage.setItem('authOrgs', JSON.stringify(sampleOrgs));

      initializeAuth();

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(true); // Assumes presence of token means authenticated for initialization
      expect(state.token).toBe('stored-token');
      expect(state.orgId).toBe('stored-org');
      expect(state.currentUser).toEqual(sampleUser);
      expect(state.orgs).toEqual(sampleOrgs);
      expect(state.authHeader).toContain('Bearer ');
    });

    it('should remain in initial state if localStorage is empty', () => {
      initializeAuth();
      expect(get(authStore)).toEqual(initialAuthState);
    });
  });

  describe('login', () => {
    it('should update state and localStorage correctly on login', () => {
      login(sampleSessionResponse);
      const state = get(authStore);

      expect(state.isAuthenticated).toBe(true);
      expect(state.currentUser).toEqual(sampleUser);
      expect(state.token).toBe(sampleSessionResponse.Token);
      expect(state.orgId).toBe(sampleSessionResponse.OrgId);
      expect(state.orgs).toEqual(sampleSessionResponse.Orgs);
      expect(state.authHeader).not.toBeNull();
      expect(state.authHeader).toContain('Bearer ');

      expect(localStorage.getItem('authToken')).toBe(sampleSessionResponse.Token);
      expect(localStorage.getItem('authOrgId')).toBe(sampleSessionResponse.OrgId);
      expect(localStorage.getItem('authUser')).toBe(JSON.stringify(sampleUser));
      expect(localStorage.getItem('authOrgs')).toBe(JSON.stringify(sampleOrgs));
    });
    
    it('should select the first org if OrgId is not in sessionResponse but Orgs array is present', () => {
        const sessionWithoutOrgId = { ...sampleSessionResponse, OrgId: null };
        login(sessionWithoutOrgId);
        const state = get(authStore);
        expect(state.orgId).toBe(sampleOrgs[0].Id);
        expect(localStorage.getItem('authOrgId')).toBe(sampleOrgs[0].Id);
         expect(state.authHeader).toContain(btoa(JSON.stringify({ Token: sessionWithoutOrgId.Token, OrgId: sampleOrgs[0].Id, Hash: ""})));
    });

    it('should handle login with no orgs', () => {
        const sessionWithoutOrgs = { ...sampleSessionResponse, Orgs: [], OrgId: null };
        login(sessionWithoutOrgs);
        const state = get(authStore);
        expect(state.isAuthenticated).toBe(true);
        expect(state.orgId).toBeNull();
        expect(state.orgs).toEqual([]);
        expect(state.authHeader).toBeNull(); // Because orgId is null
        expect(localStorage.getItem('authOrgId')).toBe("null"); // localStorage stores "null" string
    });
    
    it('should not update state if sessionResponse is invalid', () => {
        login({ Token: null, UserId: null }); // Invalid response
        expect(get(authStore)).toEqual(initialAuthState); // Should remain unchanged from initial
    });
  });

  describe('logout', () => {
    it('should reset state and clear localStorage on logout', () => {
      // First, simulate a login
      login(sampleSessionResponse);
      expect(get(authStore).isAuthenticated).toBe(true); // Ensure logged in

      logout();
      const state = get(authStore);

      expect(state).toEqual(initialAuthState);
      expect(localStorage.getItem('authToken')).toBeNull();
      expect(localStorage.getItem('authOrgId')).toBeNull();
      expect(localStorage.getItem('authUser')).toBeNull();
      expect(localStorage.getItem('authOrgs')).toBeNull();
    });
  });

  describe('selectOrg', () => {
    beforeEach(() => {
      // Ensure user is logged in before testing org selection
      login(sampleSessionResponse);
    });

    it('should update orgId and authHeader in store and localStorage', () => {
      const newOrgId = 'org2';
      selectOrg(newOrgId);
      const state = get(authStore);

      expect(state.orgId).toBe(newOrgId);
      expect(state.authHeader).not.toBeNull();
      expect(state.authHeader).toContain(btoa(JSON.stringify({ Token: sampleSessionResponse.Token, OrgId: newOrgId, Hash: ""})));
      expect(localStorage.getItem('authOrgId')).toBe(newOrgId);
    });

    it('should not update if user is not authenticated (edge case, UI should prevent)', () => {
      logout(); // Log out user
      const initialOrgIdBeforeSelect = get(authStore).orgId; // Should be null
      selectOrg('org2');
      const state = get(authStore);
      // State should remain initial (logged out)
      expect(state.orgId).toBe(initialOrgIdBeforeSelect); 
      expect(state.isAuthenticated).toBe(false);
      expect(localStorage.getItem('authOrgId')).toBeNull(); // Should not have been set
    });
  });
});
