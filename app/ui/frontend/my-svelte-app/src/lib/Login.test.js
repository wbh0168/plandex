// Note: Vitest typically expects test files in `src` or a dedicated `tests` dir.
// If `src/lib/Login.test.js` is not picked up, might need to adjust Vitest config `include`
// or move to `src/components/Login.test.js` or `src/Login.test.js` if Login.svelte is in `src`.
// Assuming Login.svelte is in `src/` for this path.
// If Login.svelte is in `src/lib/`, then this path is fine.
// For components, it's common to put them in a `components` subfolder.
// Let's assume Login.svelte is in `src/` for now and adjust if needed.
// Path should be: app/ui/frontend/my-svelte-app/src/Login.test.js

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import Login from '../Login.svelte'; // Adjust path if Login.svelte is elsewhere (e.g. ./components/Login.svelte)
import { authStore, login as processLoginCall } from './authStore'; // actual store
import api from './api'; // to mock api.login
import { navigate } from 'svelte-routing'; // to mock navigation

// Mock svelte-routing's navigate
vi.mock('svelte-routing', async (importOriginal) => {
  const original = await importOriginal();
  return {
    ...original,
    navigate: vi.fn(),
  };
});

// Mock api.login
vi.mock('./api', () => ({
  default: {
    login: vi.fn(),
    // Mock other api functions if Login.svelte uses them, but it primarily uses login.
  }
}));

describe('Login.svelte', () => {
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
    vi.clearAllMocks(); // Clears mock usage data
    // Reset authStore to logged-out state
    authStore.set({ 
      isAuthenticated: false, currentUser: null, token: null, orgId: null, orgs: [], authHeader: null 
    });
    localStorage.clear();
  });

  it('renders login form correctly', () => {
    render(Login);
    expect(screen.getByLabelText('Email:')).toBeInTheDocument();
    expect(screen.getByLabelText('PIN:')).toBeInTheDocument(); // Default label
    expect(screen.getByRole('button', { name: 'Login' })).toBeInTheDocument();
    expect(screen.getByRole('checkbox', { name: 'Use Sign-in Code' })).toBeInTheDocument();
  });

  it('toggles label for PIN/Sign-in Code when checkbox is clicked', async () => {
    render(Login);
    const checkbox = screen.getByRole('checkbox', { name: 'Use Sign-in Code' });
    
    expect(screen.getByLabelText('PIN:')).toBeInTheDocument();
    await fireEvent.click(checkbox);
    expect(screen.getByLabelText('Sign-in Code:')).toBeInTheDocument();
    await fireEvent.click(checkbox);
    expect(screen.getByLabelText('PIN:')).toBeInTheDocument();
  });

  it('submits form data and calls api.login on submit', async () => {
    api.login.mockResolvedValue(sampleSessionResponse); // Mock successful API login
    
    render(Login);
    const emailInput = screen.getByLabelText('Email:');
    const pinInput = screen.getByLabelText('PIN:'); // Default input
    const loginButton = screen.getByRole('button', { name: 'Login' });

    await fireEvent.input(emailInput, { target: { value: 'test@example.com' } });
    await fireEvent.input(pinInput, { target: { value: '12345' } });
    await fireEvent.click(loginButton);

    expect(api.login).toHaveBeenCalledOnce();
    expect(api.login).toHaveBeenCalledWith({
      email: 'test@example.com',
      pin: '12345',
      is_sign_in_code: false, // Checkbox not clicked
    });
  });
  
  it('calls api.login with is_sign_in_code true when checkbox is checked', async () => {
    api.login.mockResolvedValue(sampleSessionResponse);
    render(Login);
    const checkbox = screen.getByRole('checkbox', { name: 'Use Sign-in Code' });
    await fireEvent.click(checkbox); // Check it

    const emailInput = screen.getByLabelText('Email:');
    const signInCodeInput = screen.getByLabelText('Sign-in Code:');
    const loginButton = screen.getByRole('button', { name: 'Login' });
    
    await fireEvent.input(emailInput, { target: { value: 'test@example.com' } });
    await fireEvent.input(signInCodeInput, { target: { value: 'ABCDE' } });
    await fireEvent.click(loginButton);

    expect(api.login).toHaveBeenCalledWith({
      email: 'test@example.com',
      pin: 'ABCDE',
      is_sign_in_code: true,
    });
  });


  it('on successful login, updates authStore and navigates', async () => {
    api.login.mockResolvedValue(sampleSessionResponse);
    const authStoreLoginSpy = vi.spyOn(authStoreModule, 'login'); // Assuming authStoreModule for spying
                                                                // More direct: check store state after.

    render(Login);
    await fireEvent.input(screen.getByLabelText('Email:'), { target: { value: 'test@example.com' } });
    await fireEvent.input(screen.getByLabelText('PIN:'), { target: { value: '12345' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    // Wait for promises to resolve (e.g., api.login, then processLoginCall)
    await waitFor(() => {
      expect(api.login).toHaveBeenCalledTimes(1);
    });
    
    // Check that authStore was updated (via processLoginCall indirectly)
    // processLoginCall is imported as `login` from authStore in Login.svelte
    // We can check the effects on the store
    await waitFor(() => {
        const state = get(authStore);
        expect(state.isAuthenticated).toBe(true);
        expect(state.currentUser.email).toBe(sampleSessionResponse.Email);
        expect(state.token).toBe(sampleSessionResponse.Token);
    });

    // Check navigation
    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith('/', { replace: true });
    });
  });

  it('displays an error message on failed login', async () => {
    const error = { message: 'Invalid credentials test' };
    api.login.mockRejectedValue(error);

    render(Login);
    await fireEvent.input(screen.getByLabelText('Email:'), { target: { value: 'wrong@example.com' } });
    await fireEvent.input(screen.getByLabelText('PIN:'), { target: { value: 'invalid' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    // Wait for error message to appear
    await waitFor(() => {
      expect(screen.getByText(error.message)).toBeInTheDocument();
    });
    
    // Ensure store is not updated
    const state = get(authStore);
    expect(state.isAuthenticated).toBe(false);
    expect(navigate).not.toHaveBeenCalled();
  });
  
  it('displays specific error on 401 status from api.login', async () => {
    const error = { status: 401, message: 'Server says unauthorized' }; // api.js might reformat this
    api.login.mockRejectedValue(error);

    render(Login);
    await fireEvent.input(screen.getByLabelText('Email:'), { target: { value: 'test@example.com' } });
    await fireEvent.input(screen.getByLabelText('PIN:'), { target: { value: '123' } });
    await fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    await waitFor(() => {
      expect(screen.getByText('Invalid credentials. Please try again.')).toBeInTheDocument();
    });
  });

  // This is a conceptual spy placeholder. Actual spying on store methods needs careful setup.
  const authStoreModule = { login: processLoginCall }; 
});
