<script>
  import { onMount } from 'svelte';

  let modelPacks = [];
  let customModels = []; // To populate model dropdowns
  let showModal = false;
  let currentPack = null;
  let isEditing = false;

  // Form fields for ModelPack
  let packId = ''; // User-defined ID for the pack
  let packName = '';
  let packDescription = '';

  // For ModelRoleConfig fields (simplified for this form)
  // We'll need to allow selecting from available custom models
  let plannerModelId = '';
  let coderModelId = '';
  let planSummaryModelId = '';
  let builderModelId = '';
  let wholeFileBuilderModelId = '';
  let namerModelId = '';
  let commitMsgModelId = '';
  let execStatusModelId = '';
  let architectModelId = ''; // contextLoader

  // Default values for a ModelRoleConfig (can be expanded)
  const defaultRoleConfigSettings = {
    temperature: 0.7,
    topP: 1.0,
    reservedOutputTokens: 512, // A sensible default
    reasoningEffort: "medium",
  };

  // Default values for PlannerModelConfig
  const defaultPlannerSettings = {
    maxConvoTokens: 8000, // A sensible default
  };

  async function fetchModelPacks() {
    try {
      const response = await fetch('/api/ui/model-packs');
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      modelPacks = await response.json();
      if (!modelPacks) modelPacks = [];
    } catch (error) {
      console.error("Failed to fetch model packs:", error);
      modelPacks = [];
    }
  }

  async function fetchCustomModelsForSelect() {
    try {
      const response = await fetch('/api/ui/custom-models');
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      customModels = await response.json();
      if (!customModels) customModels = [];
    } catch (error) {
      console.error("Failed to fetch custom models for select:", error);
      customModels = [];
    }
  }

  onMount(async () => {
    await fetchModelPacks();
    await fetchCustomModelsForSelect();
  });

  function findBaseModelConfig(modelIdToFind) {
    if (!customModels || customModels.length === 0 || !modelIdToFind) return null;
    const foundModel = customModels.find(m => m.modelId === modelIdToFind);
    return foundModel ? foundModel.baseModelConfig : null;
  }

  function populateFormFromPack(pack) {
    packId = pack.id || ''; // This is the user-defined ID of the pack
    packName = pack.name || '';
    packDescription = pack.description || '';

    plannerModelId = pack.planner?.baseModelConfig?.modelId || '';
    coderModelId = pack.coder?.baseModelConfig?.modelId || '';
    planSummaryModelId = pack.planSummary?.baseModelConfig?.modelId || '';
    builderModelId = pack.builder?.baseModelConfig?.modelId || '';
    wholeFileBuilderModelId = pack.wholeFileBuilder?.baseModelConfig?.modelId || '';
    namerModelId = pack.namer?.baseModelConfig?.modelId || '';
    commitMsgModelId = pack.commitMsg?.baseModelConfig?.modelId || '';
    execStatusModelId = pack.execStatus?.baseModelConfig?.modelId || '';
    architectModelId = pack.contextLoader?.baseModelConfig?.modelId || ''; // Ensure correct field name
  }

  function openAddModal() {
    isEditing = false;
    currentPack = null;
    // Reset form fields
    packId = '';
    packName = '';
    packDescription = '';
    plannerModelId = '';
    coderModelId = '';
    planSummaryModelId = '';
    builderModelId = '';
    wholeFileBuilderModelId = '';
    namerModelId = '';
    commitMsgModelId = '';
    execStatusModelId = '';
    architectModelId = '';
    showModal = true;
  }

  function openEditModal(pack) {
    isEditing = true;
    currentPack = pack;
    populateFormFromPack(pack);
    showModal = true;
  }

  function closeModal() {
    showModal = false;
    currentPack = null;
  }

  function createModelRoleConfig(selectedModelId) {
    const baseConfig = findBaseModelConfig(selectedModelId);
    if (!baseConfig) return null; // Or handle error appropriately

    return {
      baseModelConfig: baseConfig,
      ...defaultRoleConfigSettings, // Apply other defaults
      // Fallbacks would be null by default or configured further
      largeContextFallback: null,
      largeOutputFallback: null,
      errorFallback: null,
      missingKeyFallback: null,
      strongModel: null,
    };
  }
  
  function createPlannerRoleConfig(selectedModelId) {
    const modelRoleConfig = createModelRoleConfig(selectedModelId);
    if (!modelRoleConfig) return null;

    return {
      modelRoleConfig: modelRoleConfig,
      plannerModelConfig: { // Specific planner config
        ...defaultPlannerSettings, // Apply planner defaults
      },
    };
  }


  async function handleSubmit() {
    // Construct the ModelPack payload based on shared.ModelPack
    // Each role (planner, coder, etc.) is a ModelRoleConfig or PlannerRoleConfig
    // which embeds a BaseModelConfig.
    // The form collects model IDs (e.g., 'gpt-4-turbo') for each role.
    // We need to find the corresponding BaseModelConfig from `customModels`.

    const plannerConfig = createPlannerRoleConfig(plannerModelId);
    if (!plannerConfig) {
        alert("Planner model configuration is missing or invalid.");
        return;
    }
    
    // Optional roles: Coder, WholeFileBuilder, Architect
    // Required roles: PlanSummary, Builder, Namer, CommitMsg, ExecStatus
    // If an optional role's modelId is not selected, we might omit it or use a default.
    // For now, if a model ID is not selected for an optional role, we'll set the role to null.

    const getOptionalRoleConfig = (modelId) => modelId ? createModelRoleConfig(modelId) : null;
    const getRequiredRoleConfig = (modelId, roleName) => {
        const config = createModelRoleConfig(modelId);
        if (!config) {
            alert(`${roleName} model configuration is missing or invalid.`);
            throw new Error(`${roleName} configuration error`); // Stop submission
        }
        return config;
    };
    
    let payload;
    try {
        payload = {
            id: packId, // User-defined ID for the pack
            name: packName,
            description: packDescription,
            planner: plannerConfig.modelRoleConfig, //  shared.ModelPack.Planner is PlannerRoleConfig
            coder: getOptionalRoleConfig(coderModelId),
            planSummary: getRequiredRoleConfig(planSummaryModelId, "Plan Summary"),
            builder: getRequiredRoleConfig(builderModelId, "Builder"),
            wholeFileBuilder: getOptionalRoleConfig(wholeFileBuilderModelId),
            namer: getRequiredRoleConfig(namerModelId, "Namer"),
            commitMsg: getRequiredRoleConfig(commitMsgModelId, "Commit Message"),
            execStatus: getRequiredRoleConfig(execStatusModelId, "Execution Status"),
            contextLoader: getOptionalRoleConfig(architectModelId), // 'contextLoader' is 'Architect' in shared.ModelPack
        };
        // Adjust planner to be PlannerRoleConfig
        payload.planner = {
            modelRoleConfig: plannerConfig.modelRoleConfig,
            plannerModelConfig: plannerConfig.plannerModelConfig
        }

    } catch (error) {
        // Error already alerted by getRequiredRoleConfig
        return;
    }


    try {
      let response;
      if (isEditing && currentPack && currentPack.id) {
         // The API expects the user-defined pack ID in the URL.
         // The payload for PUT should be the complete updated ModelPack object.
        response = await fetch(`/api/ui/model-packs/${currentPack.id}`, { // currentPack.id is the user-defined one
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
      } else {
        response = await fetch('/api/ui/model-packs', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
      }

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
      }
      fetchModelPacks(); // Refresh list
      closeModal();
    } catch (error) {
      console.error("Failed to save model pack:", error);
      alert(`Error saving model pack: ${error.message}`);
    }
  }

  async function deletePack(packToDelete) {
    if (!confirm(`Are you sure you want to delete model pack: ${packToDelete.name || packToDelete.id}?`)) {
      return;
    }
    try {
      // The API expects the user-defined pack ID in the URL.
      const idToDelete = packToDelete.id; // User-defined packId
      const response = await fetch(`/api/ui/model-packs/${idToDelete}`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
      }
      fetchModelPacks(); // Refresh list
    } catch (error) {
      console.error("Failed to delete model pack:", error);
      alert(`Error deleting model pack: ${error.message}`);
    }
  }

  function getModelDisplayName(modelId) {
    const model = customModels.find(m => m.modelId === modelId);
    return model ? `${model.modelName} (${model.modelId})` : modelId || 'N/A';
  }

</script>

<style>
  .container { padding: 20px; }
  .modal {
    position: fixed; top: 0; left: 0; width: 100%; height: 100%;
    background: rgba(0,0,0,0.5); display: flex;
    align-items: center; justify-content: center; z-index: 1000;
  }
  .modal-content { background: white; padding: 20px; border-radius: 5px; width: 700px; max-height: 90vh; overflow-y: auto;}
  .form-group { margin-bottom: 15px; }
  label { display: block; margin-bottom: 5px; font-weight: bold; }
  input[type="text"], input[type="number"], select, textarea { width: 100%; padding: 8px; box-sizing: border-box; border: 1px solid #ccc; border-radius: 4px;}
  table { width: 100%; border-collapse: collapse; margin-top: 20px; }
  th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
  th { background-color: #f2f2f2; }
  .actions button { margin-right: 5px; }
  .role-configs { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-top:15px; }
  .role-group { border: 1px solid #eee; padding: 10px; border-radius: 4px; }
  .role-group h4 { margin-top: 0; }
</style>

<div class="container">
  <h2>Model Packs</h2>
  <button on:click={openAddModal}>Add New Model Pack</button>

  {#if !modelPacks || modelPacks.length === 0}
    <p>No model packs found.</p>
  {:else}
    <table>
      <thead>
        <tr>
          <th>Pack ID (User Defined)</th>
          <th>Name</th>
          <th>Description</th>
          <th>Planner Model</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each modelPacks as pack}
          <tr>
            <td>{pack.id}</td>
            <td>{pack.name}</td>
            <td>{pack.description}</td>
            <td>{getModelDisplayName(pack.planner?.modelRoleConfig?.baseModelConfig?.modelId)}</td>
            <td class="actions">
              <button on:click={() => openEditModal(pack)}>Edit</button>
              <button on:click={() => deletePack(pack)}>Delete</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}

  {#if showModal}
    <div class="modal">
      <div class="modal-content">
        <h3>{isEditing ? 'Edit' : 'Add'} Model Pack</h3>
        <form on:submit|preventDefault={handleSubmit}>
          <div class="form-group">
            <label for="packId">Pack ID (unique identifier for this pack):</label>
            <input type="text" id="packId" bind:value={packId} required disabled={isEditing}>
             {#if isEditing}
              <small>Pack ID cannot be changed after creation.</small>
            {/if}
          </div>
          <div class="form-group">
            <label for="packName">Pack Name (e.g., Default Development Pack):</label>
            <input type="text" id="packName" bind:value={packName} required>
          </div>
          <div class="form-group">
            <label for="packDescription">Description:</label>
            <textarea id="packDescription" bind:value={packDescription}></textarea>
          </div>

          <h4>Role Model Assignments</h4>
          <p><small>Select from your available custom models. Ensure models are suitable for their roles.</small></p>
          
          <div class="role-configs">
            <div class="role-group">
              <h4>Core Roles</h4>
              <div class="form-group">
                <label for="plannerModelId">Planner Model:</label>
                <select id="plannerModelId" bind:value={plannerModelId} required>
                  <option value="">Select Model</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
              </div>
              <div class="form-group">
                <label for="planSummaryModelId">Plan Summary Model:</label>
                <select id="planSummaryModelId" bind:value={planSummaryModelId} required>
                  <option value="">Select Model</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
              </div>
              <div class="form-group">
                <label for="builderModelId">Builder Model:</label>
                <select id="builderModelId" bind:value={builderModelId} required>
                  <option value="">Select Model</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
              </div>
               <div class="form-group">
                <label for="namerModelId">Namer Model:</label>
                 <select id="namerModelId" bind:value={namerModelId} required>
                  <option value="">Select Model</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
              </div>
              <div class="form-group">
                <label for="commitMsgModelId">Commit Message Model:</label>
                <select id="commitMsgModelId" bind:value={commitMsgModelId} required>
                  <option value="">Select Model</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
              </div>
               <div class="form-group">
                <label for="execStatusModelId">Execution Status Model:</label>
                <select id="execStatusModelId" bind:value={execStatusModelId} required>
                  <option value="">Select Model</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
              </div>
            </div>

            <div class="role-group">
              <h4>Optional Roles (leave blank to use defaults or disable)</h4>
              <div class="form-group">
                <label for="coderModelId">Coder Model:</label>
                <select id="coderModelId" bind:value={coderModelId}>
                  <option value="">Select Model (Optional)</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
                 <small>If not set, Planner model may be used.</small>
              </div>
              <div class="form-group">
                <label for="wholeFileBuilderModelId">Whole File Builder Model:</label>
                <select id="wholeFileBuilderModelId" bind:value={wholeFileBuilderModelId}>
                  <option value="">Select Model (Optional)</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
                <small>If not set, Builder model is used.</small>
              </div>
              <div class="form-group">
                <label for="architectModelId">Architect (Context Loader) Model:</label>
                <select id="architectModelId" bind:value={architectModelId}>
                  <option value="">Select Model (Optional)</option>
                  {#each customModels as model} <option value={model.modelId}>{model.modelName} ({model.modelId})</option> {/each}
                </select>
                <small>If not set, Planner model may be used.</small>
              </div>
            </div>
          </div>
          
          <button type="submit" style="margin-top: 15px;">{isEditing ? 'Save Changes' : 'Create Pack'}</button>
          <button type="button" on:click={closeModal} style="margin-left: 10px;">Cancel</button>
        </form>
      </div>
    </div>
  {/if}
</div>
