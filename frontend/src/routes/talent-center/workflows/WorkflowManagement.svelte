<script lang="ts">
  import { onMount } from 'svelte';
  import WorkflowTemplateManager from './WorkflowTemplateManager.svelte';
  import WorkflowAssign from './WorkflowAssign.svelte';
  import WorkflowMonitoring from './WorkflowMonitoring.svelte';
  
  type Tab = 'templates' | 'assign' | 'monitoring';
  
  let activeTab: Tab = 'templates';
  let stats = {
    templates: 0,
    activeWorkflows: 0,
    completedThisMonth: 0,
    averageCompletionTime: 0,
    pendingAssignments: 0
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
          activeWorkflows: data.active_workflows || 0,
          completedThisMonth: data.completed_this_month || 0,
          averageCompletionTime: Math.round(data.avg_completion_time || 0),
          pendingAssignments: data.pending_assignments || 0
        };
      }
    } catch (err) {
      console.error('Failed to load workflow stats:', err);
    } finally {
      loading = false;
    }
  }
  
  function handleStatsUpdate() {
    loadStats();
  }
</script>

<div class="workflow-management">
  <!-- Header -->
  <div class="page-header">
    <div class="header-content">
      <h1>Workflow Management</h1>
      <p>Create templates, assign workflows to employees, and monitor progress</p>
    </div>
  </div>

  <!-- Stats Cards -->
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-icon">📋</div>
      <div class="stat-content">
        <div class="stat-value">{stats.templates}</div>
        <div class="stat-label">Templates</div>
        <div class="stat-sublabel">Available workflow templates</div>
      </div>
    </div>
    
    <div class="stat-card active">
      <div class="stat-icon">🚀</div>
      <div class="stat-content">
        <div class="stat-value">{stats.activeWorkflows}</div>
        <div class="stat-label">Active Workflows</div>
        <div class="stat-sublabel">Currently in progress</div>
      </div>
    </div>
    
    <div class="stat-card">
      <div class="stat-icon">✅</div>
      <div class="stat-content">
        <div class="stat-value">{stats.completedThisMonth}</div>
        <div class="stat-label">Completed</div>
        <div class="stat-sublabel">This month</div>
      </div>
    </div>
    
    <div class="stat-card">
      <div class="stat-icon">⏱️</div>
      <div class="stat-content">
        <div class="stat-value">{stats.averageCompletionTime}d</div>
        <div class="stat-label">Avg Time</div>
        <div class="stat-sublabel">To completion</div>
      </div>
    </div>
    
    {#if stats.pendingAssignments > 0}
      <div class="stat-card pending">
        <div class="stat-icon">⚠️</div>
        <div class="stat-content">
          <div class="stat-value">{stats.pendingAssignments}</div>
          <div class="stat-label">Pending</div>
          <div class="stat-sublabel">Requires assignment</div>
        </div>
      </div>
    {/if}
  </div>

  <!-- Tab Navigation -->
  <div class="tabs-container">
    <div class="tabs">
      <button
        class="tab"
        class:active={activeTab === 'templates'}
        on:click={() => activeTab = 'templates'}
      >
        <span class="tab-icon">📋</span>
        <div class="tab-content">
          <span class="tab-title">Add Workflow Template</span>
          <span class="tab-description">Create reusable workflow templates</span>
        </div>
      </button>

      <button
        class="tab"
        class:active={activeTab === 'assign'}
        on:click={() => activeTab = 'assign'}
      >
        <span class="tab-icon">👤</span>
        <div class="tab-content">
          <span class="tab-title">Workflow Assign</span>
          <span class="tab-description">Assign workflows to employees</span>
        </div>
        {#if stats.pendingAssignments > 0}
          <span class="tab-badge">{stats.pendingAssignments}</span>
        {/if}
      </button>

      <button
        class="tab"
        class:active={activeTab === 'monitoring'}
        on:click={() => activeTab = 'monitoring'}
      >
        <span class="tab-icon">📊</span>
        <div class="tab-content">
          <span class="tab-title">Monitoring</span>
          <span class="tab-description">Track employee workflow progress</span>
        </div>
        {#if stats.activeWorkflows > 0}
          <span class="tab-badge">{stats.activeWorkflows}</span>
        {/if}
      </button>
    </div>
  </div>

  <!-- Tab Content -->
  <div class="tab-panel">
    {#if activeTab === 'templates'}
      <WorkflowTemplateManager on:updated={handleStatsUpdate} />
    {:else if activeTab === 'assign'}
      <WorkflowAssign on:assigned={handleStatsUpdate} />
    {:else if activeTab === 'monitoring'}
      <WorkflowMonitoring on:updated={handleStatsUpdate} />
    {/if}
  </div>
</div>

<style>
  .workflow-management {
    padding: 24px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .page-header {
    margin-bottom: 32px;
  }

  .header-content h1 {
    font-size: 32px;
    font-weight: 700;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .header-content p {
    font-size: 16px;
    color: #718096;
    margin: 0;
  }

  /* Stats Grid */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 20px;
    margin-bottom: 32px;
  }

  .stat-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 24px;
    background: white;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    transition: all 0.2s;
  }

  .stat-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }

  .stat-card.active {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
  }

  .stat-card.pending {
    background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
    color: white;
    border: none;
  }

  .stat-icon {
    font-size: 40px;
    min-width: 60px;
    text-align: center;
  }

  .stat-content {
    flex: 1;
  }

  .stat-value {
    font-size: 36px;
    font-weight: 700;
    line-height: 1;
    margin-bottom: 8px;
  }

  .stat-label {
    font-size: 14px;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .stat-card.active .stat-label,
  .stat-card.pending .stat-label {
    color: rgba(255, 255, 255, 0.95);
  }

  .stat-sublabel {
    font-size: 12px;
    opacity: 0.7;
  }

  /* Tabs */
  .tabs-container {
    margin-bottom: 24px;
  }

  .tabs {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 16px;
    background: white;
    padding: 8px;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
  }

  .tab {
    position: relative;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 20px;
    background: transparent;
    border: 2px solid transparent;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
  }

  .tab:hover {
    background: #f7fafc;
    border-color: #e2e8f0;
  }

  .tab.active {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border-color: transparent;
  }

  .tab-icon {
    font-size: 32px;
    min-width: 40px;
    text-align: center;
  }

  .tab-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .tab-title {
    font-size: 16px;
    font-weight: 600;
  }

  .tab-description {
    font-size: 12px;
    opacity: 0.8;
  }

  .tab:not(.active) .tab-description {
    color: #718096;
  }

  .tab-badge {
    position: absolute;
    top: 8px;
    right: 8px;
    background: #f56565;
    color: white;
    font-size: 11px;
    font-weight: 600;
    padding: 4px 8px;
    border-radius: 12px;
    min-width: 24px;
    text-align: center;
  }

  .tab.active .tab-badge {
    background: white;
    color: #667eea;
  }

  /* Tab Panel */
  .tab-panel {
    background: white;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    min-height: 500px;
  }

  @media (max-width: 768px) {
    .workflow-management {
      padding: 16px;
    }

    .stats-grid {
      grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
      gap: 12px;
    }

    .stat-card {
      padding: 16px;
    }

    .stat-icon {
      font-size: 32px;
    }

    .stat-value {
      font-size: 28px;
    }

    .tabs {
      grid-template-columns: 1fr;
    }

    .tab {
      padding: 16px;
    }
  }
</style>
