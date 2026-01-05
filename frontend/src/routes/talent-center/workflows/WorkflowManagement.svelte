<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import WorkflowManager from './WorkflowManager.svelte';
  import OnboardingDashboard from './OnboardingDashboard.svelte';
  
  const dispatch = createEventDispatcher();
  
  let activeTab: 'templates' | 'onboarding' = 'templates';
  let stats = {
    templates: 0,
    activeOnboardings: 0,
    completedThisMonth: 0,
    averageCompletionTime: 0
  };
  let loading = true;
  
  onMount(async () => {
    await loadStats();
  });
  
  async function loadStats() {
    try {
      loading = true;
      const token = localStorage.getItem('token');
      const response = await fetch('/api/workflows/stats', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        const data = await response.json();
        stats = {
          templates: data.templates_count || 0,
          activeOnboardings: data.active_onboardings || 0,
          completedThisMonth: data.completed_this_month || 0,
          averageCompletionTime: data.avg_completion_time || 0
        };
        dispatch('statsUpdated');
      }
    } catch (err) {
      console.error('Failed to load workflow stats:', err);
    } finally {
      loading = false;
    }
  }
</script>

<div class="workflow-management">
  <!-- Header Stats -->
  <div class="workflow-stats">
    <div class="stat-card">
      <div class="stat-icon">⚙️</div>
      <div class="stat-content">
        <div class="stat-value">{stats.templates}</div>
        <div class="stat-label">Workflow Templates</div>
        <div class="stat-sublabel">Active configurations</div>
      </div>
    </div>
    
    <div class="stat-card highlight">
      <div class="stat-icon">🚀</div>
      <div class="stat-content">
        <div class="stat-value">{stats.activeOnboardings}</div>
        <div class="stat-label">Active Onboardings</div>
        <div class="stat-sublabel">Currently in progress</div>
      </div>
    </div>
    
    <div class="stat-card">
      <div class="stat-icon">✓</div>
      <div class="stat-content">
        <div class="stat-value">{stats.completedThisMonth}</div>
        <div class="stat-label">Completed This Month</div>
        <div class="stat-sublabel">Successfully finished</div>
      </div>
    </div>
    
    <div class="stat-card">
      <div class="stat-icon">⏱️</div>
      <div class="stat-content">
        <div class="stat-value">{stats.averageCompletionTime} days</div>
        <div class="stat-label">Avg Completion Time</div>
        <div class="stat-sublabel">Last 3 months</div>
      </div>
    </div>
  </div>
  
  <!-- Info Banner -->
  <div class="info-banner">
    <div class="banner-icon">💡</div>
    <div class="banner-content">
      <h4>Workflow Management</h4>
      <p>Create reusable templates for onboarding processes or monitor active new hire onboardings with real-time progress tracking.</p>
    </div>
  </div>
  
  <!-- Tab Navigation -->
  <div class="workflow-tabs">
    <button 
      class="workflow-tab"
      class:active={activeTab === 'templates'}
      on:click={() => activeTab = 'templates'}
    >
      <span class="tab-icon">⚙️</span>
      <span class="tab-text">Workflow Templates</span>
      <span class="tab-description">Create and manage reusable workflows</span>
    </button>
    
    <button 
      class="workflow-tab"
      class:active={activeTab === 'onboarding'}
      on:click={() => activeTab = 'onboarding'}
    >
      <span class="tab-icon">🚀</span>
      <span class="tab-text">New Hire Onboarding</span>
      <span class="tab-description">Monitor active onboarding processes</span>
      {#if stats.activeOnboardings > 0}
        <span class="tab-badge">{stats.activeOnboardings}</span>
      {/if}
    </button>
  </div>
  
  <!-- Tab Content -->
  <div class="workflow-content">
    {#if activeTab === 'templates'}
      <div class="content-wrapper">
        <div class="content-header">
          <div>
            <h3>Workflow Templates</h3>
            <p>Define reusable workflow templates with steps, dependencies, and automation rules</p>
          </div>
        </div>
        <WorkflowManager />
      </div>
    {:else}
      <div class="content-wrapper">
        <div class="content-header">
          <div>
            <h3>New Hire Onboarding</h3>
            <p>Track onboarding progress, tasks completion, and AI-powered insights</p>
          </div>
        </div>
        <OnboardingDashboard />
      </div>
    {/if}
  </div>
</div>

<style>
  .workflow-management {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .workflow-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 16px;
  }
  
  .stat-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    transition: all 0.2s;
  }
  
  .stat-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
  
  .stat-card.highlight {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }
  
  .stat-icon {
    font-size: 36px;
    min-width: 50px;
    text-align: center;
  }
  
  .stat-content {
    flex: 1;
  }
  
  .stat-value {
    font-size: 32px;
    font-weight: 700;
    line-height: 1;
    margin-bottom: 4px;
  }
  
  .stat-card.highlight .stat-value {
    color: white;
  }
  
  .stat-label {
    font-size: 14px;
    font-weight: 600;
    color: #111827;
    margin-bottom: 2px;
  }
  
  .stat-card.highlight .stat-label {
    color: rgba(255, 255, 255, 0.95);
  }
  
  .stat-sublabel {
    font-size: 12px;
    color: #6b7280;
  }
  
  .stat-card.highlight .stat-sublabel {
    color: rgba(255, 255, 255, 0.8);
  }
  
  .info-banner {
    display: flex;
    align-items: start;
    gap: 16px;
    padding: 20px;
    background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
    border-left: 4px solid #3b82f6;
    border-radius: 12px;
  }
  
  .banner-icon {
    font-size: 28px;
  }
  
  .banner-content h4 {
    font-size: 16px;
    font-weight: 600;
    color: #1e40af;
    margin: 0 0 8px 0;
  }
  
  .banner-content p {
    font-size: 14px;
    color: #1e3a8a;
    margin: 0;
    line-height: 1.5;
  }
  
  .workflow-tabs {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 16px;
    background: white;
    padding: 8px;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }
  
  .workflow-tab {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: start;
    gap: 8px;
    padding: 20px;
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
  }
  
  .workflow-tab:hover {
    border-color: #3b82f6;
    background: #eff6ff;
  }
  
  .workflow-tab.active {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-color: #667eea;
    color: white;
  }
  
  .tab-icon {
    font-size: 28px;
  }
  
  .tab-text {
    font-size: 18px;
    font-weight: 600;
  }
  
  .tab-description {
    font-size: 13px;
    color: #6b7280;
  }
  
  .workflow-tab.active .tab-description {
    color: rgba(255, 255, 255, 0.9);
  }
  
  .tab-badge {
    position: absolute;
    top: 12px;
    right: 12px;
    padding: 4px 10px;
    background: #ef4444;
    color: white;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
  }
  
  .workflow-content {
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    overflow: hidden;
  }
  
  .content-wrapper {
    display: flex;
    flex-direction: column;
  }
  
  .content-header {
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
    background: linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%);
  }
  
  .content-header h3 {
    font-size: 20px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .content-header p {
    font-size: 14px;
    color: #6b7280;
    margin: 0;
  }
  
  @media (max-width: 1024px) {
    .workflow-stats {
      grid-template-columns: repeat(2, 1fr);
    }
    
    .workflow-tabs {
      grid-template-columns: 1fr;
    }
  }
  
  @media (max-width: 640px) {
    .workflow-stats {
      grid-template-columns: 1fr;
    }
    
    .stat-card {
      padding: 16px;
    }
    
    .stat-icon {
      font-size: 28px;
    }
    
    .stat-value {
      font-size: 24px;
    }
    
    .info-banner {
      flex-direction: column;
    }
  }
</style>
