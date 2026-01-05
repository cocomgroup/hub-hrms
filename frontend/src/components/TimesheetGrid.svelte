<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface Project {
    id: string;
    name: string;
    description?: string;
  }

  interface TimesheetEntry {
    id?: string;
    date: string;
    project_id: string;
    project_name?: string;
    hours: number;
    description: string;
    status: 'draft' | 'submitted' | 'approved' | 'rejected';
  }

  interface WeekEntry {
    project_id: string;
    project_name: string;
    monday: number;
    tuesday: number;
    wednesday: number;
    thursday: number;
    friday: number;
    saturday: number;
    sunday: number;
    total: number;
    description: string;
    entries: { [key: string]: string }; // date -> entry_id mapping
  }

  let loading = $state(false);
  let saving = $state(false);
  let projects = $state<Project[]>([]);
  let weekEntries = $state<WeekEntry[]>([]);
  let currentWeekStart = $state(getWeekStart(new Date()));
  let successMessage = $state('');
  let errorMessage = $state('');

  // Get the dates for the current week
  const weekDates = $derived(getWeekDates(currentWeekStart));
  const weekTotal = $derived(
    weekEntries.reduce((sum, entry) => sum + entry.total, 0)
  );

  onMount(() => {
    loadProjects();
    loadWeekEntries();
  });

  function getWeekStart(date: Date): Date {
    const d = new Date(date);
    const day = d.getDay();
    const diff = d.getDate() - day + (day === 0 ? -6 : 1); // Adjust to Monday
    d.setDate(diff);
    d.setHours(0, 0, 0, 0);
    return d;
  }

  function getWeekDates(weekStart: Date): Date[] {
    const dates = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(weekStart);
      date.setDate(weekStart.getDate() + i);
      dates.push(date);
    }
    return dates;
  }

  function formatDate(date: Date): string {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function formatDateISO(date: Date): string {
    return date.toISOString().split('T')[0];
  }

  function previousWeek() {
    const newStart = new Date(currentWeekStart);
    newStart.setDate(newStart.getDate() - 7);
    currentWeekStart = newStart;
    loadWeekEntries();
  }

  function nextWeek() {
    const newStart = new Date(currentWeekStart);
    newStart.setDate(newStart.getDate() + 7);
    currentWeekStart = newStart;
    loadWeekEntries();
  }

  function currentWeek() {
    currentWeekStart = getWeekStart(new Date());
    loadWeekEntries();
  }

  async function loadProjects() {
    try {
      const response = await fetch(`${API_BASE_URL}/projects`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        projects = await response.json();
      }
    } catch (err) {
      console.error('Error loading projects:', err);
    }
  }

  async function loadWeekEntries() {
    loading = true;
    try {
      const startDate = formatDateISO(weekDates[0]);
      const endDate = formatDateISO(weekDates[6]);
      
      const response = await fetch(
        `${API_BASE_URL}/timesheet/entries?start_date=${startDate}&end_date=${endDate}`,
        { headers: { 'Authorization': `Bearer ${$authStore.token}` } }
      );
      
      if (response.ok) {
        const entries: TimesheetEntry[] = await response.json();
        weekEntries = convertToWeekView(entries);
      }
    } catch (err) {
      console.error('Error loading timesheet entries:', err);
    } finally {
      loading = false;
    }
  }

  function convertToWeekView(entries: TimesheetEntry[]): WeekEntry[] {
    const projectMap = new Map<string, WeekEntry>();
    
    entries.forEach(entry => {
      const date = new Date(entry.date);
      const dayIndex = date.getDay();
      const dayName = ['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'][dayIndex];
      
      if (!projectMap.has(entry.project_id)) {
        projectMap.set(entry.project_id, {
          project_id: entry.project_id,
          project_name: entry.project_name || 'Unknown Project',
          monday: 0,
          tuesday: 0,
          wednesday: 0,
          thursday: 0,
          friday: 0,
          saturday: 0,
          sunday: 0,
          total: 0,
          description: entry.description || '',
          entries: {}
        });
      }
      
      const weekEntry = projectMap.get(entry.project_id)!;
      weekEntry[dayName] = entry.hours;
      weekEntry.total += entry.hours;
      if (entry.id) {
        weekEntry.entries[formatDateISO(date)] = entry.id;
      }
    });
    
    return Array.from(projectMap.values());
  }

  function addEntry() {
    if (projects.length === 0) {
      errorMessage = 'No projects available. Please contact your manager.';
      return;
    }
    
    weekEntries = [...weekEntries, {
      project_id: projects[0].id,
      project_name: projects[0].name,
      monday: 0,
      tuesday: 0,
      wednesday: 0,
      thursday: 0,
      friday: 0,
      saturday: 0,
      sunday: 0,
      total: 0,
      description: '',
      entries: {}
    }];
  }

  function deleteEntry(index: number) {
    weekEntries = weekEntries.filter((_, i) => i !== index);
  }

  function updateHours(index: number, day: string, value: string) {
    const hours = parseFloat(value) || 0;
    weekEntries[index][day] = hours;
    weekEntries[index].total = calculateRowTotal(weekEntries[index]);
    weekEntries = [...weekEntries]; // Trigger reactivity
  }

  function updateProject(index: number, projectId: string) {
    const project = projects.find(p => p.id === projectId);
    if (project) {
      weekEntries[index].project_id = projectId;
      weekEntries[index].project_name = project.name;
      weekEntries = [...weekEntries];
    }
  }

  function updateDescription(index: number, description: string) {
    weekEntries[index].description = description;
  }

  function calculateRowTotal(entry: WeekEntry): number {
    return entry.monday + entry.tuesday + entry.wednesday + 
           entry.thursday + entry.friday + entry.saturday + entry.sunday;
  }

  async function saveDraft() {
    await saveEntries('draft');
  }

  async function submitTimesheet() {
    if (weekTotal === 0) {
      errorMessage = 'Please enter at least some hours before submitting.';
      return;
    }
    await saveEntries('submitted');
  }

  async function saveEntries(status: 'draft' | 'submitted') {
    saving = true;
    errorMessage = '';
    successMessage = '';
    
    try {
      const entriesToSave: TimesheetEntry[] = [];
      
      weekEntries.forEach(weekEntry => {
        const days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
        days.forEach((day, index) => {
          const hours = weekEntry[day];
          if (hours > 0) {
            const date = new Date(weekDates[index + 1]); // +1 because weekDates[0] is Sunday
            entriesToSave.push({
              id: weekEntry.entries[formatDateISO(date)],
              date: formatDateISO(date),
              project_id: weekEntry.project_id,
              hours,
              description: weekEntry.description,
              status
            });
          }
        });
      });
      
      const response = await fetch(`${API_BASE_URL}/timesheet/entries/bulk`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ entries: entriesToSave })
      });
      
      if (response.ok) {
        successMessage = status === 'submitted' 
          ? 'Timesheet submitted successfully!' 
          : 'Timesheet saved as draft.';
        loadWeekEntries();
      } else {
        const error = await response.json();
        errorMessage = error.error || 'Failed to save timesheet.';
      }
    } catch (err) {
      console.error('Error saving timesheet:', err);
      errorMessage = 'An error occurred while saving.';
    } finally {
      saving = false;
    }
  }
