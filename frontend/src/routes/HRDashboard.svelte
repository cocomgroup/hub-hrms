<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  import EmployeeDetail from '../components/EmployeeDetail.svelte';
  import Compensation from '../components/Compensation.svelte';
  import EmployeeList from '../components/EmployeeList.svelte';
  import OrganizationDashboard from '../components/OrganizationDashboard.svelte';
  import TalentCenter from './talent-center/TalentCenter.svelte';
  import PayExpenses from './pay-expenses/PayExpenses.svelte';
  import ComplianceDashboard from './compliance/ComplianceDashboard.svelte';
  import BenefitsDashboard from './benefits/BenefitsDashboard.svelte';
  
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
  let showCompensation = false;
  let showTalentCenter = false;
  let showOrganizationManagment = false;
  let showPayExpenses = false;
  let showCompliance = false;
  let showBenefits = false;
  let showEmployeeList = false; // New: for Quick Actions employee list modal

  onMount(() => {
    loadDashboard();
  });

  async function loadDashboard() {
    try {
      loading = true;
      error = '';

      const response = await fetch(`${API_BASE_URL}/employees`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });

      if (!response.ok) throw new Error('Failed to load employees');

      employees = await response.json();
      calculateStats();
    } catch (err: any) {
      error = err.message;
      console.error('Dashboard error:', err);
    } finally {
      loading = false;
    }
  }

  function calculateStats() {
    stats.totalEmployees = employees.length;
    stats.activeEmployees = employees.filter(e => e.status === 'active').length;
    stats.onboardingEmployees = employees.filter(e => e.status === 'onboarding').length;
    stats.onLeaveEmployees = employees.filter(e => e.status === 'on-leave').length;
  }

  function openEmployeeDetail(employeeId: string) {
    selectedEmployeeId = employeeId;
    showEmployeeDetail = true;
  }

  function closeEmployeeDetail() {
    showEmployeeDetail = false;
    selectedEmployeeId = null;
  }

  function openEmployeeListModal() {
    showEmployeeList = true;
  }
</script>

