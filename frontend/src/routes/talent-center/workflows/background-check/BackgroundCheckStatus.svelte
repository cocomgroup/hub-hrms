<script lang="ts">
  import { onMount } from 'svelte';
  import { fade, slide } from 'svelte/transition';
  
  export let employeeId: string;

  interface BackgroundCheck {
    id: string;
    employee_id: string;
    package_id: string;
    status: string;
    result?: string;
    check_types: string[];
    initiated_at: string;
    completed_at?: string;
    estimated_eta?: string;
    report_url?: string;
    initiated_by: string;
  }

  let checks: BackgroundCheck[] = [];
  let loading = true;
  let error = '';
  let selectedCheck: BackgroundCheck | null = null;
  let refreshInterval: number;

  onMount(() => {
    loadChecks();
    // Refresh every 30 seconds for pending/in-progress checks
    refreshInterval = setInterval(() => {
      if (checks.some(c => c.status === 'pending' || c.status === 'in_progress')) {
        loadChecks();
      }
    }, 30000);

    return () => {
      if (refreshInterval) clearInterval(refreshInterval);
    };
  });

  async function loadChecks() {
    try {
      const response = await fetch(`/api/v1/employees/${employeeId}/background-checks`);
      if (!response.ok) throw new Error('Failed to load background checks');
      checks = await response.json();
      checks.sort((a, b) => 
        new Date(b.initiated_at).getTime() - new Date(a.initiated_at).getTime()
      );
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function viewDetails(check: BackgroundCheck) {
    try {
      const response = await fetch(`/api/v1/background-checks/${check.id}`);
      if (!response.ok) throw new Error('Failed to load check details');
      selectedCheck = await response.json();
    } catch (err) {
      error = err.message;
    }
  }

  async function cancelCheck(checkId: string) {
    if (!confirm('Are you sure you want to cancel this background check?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/background-checks/${checkId}/cancel`, {
        method: 'POST'
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to cancel check');
      }
      
      await loadChecks();
      selectedCheck = null;
    } catch (err) {
      error = err.message;
    }
  }

  function closeDetails() {
    selectedCheck = null;
  }

  function getStatusColor(status: string): string {
    switch (status) {
      case 'completed': return 'green';
      case 'in_progress': return 'blue';
      case 'pending': return 'orange';
      case 'failed': return 'red';
      case 'cancelled': return 'gray';
      default: return 'gray';
    }
  }

  function getResultColor(result: string): string {
    switch (result) {
      case 'clear': return 'green';
      case 'consider': return 'orange';
      case 'suspended': return 'red';
      case 'dispute': return 'purple';
      default: return 'gray';
    }
  }

  function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function formatCheckType(type: string): string {
    return type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
  }

  function getStatusIcon(status: string): string {
    switch (status) {
      case 'completed': return '✓';
      case 'in_progress': return '⟳';
      case 'pending': return '⏱';
      case 'failed': return '✗';
      case 'cancelled': return '⊘';
      default: return '•';
    }
  }
</script>

<div class="background-checks-container">
  <div class="header">
    <h2>Background Checks</h2>
  </div>

  {#if error}
    <div class="error-message" transition:slide>
      {error}
      <button class="close-error" on:click={() => error = ''}>×</button>
    </div>
  {/if}

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Loading background checks...</p>
    </div>
  {:else if checks.length === 0}
    <div class="empty-state">
      <div class="empty-icon">📋</div>
      <h3>No Background Checks</h3>
      <p>No background checks have been initiated for this employee yet.</p>
    </div>
  {:else}
    <div class="checks-list">
      {#each checks as check (check.id)}
        <div class="check-card" transition:fade>
          <div class="check-header">
            <div class="check-status">
              <span 
                class="status-badge" 
                class:status-completed={check.status === 'completed'}
                class:status-in-progress={check.status === 'in_progress'}
                class:status-pending={check.status === 'pending'}
                class:status-failed={check.status === 'failed'}
                class:status-cancelled={check.status === 'cancelled'}
              >
                <span class="status-icon">{getStatusIcon(check.status)}</span>
                {check.status.replace(/_/g, ' ')}
              </span>
              
              {#if check.result}
                <span 
                  class="result-badge"
                  class:result-clear={check.result === 'clear'}
                  class:result-consider={check.result === 'consider'}
                  class:result-suspended={check.result === 'suspended'}
                  class:result-dispute={check.result === 'dispute'}
                >
                  {check.result}
                </span>
              {/if}
            </div>

            <div class="check-actions">
              <button class="btn-icon" on:click={() => viewDetails(check)} title="View Details">
                👁️
              </button>
              
              {#if check.report_url}
                <a 
                  href={check.report_url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  class="btn-icon"
                  title="View Report"
                >
                  📄
                </a>
              {/if}

              {#if check.status === 'pending' || check.status === 'in_progress'}
                <button 
                  class="btn-icon btn-cancel" 
                  on:click={() => cancelCheck(check.id)}
                  title="Cancel Check"
                >
                  ⊘
                </button>
              {/if}
            </div>
          </div>

          <div class="check-content">
            <div class="check-types">
              {#each check.check_types as type}
                <span class="check-type-tag">{formatCheckType(type)}</span>
              {/each}
            </div>

            <div class="check-meta">
              <div class="meta-item">
                <span class="meta-label">Initiated:</span>
                <span class="meta-value">{formatDate(check.initiated_at)}</span>
              </div>

              {#if check.completed_at}
                <div class="meta-item">
                  <span class="meta-label">Completed:</span>
                  <span class="meta-value">{formatDate(check.completed_at)}</span>
                </div>
              {:else if check.estimated_eta}
                <div class="meta-item">
                  <span class="meta-label">Estimated ETA:</span>
                  <span class="meta-value">{formatDate(check.estimated_eta)}</span>
                </div>
              {/if}

              <div class="meta-item">
                <span class="meta-label">Check ID:</span>
                <span class="meta-value mono">{check.id.substring(0, 12)}...</span>
              </div>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Details Modal -->
{#if selectedCheck}
  <div class="modal-overlay" on:click={closeDetails} transition:fade>
    <div class="modal-content" on:click|stopPropagation transition:slide>
      <div class="modal-header">
        <h3>Background Check Details</h3>
        <button class="modal-close" on:click={closeDetails}>×</button>
      </div>

      <div class="modal-body">
        <div class="detail-section">
          <h4>Status Information</h4>
          <div class="detail-grid">
            <div class="detail-item">
              <span class="detail-label">Status:</span>
              <span class="detail-value">
                <span 
                  class="status-badge" 
                  style="background: var(--{getStatusColor(selectedCheck.status)})"
                >
                  {selectedCheck.status.replace(/_/g, ' ')}
                </span>
              </span>
            </div>

            {#if selectedCheck.result}
              <div class="detail-item">
                <span class="detail-label">Result:</span>
                <span class="detail-value">
                  <span 
                    class="result-badge"
                    style="background: var(--{getResultColor(selectedCheck.result)})"
                  >
                    {selectedCheck.result}
                  </span>
                </span>
              </div>
            {/if}

            <div class="detail-item">
              <span class="detail-label">Check ID:</span>
              <span class="detail-value mono">{selectedCheck.id}</span>
            </div>

            <div class="detail-item">
              <span class="detail-label">Initiated By:</span>
              <span class="detail-value">{selectedCheck.initiated_by}</span>
            </div>

            <div class="detail-item">
              <span class="detail-label">Initiated At:</span>
              <span class="detail-value">{formatDate(selectedCheck.initiated_at)}</span>
            </div>

            {#if selectedCheck.completed_at}
              <div class="detail-item">
                <span class="detail-label">Completed At:</span>
                <span class="detail-value">{formatDate(selectedCheck.completed_at)}</span>
              </div>
            {/if}
          </div>
        </div>

        <div class="detail-section">
          <h4>Check Types</h4>
          <div class="check-types-list">
            {#each selectedCheck.check_types as type}
              <div class="check-type-item">
                <span class="check-type-icon">✓</span>
                <span class="check-type-name">{formatCheckType(type)}</span>
              </div>
            {/each}
          </div>
        </div>

        {#if selectedCheck.report_url}
          <div class="detail-section">
            <h4>Report</h4>
            <a 
              href={selectedCheck.report_url} 
              target="_blank" 
              rel="noopener noreferrer"
              class="btn btn-primary full-width"
            >
              View Full Report
            </a>
          </div>
        {/if}

        {#if selectedCheck.status === 'pending' || selectedCheck.status === 'in_progress'}
          <div class="detail-section">
            <button 
              class="btn btn-danger full-width" 
              on:click={() => cancelCheck(selectedCheck.id)}
            >
              Cancel Background Check
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .background-checks-container {
    padding: 1.5rem;
  }

  .header {
    margin-bottom: 2rem;
  }

  .header h2 {
    margin: 0;
    color: #1a1a1a;
  }

  .error-message {
    background: #fee;
    color: #c33;
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .close-error {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: #c33;
  }

  .loading-state {
    text-align: center;
    padding: 4rem 2rem;
  }

  .spinner {
    width: 50px;
    height: 50px;
    margin: 0 auto 1rem;
    border: 4px solid #f3f4f6;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .empty-state {
    text-align: center;
    padding: 4rem 2rem;
  }

  .empty-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
  }

  .empty-state h3 {
    color: #1a1a1a;
    margin-bottom: 0.5rem;
  }

  .empty-state p {
    color: #666;
  }

  .checks-list {
    display: grid;
    gap: 1.5rem;
  }

  .check-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    padding: 1.5rem;
    transition: all 0.2s;
  }

  .check-card:hover {
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    border-color: #d1d5db;
  }

  .check-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid #f3f4f6;
  }

  .check-status {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 0.75rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    text-transform: capitalize;
  }

  .status-completed { background: #d1fae5; color: #065f46; }
  .status-in-progress { background: #dbeafe; color: #1e40af; }
  .status-pending { background: #fef3c7; color: #92400e; }
  .status-failed { background: #fee2e2; color: #991b1b; }
  .status-cancelled { background: #f3f4f6; color: #4b5563; }

  .result-badge {
    padding: 0.375rem 0.75rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    text-transform: uppercase;
  }

  .result-clear { background: #d1fae5; color: #065f46; }
  .result-consider { background: #fef3c7; color: #92400e; }
  .result-suspended { background: #fee2e2; color: #991b1b; }
  .result-dispute { background: #e9d5ff; color: #6b21a8; }

  .check-actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn-icon {
    background: none;
    border: 1px solid #e5e7eb;
    padding: 0.5rem 0.75rem;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: all 0.2s;
    font-size: 1.125rem;
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }

  .btn-icon:hover {
    background: #f9fafb;
    border-color: #d1d5db;
  }

  .btn-cancel:hover {
    background: #fee2e2;
    border-color: #fca5a5;
  }

  .check-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .check-types {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .check-type-tag {
    background: #f3f4f6;
    color: #4b5563;
    padding: 0.25rem 0.75rem;
    border-radius: 0.25rem;
    font-size: 0.875rem;
  }

  .check-meta {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  .meta-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .meta-label {
    font-size: 0.75rem;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .meta-value {
    font-size: 0.875rem;
    color: #1f2937;
  }

  .mono {
    font-family: 'Courier New', monospace;
  }

  /* Modal Styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal-content {
    background: white;
    border-radius: 0.5rem;
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h3 {
    margin: 0;
    color: #1a1a1a;
  }

  .modal-close {
    background: none;
    border: none;
    font-size: 2rem;
    line-height: 1;
    cursor: pointer;
    color: #6b7280;
    transition: color 0.2s;
  }

  .modal-close:hover {
    color: #1f2937;
  }

  .modal-body {
    padding: 1.5rem;
  }

  .detail-section {
    margin-bottom: 2rem;
  }

  .detail-section:last-child {
    margin-bottom: 0;
  }

  .detail-section h4 {
    margin: 0 0 1rem 0;
    color: #1a1a1a;
    font-size: 1rem;
  }

  .detail-grid {
    display: grid;
    gap: 1rem;
  }

  .detail-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    background: #f9fafb;
    border-radius: 0.375rem;
  }

  .detail-label {
    font-weight: 500;
    color: #6b7280;
  }

  .detail-value {
    color: #1f2937;
  }

  .check-types-list {
    display: grid;
    gap: 0.5rem;
  }

  .check-type-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    background: #f9fafb;
    border-radius: 0.375rem;
  }

  .check-type-icon {
    color: #10b981;
    font-weight: bold;
  }

  .check-type-name {
    color: #1f2937;
  }

  .btn {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 0.375rem;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    text-decoration: none;
    display: inline-block;
    text-align: center;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-primary:hover {
    background: #2563eb;
  }

  .btn-danger {
    background: #ef4444;
    color: white;
  }

  .btn-danger:hover {
    background: #dc2626;
  }

  .full-width {
    width: 100%;
  }
</style>
