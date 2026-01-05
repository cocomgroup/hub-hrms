<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface TimeEntry {
    id: string;
    date: string;
    clock_in: string;
    clock_out: string | null;
    total_hours: number;
    status: 'active' | 'completed' | 'submitted' | 'approved';
    notes?: string;
  }

  let loading = $state(false);
  let currentEntry = $state<TimeEntry | null>(null);
  let recentEntries = $state<TimeEntry[]>([]);
  let currentTime = $state(new Date());
  let editingEntry = $state<string | null>(null);
  let editNotes = $state('');
  let successMessage = $state('');
  let errorMessage = $state('');

  // Update current time every second
  let timeInterval: ReturnType<typeof setInterval>;

  onMount(() => {
    loadCurrentEntry();
    loadRecentEntries();
    
    timeInterval = setInterval(() => {
      currentTime = new Date();
    }, 1000);

    return () => {
      if (timeInterval) clearInterval(timeInterval);
    };
  });

  function formatTime(date: Date): string {
    return date.toLocaleTimeString('en-US', { 
      hour: '2-digit', 
      minute: '2-digit',
      second: '2-digit'
    });
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  function formatDateTime(dateStr: string): string {
    return new Date(dateStr).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function calculateDuration(clockIn: string, clockOut: string | null): string {
    const start = new Date(clockIn);
    const end = clockOut ? new Date(clockOut) : currentTime;
    const diff = end.getTime() - start.getTime();
    
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((diff % (1000 * 60)) / 1000);
    
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
  }

  async function loadCurrentEntry() {
    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/current`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        currentEntry = await response.json();
      } else if (response.status === 404) {
        currentEntry = null;
      }
    } catch (err) {
      console.error('Error loading current entry:', err);
    }
  }

  async function loadRecentEntries() {
    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/entries?limit=10`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        const entries = await response.json();
        recentEntries = Array.isArray(entries) ? entries : [];
      }
    } catch (err) {
      console.error('Error loading entries:', err);
    } finally {
      loading = false;
    }
  }

  async function clockIn() {
    errorMessage = '';
    successMessage = '';
    
    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/clock-in`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          clock_in: new Date().toISOString()
        })
      });
      
      if (response.ok) {
        currentEntry = await response.json();
        successMessage = 'Clocked in successfully!';
        loadRecentEntries();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to clock in.';
      }
    } catch (err) {
      console.error('Error clocking in:', err);
      errorMessage = 'An error occurred while clocking in.';
    }
  }

  async function clockOut() {
    if (!currentEntry) return;
    
    errorMessage = '';
    successMessage = '';
    
    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/clock-out`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          entry_id: currentEntry.id,
          clock_out: new Date().toISOString()
        })
      });
      
      if (response.ok) {
        successMessage = 'Clocked out successfully!';
        currentEntry = null;
        loadRecentEntries();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to clock out.';
      }
    } catch (err) {
      console.error('Error clocking out:', err);
      errorMessage = 'An error occurred while clocking out.';
    }
  }

  function startEdit(entry: TimeEntry) {
    editingEntry = entry.id;
    editNotes = entry.notes || '';
  }

  function cancelEdit() {
    editingEntry = null;
    editNotes = '';
  }

  async function saveEdit(entryId: string) {
    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/entries/${entryId}`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ notes: editNotes })
      });
      
      if (response.ok) {
        successMessage = 'Notes updated successfully!';
        editingEntry = null;
        editNotes = '';
        loadRecentEntries();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to update notes.';
      }
    } catch (err) {
      console.error('Error updating notes:', err);
      errorMessage = 'An error occurred while updating.';
    }
  }

  async function deleteEntry(entryId: string) {
    if (!confirm('Are you sure you want to delete this entry?')) return;
    
    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/entries/${entryId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        successMessage = 'Entry deleted successfully!';
        loadRecentEntries();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to delete entry.';
      }
    } catch (err) {
      console.error('Error deleting entry:', err);
      errorMessage = 'An error occurred while deleting.';
    }
  }

  async function submitTimesheet() {
    const unsubmittedEntries = recentEntries.filter(e => e.status === 'completed');
    
    if (unsubmittedEntries.length === 0) {
      errorMessage = 'No completed entries to submit.';
      return;
    }
    
    try {
      const response = await fetch(`${API_BASE_URL}/timesheet/submit`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          entry_ids: unsubmittedEntries.map(e => e.id)
        })
      });
      
      if (response.ok) {
        successMessage = 'Timesheet submitted for approval!';
        loadRecentEntries();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to submit timesheet.';
      }
    } catch (err) {
      console.error('Error submitting timesheet:', err);
      errorMessage = 'An error occurred while submitting.';
    }
  }

  const isClockedIn = $derived(currentEntry !== null && currentEntry.clock_out === null);
  const currentDuration = $derived(
    currentEntry && !currentEntry.clock_out
      ? calculateDuration(currentEntry.clock_in, null)
      : '00:00:00'
  );
</script>

<div class="timesheet-clock">
  <div class="header">
    <h1>⏰ Timesheet - Clock In/Out</h1>
    <div class="current-time">{formatTime(currentTime)}</div>
  </div>

  {#if successMessage}
    <div class="alert alert-success">{successMessage}</div>
  {/if}

  {#if errorMessage}
    <div class="alert alert-error">{errorMessage}</div>
  {/if}

  <!-- Clock Card -->
  <div class="clock-card">
    <div class="status-indicator" class:active={isClockedIn}>
      <div class="pulse"></div>
      <span>{isClockedIn ? 'Currently Working' : 'Not Clocked In'}</span>
    </div>

    {#if currentEntry && !currentEntry.clock_out}
      <div class="active-session">
        <div class="session-info">
          <span class="info-label">Clocked in at:</span>
          <span class="info-value">{formatDateTime(currentEntry.clock_in)}</span>
        </div>
        
        <div class="duration-display">
          <div class="duration-label">Time Elapsed</div>
          <div class="duration-value">{currentDuration}</div>
        </div>
        
        <button type="button" class="btn btn-clock-out" onclick={clockOut}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <polyline points="12 6 12 12 16 14"></polyline>
          </svg>
          Clock Out
        </button>
      </div>
    {:else}
      <div class="clock-in-prompt">
        <p>Ready to start your shift?</p>
        <button type="button" class="btn btn-clock-in" onclick={clockIn}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <polyline points="12 6 12 12 16 14"></polyline>
          </svg>
          Clock In
        </button>
      </div>
    {/if}
  </div>

  <!-- Recent Entries -->
  <div class="recent-entries">
    <div class="section-header">
      <h2>Recent Entries</h2>
      <button type="button" class="btn btn-secondary btn-sm" onclick={submitTimesheet}>
        Submit for Approval
      </button>
    </div>

    {#if loading}
      <div class="loading">Loading entries...</div>
    {:else if recentEntries.length === 0}
      <div class="empty-state">
        <div class="empty-icon">📋</div>
        <p>No time entries yet</p>
        <p class="text-muted">Clock in to start tracking your time</p>
      </div>
    {:else}
      <div class="entries-list">
        {#each recentEntries as entry}
          <div class="entry-card" class:active={entry.id === currentEntry?.id}>
            <div class="entry-header">
              <div class="entry-date">{formatDate(entry.date)}</div>
              <span class="status-badge status-{entry.status}">{entry.status}</span>
            </div>
            
            <div class="entry-times">
              <div class="time-item">
                <span class="time-label">In:</span>
                <span class="time-value">{formatDateTime(entry.clock_in)}</span>
              </div>
              <div class="time-divider">→</div>
              <div class="time-item">
                <span class="time-label">Out:</span>
                <span class="time-value">{entry.clock_out ? formatDateTime(entry.clock_out) : 'Active'}</span>
              </div>
              <div class="time-divider">|</div>
              <div class="time-item">
                <span class="time-label">Hours:</span>
                <span class="time-value hours">{entry.total_hours.toFixed(2)}</span>
              </div>
            </div>

            {#if editingEntry === entry.id}
              <div class="edit-notes">
                <textarea 
                  bind:value={editNotes}
                  placeholder="Add notes about this shift..."
                  rows="2"
                ></textarea>
                <div class="edit-actions">
                  <button type="button" class="btn btn-sm btn-secondary" onclick={cancelEdit}>Cancel</button>
                  <button type="button" class="btn btn-sm btn-primary" onclick={() => saveEdit(entry.id)}>Save</button>
                </div>
              </div>
            {:else}
              <div class="entry-notes">
                {#if entry.notes}
                  <p>{entry.notes}</p>
                {:else}
                  <p class="no-notes">No notes</p>
                {/if}
              </div>
              
              {#if entry.status === 'completed'}
                <div class="entry-actions">
                  <button type="button" class="btn-icon" onclick={() => startEdit(entry)} title="Edit notes">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                      <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                    </svg>
                  </button>
                  <button type="button" class="btn-icon btn-delete" onclick={() => deleteEntry(entry.id)} title="Delete entry">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="3 6 5 6 21 6"></polyline>
                      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                    </svg>
                  </button>
                </div>
              {/if}
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .timesheet-clock {
    padding: 2rem;
    max-width: 1000px;
    margin: 0 auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .header h1 {
    font-size: 1.75rem;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
  }

  .current-time {
    font-size: 1.5rem;
    font-weight: 600;
    font-family: 'Courier New', monospace;
    color: #3b82f6;
    padding: 0.5rem 1rem;
    background: #eff6ff;
    border-radius: 8px;
  }

  .alert {
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
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

  /* Clock Card */
  .clock-card {
    background: white;
    border-radius: 16px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    padding: 2.5rem;
    margin-bottom: 2rem;
  }

  .status-indicator {
    display: flex;
    align-items: center;
    gap: 1rem;
    justify-content: center;
    margin-bottom: 2rem;
    padding: 1rem;
    background: #f3f4f6;
    border-radius: 12px;
  }

  .status-indicator.active {
    background: #d1fae5;
  }

  .pulse {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: #9ca3af;
  }

  .status-indicator.active .pulse {
    background: #10b981;
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; transform: scale(1); }
    50% { opacity: 0.5; transform: scale(1.1); }
  }

  .status-indicator span {
    font-weight: 600;
    font-size: 1.125rem;
    color: #374151;
  }

  .status-indicator.active span {
    color: #065f46;
  }

  .active-session {
    text-align: center;
  }

  .session-info {
    display: flex;
    justify-content: center;
    gap: 1rem;
    margin-bottom: 1.5rem;
    font-size: 1rem;
  }

  .info-label {
    color: #6b7280;
  }

  .info-value {
    font-weight: 600;
    color: #1f2937;
  }

  .duration-display {
    margin-bottom: 2rem;
  }

  .duration-label {
    font-size: 0.875rem;
    color: #6b7280;
    margin-bottom: 0.5rem;
  }

  .duration-value {
    font-size: 3rem;
    font-weight: 700;
    font-family: 'Courier New', monospace;
    color: #3b82f6;
  }

  .clock-in-prompt {
    text-align: center;
    padding: 2rem 0;
  }

  .clock-in-prompt p {
    font-size: 1.125rem;
    color: #6b7280;
    margin-bottom: 1.5rem;
  }

  /* Recent Entries */
  .recent-entries {
    background: white;
    border-radius: 16px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    padding: 2rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .section-header h2 {
    font-size: 1.25rem;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
  }

  .loading {
    text-align: center;
    padding: 2rem;
    color: #6b7280;
  }

  .empty-state {
    text-align: center;
    padding: 3rem 2rem;
    color: #9ca3af;
  }

  .empty-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
    opacity: 0.5;
  }

  .text-muted {
    font-size: 0.875rem;
    margin-top: 0.5rem;
  }

  .entries-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .entry-card {
    padding: 1.25rem;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    transition: all 0.2s;
  }

  .entry-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
  }

  .entry-card.active {
    background: #eff6ff;
    border-color: #3b82f6;
  }

  .entry-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .entry-date {
    font-weight: 600;
    color: #1f2937;
  }

  .status-badge {
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: uppercase;
  }

  .status-active {
    background: #dbeafe;
    color: #1e40af;
  }

  .status-completed {
    background: #d1d5db;
    color: #374151;
  }

  .status-submitted {
    background: #fef3c7;
    color: #92400e;
  }

  .status-approved {
    background: #d1fae5;
    color: #065f46;
  }

  .entry-times {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
    flex-wrap: wrap;
  }

  .time-item {
    display: flex;
    gap: 0.5rem;
  }

  .time-label {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .time-value {
    font-weight: 500;
    color: #1f2937;
    font-size: 0.875rem;
  }

  .time-value.hours {
    color: #3b82f6;
    font-weight: 600;
  }

  .time-divider {
    color: #d1d5db;
  }

  .entry-notes {
    margin-bottom: 0.75rem;
    padding: 0.75rem;
    background: #f9fafb;
    border-radius: 6px;
    font-size: 0.875rem;
    color: #374151;
  }

  .no-notes {
    color: #9ca3af;
    font-style: italic;
  }

  .edit-notes textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.875rem;
    font-family: inherit;
    margin-bottom: 0.75rem;
    resize: vertical;
  }

  .edit-notes textarea:focus {
    outline: none;
    border-color: #3b82f6;
    ring: 2px;
    ring-color: #dbeafe;
  }

  .edit-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
  }

  .entry-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
  }

  .btn-icon {
    padding: 0.5rem;
    border: none;
    background: transparent;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.2s;
    color: #6b7280;
  }

  .btn-icon:hover {
    background: #f3f4f6;
    color: #1f2937;
  }

  .btn-icon svg {
    width: 18px;
    height: 18px;
  }

  .btn-delete {
    color: #ef4444;
  }

  .btn-delete:hover {
    background: #fee2e2;
  }

  /* Buttons */
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn svg {
    width: 18px;
    height: 18px;
  }

  .btn-sm {
    padding: 0.5rem 1rem;
    font-size: 0.8125rem;
  }

  .btn-primary {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
  }

  .btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  }

  .btn-secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-secondary:hover {
    background: #f9fafb;
    border-color: #9ca3af;
  }

  .btn-clock-in {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    color: white;
    font-size: 1.125rem;
    padding: 1rem 2rem;
  }

  .btn-clock-in:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4);
  }

  .btn-clock-out {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
    color: white;
    font-size: 1.125rem;
    padding: 1rem 2rem;
  }

  .btn-clock-out:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(239, 68, 68, 0.4);
  }

  @media (max-width: 768px) {
    .timesheet-clock {
      padding: 1rem;
    }

    .header {
      flex-direction: column;
      gap: 1rem;
      align-items: stretch;
    }

    .current-time {
      text-align: center;
    }

    .entry-times {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }

    .time-divider {
      display: none;
    }
  }
</style>