<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../../stores/auth';
  import { getApiBaseUrl } from '../../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();
  
  let activeTab: 'overview' | 'documents' | 'training' | 'audits' = 'overview';
  
  // Compliance Overview State
  interface ComplianceMetric {
    category: string;
    status: 'compliant' | 'warning' | 'critical';
    score: number;
    issues: number;
    last_audit: string;
  }
  
  let complianceMetrics: ComplianceMetric[] = [];
  let overallScore = 0;
  let criticalIssues = 0;
  let upcomingDeadlines = 0;
  
  // Documents State
  interface ComplianceDocument {
    id: string;
    title: string;
    category: string;
    type: 'policy' | 'form' | 'certificate' | 'agreement';
    status: 'current' | 'expiring' | 'expired' | 'missing';
    expiry_date?: string;
    last_reviewed: string;
    department: string;
    owner: string;
    file_url?: string;
  }
  
  let documents: ComplianceDocument[] = [];
  let documentFilter = 'all';
  
  // Training State
  interface TrainingModule {
    id: string;
    title: string;
    category: string;
    required: boolean;
    completion_rate: number;
    total_employees: number;
    completed_employees: number;
    pending_employees: number;
    due_date?: string;
    status: 'on-track' | 'at-risk' | 'overdue';
  }
  
  let trainingModules: TrainingModule[] = [];
  
  // Audit State
  interface AuditItem {
    id: string;
    title: string;
    category: string;
    status: 'open' | 'in-progress' | 'resolved' | 'closed';
    severity: 'low' | 'medium' | 'high' | 'critical';
    assigned_to: string;
    due_date: string;
    created_date: string;
    description: string;
    findings?: string;
    resolution?: string;
  }
  
  let auditItems: AuditItem[] = [];
  let auditFilter = 'open';
  
  // Modal State
  let selectedDocument: ComplianceDocument | null = null;
  let selectedTraining: TrainingModule | null = null;
  let selectedAudit: AuditItem | null = null;
  let showDocumentModal = false;
  let showTrainingModal = false;
  let showAuditModal = false;
  
  let loading = true;
  
  onMount(async () => {
    await loadComplianceData();
  });
  
  async function loadComplianceData() {
    try {
      loading = true;
      const token = $authStore.token || localStorage.getItem('token');
      
      // Load all compliance data
      const [metricsRes, docsRes, trainingRes, auditsRes] = await Promise.all([
        fetch(`${API_BASE_URL}/compliance/metrics`, {
          headers: { 'Authorization': `Bearer ${token}` }
        }),
        fetch(`${API_BASE_URL}/compliance/documents`, {
          headers: { 'Authorization': `Bearer ${token}` }
        }),
        fetch(`${API_BASE_URL}/compliance/training`, {
          headers: { 'Authorization': `Bearer ${token}` }
        }),
        fetch(`${API_BASE_URL}/compliance/audits`, {
          headers: { 'Authorization': `Bearer ${token}` }
        })
      ]);
      
      if (metricsRes.ok) {
        const data = await metricsRes.json();
        complianceMetrics = data.metrics || [];
        overallScore = data.overall_score || 0;
        criticalIssues = data.critical_issues || 0;
        upcomingDeadlines = data.upcoming_deadlines || 0;
      }
      
      if (docsRes.ok) documents = await docsRes.json();
      if (trainingRes.ok) trainingModules = await trainingRes.json();
      if (auditsRes.ok) auditItems = await auditsRes.json();
      
    } catch (err) {
      console.error('Failed to load compliance data:', err);
    } finally {
      loading = false;
    }
  }
  
  function getStatusColor(status: string): string {
    const colors = {
      'compliant': 'bg-green-100 text-green-800',
      'current': 'bg-green-100 text-green-800',
      'on-track': 'bg-green-100 text-green-800',
      'resolved': 'bg-green-100 text-green-800',
      'closed': 'bg-gray-100 text-gray-800',
      'warning': 'bg-yellow-100 text-yellow-800',
      'expiring': 'bg-yellow-100 text-yellow-800',
      'at-risk': 'bg-yellow-100 text-yellow-800',
      'in-progress': 'bg-blue-100 text-blue-800',
      'critical': 'bg-red-100 text-red-800',
      'expired': 'bg-red-100 text-red-800',
      'missing': 'bg-red-100 text-red-800',
      'overdue': 'bg-red-100 text-red-800',
      'open': 'bg-purple-100 text-purple-800'
    };
    return colors[status] || 'bg-gray-100 text-gray-800';
  }
  
  function getSeverityColor(severity: string): string {
    const colors = {
      'low': '#10b981',
      'medium': '#f59e0b',
      'high': '#f97316',
      'critical': '#ef4444'
    };
    return colors[severity] || '#6b7280';
  }
  
  function viewDocument(doc: ComplianceDocument) {
    selectedDocument = doc;
    showDocumentModal = true;
  }
  
  function viewTraining(training: TrainingModule) {
    selectedTraining = training;
    showTrainingModal = true;
  }
  
  function viewAudit(audit: AuditItem) {
    selectedAudit = audit;
    showAuditModal = true;
  }
  
  async function updateAuditStatus(auditId: string, status: string) {
    try {
      const token = $authStore.token || localStorage.getItem('token');
      await fetch(`${API_BASE_URL}/compliance/audits/${auditId}/status`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status })
      });
      await loadComplianceData();
      showAuditModal = false;
    } catch (err) {
      console.error('Failed to update audit:', err);
    }
  }
  
  $: filteredDocuments = documents.filter(d => 
    documentFilter === 'all' || d.status === documentFilter
  );
  
  $: filteredAudits = auditItems.filter(a =>
    auditFilter === 'all' || a.status === auditFilter
  );
