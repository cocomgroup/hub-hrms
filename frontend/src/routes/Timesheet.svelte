<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api'; 
  const API_BASE_URL = getApiBaseUrl();
  
  interface TimeEntry {
    id: string;
    employee_id: string;
    entry_date: string;
    clock_in?: string;
    clock_out?: string;
    break_duration: number;
    notes: string;
    entry_type: string;
    status: string;
    total_hours?: number;
    projects?: TimeEntryProject[];
  }

  interface TimeEntryProject {
    id: string;
    project_id: string;
    project_name: string;
    project_code: string;
    hours: number;
    notes: string;
  }

  interface Project {
    id: string;
    name: string;
    code: string;
    status: string;
  }

  interface TimesheetPeriod {
    id: string;
    start_date: string;
    end_date: string;
    status: string;
    total_regular_hours: number;
    total_overtime_hours: number;
    total_pto_hours: number;
  }

  let activeEntry: TimeEntry | null = null;
  let currentPeriod: TimesheetPeriod | null = null;
  let timeEntries: TimeEntry[] = [];
  let projects: Project[] = [];
  let loading = false;
  let error = '';
  let success = '';
  
  // Filter and view state
  let viewMode: 'week' | 'period' = 'week';
  let selectedDate = new Date().toISOString().split('T')[0];
  let showAddEntry = false;
  let editingEntry: TimeEntry | null = null;
  
  // New entry form
  let newEntry = {
    entry_date: new Date().toISOString().split('T')[0],
    clock_in: '',
    clock_out: '',
    break_duration: 0,
    notes: '',
    entry_type: 'regular',
    projects: [] as { project_id: string; hours: number; notes: string }[]
  };

  // Clock in/out state
  let clockedIn = false;
  let clockInTime = '';
  let elapsedTime = '00:00:00';
  let timerInterval: number;

  onMount(() => {
    loadData();
    checkActiveClockIn();
  });

  async function loadData() {
    loading = true;
    error = '';
    
    try {
      await Promise.all([
        loadCurrentPeriod(),
        loadTimeEntries(),
        loadProjects()
      ]);
    } catch (err: any) {
      error = err.message || 'Failed to load data';
    } finally {
      loading = false;
    }
  }

  async function loadCurrentPeriod() {
    const response = await fetch('/api/timesheet/periods/current', {
      headers: {
        'Authorization': `Bearer ${$authStore.token}`
      }
    });
    
    if (!response.ok) throw new Error('Failed to load current period');
    currentPeriod = await response.json();
  }

  async function loadTimeEntries() {
    const startDate = currentPeriod?.start_date || getWeekStart();
    const endDate = currentPeriod?.end_date || getWeekEnd();
    
    const response = await fetch(
      `/api/timesheet/entries?start_date=${startDate}&end_date=${endDate}`,
      {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      }
    );
    
    if (!response.ok) throw new Error('Failed to load time entries');
    const data = await response.json();
    timeEntries = Array.isArray(data) ? data : [];
  }

  async function loadProjects() {
    const response = await fetch('/api/timesheet/projects?status=active', {
      headers: {
        'Authorization': `Bearer ${$authStore.token}`
      }
    });
    
    if (!response.ok) throw new Error('Failed to load projects');
    projects = await response.json();
  }

  async function checkActiveClockIn() {
    try {
      const response = await fetch('/api/timesheet/active', {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (response.ok) {
        activeEntry = await response.json();
        if (activeEntry && activeEntry.clock_in) {
          clockedIn = true;
          clockInTime = activeEntry.clock_in;
          startTimer();
        }
      }
    } catch (err) {
      console.error('Failed to check active clock-in:', err);
    }
  }

  async function clockIn() {
    loading = true;
    error = '';
    
    try {
      const response = await fetch('/api/timesheet/clock-in', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ notes: '' })
      });
      
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to clock in');
      }
      
      activeEntry = await response.json();
      clockedIn = true;
      clockInTime = activeEntry!.clock_in!;
      startTimer();
      success = 'Clocked in successfully!';
      setTimeout(() => success = '', 3000);
      
      await loadTimeEntries();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function clockOut() {
    loading = true;
    error = '';
    
    const breakDuration = parseInt(prompt('Break duration in minutes:', '0') || '0');
    
    try {
      const response = await fetch('/api/timesheet/clock-out', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ 
          break_duration: breakDuration,
          notes: '' 
        })
      });
      
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to clock out');
      }
      
      activeEntry = await response.json();
      clockedIn = false;
      stopTimer();
      success = `Clocked out! Total: ${activeEntry!.total_hours?.toFixed(2)} hours`;
      setTimeout(() => success = '', 3000);
      
      await loadTimeEntries();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function startTimer() {
    if (timerInterval) clearInterval(timerInterval);
    
    timerInterval = window.setInterval(() => {
      if (!clockInTime) return;
      
      const start = new Date(clockInTime);
      const now = new Date();
      const diff = now.getTime() - start.getTime();
      
      const hours = Math.floor(diff / 3600000);
      const minutes = Math.floor((diff % 3600000) / 60000);
      const seconds = Math.floor((diff % 60000) / 1000);
      
      elapsedTime = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    }, 1000);
  }

  function stopTimer() {
    if (timerInterval) {
      clearInterval(timerInterval);
      timerInterval = 0;
    }
    elapsedTime = '00:00:00';
  }

  async function saveEntry() {
    loading = true;
    error = '';
    
    try {
      const url = editingEntry 
        ? `${API_BASE_URL}/timesheet/entries/${editingEntry.id}`
        : `${API_BASE_URL}/timesheet/entries`;
      
      const method = editingEntry ? 'PUT' : 'POST';
      
      // ✅ FIX: Convert time strings to ISO datetime format
      const entryData = {
        ...newEntry,
        // Convert "HH:MM" to "YYYY-MM-DDTHH:MM:00Z"
        clock_in: newEntry.clock_in 
          ? `${newEntry.entry_date}T${newEntry.clock_in}:00Z` 
          : null,
        clock_out: newEntry.clock_out 
          ? `${newEntry.entry_date}T${newEntry.clock_out}:00Z` 
          : null
      };
      
      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(entryData)
      });
      
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to save entry');
      }
      
      success = editingEntry ? 'Entry updated!' : 'Entry created!';
      setTimeout(() => success = '', 3000);
      
      resetForm();
      await loadTimeEntries();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function deleteEntry(id: string) {
    if (!confirm('Delete this time entry?')) return;
    
    loading = true;
    error = '';
    
    try {
      const response = await fetch(`/api/timesheet/entries/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to delete entry');
      
      success = 'Entry deleted!';
      setTimeout(() => success = '', 3000);
      await loadTimeEntries();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function editEntry(entry: TimeEntry) {
    editingEntry = entry;
    newEntry = {
      entry_date: entry.entry_date.split('T')[0],
      clock_in: entry.clock_in?.substring(11, 16) || '',
      clock_out: entry.clock_out?.substring(11, 16) || '',
      break_duration: entry.break_duration,
      notes: entry.notes,
      entry_type: entry.entry_type,
      projects: entry.projects?.map(p => ({
        project_id: p.project_id,
        hours: p.hours,
        notes: p.notes
      })) || []
    };
    showAddEntry = true;
  }

  function resetForm() {
    showAddEntry = false;
    editingEntry = null;
    newEntry = {
      entry_date: new Date().toISOString().split('T')[0],
      clock_in: '',
      clock_out: '',
      break_duration: 0,
      notes: '',
      entry_type: 'regular',
      projects: []
    };
  }

  function addProjectAllocation() {
    newEntry.projects.push({ project_id: '', hours: 0, notes: '' });
    newEntry = newEntry;
  }

  function removeProjectAllocation(index: number) {
    newEntry.projects.splice(index, 1);
    newEntry = newEntry;
  }

  async function submitTimesheet() {
    if (!currentPeriod) return;
    
    if (!confirm('Submit timesheet for approval?')) return;
    
    loading = true;
    error = '';
    
    try {
      const response = await fetch(`/api/timesheet/periods/${currentPeriod.id}/submit`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to submit timesheet');
      
      success = 'Timesheet submitted for approval!';
      setTimeout(() => success = '', 3000);
      await loadCurrentPeriod();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function getWeekStart(): string {
    const date = new Date(selectedDate);
    const day = date.getDay();
    const diff = date.getDate() - day;
    const weekStart = new Date(date.setDate(diff));
    return weekStart.toISOString().split('T')[0];
  }

  function getWeekEnd(): string {
    const date = new Date(selectedDate);
    const day = date.getDay();
    const diff = date.getDate() + (6 - day);
    const weekEnd = new Date(date.setDate(diff));
    return weekEnd.toISOString().split('T')[0];
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', { 
      weekday: 'short', 
      month: 'short', 
      day: 'numeric' 
    });
  }

  function formatTime(timeStr: string | undefined): string {
    if (!timeStr) return '-';
    return new Date(timeStr).toLocaleTimeString('en-US', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  }

  $: periodTotal = currentPeriod 
    ? currentPeriod.total_regular_hours + currentPeriod.total_overtime_hours + currentPeriod.total_pto_hours
    : 0;
</script>

<div class="timesheet-container">
  <div class="timesheet-header">
    <h1>Timesheet</h1>
    {#if currentPeriod}
      <div class="period-info">
        <div class="period-dates">
          {formatDate(currentPeriod.start_date)} - {formatDate(currentPeriod.end_date)}
        </div>
        <span class="status-badge status-{currentPeriod.status}">
          {currentPeriod.status}
        </span>
      </div>
    {/if}
  </div>

  {#if error}
    <div class="alert alert-error">{error}</div>
  {/if}
  
  {#if success}
    <div class="alert alert-success">{success}</div>
  {/if}

  <!-- Clock In/Out Section -->
  <div class="clock-section">
    <div class="clock-display">
      <div class="clock-status">
        <div class="status-icon">
          {clockedIn ? '🟢' : '⚪'}
        </div>
        <div class="status-info">
          <h3>{clockedIn ? 'Clocked In' : 'Not Clocked In'}</h3>
          <p>
            {#if clockedIn && clockInTime}
              Since {formatTime(clockInTime)}
            {:else}
              Ready to clock in
            {/if}
          </p>
          {#if clockedIn}
            <div class="elapsed-time">{elapsedTime}</div>
          {/if}
        </div>
      </div>
    </div>
    <div class="clock-actions">
      {#if !clockedIn}
        <button class="btn btn-success btn-lg" on:click={clockIn} disabled={loading}>
          Clock In
        </button>
      {:else}
        <button class="btn btn-danger btn-lg" on:click={clockOut} disabled={loading}>
          Clock Out
        </button>
      {/if}
    </div>
  </div>

  <!-- Summary Cards -->
  <div class="summary-cards">
    <div class="card">
      <h4>Regular Hours</h4>
      <div class="card-value">{(currentPeriod?.total_regular_hours ?? 0).toFixed(1)}</div>
    </div>
    <div class="card">
      <h4>Overtime Hours</h4>
      <div class="card-value">{(currentPeriod?.total_overtime_hours ?? 0).toFixed(1)}</div>
    </div>
    <div class="card">
      <h4>PTO Hours</h4>
      <div class="card-value">{(currentPeriod?.total_pto_hours ?? 0).toFixed(1)}</div>
    </div>
    <div class="card">
      <h4>Total Hours</h4>
      <div class="card-value">{periodTotal.toFixed(1)}</div>
    </div>
  </div>

  <!-- Actions Bar -->
  <div class="actions-bar">
    <button class="btn btn-primary" on:click={() => showAddEntry = true}>
      + Add Entry
    </button>
    {#if currentPeriod?.status === 'draft'}
      <button class="btn btn-success" on:click={submitTimesheet} disabled={loading || timeEntries.length === 0}>
        Submit Timesheet
      </button>
    {/if}
  </div>

  <!-- Time Entries Table -->
  <div class="entries-table">
    <table>
      <thead>
        <tr>
          <th>Date</th>
          <th>Clock In</th>
          <th>Clock Out</th>
          <th>Break</th>
          <th>Total Hours</th>
          <th>Type</th>
          <th>Projects</th>
          <th>Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#if !timeEntries || timeEntries.length === 0}
          <tr>
            <td colspan="9">
              <div class="empty-state">
                No time entries for this period. Add your first entry above!
              </div>
            </td>
          </tr>
        {:else}
          {#each timeEntries as entry}
            <tr>
              <td>{formatDate(entry.entry_date)}</td>
              <td>{formatTime(entry.clock_in)}</td>
              <td>{formatTime(entry.clock_out)}</td>
              <td>{entry.break_duration}m</td>
              <td>{entry.total_hours?.toFixed(2) || '-'}</td>
              <td>
                <span class="entry-type-badge type-{entry.entry_type}">
                  {entry.entry_type}
                </span>
              </td>
              <td>
                {#if entry.projects && entry.projects.length > 0}
                  <div class="project-list">
                    {#each entry.projects as project}
                      <div class="project-item">
                        {project.project_name} ({project.hours}h)
                      </div>
                    {/each}
                  </div>
                {:else}
                  -
                {/if}
              </td>
              <td>
                <span class="status-badge status-{entry.status}">
                  {entry.status}
                </span>
              </td>
              <td>
                {#if entry.status === 'draft'}
                  <button class="btn-icon" on:click={() => editEntry(entry)} title="Edit entry" aria-label="Edit entry">
                    ✏️
                  </button>
                  <button class="btn-icon" on:click={() => deleteEntry(entry.id)} title="Delete entry" aria-label="Delete entry">
                    🗑️
                  </button>
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <!-- Add/Edit Entry Modal -->
  {#if showAddEntry}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="modal-overlay" on:click={resetForm}>
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div class="modal" on:click|stopPropagation>
        <div class="modal-header">
          <h2>{editingEntry ? 'Edit' : 'Add'} Time Entry</h2>
          <button class="close-btn" on:click={resetForm}>×</button>
        </div>
        
        <div class="modal-body">
          <div class="form-group">
            <label for="entry-date">Date</label>
            <input id="entry-date" type="date" bind:value={newEntry.entry_date} />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="clock-in">Clock In</label>
              <input id="clock-in" type="time" bind:value={newEntry.clock_in} />
            </div>
            <div class="form-group">
              <label for="clock-out">Clock Out</label>
              <input id="clock-out" type="time" bind:value={newEntry.clock_out} />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="break-duration">Break Duration (minutes)</label>
              <input id="break-duration" type="number" bind:value={newEntry.break_duration} min="0" />
            </div>
            <div class="form-group">
              <label for="entry-type">Entry Type</label>
              <select id="entry-type" bind:value={newEntry.entry_type}>
                <option value="regular">Regular</option>
                <option value="overtime">Overtime</option>
                <option value="pto">PTO</option>
                <option value="sick">Sick Leave</option>
                <option value="holiday">Holiday</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label for="entry-notes">Notes</label>
            <textarea id="entry-notes" bind:value={newEntry.notes} rows="2"></textarea>
          </div>

          <!-- Project Allocations -->
          <div class="form-section">
            <div class="section-header">
              <h3>Project Allocations</h3>
              <button class="btn btn-sm btn-secondary" on:click={addProjectAllocation}>
                + Add Project
              </button>
            </div>
            
            {#each newEntry.projects as projectAlloc, i}
              <div class="project-allocation">
                <select bind:value={projectAlloc.project_id} aria-label="Select project">
                  <option value="">Select Project</option>
                  {#each projects as project}
                    <option value={project.id}>{project.name} ({project.code})</option>
                  {/each}
                </select>
                <input 
                  type="number" 
                  bind:value={projectAlloc.hours} 
                  placeholder="Hours" 
                  min="0" 
                  step="0.5"
                  aria-label="Project hours"
                />
                <button 
                  class="btn-icon" 
                  on:click={() => removeProjectAllocation(i)}
                  title="Remove project allocation"
                  aria-label="Remove project allocation"
                >
                  ❌
                </button>
              </div>
            {/each}
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" on:click={resetForm}>
            Cancel
          </button>
          <button class="btn btn-primary" on:click={saveEntry} disabled={loading}>
            {editingEntry ? 'Update' : 'Save'} Entry
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .timesheet-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 2rem;
  }

  .timesheet-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .period-info {
    display: flex;
    gap: 1rem;
    align-items: center;
  }

  .period-dates {
    font-size: 1.1rem;
    font-weight: 500;
  }

  .alert {
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1rem;
  }

  .alert-error {
    background-color: #fee;
    color: #c33;
    border: 1px solid #fcc;
  }

  .alert-success {
    background-color: #efe;
    color: #3c3;
    border: 1px solid #cfc;
  }

  .clock-section {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 12px;
    padding: 2rem;
    color: white;
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .clock-display {
    flex: 1;
  }

  .clock-status {
    display: flex;
    gap: 1.5rem;
    align-items: center;
  }

  .status-icon {
    font-size: 3rem;
  }

  .status-info h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.5rem;
  }

  .status-info p {
    margin: 0;
    opacity: 0.9;
  }

  .elapsed-time {
    font-size: 2rem;
    font-weight: bold;
    margin-top: 0.5rem;
    font-family: 'Courier New', monospace;
  }

  .clock-actions button {
    padding: 1rem 2.5rem;
    font-size: 1.2rem;
  }

  .summary-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .card {
    background: white;
    border-radius: 8px;
    padding: 1.5rem;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }

  .card h4 {
    margin: 0 0 0.5rem 0;
    color: #666;
    font-size: 0.9rem;
    font-weight: 500;
  }

  .card-value {
    font-size: 2rem;
    font-weight: bold;
    color: #333;
  }

  .actions-bar {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .entries-table {
    background: white;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    overflow: hidden;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th, td {
    padding: 1rem;
    text-align: left;
    border-bottom: 1px solid #eee;
  }

  th {
    background-color: #f8f9fa;
    font-weight: 600;
    color: #495057;
  }

  tbody tr:hover {
    background-color: #f8f9fa;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #999;
  }

  .status-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .status-draft {
    background-color: #e3f2fd;
    color: #1976d2;
  }

  .status-submitted {
    background-color: #fff3e0;
    color: #f57c00;
  }

  .status-approved {
    background-color: #e8f5e9;
    color: #388e3c;
  }

  .status-rejected {
    background-color: #ffebee;
    color: #d32f2f;
  }

  .entry-type-badge {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.875rem;
  }

  .type-regular {
    background-color: #e3f2fd;
    color: #1976d2;
  }

  .type-overtime {
    background-color: #fff3e0;
    color: #f57c00;
  }

  .type-pto, .type-sick, .type-holiday {
    background-color: #f3e5f5;
    color: #7b1fa2;
  }

  .project-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .project-item {
    font-size: 0.875rem;
    color: #666;
  }

  .btn-icon {
    background: none;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0.25rem;
    opacity: 0.7;
  }

  .btn-icon:hover {
    opacity: 1;
  }

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
    border-radius: 12px;
    width: 90%;
    max-width: 700px;
    max-height: 90vh;
    overflow-y: auto;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid #eee;
  }

  .modal-header h2 {
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 2rem;
    cursor: pointer;
    color: #999;
    line-height: 1;
  }

  .modal-body {
    padding: 1.5rem;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding: 1.5rem;
    border-top: 1px solid #eee;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: #333;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 6px;
    font-size: 1rem;
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .form-section {
    margin-top: 2rem;
    padding-top: 1.5rem;
    border-top: 1px solid #eee;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .section-header h3 {
    margin: 0;
    font-size: 1.1rem;
  }

  .project-allocation {
    display: grid;
    grid-template-columns: 2fr 1fr auto;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .btn {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 6px;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: #667eea;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: #5568d3;
  }

  .btn-success {
    background-color: #48bb78;
    color: white;
  }

  .btn-success:hover:not(:disabled) {
    background-color: #38a169;
  }

  .btn-danger {
    background-color: #f56565;
    color: white;
  }

  .btn-danger:hover:not(:disabled) {
    background-color: #e53e3e;
  }

  .btn-secondary {
    background-color: #e2e8f0;
    color: #4a5568;
  }

  .btn-secondary:hover:not(:disabled) {
    background-color: #cbd5e0;
  }

  .btn-sm {
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
  }

  .btn-lg {
    padding: 1rem 2rem;
    font-size: 1.1rem;
  }
</style>