<div class="hr-dashboard">
  <div class="dashboard-header">
    <h1>HR Dashboard</h1>
  </div>

  {#if loading && employees.length === 0}
    <div class="loading">Loading dashboard...</div>
  {:else if error}
    <div class="error">
      <p>Error: {error}</p>
      <button onclick={loadDashboard}>Retry</button>
    </div>
  {:else}
    <!-- Stats Cards -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">👥</div>
        <div class="stat-content">
          <div class="stat-value">{stats.totalEmployees}</div>
          <div class="stat-label">Total Employees</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">✓</div>
        <div class="stat-content">
          <div class="stat-value">{stats.activeEmployees}</div>
          <div class="stat-label">Active</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">🚀</div>
        <div class="stat-content">
          <div class="stat-value">{stats.onboardingEmployees}</div>
          <div class="stat-label">Onboarding</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">🏖️</div>
        <div class="stat-content">
          <div class="stat-value">{stats.onLeaveEmployees}</div>
          <div class="stat-label">On Leave</div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="section">
      <h2>Quick Actions</h2>
      <div class="quick-actions-grid">
        <!--
        <button class="action-card" onclick={openEmployeeListModal}>
          <div class="action-icon">👥</div>
          <div class="action-content">
            <h3>View All Employees</h3>
            <p>Search, filter & manage employees</p>
          </div>
        </button>
      -->
        <button class="action-card" onclick={() => showOrganizationManagment = true}>
          <div class="action-icon">👥</div>
          <div class="action-content">
            <h3>Organization Management</h3>
            <p>Manage employees & organization</p>
          </div>
        </button>

      <button class="action-card" onclick={() => showTalentCenter = true}>
        <div class="action-icon">🎯</div>
        <div class="action-content">
          <h3>Talent Center</h3>
          <p>Recruiting, onboarding & workflows</p>
        </div>
      </button>

        <button class="action-card" onclick={() => showPayExpenses = true}>
          <div class="action-icon">💳</div>
          <div class="action-content">
            <h3>Pay & Expenses</h3>
            <p>Payroll & expense approvals</p>
          </div>
        </button>

        <button class="action-card" onclick={() => showCompliance = true}>
          <div class="action-icon">⚖️</div>
          <div class="action-content">
            <h3>Compliance</h3>
            <p>Documents, training & audits</p>
          </div>
        </button>

        <button class="action-card" onclick={() => showBenefits = true}>
          <div class="action-icon">🏥</div>
          <div class="action-content">
            <h3>Benefits</h3>
            <p>Healthcare, retirement & enrollment</p>
          </div>
        </button>

        <button class="action-card" onclick={() => showCompensation = true}>
          <div class="action-icon">💰</div>
          <div class="action-content">
            <h3>Compensation</h3>
            <p>Manage salaries & bonuses</p>
          </div>
        </button>

      </div>
    </div>

    <!-- Recent Activity / Summary can go here -->
    <div class="section">
      <h2>Recent Activity</h2>
      <div class="recent-activity">
        <p class="text-muted">Recent employee activities will appear here</p>
      </div>
    </div>
  {/if}
</div>

<!-- Employee List Modal -->
{#if showEmployeeList}
  <div class="modal-overlay" onclick={() => showEmployeeList = false} onkeydown={(e) => e.key === 'Escape' && (showEmployeeList = false)} role="button" tabindex="0">
    <div class="modal full-screen" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showEmployeeList = false)} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>All Employees</h2>
        <button class="close-btn" onclick={() => showEmployeeList = false}>×</button>
      </div>
      <div class="modal-body">
        <EmployeeList 
          {employees} 
          loading={false}
          onViewDetails={openEmployeeDetail}
        />
      </div>
    </div>
  </div>
{/if}

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

<!-- Organization Management Modal -->
{#if showOrganizationManagment}
  <div class="modal-overlay" onclick={() => showOrganizationManagment = false} onkeydown={(e) => e.key === 'Escape' && (showOrganizationManagment = false)} role="button" tabindex="0">
    <div class="modal full-screen" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showOrganizationManagment = false)} role="dialog" aria-modal="true" tabindex="-1">
      <div class="title-section h1">
        <h1>🎯 Talent Center</h1>
        <p class="subtitle">Employee and Organization Management</p>

        <button class="close-btn" onclick={() => showOrganizationManagment = false}>×</button>
      </div>
      <OrganizationDashboard />
    </div>
  </div>
{/if}

<!-- Talent Center Modal -->
{#if showTalentCenter}
  <div class="modal-overlay" onclick={() => showTalentCenter = false} onkeydown={(e) => e.key === 'Escape' && (showTalentCenter = false)} role="button" tabindex="0">
    <div class="modal full-screen" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showTalentCenter = false)} role="dialog" aria-modal="true" tabindex="-1">
      <div class="title-section h1">
        <h1>🎯 Talent Center</h1>
        <p class="subtitle">Recruiting, Onboarding & Workflow Management</p>

        <button class="close-btn" onclick={() => showTalentCenter = false}>×</button>
      </div>
      <TalentCenter />
    </div>
  </div>
{/if}

<!-- Pay & Expenses Modal -->
{#if showPayExpenses}
  <div class="modal-overlay" onclick={() => showPayExpenses = false} onkeydown={(e) => e.key === 'Escape' && (showPayExpenses = false)} role="button" tabindex="0">
    <div class="modal full-screen" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showPayExpenses = false)} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Pay & Expenses</h2>
        <button class="close-btn" onclick={() => showPayExpenses = false}>×</button>
      </div>
      <div class="modal-body">
        <PayExpenses />
      </div>
    </div>
  </div>
{/if}

