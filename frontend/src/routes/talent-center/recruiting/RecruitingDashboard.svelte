<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import RecruitingProviders from './RecruitingProviders.svelte';
  import JobPostingManager from './JobPostingManager.svelte';
  import ApplicantLeaderboard from './ApplicantLeaderboard.svelte';
  
  const dispatch = createEventDispatcher();
  
  let activeView: 'overview' | 'providers' | 'jobs' | 'applicants' = 'overview';
  let stats = {
    activeJobs: 0,
    totalApplicants: 0,
    pendingReviews: 0,
    scheduledInterviews: 0,
    avgTimeToHire: 0,
    topSourceByApplicants: '',
    connectedProviders: 0
  };
  let recentJobs: any[] = [];
  let topApplicants: any[] = [];
  let loading = true;
  
  onMount(async () => {
    await loadDashboard();
  });
  
  async function loadDashboard() {
    try {
      loading = true;
      const token = localStorage.getItem('token');
      
      // Load recruiting stats
      const statsRes = await fetch('/api/recruiting/dashboard', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (statsRes.ok) {
        const data = await statsRes.json();
        stats = {
          activeJobs: data.active_jobs || 0,
          totalApplicants: data.total_applicants || 0,
          pendingReviews: data.pending_reviews || 0,
          scheduledInterviews: data.scheduled_interviews || 0,
          avgTimeToHire: data.avg_time_to_hire || 0,
          topSourceByApplicants: data.top_source || 'N/A',
          connectedProviders: data.connected_providers || 0
        };
        recentJobs = data.recent_jobs || [];
        topApplicants = data.top_applicants || [];
      }
      
      dispatch('statsUpdated');
    } catch (err) {
      console.error('Failed to load recruiting dashboard:', err);
    } finally {
      loading = false;
    }
  }
  
  function handleProviderUpdated() {
    loadDashboard();
  }
</script>

<div class="recruiting-dashboard">
  {#if activeView === 'overview'}
    <!-- Overview Section -->
    <div class="overview-container">
      <!-- Stats Grid -->
      <div class="stats-grid">
        <div class="stat-card primary">
          <div class="stat-header">
            <span class="stat-icon">📋</span>
            <h3>Active Jobs</h3>
          </div>
          <div class="stat-value">{stats.activeJobs}</div>
          <button class="stat-action" on:click={() => activeView = 'jobs'}>
            Manage Jobs →
          </button>
        </div>
        
        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-icon">👥</span>
            <h3>Total Applicants</h3>
          </div>
          <div class="stat-value">{stats.totalApplicants}</div>
          <button class="stat-action" on:click={() => activeView = 'applicants'}>
            View Leaderboard →
          </button>
        </div>
        
        <div class="stat-card warning">
          <div class="stat-header">
            <span class="stat-icon">⏳</span>
            <h3>Pending Reviews</h3>
          </div>
          <div class="stat-value">{stats.pendingReviews}</div>
          <div class="stat-subtitle">Need attention</div>
        </div>
        
        <div class="stat-card success">
          <div class="stat-header">
            <span class="stat-icon">📅</span>
            <h3>Scheduled Interviews</h3>
          </div>
          <div class="stat-value">{stats.scheduledInterviews}</div>
          <div class="stat-subtitle">Upcoming</div>
        </div>
        
        <div class="stat-card info">
          <div class="stat-header">
            <span class="stat-icon">⚡</span>
            <h3>Avg Time to Hire</h3>
          </div>
          <div class="stat-value">{stats.avgTimeToHire} days</div>
          <div class="stat-subtitle">Last 90 days</div>
        </div>
        
        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-icon">🔗</span>
            <h3>Connected Providers</h3>
          </div>
          <div class="stat-value">{stats.connectedProviders}</div>
          <button class="stat-action" on:click={() => activeView = 'providers'}>
            Manage Providers →
          </button>
        </div>
      </div>
      
      <!-- Quick Actions -->
      <div class="section-card">
        <h2 class="section-title">Quick Actions</h2>
        <div class="action-buttons">
          <button class="action-btn primary" on:click={() => activeView = 'jobs'}>
            <span class="btn-icon">➕</span>
            Post New Job
          </button>
          <button class="action-btn" on:click={() => activeView = 'providers'}>
            <span class="btn-icon">🔗</span>
            Configure Providers
          </button>
          <button class="action-btn" on:click={() => activeView = 'applicants'}>
            <span class="btn-icon">📊</span>
            Review Applicants
          </button>
        </div>
      </div>
      
      <!-- Recent Jobs -->
      <div class="section-card">
        <div class="section-header">
          <h2 class="section-title">Recent Job Postings</h2>
          <button class="view-all-btn" on:click={() => activeView = 'jobs'}>
            View All →
          </button>
        </div>
        
        {#if recentJobs.length > 0}
          <div class="jobs-list">
            {#each recentJobs as job}
              <div class="job-item">
                <div class="job-header">
                  <h3 class="job-title">{job.title}</h3>
                  <span class="job-status status-{job.status}">{job.status}</span>
                </div>
                <div class="job-meta">
                  <span class="meta-item">
                    <span class="meta-icon">📍</span>
                    {job.location}
                  </span>
                  <span class="meta-item">
                    <span class="meta-icon">💰</span>
                    {job.salary_range}
                  </span>
                  <span class="meta-item">
                    <span class="meta-icon">👥</span>
                    {job.applicant_count} applicants
                  </span>
                </div>
                {#if job.providers && job.providers.length > 0}
                  <div class="job-providers">
                    {#each job.providers as provider}
                      <span class="provider-badge">{provider}</span>
                    {/each}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {:else}
          <div class="empty-state">
            <span class="empty-icon">📋</span>
            <p>No job postings yet</p>
            <button class="empty-action" on:click={() => activeView = 'jobs'}>
              Create Your First Job
            </button>
          </div>
        {/if}
      </div>
      
      <!-- Top Applicants Preview -->
      <div class="section-card">
        <div class="section-header">
          <h2 class="section-title">🏆 Top Ranked Applicants</h2>
          <button class="view-all-btn" on:click={() => activeView = 'applicants'}>
            View Leaderboard →
          </button>
        </div>
        
        {#if topApplicants.length > 0}
          <div class="applicants-preview">
            {#each topApplicants.slice(0, 5) as applicant, index}
              <div class="applicant-item">
                <div class="applicant-rank">#{index + 1}</div>
                <div class="applicant-info">
                  <div class="applicant-name">{applicant.name}</div>
                  <div class="applicant-role">{applicant.position}</div>
                </div>
                <div class="applicant-score">
                  <div class="score-value">{applicant.ai_score}/100</div>
                  <div class="score-bar">
                    <div class="score-fill" style="width: {applicant.ai_score}%"></div>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {:else}
          <div class="empty-state small">
            <p>No applicants to rank yet</p>
          </div>
        {/if}
      </div>
    </div>
  {:else if activeView === 'providers'}
    <!-- Providers Management -->
    <div class="view-header">
      <button class="back-btn" on:click={() => activeView = 'overview'}>
        ← Back to Overview
      </button>
      <h2>Recruiting Providers</h2>
    </div>
    <RecruitingProviders on:providerUpdated={handleProviderUpdated} />
  {:else if activeView === 'jobs'}
    <!-- Job Posting Management -->
    <div class="view-header">
      <button class="back-btn" on:click={() => activeView = 'overview'}>
        ← Back to Overview
      </button>
      <h2>Job Postings</h2>
    </div>
    <JobPostingManager on:jobUpdated={loadDashboard} />
  {:else if activeView === 'applicants'}
    <!-- Applicant Leaderboard -->
    <div class="view-header">
      <button class="back-btn" on:click={() => activeView = 'overview'}>
        ← Back to Overview
      </button>
      <h2>Applicant Leaderboard</h2>
    </div>
    <ApplicantLeaderboard />
  {/if}
</div>

<style>
  .recruiting-dashboard {
    min-height: 100%;
  }
  
  .overview-container {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 20px;
  }
  
  .stat-card {
    background: white;
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    transition: transform 0.2s, box-shadow 0.2s;
  }
  
  .stat-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
  
  .stat-card.primary {
    border-left: 4px solid #3b82f6;
  }
  
  .stat-card.success {
    border-left: 4px solid #10b981;
  }
  
  .stat-card.warning {
    border-left: 4px solid #f59e0b;
  }
  
  .stat-card.info {
    border-left: 4px solid #6366f1;
  }
  
  .stat-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }
  
  .stat-icon {
    font-size: 24px;
  }
  
  .stat-header h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    color: #6b7280;
  }
  
  .stat-value {
    font-size: 36px;
    font-weight: 700;
    color: #111827;
    margin-bottom: 8px;
  }
  
  .stat-subtitle {
    font-size: 13px;
    color: #6b7280;
  }
  
  .stat-action {
    margin-top: 12px;
    padding: 8px 0;
    background: none;
    border: none;
    color: #3b82f6;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: color 0.2s;
    width: 100%;
    text-align: left;
  }
  
  .stat-action:hover {
    color: #2563eb;
  }
  
  .section-card {
    background: white;
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }
  
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }
  
  .section-title {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }
  
  .view-all-btn {
    background: none;
    border: none;
    color: #3b82f6;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: color 0.2s;
  }
  
  .view-all-btn:hover {
    color: #2563eb;
  }
  
  .action-buttons {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;
  }
  
  .action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 14px 20px;
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    color: #374151;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .action-btn:hover {
    border-color: #3b82f6;
    color: #3b82f6;
    transform: translateY(-1px);
  }
  
  .action-btn.primary {
    background: #3b82f6;
    border-color: #3b82f6;
    color: white;
  }
  
  .action-btn.primary:hover {
    background: #2563eb;
    border-color: #2563eb;
    color: white;
  }
  
  .btn-icon {
    font-size: 18px;
  }
  
  .jobs-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  
  .job-item {
    padding: 16px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    transition: all 0.2s;
  }
  
  .job-item:hover {
    border-color: #3b82f6;
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
  }
  
  .job-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 12px;
  }
  
  .job-title {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #111827;
  }
  
  .job-status {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;
  }
  
  .job-status.status-active {
    background: #d1fae5;
    color: #065f46;
  }
  
  .job-status.status-draft {
    background: #fef3c7;
    color: #92400e;
  }
  
  .job-status.status-closed {
    background: #e5e7eb;
    color: #374151;
  }
  
  .job-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    margin-bottom: 12px;
  }
  
  .meta-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #6b7280;
  }
  
  .meta-icon {
    font-size: 14px;
  }
  
  .job-providers {
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
  
  .applicants-preview {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .applicant-item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
  }
  
  .applicant-rank {
    font-size: 20px;
    font-weight: 700;
    color: #3b82f6;
    min-width: 40px;
    text-align: center;
  }
  
  .applicant-info {
    flex: 1;
  }
  
  .applicant-name {
    font-size: 14px;
    font-weight: 600;
    color: #111827;
  }
  
  .applicant-role {
    font-size: 12px;
    color: #6b7280;
  }
  
  .applicant-score {
    min-width: 120px;
  }
  
  .score-value {
    font-size: 16px;
    font-weight: 600;
    color: #3b82f6;
    margin-bottom: 4px;
  }
  
  .score-bar {
    width: 100%;
    height: 6px;
    background: #e5e7eb;
    border-radius: 3px;
    overflow: hidden;
  }
  
  .score-fill {
    height: 100%;
    background: linear-gradient(90deg, #3b82f6 0%, #8b5cf6 100%);
    transition: width 0.3s;
  }
  
  .empty-state {
    text-align: center;
    padding: 48px 24px;
  }
  
  .empty-state.small {
    padding: 24px;
  }
  
  .empty-icon {
    font-size: 48px;
    display: block;
    margin-bottom: 16px;
  }
  
  .empty-state p {
    color: #6b7280;
    margin: 0 0 16px 0;
  }
  
  .empty-action {
    padding: 10px 20px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
  }
  
  .empty-action:hover {
    background: #2563eb;
  }
  
  .view-header {
    margin-bottom: 24px;
  }
  
  .back-btn {
    background: none;
    border: none;
    color: #3b82f6;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    padding: 8px 0;
    margin-bottom: 12px;
    transition: color 0.2s;
  }
  
  .back-btn:hover {
    color: #2563eb;
  }
  
  .view-header h2 {
    font-size: 24px;
    font-weight: 700;
    color: #111827;
    margin: 0;
  }
  
  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }
    
    .action-buttons {
      grid-template-columns: 1fr;
    }
    
    .job-header {
      flex-direction: column;
      gap: 8px;
    }
    
    .applicant-item {
      flex-direction: column;
      align-items: flex-start;
    }
    
    .applicant-score {
      width: 100%;
    }
  }
</style>
