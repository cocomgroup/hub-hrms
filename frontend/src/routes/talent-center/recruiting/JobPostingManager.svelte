<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  interface Job {
    id: string;
    title: string;
    description: string;
    department: string;
    location: string;
    employment_type: 'full-time' | 'part-time' | 'contract' | 'internship';
    salary_min: number;
    salary_max: number;
    salary_currency: string;
    requirements: string[];
    responsibilities: string[];
    benefits: string[];
    status: 'draft' | 'active' | 'closed' | 'archived';
    providers: string[];
    applicant_count: number;
    created_at: string;
    posted_at?: string;
    closed_at?: string;
  }
  
  interface Provider {
    id: string;
    name: string;
    type: string;
    status: string;
  }
  
  let jobs: Job[] = [];
  let providers: Provider[] = [];
  let loading = true;
  let showJobModal = false;
  let editingJob: Job | null = null;
  let filterStatus = 'all';
  let searchQuery = '';
  
  const employmentTypes = [
    { value: 'full-time', label: 'Full-time' },
    { value: 'part-time', label: 'Part-time' },
    { value: 'contract', label: 'Contract' },
    { value: 'internship', label: 'Internship' }
  ];
  
  $: filteredJobs = jobs
    .filter(j => filterStatus === 'all' || j.status === filterStatus)
    .filter(j => !searchQuery || 
      j.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      j.department.toLowerCase().includes(searchQuery.toLowerCase()) ||
      j.location.toLowerCase().includes(searchQuery.toLowerCase())
    );
  
  onMount(async () => {
    await loadJobs();
    await loadProviders();
  });
  
  async function loadJobs() {
    try {
      loading = true;
      const token = localStorage.getItem('token');
      const response = await fetch('/api/recruiting/jobs', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        jobs = await response.json();
      }
    } catch (err) {
      console.error('Failed to load jobs:', err);
    } finally {
      loading = false;
    }
  }
  
  async function loadProviders() {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/recruiting/providers', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        providers = await response.json();
      }
    } catch (err) {
      console.error('Failed to load providers:', err);
    }
  }
  
  function createNewJob() {
    editingJob = {
      id: '',
      title: '',
      description: '',
      department: '',
      location: '',
      employment_type: 'full-time',
      salary_min: 0,
      salary_max: 0,
      salary_currency: 'USD',
      requirements: [''],
      responsibilities: [''],
      benefits: [''],
      status: 'draft',
      providers: [],
      applicant_count: 0,
      created_at: new Date().toISOString()
    };
    showJobModal = true;
  }
  
  function editJob(job: Job) {
    editingJob = { ...job };
    showJobModal = true;
  }
  
  async function saveJob() {
    try {
      const token = localStorage.getItem('token');
      const method = editingJob?.id ? 'PUT' : 'POST';
      const url = editingJob?.id 
        ? `/api/recruiting/jobs/${editingJob.id}`
        : '/api/recruiting/jobs';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(editingJob)
      });
      
      if (response.ok) {
        await loadJobs();
        showJobModal = false;
        editingJob = null;
        dispatch('jobUpdated');
      } else {
        const error = await response.json();
        alert('Failed to save job: ' + (error.message || 'Unknown error'));
      }
    } catch (err) {
      console.error('Failed to save job:', err);
      alert('Failed to save job: ' + err.message);
    }
  }
  
  async function postJob(job: Job) {
    if (job.providers.length === 0) {
      alert('Please select at least one provider to post this job');
      return;
    }
    
    if (!confirm(`Post "${job.title}" to ${job.providers.length} provider(s)?`)) {
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/recruiting/jobs/${job.id}/post`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ providers: job.providers })
      });
      
      if (response.ok) {
        await loadJobs();
        alert('Job posted successfully!');
        dispatch('jobUpdated');
      }
    } catch (err) {
      console.error('Failed to post job:', err);
      alert('Failed to post job: ' + err.message);
    }
  }
  
  async function closeJob(job: Job) {
    if (!confirm(`Close job posting "${job.title}"?\n\nThis will stop accepting new applicants.`)) {
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      await fetch(`/api/recruiting/jobs/${job.id}/close`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      await loadJobs();
      dispatch('jobUpdated');
    } catch (err) {
      console.error('Failed to close job:', err);
      alert('Failed to close job: ' + err.message);
    }
  }
  
  function addListItem(list: string[]) {
    list.push('');
    editingJob = editingJob; // Trigger reactivity
  }
  
  function removeListItem(list: string[], index: number) {
    list.splice(index, 1);
    editingJob = editingJob;
  }
  
  function getStatusColor(status: string) {
    const colors = {
      'draft': '#6b7280',
      'active': '#10b981',
      'closed': '#ef4444',
      'archived': '#9ca3af'
    };
    return colors[status] || colors.draft;
  }
</script>

<div class="jobs-container">
  <!-- Header -->
  <div class="header-section">
    <div>
      <h2>Job Postings</h2>
      <p class="subtitle">Create and manage job postings across multiple platforms</p>
    </div>
    <button class="btn primary" on:click={createNewJob}>
      <span class="btn-icon">➕</span>
      Create New Job
    </button>
  </div>
  
  <!-- Filters -->
  <div class="filters-section">
    <div class="search-box">
      <span class="search-icon">🔍</span>
      <input
        type="text"
        placeholder="Search jobs..."
        bind:value={searchQuery}
      />
    </div>
    
    <div class="filter-tabs">
      <button 
        class="filter-tab"
        class:active={filterStatus === 'all'}
        on:click={() => filterStatus = 'all'}
      >
        All ({jobs.length})
      </button>
      <button 
        class="filter-tab"
        class:active={filterStatus === 'draft'}
        on:click={() => filterStatus = 'draft'}
      >
        Draft ({jobs.filter(j => j.status === 'draft').length})
      </button>
      <button 
        class="filter-tab"
        class:active={filterStatus === 'active'}
        on:click={() => filterStatus = 'active'}
      >
        Active ({jobs.filter(j => j.status === 'active').length})
      </button>
      <button 
        class="filter-tab"
        class:active={filterStatus === 'closed'}
        on:click={() => filterStatus = 'closed'}
      >
        Closed ({jobs.filter(j => j.status === 'closed').length})
      </button>
    </div>
  </div>
  
  <!-- Jobs List -->
  <div class="jobs-grid">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        Loading jobs...
      </div>
    {:else if filteredJobs.length === 0}
      <div class="empty-state">
        <span class="empty-icon">📋</span>
        <h3>No Jobs Found</h3>
        <p>{searchQuery ? 'Try adjusting your search' : 'Create your first job posting to get started'}</p>
        {#if !searchQuery}
          <button class="btn primary" on:click={createNewJob}>
            <span class="btn-icon">➕</span>
            Create Job
          </button>
        {/if}
      </div>
    {:else}
      {#each filteredJobs as job}
        <div class="job-card">
          <div class="job-header">
            <div class="job-title-section">
              <h3 class="job-title">{job.title}</h3>
              <div class="job-meta">
                <span class="meta-badge">
                  <span class="meta-icon">📍</span>
                  {job.location}
                </span>
                <span class="meta-badge">
                  <span class="meta-icon">💼</span>
                  {employmentTypes.find(t => t.value === job.employment_type)?.label}
                </span>
                <span class="meta-badge">
                  <span class="meta-icon">🏢</span>
                  {job.department}
                </span>
              </div>
            </div>
            <div class="status-badge" style="background-color: {getStatusColor(job.status)}20; color: {getStatusColor(job.status)}">
              {job.status}
            </div>
          </div>
          
          <div class="job-body">
            <div class="job-salary">
              💰 ${job.salary_min.toLocaleString()} - ${job.salary_max.toLocaleString()} {job.salary_currency}
            </div>
            
            {#if job.description}
              <p class="job-description">{job.description.substring(0, 150)}{job.description.length > 150 ? '...' : ''}</p>
            {/if}
            
            <div class="job-stats">
              <div class="stat-item">
                <span class="stat-icon">👥</span>
                <span class="stat-text">{job.applicant_count} applicants</span>
              </div>
              {#if job.providers.length > 0}
                <div class="stat-item">
                  <span class="stat-icon">🔗</span>
                  <span class="stat-text">{job.providers.length} provider{job.providers.length !== 1 ? 's' : ''}</span>
                </div>
              {/if}
              <div class="stat-item">
                <span class="stat-icon">📅</span>
                <span class="stat-text">Posted {new Date(job.created_at).toLocaleDateString()}</span>
              </div>
            </div>
            
            {#if job.providers.length > 0}
              <div class="provider-badges">
                {#each job.providers as providerType}
                  {@const provider = providers.find(p => p.type === providerType)}
                  {#if provider}
                    <span class="provider-badge">{provider.name}</span>
                  {/if}
                {/each}
              </div>
            {/if}
          </div>
          
          <div class="job-actions">
            <button class="action-btn secondary" on:click={() => editJob(job)}>
              <span class="btn-icon">✏️</span>
              Edit
            </button>
            
            {#if job.status === 'draft'}
              <button class="action-btn primary" on:click={() => postJob(job)}>
                <span class="btn-icon">🚀</span>
                Post Job
              </button>
            {:else if job.status === 'active'}
              <button class="action-btn warning" on:click={() => closeJob(job)}>
                <span class="btn-icon">⏸️</span>
                Close
              </button>
            {/if}
            
            <button class="action-btn" on:click={() => window.location.href = `/recruiting/jobs/${job.id}/applicants`}>
              <span class="btn-icon">👥</span>
              View Applicants ({job.applicant_count})
            </button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<!-- Job Modal -->
{#if showJobModal && editingJob}
  <div class="modal-overlay" on:click={() => showJobModal = false}>
    <div class="modal-large" on:click|stopPropagation>
      <div class="modal-header">
        <h2>{editingJob.id ? 'Edit Job' : 'Create New Job'}</h2>
        <button class="close-btn" on:click={() => showJobModal = false}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="form-grid">
          <!-- Basic Info -->
          <div class="form-section full-width">
            <h4 class="section-title">Basic Information</h4>
            
            <div class="form-group">
              <label>Job Title *</label>
              <input
                type="text"
                bind:value={editingJob.title}
                placeholder="e.g., Senior Software Engineer"
                required
              />
            </div>
            
            <div class="form-row">
              <div class="form-group">
                <label>Department *</label>
                <input
                  type="text"
                  bind:value={editingJob.department}
                  placeholder="Engineering"
                  required
                />
              </div>
              
              <div class="form-group">
                <label>Location *</label>
                <input
                  type="text"
                  bind:value={editingJob.location}
                  placeholder="Remote / San Francisco, CA"
                  required
                />
              </div>
              
              <div class="form-group">
                <label>Employment Type *</label>
                <select bind:value={editingJob.employment_type}>
                  {#each employmentTypes as type}
                    <option value={type.value}>{type.label}</option>
                  {/each}
                </select>
              </div>
            </div>
            
            <div class="form-group">
              <label>Description *</label>
              <textarea
                bind:value={editingJob.description}
                placeholder="Describe the role, team, and company..."
                rows="5"
                required
              ></textarea>
            </div>
          </div>
          
          <!-- Compensation -->
          <div class="form-section">
            <h4 class="section-title">Compensation</h4>
            
            <div class="form-row">
              <div class="form-group">
                <label>Min Salary</label>
                <input
                  type="number"
                  bind:value={editingJob.salary_min}
                  placeholder="80000"
                />
              </div>
              
              <div class="form-group">
                <label>Max Salary</label>
                <input
                  type="number"
                  bind:value={editingJob.salary_max}
                  placeholder="120000"
                />
              </div>
              
              <div class="form-group">
                <label>Currency</label>
                <select bind:value={editingJob.salary_currency}>
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                </select>
              </div>
            </div>
          </div>
          
          <!-- Requirements -->
          <div class="form-section">
            <h4 class="section-title">Requirements</h4>
            {#each editingJob.requirements as req, i}
              <div class="list-item">
                <input
                  type="text"
                  bind:value={editingJob.requirements[i]}
                  placeholder="e.g., 5+ years of React experience"
                />
                <button class="remove-btn" on:click={() => removeListItem(editingJob.requirements, i)}>
                  ×
                </button>
              </div>
            {/each}
            <button class="add-btn" on:click={() => addListItem(editingJob.requirements)}>
              + Add Requirement
            </button>
          </div>
          
          <!-- Responsibilities -->
          <div class="form-section">
            <h4 class="section-title">Responsibilities</h4>
            {#each editingJob.responsibilities as resp, i}
              <div class="list-item">
                <input
                  type="text"
                  bind:value={editingJob.responsibilities[i]}
                  placeholder="e.g., Lead frontend architecture decisions"
                />
                <button class="remove-btn" on:click={() => removeListItem(editingJob.responsibilities, i)}>
                  ×
                </button>
              </div>
            {/each}
            <button class="add-btn" on:click={() => addListItem(editingJob.responsibilities)}>
              + Add Responsibility
            </button>
          </div>
          
          <!-- Benefits -->
          <div class="form-section full-width">
            <h4 class="section-title">Benefits</h4>
            {#each editingJob.benefits as benefit, i}
              <div class="list-item">
                <input
                  type="text"
                  bind:value={editingJob.benefits[i]}
                  placeholder="e.g., Health insurance, 401k matching"
                />
                <button class="remove-btn" on:click={() => removeListItem(editingJob.benefits, i)}>
                  ×
                </button>
              </div>
            {/each}
            <button class="add-btn" on:click={() => addListItem(editingJob.benefits)}>
              + Add Benefit
            </button>
          </div>
          
          <!-- Providers -->
          <div class="form-section full-width">
            <h4 class="section-title">Post to Providers</h4>
            <div class="provider-checkboxes">
              {#each providers as provider}
                <label class="checkbox-label">
                  <input
                    type="checkbox"
                    value={provider.type}
                    bind:group={editingJob.providers}
                  />
                  <span>{provider.name}</span>
                </label>
              {/each}
              {#if providers.length === 0}
                <p class="no-providers">No providers connected. <a href="#" on:click|preventDefault={() => {/* Navigate to providers */}}>Connect providers</a></p>
              {/if}
            </div>
          </div>
        </div>
        
        <div class="modal-actions">
          <button class="btn secondary" on:click={() => showJobModal = false}>
            Cancel
          </button>
          <button class="btn primary" on:click={saveJob}>
            <span class="btn-icon">💾</span>
            {editingJob.id ? 'Save Changes' : 'Create Job'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .jobs-container {
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
  
  .filters-section {
    background: white;
    border-radius: 12px;
    padding: 20px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    display: flex;
    gap: 16px;
    flex-wrap: wrap;
    align-items: center;
  }
  
  .search-box {
    position: relative;
    flex: 1;
    min-width: 250px;
  }
  
  .search-icon {
    position: absolute;
    left: 12px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 16px;
  }
  
  .search-box input {
    width: 100%;
    padding: 10px 12px 10px 36px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
  }
  
  .search-box input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }
  
  .filter-tabs {
    display: flex;
    gap: 8px;
  }
  
  .filter-tab {
    padding: 8px 16px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    color: #6b7280;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .filter-tab:hover {
    border-color: #3b82f6;
    color: #3b82f6;
  }
  
  .filter-tab.active {
    background: #3b82f6;
    border-color: #3b82f6;
    color: white;
  }
  
  .jobs-grid {
    display: grid;
    gap: 20px;
  }
  
  .job-card {
    background: white;
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    transition: all 0.2s;
  }
  
  .job-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }
  
  .job-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 16px;
  }
  
  .job-title {
    font-size: 20px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .job-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .meta-badge {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    background: #f3f4f6;
    border-radius: 6px;
    font-size: 13px;
    color: #6b7280;
  }
  
  .meta-icon {
    font-size: 14px;
  }
  
  .status-badge {
    padding: 6px 12px;
    border-radius: 8px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
  }
  
  .job-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 16px;
  }
  
  .job-salary {
    font-size: 16px;
    font-weight: 600;
    color: #10b981;
  }
  
  .job-description {
    color: #6b7280;
    line-height: 1.5;
    margin: 0;
  }
  
  .job-stats {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
  }
  
  .stat-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #6b7280;
  }
  
  .stat-icon {
    font-size: 14px;
  }
  
  .provider-badges {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  
  .provider-badge {
    padding: 4px 10px;
    background: #ede9fe;
    color: #5b21b6;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 500;
  }
  
  .job-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    padding-top: 16px;
    border-top: 1px solid #e5e7eb;
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
  }
  
  .action-btn.secondary {
    background: white;
    color: #374151;
  }
  
  .action-btn.secondary:hover {
    border-color: #3b82f6;
    color: #3b82f6;
  }
  
  .action-btn.primary {
    background: #3b82f6;
    border-color: #3b82f6;
    color: white;
  }
  
  .action-btn.primary:hover {
    background: #2563eb;
  }
  
  .action-btn.warning {
    background: #fef3c7;
    border-color: #f59e0b;
    color: #92400e;
  }
  
  .action-btn.warning:hover {
    background: #f59e0b;
    color: white;
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
  }
  
  .btn-icon {
    font-size: 16px;
  }
  
  .loading-state,
  .empty-state {
    text-align: center;
    padding: 60px 20px;
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
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
    margin: 0 0 20px 0;
  }
  
  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 16px;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
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
  
  .modal-large {
    background: white;
    border-radius: 12px;
    max-width: 900px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .modal-header h2 {
    font-size: 20px;
    font-weight: 600;
    margin: 0;
  }
  
  .close-btn {
    background: none;
    border: none;
    font-size: 28px;
    cursor: pointer;
    color: #6b7280;
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
  }
  
  .modal-body {
    padding: 24px;
  }
  
  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 24px;
  }
  
  .form-section {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  
  .form-section.full-width {
    grid-column: 1 / -1;
  }
  
  .section-title {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0;
    padding-bottom: 12px;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  
  .form-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 12px;
  }
  
  .form-group label {
    font-size: 13px;
    font-weight: 500;
    color: #374151;
  }
  
  .form-group input,
  .form-group select,
  .form-group textarea {
    padding: 10px 12px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    font-family: inherit;
  }
  
  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }
  
  .list-item {
    display: flex;
    gap: 8px;
  }
  
  .list-item input {
    flex: 1;
    padding: 8px 12px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
  }
  
  .remove-btn {
    padding: 8px 12px;
    background: #fee2e2;
    color: #991b1b;
    border: none;
    border-radius: 8px;
    font-size: 18px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .remove-btn:hover {
    background: #dc2626;
    color: white;
  }
  
  .add-btn {
    padding: 8px 16px;
    background: #eff6ff;
    color: #3b82f6;
    border: 1px dashed #3b82f6;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .add-btn:hover {
    background: #3b82f6;
    color: white;
  }
  
  .provider-checkboxes {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 12px;
  }
  
  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .checkbox-label:hover {
    border-color: #3b82f6;
    background: #eff6ff;
  }
  
  .checkbox-label input[type="checkbox"] {
    width: 18px;
    height: 18px;
    cursor: pointer;
  }
  
  .no-providers {
    color: #6b7280;
    font-size: 14px;
  }
  
  .no-providers a {
    color: #3b82f6;
    text-decoration: none;
  }
  
  .no-providers a:hover {
    text-decoration: underline;
  }
  
  .modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid #e5e7eb;
  }
  
  @media (max-width: 768px) {
    .form-grid {
      grid-template-columns: 1fr;
    }
    
    .filters-section {
      flex-direction: column;
      align-items: stretch;
    }
    
    .search-box {
      width: 100%;
    }
    
    .filter-tabs {
      overflow-x: auto;
      -webkit-overflow-scrolling: touch;
    }
    
    .filter-tab {
      white-space: nowrap;
    }
    
    .job-actions {
      flex-direction: column;
    }
    
    .action-btn {
      justify-content: center;
    }
  }
</style>
