<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth-store';
  
  // Types
  interface EmployeeCompensation {
    id: string;
    employee_id: string;
    employment_type: 'W2' | '1099';
    pay_type: 'hourly' | 'salary' | 'commission';
    hourly_rate?: number;
    annual_salary?: number;
    pay_frequency: 'weekly' | 'biweekly' | 'semimonthly' | 'monthly';
    effective_date: string;
    overtime_eligible: boolean;
    standard_hours_per_week: number;
  }

  interface W2TaxWithholding {
    id: string;
    employee_id: string;
    filing_status: 'single' | 'married' | 'head_of_household';
    federal_allowances: number;
    state_allowances: number;
    additional_withholding: number;
    exempt_federal: boolean;
    exempt_state: boolean;
    exempt_fica: boolean;
  }

  interface PayrollPeriod {
    id: string;
    start_date: string;
    end_date: string;
    pay_date: string;
    status: 'draft' | 'processing' | 'processed' | 'failed';
    processed_by?: string;
    processed_at?: string;
    created_at: string;
  }

  interface PayStub {
    id: string;
    employee_id: string;
    payroll_period_id: string;
    gross_pay: number;
    federal_tax: number;
    state_tax: number;
    social_security: number;
    medicare: number;
    other_deductions: number;
    benefits_deductions: number;
    net_pay: number;
    hours_worked?: number;
    overtime_hours?: number;
    hourly_rate?: number;
    created_at: string;
  }

  interface PayStubDetail {
    pay_stub: PayStub;
    employee: any;
    payroll_period: PayrollPeriod;
    earnings: PayStubEarning[];
    deductions: PayStubDeduction[];
    taxes: PayStubTax[];
    ytd_gross_pay: number;
    ytd_net_pay: number;
    ytd_federal_tax: number;
    ytd_state_tax: number;
  }

  interface PayStubEarning {
    id: string;
    earning_type: string;
    description: string;
    hours?: number;
    rate?: number;
    amount: number;
  }

  interface PayStubDeduction {
    id: string;
    deduction_type: string;
    description: string;
    amount: number;
    pre_tax: boolean;
  }

  interface PayStubTax {
    id: string;
    tax_type: string;
    description: string;
    amount: number;
    taxable_wage: number;
    tax_rate?: number;
  }

  interface Form1099 {
    id: string;
    employee_id: string;
    tax_year: number;
    total_payments: number;
    federal_tax_withheld: number;
    state_tax_withheld: number;
    status: 'draft' | 'filed' | 'corrected';
    filed_date?: string;
  }

  // State
  let loading = false;
  let error = '';
  let success = '';
  let activeTab: 'paystubs' | 'compensation' | 'tax-withholding' | 'admin' = 'paystubs';
  let isAdmin = false;
  
  // Pay Stubs
  let payStubs: PayStub[] = [];
  let selectedPayStub: PayStubDetail | null = null;
  let showPayStubDetail = false;
  
  // Compensation
  let compensation: EmployeeCompensation | null = null;
  let showCompensationForm = false;
  let compensationForm = {
    employment_type: 'W2' as 'W2' | '1099',
    pay_type: 'hourly' as 'hourly' | 'salary',
    hourly_rate: 0,
    annual_salary: 0,
    pay_frequency: 'biweekly' as any,
    overtime_eligible: true,
    standard_hours_per_week: 40
  };
  
  // Tax Withholding
  let taxWithholding: W2TaxWithholding | null = null;
  let showTaxForm = false;
  let taxForm = {
    filing_status: 'single' as 'single' | 'married' | 'head_of_household',
    federal_allowances: 0,
    state_allowances: 0,
    additional_withholding: 0,
    exempt_federal: false,
    exempt_state: false,
    exempt_fica: false
  };
  
  // Admin - Payroll Periods
  let payrollPeriods: PayrollPeriod[] = [];
  let showPeriodForm = false;
  let periodForm = {
    start_date: '',
    end_date: '',
    pay_date: ''
  };
  let processingPeriod = false;
  
  // Admin - 1099 Forms
  let forms1099: Form1099[] = [];
  let selectedYear = new Date().getFullYear();

  // Computed
  $: totalTaxes = selectedPayStub ? (
    selectedPayStub.pay_stub.federal_tax +
    selectedPayStub.pay_stub.state_tax +
    selectedPayStub.pay_stub.social_security +
    selectedPayStub.pay_stub.medicare
  ) : 0;

  $: totalDeductions = selectedPayStub ? (
    selectedPayStub.pay_stub.other_deductions +
    selectedPayStub.pay_stub.benefits_deductions
  ) : 0;

  // API Calls - Pay Stubs
  async function loadPayStubs() {
    try {
      loading = true;
      error = '';
      
      const employeeId = $authStore.user?.employee_id;
      if (!employeeId) {
        error = 'Employee ID not found';
        return;
      }
      
      const response = await fetch(`/api/payroll/paystubs/employee/${employeeId}`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to load pay stubs');
      
      payStubs = await response.json();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function viewPayStubDetail(payStubId: string) {
    try {
      loading = true;
      error = '';
      
      const response = await fetch(`/api/payroll/paystubs/${payStubId}`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to load pay stub details');
      
      selectedPayStub = await response.json();
      showPayStubDetail = true;
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function downloadPayStubPDF(payStubId: string) {
    try {
      const response = await fetch(`/api/payroll/paystubs/${payStubId}/pdf`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to download PDF');
      
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `paystub_${payStubId}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err: any) {
      error = err.message;
    }
  }

  // API Calls - Compensation
  async function loadCompensation() {
    try {
      loading = true;
      error = '';
      
      const employeeId = $authStore.user?.employee_id;
      if (!employeeId) {
        error = 'Employee ID not found';
        return;
      }
      
      const response = await fetch(`/api/payroll/compensation/employee/${employeeId}`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) {
        // No compensation setup yet
        compensation = null;
        return;
      }
      
      compensation = await response.json();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function saveCompensation() {
    try {
      loading = true;
      error = '';
      success = '';
      
      const employeeId = $authStore.user?.employee_id;
      if (!employeeId) {
        error = 'Employee ID not found';
        return;
      }
      
      const payload = {
        employee_id: employeeId,
        ...compensationForm,
        effective_date: new Date().toISOString()
      };
      
      const response = await fetch('/api/payroll/compensation', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });
      
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to save compensation');
      }
      
      success = 'Compensation saved successfully';
      showCompensationForm = false;
      await loadCompensation();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  // API Calls - Tax Withholding
  async function loadTaxWithholding() {
    try {
      loading = true;
      error = '';
      
      const employeeId = $authStore.user?.employee_id;
      if (!employeeId) {
        error = 'Employee ID not found';
        return;
      }
      
      const response = await fetch(`/api/payroll/tax-withholding/${employeeId}`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) {
        taxWithholding = null;
        return;
      }
      
      taxWithholding = await response.json();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function saveTaxWithholding() {
    try {
      loading = true;
      error = '';
      success = '';
      
      const employeeId = $authStore.user?.employee_id;
      if (!employeeId) {
        error = 'Employee ID not found';
        return;
      }
      
      const response = await fetch(`/api/payroll/tax-withholding/${employeeId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(taxForm)
      });
      
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to save tax withholding');
      }
      
      success = 'Tax withholding saved successfully';
      showTaxForm = false;
      await loadTaxWithholding();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  // API Calls - Admin: Payroll Periods
  async function loadPayrollPeriods() {
    try {
      loading = true;
      error = '';
      
      const response = await fetch('/api/payroll/periods', {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to load payroll periods');
      
      payrollPeriods = await response.json();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function createPayrollPeriod() {
    try {
      loading = true;
      error = '';
      success = '';
      
      const response = await fetch('/api/payroll/periods', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(periodForm)
      });
      
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to create payroll period');
      }
      
      success = 'Payroll period created successfully';
      showPeriodForm = false;
      periodForm = { start_date: '', end_date: '', pay_date: '' };
      await loadPayrollPeriods();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function processPayroll(periodId: string) {
    if (!confirm('Are you sure you want to process this payroll period? This will create pay stubs for all employees.')) {
      return;
    }
    
    try {
      processingPeriod = true;
      error = '';
      success = '';
      
      const response = await fetch(`/api/payroll/periods/${periodId}/process`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ dry_run: false })
      });
      
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to process payroll');
      }
      
      const summary = await response.json();
      success = `Payroll processed successfully! ${summary.total_employees} employees processed.`;
      await loadPayrollPeriods();
    } catch (err: any) {
      error = err.message;
    } finally {
      processingPeriod = false;
    }
  }

  // API Calls - Admin: 1099 Forms
  async function generate1099Forms() {
    if (!confirm(`Generate 1099 forms for ${selectedYear}? This will create forms for all 1099 contractors.`)) {
      return;
    }
    
    try {
      loading = true;
      error = '';
      success = '';
      
      const response = await fetch(`/api/payroll/1099/generate/${selectedYear}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to generate 1099 forms');
      }
      
      const result = await response.json();
      success = result.message || `${result.count} 1099 forms generated`;
      await load1099Forms();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function load1099Forms() {
    try {
      loading = true;
      error = '';
      
      const response = await fetch(`/api/payroll/1099/${selectedYear}`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to load 1099 forms');
      
      const data = await response.json();
      forms1099 = data.forms || [];
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  // Utility Functions
  function formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount);
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return 'N/A';
    return new Date(dateStr).toLocaleDateString();
  }

  function formatPercentage(rate: number): string {
    return (rate * 100).toFixed(2) + '%';
  }

  function getStatusBadgeClass(status: string): string {
    const classes = {
      draft: 'badge-info',
      processing: 'badge-warning',
      processed: 'badge-success',
      failed: 'badge-error'
    };
    return classes[status] || 'badge-ghost';
  }

  function openCompensationForm() {
    if (compensation) {
      compensationForm = {
        employment_type: compensation.employment_type,
        pay_type: compensation.pay_type,
        hourly_rate: compensation.hourly_rate || 0,
        annual_salary: compensation.annual_salary || 0,
        pay_frequency: compensation.pay_frequency,
        overtime_eligible: compensation.overtime_eligible,
        standard_hours_per_week: compensation.standard_hours_per_week
      };
    }
    showCompensationForm = true;
  }

  function openTaxForm() {
    if (taxWithholding) {
      taxForm = {
        filing_status: taxWithholding.filing_status,
        federal_allowances: taxWithholding.federal_allowances,
        state_allowances: taxWithholding.state_allowances,
        additional_withholding: taxWithholding.additional_withholding,
        exempt_federal: taxWithholding.exempt_federal,
        exempt_state: taxWithholding.exempt_state,
        exempt_fica: taxWithholding.exempt_fica
      };
    }
    showTaxForm = true;
  }

  onMount(() => {
    // Check if user is admin
    isAdmin = $authStore.user?.role === 'admin';
    
    // Load initial data
    loadPayStubs();
    loadCompensation();
    
    if (compensation?.employment_type === 'W2') {
      loadTaxWithholding();
    }
    
    if (isAdmin) {
      loadPayrollPeriods();
    }
  });
</script>

<div class="payroll-container">
  <!-- Header -->
  <div class="payroll-header">
    <h1>💰 Payroll</h1>
    <p class="text-muted">View pay stubs and manage compensation</p>
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
      class="tab {activeTab === 'paystubs' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'paystubs'}>
      Pay Stubs
    </button>
    <button 
      class="tab {activeTab === 'compensation' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'compensation'}>
      Compensation
    </button>
    {#if compensation?.employment_type === 'W2'}
      <button 
        class="tab {activeTab === 'tax-withholding' ? 'tab-active' : ''}"
        on:click={() => activeTab = 'tax-withholding'}>
        Tax Withholding
      </button>
    {/if}
    {#if isAdmin}
      <button 
        class="tab {activeTab === 'admin' ? 'tab-active' : ''}"
        on:click={() => activeTab = 'admin'}>
        Admin
      </button>
    {/if}
  </div>

  <!-- Content -->
  <div class="tab-content">
    <!-- Pay Stubs Tab -->
    {#if activeTab === 'paystubs'}
      <div class="section">
        <div class="section-header">
          <h2>Your Pay Stubs</h2>
          <p class="text-muted">View and download your payment history</p>
        </div>

        {#if loading}
          <div class="loading">Loading...</div>
        {:else if payStubs.length === 0}
          <div class="empty-state">
            <span class="empty-icon">📄</span>
            <p>No pay stubs available yet</p>
          </div>
        {:else}
          <div class="paystubs-grid">
            {#each payStubs as stub}
              <div class="paystub-card">
                <div class="paystub-header">
                  <div class="paystub-icon">💵</div>
                  <div class="paystub-info">
                    <div class="paystub-amount">{formatCurrency(stub.net_pay)}</div>
                    <div class="paystub-date">{formatDate(stub.created_at)}</div>
                  </div>
                </div>
                
                <div class="paystub-details">
                  <div class="detail-row">
                    <span>Gross Pay</span>
                    <strong>{formatCurrency(stub.gross_pay)}</strong>
                  </div>
                  <div class="detail-row">
                    <span>Taxes</span>
                    <span class="text-error">
                      -{formatCurrency(stub.federal_tax + stub.state_tax + stub.social_security + stub.medicare)}
                    </span>
                  </div>
                  <div class="detail-row">
                    <span>Deductions</span>
                    <span class="text-error">-{formatCurrency(stub.other_deductions + stub.benefits_deductions)}</span>
                  </div>
                  {#if stub.hours_worked}
                    <div class="detail-row">
                      <span>Hours Worked</span>
                      <span>{stub.hours_worked.toFixed(2)} hrs</span>
                    </div>
                  {/if}
                </div>

                <div class="paystub-actions">
                  <button class="btn btn-sm btn-primary" on:click={() => viewPayStubDetail(stub.id)}>
                    View Details
                  </button>
                  <button class="btn btn-sm btn-ghost" on:click={() => downloadPayStubPDF(stub.id)}>
                    📥 Download
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    <!-- Compensation Tab -->
    {#if activeTab === 'compensation'}
      <div class="section">
        <div class="section-header">
          <h2>Compensation Information</h2>
          <button class="btn btn-primary" on:click={openCompensationForm}>
            {compensation ? 'Update' : 'Set Up'} Compensation
          </button>
        </div>

        {#if compensation}
          <div class="info-card">
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">Employment Type</span>
                <span class="info-value badge badge-{compensation.employment_type === 'W2' ? 'success' : 'info'}">
                  {compensation.employment_type}
                </span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Pay Type</span>
                <span class="info-value">{compensation.pay_type.charAt(0).toUpperCase() + compensation.pay_type.slice(1)}</span>
              </div>
              
              {#if compensation.hourly_rate}
                <div class="info-item">
                  <span class="info-label">Hourly Rate</span>
                  <span class="info-value">{formatCurrency(compensation.hourly_rate)}/hr</span>
                </div>
              {/if}
              
              {#if compensation.annual_salary}
                <div class="info-item">
                  <span class="info-label">Annual Salary</span>
                  <span class="info-value">{formatCurrency(compensation.annual_salary)}</span>
                </div>
              {/if}
              
              <div class="info-item">
                <span class="info-label">Pay Frequency</span>
                <span class="info-value">{compensation.pay_frequency.charAt(0).toUpperCase() + compensation.pay_frequency.slice(1)}</span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Overtime Eligible</span>
                <span class="info-value">{compensation.overtime_eligible ? 'Yes' : 'No'}</span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Standard Hours/Week</span>
                <span class="info-value">{compensation.standard_hours_per_week}</span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Effective Date</span>
                <span class="info-value">{formatDate(compensation.effective_date)}</span>
              </div>
            </div>
          </div>
        {:else}
          <div class="empty-state">
            <span class="empty-icon">💼</span>
            <p>No compensation information set up</p>
            <button class="btn btn-primary" on:click={() => showCompensationForm = true}>
              Set Up Compensation
            </button>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Tax Withholding Tab (W2 only) -->
    {#if activeTab === 'tax-withholding'}
      <div class="section">
        <div class="section-header">
          <h2>Tax Withholding (W-4)</h2>
          <button class="btn btn-primary" on:click={openTaxForm}>
            {taxWithholding ? 'Update' : 'Set Up'} Tax Withholding
          </button>
        </div>

        {#if taxWithholding}
          <div class="info-card">
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">Filing Status</span>
                <span class="info-value">
                  {taxWithholding.filing_status.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')}
                </span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Federal Allowances</span>
                <span class="info-value">{taxWithholding.federal_allowances}</span>
              </div>
              
              <div class="info-item">
                <span class="info-label">State Allowances</span>
                <span class="info-value">{taxWithholding.state_allowances}</span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Additional Withholding</span>
                <span class="info-value">{formatCurrency(taxWithholding.additional_withholding)}</span>
              </div>
              
              <div class="info-item">
                <span class="info-label">Federal Exempt</span>
                <span class="info-value badge badge-{taxWithholding.exempt_federal ? 'warning' : 'ghost'}">
                  {taxWithholding.exempt_federal ? 'Yes' : 'No'}
                </span>
              </div>
              
              <div class="info-item">
                <span class="info-label">State Exempt</span>
                <span class="info-value badge badge-{taxWithholding.exempt_state ? 'warning' : 'ghost'}">
                  {taxWithholding.exempt_state ? 'Yes' : 'No'}
                </span>
              </div>
              
              <div class="info-item">
                <span class="info-label">FICA Exempt</span>
                <span class="info-value badge badge-{taxWithholding.exempt_fica ? 'warning' : 'ghost'}">
                  {taxWithholding.exempt_fica ? 'Yes' : 'No'}
                </span>
              </div>
            </div>
          </div>
        {:else}
          <div class="empty-state">
            <span class="empty-icon">📝</span>
            <p>No tax withholding information set up</p>
            <button class="btn btn-primary" on:click={() => showTaxForm = true}>
              Set Up Tax Withholding
            </button>
          </div>
        {/if}
      </div>
    {/if}

    <!-- Admin Tab -->
    {#if activeTab === 'admin' && isAdmin}
      <div class="section">
        <!-- Payroll Periods -->
        <div class="admin-section">
          <div class="section-header">
            <h2>Payroll Periods</h2>
            <button class="btn btn-primary" on:click={() => showPeriodForm = true}>
              + Create Period
            </button>
          </div>

          <div class="periods-table">
            <table>
              <thead>
                <tr>
                  <th>Period</th>
                  <th>Pay Date</th>
                  <th>Status</th>
                  <th>Processed</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {#each payrollPeriods as period}
                  <tr>
                    <td>{formatDate(period.start_date)} - {formatDate(period.end_date)}</td>
                    <td>{formatDate(period.pay_date)}</td>
                    <td>
                      <span class="badge {getStatusBadgeClass(period.status)}">
                        {period.status}
                      </span>
                    </td>
                    <td>{period.processed_at ? formatDate(period.processed_at) : '-'}</td>
                    <td>
                      {#if period.status === 'draft'}
                        <button 
                          class="btn btn-sm btn-success" 
                          on:click={() => processPayroll(period.id)}
                          disabled={processingPeriod}>
                          {processingPeriod ? 'Processing...' : 'Process'}
                        </button>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>

        <!-- 1099 Forms -->
        <div class="admin-section">
          <div class="section-header">
            <h2>1099 Forms</h2>
            <div style="display: flex; gap: 1rem; align-items: center;">
              <select bind:value={selectedYear} class="select select-sm">
                {#each Array(5).fill(0).map((_, i) => new Date().getFullYear() - i) as year}
                  <option value={year}>{year}</option>
                {/each}
              </select>
              <button class="btn btn-primary btn-sm" on:click={generate1099Forms}>
                Generate {selectedYear} Forms
              </button>
            </div>
          </div>

          <div class="forms-grid">
            {#each forms1099 as form}
              <div class="form-card">
                <div class="form-header">
                  <span class="form-year">1099-NEC {form.tax_year}</span>
                  <span class="badge {getStatusBadgeClass(form.status)}">{form.status}</span>
                </div>
                <div class="form-amount">
                  {formatCurrency(form.total_payments)}
                </div>
                <div class="form-footer">
                  {#if form.filed_date}
                    <span class="text-muted">Filed {formatDate(form.filed_date)}</span>
                  {:else}
                    <span class="text-muted">Not filed</span>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

<!-- Pay Stub Detail Modal -->
{#if showPayStubDetail && selectedPayStub}
  <div class="modal" on:click={() => showPayStubDetail = false}>
    <div class="modal-box max-w-4xl" on:click|stopPropagation>
      <div class="modal-header">
        <h2>Pay Stub Details</h2>
        <button class="btn btn-circle btn-sm" on:click={() => showPayStubDetail = false}>✕</button>
      </div>

      <div class="paystub-detail">
        <!-- Summary -->
        <div class="detail-summary">
          <div class="summary-item">
            <span class="summary-label">Gross Pay</span>
            <span class="summary-value">{formatCurrency(selectedPayStub.pay_stub.gross_pay)}</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">Total Taxes</span>
            <span class="summary-value text-error">-{formatCurrency(totalTaxes)}</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">Total Deductions</span>
            <span class="summary-value text-error">-{formatCurrency(totalDeductions)}</span>
          </div>
          <div class="summary-item highlight">
            <span class="summary-label">Net Pay</span>
            <span class="summary-value">{formatCurrency(selectedPayStub.pay_stub.net_pay)}</span>
          </div>
        </div>

        <!-- Earnings -->
        {#if selectedPayStub.earnings.length > 0}
          <div class="detail-section">
            <h3>Earnings</h3>
            <table class="detail-table">
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Description</th>
                  <th>Hours</th>
                  <th>Rate</th>
                  <th>Amount</th>
                </tr>
              </thead>
              <tbody>
                {#each selectedPayStub.earnings as earning}
                  <tr>
                    <td>{earning.earning_type}</td>
                    <td>{earning.description}</td>
                    <td>{earning.hours ? earning.hours.toFixed(2) : '-'}</td>
                    <td>{earning.rate ? formatCurrency(earning.rate) : '-'}</td>
                    <td>{formatCurrency(earning.amount)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}

        <!-- Taxes -->
        {#if selectedPayStub.taxes.length > 0}
          <div class="detail-section">
            <h3>Taxes</h3>
            <table class="detail-table">
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Description</th>
                  <th>Taxable Wage</th>
                  <th>Rate</th>
                  <th>Amount</th>
                </tr>
              </thead>
              <tbody>
                {#each selectedPayStub.taxes as tax}
                  <tr>
                    <td>{tax.tax_type}</td>
                    <td>{tax.description}</td>
                    <td>{formatCurrency(tax.taxable_wage)}</td>
                    <td>{tax.tax_rate ? formatPercentage(tax.tax_rate) : '-'}</td>
                    <td>{formatCurrency(tax.amount)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}

        <!-- Deductions -->
        {#if selectedPayStub.deductions.length > 0}
          <div class="detail-section">
            <h3>Deductions</h3>
            <table class="detail-table">
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Description</th>
                  <th>Pre-Tax</th>
                  <th>Amount</th>
                </tr>
              </thead>
              <tbody>
                {#each selectedPayStub.deductions as deduction}
                  <tr>
                    <td>{deduction.deduction_type}</td>
                    <td>{deduction.description}</td>
                    <td>{deduction.pre_tax ? 'Yes' : 'No'}</td>
                    <td>{formatCurrency(deduction.amount)}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}

        <!-- YTD Summary -->
        <div class="detail-section">
          <h3>Year-to-Date Summary</h3>
          <div class="ytd-grid">
            <div class="ytd-item">
              <span class="ytd-label">YTD Gross Pay</span>
              <span class="ytd-value">{formatCurrency(selectedPayStub.ytd_gross_pay)}</span>
            </div>
            <div class="ytd-item">
              <span class="ytd-label">YTD Net Pay</span>
              <span class="ytd-value">{formatCurrency(selectedPayStub.ytd_net_pay)}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button class="btn btn-primary" on:click={() => downloadPayStubPDF(selectedPayStub.pay_stub.id)}>
          📥 Download PDF
        </button>
        <button class="btn btn-ghost" on:click={() => showPayStubDetail = false}>Close</button>
      </div>
    </div>
  </div>
{/if}

<!-- Compensation Form Modal -->
{#if showCompensationForm}
  <div class="modal" on:click={() => showCompensationForm = false}>
    <div class="modal-box max-w-2xl" on:click|stopPropagation>
      <h2>Compensation Setup</h2>
      
      <form on:submit|preventDefault={saveCompensation} class="form">
        <div class="form-control">
          <label class="label">
            <span class="label-text">Employment Type</span>
          </label>
          <select bind:value={compensationForm.employment_type} class="select w-full" required>
            <option value="W2">W2 Employee</option>
            <option value="1099">1099 Contractor</option>
          </select>
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Pay Type</span>
          </label>
          <select bind:value={compensationForm.pay_type} class="select w-full" required>
            <option value="hourly">Hourly</option>
            <option value="salary">Salary</option>
            <option value="commission">Commission</option>
          </select>
        </div>

        {#if compensationForm.pay_type === 'hourly'}
          <div class="form-control">
            <label class="label">
              <span class="label-text">Hourly Rate</span>
            </label>
            <input 
              type="number" 
              bind:value={compensationForm.hourly_rate} 
              class="input w-full" 
              min="0" 
              step="0.01" 
              required 
            />
          </div>
        {/if}

        {#if compensationForm.pay_type === 'salary'}
          <div class="form-control">
            <label class="label">
              <span class="label-text">Annual Salary</span>
            </label>
            <input 
              type="number" 
              bind:value={compensationForm.annual_salary} 
              class="input w-full" 
              min="0" 
              step="1000" 
              required 
            />
          </div>
        {/if}

        <div class="form-control">
          <label class="label">
            <span class="label-text">Pay Frequency</span>
          </label>
          <select bind:value={compensationForm.pay_frequency} class="select w-full" required>
            <option value="weekly">Weekly</option>
            <option value="biweekly">Bi-Weekly</option>
            <option value="semimonthly">Semi-Monthly</option>
            <option value="monthly">Monthly</option>
          </select>
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Standard Hours Per Week</span>
          </label>
          <input 
            type="number" 
            bind:value={compensationForm.standard_hours_per_week} 
            class="input w-full" 
            min="0" 
            max="80" 
            required 
          />
        </div>

        <div class="form-control">
          <label class="label cursor-pointer">
            <span class="label-text">Overtime Eligible</span>
            <input 
              type="checkbox" 
              bind:checked={compensationForm.overtime_eligible} 
              class="checkbox" 
            />
          </label>
        </div>

        <div class="modal-actions">
          <button type="submit" class="btn btn-primary" disabled={loading}>
            {loading ? 'Saving...' : 'Save'}
          </button>
          <button type="button" class="btn btn-ghost" on:click={() => showCompensationForm = false}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Tax Withholding Form Modal -->
{#if showTaxForm}
  <div class="modal" on:click={() => showTaxForm = false}>
    <div class="modal-box max-w-2xl" on:click|stopPropagation>
      <h2>Tax Withholding (W-4)</h2>
      
      <form on:submit|preventDefault={saveTaxWithholding} class="form">
        <div class="form-control">
          <label class="label">
            <span class="label-text">Filing Status</span>
          </label>
          <select bind:value={taxForm.filing_status} class="select w-full" required>
            <option value="single">Single</option>
            <option value="married">Married Filing Jointly</option>
            <option value="head_of_household">Head of Household</option>
          </select>
        </div>

        <div class="form-row">
          <div class="form-control">
            <label class="label">
              <span class="label-text">Federal Allowances</span>
            </label>
            <input 
              type="number" 
              bind:value={taxForm.federal_allowances} 
              class="input w-full" 
              min="0" 
              required 
            />
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">State Allowances</span>
            </label>
            <input 
              type="number" 
              bind:value={taxForm.state_allowances} 
              class="input w-full" 
              min="0" 
              required 
            />
          </div>
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Additional Withholding (per paycheck)</span>
          </label>
          <input 
            type="number" 
            bind:value={taxForm.additional_withholding} 
            class="input w-full" 
            min="0" 
            step="0.01" 
          />
        </div>

        <div class="form-control">
          <label class="label cursor-pointer">
            <span class="label-text">Exempt from Federal Tax</span>
            <input 
              type="checkbox" 
              bind:checked={taxForm.exempt_federal} 
              class="checkbox" 
            />
          </label>
        </div>

        <div class="form-control">
          <label class="label cursor-pointer">
            <span class="label-text">Exempt from State Tax</span>
            <input 
              type="checkbox" 
              bind:checked={taxForm.exempt_state} 
              class="checkbox" 
            />
          </label>
        </div>

        <div class="form-control">
          <label class="label cursor-pointer">
            <span class="label-text">Exempt from FICA (Social Security & Medicare)</span>
            <input 
              type="checkbox" 
              bind:checked={taxForm.exempt_fica} 
              class="checkbox" 
            />
          </label>
        </div>

        <div class="modal-actions">
          <button type="submit" class="btn btn-primary" disabled={loading}>
            {loading ? 'Saving...' : 'Save'}
          </button>
          <button type="button" class="btn btn-ghost" on:click={() => showTaxForm = false}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Period Form Modal -->
{#if showPeriodForm}
  <div class="modal" on:click={() => showPeriodForm = false}>
    <div class="modal-box" on:click|stopPropagation>
      <h2>Create Payroll Period</h2>
      
      <form on:submit|preventDefault={createPayrollPeriod} class="form">
        <div class="form-control">
          <label class="label">
            <span class="label-text">Start Date</span>
          </label>
          <input 
            type="date" 
            bind:value={periodForm.start_date} 
            class="input w-full" 
            required 
          />
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">End Date</span>
          </label>
          <input 
            type="date" 
            bind:value={periodForm.end_date} 
            class="input w-full" 
            required 
          />
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Pay Date</span>
          </label>
          <input 
            type="date" 
            bind:value={periodForm.pay_date} 
            class="input w-full" 
            required 
          />
        </div>

        <div class="modal-actions">
          <button type="submit" class="btn btn-primary" disabled={loading}>
            {loading ? 'Creating...' : 'Create'}
          </button>
          <button type="button" class="btn btn-ghost" on:click={() => showPeriodForm = false}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<style>
  .payroll-container {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .payroll-header {
    margin-bottom: 2rem;
  }

  .payroll-header h1 {
    font-size: 2rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 0.5rem;
  }

  .text-muted {
    color: #6b7280;
    font-size: 0.875rem;
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
    padding: 0;
    line-height: 1;
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

  .section {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 2rem;
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

  .paystubs-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }

  .paystub-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
    transition: all 0.2s;
  }

  .paystub-card:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  .paystub-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1.5rem;
    padding-bottom: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .paystub-icon {
    font-size: 2.5rem;
  }

  .paystub-info {
    flex: 1;
  }

  .paystub-amount {
    font-size: 1.75rem;
    font-weight: 700;
    color: #059669;
  }

  .paystub-date {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .paystub-details {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 1.5rem;
  }

  .detail-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .detail-row span:first-child {
    color: #6b7280;
  }

  .detail-row strong {
    color: #111827;
    font-weight: 600;
  }

  .text-error {
    color: #dc2626;
  }

  .paystub-actions {
    display: flex;
    gap: 0.75rem;
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
    border-color: #3b82f6;
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

  .btn-success {
    background: #059669;
    color: white;
  }

  .btn-success:hover:not(:disabled) {
    background: #047857;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .info-card {
    background: #f9fafb;
    border-radius: 0.75rem;
    padding: 1.5rem;
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .info-label {
    color: #6b7280;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .info-value {
    color: #111827;
    font-weight: 600;
    font-size: 1rem;
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

  .admin-section {
    margin-bottom: 3rem;
  }

  .admin-section:last-child {
    margin-bottom: 0;
  }

  .periods-table {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    overflow: hidden;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th {
    background: #f9fafb;
    padding: 0.75rem 1rem;
    text-align: left;
    font-weight: 600;
    color: #374151;
    border-bottom: 1px solid #e5e7eb;
    font-size: 0.875rem;
  }

  td {
    padding: 1rem;
    border-bottom: 1px solid #e5e7eb;
  }

  tbody tr:last-child td {
    border-bottom: none;
  }

  tbody tr:hover {
    background: #f9fafb;
  }

  .forms-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1rem;
  }

  .form-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
  }

  .form-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .form-year {
    font-weight: 600;
    color: #111827;
  }

  .form-amount {
    font-size: 1.5rem;
    font-weight: 700;
    color: #059669;
    margin-bottom: 0.5rem;
  }

  .form-footer {
    font-size: 0.875rem;
    color: #6b7280;
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
  }

  .max-w-2xl {
    max-width: 42rem;
  }

  .max-w-4xl {
    max-width: 56rem;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .modal-header h2 {
    font-size: 1.5rem;
    font-weight: 700;
    color: #111827;
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

  .paystub-detail {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .detail-summary {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1rem;
    background: #f9fafb;
    padding: 1.5rem;
    border-radius: 0.75rem;
  }

  .summary-item {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .summary-item.highlight {
    background: #3b82f6;
    color: white;
    padding: 1rem;
    border-radius: 0.5rem;
  }

  .summary-label {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .summary-item.highlight .summary-label {
    color: rgba(255, 255, 255, 0.9);
  }

  .summary-value {
    font-size: 1.25rem;
    font-weight: 700;
    color: #111827;
  }

  .summary-item.highlight .summary-value {
    color: white;
  }

  .detail-section {
    border-top: 1px solid #e5e7eb;
    padding-top: 1.5rem;
  }

  .detail-section h3 {
    font-size: 1.125rem;
    font-weight: 600;
    color: #111827;
    margin-bottom: 1rem;
  }

  .detail-table {
    font-size: 0.875rem;
  }

  .detail-table th {
    background: #f9fafb;
  }

  .ytd-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  .ytd-item {
    background: #f0fdf4;
    padding: 1rem;
    border-radius: 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .ytd-label {
    font-size: 0.875rem;
    color: #166534;
  }

  .ytd-value {
    font-size: 1.25rem;
    font-weight: 700;
    color: #059669;
  }

  .modal-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
    margin-top: 1.5rem;
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

  .label.cursor-pointer {
    cursor: pointer;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }

  .select,
  .input {
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 1rem;
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

  .checkbox {
    width: 1.25rem;
    height: 1.25rem;
    border-radius: 0.25rem;
    border: 1px solid #d1d5db;
    cursor: pointer;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .select-sm {
    padding: 0.375rem;
    font-size: 0.875rem;
  }

  @media (max-width: 768px) {
    .payroll-container {
      padding: 1rem;
    }

    .paystubs-grid {
      grid-template-columns: 1fr;
    }

    .info-grid {
      grid-template-columns: 1fr;
    }

    .detail-summary {
      grid-template-columns: 1fr;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .modal-box {
      padding: 1.5rem;
    }
  }
</style>