</script>

<div class="compliance-dashboard">
  <!-- Header -->
  <div class="header">
    <div>
      <h1>⚖️ Compliance Dashboard</h1>
      <p class="subtitle">Monitor regulatory compliance, training, and audit requirements</p>
    </div>
  </div>
  
  <!-- Tab Navigation -->
  <div class="tabs">
    <button 
      class="tab"
      class:active={activeTab === 'overview'}
      on:click={() => activeTab = 'overview'}
    >
      <span class="tab-icon">📊</span>
      <span class="tab-text">Overview</span>
    </button>
    
    <button 
      class="tab"
      class:active={activeTab === 'documents'}
      on:click={() => activeTab = 'documents'}
    >
      <span class="tab-icon">📄</span>
      <span class="tab-text">Documents</span>
      {#if documents.filter(d => d.status === 'expired' || d.status === 'expiring').length > 0}
        <span class="tab-badge">{documents.filter(d => d.status === 'expired' || d.status === 'expiring').length}</span>
      {/if}
    </button>
    
    <button 
      class="tab"
      class:active={activeTab === 'training'}
      on:click={() => activeTab = 'training'}
    >
      <span class="tab-icon">🎓</span>
      <span class="tab-text">Training</span>
      {#if trainingModules.filter(t => t.status === 'overdue' || t.status === 'at-risk').length > 0}
        <span class="tab-badge">{trainingModules.filter(t => t.status === 'overdue' || t.status === 'at-risk').length}</span>
      {/if}
    </button>
    
    <button 
      class="tab"
      class:active={activeTab === 'audits'}
      on:click={() => activeTab = 'audits'}
    >
      <span class="tab-icon">🔍</span>
      <span class="tab-text">Audits</span>
      {#if auditItems.filter(a => a.status === 'open' || a.status === 'in-progress').length > 0}
        <span class="tab-badge">{auditItems.filter(a => a.status === 'open' || a.status === 'in-progress').length}</span>
      {/if}
    </button>
  </div>
  
  <!-- Tab Content -->
  {#if loading}
    <div class="loading">Loading compliance data...</div>
  {:else}
    
    {#if activeTab === 'overview'}
      <!-- Overview Tab -->
      <div class="tab-content">
        <!-- Key Metrics -->
        <div class="stats-grid">
          <div class="stat-card large">
            <div class="score-circle" class:critical={overallScore < 70} class:warning={overallScore >= 70 && overallScore < 85} class:good={overallScore >= 85}>
              <div class="score-value">{overallScore}%</div>
              <div class="score-label">Overall Compliance</div>
            </div>
          </div>
          
          <div class="stat-card" class:highlight={criticalIssues > 0}>
            <div class="stat-icon">⚠️</div>
            <div class="stat-content">
              <div class="stat-value">{criticalIssues}</div>
              <div class="stat-label">Critical Issues</div>
              <div class="stat-sublabel">Requires immediate attention</div>
            </div>
          </div>
          
          <div class="stat-card" class:highlight={upcomingDeadlines > 0}>
            <div class="stat-icon">📅</div>
            <div class="stat-content">
              <div class="stat-value">{upcomingDeadlines}</div>
              <div class="stat-label">Upcoming Deadlines</div>
              <div class="stat-sublabel">Next 30 days</div>
            </div>
          </div>
        </div>
        
        <!-- Compliance Categories -->
        <div class="section-card">
          <div class="section-header">
            <h2>Compliance Categories</h2>
          </div>
          
          <div class="categories-grid">
            {#each complianceMetrics as metric}
              <div class="category-card">
                <div class="category-header">
                  <h3>{metric.category}</h3>
                  <span class="status-badge {getStatusColor(metric.status)}">
                    {metric.status}
                  </span>
                </div>
                
                <div class="category-score">
                  <div class="score-bar">
                    <div class="score-fill" style="width: {metric.score}%; background: {getSeverityColor(metric.status === 'critical' ? 'critical' : metric.status === 'warning' ? 'medium' : 'low')}"></div>
                  </div>
                  <span class="score-text">{metric.score}%</span>
                </div>
                
                <div class="category-details">
                  <span class="detail-item">
                    <span class="detail-icon">⚠️</span>
                    {metric.issues} issues
                  </span>
                  <span class="detail-item">
                    <span class="detail-icon">🔍</span>
                    Last audit: {new Date(metric.last_audit).toLocaleDateString()}
                  </span>
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
      
    {:else if activeTab === 'documents'}
      <!-- Documents Tab -->
      <div class="tab-content">
        <!-- Document Filters -->
        <div class="filter-bar">
          <button class="filter-btn" class:active={documentFilter === 'all'} on:click={() => documentFilter = 'all'}>
            All ({documents.length})
          </button>
          <button class="filter-btn" class:active={documentFilter === 'current'} on:click={() => documentFilter = 'current'}>
            Current ({documents.filter(d => d.status === 'current').length})
          </button>
          <button class="filter-btn" class:active={documentFilter === 'expiring'} on:click={() => documentFilter = 'expiring'}>
            Expiring ({documents.filter(d => d.status === 'expiring').length})
          </button>
          <button class="filter-btn" class:active={documentFilter === 'expired'} on:click={() => documentFilter = 'expired'}>
            Expired ({documents.filter(d => d.status === 'expired').length})
          </button>
          <button class="filter-btn" class:active={documentFilter === 'missing'} on:click={() => documentFilter = 'missing'}>
            Missing ({documents.filter(d => d.status === 'missing').length})
          </button>
        </div>
        
        <!-- Documents List -->
        <div class="section-card">
          {#if filteredDocuments.length === 0}
            <div class="empty-state">
              <span class="empty-icon">📄</span>
              <p>No {documentFilter === 'all' ? '' : documentFilter} documents found</p>
            </div>
          {:else}
            <div class="documents-grid">
              {#each filteredDocuments as doc}
                <div class="document-card" on:click={() => viewDocument(doc)}>
                  <div class="document-header">
                    <div class="document-icon">
                      {doc.type === 'policy' ? '📋' : doc.type === 'certificate' ? '🏆' : doc.type === 'form' ? '📝' : '📄'}
                    </div>
                    <span class="status-badge {getStatusColor(doc.status)}">
                      {doc.status}
                    </span>
                  </div>
                  
                  <h3 class="document-title">{doc.title}</h3>
                  
                  <div class="document-meta">
                    <span class="meta-item">
                      <span class="meta-icon">🏷️</span>
                      {doc.category}
                    </span>
                    <span class="meta-item">
                      <span class="meta-icon">🏢</span>
                      {doc.department}
                    </span>
                  </div>
                  
                  <div class="document-dates">
                    {#if doc.expiry_date}
                      <div class="date-item">
                        <span class="date-label">Expires:</span>
                        <span class="date-value">{new Date(doc.expiry_date).toLocaleDateString()}</span>
                      </div>
                    {/if}
                    <div class="date-item">
                      <span class="date-label">Reviewed:</span>
                      <span class="date-value">{new Date(doc.last_reviewed).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
      
    {:else if activeTab === 'training'}
      <!-- Training Tab -->
      <div class="tab-content">
        <div class="section-card">
          <div class="section-header">
            <h2>Training Compliance</h2>
            <button class="btn-primary">+ Assign Training</button>
          </div>
          
          {#if trainingModules.length === 0}
            <div class="empty-state">
              <span class="empty-icon">🎓</span>
              <p>No training modules found</p>
            </div>
          {:else}
            <div class="training-list">
              {#each trainingModules as training}
                <div class="training-card" on:click={() => viewTraining(training)}>
                  <div class="training-header">
                    <div>
                      <h3 class="training-title">{training.title}</h3>
                      <p class="training-category">{training.category}</p>
                    </div>
                    <span class="status-badge {getStatusColor(training.status)}">
                      {training.status}
                    </span>
                  </div>
                  
                  <div class="training-progress">
                    <div class="progress-header">
                      <span class="progress-label">Completion Rate</span>
                      <span class="progress-value">{training.completion_rate}%</span>
                    </div>
                    <div class="progress-bar">
                      <div class="progress-fill" style="width: {training.completion_rate}%; background: {training.completion_rate >= 90 ? '#10b981' : training.completion_rate >= 70 ? '#f59e0b' : '#ef4444'}"></div>
                    </div>
                    <div class="progress-stats">
                      <span class="stat-text">{training.completed_employees} completed</span>
                      <span class="stat-divider">•</span>
                      <span class="stat-text">{training.pending_employees} pending</span>
                      <span class="stat-divider">•</span>
                      <span class="stat-text">{training.total_employees} total</span>
                    </div>
                  </div>
                  
                  {#if training.due_date}
                    <div class="training-due">
                      <span class="due-icon">📅</span>
                      Due: {new Date(training.due_date).toLocaleDateString()}
                    </div>
                  {/if}
                  
                  {#if training.required}
                    <div class="training-badge">
                      <span class="required-badge">Required</span>
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
      
    {:else if activeTab === 'audits'}
      <!-- Audits Tab -->
      <div class="tab-content">
        <!-- Audit Filters -->
        <div class="filter-bar">
          <button class="filter-btn" class:active={auditFilter === 'all'} on:click={() => auditFilter = 'all'}>
            All ({auditItems.length})
          </button>
          <button class="filter-btn" class:active={auditFilter === 'open'} on:click={() => auditFilter = 'open'}>
            Open ({auditItems.filter(a => a.status === 'open').length})
          </button>
          <button class="filter-btn" class:active={auditFilter === 'in-progress'} on:click={() => auditFilter = 'in-progress'}>
            In Progress ({auditItems.filter(a => a.status === 'in-progress').length})
          </button>
          <button class="filter-btn" class:active={auditFilter === 'resolved'} on:click={() => auditFilter = 'resolved'}>
            Resolved ({auditItems.filter(a => a.status === 'resolved').length})
          </button>
        </div>
        
        <!-- Audits List -->
        <div class="section-card">
          {#if filteredAudits.length === 0}
            <div class="empty-state">
              <span class="empty-icon">🔍</span>
              <p>No {auditFilter === 'all' ? '' : auditFilter} audit items found</p>
            </div>
          {:else}
            <div class="audit-list">
              {#each filteredAudits as audit}
                <div class="audit-card" on:click={() => viewAudit(audit)}>
                  <div class="audit-header">
                    <div class="audit-severity" style="background-color: {getSeverityColor(audit.severity)}20; color: {getSeverityColor(audit.severity)}">
                      {audit.severity.toUpperCase()}
                    </div>
                    <span class="status-badge {getStatusColor(audit.status)}">
                      {audit.status}
                    </span>
                  </div>
                  
                  <h3 class="audit-title">{audit.title}</h3>
                  <p class="audit-description">{audit.description}</p>
                  
                  <div class="audit-meta">
                    <span class="meta-item">
                      <span class="meta-icon">🏷️</span>
                      {audit.category}
                    </span>
                    <span class="meta-item">
                      <span class="meta-icon">👤</span>
                      {audit.assigned_to}
                    </span>
                    <span class="meta-item">
                      <span class="meta-icon">📅</span>
                      Due: {new Date(audit.due_date).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}
    
  {/if}
</div>

<!-- Document Detail Modal -->
{#if showDocumentModal && selectedDocument}
  <div class="modal-overlay" on:click={() => showDocumentModal = false}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{selectedDocument.title}</h2>
          <p class="modal-subtitle">{selectedDocument.category}</p>
        </div>
        <button class="close-btn" on:click={() => showDocumentModal = false}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="detail-grid">
          <div class="detail-row">
            <span class="detail-label">Type</span>
            <span class="detail-value">{selectedDocument.type}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Status</span>
            <span class="status-badge {getStatusColor(selectedDocument.status)}">
              {selectedDocument.status}
            </span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Department</span>
            <span class="detail-value">{selectedDocument.department}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Owner</span>
            <span class="detail-value">{selectedDocument.owner}</span>
          </div>
          {#if selectedDocument.expiry_date}
            <div class="detail-row">
              <span class="detail-label">Expiry Date</span>
              <span class="detail-value">{new Date(selectedDocument.expiry_date).toLocaleDateString()}</span>
            </div>
          {/if}
          <div class="detail-row">
            <span class="detail-label">Last Reviewed</span>
            <span class="detail-value">{new Date(selectedDocument.last_reviewed).toLocaleDateString()}</span>
          </div>
        </div>
        
        {#if selectedDocument.file_url}
          <div class="modal-actions">
            <button class="btn-primary" on:click={() => window.open(selectedDocument.file_url, '_blank')}>
              📄 View Document
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Training Detail Modal -->
{#if showTrainingModal && selectedTraining}
  <div class="modal-overlay" on:click={() => showTrainingModal = false}>
    <div class="modal large" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{selectedTraining.title}</h2>
          <p class="modal-subtitle">{selectedTraining.category}</p>
        </div>
        <button class="close-btn" on:click={() => showTrainingModal = false}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="training-detail">
          <div class="completion-summary">
            <div class="summary-circle" style="background: conic-gradient(#10b981 0% {selectedTraining.completion_rate}%, #e5e7eb {selectedTraining.completion_rate}% 100%)">
              <div class="summary-inner">
                <div class="summary-value">{selectedTraining.completion_rate}%</div>
                <div class="summary-label">Complete</div>
              </div>
            </div>
            
            <div class="summary-stats">
              <div class="summary-stat">
                <span class="stat-value">{selectedTraining.completed_employees}</span>
                <span class="stat-label">Completed</span>
              </div>
              <div class="summary-stat">
                <span class="stat-value">{selectedTraining.pending_employees}</span>
                <span class="stat-label">Pending</span>
              </div>
              <div class="summary-stat">
                <span class="stat-value">{selectedTraining.total_employees}</span>
                <span class="stat-label">Total</span>
              </div>
            </div>
          </div>
          
          {#if selectedTraining.due_date}
            <div class="info-banner">
              <span class="banner-icon">📅</span>
              <div class="banner-content">
                <strong>Due Date:</strong> {new Date(selectedTraining.due_date).toLocaleDateString()}
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Audit Detail Modal -->
{#if showAuditModal && selectedAudit}
  <div class="modal-overlay" on:click={() => showAuditModal = false}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{selectedAudit.title}</h2>
          <p class="modal-subtitle">{selectedAudit.category}</p>
        </div>
        <button class="close-btn" on:click={() => showAuditModal = false}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="audit-severity-banner" style="background-color: {getSeverityColor(selectedAudit.severity)}20; border-left-color: {getSeverityColor(selectedAudit.severity)}">
          <span class="severity-icon">⚠️</span>
          <strong>{selectedAudit.severity.toUpperCase()} SEVERITY</strong>
        </div>
        
        <div class="detail-grid">
          <div class="detail-row">
            <span class="detail-label">Status</span>
            <span class="status-badge {getStatusColor(selectedAudit.status)}">
              {selectedAudit.status}
            </span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Assigned To</span>
            <span class="detail-value">{selectedAudit.assigned_to}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Due Date</span>
            <span class="detail-value">{new Date(selectedAudit.due_date).toLocaleDateString()}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Created</span>
            <span class="detail-value">{new Date(selectedAudit.created_date).toLocaleDateString()}</span>
          </div>
          <div class="detail-row full">
            <span class="detail-label">Description</span>
            <p class="detail-description">{selectedAudit.description}</p>
          </div>
          {#if selectedAudit.findings}
            <div class="detail-row full">
              <span class="detail-label">Findings</span>
              <p class="detail-description">{selectedAudit.findings}</p>
            </div>
          {/if}
          {#if selectedAudit.resolution}
            <div class="detail-row full">
              <span class="detail-label">Resolution</span>
              <p class="detail-description">{selectedAudit.resolution}</p>
            </div>
          {/if}
        </div>
        
        {#if selectedAudit.status === 'open'}
          <div class="modal-actions">
            <button class="btn-primary" on:click={() => updateAuditStatus(selectedAudit.id, 'in-progress')}>
              Start Progress
            </button>
          </div>
        {:else if selectedAudit.status === 'in-progress'}
          <div class="modal-actions">
            <button class="btn-primary" on:click={() => updateAuditStatus(selectedAudit.id, 'resolved')}>
              Mark Resolved
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .compliance-dashboard {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .header h1 {
    font-size: 28px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .subtitle {
    font-size: 14px;
    color: #6b7280;
    margin: 0;
  }
  
  .tabs {
    display: flex;
    gap: 8px;
    background: white;
    padding: 8px;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }
  
  .tab {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 16px;
    background: none;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    color: #6b7280;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
  }
  
  .tab:hover {
    background: #f9fafb;
    color: #111827;
  }
  
  .tab.active {
    background: #8b5cf6;
    color: white;
  }
  
  .tab-icon {
    font-size: 18px;
  }
  
  .tab-badge {
    position: absolute;
    top: 4px;
    right: 4px;
    padding: 2px 8px;
    background: #ef4444;
    color: white;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
  }
  
  .tab-content {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 16px;
  }
  
  .stat-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 24px;
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }
  
  .stat-card.large {
    justify-content: center;
  }
  
  .stat-card.highlight {
    background: linear-gradient(135deg, #fee2e2 0%, #fecaca 100%);
    border-left: 4px solid #ef4444;
  }
  
  .score-circle {
    width: 140px;
    height: 140px;
    border-radius: 50%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    border: 8px solid;
  }
  
  .score-circle.critical {
    border-color: #ef4444;
    background: #fee2e2;
  }
  
  .score-circle.warning {
    border-color: #f59e0b;
    background: #fef3c7;
  }
  
  .score-circle.good {
    border-color: #10b981;
    background: #d1fae5;
  }
  
  .score-value {
    font-size: 36px;
    font-weight: 700;
  }
  
  .score-circle.critical .score-value {
    color: #ef4444;
  }
  
  .score-circle.warning .score-value {
    color: #f59e0b;
  }
  
  .score-circle.good .score-value {
    color: #10b981;
  }
  
  .score-label {
    font-size: 12px;
    color: #6b7280;
    margin-top: 4px;
    text-align: center;
  }
  
  .stat-icon {
    font-size: 32px;
  }
  
  .stat-content {
    flex: 1;
  }
  
  .stat-value {
    font-size: 28px;
    font-weight: 700;
    color: #111827;
    margin-bottom: 4px;
  }
  
  .stat-label {
    font-size: 14px;
    color: #111827;
    font-weight: 500;
  }
  
  .stat-sublabel {
    font-size: 12px;
    color: #6b7280;
    margin-top: 2px;
  }
  
  .section-card {
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    overflow: hidden;
  }
  
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .section-header h2 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }
  
  .btn-primary {
    padding: 10px 20px;
    background: #8b5cf6;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
  }
  
  .btn-primary:hover {
    background: #7c3aed;
  }
  
  .categories-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 16px;
    padding: 24px;
  }
  
  .category-card {
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    transition: all 0.2s;
  }
  
  .category-card:hover {
    border-color: #8b5cf6;
    box-shadow: 0 2px 8px rgba(139, 92, 246, 0.1);
  }
  
  .category-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 16px;
  }
  
  .category-header h3 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }
  
  .status-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }
  
  .bg-green-100 { background: #d1fae5; }
  .text-green-800 { color: #065f46; }
  .bg-yellow-100 { background: #fef3c7; }
  .text-yellow-800 { color: #92400e; }
  .bg-red-100 { background: #fee2e2; }
  .text-red-800 { color: #991b1b; }
  .bg-blue-100 { background: #dbeafe; }
  .text-blue-800 { color: #1e40af; }
  .bg-purple-100 { background: #ede9fe; }
  .text-purple-800 { color: #5b21b6; }
  .bg-gray-100 { background: #f3f4f6; }
  .text-gray-800 { color: #1f2937; }
  
  .category-score {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }
  
  .score-bar {
    flex: 1;
    height: 8px;
    background: #e5e7eb;
    border-radius: 4px;
    overflow: hidden;
  }
  
  .score-fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.3s;
  }
  
  .score-text {
    font-size: 14px;
    font-weight: 600;
    color: #374151;
    min-width: 40px;
  }
  
  .category-details {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 13px;
    color: #6b7280;
  }
  
  .detail-item {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  
  .detail-icon {
    font-size: 14px;
  }
  
  .filter-bar {
    display: flex;
    gap: 8px;
    background: white;
    padding: 16px;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    flex-wrap: wrap;
  }
  
  .filter-btn {
    padding: 8px 16px;
    background: white;
    color: #6b7280;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .filter-btn:hover {
    border-color: #8b5cf6;
    color: #8b5cf6;
  }
  
  .filter-btn.active {
    background: #8b5cf6;
    color: white;
    border-color: #8b5cf6;
  }
  
  .documents-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
    padding: 24px;
  }
  
  .document-card {
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .document-card:hover {
    border-color: #8b5cf6;
    box-shadow: 0 2px 8px rgba(139, 92, 246, 0.1);
    transform: translateY(-2px);
  }
  
  .document-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 12px;
  }
  
  .document-icon {
    font-size: 32px;
  }
  
  .document-title {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 12px 0;
  }
  
  .document-meta,
  .audit-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 12px;
    font-size: 13px;
    color: #6b7280;
  }
  
  .meta-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  
  .meta-icon {
    font-size: 14px;
  }
  
  .document-dates {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-top: 12px;
    border-top: 1px solid #e5e7eb;
    font-size: 12px;
  }
  
  .date-item {
    display: flex;
    justify-content: space-between;
  }
  
  .date-label {
    color: #6b7280;
  }
  
  .date-value {
    color: #111827;
    font-weight: 500;
  }
  
  .training-list,
  .audit-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 24px;
  }
  
  .training-card,
  .audit-card {
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .training-card:hover,
  .audit-card:hover {
    border-color: #8b5cf6;
    box-shadow: 0 2px 8px rgba(139, 92, 246, 0.1);
  }
  
  .training-header,
  .audit-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 16px;
  }
  
  .training-title,
  .audit-title {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .training-category,
  .audit-description {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .training-progress {
    margin-bottom: 12px;
  }
  
  .progress-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
    font-size: 13px;
  }
  
  .progress-label {
    color: #6b7280;
  }
  
  .progress-value {
    color: #111827;
    font-weight: 600;
  }
  
  .progress-bar {
    width: 100%;
    height: 8px;
    background: #e5e7eb;
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 8px;
  }
  
  .progress-fill {
    height: 100%;
    border-radius: 4px;
    transition: width 0.3s;
  }
  
  .progress-stats {
    display: flex;
    gap: 8px;
    font-size: 12px;
    color: #6b7280;
  }
  
  .stat-text {
    color: #6b7280;
  }
  
  .stat-divider {
    color: #d1d5db;
  }
  
  .training-due {
    display: flex;
    align-items: center;
    gap: 6px;
    padding-top: 12px;
    border-top: 1px solid #e5e7eb;
    font-size: 13px;
    color: #6b7280;
  }
  
  .due-icon {
    font-size: 14px;
  }
  
  .training-badge {
    margin-top: 8px;
  }
  
  .required-badge {
    display: inline-block;
    padding: 4px 12px;
    background: #ede9fe;
    color: #5b21b6;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }
  
  .audit-severity {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 700;
  }
  
  .loading,
  .empty-state {
    text-align: center;
    padding: 48px 24px;
    color: #6b7280;
  }
  
  .empty-icon {
    font-size: 64px;
    display: block;
    margin-bottom: 16px;
  }
  
  /* Modal Styles */
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
  
  .modal {
    background: white;
    border-radius: 12px;
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
  }
  
  .modal.large {
    max-width: 800px;
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
    margin: 0 0 4px 0;
  }
  
  .modal-subtitle {
    font-size: 13px;
    color: #6b7280;
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
  }
  
  .close-btn:hover {
    background: #f3f4f6;
  }
  
  .modal-body {
    padding: 24px;
  }
  
  .detail-grid {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  
  .detail-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .detail-row.full {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .detail-label {
    font-size: 13px;
    font-weight: 500;
    color: #6b7280;
  }
  
  .detail-value {
    font-size: 14px;
    color: #111827;
  }
  
  .detail-description {
    font-size: 14px;
    color: #374151;
    line-height: 1.5;
    margin: 0;
  }
  
  .modal-actions {
    display: flex;
    gap: 12px;
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid #e5e7eb;
  }
  
  .training-detail {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .completion-summary {
    display: flex;
    gap: 32px;
    align-items: center;
  }
  
  .summary-circle {
    width: 150px;
    height: 150px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
  }
  
  .summary-inner {
    width: 110px;
    height: 110px;
    background: white;
    border-radius: 50%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
  
  .summary-value {
    font-size: 32px;
    font-weight: 700;
    color: #111827;
  }
  
  .summary-label {
    font-size: 13px;
    color: #6b7280;
  }
  
  .summary-stats {
    flex: 1;
    display: flex;
    gap: 24px;
  }
  
  .summary-stat {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  
  .summary-stat .stat-value {
    font-size: 24px;
    font-weight: 700;
    color: #111827;
  }
  
  .summary-stat .stat-label {
    font-size: 13px;
    color: #6b7280;
  }
  
  .info-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;
    background: #eff6ff;
    border-left: 4px solid #3b82f6;
    border-radius: 8px;
  }
  
  .banner-icon {
    font-size: 24px;
  }
  
  .banner-content {
    font-size: 14px;
    color: #1e40af;
  }
  
  .audit-severity-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;
    border-left: 4px solid;
    border-radius: 8px;
    margin-bottom: 20px;
    font-size: 14px;
  }
  
  .severity-icon {
    font-size: 24px;
  }
  
  @media (max-width: 1024px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }
    
    .categories-grid,
    .documents-grid {
      grid-template-columns: 1fr;
    }
    
    .completion-summary {
      flex-direction: column;
    }
  }
  
  @media (max-width: 768px) {
    .tabs {
      flex-wrap: wrap;
    }
    
    .tab {
      flex: 1 1 calc(50% - 4px);
      min-width: 140px;
    }
    
    .filter-bar {
      flex-direction: column;
    }
    
    .filter-btn {
      width: 100%;
    }
  }
</style>