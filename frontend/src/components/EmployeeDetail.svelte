<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  export let employeeId: string;

  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    phone: string;
    hire_date: string;
    status: 'onboarding' | 'active' | 'on-leave' | 'separated';
    department: string;
    position: string;
    manager_id?: string;
    manager_name?: string;
  }

  interface Compensation {
    id: string;
    employee_id: string;
    employment_type: 'W2-salary' | 'W2-hourly' | '1099' | 'PRN';
    pay_type: 'hourly' | 'salary';
    hourly_rate?: number;
    annual_salary?: number;
    pay_frequency: string;
    effective_date: string;
  }

  interface Benefit {
    id: string;
    name: string;
    type: string;
    provider: string;
    enrolled_date: string;
    status: 'active' | 'pending' | 'terminated';
    employee_contribution?: number;
    employer_contribution?: number;
  }

  let employee: Employee | null = null;
  let compensation: Compensation | null = null;
  let benefits: Benefit[] = [];
  let loading = true;
  let error = '';
  let activeTab: 'general' | 'compensation' | 'benefits' = 'general';

  onMount(() => {
    loadEmployeeData();
  });

  async function loadEmployeeData() {
    try {
      loading = true;
      error = '';

      // Load employee general info
      const empResponse = await fetch(`${API_BASE_URL}/employees/${employeeId}`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      if (!empResponse.ok) throw new Error('Failed to load employee');
      employee = await empResponse.json();

      // Load compensation
      const compResponse = await fetch(`${API_BASE_URL}/payroll/compensation/${employeeId}`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      if (compResponse.ok) {
        compensation = await compResponse.json();
      }

      // Load benefits
      const benefitsResponse = await fetch(`${API_BASE_URL}/benefits/employee/${employeeId}`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      if (benefitsResponse.ok) {
        benefits = await benefitsResponse.json();
      }

    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
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
</script>

<div class="employee-detail">
  {#if loading}
    <div class="loading">Loading employee details...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if employee}
    <!-- Header -->
    <div class="employee-header">
      <div class="employee-name">
        <h2>{employee.first_name} {employee.last_name}</h2>
        <span class="status-badge {getStatusBadgeClass(employee.status)}">
          {getStatusLabel(employee.status)}
        </span>
      </div>
      <div class="employee-meta">
        <span>{employee.position}</span>
        <span>•</span>
        <span>{employee.department}</span>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button 
        class="tab" 
        class:active={activeTab === 'general'}
        onclick={() => activeTab = 'general'}
      >
        General Information
      </button>
      <button 
        class="tab" 
        class:active={activeTab === 'compensation'}
        onclick={() => activeTab = 'compensation'}
      >
        Compensation
      </button>
      <button 
        class="tab" 
        class:active={activeTab === 'benefits'}
        onclick={() => activeTab = 'benefits'}
      >
        Benefits
      </button>
    </div>

    <!-- Tab Content -->
    <div class="tab-content">
      {#if activeTab === 'general'}
        <div class="info-section">
          <h3>Contact Information</h3>
          <div class="info-grid">
            <div class="info-item">
              <label>Email</label>
              <div>{employee.email}</div>
            </div>
            <div class="info-item">
              <label>Phone</label>
              <div>{employee.phone || 'Not provided'}</div>
            </div>
          </div>

          <h3>Employment Details</h3>
          <div class="info-grid">
            <div class="info-item">
              <label>Hire Date</label>
              <div>{new Date(employee.hire_date).toLocaleDateString()}</div>
            </div>
            <div class="info-item">
              <label>Status</label>
              <div>{getStatusLabel(employee.status)}</div>
            </div>
            <div class="info-item">
              <label>Department</label>
              <div>{employee.department}</div>
            </div>
            <div class="info-item">
              <label>Position</label>
              <div>{employee.position}</div>
            </div>
            {#if employee.manager_name}
              <div class="info-item">
                <label>Manager</label>
                <div>{employee.manager_name}</div>
              </div>
            {/if}
          </div>
        </div>

      {:else if activeTab === 'compensation'}
        <div class="info-section">
          {#if compensation}
            <h3>Current Compensation</h3>
            <div class="info-grid">
              <div class="info-item">
                <label>Employment Type</label>
                <div class="employment-type">
                  <span class="badge badge-{compensation.employment_type}">
                    {compensation.employment_type}
                  </span>
                </div>
              </div>
              <div class="info-item">
                <label>Pay Type</label>
                <div>{compensation.pay_type === 'hourly' ? 'Hourly' : 'Salary'}</div>
              </div>
              
              {#if compensation.pay_type === 'hourly'}
                <div class="info-item">
                  <label>Hourly Rate</label>
                  <div class="pay-amount">${compensation.hourly_rate?.toFixed(2)}/hr</div>
                </div>
              {:else}
                <div class="info-item">
                  <label>Annual Salary</label>
                  <div class="pay-amount">${compensation.annual_salary?.toLocaleString()}/year</div>
                </div>
              {/if}

              <div class="info-item">
                <label>Pay Frequency</label>
                <div>{compensation.pay_frequency}</div>
              </div>
              <div class="info-item">
                <label>Effective Date</label>
                <div>{new Date(compensation.effective_date).toLocaleDateString()}</div>
              </div>
            </div>
          {:else}
            <div class="empty-state">
              No compensation information available
            </div>
          {/if}
        </div>

      {:else if activeTab === 'benefits'}
        <div class="info-section">
          <h3>Enrolled Benefits</h3>
          {#if benefits.length > 0}
            <div class="benefits-list">
              {#each benefits as benefit}
                <div class="benefit-card">
                  <div class="benefit-header">
                    <h4>{benefit.name}</h4>
                    <span class="benefit-status status-{benefit.status}">
                      {benefit.status}
                    </span>
                  </div>
                  <div class="benefit-details">
                    <div class="benefit-item">
                      <label>Type</label>
                      <div>{benefit.type}</div>
                    </div>
                    <div class="benefit-item">
                      <label>Provider</label>
                      <div>{benefit.provider}</div>
                    </div>
                    <div class="benefit-item">
                      <label>Enrolled</label>
                      <div>{new Date(benefit.enrolled_date).toLocaleDateString()}</div>
                    </div>
                    {#if benefit.employee_contribution}
                      <div class="benefit-item">
                        <label>Employee Contribution</label>
                        <div>${benefit.employee_contribution}/month</div>
                      </div>
                    {/if}
                    {#if benefit.employer_contribution}
                      <div class="benefit-item">
                        <label>Employer Contribution</label>
                        <div>${benefit.employer_contribution}/month</div>
                      </div>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="empty-state">
              No benefits enrolled
            </div>
          {/if}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .employee-detail {
    background: #ffffff; border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 24px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .loading, .error {
    text-align: center;
    padding: 48px;
    color: #333;
  }

  .error {
    color: #dc3545;
  }

  .employee-header {
    margin-bottom: 24px;
    padding-bottom: 24px;
    border-bottom: 2px solid #e9ecef;
  }

  .employee-name {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 8px;
  }

  .employee-name h2 {
    margin: 0;
    font-size: 24px;
    font-weight: 600;
  }

  .employee-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #333;
    font-size: 14px;
  }

  .status-badge {
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-onboarding { background: #fff3cd; color: #856404; }
  .status-active { background: #d4edda; color: #155724; }
  .status-leave { background: #d1ecf1; color: #0c5460; }
  .status-separated { background: #f8d7da; color: #721c24; }

  /* Tabs */
  .tabs {
    display: flex;
    gap: 8px;
    margin-bottom: 24px;
    border-bottom: 2px solid #e9ecef;
  }

  .tab {
    padding: 12px 24px;
    background: none;
    border: none;
    border-bottom: 3px solid transparent;
    color: #333;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .tab:hover {
    color: #333;
    background: #e9ecef;
  }

  .tab.active {
    color: #007bff;
    border-bottom-color: #007bff;
  }

  /* Tab Content */
  .tab-content {
    min-height: 400px;
  }

  .info-section h3 {
    margin: 0 0 16px 0;
    font-size: 18px;
    font-weight: 600;
    color: #333;
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 20px;
    margin-bottom: 32px;
  }

  .info-item label {
    display: block;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    color: #555;
    margin-bottom: 4px;
    letter-spacing: 0.5px;
  }

  .info-item div {
    font-size: 14px;
    color: #333;
  }

  .employment-type .badge {
    padding: 6px 12px;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 600;
  }

  .badge-W2-salary { background: #d4edda; color: #155724; }
  .badge-W2-hourly { background: #d1ecf1; color: #0c5460; }
  .badge-1099 { background: #fff3cd; color: #856404; }
  .badge-PRN { background: #e2e3e5; color: #383d41; }

  .pay-amount {
    font-size: 18px;
    font-weight: 600;
    color: #28a745;
  }

  /* Benefits */
  .benefits-list {
    display: grid;
    gap: 16px;
  }

  .benefit-card {
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 16px;
    background: #e9ecef;
  }

  .benefit-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .benefit-header h4 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  .benefit-status {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .benefit-status.status-active { background: #d4edda; color: #155724; }
  .benefit-status.status-pending { background: #fff3cd; color: #856404; }
  .benefit-status.status-terminated { background: #f8d7da; color: #721c24; }

  .benefit-details {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;
  }

  .benefit-item label {
    display: block;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    color: #555;
    margin-bottom: 2px;
  }

  .benefit-item div {
    font-size: 13px;
    color: #333;
  }

  .empty-state {
    text-align: center;
    padding: 48px;
    color: #555;
    font-style: italic;
  }
</style>