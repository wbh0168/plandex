<script>
  import { authStore, selectOrg } from './authStore';
  import { onMount } from 'svelte';

  let orgs = [];
  let selectedOrgId = '';

  // Subscribe to authStore to get organizations and current selection
  authStore.subscribe(value => {
    orgs = value.orgs || [];
    selectedOrgId = value.orgId || '';
  });

  function handleOrgChange(event) {
    const newOrgId = event.target.value;
    if (newOrgId && newOrgId !== selectedOrgId) {
      selectOrg(newOrgId); // This will update the store and localStorage
      // Optionally, trigger a page reload or data refresh if needed after org change
      // window.location.reload(); // Or more specific data fetching
    }
  }
</script>

<style>
  .org-selector-container {
    display: inline-block; /* Or flex, depending on layout needs */
  }
  select {
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    border: 1px solid #4f5865; /* Matches other controls */
    background-color: #21252b; /* Matches other controls */
    color: #abb2bf; /* Matches other controls */
    font-family: inherit;
    font-size: 14px;
    min-width: 150px; /* Give it some space */
  }
  select:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
</style>

<div class="org-selector-container">
  {#if orgs.length > 0}
    <select bind:value={selectedOrgId} on:change={handleOrgChange} disabled={orgs.length <= 1}>
      {#each orgs as org}
        <option value={org.Id}>{org.Name}</option>
      {/each}
    </select>
  {:else if $authStore.isAuthenticated}
    <span>No organizations available.</span>
  {/if}
  <!-- If not authenticated, this component might not be shown, or show nothing -->
</div>
