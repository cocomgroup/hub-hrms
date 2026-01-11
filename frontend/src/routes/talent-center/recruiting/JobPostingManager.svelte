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
  let activeTab = 'manual'; // 'manual' or 'ai-generate'
  let aiGenerating = false;
  let generatedJobBoardVersion = '';
  let generatedInternalVersion = '';
  
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
      const data = await response.json();
      providers = Array.isArray(data) ? data : [];
    } else {
      providers = [];
    }
  } catch (err) {
    console.error('Failed to load providers:', err);
    providers = []; // Ensure it's always an array
  }
}
  
async function uploadJobFromFile(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  
  if (!file) return;
  
  // Validate file type
  const validTypes = ['.json', '.md', '.txt'];
  const fileExt = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
  
  if (!validTypes.includes(fileExt)) {
    alert(`Invalid file type. Please upload: ${validTypes.join(', ')}`);
    input.value = '';
    return;
  }
  
  try {
    // Show loading state
    loading = true;
    
    // Create FormData
    const formData = new FormData();
    formData.append('file', file);
    
    // Upload to backend for parsing
    const token = localStorage.getItem('token');
    const response = await fetch('/api/recruiting/jobs/upload', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      },
      body: formData
    });
    
    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || 'Failed to parse job file');
    }
    
    // Get parsed job data from backend
    const jobData = await response.json();
    
    // Populate the form with backend-parsed data
    editingJob = {
      id: '',
      title: jobData.title || '',
      description: jobData.description || '',
      department: jobData.department || '',
      location: jobData.location || '',
      employment_type: jobData.employment_type || 'full-time',
      salary_min: jobData.salary_min || 0,
      salary_max: jobData.salary_max || 0,
      salary_currency: jobData.salary_currency || 'USD',
      requirements: ensureArrayWithDefault(jobData.requirements),
      responsibilities: ensureArrayWithDefault(jobData.responsibilities),
      benefits: ensureArrayWithDefault(jobData.benefits),
      status: 'draft',
      providers: [],
      applicant_count: 0,
      created_at: new Date().toISOString()
    };
    
    showJobModal = true;
    
  } catch (err: any) {
    console.error('Failed to upload job file:', err);
    alert(`Failed to parse job file:\n\n${err.message}\n\nSupported formats:\n- JSON (.json)\n- Markdown (.md)\n- Text (.txt)`);
  } finally {
    loading = false;
  }
  
  // Reset input
  input.value = '';
}

async function uploadAndSaveJob(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  
  if (!file) return;
  
  // Validate file type
  const validTypes = ['.json', '.md', '.txt'];
  const fileExt = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
  
  if (!validTypes.includes(fileExt)) {
    alert(`Invalid file type. Please upload: ${validTypes.join(', ')}`);
    input.value = '';
    return;
  }
  
  // Confirm direct save
  if (!confirm(`Upload and save "${file.name}" directly to database?\n\nThe job will be created immediately as a draft.`)) {
    input.value = '';
    return;
  }
  
  try {
    // Show loading state
    loading = true;
    
    // Create FormData
    const formData = new FormData();
    formData.append('file', file);
    
    // Upload to backend with save=true parameter
    const token = localStorage.getItem('token');
    const response = await fetch('/api/recruiting/jobs/upload?save=true', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      },
      body: formData
    });
    
    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || 'Failed to create job');
    }
    
    // Get created job from database
    const createdJob = await response.json();
    
    // Refresh job list
    await loadJobs();
    
    // Show success message
    alert(`✅ Job "${createdJob.title}" created successfully!\n\nStatus: ${createdJob.status}\nID: ${createdJob.id}`);
    
    // Dispatch event
    dispatch('jobUpdated');
    
  } catch (err: any) {
    console.error('Failed to upload and save job:', err);
    alert(`Failed to create job:\n\n${err.message}`);
  } finally {
    loading = false;
  }
  
  // Reset input
  input.value = '';
}

