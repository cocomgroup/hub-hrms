<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface Props {
    onAction?: () => void;
  }

  let { onAction }: Props = $props();

  interface Timesheet {
    id: string;
    employee_id: string;
    employee_name: string;
    start_date: string;
    end_date: string;
    status: string;
    total_hours: number;
    regular_hours: number;
    overtime_hours?: number;
    submitted_at: string;
    time_entries?: TimeEntry[];
  }

  interface TimeEntry {
    id: string;
    date: string;
    hours: number;
    project_id?: string;
    project_name?: string;
    type: string;
    notes?: string;
  }

  let loading = $state(false);
  let processing = $state(false);
  let timesheets = $state<Timesheet[]>([]);
  let selectedTimesheet = $state<Timesheet | null>(null);
  let showModal = $state(false);
  let modalMode = $state<'approve' | 'reject'>('approve');
  let rejectionReason = $state('');
  let successMessage = $state('');
  let errorMessage = $state('');
  let expandedTimesheet = $state<string | null>(null);
  let loadingDetails = $state<string | null>(null);

  onMount(() => {
    loadPendingTimesheets();
  });

  async function loadPendingTimesheets() {
    loading = true;
    errorMessage = '';
    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/pending`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const data = await response.json();
        timesheets = Array.isArray(data) ? data : [];
        console.log('Loaded timesheets:', timesheets);
      } else {
        console.error('Failed to load timesheets:', response.status);
        timesheets = [];
      }
    } catch (err) {
      console.error('Error loading timesheets:', err);
      errorMessage = 'Failed to load pending timesheets';
      timesheets = [];
    } finally {
      loading = false;
    }
  }

  async function loadTimesheetDetails(timesheetId: string) {
    loadingDetails = timesheetId;
    try {
      console.log('Loading details for timesheet:', timesheetId);
      const response = await fetch(`${API_BASE_URL}/timesheet/timesheets/${timesheetId}`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const data = await response.json();
        console.log('Timesheet details:', data);
        
        // Update the timesheet with full details
        const index = timesheets.findIndex(t => t.id === timesheetId);
        if (index !== -1) {
          timesheets[index] = { 
            ...timesheets[index], 
            time_entries: data.time_entries || []
          };
          timesheets = [...timesheets];
        }
      } else {
        console.error('Failed to load timesheet details:', response.status);
        errorMessage = 'Failed to load timesheet details';
      }
    } catch (err) {
      console.error('Error loading timesheet details:', err);
      errorMessage = 'Error loading timesheet details';
    } finally {
      loadingDetails = null;
    }
  }

  function openApproveModal(timesheet: Timesheet) {
    selectedTimesheet = timesheet;
    modalMode = 'approve';
    rejectionReason = '';
    showModal = true;
    errorMessage = '';
  }

  function openRejectModal(timesheet: Timesheet) {
    selectedTimesheet = timesheet;
    modalMode = 'reject';
    rejectionReason = '';
    showModal = true;
    errorMessage = '';
  }

  function closeModal() {
    showModal = false;
    selectedTimesheet = null;
    rejectionReason = '';
    errorMessage = '';
  }

  async function handleApprove() {
    if (!selectedTimesheet) return;

    processing = true;
    errorMessage = '';
    successMessage = '';

    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/timesheets/${selectedTimesheet.id}/approve`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          action: 'approve'
        })
      });

      if (response.ok) {
        successMessage = `Timesheet approved for ${selectedTimesheet.employee_name}`;
        closeModal();
        await loadPendingTimesheets();
        onAction?.();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to approve timesheet';
      }
    } catch (err) {
      console.error('Error approving timesheet:', err);
      errorMessage = 'An error occurred while approving the timesheet';
    } finally {
      processing = false;
    }
  }

  async function handleReject() {
    if (!selectedTimesheet) return;
    
    if (!rejectionReason.trim()) {
      errorMessage = 'Please provide a reason for rejection';
      return;
    }

    processing = true;
    errorMessage = '';
    successMessage = '';

    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/timesheets/${selectedTimesheet.id}/approve`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          action: 'reject',
          rejection_reason: rejectionReason
        })
      });

      if (response.ok) {
        successMessage = `Timesheet rejected for ${selectedTimesheet.employee_name}`;
        closeModal();
        await loadPendingTimesheets();
        onAction?.();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to reject timesheet';
      }
    } catch (err) {
      console.error('Error rejecting timesheet:', err);
      errorMessage = 'An error occurred while rejecting the timesheet';
    } finally {
      processing = false;
    }
  }

  function toggleTimesheetDetails(timesheetId: string) {
    if (expandedTimesheet === timesheetId) {
      expandedTimesheet = null;
    } else {
      expandedTimesheet = timesheetId;
      const timesheet = timesheets.find(t => t.id === timesheetId);
      if (timesheet && !timesheet.time_entries) {
        loadTimesheetDetails(timesheetId);
      }
    }
  }

  // ✅ FIXED: Better date validation and formatting
  function formatDate(dateString: string | null | undefined): string {
    if (!dateString) return 'N/A';
    
    try {
      // Handle different date formats
      let date: Date;
      
      // Check if it's already a valid date string
      if (dateString.includes('T') || dateString.includes('-')) {
        date = new Date(dateString);
      } else {
        // Try parsing as is
        date = new Date(dateString);
      }
      
      // Validate the date
      if (isNaN(date.getTime())) {
        console.error('Invalid date:', dateString);
        return 'Invalid Date';
      }
      
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      });
    } catch (err) {
      console.error('Error formatting date:', dateString, err);
      return 'Invalid Date';
    }
  }

  // ✅ FIXED: Better datetime formatting
  function formatDateTime(dateString: string | null | undefined): string {
    if (!dateString) return 'N/A';
    
    try {
      const date = new Date(dateString);
      
      if (isNaN(date.getTime())) {
        console.error('Invalid datetime:', dateString);
        return 'Invalid Date';
      }
      
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: 'numeric',
        minute: '2-digit'
      });
    } catch (err) {
      console.error('Error formatting datetime:', dateString, err);
      return 'Invalid Date';
    }
  }

  function getWeekLabel(startDate: string, endDate: string): string {
    const start = formatDate(startDate);
    const end = formatDate(endDate);
    
    if (start === 'Invalid Date' || end === 'Invalid Date') {
      return 'Invalid Date Range';
    }
    
    return `${start} - ${end}`;
  }

  function getInitials(name: string): string {
    if (!name) return 'NA';
    return name.split(' ').map(n => n[0]).join('').toUpperCase();
  }
</script>

<div class="timesheet-approval">
  <div class="section-header">
    <h2>Timesheet Approvals</h2>
    <p class="section-subtitle">Review and approve timesheets submitted by your team</p>
  </div>

  {#if successMessage}
    <div class="alert alert-success">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
        <polyline points="22 4 12 14.01 9 11.01"></polyline>
      </svg>
      {successMessage}
      <button type="button" class="alert-close" onclick={() => successMessage = ''}>×</button>
    </div>
  {/if}

  {#if errorMessage}
    <div class="alert alert-error">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="8" x2="12" y2="12"></line>
        <line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      {errorMessage}
      <button type="button" class="alert-close" onclick={() => errorMessage = ''}>×</button>
    </div>
  {/if}

  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading pending timesheets...</p>
    </div>
  {:else if timesheets.length === 0}
    <div class="empty-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
        <polyline points="22 4 12 14.01 9 11.01"></polyline>
      </svg>
      <h3>All caught up!</h3>
      <p>There are no pending timesheets to review at this time.</p>
    </div>
  {:else}
    <div class="timesheets-list">
      {#each timesheets as timesheet}
        <div class="timesheet-card">
          <div class="timesheet-header">
            <div class="employee-info">
              <div class="employee-avatar">
                {getInitials(timesheet.employee_name)}
              </div>
              <div>
                <h3>{timesheet.employee_name || 'Unknown Employee'}</h3>
                <p class="week-label">{getWeekLabel(timesheet.start_date, timesheet.end_date)}</p>
              </div>
            </div>

            <div class="timesheet-summary">
              <div class="summary-item">
                <span class="summary-label">Total Hours</span>
                <span class="summary-value">{timesheet.total_hours?.toFixed(2) || '0.00'}</span>
              </div>
              {#if timesheet.overtime_hours && timesheet.overtime_hours > 0}
                <div class="summary-item">
                  <span class="summary-label">Overtime</span>
                  <span class="summary-value overtime">{timesheet.overtime_hours.toFixed(2)}</span>
                </div>
              {/if}
              <div class="summary-item">
                <span class="summary-label">Submitted</span>
                <span class="summary-value">{formatDateTime(timesheet.submitted_at)}</span>
              </div>
            </div>
          </div>

          <button 
            type="button"
            class="details-toggle"
            onclick={() => toggleTimesheetDetails(timesheet.id)}
          >
            {expandedTimesheet === timesheet.id ? 'Hide' : 'View'} Details
            <svg 
              class="toggle-icon {expandedTimesheet === timesheet.id ? 'expanded' : ''}"
              viewBox="0 0 24 24" 
              fill="none" 
              stroke="currentColor" 
              stroke-width="2"
            >
              <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
          </button>

          {#if expandedTimesheet === timesheet.id}
            <div class="timesheet-details">
              {#if loadingDetails === timesheet.id}
                <div class="details-loading">
                  <div class="spinner-sm"></div>
                  <span>Loading details...</span>
                </div>
              {:else if timesheet.time_entries && timesheet.time_entries.length > 0}
                <table class="entries-table">
                  <thead>
                    <tr>
                      <th>Date</th>
                      <th>Project</th>
                      <th>Type</th>
                      <th>Hours</th>
                      <th>Notes</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each timesheet.time_entries as entry}
                      <tr>
                        <td>{formatDate(entry.date)}</td>
                        <td>{entry.project_name || 'No Project'}</td>
                        <td><span class="type-badge type-{entry.type}">{entry.type}</span></td>
                        <td class="hours-cell">{entry.hours?.toFixed(2) || '0.00'}</td>
                        <td class="notes-cell">{entry.notes || '-'}</td>
                      </tr>
                    {/each}
                  </tbody>
                  <tfoot>
                    <tr>
                      <td colspan="3" class="total-label">Total</td>
                      <td class="total-hours">
                        {timesheet.time_entries.reduce((sum, e) => sum + (e.hours || 0), 0).toFixed(2)}
                      </td>
                      <td></td>
                    </tr>
                  </tfoot>
                </table>
              {:else}
                <p class="no-entries">No time entries found</p>
              {/if}
            </div>
          {/if}

          <div class="timesheet-actions">
            <button 
              type="button"
              class="btn btn-reject"
              onclick={() => openRejectModal(timesheet)}
              disabled={processing}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
              Reject
            </button>
            <button 
              type="button"
              class="btn btn-approve"
              onclick={() => openApproveModal(timesheet)}
              disabled={processing}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12"></polyline>
              </svg>
              Approve
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Approval/Rejection Modal -->
{#if showModal && selectedTimesheet}
  <div class="modal-overlay" onclick={closeModal}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h3>{modalMode === 'approve' ? 'Approve' : 'Reject'} Timesheet</h3>
        <button type="button" class="modal-close" onclick={closeModal}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <div class="modal-body">
        <div class="modal-summary">
          <p><strong>Employee:</strong> {selectedTimesheet.employee_name}</p>
          <p><strong>Period:</strong> {getWeekLabel(selectedTimesheet.start_date, selectedTimesheet.end_date)}</p>
          <p><strong>Total Hours:</strong> {selectedTimesheet.total_hours?.toFixed(2) || '0.00'}</p>
          {#if selectedTimesheet.overtime_hours && selectedTimesheet.overtime_hours > 0}
            <p><strong>Overtime Hours:</strong> {selectedTimesheet.overtime_hours.toFixed(2)}</p>
          {/if}
        </div>

        {#if modalMode === 'reject'}
          <div class="rejection-reason">
            <label for="rejection-reason">
              Rejection Reason <span class="required">*</span>
            </label>
            <textarea
              id="rejection-reason"
              bind:value={rejectionReason}
              placeholder="Please provide a reason for rejecting this timesheet..."
              rows="4"
              required
            ></textarea>
          </div>
        {/if}

        {#if errorMessage}
          <div class="alert alert-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            {errorMessage}
          </div>
        {/if}
      </div>

      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" onclick={closeModal} disabled={processing}>
          Cancel
        </button>
        {#if modalMode === 'approve'}
          <button type="button" class="btn btn-approve" onclick={handleApprove} disabled={processing}>
            {processing ? 'Approving...' : 'Approve Timesheet'}
          </button>
        {:else}
          <button type="button" class="btn btn-reject" onclick={handleReject} disabled={processing}>
            {processing ? 'Rejecting...' : 'Reject Timesheet'}
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .timesheet-approval {
    padding: 24px;
  }

  .section-header {
    margin-bottom: 24px;
  }

  .section-header h2 {
    font-size: 24px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .section-subtitle {
    font-size: 14px;
    color: #6b7280;
    margin: 0;
  }

  /* Alerts */
  .alert {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 20px;
    position: relative;
  }

  .alert svg {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
  }

  .alert-success {
    background: #d1fae5;
    color: #065f46;
    border: 1px solid #a7f3d0;
  }

  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .alert-close {
    margin-left: auto;
    background: none;
    border: none;
    font-size: 24px;
    cursor: pointer;
    color: inherit;
    padding: 0 8px;
  }

  /* Loading States */
  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 20px;
    color: #6b7280;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin-bottom: 16px;
  }

  .spinner-sm {
    width: 20px;
    height: 20px;
    border: 3px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .details-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 40px;
    color: #6b7280;
    font-size: 14px;
  }

  /* Empty State */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 20px;
    text-align: center;
  }

  .empty-state svg {
    width: 64px;
    height: 64px;
    color: #d1d5db;
    margin-bottom: 16px;
  }

  .empty-state h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .empty-state p {
    font-size: 14px;
    color: #6b7280;
    margin: 0;
  }

  /* Timesheets List */
  .timesheets-list {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .timesheet-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    transition: all 0.2s;
  }

  .timesheet-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  }

  .timesheet-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid #e5e7eb;
  }

  .employee-info {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .employee-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .employee-info h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }

  .week-label {
    font-size: 14px;
    color: #6b7280;
    margin: 0;
  }

  .timesheet-summary {
    display: flex;
    gap: 32px;
  }

  .summary-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    text-align: right;
  }

  .summary-label {
    font-size: 12px;
    color: #6b7280;
    font-weight: 500;
  }

  .summary-value {
    font-size: 20px;
    font-weight: 700;
    color: #111827;
  }

  .summary-value.overtime {
    color: #dc2626;
  }

  /* Details Toggle */
  .details-toggle {
    width: 100%;
    padding: 12px 16px;
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    color: #374151;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: all 0.2s;
    margin-bottom: 16px;
  }

  .details-toggle:hover {
    background: #f3f4f6;
    border-color: #d1d5db;
  }

  .toggle-icon {
    width: 16px;
    height: 16px;
    transition: transform 0.2s;
  }

  .toggle-icon.expanded {
    transform: rotate(180deg);
  }

  /* Timesheet Details */
  .timesheet-details {
    margin-bottom: 16px;
  }

  .entries-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
  }

  .entries-table thead {
    background: #f9fafb;
  }

  .entries-table th {
    padding: 12px;
    text-align: left;
    font-weight: 600;
    color: #374151;
    border-bottom: 2px solid #e5e7eb;
  }

  .entries-table td {
    padding: 12px;
    border-bottom: 1px solid #f3f4f6;
    color: #111827;
  }

  .entries-table tbody tr:hover {
    background: #f9fafb;
  }

  .entries-table tfoot {
    background: #f9fafb;
    font-weight: 600;
  }

  .entries-table tfoot td {
    border-top: 2px solid #e5e7eb;
    border-bottom: none;
  }

  .hours-cell {
    text-align: right;
    font-weight: 600;
  }

  .total-label {
    text-align: right !important;
  }

  .total-hours {
    text-align: right;
    color: #3b82f6;
    font-size: 16px;
  }

  .notes-cell {
    color: #6b7280;
    font-size: 13px;
    max-width: 200px;
  }

  .type-badge {
    padding: 4px 8px;
    background: #f3f4f6;
    color: #374151;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 500;
    text-transform: capitalize;
  }

  .type-badge.type-regular {
    background: #dbeafe;
    color: #1e40af;
  }

  .type-badge.type-overtime {
    background: #fecaca;
    color: #991b1b;
  }

  .no-entries {
    text-align: center;
    color: #9ca3af;
    padding: 40px 20px;
    margin: 0;
    font-size: 14px;
  }

  /* Actions */
  .timesheet-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .btn {
    display: inline-flex;
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

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn svg {
    width: 16px;
    height: 16px;
  }

  .btn-approve {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    color: white;
  }

  .btn-approve:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
  }

  .btn-reject {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .btn-reject:hover:not(:disabled) {
    background: #fecaca;
  }

  .btn-secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #f9fafb;
  }

  /* Modal */
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
  }

  .modal {
    background: white;
    border-radius: 16px;
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h3 {
    font-size: 20px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .modal-close {
    padding: 8px;
    background: transparent;
    border: none;
    cursor: pointer;
    border-radius: 6px;
    color: #6b7280;
    transition: all 0.2s;
  }

  .modal-close:hover {
    background: #f3f4f6;
    color: #111827;
  }

  .modal-close svg {
    width: 20px;
    height: 20px;
  }

  .modal-body {
    padding: 24px;
    overflow-y: auto;
  }

  .modal-summary {
    background: #f9fafb;
    padding: 16px;
    border-radius: 8px;
    margin-bottom: 20px;
  }

  .modal-summary p {
    margin: 8px 0;
    font-size: 14px;
    color: #374151;
  }

  .rejection-reason {
    margin-bottom: 20px;
  }

  .rejection-reason label {
    display: block;
    font-size: 14px;
    font-weight: 600;
    color: #374151;
    margin-bottom: 8px;
  }

  .required {
    color: #ef4444;
  }

  .rejection-reason textarea {
    width: 100%;
    padding: 12px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    font-family: inherit;
    resize: vertical;
  }

  .rejection-reason textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: 1px solid #e5e7eb;
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  @media (max-width: 768px) {
    .timesheet-header {
      flex-direction: column;
      gap: 16px;
    }

    .timesheet-summary {
      flex-direction: column;
      gap: 12px;
    }

    .summary-item {
      flex-direction: row;
      justify-content: space-between;
      text-align: left;
    }

    .timesheet-actions {
      flex-direction: column;
    }

    .btn {
      width: 100%;
      justify-content: center;
    }
  }
</style>