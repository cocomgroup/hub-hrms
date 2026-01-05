<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  interface Provider {
    id: string;
    name: string;
    type: 'linkedin' | 'indeed' | 'ziprecruiter' | 'glassdoor' | 'monster' | 'custom';
    status: 'connected' | 'disconnected' | 'error';
    api_key?: string;
    config: any;
    jobs_posted: number;
    applicants_received: number;
    last_sync: string;
  }
  
  let providers: Provider[] = [];
  let loading = true;
  let showAddProvider = false;
  let editingProvider: Provider | null = null;
  
  const providerTemplates = {
    linkedin: {
      name: 'LinkedIn',
      icon: '💼',
      color: '#0077b5',
      description: 'Professional networking platform',
      fields: [
        { name: 'client_id', label: 'Client ID', type: 'text', placeholder: 'Enter LinkedIn Client ID' },
        { name: 'client_secret', label: 'Client Secret', type: 'password', placeholder: 'Enter Client Secret' },
        { name: 'company_id', label: 'Company ID', type: 'text', placeholder: 'Your LinkedIn Company ID' }
      ]
    },
    indeed: {
      name: 'Indeed',
      icon: '🔍',
      color: '#2164f3',
      description: 'World\'s #1 job site',
      fields: [
        { name: 'publisher_id', label: 'Publisher ID', type: 'text', placeholder: 'Your Indeed Publisher ID' },
        { name: 'api_token', label: 'API Token', type: 'password', placeholder: 'Indeed API Token' }
      ]
    },
    ziprecruiter: {
      name: 'ZipRecruiter',
      icon: '⚡',
      color: '#1ca774',
      description: 'Smart matching technology',
      fields: [
        { name: 'api_key', label: 'API Key', type: 'password', placeholder: 'ZipRecruiter API Key' },
        { name: 'company_id', label: 'Company ID', type: 'text', placeholder: 'Your Company ID' }
      ]
    },
    glassdoor: {
      name: 'Glassdoor',
      icon: '🏢',
      color: '#0caa41',
      description: 'Company reviews & salaries',
      fields: [
        { name: 'partner_id', label: 'Partner ID', type: 'text', placeholder: 'Glassdoor Partner ID' },
        { name: 'partner_key', label: 'Partner Key', type: 'password', placeholder: 'Partner Key' }
      ]
    },
    monster: {
      name: 'Monster',
      icon: '👹',
      color: '#6f42c1',
      description: 'Find better candidates faster',
      fields: [
        { name: 'api_key', label: 'API Key', type: 'password', placeholder: 'Monster API Key' },
        { name: 'account_id', label: 'Account ID', type: 'text', placeholder: 'Your Account ID' }
      ]
    },
    custom: {
      name: 'Custom Provider',
      icon: '🔧',
      color: '#6b7280',
      description: 'Configure custom integration',
      fields: [
        { name: 'api_url', label: 'API URL', type: 'text', placeholder: 'https://api.example.com' },
        { name: 'api_key', label: 'API Key', type: 'password', placeholder: 'Your API Key' },
        { name: 'webhook_url', label: 'Webhook URL', type: 'text', placeholder: 'Webhook endpoint (optional)' }
      ]
    }
  };
  
  onMount(async () => {
    await loadProviders();
  });
  
  async function loadProviders() {
    try {
      loading = true;
      const token = localStorage.getItem('token');
      const response = await fetch('/api/recruiting/providers', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        providers = await response.json();
      }
    } catch (err) {
      console.error('Failed to load providers:', err);
    } finally {
      loading = false;
    }
  }
  
  function startConnectProvider(type: string) {
    const template = providerTemplates[type];
    editingProvider = {
      id: '',
      name: template.name,
      type: type as any,
      status: 'disconnected',
      config: {},
      jobs_posted: 0,
      applicants_received: 0,
      last_sync: new Date().toISOString()
    };
    
    // Initialize config fields
    template.fields.forEach(field => {
      editingProvider.config[field.name] = '';
    });
    
    showAddProvider = true;
  }
  
  function editProvider(provider: Provider) {
    editingProvider = { ...provider };
    showAddProvider = true;
  }
  
  async function saveProvider() {
    try {
      const token = localStorage.getItem('token');
      const method = editingProvider?.id ? 'PUT' : 'POST';
      const url = editingProvider?.id 
        ? `/api/recruiting/providers/${editingProvider.id}`
        : '/api/recruiting/providers';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(editingProvider)
      });
      
      if (response.ok) {
        await loadProviders();
        showAddProvider = false;
        editingProvider = null;
        dispatch('providerUpdated');
      } else {
        const error = await response.json();
        alert('Failed to save provider: ' + (error.message || 'Unknown error'));
      }
    } catch (err) {
      console.error('Failed to save provider:', err);
      alert('Failed to save provider: ' + err.message);
    }
  }
  
  async function testConnection(provider: Provider) {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/recruiting/providers/${provider.id}/test`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        const result = await response.json();
        if (result.success) {
          alert('✓ Connection successful!\n\n' + (result.message || 'Provider is properly configured.'));
        } else {
          alert('✗ Connection failed!\n\n' + (result.error || 'Unknown error'));
        }
        await loadProviders();
      }
    } catch (err) {
      alert('✗ Connection test failed!\n\n' + err.message);
    }
  }
  
  async function disconnectProvider(provider: Provider) {
    if (!confirm(`Are you sure you want to disconnect ${provider.name}?\n\nThis will remove all configuration and stop syncing jobs and applicants.`)) {
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      await fetch(`/api/recruiting/providers/${provider.id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      await loadProviders();
      dispatch('providerUpdated');
    } catch (err) {
      console.error('Failed to disconnect provider:', err);
      alert('Failed to disconnect provider: ' + err.message);
    }
  }
  
  function isProviderConnected(type: string): boolean {
    return providers.some(p => p.type === type);
  }