function ensureArrayWithDefault(data: any): string[] {
  if (Array.isArray(data) && data.length > 0) {
    return data;
  }
  return [''];
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
  
  async function deleteJob(job: Job) {
    if (!confirm(`Are you sure you want to delete "${job.title}"?\n\nThis action cannot be undone.`)) {
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/recruiting/jobs/${job.id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        await loadJobs();
        dispatch('jobUpdated');
        alert('Job deleted successfully');
      } else {
        const error = await response.json();
        alert('Failed to delete job: ' + (error.message || 'Unknown error'));
      }
    } catch (err) {
      console.error('Failed to delete job:', err);
      alert('Failed to delete job: ' + err.message);
    }
  }
  
  async function toggleJobStatus(job: Job) {
    // Cycle through: draft -> active -> closed -> draft
    const statusCycle = {
      'draft': 'active',
      'active': 'closed',
      'closed': 'draft',
      'archived': 'draft'
    };
    
    const newStatus = statusCycle[job.status] || 'draft';
    
    if (!confirm(`Change status of "${job.title}" from ${job.status} to ${newStatus}?`)) {
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/recruiting/jobs/${job.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ ...job, status: newStatus })
      });
      
      if (response.ok) {
        await loadJobs();
        dispatch('jobUpdated');
      } else {
        const error = await response.json();
        alert('Failed to update status: ' + (error.message || 'Unknown error'));
      }
    } catch (err) {
      console.error('Failed to update status:', err);
      alert('Failed to update status: ' + err.message);
    }
  }
  
  async function generateAIJobDescriptions() {
    console.log('🚀 FUNCTION CALLED - generateAIJobDescriptions');
    alert('Button clicked! Check console for details.');
    
    if (!editingJob) {
      console.error('❌ No editing job');
      alert('Error: No job data available');
      return;
    }
    
    console.log('✅ EditingJob exists:', {
      title: editingJob.title,
      description: editingJob.description,
      department: editingJob.department
    });
    
    aiGenerating = true;
    generatedJobBoardVersion = '';
    generatedInternalVersion = '';
    
    try {
      const token = localStorage.getItem('token');
      
      console.log('🔑 Token check:', token ? 'Token exists' : '❌ NO TOKEN');
      
      if (!token) {
        throw new Error('No authentication token found');
      }
      
      // Prepare job data
      const jobData = {
        title: editingJob.title,
        department: editingJob.department,
        location: editingJob.location,
        employment_type: editingJob.employment_type,
        salary_min: editingJob.salary_min || 0,
        salary_max: editingJob.salary_max || 0,
        description: editingJob.description,
        requirements: editingJob.requirements.filter(r => r.trim()),
        responsibilities: editingJob.responsibilities.filter(r => r.trim()),
        benefits: editingJob.benefits.filter(b => b.trim())
      };

      console.log('📦 Job data prepared:', jobData);
      console.log('🌐 About to call: /api/ai/generate-job-description');
      
      // Generate Job Board version
      const jobBoardResponse = await fetch('/api/ai/generate-job-description', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...jobData,
          version_type: 'jobboard'
        })
      });
      
      console.log('📨 Job Board response status:', jobBoardResponse.status);
      console.log('📨 Job Board response ok:', jobBoardResponse.ok);
      
      if (!jobBoardResponse.ok) {
        const errorText = await jobBoardResponse.text();
        console.error('❌ Job Board error:', errorText);
        throw new Error(`Failed to generate job board version: ${errorText}`);
      }
      
      const jobBoardData = await jobBoardResponse.json();
      generatedJobBoardVersion = jobBoardData.generated_text;
      console.log('✅ Job Board version generated, length:', generatedJobBoardVersion.length);

      console.log('🌐 Generating Internal version...');
      
      // Generate Internal version
      const internalResponse = await fetch('/api/ai/generate-job-description', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...jobData,
          version_type: 'internal'
        })
      });
      
      console.log('📨 Internal response status:', internalResponse.status);
      
      if (!internalResponse.ok) {
        const errorText = await internalResponse.text();
        console.error('❌ Internal error:', errorText);
        throw new Error(`Failed to generate internal version: ${errorText}`);
      }
      
      const internalData = await internalResponse.json();
      generatedInternalVersion = internalData.generated_text;
      console.log('✅ Internal version generated, length:', generatedInternalVersion.length);
      
      alert('✅ Both versions generated successfully!');
      
    } catch (err) {
      console.error('💥 ERROR in generateAIJobDescriptions:', err);
      alert('❌ Failed to generate AI descriptions:\n\n' + err.message + '\n\nCheck the browser console for more details.');
    } finally {
      aiGenerating = false;
      console.log('🏁 Generation complete, aiGenerating set to false');
    }
  }
  
  function useJobBoardVersion() {
    if (editingJob && generatedJobBoardVersion) {
      editingJob.description = generatedJobBoardVersion;
      editingJob = editingJob; // Trigger reactivity
      alert('Job Board version applied to description!');
    }
  }
  
  function useInternalVersion() {
    if (editingJob && generatedInternalVersion) {
      editingJob.description = generatedInternalVersion;
      editingJob = editingJob; // Trigger reactivity
      alert('Internal HR version applied to description!');
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
<div class="header-actions">
  <button class="btn primary" on:click={createNewJob}>
    <span class="btn-icon">➕</span>
    Create New Job
  </button>

  <label class="btn secondary" for="job-file-upload-review" title="Upload file and review before saving">
    <span class="btn-icon">📁</span>
    Upload & Review
    <input
      id="job-file-upload-review"
      type="file"
      accept=".json,.txt,.md"
      style="display: none;"
      on:change={uploadJobFromFile}
    />
  </label>

  <label class="btn primary" for="job-file-upload-save" title="Upload file and save immediately">
    <span class="btn-icon">⚡</span>
    Upload & Save
    <input
      id="job-file-upload-save"
      type="file"
      accept=".json,.txt,.md"
      style="display: none;"
      on:change={uploadAndSaveJob}
    />
  </label>
</div>

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
            
            <button 
              class="action-btn status-toggle" 
              class:draft={job.status === 'draft'}
              class:active={job.status === 'active'}
              class:closed={job.status === 'closed'}
              on:click={() => toggleJobStatus(job)}
              title="Click to cycle status: {job.status} → {job.status === 'draft' ? 'active' : job.status === 'active' ? 'closed' : 'draft'}"
            >
              <span class="btn-icon">
                {#if job.status === 'draft'}📝{:else if job.status === 'active'}✅{:else if job.status === 'closed'}⏸️{:else}📋{/if}
              </span>
              {job.status.charAt(0).toUpperCase() + job.status.slice(1)}
            </button>
            
            <button class="action-btn" on:click={() => window.location.href = `/recruiting/jobs/${job.id}/applicants`}>
              <span class="btn-icon">👥</span>
              Applicants ({job.applicant_count})
            </button>
            
            <button class="action-btn danger" on:click={() => deleteJob(job)}>
              <span class="btn-icon">🗑️</span>
              Delete
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
        <div>
          <h2>{editingJob.id ? 'Edit Job' : 'Create New Job'}</h2>
          <div class="modal-tabs">
            <button 
              class="modal-tab"
              class:active={activeTab === 'manual'}
              on:click={() => activeTab = 'manual'}
            >
              📝 Manual Entry
            </button>
            <button 
              class="modal-tab"
              class:active={activeTab === 'ai-generate'}
              on:click={() => activeTab = 'ai-generate'}
            >
              ✨ AI Generate
            </button>
          </div>
        </div>
        <button class="close-btn" on:click={() => showJobModal = false}>×</button>
      </div>
      
      <div class="modal-body">
        {#if activeTab === 'manual'}
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
              {#if providers && Array.isArray(providers)}
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
              {/if}
              {#if providers && providers.length === 0}
                <p class="no-providers">No providers connected. <a href="#" on:click|preventDefault={() => {/* Navigate to providers */}}>Connect providers</a></p>
              {:else}
                <p class="no-providers">Loading providers...</p>
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
        {:else if activeTab === 'ai-generate'}
        <!-- AI Generate Tab -->
        <div class="ai-generate-container">
          {#if !editingJob.title?.trim() || !editingJob.description?.trim()}
          <!-- Prerequisites Not Met Warning -->
          <div class="prerequisites-warning">
            <div class="warning-icon">⚠️</div>
            <div class="warning-content">
              <h3>Prerequisites Required</h3>
              <p>Before you can generate AI descriptions, please fill in the following fields in the <strong>Manual Entry</strong> tab:</p>
              <ul class="warning-list">
                <li class:complete={editingJob.title?.trim()}>
                  {editingJob.title?.trim() ? '✅' : '❌'} <strong>Job Title</strong> {!editingJob.title?.trim() ? '← Required' : ''}
                </li>
                <li class:complete={editingJob.description?.trim()}>
                  {editingJob.description?.trim() ? '✅' : '❌'} <strong>Base Description</strong> {!editingJob.description?.trim() ? '← Required' : ''}
                </li>
                <li class:complete={editingJob.department?.trim()}>
                  {editingJob.department?.trim() ? '✅' : '💡'} Department (recommended)
                </li>
                <li class:complete={editingJob.requirements?.some(r => r.trim())}>
                  {editingJob.requirements?.some(r => r.trim()) ? '✅' : '💡'} At least one Requirement (recommended)
                </li>
              </ul>
              <button class="btn primary" on:click={() => activeTab = 'manual'} style="margin-top: 16px;">
                <span class="btn-icon">📝</span>
                Go to Manual Entry Tab
              </button>
            </div>
          </div>
          {:else}
          <!-- Prerequisites Met - Show Generator -->
          <div class="ai-intro">
            <h3>🤖 AI-Powered Job Description Generator</h3>
            <p>Generate two optimized versions of your job description:</p>
            <ul>
              <li><strong>Job Board Version:</strong> Concise, engaging posting for external platforms (LinkedIn, Indeed)</li>
              <li><strong>Internal HR Version:</strong> Comprehensive, detailed description for internal use</li>
            </ul>
          </div>
          
          <div class="ai-requirements">
            <h4>Prerequisites</h4>
            <div class="requirement-checklist">
              <div class="requirement-item" class:complete={editingJob.title?.trim()}>
                {editingJob.title?.trim() ? '✅' : '⬜'} Job Title
              </div>
              <div class="requirement-item" class:complete={editingJob.department?.trim()}>
                {editingJob.department?.trim() ? '✅' : '⬜'} Department
              </div>
              <div class="requirement-item" class:complete={editingJob.description?.trim()}>
                {editingJob.description?.trim() ? '✅' : '⬜'} Base Description
              </div>
              <div class="requirement-item" class:complete={editingJob.requirements?.some(r => r.trim())}>
                {editingJob.requirements?.some(r => r.trim()) ? '✅' : '⬜'} At least one Requirement
              </div>
            </div>
            <p class="requirement-note">💡 Fill in the Manual Entry tab first to get better AI-generated descriptions</p>
          </div>
          
          <!-- TEST BUTTON - Always Enabled -->
          <button 
            class="btn secondary" 
            on:click={() => {
              console.log('TEST BUTTON CLICKED');
              alert('Test button works! editingJob=' + (editingJob ? 'exists' : 'null'));
            }}
            style="margin-bottom: 10px;"
          >
            🧪 TEST: Click Me First
          </button>
          
          <button 
            class="btn primary generate-btn" 
            on:click={generateAIJobDescriptions}
            disabled={aiGenerating || !editingJob.title?.trim() || !editingJob.description?.trim()}
          >
            {#if aiGenerating}
              <span class="spinner-small"></span>
              Generating...
            {:else}
              <span class="btn-icon">✨</span>
              Generate Both Versions
            {/if}
          </button>
          
          <!-- DEBUG INFO -->
          <div style="margin-top: 10px; padding: 10px; background: #f0f0f0; border-radius: 5px; font-size: 12px;">
            <strong>Debug Info:</strong><br/>
            aiGenerating: {aiGenerating}<br/>
            title exists: {editingJob.title ? 'yes' : 'no'}<br/>
            description exists: {editingJob.description ? 'yes' : 'no'}<br/>
            Button disabled: {aiGenerating || !editingJob.title?.trim() || !editingJob.description?.trim()}
          </div>
          
          {#if generatedJobBoardVersion || generatedInternalVersion}
          <div class="generated-versions">
            <!-- Job Board Version -->
            {#if generatedJobBoardVersion}
            <div class="generated-version-card">
              <div class="version-header">
                <h4>📢 Job Board Version</h4>
                <span class="version-badge">External</span>
              </div>
              <div class="version-preview">
                {generatedJobBoardVersion}
              </div>
              <button class="btn secondary use-btn" on:click={useJobBoardVersion}>
                <span class="btn-icon">✓</span>
                Use This Version
              </button>
            </div>
            {/if}
            
            <!-- Internal HR Version -->
            {#if generatedInternalVersion}
            <div class="generated-version-card">
              <div class="version-header">
                <h4>📋 Internal HR Version</h4>
                <span class="version-badge internal">Internal</span>
              </div>
              <div class="version-preview">
                {generatedInternalVersion}
              </div>
              <button class="btn secondary use-btn" on:click={useInternalVersion}>
                <span class="btn-icon">✓</span>
                Use This Version
              </button>
            </div>
            {/if}
          </div>
          {/if}
          {/if}
        </div>
        
        <div class="modal-actions">
          <button class="btn secondary" on:click={() => showJobModal = false}>
            Cancel
          </button>
          {#if activeTab === 'manual'}
          <button class="btn primary" on:click={saveJob}>
            <span class="btn-icon">💾</span>
            {editingJob.id ? 'Save Changes' : 'Create Job'}
          </button>
          {/if}
        </div>
        {/if}
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
  
  .header-actions {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
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
  
  .action-btn.danger {
    background: #fee2e2;
    border-color: #ef4444;
    color: #991b1b;
  }
  
  .action-btn.danger:hover {
    background: #ef4444;
    color: white;
  }
  
  .action-btn.status-toggle {
    background: white;
    color: #374151;
    cursor: pointer;
    position: relative;
  }
  
  .action-btn.status-toggle:hover {
    border-color: #3b82f6;
    color: #3b82f6;
  }
  
  .action-btn.status-toggle.draft {
    background: #f3f4f6;
    border-color: #9ca3af;
    color: #4b5563;
  }
  
  .action-btn.status-toggle.active {
    background: #d1fae5;
    border-color: #10b981;
    color: #065f46;
  }
  
  .action-btn.status-toggle.closed {
    background: #fef3c7;
    border-color: #f59e0b;
    color: #92400e;
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
    align-items: start;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .modal-header h2 {
    font-size: 20px;
    font-weight: 600;
    margin: 0 0 16px 0;
  }
  
  .modal-tabs {
    display: flex;
    gap: 8px;
  }
  
  .modal-tab {
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
  
  .modal-tab:hover {
    border-color: #3b82f6;
    color: #3b82f6;
  }
  
  .modal-tab.active {
    background: #3b82f6;
    border-color: #3b82f6;
    color: white;
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
    .header-section {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  gap: 16px;
}

.header-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.upload-button-group {
  display: flex;
  gap: 8px;
  background: #f3f4f6;
  padding: 4px;
  border-radius: 8px;
}

.upload-button-group .btn {
  margin: 0;
  white-space: nowrap;
}

.btn.secondary {
  background: white;
}

.btn.primary .btn-icon {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

/* Tooltip on hover */
.upload-button-group label {
  position: relative;
}

.upload-button-group label:hover::after {
  content: attr(title);
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: #1f2937;
  color: white;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  white-space: nowrap;
  margin-bottom: 8px;
  z-index: 1000;
  pointer-events: none;
}

@media (max-width: 768px) {
  .header-section {
    flex-direction: column;
  }
  
  .header-actions {
    width: 100%;
    flex-direction: column;
  }
  
  .upload-button-group {
    width: 100%;
  }
  
  .upload-button-group .btn {
    flex: 1;
  }
}

  /* AI Generate Tab Styles */
  .ai-generate-container {
    display: flex;
    flex-direction: column;
    gap: 24px;
    max-height: 600px;
    overflow-y: auto;
  }
  
  .prerequisites-warning {
    background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
    border: 2px solid #f59e0b;
    border-radius: 12px;
    padding: 32px;
    display: flex;
    gap: 24px;
    align-items: start;
  }
  
  .warning-icon {
    font-size: 48px;
    flex-shrink: 0;
  }
  
  .warning-content h3 {
    margin: 0 0 12px 0;
    font-size: 20px;
    color: #92400e;
  }
  
  .warning-content p {
    margin: 0 0 16px 0;
    color: #78350f;
    line-height: 1.6;
  }
  
  .warning-list {
    list-style: none;
    padding: 0;
    margin: 0 0 16px 0;
  }
  
  .warning-list li {
    padding: 12px;
    background: white;
    border: 2px solid #fbbf24;
    border-radius: 8px;
    margin-bottom: 8px;
    font-size: 15px;
    color: #92400e;
    font-weight: 500;
  }
  
  .warning-list li.complete {
    border-color: #10b981;
    background: #d1fae5;
    color: #065f46;
  }
  
  .ai-intro {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 24px;
    border-radius: 12px;
  }
  
  .ai-intro h3 {
    margin: 0 0 12px 0;
    font-size: 20px;
  }
  
  .ai-intro p {
    margin: 0 0 16px 0;
    opacity: 0.95;
  }
  
  .ai-intro ul {
    margin: 0;
    padding-left: 20px;
  }
  
  .ai-intro li {
    margin-bottom: 8px;
    opacity: 0.95;
  }
  
  .ai-requirements {
    background: #f9fafb;
    padding: 20px;
    border-radius: 12px;
    border: 1px solid #e5e7eb;
  }
  
  .ai-requirements h4 {
    margin: 0 0 16px 0;
    font-size: 16px;
    color: #111827;
  }
  
  .requirement-checklist {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
    margin-bottom: 16px;
  }
  
  .requirement-item {
    padding: 8px 12px;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    font-size: 14px;
    color: #374151;
    font-weight: 500;
  }
  
  .requirement-item.complete {
    background: #d1fae5;
    border-color: #10b981;
    color: #065f46;
    font-weight: 600;
  }
  
  .requirement-note {
    margin: 0;
    font-size: 13px;
    color: #4b5563;
    font-style: italic;
  }
  
  .generate-btn {
    width: 100%;
    padding: 14px 24px;
    font-size: 16px;
    font-weight: 600;
  }
  
  .generate-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .spinner-small {
    display: inline-block;
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  
  .generated-versions {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }
  
  .generated-version-card {
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  
  .version-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 12px;
    border-bottom: 2px solid #e5e7eb;
  }
  
  .version-header h4 {
    margin: 0;
    font-size: 16px;
    color: #111827;
  }
  
  .version-badge {
    padding: 4px 12px;
    background: #dbeafe;
    color: #1e40af;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }
  
  .version-badge.internal {
    background: #fef3c7;
    color: #92400e;
  }
  
  .version-preview {
    background: #ffffff;
    padding: 16px;
    border-radius: 8px;
    border: 1px solid #e5e7eb;
    font-size: 14px;
    line-height: 1.7;
    color: #1f2937;
    max-height: 300px;
    overflow-y: auto;
    white-space: pre-wrap;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }
  
  .use-btn {
    width: 100%;
  }
  
  @media (max-width: 1024px) {
    .generated-versions {
      grid-template-columns: 1fr;
    }
  }

  }
</style>