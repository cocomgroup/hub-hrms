<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  import EmployeeDetail from '../components/EmployeeDetail.svelte';
  import WorkflowManager from '../routes/talent-center/workflows/WorkflowManager.svelte';
  import Compensation from '../components/Compensation.svelte';
  
  const API_BASE_URL = getApiBaseUrl();

  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    department: string;
    position: string;
    status: 'onboarding' | 'active' | 'on-leave' | 'separated';
    hire_date: string;
  }

  interface DashboardStats {
    totalEmployees: number;
    activeEmployees: number;
    onboardingEmployees: number;
    onLeaveEmployees: number;
  }

  let employees: Employee[] = [];
  let stats: DashboardStats = {
    totalEmployees: 0,
    activeEmployees: 0,
    onboardingEmployees: 0,
    onLeaveEmployees: 0
  };
  let loading = true;
  let error = '';
  let selectedEmployeeId: string | null = null;
  let showEmployeeDetail = false;
  let filterStatus: string = 'all';
  let searchQuery = '';
  let showCompensation = false;
  let showWorkflowManager = false;
  let currentPage = 1;
  let itemsPerPage = 10;

  onMount(() => {
    loadDashboard();
  });

  async function loadDashboard() {
    try {
      loading = true;
      error = '';

      const response = await fetch(`${API_BASE_URL}/employees`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });

      if (!response.ok) throw new Error('Failed to load employees');
      
      employees = await response.json();
      
      // Calculate stats
      stats.totalEmployees = employees.length;
      stats.activeEmployees = employees.filter(e => e.status === 'active').length;
      stats.onboardingEmployees = employees.filter(e => e.status === 'onboarding').length;
      stats.onLeaveEmployees = employees.filter(e => e.status === 'on-leave').length;

    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function openEmployeeDetail(employeeId: string) {
    selectedEmployeeId = employeeId;
    showEmployeeDetail = true;
  }

  function closeEmployeeDetail() {
    showEmployeeDetail = false;
    selectedEmployeeId = null;
  }

  function getStatusBadgeClass(status: string): string {
    const classes = {
      'onboarding': 'status-onboarding',
      'active': 'status-active',
      'on-leave': 'status-leave',
      'separated': 'status-separated'
    };
    return classes[status] || 'status-default';
  }

  function getStatusLabel(status: string): string {
    const labels = {
      'onboarding': 'Onboarding',
      'active': 'Active',
      'on-leave': 'On Leave',
      'separated': 'Separated'
    };
    return labels[status] || status;
  }
  $: filteredEmployees = employees.filter(emp => {
    const matchesStatus = filterStatus === 'all' || emp.status === filterStatus;
    const matchesSearch = !searchQuery || 
      emp.first_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      emp.last_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      emp.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
      emp.department.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesStatus && matchesSearch;
  });

  $: totalPages = Math.ceil(filteredEmployees.length / itemsPerPage);
  $: paginatedEmployees = filteredEmployees.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );

  function nextPage() {
    if (currentPage < totalPages) {
      currentPage++;
    }
  }

  function prevPage() {
    if (currentPage > 1) {
      currentPage--;
    }
  }

  function goToPage(page: number) {
    currentPage = page;
  }

  // Reset to page 1 when filter changes
  $: if (filterStatus || searchQuery) {
    currentPage = 1;
  }
</script>

