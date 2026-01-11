<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import RecruitingProviders from './RecruitingProviders.svelte';
  import JobPostingManager from './JobPostingManager.svelte';
  import ApplicantLeaderboard from './ApplicantLeaderboard.svelte';
  import CandidateSummary from './CandidateSummary.svelte';
  
  const dispatch = createEventDispatcher();
  
  let activeView: 'overview' | 'providers' | 'jobs' | 'leaderboard' | 'summary' = 'overview';
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
          <button class="stat-action" on:click={() => activeView = 'leaderboard'}>
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
          <div class="stat-subtitle">This week</div>
        </div>
      </div>
      
      <!-- Quick Actions -->
      <div class="quick-actions-section">
        <h2>Quick Actions</h2>
        <div class="quick-actions">
          <button class="action-card" on:click={() => activeView = 'jobs'}>
            <span class="action-icon">➕</span>
            <div class="action-content">
              <h3>Post New Job</h3>
              <p>Create and publish job posting</p>
            </div>
          </button>
          
          <button class="action-card" on:click={() => activeView = 'leaderboard'}>
            <span class="action-icon">🏆</span>
            <div class="action-content">
              <h3>Applicant Leaderboard</h3>
              <p>View ranked candidates by job</p>
            </div>
          </button>
          
          <button class="action-card" on:click={() => activeView = 'summary'}>
            <span class="action-icon">📊</span>
            <div class="action-content">
              <h3>Candidate Summary</h3>
              <p>Generate AI summary of top candidates</p>
            </div>
          </button>
          
          <button class="action-card" on:click={() => activeView = 'providers'}>
            <span class="action-icon">🔗</span>
            <div class="action-content">
              <h3>Connect Providers</h3>
              <p>Integrate with job boards</p>
            </div>
          </button>
        </div>
      </div>
      
      <!-- Recent Activity -->
      <div class="activity-section">
        <div class="section-header">
          <h2>Recent Job Postings</h2>
          <button class="btn-secondary" on:click={() => activeView = 'jobs'}>
            View All
          </button>
        </div>
        <div class="recent-jobs">
          {#if recentJobs.length > 0}
            {#each recentJobs.slice(0, 5) as job}
              <div class="job-item">
                <div class="job-info">
                  <h3>{job.title}</h3>
                  <p>{job.department} • {job.location}</p>
                </div>
                <div class="job-stats">
                  <span class="applicant-count">{job.applicant_count || 0} applicants</span>
                  <span class="status-badge" class:active={job.status === 'active'}>
                    {job.status}
                  </span>
                </div>
              </div>
            {/each}
          {:else}
            <div class="empty-state">
              <p>No recent jobs. Post your first job to get started!</p>
            </div>
          {/if}
        </div>
      </div>
      
      <!-- Top Applicants Preview -->
      <div class="activity-section">
        <div class="section-header">
          <h2>Top Applicants This Week</h2>
          <button class="btn-secondary" on:click={() => activeView = 'leaderboard'}>
            View Leaderboard
          </button>
        </div>
        <div class="top-applicants">
          {#if topApplicants.length > 0}
            {#each topApplicants.slice(0, 5) as applicant, index}
              <div class="applicant-item">
                <div class="rank"># {index + 1}</div>
                <div class="applicant-info">
                  <h4>{applicant.name}</h4>
                  <p>{applicant.position}</p>
                </div>
                <div class="applicant-score">
                  <span class="score">{applicant.score || 0}</span>
                  <span class="score-label">Score</span>
                </div>
              </div>
            {/each}
          {:else}
            <div class="empty-state">
              <p>No applicants yet. They'll appear here once you start receiving applications!</p>
            </div>
          {/if}
        </div>
      </div>
    </div>
    
  {:else if activeView === 'providers'}
    <div class="view-container">
      <div class="view-header">
        <button class="btn-back" on:click={() => activeView = 'overview'}>
          ← Back to Overview
        </button>
        <h2>Recruiting Providers</h2>
      </div>
      <RecruitingProviders on:providerUpdated={handleProviderUpdated} />
    </div>
    
  {:else if activeView === 'jobs'}
    <div class="view-container">
      <div class="view-header">
        <button class="btn-back" on:click={() => activeView = 'overview'}>
          ← Back to Overview
        </button>
        <h2>Job Posting Manager</h2>
      </div>
      <JobPostingManager on:jobUpdated={loadDashboard} />
    </div>
    
  {:else if activeView === 'leaderboard'}
    <div class="view-container">
      <div class="view-header">
        <button class="btn-back" on:click={() => activeView = 'overview'}>
          ← Back to Overview
        </button>
        <h2>📊 Applicant Leaderboard</h2>
      </div>
      <ApplicantLeaderboard />
    </div>
    
  {:else if activeView === 'summary'}
    <div class="view-container">
      <div class="view-header">
        <button class="btn-back" on:click={() => activeView = 'overview'}>
          ← Back to Overview
        </button>
        <h2>🎯 Candidate Summary</h2>
      </div>
      <CandidateSummary />
    </div>
  {/if}
</div>

<style>
  .recruiting-dashboard {
    min-height: 600px;
  }
  
  .overview-container {
    display: flex;
    flex-direction: column;
    gap: 32px;
  }
  
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 20px;
  }
  
  .stat-card {
    background: white;
    border-radius: 12px;
    padding: 24px;
    border: 1px solid #e5e7eb;
    transition: all 0.2s;
  }
  
  .stat-card:hover {
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07);
    transform: translateY(-2px);
  }
  
  .stat-card.primary {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
  }
  
  .stat-card.warning {
    background: #fef3c7;
    border-color: #fbbf24;
  }
  
  .stat-card.success {
    background: #d1fae5;
    border-color: #10b981;
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
    font-size: 14px;
    font-weight: 600;
    opacity: 0.9;
    margin: 0;
  }
  
  .stat-value {
    font-size: 36px;
    font-weight: 700;
    margin-bottom: 8px;
  }
  
  .stat-subtitle {
    font-size: 13px;
    opacity: 0.7;
  }
  
  .stat-action {
    background: rgba(255, 255, 255, 0.2);
    color: inherit;
    border: 1px solid rgba(255, 255, 255, 0.3);
    padding: 8px 16px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    margin-top: 12px;
  }
  
  .stat-action:hover {
    background: rgba(255, 255, 255, 0.3);
  }
  
  .stat-card.primary .stat-action {
    background: rgba(255, 255, 255, 0.2);
    border-color: rgba(255, 255, 255, 0.3);
  }
  
  .stat-card.warning .stat-action,
  .stat-card.success .stat-action {
    background: white;
    color: #111827;
    border-color: #d1d5db;
  }
  
  .quick-actions-section h2 {
    font-size: 20px;
    font-weight: 700;
    color: #111827;
    margin-bottom: 16px;
  }
  
  .quick-actions {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 16px;
  }
  
  .action-card {
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    display: flex;
    align-items: center;
    gap: 16px;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
  }
  
  .action-card:hover {
    border-color: #667eea;
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
    transform: translateY(-2px);
  }
  
  .action-icon {
    font-size: 32px;
    flex-shrink: 0;
  }
  
  .action-content h3 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .action-content p {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .activity-section {
    background: white;
    border-radius: 12px;
    padding: 24px;
    border: 1px solid #e5e7eb;
  }
  
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }
  
  .section-header h2 {
    font-size: 18px;
    font-weight: 700;
    color: #111827;
    margin: 0;
  }
  
  .btn-secondary {
    background: #f3f4f6;
    color: #374151;
    border: 1px solid #d1d5db;
    padding: 8px 16px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-secondary:hover {
    background: #e5e7eb;
  }
  
  .recent-jobs {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .job-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    background: #f9fafb;
    border-radius: 8px;
    border: 1px solid #e5e7eb;
  }
  
  .job-info h3 {
    font-size: 15px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .job-info p {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .job-stats {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  
  .applicant-count {
    font-size: 13px;
    color: #6b7280;
  }
  
  .status-badge {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
    background: #e5e7eb;
    color: #6b7280;
  }
  
  .status-badge.active {
    background: #d1fae5;
    color: #065f46;
  }
  
  .top-applicants {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .applicant-item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px;
    background: #f9fafb;
    border-radius: 8px;
    border: 1px solid #e5e7eb;
  }
  
  .rank {
    font-size: 18px;
    font-weight: 700;
    color: #667eea;
    min-width: 40px;
  }
  
  .applicant-info {
    flex: 1;
  }
  
  .applicant-info h4 {
    font-size: 15px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .applicant-info p {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .applicant-score {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 8px 16px;
    background: white;
    border-radius: 8px;
  }
  
  .score {
    font-size: 20px;
    font-weight: 700;
    color: #667eea;
  }
  
  .score-label {
    font-size: 11px;
    color: #6b7280;
    text-transform: uppercase;
  }
  
  .empty-state {
    text-align: center;
    padding: 48px 24px;
    color: #6b7280;
  }
  
  .empty-state p {
    margin: 0;
  }
  
  .view-container {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .view-header {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  
  .view-header h2 {
    font-size: 24px;
    font-weight: 700;
    color: #111827;
    margin: 0;
  }
  
  .btn-back {
    background: #f3f4f6;
    color: #374151;
    border: 1px solid #d1d5db;
    padding: 8px 16px;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-back:hover {
    background: #e5e7eb;
  }
  
  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }
    
    .quick-actions {
      grid-template-columns: 1fr;
    }
    
    .job-item {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }
    
    .job-stats {
      width: 100%;
      justify-content: space-between;
    }
  }
</style>