<!-- Compliance Modal -->
{#if showCompliance}
  <div class="modal-overlay" onclick={() => showCompliance = false} onkeydown={(e) => e.key === 'Escape' && (showCompliance = false)} role="button" tabindex="0">
    <div class="modal full-screen" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showCompliance = false)} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Compliance Dashboard</h2>
        <button class="close-btn" onclick={() => showCompliance = false}>×</button>
      </div>
      <div class="modal-body">
        <ComplianceDashboard />
      </div>
    </div>
  </div>
{/if}

<!-- Benefits Modal -->
{#if showBenefits}
  <div class="modal-overlay" onclick={() => showBenefits = false} onkeydown={(e) => e.key === 'Escape' && (showBenefits = false)} role="button" tabindex="0">
    <div class="modal full-screen" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showBenefits = false)} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Benefits Dashboard</h2>
        <button class="close-btn" onclick={() => showBenefits = false}>×</button>
      </div>
      <div class="modal-body">
        <BenefitsDashboard />
      </div>
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
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
  }

  .dashboard-header {
    margin-bottom: 32px;
  }

  .dashboard-header h1 {
    font-size: 32px;
    font-weight: 700;
    color: #111827;
    margin: 0;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 20px;
    margin-bottom: 32px;
  }

  .stat-card {
    background: white;
    padding: 24px;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .stat-icon {
    font-size: 36px;
    width: 60px;
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f3f4f6;
    border-radius: 12px;
  }

  .stat-content {
    flex: 1;
  }

  .stat-value {
    font-size: 28px;
    font-weight: 600;
    color: #111827;
    margin-bottom: 4px;
  }

  .stat-label {
    font-size: 14px;
    color: #6b7280;
  }

  .section {
    background: white;
    padding: 24px;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    margin-bottom: 24px;
  }

  .section h2 {
    font-size: 20px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 20px 0;
  }

  .title-section h1 {
    font-size: 32px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .subtitle {
    font-size: 16px;
    opacity: 0.9;
    margin: 0 0 24px 0;
    color: #111827
  }

  .quick-actions-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 16px;
  }

  .action-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    background: white;
    border: 2px solid #e5e7eb;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
    width: 100%;
  }

  .action-card:hover {
    border-color: #3b82f6;
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  }

  .action-icon {
    font-size: 32px;
    width: 60px;
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f3f4f6;
    border-radius: 12px;
  }

  .action-content h3 {
    margin: 0 0 4px 0;
    font-size: 16px;
    font-weight: 600;
    color: #111827;
  }

  .action-content p {
    margin: 0;
    font-size: 14px;
    color: #6b7280;
  }

  .recent-activity {
    padding: 32px;
    text-align: center;
  }

  .text-muted {
    color: #6b7280;
    font-style: italic;
  }

  .loading, .error {
    text-align: center;
    padding: 48px;
    color: #666;
  }

  .error {
    color: #dc3545;
  }

  .error button {
    margin-top: 16px;
    padding: 10px 20px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
  }

  /* Modal Styles */
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
    padding: 20px;
  }

  .modal {
    background: white;
    border-radius: 12px;
    max-width: 800px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }

  .modal.full-screen {
    max-width: 95vw;
    max-height: 95vh;
    width: 95vw;
    height: 95vh;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
    color: #111827;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 32px;
    cursor: pointer;
    color: #6b7280;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 6px;
    transition: background 0.2s;
  }

  .close-btn:hover {
    background: #f3f4f6;
  }

  .modal-body {
    padding: 24px;
  }

  @media (max-width: 768px) {
    .hr-dashboard {
      padding: 16px;
    }

    .dashboard-header h1 {
      font-size: 24px;
    }

    .stats-grid {
      grid-template-columns: 1fr;
    }

    .quick-actions-grid {
      grid-template-columns: 1fr;
    }

    .modal.full-screen {
      max-width: 100vw;
      max-height: 100vh;
      width: 100vw;
      height: 100vh;
      border-radius: 0;
    }
  }
</style>