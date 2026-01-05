<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../../stores/auth';
  import { getApiBaseUrl } from '../../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();
  
  let activeTab: 'payroll' | 'expenses' = 'payroll';
  
  // Payroll State
  interface PayrollRun {
    id: string;
    period_start: string;
    period_end: string;
    pay_date: string;
    status: 'draft' | 'processing' | 'completed' | 'failed';
    total_gross: number;
    total_net: number;
    total_taxes: number;
    employee_count: number;
    created_at: string;
  }
  
  interface PayrollEmployee {
    id: string;
    employee_id: string;
    employee_name: string;
    department: string;
    gross_pay: number;
    deductions: number;
    taxes: number;
    net_pay: number;
    hours_worked?: number;
    hourly_rate?: number;
  }
  
  let payrollRuns: PayrollRun[] = [];
  let selectedPayrollRun: PayrollRun | null = null;
  let payrollEmployees: PayrollEmployee[] = [];
  let showPayrollDetail = false;
  let payrollLoading = true;
  
  // Expenses State
  interface Expense {
    id: string;
    employee_id: string;
    employee_name: string;
    category: string;
    amount: number;
    currency: string;
    description: string;
    receipt_url?: string;
    submitted_date: string;
    status: 'pending' | 'approved' | 'rejected' | 'paid';
    approver_name?: string;
    approval_date?: string;
    notes?: string;
  }
  
  let expenses: Expense[] = [];
  let selectedExpense: Expense | null = null;
  let showExpenseDetail = false;
  let expenseFilter = 'pending';
  let expensesLoading = true;
  
  // Stats
  let payrollStats = {
    nextPayDate: '',
    totalMonthly: 0,
    pendingApprovals: 0,
    lastRunDate: ''
  };
  
  let expenseStats = {
    pending: 0,
    pendingAmount: 0,
    approvedThisMonth: 0,
    approvedAmount: 0
  };
  
  onMount(async () => {
    await loadPayroll();
    await loadExpenses();
  });
  
  // Payroll Functions
  async function loadPayroll() {
    try {
      payrollLoading = true;
      const token = $authStore.token || localStorage.getItem('token');
      
      const response = await fetch(`${API_BASE_URL}/payroll/runs`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        payrollRuns = await response.json();
        calculatePayrollStats();
      }
    } catch (err) {
      console.error('Failed to load payroll:', err);
    } finally {
      payrollLoading = false;
    }
  }
  
  function calculatePayrollStats() {
    const now = new Date();
    const upcoming = payrollRuns.find(r => new Date(r.pay_date) > now && r.status !== 'completed');
    const recent = payrollRuns.filter(r => r.status === 'completed').sort((a, b) => 
      new Date(b.pay_date).getTime() - new Date(a.pay_date).getTime()
    )[0];
    
    payrollStats.nextPayDate = upcoming?.pay_date || 'Not scheduled';
    payrollStats.lastRunDate = recent?.pay_date || 'N/A';
    payrollStats.totalMonthly = payrollRuns
      .filter(r => {
        const date = new Date(r.pay_date);
        return date.getMonth() === now.getMonth() && date.getFullYear() === now.getFullYear();
      })
      .reduce((sum, r) => sum + r.total_net, 0);
    payrollStats.pendingApprovals = payrollRuns.filter(r => r.status === 'draft').length;
  }
  
  async function viewPayrollRun(run: PayrollRun) {
    selectedPayrollRun = run;
    try {
      const token = $authStore.token || localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/payroll/runs/${run.id}/employees`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        payrollEmployees = await response.json();
        showPayrollDetail = true;
      }
    } catch (err) {
      console.error('Failed to load payroll details:', err);
    }
  }
  
  function closePayrollDetail() {
    showPayrollDetail = false;
    selectedPayrollRun = null;
    payrollEmployees = [];
  }
  
  async function processPayroll(runId: string) {
    if (!confirm('Process this payroll run? This will initiate payments to all employees.')) return;
    
    try {
      const token = $authStore.token || localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/payroll/runs/${runId}/process`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        await loadPayroll();
        alert('Payroll processing initiated successfully!');
      }
    } catch (err) {
      console.error('Failed to process payroll:', err);
      alert('Failed to process payroll: ' + err.message);
    }
  }
  
  // Expense Functions
  async function loadExpenses() {
    try {
      expensesLoading = true;
      const token = $authStore.token || localStorage.getItem('token');
      
      const response = await fetch(`${API_BASE_URL}/expenses`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (response.ok) {
        expenses = await response.json();
        calculateExpenseStats();
      }
    } catch (err) {
      console.error('Failed to load expenses:', err);
    } finally {
      expensesLoading = false;
    }
  }
  
  function calculateExpenseStats() {
    const now = new Date();
    const thisMonth = expenses.filter(e => {
      const date = new Date(e.submitted_date);
      return date.getMonth() === now.getMonth() && date.getFullYear() === now.getFullYear();
    });
    
    expenseStats.pending = expenses.filter(e => e.status === 'pending').length;
    expenseStats.pendingAmount = expenses
      .filter(e => e.status === 'pending')
      .reduce((sum, e) => sum + e.amount, 0);
    expenseStats.approvedThisMonth = thisMonth.filter(e => e.status === 'approved' || e.status === 'paid').length;
    expenseStats.approvedAmount = thisMonth
      .filter(e => e.status === 'approved' || e.status === 'paid')
      .reduce((sum, e) => sum + e.amount, 0);
  }
  
  function viewExpense(expense: Expense) {
    selectedExpense = expense;
    showExpenseDetail = true;
  }
  
  function closeExpenseDetail() {
    showExpenseDetail = false;
    selectedExpense = null;
  }
  
  async function updateExpenseStatus(expenseId: string, status: 'approved' | 'rejected', notes?: string) {
    try {
      const token = $authStore.token || localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/expenses/${expenseId}/status`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status, notes })
      });
      
      if (response.ok) {
        await loadExpenses();
        closeExpenseDetail();
      }
    } catch (err) {
      console.error('Failed to update expense:', err);
      alert('Failed to update expense: ' + err.message);
    }
  }
  
  function getStatusColor(status: string): string {
    const colors = {
      'draft': 'bg-gray-100 text-gray-800',
      'pending': 'bg-yellow-100 text-yellow-800',
      'processing': 'bg-blue-100 text-blue-800',
      'approved': 'bg-green-100 text-green-800',
      'completed': 'bg-green-100 text-green-800',
      'paid': 'bg-green-100 text-green-800',
      'rejected': 'bg-red-100 text-red-800',
      'failed': 'bg-red-100 text-red-800'
    };
    return colors[status] || 'bg-gray-100 text-gray-800';
  }
  
  $: filteredExpenses = expenses.filter(e => 
    expenseFilter === 'all' || e.status === expenseFilter
  );
</script>

<div class="pay-expenses-container">
  <!-- Header -->
  <div class="header">
    <div>
      <h1>💳 Pay & Expenses</h1>
      <p class="subtitle">Manage payroll processing and expense approvals</p>
    </div>
  </div>
  
  <!-- Tab Navigation -->
  <div class="tabs">
    <button 
      class="tab"
      class:active={activeTab === 'payroll'}
      on:click={() => activeTab = 'payroll'}
    >
      <span class="tab-icon">💵</span>
      <span class="tab-text">Payroll</span>
      {#if payrollStats.pendingApprovals > 0}
        <span class="tab-badge">{payrollStats.pendingApprovals}</span>
      {/if}
    </button>
    
    <button 
      class="tab"
      class:active={activeTab === 'expenses'}
      on:click={() => activeTab = 'expenses'}
    >
      <span class="tab-icon">🧾</span>
      <span class="tab-text">Expenses</span>
      {#if expenseStats.pending > 0}
        <span class="tab-badge">{expenseStats.pending}</span>
      {/if}
    </button>
  </div>
  
  <!-- Tab Content -->
  {#if activeTab === 'payroll'}
    <!-- Payroll Tab -->
    <div class="tab-content">
      <!-- Payroll Stats -->
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-icon">📅</div>
          <div class="stat-content">
            <div class="stat-label">Next Pay Date</div>
            <div class="stat-value">
              {payrollStats.nextPayDate === 'Not scheduled' 
                ? payrollStats.nextPayDate 
                : new Date(payrollStats.nextPayDate).toLocaleDateString()}
            </div>
          </div>
        </div>
        
        <div class="stat-card">
          <div class="stat-icon">💰</div>
          <div class="stat-content">
            <div class="stat-label">Monthly Payroll</div>
            <div class="stat-value">${payrollStats.totalMonthly.toLocaleString()}</div>
          </div>
        </div>
        
        <div class="stat-card highlight">
          <div class="stat-icon">⏳</div>
          <div class="stat-content">
            <div class="stat-label">Pending Approval</div>
            <div class="stat-value">{payrollStats.pendingApprovals}</div>
          </div>
        </div>
        
        <div class="stat-card">
          <div class="stat-icon">✓</div>
          <div class="stat-content">
            <div class="stat-label">Last Run</div>
            <div class="stat-value">
              {payrollStats.lastRunDate === 'N/A' 
                ? payrollStats.lastRunDate 
                : new Date(payrollStats.lastRunDate).toLocaleDateString()}
            </div>
          </div>
        </div>
      </div>
      
      <!-- Payroll Runs List -->
      <div class="section-card">
        <div class="section-header">
          <h2>Payroll Runs</h2>
          <button class="btn-primary">+ New Payroll Run</button>
        </div>
        
        {#if payrollLoading}
          <div class="loading">Loading payroll runs...</div>
        {:else if payrollRuns.length === 0}
          <div class="empty-state">
            <span class="empty-icon">💵</span>
            <p>No payroll runs yet</p>
          </div>
        {:else}
          <div class="payroll-list">
            {#each payrollRuns as run}
              <div class="payroll-card" on:click={() => viewPayrollRun(run)}>
                <div class="payroll-header">
                  <div class="payroll-period">
                    <h3>Pay Period</h3>
                    <p>{new Date(run.period_start).toLocaleDateString()} - {new Date(run.period_end).toLocaleDateString()}</p>
                  </div>
                  <span class="status-badge {getStatusColor(run.status)}">
                    {run.status}
                  </span>
                </div>
                
                <div class="payroll-body">
                  <div class="payroll-stat">
                    <span class="stat-label">Pay Date</span>
                    <span class="stat-value">{new Date(run.pay_date).toLocaleDateString()}</span>
                  </div>
                  <div class="payroll-stat">
                    <span class="stat-label">Employees</span>
                    <span class="stat-value">{run.employee_count}</span>
                  </div>
                  <div class="payroll-stat">
                    <span class="stat-label">Gross Pay</span>
                    <span class="stat-value">${run.total_gross.toLocaleString()}</span>
                  </div>
                  <div class="payroll-stat">
                    <span class="stat-label">Net Pay</span>
                    <span class="stat-value highlight">${run.total_net.toLocaleString()}</span>
                  </div>
                </div>
                
                {#if run.status === 'draft'}
                  <div class="payroll-actions">
                    <button 
                      class="btn-small primary"
                      on:click|stopPropagation={() => processPayroll(run.id)}
                    >
                      Process Payroll
                    </button>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    
  {:else}
    <!-- Expenses Tab -->
    <div class="tab-content">
      <!-- Expense Stats -->
      <div class="stats-grid">
        <div class="stat-card highlight">
          <div class="stat-icon">⏳</div>
          <div class="stat-content">
            <div class="stat-label">Pending Approval</div>
            <div class="stat-value">{expenseStats.pending}</div>
            <div class="stat-sublabel">${expenseStats.pendingAmount.toLocaleString()}</div>
          </div>
        </div>
        
        <div class="stat-card">
          <div class="stat-icon">✓</div>
          <div class="stat-content">
            <div class="stat-label">Approved This Month</div>
            <div class="stat-value">{expenseStats.approvedThisMonth}</div>
            <div class="stat-sublabel">${expenseStats.approvedAmount.toLocaleString()}</div>
          </div>
        </div>
        
        <div class="stat-card">
          <div class="stat-icon">📊</div>
          <div class="stat-content">
            <div class="stat-label">Total Expenses</div>
            <div class="stat-value">{expenses.length}</div>
          </div>
        </div>
      </div>
      
      <!-- Expense Filters -->
      <div class="filter-bar">
        <button 
          class="filter-btn"
          class:active={expenseFilter === 'all'}
          on:click={() => expenseFilter = 'all'}
        >
          All ({expenses.length})
        </button>
        <button 
          class="filter-btn"
          class:active={expenseFilter === 'pending'}
          on:click={() => expenseFilter = 'pending'}
        >
          Pending ({expenses.filter(e => e.status === 'pending').length})
        </button>
        <button 
          class="filter-btn"
          class:active={expenseFilter === 'approved'}
          on:click={() => expenseFilter = 'approved'}
        >
          Approved ({expenses.filter(e => e.status === 'approved').length})
        </button>
        <button 
          class="filter-btn"
          class:active={expenseFilter === 'rejected'}
          on:click={() => expenseFilter = 'rejected'}
        >
          Rejected ({expenses.filter(e => e.status === 'rejected').length})
        </button>
      </div>
      
      <!-- Expenses List -->
      <div class="section-card">
        {#if expensesLoading}
          <div class="loading">Loading expenses...</div>
        {:else if filteredExpenses.length === 0}
          <div class="empty-state">
            <span class="empty-icon">🧾</span>
            <p>No {expenseFilter === 'all' ? '' : expenseFilter} expenses found</p>
          </div>
        {:else}
          <div class="expense-list">
            {#each filteredExpenses as expense}
              <div class="expense-card" on:click={() => viewExpense(expense)}>
                <div class="expense-header">
                  <div class="expense-info">
                    <h3 class="expense-employee">{expense.employee_name}</h3>
                    <p class="expense-description">{expense.description}</p>
                  </div>
                  <div class="expense-amount">${expense.amount.toLocaleString()}</div>
                </div>
                
                <div class="expense-meta">
                  <span class="expense-category">
                    <span class="meta-icon">🏷️</span>
                    {expense.category}
                  </span>
                  <span class="expense-date">
                    <span class="meta-icon">📅</span>
                    {new Date(expense.submitted_date).toLocaleDateString()}
                  </span>
                  <span class="status-badge {getStatusColor(expense.status)}">
                    {expense.status}
                  </span>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<!-- Payroll Detail Modal -->
{#if showPayrollDetail && selectedPayrollRun}
  <div class="modal-overlay" on:click={closePayrollDetail}>
    <div class="modal large" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>Payroll Run Details</h2>
          <p class="modal-subtitle">
            Pay Period: {new Date(selectedPayrollRun.period_start).toLocaleDateString()} - 
            {new Date(selectedPayrollRun.period_end).toLocaleDateString()}
          </p>
        </div>
        <button class="close-btn" on:click={closePayrollDetail}>×</button>
      </div>
      
      <div class="modal-body">
        <!-- Payroll Summary -->
        <div class="payroll-summary">
          <div class="summary-item">
            <span class="summary-label">Total Gross</span>
            <span class="summary-value">${selectedPayrollRun.total_gross.toLocaleString()}</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">Total Taxes</span>
            <span class="summary-value">${selectedPayrollRun.total_taxes.toLocaleString()}</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">Total Net</span>
            <span class="summary-value highlight">${selectedPayrollRun.total_net.toLocaleString()}</span>
          </div>
        </div>
        
        <!-- Employee Payroll Table -->
        <div class="employee-payroll-table">
          <table>
            <thead>
              <tr>
                <th>Employee</th>
                <th>Department</th>
                <th>Hours</th>
                <th>Gross Pay</th>
                <th>Taxes</th>
                <th>Deductions</th>
                <th>Net Pay</th>
              </tr>
            </thead>
            <tbody>
              {#each payrollEmployees as employee}
                <tr>
                  <td>{employee.employee_name}</td>
                  <td>{employee.department}</td>
                  <td>{employee.hours_worked || '-'}</td>
                  <td>${employee.gross_pay.toLocaleString()}</td>
                  <td>${employee.taxes.toLocaleString()}</td>
                  <td>${employee.deductions.toLocaleString()}</td>
                  <td class="net-pay">${employee.net_pay.toLocaleString()}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Expense Detail Modal -->
{#if showExpenseDetail && selectedExpense}
  <div class="modal-overlay" on:click={closeExpenseDetail}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>Expense Details</h2>
          <p class="modal-subtitle">{selectedExpense.employee_name}</p>
        </div>
        <button class="close-btn" on:click={closeExpenseDetail}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="expense-detail">
          <div class="detail-row">
            <span class="detail-label">Category</span>
            <span class="detail-value">{selectedExpense.category}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Amount</span>
            <span class="detail-value">${selectedExpense.amount.toLocaleString()} {selectedExpense.currency}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Submitted</span>
            <span class="detail-value">{new Date(selectedExpense.submitted_date).toLocaleDateString()}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Status</span>
            <span class="status-badge {getStatusColor(selectedExpense.status)}">
              {selectedExpense.status}
            </span>
          </div>
          <div class="detail-row full">
            <span class="detail-label">Description</span>
            <p class="detail-description">{selectedExpense.description}</p>
          </div>
          {#if selectedExpense.notes}
            <div class="detail-row full">
              <span class="detail-label">Notes</span>
              <p class="detail-description">{selectedExpense.notes}</p>
            </div>
          {/if}
          {#if selectedExpense.receipt_url}
            <div class="detail-row full">
              <button class="btn-link" on:click={() => window.open(selectedExpense.receipt_url, '_blank')}>
                📎 View Receipt
              </button>
            </div>
          {/if}
        </div>
        
        {#if selectedExpense.status === 'pending'}
          <div class="expense-actions">
            <button 
              class="btn-primary"
              on:click={() => updateExpenseStatus(selectedExpense.id, 'approved')}
            >
              ✓ Approve
            </button>
            <button 
              class="btn-danger"
              on:click={() => {
                const notes = prompt('Reason for rejection (optional):');
                updateExpenseStatus(selectedExpense.id, 'rejected', notes || undefined);
              }}
            >
              ✗ Reject
            </button>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .pay-expenses-container {
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
    padding: 12px 24px;
    background: none;
    border: none;
    border-radius: 8px;
    font-size: 15px;
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
    background: #3b82f6;
    color: white;
  }
  
  .tab-icon {
    font-size: 20px;
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
  }
  
  .stat-card.highlight {
    background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
    border-left: 4px solid #f59e0b;
  }
  
  .stat-icon {
    font-size: 32px;
  }
  
  .stat-content {
    flex: 1;
  }
  
  .stat-label {
    font-size: 13px;
    color: #6b7280;
    margin-bottom: 4px;
  }
  
  .stat-value {
    font-size: 24px;
    font-weight: 700;
    color: #111827;
  }
  
  .stat-value.highlight {
    color: #10b981;
  }
  
  .stat-sublabel {
    font-size: 12px;
    color: #9ca3af;
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
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
  }
  
  .btn-primary:hover {
    background: #2563eb;
  }
  
  .btn-small {
    padding: 6px 16px;
    font-size: 13px;
    border-radius: 6px;
  }
  
  .btn-small.primary {
    background: #3b82f6;
    color: white;
    border: none;
  }
  
  .btn-danger {
    padding: 10px 20px;
    background: #ef4444;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
  }
  
  .btn-danger:hover {
    background: #dc2626;
  }
  
  .btn-link {
    padding: 8px 16px;
    background: none;
    color: #3b82f6;
    border: 1px solid #3b82f6;
    border-radius: 6px;
    font-size: 14px;
    cursor: pointer;
  }
  
  .btn-link:hover {
    background: #eff6ff;
  }
  
  .payroll-list,
  .expense-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 24px;
  }
  
  .payroll-card,
  .expense-card {
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .payroll-card:hover,
  .expense-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
  }
  
  .payroll-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 16px;
  }
  
  .payroll-period h3 {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .payroll-period p {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .status-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;
    text-transform: capitalize;
  }
  
  .bg-gray-100 { background: #f3f4f6; }
  .text-gray-800 { color: #1f2937; }
  .bg-yellow-100 { background: #fef3c7; }
  .text-yellow-800 { color: #92400e; }
  .bg-blue-100 { background: #dbeafe; }
  .text-blue-800 { color: #1e40af; }
  .bg-green-100 { background: #d1fae5; }
  .text-green-800 { color: #065f46; }
  .bg-red-100 { background: #fee2e2; }
  .text-red-800 { color: #991b1b; }
  
  .payroll-body {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin-bottom: 16px;
  }
  
  .payroll-stat {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  
  .payroll-stat .stat-label {
    font-size: 12px;
    color: #6b7280;
  }
  
  .payroll-stat .stat-value {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
  }
  
  .payroll-actions {
    padding-top: 16px;
    border-top: 1px solid #e5e7eb;
  }
  
  .filter-bar {
    display: flex;
    gap: 8px;
    background: white;
    padding: 16px;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
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
    border-color: #3b82f6;
    color: #3b82f6;
  }
  
  .filter-btn.active {
    background: #3b82f6;
    color: white;
    border-color: #3b82f6;
  }
  
  .expense-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 12px;
  }
  
  .expense-info {
    flex: 1;
  }
  
  .expense-employee {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .expense-description {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .expense-amount {
    font-size: 20px;
    font-weight: 700;
    color: #10b981;
  }
  
  .expense-meta {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
    font-size: 13px;
    color: #6b7280;
  }
  
  .expense-category,
  .expense-date {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  
  .meta-icon {
    font-size: 14px;
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
    max-width: 1000px;
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
  
  .payroll-summary {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
    margin-bottom: 24px;
    padding: 20px;
    background: #f9fafb;
    border-radius: 8px;
  }
  
  .summary-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  
  .summary-label {
    font-size: 12px;
    color: #6b7280;
  }
  
  .summary-value {
    font-size: 20px;
    font-weight: 700;
    color: #111827;
  }
  
  .employee-payroll-table {
    overflow-x: auto;
  }
  
  .employee-payroll-table table {
    width: 100%;
    border-collapse: collapse;
  }
  
  .employee-payroll-table th {
    padding: 12px;
    text-align: left;
    font-size: 12px;
    font-weight: 600;
    color: #6b7280;
    text-transform: uppercase;
    background: #f9fafb;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .employee-payroll-table td {
    padding: 12px;
    font-size: 14px;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .net-pay {
    font-weight: 600;
    color: #10b981;
  }
  
  .expense-detail {
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
  
  .expense-actions {
    display: flex;
    gap: 12px;
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid #e5e7eb;
  }
  
  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }
    
    .payroll-body {
      grid-template-columns: repeat(2, 1fr);
    }
    
    .payroll-summary {
      grid-template-columns: 1fr;
    }
    
    .tabs {
      flex-direction: column;
    }
    
    .tab {
      justify-content: flex-start;
    }
  }
</style>