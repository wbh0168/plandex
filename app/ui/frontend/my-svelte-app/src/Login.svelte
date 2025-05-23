<script>
  import { onMount } from 'svelte';
  import { authStore, login as processLogin, initializeAuth } from './lib/authStore';
  import api from './lib/api';
  import { navigate } from 'svelte-routing'; // Assuming svelte-routing is or will be used

  let email = '';
  let pin = ''; // Can be PIN or sign-in code
  let isSignInCode = false; // Toggle for PIN vs Sign-in code
  let isLoading = false;
  let errorMessage = '';

  // Redirect if already authenticated
  onMount(() => {
    // Initialize auth to check if already logged in from localStorage
    // initializeAuth(); // This should be called once in App.svelte or main.js
    if ($authStore.isAuthenticated) {
      navigate('/', { replace: true }); // Or to a dashboard route
    }
  });

  async function handleLogin() {
    isLoading = true;
    errorMessage = '';
    try {
      const credentials = {
        email: email,
        pin: pin,
        is_sign_in_code: isSignInCode,
      };
      // The api.login function calls the UI backend's /api/ui/auth/login
      const sessionResponse = await api.login(credentials); 
      
      if (sessionResponse && sessionResponse.Token) {
        processLogin(sessionResponse); // Update authStore and localStorage
        // Navigate to a protected route, e.g., dashboard or the main view
        // If using svelte-routing, navigate might be imported or passed as prop
        navigate('/', { replace: true }); // Replace '/ 'with your main app route
      } else {
        // This case should ideally be handled by api.js throwing an error for non-ok responses
        errorMessage = 'Login failed: Invalid response from server.';
      }
    } catch (error) {
      console.error('Login error:', error);
      errorMessage = error.message || (error.body && error.body.message) || 'Login failed. Please check your credentials or network.';
      if (error.status === 401) {
        errorMessage = 'Invalid credentials. Please try again.';
      }
    } finally {
      isLoading = false;
    }
  }
</script>

<style>
  .login-container {
    max-width: 400px;
    margin: 50px auto;
    padding: 30px;
    border: 1px solid #ccc;
    border-radius: 8px;
    box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    background-color: #fff;
  }
  h2 {
    text-align: center;
    color: #333;
    margin-bottom: 25px;
  }
  .form-group {
    margin-bottom: 20px;
  }
  label {
    display: block;
    margin-bottom: 8px;
    color: #555;
    font-weight: bold;
  }
  input[type="email"],
  input[type="password"], /* Using type="password" for PIN for masking */
  input[type="text"] {
    width: 100%;
    padding: 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    box-sizing: border-box;
    font-size: 16px;
  }
  .checkbox-group {
    margin-bottom: 20px;
    display: flex;
    align-items: center;
  }
  .checkbox-group input[type="checkbox"] {
    margin-right: 8px;
  }
  .checkbox-group label {
    margin-bottom: 0; /* Reset bottom margin for checkbox label */
    font-weight: normal;
    color: #555;
  }
  button {
    width: 100%;
    padding: 12px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
    font-weight: bold;
    transition: background-color 0.2s;
  }
  button:hover {
    background-color: #0056b3;
  }
  button:disabled {
    background-color: #aaa;
    cursor: not-allowed;
  }
  .error-message {
    color: red;
    margin-top: 15px;
    text-align: center;
    font-size: 14px;
  }
</style>

<div class="login-container">
  <h2>Plandex UI Login</h2>
  <form on:submit|preventDefault={handleLogin}>
    <div class="form-group">
      <label for="email">Email:</label>
      <input type="email" id="email" bind:value={email} required disabled={isLoading} />
    </div>
    <div class="form-group">
      <label for="pin">{isSignInCode ? 'Sign-in Code' : 'PIN'}:</label>
      <input type={isSignInCode ? "text" : "password"} id="pin" bind:value={pin} required disabled={isLoading} />
    </div>
    <div class="checkbox-group">
      <input type="checkbox" id="isSignInCode" bind:checked={isSignInCode} disabled={isLoading} />
      <label for="isSignInCode">Use Sign-in Code</label>
    </div>
    <button type="submit" disabled={isLoading}>
      {isLoading ? 'Logging in...' : 'Login'}
    </button>
    {#if errorMessage}
      <p class="error-message">{errorMessage}</p>
    {/if}
  </form>
</div>