</script>

<div class="timesheet-grid">
  <div class="header">
    <h1>📊 Timesheet - Grid View</h1>
    <div class="week-nav">
      <button type="button" class="btn btn-secondary" onclick={previousWeek}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="15 18 9 12 15 6"></polyline>
        </svg>
        Previous
      </button>
      <button type="button" class="btn btn-secondary" onclick={currentWeek}>Today</button>
      <button type="button" class="btn btn-secondary" onclick={nextWeek}>
        Next
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="9 18 15 12 9 6"></polyline>
        </svg>
      </button>
    </div>
  </div>

  <div class="week-display">
    Week of {formatDate(weekDates[0])} - {formatDate(weekDates[6])}
  </div>

  {#if successMessage}
    <div class="alert alert-success">{successMessage}</div>
  {/if}

  {#if errorMessage}
    <div class="alert alert-error">{errorMessage}</div>
  {/if}

  {#if loading}
    <div class="loading">Loading timesheet...</div>
  {:else}
    <div class="grid-container">
      <table class="timesheet-table">
        <thead>
          <tr>
            <th class="project-col">Project</th>
            <th>Mon<br/><span class="date-label">{formatDate(weekDates[1])}</span></th>
            <th>Tue<br/><span class="date-label">{formatDate(weekDates[2])}</span></th>
            <th>Wed<br/><span class="date-label">{formatDate(weekDates[3])}</span></th>
            <th>Thu<br/><span class="date-label">{formatDate(weekDates[4])}</span></th>
            <th>Fri<br/><span class="date-label">{formatDate(weekDates[5])}</span></th>
            <th>Sat<br/><span class="date-label">{formatDate(weekDates[6])}</span></th>
            <th>Sun<br/><span class="date-label">{formatDate(weekDates[0])}</span></th>
            <th class="total-col">Total</th>
            <th class="actions-col">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each weekEntries as entry, index}
            <tr class="entry-row">
              <td class="project-cell">
                <select 
                  value={entry.project_id}
                  onchange={(e) => updateProject(index, e.currentTarget.value)}
                  class="project-select"
                >
                  {#each projects as project}
                    <option value={project.id}>{project.name}</option>
                  {/each}
                </select>
                <input 
                  type="text"
                  value={entry.description}
                  oninput={(e) => updateDescription(index, e.currentTarget.value)}
                  placeholder="Description (optional)"
                  class="description-input"
                />
              </td>
              <td><input type="number" step="0.25" min="0" max="24" value={entry.monday} oninput={(e) => updateHours(index, 'monday', e.currentTarget.value)} /></td>
              <td><input type="number" step="0.25" min="0" max="24" value={entry.tuesday} oninput={(e) => updateHours(index, 'tuesday', e.currentTarget.value)} /></td>
              <td><input type="number" step="0.25" min="0" max="24" value={entry.wednesday} oninput={(e) => updateHours(index, 'wednesday', e.currentTarget.value)} /></td>
              <td><input type="number" step="0.25" min="0" max="24" value={entry.thursday} oninput={(e) => updateHours(index, 'thursday', e.currentTarget.value)} /></td>
              <td><input type="number" step="0.25" min="0" max="24" value={entry.friday} oninput={(e) => updateHours(index, 'friday', e.currentTarget.value)} /></td>
              <td><input type="number" step="0.25" min="0" max="24" value={entry.saturday} oninput={(e) => updateHours(index, 'saturday', e.currentTarget.value)} /></td>
              <td><input type="number" step="0.25" min="0" max="24" value={entry.sunday} oninput={(e) => updateHours(index, 'sunday', e.currentTarget.value)} /></td>
              <td class="total-cell">{entry.total.toFixed(2)}</td>
              <td class="actions-cell">
                <button type="button" class="btn-icon btn-delete" onclick={() => deleteEntry(index)} title="Delete row">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="3 6 5 6 21 6"></polyline>
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                  </svg>
                </button>
              </td>
            </tr>
          {/each}
          
          {#if weekEntries.length === 0}
            <tr>
              <td colspan="10" class="empty-state">
                No entries for this week. Click "Add Entry" to get started.
              </td>
            </tr>
          {/if}
          
          <tr class="total-row">
            <td class="total-label">Week Total</td>
            <td colspan="7"></td>
            <td class="week-total">{weekTotal.toFixed(2)}</td>
            <td></td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="actions">
      <button type="button" class="btn btn-secondary" onclick={addEntry}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        Add Entry
      </button>
      
      <div class="action-group">
        <button type="button" class="btn btn-secondary" onclick={saveDraft} disabled={saving}>
          {saving ? 'Saving...' : 'Save Draft'}
        </button>
        <button type="button" class="btn btn-primary" onclick={submitTimesheet} disabled={saving || weekTotal === 0}>
          {saving ? 'Submitting...' : 'Submit Timesheet'}
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .timesheet-grid {
    padding: 2rem;
    max-width: 1600px;
    margin: 0 auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .header h1 {
    font-size: 1.75rem;
    font-weight: 600;
    color: #1f2937;
    margin: 0;
  }

  .week-nav {
    display: flex;
    gap: 0.5rem;
  }

  .week-display {
    text-align: center;
    font-size: 1.125rem;
    font-weight: 500;
    color: #4b5563;
    margin-bottom: 1.5rem;
    padding: 0.75rem;
    background: #f9fafb;
    border-radius: 8px;
  }

  .alert {
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1rem;
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

  .loading {
    text-align: center;
    padding: 3rem;
    color: #6b7280;
  }

  .grid-container {
    overflow-x: auto;
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    margin-bottom: 1.5rem;
  }

  .timesheet-table {
    width: 100%;
    border-collapse: collapse;
    min-width: 1200px;
  }

  thead {
    background: #f9fafb;
    border-bottom: 2px solid #e5e7eb;
  }

  th {
    padding: 1rem 0.75rem;
    text-align: center;
    font-weight: 600;
    font-size: 0.875rem;
    color: #374151;
    white-space: nowrap;
  }

  .date-label {
    font-size: 0.75rem;
    font-weight: 400;
    color: #6b7280;
  }

  .project-col {
    text-align: left;
    min-width: 300px;
  }

  .total-col {
    min-width: 80px;
  }

  .actions-col {
    min-width: 60px;
  }

  td {
    padding: 0.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .entry-row:hover {
    background: #f9fafb;
  }

  .project-cell {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .project-select {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.875rem;
    background: white;
  }

  .project-select:focus {
    outline: none;
    border-color: #3b82f6;
    ring: 2px;
    ring-color: #dbeafe;
  }

  .description-input {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.875rem;
  }

  .description-input:focus {
    outline: none;
    border-color: #3b82f6;
    ring: 2px;
    ring-color: #dbeafe;
  }

  input[type="number"] {
    width: 60px;
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    text-align: center;
    font-size: 0.875rem;
  }

  input[type="number"]:focus {
    outline: none;
    border-color: #3b82f6;
    ring: 2px;
    ring-color: #dbeafe;
  }

  /* Remove spinner arrows */
  input[type="number"]::-webkit-inner-spin-button,
  input[type="number"]::-webkit-outer-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  .total-cell {
    font-weight: 600;
    color: #3b82f6;
    text-align: center;
    font-size: 1rem;
  }

  .actions-cell {
    text-align: center;
  }

  .btn-icon {
    padding: 0.5rem;
    border: none;
    background: transparent;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.2s;
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

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #9ca3af;
  }

  .total-row {
    background: #f3f4f6;
    font-weight: 600;
  }

  .total-label {
    text-align: right;
    padding-right: 1rem;
    color: #374151;
  }

  .week-total {
    font-size: 1.125rem;
    color: #1f2937;
    text-align: center;
  }

  .actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .action-group {
    display: flex;
    gap: 1rem;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn svg {
    width: 18px;
    height: 18px;
  }

  .btn-primary {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  }

  .btn-secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #f9fafb;
    border-color: #9ca3af;
  }

  @media (max-width: 768px) {
    .timesheet-grid {
      padding: 1rem;
    }

    .header {
      flex-direction: column;
      gap: 1rem;
      align-items: stretch;
    }

    .actions {
      flex-direction: column;
      gap: 1rem;
    }

    .action-group {
      flex-direction: column;
    }
  }
</style>