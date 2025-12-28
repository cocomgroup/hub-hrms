<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  
  interface PTOBalance {
    id: string;
    employee_id: string;
    vacation_days: number;
    sick_days: number;
    personal_days: number;
    year: number;
    created_at: string;
    updated_at: string;
  }

  interface PTORequest {
    id: string;
    employee_id: string;
    employee_name?: string;
    department?: string;
    pto_type: string;
    start_date: string;
    end_date: string;
    days_requested: number;
    reason: string;
    status: string;
    reviewed_by?: string;
    reviewer_name?: string;
    reviewed_at?: string;
    review_notes?: string;
    created_at: string;
    updated_at: string;
  }

  let balance: PTOBalance | null = null;
  let requests: PTORequest[] = [];
  let loading = false;
  let error = '';
  let success = '';
  
  // New request form
  let showRequestForm = false;
  let newRequest = {
    pto_type: 'vacation',
    start_date: '',
    end_date: '',
    days_requested: 0,
    reason: ''
  };

  // Request detail modal
  let selectedRequest: PTORequest | null = null;
  let showRequestDetail = false;

  // Filter
  let statusFilter: string = 'all';

  onMount(() => {
    loadData();
  });

  async function loadData() {
    loading = true;
    error = '';
    
    try {
      await Promise.all([
        loadBalance(),
        loadRequests()
      ]);
    } catch (err: any) {
      error = err.message || 'Failed to load data';
    } finally {
      loading = false;
    }
  }

  async function loadBalance() {
    const response = await fetch('/api/pto/balance', {
      headers: {
        'Authorization': `Bearer ${$authStore.token}`
      }
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to load PTO balance');
    }
    
    balance = await response.json();
  }

  async function loadRequests() {
    const response = await fetch('/api/pto/requests', {
      headers: {
        'Authorization': `Bearer ${$authStore.token}`
      }
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to load PTO requests');
    }
    
    requests = await response.json() || [];
  }

  async function submitRequest() {
    loading = true;
    error = '';
    success = '';
    
    // Validate
    if (!newRequest.start_date || !newRequest.end_date) {
      error = 'Please select start and end dates';
      loading = false;
      return;
    }

    const startDate = new Date(newRequest.start_date);
    const endDate = new Date(newRequest.end_date);

    if (endDate < startDate) {
      error = 'End date must be after start date';
      loading = false;
      return;
    }

    if (newRequest.days_requested <= 0) {
      error = 'Days requested must be greater than 0';
      loading = false;
      return;
    }

    // Check balance
    if (balance) {
      let available = 0;
      switch (newRequest.pto_type) {
        case 'vacation':
          available = balance.vacation_days;
          break;
        case 'sick':
          available = balance.sick_days;
          break;
        case 'personal':
          available = balance.personal_days;
          break;
      }

      if (newRequest.days_requested > available) {
        error = `Insufficient balance. You have ${available.toFixed(1)} ${newRequest.pto_type} days available.`;
        loading = false;
        return;
      }
    }

    try {
      const response = await fetch('/api/pto/requests', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(newRequest)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to submit request');
      }

      success = 'PTO request submitted successfully!';
      setTimeout(() => success = '', 3000);
      
      // Reset form
      resetForm();
      
      // Reload data
      await loadData();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function resetForm() {
    showRequestForm = false;
    newRequest = {
      pto_type: 'vacation',
      start_date: '',
      end_date: '',
      days_requested: 0,
      reason: ''
    };
  }

  function calculateBusinessDays() {
    if (!newRequest.start_date || !newRequest.end_date) return;
    
    const start = new Date(newRequest.start_date);
    const end = new Date(newRequest.end_date);
    
    if (end < start) {
      error = 'End date must be after start date';
      newRequest.days_requested = 0;
      return;
    }

    let count = 0;
    const current = new Date(start);

    while (current <= end) {
      const day = current.getDay();
      // Skip weekends (0 = Sunday, 6 = Saturday)
      if (day !== 0 && day !== 6) {
        count++;
      }
      current.setDate(current.getDate() + 1);
    }

    newRequest.days_requested = count;
    error = ''; // Clear any previous errors
  }

  function viewRequestDetail(request: PTORequest) {
    selectedRequest = request;
    showRequestDetail = true;
  }

  function closeRequestDetail() {
    selectedRequest = null;
    showRequestDetail = false;
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function formatDateTime(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function getStatusColor(status: string): string {
    switch (status) {
      case 'approved':
        return 'status-approved';
      case 'denied':
        return 'status-denied';
      case 'pending':
        return 'status-pending';
      case 'cancelled':
        return 'status-cancelled';
      default:
        return '';
    }
  }

  function getTypeLabel(type: string): string {
    return type.charAt(0).toUpperCase() + type.slice(1);
  }

  function getTypeIcon(type: string): string {
    switch (type) {
      case 'vacation':
        return '🏖️';
      case 'sick':
        return '🏥';
      case 'personal':
        return '⏰';
      default:
        return '📅';
    }
  }

  $: totalAvailable = balance ? 
    balance.vacation_days + balance.sick_days + balance.personal_days : 0;

  $: filteredRequests = statusFilter === 'all' 
    ? requests 
    : requests.filter(r => r.status === statusFilter);

  $: pendingCount = requests.filter(r => r.status === 'pending').length;
  $: approvedCount = requests.filter(r => r.status === 'approved').length;
</script>

<div class="pto-container">
  <!-- Header -->
  <div class="header">
    <div>
      <h1>Time Off Management</h1>
      <p class="subtitle">Manage your paid time off requests and balance</p>
    </div>
    <button class="btn btn-primary" on:click={() => showRequestForm = true}>
      <span class="btn-icon">+</span> Request Time Off
    </button>
  </div>

  <!-- Messages -->
  {#if error}
    <div class="alert alert-error">
      <span class="alert-icon">⚠️</span>
      {error}
    </div>
  {/if}
  {#if success}
    <div class="alert alert-success">
      <span class="alert-icon">✓</span>
      {success}
    </div>
  {/if}

  {#if loading && !balance}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading your PTO information...</p>
    </div>
  {:else}
    <!-- Balance Cards -->
    {#if balance}
      <div class="balance-section">
        <h2>Your PTO Balance ({balance.year})</h2>
        <div class="balance-cards">
          <div class="balance-card vacation">
            <div class="card-icon">🏖️</div>
            <div class="card-content">
              <h3>Vacation Days</h3>
              <div class="balance-amount">{balance.vacation_days.toFixed(1)}</div>
              <div class="card-label">days available</div>
            </div>
          </div>

          <div class="balance-card sick">
            <div class="card-icon">🏥</div>
            <div class="card-content">
              <h3>Sick Days</h3>
              <div class="balance-amount">{balance.sick_days.toFixed(1)}</div>
              <div class="card-label">days available</div>
            </div>
          </div>

          <div class="balance-card personal">
            <div class="card-icon">⏰</div>
            <div class="card-content">
              <h3>Personal Days</h3>
              <div class="balance-amount">{balance.personal_days.toFixed(1)}</div>
              <div class="card-label">days available</div>
            </div>
          </div>

          <div class="balance-card total">
            <div class="card-icon">📊</div>
            <div class="card-content">
              <h3>Total Available</h3>
              <div class="balance-amount">{totalAvailable.toFixed(1)}</div>
              <div class="card-label">days total</div>
            </div>
          </div>
        </div>
      </div>
    {/if}

    <!-- Request Form Modal -->
    {#if showRequestForm}
      <div class="modal-overlay" on:click={resetForm} on:keydown={(e) => e.key === 'Escape' && resetForm()} role="button" tabindex="0">
        <div class="modal" on:click|stopPropagation on:keydown|stopPropagation role="dialog" aria-modal="true" tabindex="-1">
          <div class="modal-header">
            <h2>Request Time Off</h2>
            <button class="close-btn" on:click={resetForm}>×</button>
          </div>

          <form on:submit|preventDefault={submitRequest}>
            <div class="modal-body">
              <div class="form-group">
                <label for="pto-type">Time Off Type *</label>
                <select id="pto-type" bind:value={newRequest.pto_type} required>
                  <option value="vacation">🏖️ Vacation</option>
                  <option value="sick">🏥 Sick Leave</option>
                  <option value="personal">⏰ Personal Day</option>
                </select>
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label for="start-date">Start Date *</label>
                  <input 
                    id="start-date"
                    type="date" 
                    bind:value={newRequest.start_date}
                    on:change={calculateBusinessDays}
                    required
                  />
                </div>

                <div class="form-group">
                  <label for="end-date">End Date *</label>
                  <input 
                    id="end-date"
                    type="date" 
                    bind:value={newRequest.end_date}
                    on:change={calculateBusinessDays}
                    required
                  />
                </div>
              </div>

              <div class="form-group">
                <label for="days-requested">Days Requested *</label>
                <input 
                  id="days-requested"
                  type="number" 
                  bind:value={newRequest.days_requested}
                  step="0.5"
                  min="0.5"
                  required
                />
                <small>Business days calculated automatically (excluding weekends). You can adjust if needed.</small>
              </div>

              <div class="form-group">
                <label for="reason">Reason (Optional)</label>
                <textarea 
                  id="reason"
                  bind:value={newRequest.reason}
                  rows="3"
                  placeholder="Briefly explain your reason for this request..."
                ></textarea>
              </div>

              {#if balance}
                <div class="balance-check">
                  <div class="balance-check-header">
                    <strong>Available Balance</strong>
                  </div>
                  <div class="balance-check-content">
                    {#if newRequest.pto_type === 'vacation'}
                      <span class="balance-value">{balance.vacation_days.toFixed(1)}</span> vacation days
                    {:else if newRequest.pto_type === 'sick'}
                      <span class="balance-value">{balance.sick_days.toFixed(1)}</span> sick days
                    {:else}
                      <span class="balance-value">{balance.personal_days.toFixed(1)}</span> personal days
                    {/if}
                  </div>
                </div>
              {/if}
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" on:click={resetForm}>Cancel</button>
              <button 
                type="submit"
                class="btn btn-primary" 
                disabled={loading}
              >
                {loading ? 'Submitting...' : 'Submit Request'}
              </button>
            </div>
          </form>
        </div>
      </div>
    {/if}

    <!-- Request Detail Modal -->
    {#if showRequestDetail && selectedRequest}
      <div class="modal-overlay" on:click={closeRequestDetail} on:keydown={(e) => e.key === 'Escape' && closeRequestDetail()} role="button" tabindex="0">
        <div class="modal modal-detail" on:click|stopPropagation on:keydown|stopPropagation role="dialog" aria-modal="true" tabindex="-1">
          <div class="modal-header">
            <h2>Request Details</h2>
            <button class="close-btn" on:click={closeRequestDetail}>×</button>
          </div>

          <div class="modal-body">
            <div class="detail-grid">
              <div class="detail-item">
                <span class="detail-label">Type</span>
                <div class="detail-value">
                  <span class="type-badge type-{selectedRequest.pto_type}">
                    {getTypeIcon(selectedRequest.pto_type)} {getTypeLabel(selectedRequest.pto_type)}
                  </span>
                </div>
              </div>

              <div class="detail-item">
                <span class="detail-label">Status</span>
                <div class="detail-value">
                  <span class="status-badge {getStatusColor(selectedRequest.status)}">
                    {selectedRequest.status}
                  </span>
                </div>
              </div>

              <div class="detail-item">
                <span class="detail-label">Start Date</span>
                <div class="detail-value">{formatDate(selectedRequest.start_date)}</div>
              </div>

              <div class="detail-item">
                <span class="detail-label">End Date</span>
                <div class="detail-value">{formatDate(selectedRequest.end_date)}</div>
              </div>

              <div class="detail-item">
                <span class="detail-label">Days Requested</span>
                <div class="detail-value"><strong>{selectedRequest.days_requested}</strong></div>
              </div>

              <div class="detail-item">
                <span class="detail-label">Submitted</span>
                <div class="detail-value">{formatDateTime(selectedRequest.created_at)}</div>
              </div>

              {#if selectedRequest.reason}
                <div class="detail-item full-width">
                  <span class="detail-label">Reason</span>
                  <div class="detail-value reason">{selectedRequest.reason}</div>
                </div>
              {/if}

              {#if selectedRequest.reviewed_by}
                <div class="detail-item full-width review-section">
                  <span class="detail-label">Review Information</span>
                  <div class="review-info">
                    <div class="review-item">
                      <span class="review-label">Reviewed by:</span>
                      <span class="review-value">{selectedRequest.reviewer_name || 'Manager'}</span>
                    </div>
                    <div class="review-item">
                      <span class="review-label">Reviewed at:</span>
                      <span class="review-value">{formatDateTime(selectedRequest.reviewed_at || '')}</span>
                    </div>
                    {#if selectedRequest.review_notes}
                      <div class="review-item">
                        <span class="review-label">Notes:</span>
                        <span class="review-value">{selectedRequest.review_notes}</span>
                      </div>
                    {/if}
                  </div>
                </div>
              {/if}
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" on:click={closeRequestDetail}>Close</button>
          </div>
        </div>
      </div>
    {/if}

    <!-- Requests History -->
    <div class="requests-section">
      <div class="section-header">
        <div>
          <h2>Your Time Off Requests</h2>
          <div class="request-stats">
            <span class="stat">
              <span class="stat-icon">📋</span>
              {requests.length} total
            </span>
            <span class="stat">
              <span class="stat-icon">⏳</span>
              {pendingCount} pending
            </span>
            <span class="stat">
              <span class="stat-icon">✓</span>
              {approvedCount} approved
            </span>
          </div>
        </div>
        <div class="filter-group">
          <label for="status-filter">Filter:</label>
          <select id="status-filter" bind:value={statusFilter}>
            <option value="all">All Requests</option>
            <option value="pending">Pending</option>
            <option value="approved">Approved</option>
            <option value="denied">Denied</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>
      </div>
      
      {#if filteredRequests.length === 0}
        <div class="empty-state">
          <div class="empty-icon">📅</div>
          <h3>
            {statusFilter === 'all' 
              ? 'No Time Off Requests' 
              : `No ${statusFilter} requests`}
          </h3>
          <p>
            {statusFilter === 'all'
              ? "You haven't submitted any time off requests yet."
              : `You don't have any ${statusFilter} requests.`}
          </p>
          {#if statusFilter === 'all'}
            <button class="btn btn-primary" on:click={() => showRequestForm = true}>
              Create Your First Request
            </button>
          {:else}
            <button class="btn btn-secondary" on:click={() => statusFilter = 'all'}>
              View All Requests
            </button>
          {/if}
        </div>
      {:else}
        <div class="requests-table">
          <table>
            <thead>
              <tr>
                <th>Type</th>
                <th>Start Date</th>
                <th>End Date</th>
                <th>Days</th>
                <th>Status</th>
                <th>Submitted</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {#each filteredRequests as request}
                <tr on:click={() => viewRequestDetail(request)} class="clickable-row">
                  <td>
                    <span class="type-badge type-{request.pto_type}">
                      {getTypeIcon(request.pto_type)} {getTypeLabel(request.pto_type)}
                    </span>
                  </td>
                  <td>{formatDate(request.start_date)}</td>
                  <td>{formatDate(request.end_date)}</td>
                  <td><strong>{request.days_requested}</strong></td>
                  <td>
                    <span class="status-badge {getStatusColor(request.status)}">
                      {request.status}
                    </span>
                  </td>
                  <td>{formatDate(request.created_at)}</td>
                  <td>
                    <div class="action-icons">
                      {#if request.reason}
                        <button class="btn-icon" title="Has reason" on:click|stopPropagation>📝</button>
                      {/if}
                      {#if request.review_notes}
                        <button class="btn-icon" title="Has review notes" on:click|stopPropagation>💬</button>
                      {/if}
                      <button class="btn-icon" title="View details" on:click|stopPropagation={() => viewRequestDetail(request)}>👁️</button>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .pto-container {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
    min-height: 100vh;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    gap: 2rem;
  }

  .header h1 {
    font-size: 2rem;
    color: #e4e7eb;
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: #999;
    margin: 0;
    font-size: 0.95rem;
  }

  .alert {
    padding: 1rem 1.25rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    animation: slideDown 0.3s ease;
  }

  @keyframes slideDown {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .alert-icon {
    font-size: 1.25rem;
  }

  .alert-error {
    background-color: rgba(244, 67, 54, 0.1);
    color: #f44336;
    border: 1px solid rgba(244, 67, 54, 0.3);
  }

  .alert-success {
    background-color: rgba(76, 175, 80, 0.1);
    color: #4caf50;
    border: 1px solid rgba(76, 175, 80, 0.3);
  }

  .loading {
    text-align: center;
    padding: 4rem 2rem;
    color: #999;
  }

  .spinner {
    border: 4px solid #2d3139;
    border-top: 4px solid #667eea;
    border-radius: 50%;
    width: 48px;
    height: 48px;
    animation: spin 1s linear infinite;
    margin: 0 auto 1rem;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* Balance Section */
  .balance-section {
    margin-bottom: 3rem;
  }

  .balance-section h2 {
    font-size: 1.5rem;
    color: #e4e7eb;
    margin-bottom: 1.5rem;
  }

  .balance-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 1.5rem;
  }

  .balance-card {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 16px;
    padding: 1.75rem;
    color: white;
    display: flex;
    gap: 1.25rem;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    transition: all 0.3s ease;
    position: relative;
    overflow: hidden;
  }

  .balance-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0) 100%);
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  .balance-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 20px rgba(0, 0, 0, 0.2);
  }

  .balance-card:hover::before {
    opacity: 1;
  }

  .balance-card.vacation {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  }

  .balance-card.sick {
    background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  }

  .balance-card.personal {
    background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  }

  .balance-card.total {
    background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  }

  .card-icon {
    font-size: 3.5rem;
    line-height: 1;
    filter: drop-shadow(0 2px 4px rgba(0,0,0,0.2));
  }

  .card-content {
    flex: 1;
    position: relative;
    z-index: 1;
  }

  .card-content h3 {
    margin: 0 0 0.75rem 0;
    font-size: 1rem;
    font-weight: 600;
    opacity: 0.95;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .balance-amount {
    font-size: 2.75rem;
    font-weight: bold;
    line-height: 1;
    margin-bottom: 0.5rem;
  }

  .card-label {
    font-size: 0.875rem;
    opacity: 0.85;
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .modal {
    background: #1e2128;
    border-radius: 16px;
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
    animation: slideUp 0.3s ease;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-detail {
    max-width: 700px;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem 2rem;
    border-bottom: 1px solid #2d3139;
  }

  .modal-header h2 {
    margin: 0;
    color: #e4e7eb;
    font-size: 1.5rem;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 2rem;
    color: #999;
    cursor: pointer;
    line-height: 1;
    padding: 0;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    transition: all 0.2s;
  }

  .close-btn:hover {
    color: #fff;
    background: rgba(255, 255, 255, 0.1);
  }

  .modal-body {
    padding: 2rem;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding: 1.5rem 2rem;
    border-top: 1px solid #2d3139;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: #e4e7eb;
    font-size: 0.95rem;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    width: 100%;
    padding: 0.75rem 1rem;
    background: #2d3139;
    border: 2px solid #3d4149;
    border-radius: 8px;
    color: #e4e7eb;
    font-size: 1rem;
    transition: border-color 0.2s;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #667eea;
  }

  .form-group small {
    display: block;
    margin-top: 0.5rem;
    color: #999;
    font-size: 0.875rem;
    line-height: 1.4;
  }

  .form-group textarea {
    resize: vertical;
    font-family: inherit;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .balance-check {
    background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
    padding: 1.25rem;
    border-radius: 8px;
    border-left: 4px solid #667eea;
    margin-top: 1rem;
  }

  .balance-check-header {
    color: #999;
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
  }

  .balance-check-content {
    color: #e4e7eb;
    font-size: 1.1rem;
  }

  .balance-value {
    font-size: 1.5rem;
    font-weight: bold;
    color: #667eea;
  }

  /* Detail Grid */
  .detail-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1.5rem;
  }

  .detail-item {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .detail-item.full-width {
    grid-column: 1 / -1;
  }

  .detail-item label {
    color: #999;
    font-size: 0.875rem;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .detail-label {
    color: #999;
    font-size: 0.875rem;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    display: block;
  }

  .detail-value {
    color: #e4e7eb;
    font-size: 1rem;
  }

  .detail-value.reason {
    background: #2d3139;
    padding: 1rem;
    border-radius: 8px;
    line-height: 1.6;
  }

  .review-section {
    background: #2d3139;
    padding: 1.5rem;
    border-radius: 8px;
  }

  .review-info {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-top: 1rem;
  }

  .review-item {
    display: flex;
    gap: 0.75rem;
  }

  .review-label {
    color: #999;
    font-weight: 500;
    min-width: 120px;
  }

  .review-value {
    color: #e4e7eb;
    flex: 1;
  }

  /* Requests Section */
  .requests-section {
    margin-top: 3rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
    gap: 2rem;
  }

  .section-header h2 {
    font-size: 1.5rem;
    color: #e4e7eb;
    margin: 0 0 0.75rem 0;
  }

  .request-stats {
    display: flex;
    gap: 1.5rem;
    flex-wrap: wrap;
  }

  .stat {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: #999;
    font-size: 0.9rem;
  }

  .stat-icon {
    font-size: 1.1rem;
  }

  .filter-group {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .filter-group label {
    color: #999;
    font-size: 0.9rem;
    font-weight: 500;
  }

  .filter-group select {
    padding: 0.5rem 1rem;
    background: #2d3139;
    border: 2px solid #3d4149;
    border-radius: 8px;
    color: #e4e7eb;
    font-size: 0.9rem;
    cursor: pointer;
  }

  .requests-table {
    background: #1e2128;
    border-radius: 16px;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th {
    background: #2d3139;
    color: #999;
    font-weight: 600;
    text-align: left;
    padding: 1rem 1.25rem;
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  td {
    padding: 1rem 1.25rem;
    border-top: 1px solid #2d3139;
    color: #e4e7eb;
  }

  .clickable-row {
    cursor: pointer;
    transition: background-color 0.2s;
  }

  tbody tr:hover {
    background: #2d3139;
  }

  .type-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.35rem 0.85rem;
    border-radius: 12px;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .type-vacation {
    background: rgba(102, 126, 234, 0.15);
    color: #667eea;
  }

  .type-sick {
    background: rgba(245, 87, 108, 0.15);
    color: #f5576c;
  }

  .type-personal {
    background: rgba(79, 172, 254, 0.15);
    color: #4facfe;
  }

  .status-badge {
    display: inline-block;
    padding: 0.35rem 0.85rem;
    border-radius: 12px;
    font-size: 0.875rem;
    font-weight: 500;
    text-transform: capitalize;
  }

  .status-pending {
    background: rgba(255, 193, 7, 0.15);
    color: #ffc107;
  }

  .status-approved {
    background: rgba(76, 175, 80, 0.15);
    color: #4caf50;
  }

  .status-denied {
    background: rgba(244, 67, 54, 0.15);
    color: #f44336;
  }

  .status-cancelled {
    background: rgba(158, 158, 158, 0.15);
    color: #9e9e9e;
  }

  .action-icons {
    display: flex;
    gap: 0.5rem;
  }

  .btn-icon {
    background: none;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    opacity: 0.6;
    padding: 0.25rem;
    transition: opacity 0.2s, transform 0.2s;
  }

  .btn-icon:hover {
    opacity: 1;
    transform: scale(1.1);
  }

  .empty-state {
    text-align: center;
    padding: 4rem 2rem;
    color: #999;
    background: #1e2128;
    border-radius: 16px;
  }

  .empty-icon {
    font-size: 5rem;
    margin-bottom: 1.5rem;
    opacity: 0.5;
  }

  .empty-state h3 {
    color: #e4e7eb;
    margin-bottom: 0.75rem;
    font-size: 1.5rem;
  }

  .empty-state p {
    margin-bottom: 2rem;
    font-size: 1rem;
  }

  /* Buttons */
  .btn {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 16px rgba(102, 126, 234, 0.4);
  }

  .btn-secondary {
    background: #2d3139;
    color: #e4e7eb;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #3d4149;
  }

  @media (max-width: 968px) {
    .pto-container {
      padding: 1rem;
    }

    .header {
      flex-direction: column;
      align-items: stretch;
    }

    .balance-cards {
      grid-template-columns: 1fr;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .section-header {
      flex-direction: column;
      align-items: stretch;
    }

    .filter-group {
      justify-content: space-between;
    }

    .requests-table {
      overflow-x: auto;
    }

    table {
      min-width: 700px;
    }

    .detail-grid {
      grid-template-columns: 1fr;
    }

    .modal {
      max-width: 95%;
    }
  }
</style>