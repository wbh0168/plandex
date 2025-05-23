<script>
  import CustomModels from './CustomModels.svelte';
  import ModelPacks from './ModelPacks.svelte';
  import FileManager from './FileManager.svelte';
<script>
  import { Router, Route, link, navigate } from 'svelte-routing';
  import { onMount } from 'svelte';
  import { authStore, initializeAuth, logout as processLogout, selectOrg } from './lib/authStore';
  import api from './lib/api';

  // Import page components
  import Login from './Login.svelte';
  import CustomModels from './CustomModels.svelte';
  import ModelPacks from './ModelPacks.svelte';
  import FileManager from './FileManager.svelte';
  import RealtimePreview from './RealtimePreview.svelte';
  import TerminalComponent from './Terminal.svelte';
  import CodeEditorView from './CodeEditorView.svelte';
  import OrgSelector from './lib/OrgSelector.svelte'; // Will create this

  let appReady = false;

  onMount(async () => {
    initializeAuth(); // Initialize from localStorage
    const currentAuth = $authStore;

    if (currentAuth.isAuthenticated && currentAuth.token) {
      try {
        // Validate session with the backend
        const sessionData = await api.getSession();
        if (sessionData && sessionData.UserId) {
          // Potentially refresh authStore with fresh sessionData if needed,
          // for now, initializeAuth + this check confirms validity.
          // If orgs list changed, authStore.login(sessionData) could update it.
          if (currentAuth.orgId && !sessionData.Orgs.find(o => o.Id === currentAuth.orgId)) {
            // Current orgId is invalid, select a new one or clear it
            const newOrgId = sessionData.Orgs.length > 0 ? sessionData.Orgs[0].Id : null;
            selectOrg(newOrgId);
          }
        } else {
          // Session invalid or unexpected response
          processLogout(); // Clear auth state
          navigate('/login', { replace: true });
        }
      } catch (error) {
        console.error('Session validation failed:', error);
        processLogout();
        if (error.status !== 401) { // 401 already handled by api.js to redirect
            // navigate('/login', { replace: true }); 
            // api.js will handle redirect for 401, for other errors we might stay or redirect
        }
      }
    } else if (!currentAuth.isAuthenticated && window.location.pathname !== '/login') {
        // If not authenticated and not already on login page, redirect.
        // This handles the case where there's nothing in localStorage.
        navigate('/login', { replace: true });
    }
    appReady = true;
  });

  async function handleLogout() {
    try {
      await api.logout(); // Call backend logout
    } catch (error) {
      console.error('Logout API call failed:', error);
      // Still proceed with client-side logout even if API fails
    } finally {
      processLogout(); // Clear client-side auth state & localStorage
      navigate('/login', { replace: true });
    }
  }

  // Reactive statement to redirect if auth state changes to unauthenticated
  // and we are not on the login page.
  $: if (appReady && !$authStore.isAuthenticated && window.location.pathname !== '/login') {
    navigate('/login', { replace: true });
  }

</script>

{#if !appReady}
  <div>Loading application...</div>
{:else}
  <Router>
    {#if $authStore.isAuthenticated}
      <div class="app-layout">
        <nav class="main-nav">
          <a href="/custom-models" use:link class:active={location.pathname === '/custom-models' || location.pathname === '/'}>Custom Models</a>
          <a href="/model-packs" use:link class:active={location.pathname === '/model-packs'}>Model Packs</a>
          <a href="/files" use:link class:active={location.pathname === '/files'}>File Manager</a>
          <a href="/preview" use:link class:active={location.pathname === '/preview'}>Real-time Preview</a>
          <a href="/terminal" use:link class:active={location.pathname === '/terminal'}>Terminal</a>
          <a href="/code-interact" use:link class:active={location.pathname === '/code-interact'}>Code Interaction</a>
          <div class="nav-actions">
            <OrgSelector />
            <button on:click={handleLogout}>Logout ({$authStore.currentUser && $authStore.currentUser.email})</button>
          </div>
        </nav>
        <main class="content">
          <Route path="/login" component={Login} /> {/* Should redirect if auth */}
          <Route path="/custom-models" component={CustomModels} />
          <Route path="/model-packs" component={ModelPacks} />
          <Route path="/files" component={FileManager} />
          <Route path="/preview" component={RealtimePreview} />
          <Route path="/terminal" component={TerminalComponent} />
          <Route path="/code-interact" component={CodeEditorView} />
          <Route path="/" component={CustomModels} /> {/* Default route */}
        </main>
      </div>
    {:else}
      <!-- Public part of the app, only Login for now -->
      <Route path="/login" component={Login} />
      {/* Catch-all to redirect to login if not authenticated and no other route matches,
          though the reactive statement and onMount should handle most cases.
          This helps if trying to access a non-login route directly while unauthenticated.
      */}
      <Route path="*">
        <svelte:component this="{() => { if (appReady && window.location.pathname !== '/login') navigate('/login', { replace: true }); return null; }}" />
      </Route>
    {/if}
  </Router>
{/if}

<style>
  .app-layout {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }
  .main-nav {
    display: flex;
    align-items: center; /* Vertically center items */
    padding: 0.5rem 1rem;
    background-color: #333;
    color: white;
  }
  .main-nav a {
    color: white;
    text-decoration: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    transition: background-color 0.2s;
  }
  .main-nav a:hover, .main-nav a.active {
    background-color: #555;
  }
  .nav-actions {
    margin-left: auto; /* Pushes org selector and logout button to the right */
    display: flex;
    align-items: center;
    gap: 15px;
  }
  .nav-actions button {
    background-color: #61afef;
    color: #282c34;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
  }
  .nav-actions button:hover {
    background-color: #5295cf;
  }
  .content {
    flex-grow: 1;
    padding: 1rem;
    background-color: #f4f4f4; /* Default page background */
  }

  /* Reset global styles if needed, or ensure App.svelte is the root */
  :global(body) {
    font-family: Arial, sans-serif;
    margin: 0;
    padding: 0;
    /* background-color: #f4f4f4; */ /* Moved to .content for better control */
  }
  :global(main.content > div) { /* Target the div rendered by Svelte Router's Route */
    background-color: white;
    padding: 20px;
    border-radius: 5px;
    box-shadow: 0 0 10px rgba(0,0,0,0.1);
  }
</style>
</main>

<style>
  :global(body) {
    font-family: Arial, sans-serif;
    margin: 0;
    padding: 0;
    background-color: #f4f4f4;
  }

  main {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1em;
  }

  nav {
    display: flex;
    justify-content: center;
    margin-bottom: 20px;
    background-color: #333;
    padding: 10px 0;
    border-radius: 5px;
  }

  nav button {
    background-color: #555;
    color: white;
    border: none;
    padding: 10px 20px;
    margin: 0 10px;
    cursor: pointer;
    border-radius: 4px;
    font-size: 16px;
    transition: background-color 0.3s ease;
  }

  nav button.active {
    background-color: #ff3e00; /* Svelte orange */
  }

  nav button:hover {
    background-color: #777;
  }

  nav button.active:hover {
    background-color: #ff3e00; 
  }

  .content {
    background-color: white;
    padding: 20px;
    border-radius: 5px;
    box-shadow: 0 0 10px rgba(0,0,0,0.1);
  }
</style>