<div class="hr-dashboard">
  <div class="dashboard-header">
    <h1>HR Dashboard</h1>
  </div>



  {#if loading}
    <div class="loading">Loading dashboard...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else}
    <!-- Stats Cards -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">{stats.totalEmployees}</div>
        <div class="stat-label">Total Employees</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{stats.activeEmployees}</div>
        <div class="stat-label">Active</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{stats.onboardingEmployees}</div>
        <div class="stat-label">Onboarding</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{stats.onLeaveEmployees}</div>
        <div class="stat-label">On Leave</div>
      </div>
    </div>

    
    <!-- Quick Actions -->
    <div class="section">
      <div class="section-header">
        <h2>Quick Actions</h2>
      </div>

      <div class="actions-grid">

        <button class="action-card" onclick={()  => showWorkflowManager = true}> 
          <div class="action-icon">💰</div>
          <div class="action-content">
            <h3>Workflow Manager</h3>
            <p>Manage onboarding and HR workflows</p>
          </div>
        </button>

        <button class="action-card" onclick={() => showCompensation = true}>
          <div class="action-icon">📋</div>
          <div class="action-content">
            <h3>Compensation</h3>
            <p>Manage salaries, bonuses & pay</p>
          </div>
        </button>

      </div>
    </div>

    <!-- Filters -->
    <div class="filters">
      <input 
        type="text" 
        placeholder="Search employees..."
        bind:value={searchQuery}
        class="search-input"
      />
      <select bind:value={filterStatus} class="filter-select">
        <option value="all">All Statuses</option>
        <option value="active">Active</option>
        <option value="onboarding">Onboarding</option>
        <option value="on-leave">On Leave</option>
        <option value="separated">Separated</option>
      </select>
    </div>

    <!-- Employees Table -->
    <div class="employees-table">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Department</th>
            <th>Position</th>
            <th>Status</th>
            <th>Hire Date</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#if paginatedEmployees.length === 0 && filteredEmployees.length === 0}
            <tr>
              <td colspan="7" class="empty-state">
                No employees found
              </td>
            </tr>
          {:else}
            {#each paginatedEmployees as employee}
              <tr>
                <td class="employee-name">
                  {employee.first_name} {employee.last_name}
                </td>
                <td>{employee.email}</td>
                <td>{employee.department}</td>
                <td>{employee.position}</td>
                <td>
                  <span class="status-badge {getStatusBadgeClass(employee.status)}">
                    {getStatusLabel(employee.status)}
                  </span>
                </td>
                <td>{new Date(employee.hire_date).toLocaleDateString()}</td>
                <td>
                  <button 
                    class="btn-view"
                    onclick={() => openEmployeeDetail(employee.id)}
                  >
                    View Details
                  </button>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>

    <!-- Pagination Controls -->
    {#if filteredEmployees.length > itemsPerPage}
      <div class="pagination">
        <div class="pagination-info">
          Showing {((currentPage - 1) * itemsPerPage) + 1} to {Math.min(currentPage * itemsPerPage, filteredEmployees.length)} of {filteredEmployees.length} employees
        </div>
        <div class="pagination-controls">
          <button 
            class="btn-page" 
            onclick={prevPage} 
            disabled={currentPage === 1}
          >
            ← Previous
          </button>
          
          <div class="page-numbers">
            {#each Array(totalPages) as _, i}
              <button
                class="btn-page-num {currentPage === i + 1 ? 'active' : ''}"
                onclick={() => goToPage(i + 1)}
              >
                {i + 1}
              </button>
            {/each}
          </div>
          
          <button 
            class="btn-page" 
            onclick={nextPage} 
            disabled={currentPage === totalPages}
          >
            Next →
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>

<!-- Employee Detail Modal -->
{#if showEmployeeDetail && selectedEmployeeId}
  <div class="modal-overlay" onclick={closeEmployeeDetail} onkeydown={(e) => e.key === 'Escape' && closeEmployeeDetail()} role="button" tabindex="0">
    <div class="modal" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && closeEmployeeDetail()} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Employee Details</h2>
        <button class="close-btn" onclick={closeEmployeeDetail}>×</button>
      </div>
      <div class="modal-body">
        <EmployeeDetail employeeId={selectedEmployeeId} />
      </div>
    </div>
  </div>
{/if}

<!-- Workflow Manager Modal -->
{#if showWorkflowManager}
  <div class="modal-overlay" onclick={() => showWorkflowManager = false} onkeydown={(e) => e.key === 'Escape' && (showWorkflowManager = false)} role="button" tabindex="0">
    <div class="modal full-screen" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showWorkflowManager = false)} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Workflow Management</h2>
        <button class="close-btn" onclick={() => showWorkflowManager = false}>×</button>
      </div>
      <WorkflowManager />
    </div>
  </div>
{/if}

<!-- Compensation Modal -->
{#if showCompensation}
  <div class="modal-overlay" onclick={() => showCompensation = false} onkeydown={(e) => e.key === 'Escape' && (showCompensation = false)} role="button" tabindex="0">
    <div class="modal full-screen" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showCompensation = false)} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Compensation Management</h2>
        <button class="close-btn" onclick={() => showCompensation = false}>×</button>
      </div>
      <Compensation />
    </div>
  </div>
{/if}

<style>
  .hr-dashboard {
    padding: 24px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .dashboard-header {
    margin-bottom: 24px;
  }

  .dashboard-header h1 {
    margin: 0;
    font-size: 28px;
    font-weight: 600;
  }

  .loading, .error {
    text-align: center;
    padding: 48px;
    color: #666;
  }

  .error {
    color: #dc3545;
  }

  /* Stats Cards */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
    margin-bottom: 32px;
  }

  .stat-card {
    background: white;
    padding: 24px;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    text-align: center;
  }

  .stat-value {
    font-size: 36px;
    font-weight: 700;
    color: #007bff;
    margin-bottom: 8px;
  }

  .stat-label {
    font-size: 14px;
    color: #666;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  /* Filters */
  .filters {
    display: flex;
    gap: 16px;
    margin-bottom: 24px;
  }

  .search-input {
    flex: 1;
    padding: 10px 16px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
  }

  .filter-select {
    padding: 10px 16px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 14px;
    background: white;
    cursor: pointer;
  }

  /* Table */
  .employees-table {
    background: white;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    overflow: hidden;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  thead {
    background: #f8f9fa;
  }

  th {
    padding: 16px;
    text-align: left;
    font-size: 12px;
    font-weight: 600;
    color: #666;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    border-bottom: 2px solid #e9ecef;
  }

  td {
    color: #212529; 
    padding: 16px;
    border-bottom: 1px solid #e9ecef;
    font-size: 14px;
  }

  tr:hover {
    background: #f8f9fa;
  }

  .employee-name {
    font-weight: 500;
  }

  .status-badge {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-onboarding { background: #fff3cd; color: #856404; }
  .status-active { background: #d4edda; color: #155724; }
  .status-leave { background: #d1ecf1; color: #0c5460; }
  .status-separated { background: #f8d7da; color: #721c24; }

  .btn-view {
    padding: 6px 16px;
    background: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
  }

  .btn-view:hover {
    background: #0056b3;
  }

  .empty-state {
    text-align: center;
    padding: 48px;
    color: #999;
    font-style: italic;
  }

  /* Modal */
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
    border-radius: 8px;
    width: 90%;
    max-width: 900px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #e9ecef;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 20px;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 32px;
    line-height: 1;
    color: #999;
    cursor: pointer;
    padding: 0;
    width: 32px;
    height: 32px;
  }

  .close-btn:hover {
    color: #333;
  }

  .modal-body {
    padding: 0;
  }
  .pagination {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    background: white;
    border-top: 1px solid #e9ecef;
    border-radius: 0 0 8px 8px;
  }

  .pagination-info {
    font-size: 14px;
    color: #6b7280;
  }

  .pagination-controls {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .page-numbers {
    display: flex;
    gap: 4px;
  }

  .btn-page {
    padding: 8px 16px;
    border: 1px solid #d1d5db;
    background: white;
    color: #374151;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    transition: all 0.2s;
  }

  .btn-page:hover:not(:disabled) {
    background: #f3f4f6;
    border-color: #9ca3af;
  }

  .btn-page:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-page-num {
    padding: 8px 12px;
    border: 1px solid #d1d5db;
    background: white;
    color: #374151;
    border-radius: 6px;
    cursor: pointer;
    font-size: 14px;
    min-width: 40px;
    transition: all 0.2s;
  }

  .btn-page-num:hover {
    background: #f3f4f6;
    border-color: #9ca3af;
  }

  .btn-page-num.active {
    background: #3b82f6;
    color: white;
    border-color: #3b82f6;
  }

    /* Actions Grid */
  .actions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1rem;
  }

  .action-card {
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 12px;
    padding: 1.5rem;
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .action-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.1);
    transform: translateY(-2px);
  }

  .action-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .action-content {
    text-align: left;
  }

  .action-content h3 {
    margin: 0 0 0.25rem 0;
    font-size: 1rem;
    color: #1f2937;
  }

  .action-content p {
    margin: 0;
    font-size: 0.875rem;
    color: #6b7280;
  }
</style>