</script>

<div class="providers-container">
  <!-- Header -->
  <div class="header-section">
    <div>
      <h2>Recruiting Providers</h2>
      <p class="subtitle">Connect job boards and recruiting platforms to post jobs and manage applicants</p>
    </div>
  </div>
  
  <!-- Available Providers -->
  <div class="section-card">
    <h3>Available Providers</h3>
    <p class="section-description">Click on a provider to connect and configure</p>
    
    <div class="providers-grid">
      {#each Object.entries(providerTemplates) as [type, template]}
        {@const connected = providers.find(p => p.type === type)}
        <button 
          class="provider-card"
          class:connected={connected}
          on:click={() => !connected && startConnectProvider(type)}
          disabled={!!connected}
        >
          <div class="provider-icon" style="background-color: {template.color}20; color: {template.color}">
            {template.icon}
          </div>
          <div class="provider-info">
            <div class="provider-name">{template.name}</div>
            <div class="provider-description">{template.description}</div>
            {#if connected}
              <div class="provider-status connected">
                ✓ Connected
              </div>
            {:else}
              <div class="provider-status">
                Click to Connect →
              </div>
            {/if}
          </div>
        </button>
      {/each}
    </div>
  </div>
  
  <!-- Connected Providers -->
  {#if providers.length > 0}
    <div class="section-card">
      <h3>Connected Providers ({providers.length})</h3>
      <p class="section-description">Manage your active integrations</p>
      
      <div class="connected-list">
        {#each providers as provider}
          {@const template = providerTemplates[provider.type]}
          <div class="connected-provider">
            <div class="provider-main">
              <div class="provider-icon large" style="background-color: {template.color}20; color: {template.color}">
                {template.icon}
              </div>
              <div class="provider-details">
                <div class="provider-name">{provider.name}</div>
                <div class="provider-stats">
                  <span class="stat-item">
                    <span class="stat-icon">📋</span>
                    {provider.jobs_posted} jobs posted
                  </span>
                  <span class="stat-divider">•</span>
                  <span class="stat-item">
                    <span class="stat-icon">👥</span>
                    {provider.applicants_received} applicants
                  </span>
                  <span class="stat-divider">•</span>
                  <span class="status-indicator status-{provider.status}">
                    {provider.status === 'connected' ? '✓ Connected' : 
                     provider.status === 'error' ? '✗ Error' : 
                     '○ Disconnected'}
                  </span>
                </div>
                {#if provider.last_sync}
                  <div class="last-sync">
                    Last synced: {new Date(provider.last_sync).toLocaleString()}
                  </div>
                {/if}
              </div>
            </div>
            <div class="provider-actions">
              <button class="action-btn secondary" on:click={() => testConnection(provider)}>
                <span class="btn-icon">🔌</span>
                Test
              </button>
              <button class="action-btn secondary" on:click={() => editProvider(provider)}>
                <span class="btn-icon">⚙️</span>
                Settings
              </button>
              <button class="action-btn danger" on:click={() => disconnectProvider(provider)}>
                <span class="btn-icon">✗</span>
                Disconnect
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {:else}
    <div class="section-card empty">
      <div class="empty-state">
        <span class="empty-icon">🔗</span>
        <h3>No Providers Connected</h3>
        <p>Connect a recruiting provider to start posting jobs and receiving applicants</p>
      </div>
    </div>
  {/if}
</div>

<!-- Add/Edit Provider Modal -->
{#if showAddProvider && editingProvider}
  {@const template = providerTemplates[editingProvider.type]}
  <div class="modal-overlay" on:click={() => showAddProvider = false}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{editingProvider.id ? 'Edit' : 'Connect'} {editingProvider.name}</h2>
          <p class="modal-subtitle">{template.description}</p>
        </div>
        <button class="close-btn" on:click={() => showAddProvider = false}>×</button>
      </div>
      
      <div class="modal-body">
        
        <div class="form-section">
          <div class="form-header">
            <div class="provider-icon modal-icon" style="background-color: {template.color}20; color: {template.color}">
              {template.icon}
            </div>
            <div>
              <h4>Configuration</h4>
              <p class="form-description">Enter your API credentials to connect {template.name}</p>
            </div>
          </div>
          
          {#each template.fields as field}
            <div class="form-group">
              <label for={field.name}>{field.label}</label>
              <input
                id={field.name}
                type={field.type}
                bind:value={editingProvider.config[field.name]}
                placeholder={field.placeholder}
                required
              />
              {#if field.name === 'api_url'}
                <div class="field-hint">Enter the base URL for API requests</div>
              {:else if field.name === 'webhook_url'}
                <div class="field-hint">Optional: URL to receive applicant notifications</div>
              {/if}
            </div>
          {/each}
        </div>
        
        <div class="modal-actions">
          <button class="btn secondary" on:click={() => showAddProvider = false}>
            Cancel
          </button>
          <button class="btn primary" on:click={saveProvider}>
            <span class="btn-icon">💾</span>
            {editingProvider.id ? 'Save Changes' : 'Save & Connect'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .providers-container {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .header-section {
    display: flex;
    justify-content: space-between;
    align-items: start;
  }
  
  .header-section h2 {
    font-size: 24px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .subtitle {
    font-size: 14px;
    color: #6b7280;
    margin: 0;
  }
  
  .section-card {
    background: white;
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }
  
  .section-card.empty {
    background: linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%);
  }
  
  .section-card h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .section-description {
    font-size: 13px;
    color: #6b7280;
    margin: 0 0 20px 0;
  }
  
  .providers-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }
  
  .provider-card {
    display: flex;
    align-items: start;
    gap: 16px;
    padding: 20px;
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
    width: 100%;
  }
  
  .provider-card:not(.connected):hover {
    border-color: #3b82f6;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  }
  
  .provider-card.connected {
    border-color: #10b981;
    background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
    cursor: not-allowed;
  }
  
  .provider-icon {
    font-size: 32px;
    min-width: 56px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
  }
  
  .provider-icon.large {
    min-width: 64px;
    height: 64px;
    font-size: 36px;
  }
  
  .provider-icon.modal-icon {
    min-width: 48px;
    height: 48px;
    font-size: 28px;
  }
  
  .provider-info {
    flex: 1;
  }
  
  .provider-name {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin-bottom: 4px;
  }
  
  .provider-description {
    font-size: 13px;
    color: #6b7280;
    margin-bottom: 8px;
  }
  
  .provider-status {
    font-size: 13px;
    font-weight: 500;
    color: #6b7280;
  }
  
  .provider-status.connected {
    color: #10b981;
  }
  
  .connected-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  
  .connected-provider {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    transition: all 0.2s;
    gap: 20px;
  }
  
  .connected-provider:hover {
    border-color: #d1d5db;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  }
  
  .provider-main {
    display: flex;
    align-items: center;
    gap: 16px;
    flex: 1;
  }
  
  .provider-details {
    flex: 1;
  }
  
  .provider-stats {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    font-size: 13px;
    color: #6b7280;
    margin-top: 8px;
  }
  
  .stat-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  
  .stat-icon {
    font-size: 14px;
  }
  
  .stat-divider {
    color: #d1d5db;
  }
  
  .status-indicator {
    padding: 3px 10px;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 500;
  }
  
  .status-indicator.status-connected {
    background: #d1fae5;
    color: #065f46;
  }
  
  .status-indicator.status-error {
    background: #fee2e2;
    color: #991b1b;
  }
  
  .status-indicator.status-disconnected {
    background: #f3f4f6;
    color: #6b7280;
  }
  
  .last-sync {
    font-size: 12px;
    color: #9ca3af;
    margin-top: 4px;
  }
  
  .provider-actions {
    display: flex;
    gap: 8px;
  }
  
  .action-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
  }
  
  .action-btn.secondary {
    background: white;
    color: #374151;
  }
  
  .action-btn.secondary:hover {
    border-color: #3b82f6;
    color: #3b82f6;
    background: #eff6ff;
  }
  
  .action-btn.danger {
    background: white;
    color: #dc2626;
    border-color: #fecaca;
  }
  
  .action-btn.danger:hover {
    background: #dc2626;
    color: white;
    border-color: #dc2626;
  }
  
  .btn-icon {
    font-size: 14px;
  }
  
  .empty-state {
    text-align: center;
    padding: 60px 20px;
  }
  
  .empty-icon {
    font-size: 64px;
    display: block;
    margin-bottom: 16px;
  }
  
  .empty-state h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .empty-state p {
    color: #6b7280;
    margin: 0;
  }
  
  /* Modal styles */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
  }
  
  .modal {
    background: white;
    border-radius: 12px;
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .modal-header h2 {
    font-size: 20px;
    font-weight: 600;
    margin: 0 0 4px 0;
    color: #111827;
  }
  
  .modal-subtitle {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .close-btn {
    background: none;
    border: none;
    font-size: 28px;
    line-height: 1;
    cursor: pointer;
    color: #6b7280;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    transition: all 0.2s;
  }
  
  .close-btn:hover {
    background: #f3f4f6;
    color: #111827;
  }
  
  .modal-body {
    padding: 24px;
  }
  
  .form-section {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }
  
  .form-header {
    display: flex;
    align-items: start;
    gap: 16px;
    padding-bottom: 20px;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .form-header h4 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .form-description {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  
  .form-group label {
    font-size: 13px;
    font-weight: 500;
    color: #374151;
  }
  
  .form-group input {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    transition: all 0.2s;
  }
  
  .form-group input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }
  
  .field-hint {
    font-size: 12px;
    color: #6b7280;
    font-style: italic;
  }
  
  .modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid #e5e7eb;
  }
  
  .btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn.primary {
    background: #3b82f6;
    color: white;
  }
  
  .btn.primary:hover {
    background: #2563eb;
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  }
  
  .btn.secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }
  
  .btn.secondary:hover {
    background: #f9fafb;
    border-color: #9ca3af;
  }
  
  @media (max-width: 768px) {
    .providers-grid {
      grid-template-columns: 1fr;
    }
    
    .connected-provider {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;
    }
    
    .provider-main {
      width: 100%;
    }
    
    .provider-actions {
      width: 100%;
      flex-wrap: wrap;
    }
    
    .action-btn {
      flex: 1;
      justify-content: center;
    }
  }
</style>
