<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';

  // Types
  interface JobPosting {
    id: string;
    title: string;
    department: string;
    location: string;
    employment_type: 'full-time' | 'part-time' | 'contract' | 'internship';
    salary_min?: number;
    salary_max?: number;
    description: string;
    requirements: string[];
    responsibilities: string[];
    benefits: string[];
    status: 'draft' | 'active' | 'closed' | 'filled';
    posted_date?: string;
    applications_count: number;
    created_at: string;
  }

  interface Candidate {
    id: string;
    job_posting_id: string;
    first_name: string;
    last_name: string;
    email: string;
    phone: string;
    resume_url?: string;
    cover_letter?: string;
    linkedin_url?: string;
    portfolio_url?: string;
    status: 'new' | 'screening' | 'interview' | 'offered' | 'rejected' | 'hired';
    score?: number;
    ai_summary?: string;
    strengths?: string[];
    weaknesses?: string[];
    experience_years?: number;
    skills: string[];
    applied_date: string;
    notes?: string;
  }

  interface JobBoard {
    id: string;
    name: string;
    logo: string;
    enabled: boolean;
    api_connected: boolean;
  }

  // State
  let activeTab: 'jobs' | 'candidates' | 'analytics' = 'jobs';
  let loading = false;
  let error = '';
  let success = '';

  // Jobs State
  let jobs: JobPosting[] = [];
  let showJobModal = false;
  let editingJob: JobPosting | null = null;
  let jobStatusFilter = 'all';

  let jobForm = {
    title: '',
    department: '',
    location: '',
    employment_type: 'full-time' as 'full-time' | 'part-time' | 'contract' | 'internship',
    salary_min: 0,
    salary_max: 0,
    description: '',
    requirements: [] as string[],
    responsibilities: [] as string[],
    benefits: [] as string[]
  };

  // Candidates State
  let candidates: Candidate[] = [];
  let selectedJob: JobPosting | null = null;
  let selectedCandidate: Candidate | null = null;
  let showCandidateDetail = false;
  let showResumeAnalysis = false;
  let candidateStatusFilter = 'all';
  let analyzingResume = false;

  // Job Boards State
  let showJobBoardsModal = false;
  let jobBoards: JobBoard[] = [
    { id: '1', name: 'LinkedIn', logo: '💼', enabled: false, api_connected: false },
    { id: '2', name: 'Indeed', logo: '🔍', enabled: false, api_connected: false },
    { id: '3', name: 'Glassdoor', logo: '🏢', enabled: false, api_connected: false },
    { id: '4', name: 'ZipRecruiter', logo: '📮', enabled: false, api_connected: false },
    { id: '5', name: 'Monster', logo: '👹', enabled: false, api_connected: false }
  ];

  // Email State
  let showEmailModal = false;
  let emailTemplate = '';
  let emailSubject = '';
  let emailBody = '';
  let generatingEmail = false;

  const emailTemplates = {
    screening: {
      subject: 'Next Steps - {{job_title}} Position at {{company}}',
      body: `Dear {{candidate_name}},

Thank you for your application for the {{job_title}} position. We've reviewed your resume and would like to proceed with the next step in our hiring process.

We'd like to schedule a {{interview_type}} to discuss your qualifications and learn more about your experience.

Please let us know your availability for a {{duration}} conversation in the next week.

Best regards,
{{recruiter_name}}
{{company}}`
    },
    interview: {
      subject: 'Interview Invitation - {{job_title}} Position',
      body: `Dear {{candidate_name}},

We're pleased to invite you to interview for the {{job_title}} position at {{company}}.

Interview Details:
- Date & Time: {{interview_datetime}}
- Duration: {{duration}}
- Format: {{interview_format}}
- Location/Link: {{location_or_link}}

Please confirm your attendance by {{confirmation_deadline}}.

We look forward to speaking with you!

Best regards,
{{recruiter_name}}`
    },
    offer: {
      subject: '🎉 Job Offer - {{job_title}} at {{company}}',
      body: `Dear {{candidate_name}},

Congratulations! We're thrilled to extend an offer for the {{job_title}} position at {{company}}.

Offer Details:
- Start Date: {{start_date}}
- Salary: {{salary}}
- Benefits: {{benefits_summary}}

Please review the attached formal offer letter. We'd appreciate your response by {{response_deadline}}.

Welcome to the team!

Best regards,
{{recruiter_name}}`
    },
    rejection: {
      subject: 'Update on Your Application - {{job_title}}',
      body: `Dear {{candidate_name}},

Thank you for taking the time to apply for the {{job_title}} position at {{company}} and for your interest in joining our team.

After careful consideration, we've decided to move forward with other candidates whose qualifications more closely match our current needs.

We were impressed by your background in {{candidate_strength}}, and we encourage you to apply for future positions that align with your skills.

We wish you the best in your job search.

Best regards,
{{recruiter_name}}`
    }
  };

  // Computed
  $: filteredJobs = jobs.filter(job => {
    return jobStatusFilter === 'all' || job.status === jobStatusFilter;
  });

  $: filteredCandidates = selectedJob 
    ? candidates.filter(c => {
        const matchesJob = c.job_posting_id === selectedJob.id;
        const matchesStatus = candidateStatusFilter === 'all' || c.status === candidateStatusFilter;
        return matchesJob && matchesStatus;
      })
    : [];

  $: jobStats = {
    total: jobs.length,
    active: jobs.filter(j => j.status === 'active').length,
    filled: jobs.filter(j => j.status === 'filled').length,
    total_applicants: jobs.reduce((sum, j) => sum + j.applications_count, 0)
  };

  $: candidateStats = selectedJob ? {
    total: filteredCandidates.length,
    new: filteredCandidates.filter(c => c.status === 'new').length,
    screening: filteredCandidates.filter(c => c.status === 'screening').length,
    interview: filteredCandidates.filter(c => c.status === 'interview').length,
    offered: filteredCandidates.filter(c => c.status === 'offered').length
  } : null;

  // API Calls - Jobs
  async function loadJobs() {
    try {
      loading = true;
      error = '';

      const response = await fetch('/api/recruiting/jobs', {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) throw new Error('Failed to load jobs');

      jobs = await response.json();
    } catch (err: any) {
      error = err.message;
      jobs = [];
    } finally {
      loading = false;
    }
  }

  async function createJob() {
    try {
      loading = true;
      error = '';
      success = '';

      const response = await fetch('/api/recruiting/jobs', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(jobForm)
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to create job');
      }

      success = 'Job posting created successfully';
      showJobModal = false;
      resetJobForm();
      await loadJobs();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function updateJob() {
    if (!editingJob) return;

    try {
      loading = true;
      error = '';
      success = '';

      const response = await fetch(`/api/recruiting/jobs/${editingJob.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(jobForm)
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to update job');
      }

      success = 'Job posting updated successfully';
      showJobModal = false;
      editingJob = null;
      resetJobForm();
      await loadJobs();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function postToJobBoards(jobId: string) {
    showJobBoardsModal = true;
    // In production, this would integrate with job board APIs
  }

  function openJobModal(job?: JobPosting) {
    if (job) {
      editingJob = job;
      jobForm = {
        title: job.title,
        department: job.department,
        location: job.location,
        employment_type: job.employment_type,
        salary_min: job.salary_min || 0,
        salary_max: job.salary_max || 0,
        description: job.description,
        requirements: [...job.requirements],
        responsibilities: [...job.responsibilities],
        benefits: [...job.benefits]
      };
    }
    showJobModal = true;
  }

  function resetJobForm() {
    jobForm = {
      title: '',
      department: '',
      location: '',
      employment_type: 'full-time',
      salary_min: 0,
      salary_max: 0,
      description: '',
      requirements: [],
      responsibilities: [],
      benefits: []
    };
    editingJob = null;
  }

  // API Calls - Candidates
  async function loadCandidates(jobId: string) {
    try {
      loading = true;
      error = '';

      const response = await fetch(`/api/recruiting/jobs/${jobId}/candidates`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) throw new Error('Failed to load candidates');

      candidates = await response.json();
    } catch (err: any) {
      error = err.message;
      candidates = [];
    } finally {
      loading = false;
    }
  }

  async function analyzeResume(candidate: Candidate) {
    try {
      analyzingResume = true;
      error = '';

      const response = await fetch(`/api/recruiting/candidates/${candidate.id}/analyze`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) throw new Error('Failed to analyze resume');

      const analysis = await response.json();
      
      // Update candidate with AI analysis
      selectedCandidate = {
        ...candidate,
        score: analysis.score,
        ai_summary: analysis.summary,
        strengths: analysis.strengths,
        weaknesses: analysis.weaknesses,
        experience_years: analysis.experience_years,
        skills: analysis.skills
      };

      showResumeAnalysis = true;
    } catch (err: any) {
      error = err.message;
    } finally {
      analyzingResume = false;
    }
  }

  async function updateCandidateStatus(candidateId: string, status: string) {
    try {
      loading = true;
      error = '';
      success = '';

      const response = await fetch(`/api/recruiting/candidates/${candidateId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status })
      });

      if (!response.ok) throw new Error('Failed to update candidate');

      success = 'Candidate status updated';
      if (selectedJob) {
        await loadCandidates(selectedJob.id);
      }
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function viewCandidateDetail(candidate: Candidate) {
    selectedCandidate = candidate;
    showCandidateDetail = true;
  }

  function selectJob(job: JobPosting) {
    selectedJob = job;
    loadCandidates(job.id);
  }

  // Email Generation
  function generateEmail(candidate: Candidate, template: string) {
    const templates = emailTemplates[template];
    if (!templates) return;

    emailTemplate = template;
    emailSubject = templates.subject
      .replace('{{job_title}}', selectedJob?.title || '')
      .replace('{{company}}', 'Your Company');

    emailBody = templates.body
      .replace(/{{candidate_name}}/g, `${candidate.first_name} ${candidate.last_name}`)
      .replace(/{{job_title}}/g, selectedJob?.title || '')
      .replace(/{{company}}/g, 'Your Company')
      .replace(/{{recruiter_name}}/g, $authStore.user?.name || 'Hiring Team')
      .replace(/{{candidate_strength}}/g, candidate.strengths?.[0] || 'your experience');

    showEmailModal = true;
  }

  async function generateAIEmail(candidate: Candidate, context: string) {
    try {
      generatingEmail = true;
      error = '';

      const response = await fetch('/api/recruiting/email/generate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          candidate_id: candidate.id,
          job_id: selectedJob?.id,
          context: context,
          tone: 'professional_friendly'
        })
      });

      if (!response.ok) throw new Error('Failed to generate email');

      const result = await response.json();
      emailSubject = result.subject;
      emailBody = result.body;
      showEmailModal = true;
    } catch (err: any) {
      error = err.message;
    } finally {
      generatingEmail = false;
    }
  }

  async function sendEmail() {
    if (!selectedCandidate) return;

    try {
      loading = true;
      error = '';
      success = '';

      const response = await fetch('/api/recruiting/email/send', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          candidate_id: selectedCandidate.id,
          subject: emailSubject,
          body: emailBody
        })
      });

      if (!response.ok) throw new Error('Failed to send email');

      success = 'Email sent successfully';
      showEmailModal = false;
      resetEmailForm();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function resetEmailForm() {
    emailTemplate = '';
    emailSubject = '';
    emailBody = '';
  }

  // Utility Functions
  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0
    }).format(amount);
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return 'N/A';
    return new Date(dateStr).toLocaleDateString();
  }

  function getStatusBadgeClass(status: string): string {
    const classes = {
      draft: 'badge-ghost',
      active: 'badge-success',
      closed: 'badge-error',
      filled: 'badge-info',
      new: 'badge-warning',
      screening: 'badge-info',
      interview: 'badge-primary',
      offered: 'badge-success',
      rejected: 'badge-error',
      hired: 'badge-success'
    };
    return classes[status] || 'badge-ghost';
  }

  function getScoreColor(score: number): string {
    if (score >= 80) return '#10b981';
    if (score >= 60) return '#3b82f6';
    if (score >= 40) return '#f59e0b';
    return '#ef4444';
  }

  function addArrayItem(array: string[], value: string) {
    if (value.trim()) {
      array.push(value.trim());
      return array;
    }
    return array;
  }

  function removeArrayItem(array: string[], index: number) {
    array.splice(index, 1);
    return array;
  }

  onMount(() => {
    loadJobs();
  });
</script>

<div class="recruiting-container">
  <!-- Header -->
  <div class="recruiting-header">
    <h1>🎯 Recruiting</h1>
    <p class="text-muted">Post jobs, review candidates, and manage hiring pipeline</p>
  </div>

  <!-- Alerts -->
  {#if error}
    <div class="alert alert-error">
      <span>{error}</span>
      <button on:click={() => error = ''}>✕</button>
    </div>
  {/if}

  {#if success}
    <div class="alert alert-success">
      <span>{success}</span>
      <button on:click={() => success = ''}>✕</button>
    </div>
  {/if}

  <!-- Tabs -->
  <div class="tabs">
    <button 
      class="tab {activeTab === 'jobs' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'jobs'}>
      📋 Job Postings
    </button>
    <button 
      class="tab {activeTab === 'candidates' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'candidates'}>
      👥 Candidates
    </button>
    <button 
      class="tab {activeTab === 'analytics' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'analytics'}>
      📊 Analytics
    </button>
  </div>

  <!-- Content -->
  <div class="tab-content">
    <!-- Jobs Tab -->
    {#if activeTab === 'jobs'}
      <div class="jobs-section">
        <div class="section-header">
          <div>
            <h2>Job Postings</h2>
            <p class="text-muted">Create and manage open positions</p>
          </div>
          <button class="btn btn-primary" on:click={() => openJobModal()}>
            + Post New Job
          </button>
        </div>

        <!-- Stats -->
        <div class="stats-grid">
          <div class="stat-card">
            <span class="stat-icon">📋</span>
            <div class="stat-info">
              <div class="stat-value">{jobStats.total}</div>
              <div class="stat-label">Total Jobs</div>
            </div>
          </div>
          <div class="stat-card">
            <span class="stat-icon">✅</span>
            <div class="stat-info">
              <div class="stat-value">{jobStats.active}</div>
              <div class="stat-label">Active</div>
            </div>
          </div>
          <div class="stat-card">
            <span class="stat-icon">🎉</span>
            <div class="stat-info">
              <div class="stat-value">{jobStats.filled}</div>
              <div class="stat-label">Filled</div>
            </div>
          </div>
          <div class="stat-card">
            <span class="stat-icon">👥</span>
            <div class="stat-info">
              <div class="stat-value">{jobStats.total_applicants}</div>
              <div class="stat-label">Total Applicants</div>
            </div>
          </div>
        </div>

        <!-- Filter -->
        <div class="filters">
          <select bind:value={jobStatusFilter} class="select select-sm">
            <option value="all">All Status</option>
            <option value="draft">Draft</option>
            <option value="active">Active</option>
            <option value="closed">Closed</option>
            <option value="filled">Filled</option>
          </select>
        </div>

        <!-- Jobs List -->
        {#if loading}
          <div class="loading">Loading jobs...</div>
        {:else if filteredJobs.length === 0}
          <div class="empty-state">
            <span class="empty-icon">📋</span>
            <p>No job postings found</p>
            <button class="btn btn-primary" on:click={() => openJobModal()}>
              Post Your First Job
            </button>
          </div>
        {:else}
          <div class="jobs-grid">
            {#each filteredJobs as job}
              <div class="job-card">
                <div class="job-header">
                  <div>
                    <h3 class="job-title">{job.title}</h3>
                    <p class="job-meta">{job.department} • {job.location}</p>
                  </div>
                  <span class="badge {getStatusBadgeClass(job.status)}">
                    {job.status}
                  </span>
                </div>

                <div class="job-details">
                  <span class="job-detail">
                    <span class="detail-icon">💼</span>
                    {job.employment_type.replace('-', ' ')}
                  </span>
                  {#if job.salary_min && job.salary_max}
                    <span class="job-detail">
                      <span class="detail-icon">💰</span>
                      {formatCurrency(job.salary_min)} - {formatCurrency(job.salary_max)}
                    </span>
                  {/if}
                  <span class="job-detail">
                    <span class="detail-icon">📅</span>
                    Posted {formatDate(job.posted_date || job.created_at)}
                  </span>
                </div>

                <div class="job-applicants">
                  <span class="applicants-count">{job.applications_count}</span>
                  <span class="applicants-label">Applicants</span>
                </div>

                <div class="job-actions">
                  <button class="btn btn-sm btn-primary" on:click={() => selectJob(job)}>
                    View Candidates
                  </button>
                  <button class="btn btn-sm btn-ghost" on:click={() => openJobModal(job)}>
                    ✏️ Edit
                  </button>
                  <button class="btn btn-sm btn-ghost" on:click={() => postToJobBoards(job.id)}>
                    📤 Post to Boards
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    <!-- Candidates Tab -->
    {#if activeTab === 'candidates'}
      <div class="candidates-section">
        {#if selectedJob}
          <div class="section-header">
            <div>
              <h2>Candidates for {selectedJob.title}</h2>
              <p class="text-muted">{selectedJob.department} • {selectedJob.location}</p>
            </div>
            <button class="btn btn-ghost" on:click={() => { selectedJob = null; candidates = []; }}>
              ← Back to Jobs
            </button>
          </div>

          <!-- Candidate Stats -->
          {#if candidateStats}
            <div class="stats-grid">
              <div class="stat-card">
                <span class="stat-icon">👥</span>
                <div class="stat-info">
                  <div class="stat-value">{candidateStats.total}</div>
                  <div class="stat-label">Total</div>
                </div>
              </div>
              <div class="stat-card">
                <span class="stat-icon">🆕</span>
                <div class="stat-info">
                  <div class="stat-value">{candidateStats.new}</div>
                  <div class="stat-label">New</div>
                </div>
              </div>
              <div class="stat-card">
                <span class="stat-icon">🔍</span>
                <div class="stat-info">
                  <div class="stat-value">{candidateStats.screening}</div>
                  <div class="stat-label">Screening</div>
                </div>
              </div>
              <div class="stat-card">
                <span class="stat-icon">💬</span>
                <div class="stat-info">
                  <div class="stat-value">{candidateStats.interview}</div>
                  <div class="stat-label">Interview</div>
                </div>
              </div>
              <div class="stat-card">
                <span class="stat-icon">🎁</span>
                <div class="stat-info">
                  <div class="stat-value">{candidateStats.offered}</div>
                  <div class="stat-label">Offered</div>
                </div>
              </div>
            </div>
          {/if}

          <!-- Filter -->
          <div class="filters">
            <select bind:value={candidateStatusFilter} class="select select-sm">
              <option value="all">All Status</option>
              <option value="new">New</option>
              <option value="screening">Screening</option>
              <option value="interview">Interview</option>
              <option value="offered">Offered</option>
              <option value="rejected">Rejected</option>
              <option value="hired">Hired</option>
            </select>
          </div>

          <!-- Candidates List -->
          {#if loading}
            <div class="loading">Loading candidates...</div>
          {:else if filteredCandidates.length === 0}
            <div class="empty-state">
              <span class="empty-icon">👥</span>
              <p>No candidates found for this position</p>
            </div>
          {:else}
            <div class="candidates-grid">
              {#each filteredCandidates as candidate}
                <div class="candidate-card">
                  <div class="candidate-header">
                    <div class="candidate-avatar">
                      {candidate.first_name[0]}{candidate.last_name[0]}
                    </div>
                    <div class="candidate-info">
                      <h3 class="candidate-name">{candidate.first_name} {candidate.last_name}</h3>
                      <p class="candidate-meta">{candidate.email}</p>
                      {#if candidate.phone}
                        <p class="candidate-meta">{candidate.phone}</p>
                      {/if}
                    </div>
                  </div>

                  {#if candidate.score !== undefined}
                    <div class="candidate-score">
                      <div class="score-circle" style="border-color: {getScoreColor(candidate.score)}">
                        <span class="score-value" style="color: {getScoreColor(candidate.score)}">{candidate.score}</span>
                      </div>
                      <span class="score-label">AI Match Score</span>
                    </div>
                  {/if}

                  <div class="candidate-skills">
                    {#each candidate.skills.slice(0, 3) as skill}
                      <span class="skill-badge">{skill}</span>
                    {/each}
                    {#if candidate.skills.length > 3}
                      <span class="skill-badge">+{candidate.skills.length - 3}</span>
                    {/if}
                  </div>

                  <div class="candidate-meta-row">
                    <span class="badge {getStatusBadgeClass(candidate.status)}">
                      {candidate.status}
                    </span>
                    <span class="candidate-date">
                      Applied {formatDate(candidate.applied_date)}
                    </span>
                  </div>

                  <div class="candidate-actions">
                    <button class="btn btn-sm btn-primary" on:click={() => viewCandidateDetail(candidate)}>
                      View Profile
                    </button>
                    <button 
                      class="btn btn-sm btn-ghost" 
                      on:click={() => analyzeResume(candidate)}
                      disabled={analyzingResume}>
                      {analyzingResume ? '🔄' : '🤖'} AI Review
                    </button>
                    <button class="btn btn-sm btn-ghost" on:click={() => generateEmail(candidate, 'screening')}>
                      ✉️ Email
                    </button>
                  </div>

                  <div class="candidate-status-select">
                    <select
                      value={candidate.status}
                      on:change={(e) => updateCandidateStatus(candidate.id, e.currentTarget.value)}
                      class="select select-sm w-full">
                      <option value="new">New</option>
                      <option value="screening">Screening</option>
                      <option value="interview">Interview</option>
                      <option value="offered">Offered</option>
                      <option value="rejected">Rejected</option>
                      <option value="hired">Hired</option>
                    </select>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        {:else}
          <div class="empty-state">
            <span class="empty-icon">📋</span>
            <p>Select a job posting to view candidates</p>
            <button class="btn btn-primary" on:click={() => activeTab = 'jobs'}>
              View Job Postings
            </button>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Analytics Tab -->
    {#if activeTab === 'analytics'}
      <div class="analytics-section">
        <div class="section-header">
          <h2>Recruiting Analytics</h2>
          <p class="text-muted">Track hiring performance and metrics</p>
        </div>

        <div class="analytics-grid">
          <div class="analytics-card">
            <h3>Time to Hire</h3>
            <div class="analytics-value">24 days</div>
            <div class="analytics-trend text-success">↓ 15% vs last month</div>
          </div>

          <div class="analytics-card">
            <h3>Applications per Job</h3>
            <div class="analytics-value">
              {jobStats.total > 0 ? Math.round(jobStats.total_applicants / jobStats.total) : 0}
            </div>
            <div class="analytics-trend">Average</div>
          </div>

          <div class="analytics-card">
            <h3>Interview Rate</h3>
            <div class="analytics-value">32%</div>
            <div class="analytics-trend text-success">↑ 8% vs last month</div>
          </div>

          <div class="analytics-card">
            <h3>Offer Acceptance</h3>
            <div class="analytics-value">85%</div>
            <div class="analytics-trend">Above average</div>
          </div>
        </div>

        <div class="empty-state">
          <span class="empty-icon">📊</span>
          <p>More detailed analytics coming soon...</p>
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Job Modal -->
{#if showJobModal}
  <div class="modal" on:click={() => { showJobModal = false; resetJobForm(); }}>
    <div class="modal-box max-w-4xl" on:click|stopPropagation>
      <div class="modal-header">
        <h2>{editingJob ? 'Edit' : 'Create'} Job Posting</h2>
        <button class="btn btn-circle btn-sm" on:click={() => { showJobModal = false; resetJobForm(); }}>✕</button>
      </div>

      <form on:submit|preventDefault={editingJob ? updateJob : createJob} class="form">
        <div class="form-row">
          <div class="form-control">
            <label class="label">
              <span class="label-text">Job Title</span>
            </label>
            <input 
              type="text" 
              bind:value={jobForm.title} 
              class="input w-full" 
              required 
              placeholder="e.g., Senior Software Engineer"
            />
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">Department</span>
            </label>
            <input 
              type="text" 
              bind:value={jobForm.department} 
              class="input w-full" 
              required 
              placeholder="e.g., Engineering"
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-control">
            <label class="label">
              <span class="label-text">Location</span>
            </label>
            <input 
              type="text" 
              bind:value={jobForm.location} 
              class="input w-full" 
              required 
              placeholder="e.g., San Francisco, CA / Remote"
            />
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">Employment Type</span>
            </label>
            <select bind:value={jobForm.employment_type} class="select w-full" required>
              <option value="full-time">Full-Time</option>
              <option value="part-time">Part-Time</option>
              <option value="contract">Contract</option>
              <option value="internship">Internship</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-control">
            <label class="label">
              <span class="label-text">Minimum Salary</span>
            </label>
            <input 
              type="number" 
              bind:value={jobForm.salary_min} 
              class="input w-full" 
              min="0"
              step="1000"
            />
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">Maximum Salary</span>
            </label>
            <input 
              type="number" 
              bind:value={jobForm.salary_max} 
              class="input w-full" 
              min="0"
              step="1000"
            />
          </div>
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Job Description</span>
          </label>
          <textarea 
            bind:value={jobForm.description} 
            class="textarea w-full" 
            rows="5"
            required
            placeholder="Describe the role..."
          ></textarea>
        </div>

        <!-- Requirements -->
        <div class="form-control">
          <label class="label">
            <span class="label-text">Requirements</span>
          </label>
          <div class="array-input">
            {#each jobForm.requirements as req, i}
              <div class="array-item">
                <input type="text" bind:value={jobForm.requirements[i]} class="input w-full" />
                <button type="button" class="btn btn-sm btn-ghost" on:click={() => jobForm.requirements = removeArrayItem(jobForm.requirements, i)}>
                  🗑️
                </button>
              </div>
            {/each}
            <button 
              type="button" 
              class="btn btn-sm btn-ghost" 
              on:click={() => jobForm.requirements = [...jobForm.requirements, '']}>
              + Add Requirement
            </button>
          </div>
        </div>

        <!-- Responsibilities -->
        <div class="form-control">
          <label class="label">
            <span class="label-text">Responsibilities</span>
          </label>
          <div class="array-input">
            {#each jobForm.responsibilities as resp, i}
              <div class="array-item">
                <input type="text" bind:value={jobForm.responsibilities[i]} class="input w-full" />
                <button type="button" class="btn btn-sm btn-ghost" on:click={() => jobForm.responsibilities = removeArrayItem(jobForm.responsibilities, i)}>
                  🗑️
                </button>
              </div>
            {/each}
            <button 
              type="button" 
              class="btn btn-sm btn-ghost" 
              on:click={() => jobForm.responsibilities = [...jobForm.responsibilities, '']}>
              + Add Responsibility
            </button>
          </div>
        </div>

        <div class="modal-actions">
          <button type="submit" class="btn btn-primary" disabled={loading}>
            {loading ? 'Saving...' : editingJob ? 'Update Job' : 'Post Job'}
          </button>
          <button type="button" class="btn btn-ghost" on:click={() => { showJobModal = false; resetJobForm(); }}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Candidate Detail Modal -->
{#if showCandidateDetail && selectedCandidate}
  <div class="modal" on:click={() => showCandidateDetail = false}>
    <div class="modal-box max-w-4xl" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{selectedCandidate.first_name} {selectedCandidate.last_name}</h2>
          <p class="text-muted">{selectedCandidate.email}</p>
        </div>
        <button class="btn btn-circle btn-sm" on:click={() => showCandidateDetail = false}>✕</button>
      </div>

      <div class="candidate-detail">
        <!-- Contact Info -->
        <div class="detail-section">
          <h3>Contact Information</h3>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">Email</span>
              <span class="info-value">{selectedCandidate.email}</span>
            </div>
            {#if selectedCandidate.phone}
              <div class="info-item">
                <span class="info-label">Phone</span>
                <span class="info-value">{selectedCandidate.phone}</span>
              </div>
            {/if}
            {#if selectedCandidate.linkedin_url}
              <div class="info-item">
                <span class="info-label">LinkedIn</span>
                <a href={selectedCandidate.linkedin_url} target="_blank" class="link">View Profile</a>
              </div>
            {/if}
            {#if selectedCandidate.portfolio_url}
              <div class="info-item">
                <span class="info-label">Portfolio</span>
                <a href={selectedCandidate.portfolio_url} target="_blank" class="link">View Portfolio</a>
              </div>
            {/if}
          </div>
        </div>

        <!-- Skills -->
        {#if selectedCandidate.skills.length > 0}
          <div class="detail-section">
            <h3>Skills</h3>
            <div class="skills-list">
              {#each selectedCandidate.skills as skill}
                <span class="skill-badge-large">{skill}</span>
              {/each}
            </div>
          </div>
        {/if}

        <!-- AI Analysis -->
        {#if selectedCandidate.ai_summary}
          <div class="detail-section">
            <h3>AI Analysis</h3>
            <div class="ai-analysis">
              <p>{selectedCandidate.ai_summary}</p>
              
              {#if selectedCandidate.strengths && selectedCandidate.strengths.length > 0}
                <div class="analysis-category">
                  <h4>✅ Strengths</h4>
                  <ul>
                    {#each selectedCandidate.strengths as strength}
                      <li>{strength}</li>
                    {/each}
                  </ul>
                </div>
              {/if}

              {#if selectedCandidate.weaknesses && selectedCandidate.weaknesses.length > 0}
                <div class="analysis-category">
                  <h4>⚠️ Areas for Consideration</h4>
                  <ul>
                    {#each selectedCandidate.weaknesses as weakness}
                      <li>{weakness}</li>
                    {/each}
                  </ul>
                </div>
              {/if}
            </div>
          </div>
        {/if}

        <!-- Cover Letter -->
        {#if selectedCandidate.cover_letter}
          <div class="detail-section">
            <h3>Cover Letter</h3>
            <div class="cover-letter">
              {selectedCandidate.cover_letter}
            </div>
          </div>
        {/if}

        <!-- Actions -->
        <div class="detail-actions">
          {#if selectedCandidate.resume_url}
            <a href={selectedCandidate.resume_url} target="_blank" class="btn btn-primary">
              📄 View Resume
            </a>
          {/if}
          <button class="btn btn-ghost" on:click={() => analyzeResume(selectedCandidate)}>
            🤖 AI Analysis
          </button>
          <button class="btn btn-ghost" on:click={() => generateEmail(selectedCandidate, 'interview')}>
            ✉️ Send Email
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Email Modal -->
{#if showEmailModal}
  <div class="modal" on:click={() => { showEmailModal = false; resetEmailForm(); }}>
    <div class="modal-box max-w-3xl" on:click|stopPropagation>
      <div class="modal-header">
        <h2>Compose Email</h2>
        <button class="btn btn-circle btn-sm" on:click={() => { showEmailModal = false; resetEmailForm(); }}>✕</button>
      </div>

      <div class="email-templates">
        <button class="btn btn-sm btn-ghost" on:click={() => generateEmail(selectedCandidate, 'screening')}>
          📧 Screening
        </button>
        <button class="btn btn-sm btn-ghost" on:click={() => generateEmail(selectedCandidate, 'interview')}>
          💬 Interview
        </button>
        <button class="btn btn-sm btn-ghost" on:click={() => generateEmail(selectedCandidate, 'offer')}>
          🎁 Offer
        </button>
        <button class="btn btn-sm btn-ghost" on:click={() => generateEmail(selectedCandidate, 'rejection')}>
          ✉️ Rejection
        </button>
        <button 
          class="btn btn-sm btn-primary" 
          on:click={() => generateAIEmail(selectedCandidate, 'custom')}
          disabled={generatingEmail}>
          {generatingEmail ? '🔄 Generating...' : '🤖 AI Generate'}
        </button>
      </div>

      <form on:submit|preventDefault={sendEmail} class="form">
        <div class="form-control">
          <label class="label">
            <span class="label-text">To: {selectedCandidate?.email}</span>
          </label>
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Subject</span>
          </label>
          <input 
            type="text" 
            bind:value={emailSubject} 
            class="input w-full" 
            required 
          />
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Message</span>
          </label>
          <textarea 
            bind:value={emailBody} 
            class="textarea w-full" 
            rows="15"
            required
          ></textarea>
        </div>

        <div class="modal-actions">
          <button type="submit" class="btn btn-primary" disabled={loading}>
            {loading ? 'Sending...' : '📤 Send Email'}
          </button>
          <button type="button" class="btn btn-ghost" on:click={() => { showEmailModal = false; resetEmailForm(); }}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Job Boards Modal -->
{#if showJobBoardsModal}
  <div class="modal" on:click={() => showJobBoardsModal = false}>
    <div class="modal-box" on:click|stopPropagation>
      <div class="modal-header">
        <h2>Post to Job Boards</h2>
        <button class="btn btn-circle btn-sm" on:click={() => showJobBoardsModal = false}>✕</button>
      </div>

      <div class="job-boards-list">
        {#each jobBoards as board}
          <div class="job-board-item">
            <div class="board-info">
              <span class="board-logo">{board.logo}</span>
              <span class="board-name">{board.name}</span>
            </div>
            <label class="switch">
              <input type="checkbox" bind:checked={board.enabled} />
              <span class="slider"></span>
            </label>
          </div>
        {/each}
      </div>

      <div class="modal-actions">
        <button class="btn btn-primary" on:click={() => showJobBoardsModal = false}>
          Post Selected
        </button>
        <button class="btn btn-ghost" on:click={() => showJobBoardsModal = false}>
          Cancel
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Previous styles from Onboarding.svelte remain the same through line ~900 */
  /* Adding new styles specific to recruiting */

  .recruiting-container {
    padding: 2rem;
    max-width: 1600px;
    margin: 0 auto;
  }

  .recruiting-header {
    margin-bottom: 2rem;
  }

  .recruiting-header h1 {
    font-size: 2rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 0.5rem;
  }

  .text-muted {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .text-success {
    color: #059669;
  }

  .alert {
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .alert-error {
    background: #fef2f2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .alert-success {
    background: #f0fdf4;
    color: #166534;
    border: 1px solid #bbf7d0;
  }

  .alert button {
    background: none;
    border: none;
    font-size: 1.25rem;
    cursor: pointer;
  }

  .tabs {
    display: flex;
    gap: 0.5rem;
    border-bottom: 2px solid #e5e7eb;
    margin-bottom: 2rem;
  }

  .tab {
    padding: 0.75rem 1.5rem;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    color: #6b7280;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: -2px;
  }

  .tab:hover {
    color: #111827;
  }

  .tab-active {
    color: #3b82f6;
    border-bottom-color: #3b82f6;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .section-header h2 {
    font-size: 1.5rem;
    font-weight: 700;
    color: #111827;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .stat-card {
    display: flex;
    align-items: center;
    gap: 1rem;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    padding: 1.5rem;
  }

  .stat-icon {
    font-size: 2rem;
  }

  .stat-info {
    display: flex;
    flex-direction: column;
  }

  .stat-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: #111827;
  }

  .stat-label {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .loading {
    text-align: center;
    padding: 3rem;
    color: #6b7280;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #6b7280;
  }

  .empty-icon {
    font-size: 4rem;
    display: block;
    margin-bottom: 1rem;
  }

  .filters {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .jobs-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
    gap: 1.5rem;
  }

  .job-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
    transition: all 0.2s;
  }

  .job-card:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  .job-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
  }

  .job-title {
    font-size: 1.125rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 0.25rem;
  }

  .job-meta {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .job-details {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .job-detail {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: #6b7280;
  }

  .detail-icon {
    font-size: 1rem;
  }

  .job-applicants {
    background: #f0f9ff;
    padding: 1rem;
    border-radius: 0.5rem;
    text-align: center;
    margin-bottom: 1rem;
  }

  .applicants-count {
    display: block;
    font-size: 2rem;
    font-weight: 700;
    color: #0284c7;
  }

  .applicants-label {
    font-size: 0.875rem;
    color: #075985;
  }

  .job-actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .candidates-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }

  .candidate-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
    transition: all 0.2s;
  }

  .candidate-card:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  .candidate-header {
    display: flex;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .candidate-avatar {
    width: 4rem;
    height: 4rem;
    border-radius: 50%;
    background: #3b82f6;
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 1.5rem;
    flex-shrink: 0;
  }

  .candidate-info {
    flex: 1;
  }

  .candidate-name {
    font-size: 1.125rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 0.25rem;
  }

  .candidate-meta {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .candidate-score {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background: #f9fafb;
    border-radius: 0.5rem;
    margin-bottom: 1rem;
  }

  .score-circle {
    width: 4rem;
    height: 4rem;
    border-radius: 50%;
    border: 4px solid;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .score-value {
    font-size: 1.5rem;
    font-weight: 700;
  }

  .score-label {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .candidate-skills {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-bottom: 1rem;
  }

  .skill-badge {
    padding: 0.25rem 0.75rem;
    background: #eff6ff;
    color: #1e40af;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .skill-badge-large {
    padding: 0.5rem 1rem;
    background: #eff6ff;
    color: #1e40af;
    border-radius: 9999px;
    font-size: 0.875rem;
    font-weight: 600;
  }

  .candidate-meta-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .candidate-date {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .candidate-actions {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .candidate-status-select {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e5e7eb;
  }

  .btn {
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-weight: 500;
    cursor: pointer;
    border: 1px solid transparent;
    transition: all 0.2s;
    background: none;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-sm {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
  }

  .btn-ghost {
    color: #6b7280;
    border-color: #d1d5db;
  }

  .btn-ghost:hover {
    background: #f9fafb;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-circle {
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
  }

  .badge-success {
    background: #d1fae5;
    color: #065f46;
  }

  .badge-info {
    background: #dbeafe;
    color: #1e40af;
  }

  .badge-warning {
    background: #fef3c7;
    color: #92400e;
  }

  .badge-error {
    background: #fee2e2;
    color: #991b1b;
  }

  .badge-ghost {
    background: #f3f4f6;
    color: #4b5563;
  }

  .badge-primary {
    background: #dbeafe;
    color: #1e40af;
  }

  .select,
  .input {
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 1rem;
  }

  .select-sm {
    padding: 0.375rem;
    font-size: 0.875rem;
  }

  .select:focus,
  .input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .w-full {
    width: 100%;
  }

  .modal {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal-box {
    background: white;
    border-radius: 0.75rem;
    padding: 2rem;
    max-height: 90vh;
    overflow-y: auto;
    width: 100%;
    max-width: 32rem;
  }

  .max-w-3xl {
    max-width: 48rem;
  }

  .max-w-4xl {
    max-width: 56rem;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1.5rem;
  }

  .modal-header h2 {
    font-size: 1.5rem;
    font-weight: 700;
    color: #111827;
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .form-control {
    display: flex;
    flex-direction: column;
  }

  .label {
    margin-bottom: 0.5rem;
  }

  .label-text {
    font-weight: 500;
    color: #374151;
  }

  .textarea {
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 1rem;
    font-family: inherit;
    resize: vertical;
  }

  .textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .array-input {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .array-item {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .modal-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
    margin-top: 1.5rem;
  }

  .candidate-detail {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .detail-section {
    border-bottom: 1px solid #e5e7eb;
    padding-bottom: 1.5rem;
  }

  .detail-section:last-child {
    border-bottom: none;
  }

  .detail-section h3 {
    font-size: 1.125rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 1rem;
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1rem;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .info-label {
    font-size: 0.875rem;
    color: #6b7280;
    font-weight: 500;
  }

  .info-value {
    font-size: 1rem;
    color: #111827;
  }

  .link {
    color: #3b82f6;
    text-decoration: underline;
  }

  .skills-list {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .ai-analysis {
    background: #f0f9ff;
    border: 1px solid #bae6fd;
    border-radius: 0.5rem;
    padding: 1rem;
  }

  .ai-analysis p {
    color: #075985;
    margin-bottom: 1rem;
  }

  .analysis-category {
    margin-top: 1rem;
  }

  .analysis-category h4 {
    font-size: 0.875rem;
    font-weight: 600;
    color: #0c4a6e;
    margin-bottom: 0.5rem;
  }

  .analysis-category ul {
    list-style: disc;
    padding-left: 1.5rem;
    color: #075985;
  }

  .analysis-category li {
    margin-bottom: 0.25rem;
  }

  .cover-letter {
    background: #f9fafb;
    padding: 1rem;
    border-radius: 0.5rem;
    white-space: pre-wrap;
    color: #374151;
  }

  .detail-actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .email-templates {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-bottom: 1.5rem;
    padding-bottom: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .analytics-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .analytics-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
  }

  .analytics-card h3 {
    font-size: 0.875rem;
    color: #6b7280;
    margin-bottom: 1rem;
  }

  .analytics-value {
    font-size: 2.5rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 0.5rem;
  }

  .analytics-trend {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .job-boards-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .job-board-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: #f9fafb;
    border-radius: 0.5rem;
  }

  .board-info {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .board-logo {
    font-size: 1.5rem;
  }

  .board-name {
    font-weight: 600;
    color: #111827;
  }

  .switch {
    position: relative;
    display: inline-block;
    width: 3rem;
    height: 1.75rem;
  }

  .switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider {
    position: absolute;
    cursor: pointer;
    inset: 0;
    background-color: #cbd5e1;
    transition: 0.4s;
    border-radius: 9999px;
  }

  .slider:before {
    position: absolute;
    content: "";
    height: 1.25rem;
    width: 1.25rem;
    left: 0.25rem;
    bottom: 0.25rem;
    background-color: white;
    transition: 0.4s;
    border-radius: 50%;
  }

  input:checked + .slider {
    background-color: #3b82f6;
  }

  input:checked + .slider:before {
    transform: translateX(1.25rem);
  }

  @media (max-width: 1024px) {
    .jobs-grid,
    .candidates-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .recruiting-container {
      padding: 1rem;
    }

    .stats-grid,
    .analytics-grid {
      grid-template-columns: 1fr;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .candidate-actions,
    .job-actions {
      flex-direction: column;
    }
  }
</style>