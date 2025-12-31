<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface CompensationPlan {
    id: string;
    employee_id: string;
    employee_name?: string;
    compensation_type: 'salary' | 'hourly' | 'contract';
    base_amount: number;
    currency: string;
    pay_frequency: 'hourly' | 'weekly' | 'biweekly' | 'monthly' | 'annually';
    effective_date: string;
    end_date?: string;
    status: 'active' | 'pending' | 'expired';
    created_at: string;
    updated_at: string;
  }

  interface Bonus {
    id: string;
    employee_id: string;
    employee_name?: string;
    bonus_type: 'monthly' | 'quarterly' | 'annual' | 'performance' | 'signing' | 'retention';
    amount: number;
    currency: string;
    description: string;
    payment_date: string;
    status: 'pending' | 'approved' | 'paid' | 'cancelled';
    created_at: string;
  }

  let compensationPlans: CompensationPlan[] = [];
  let bonuses: Bonus[] = [];
  let employees: any[] = [];
  let loading = false;
  let error = '';
  let success = '';
  let activeTab: 'plans' | 'bonuses' = 'plans';
  
  // Modals
  let showAddCompensation = false;
  let showAddBonus = false;
  let editingPlan: CompensationPlan | null = null;
  let editingBonus: Bonus | null = null;

  // Form data - Compensation
  let newCompensation = {
    employee_id: '',
    compensation_type: 'salary' as 'salary' | 'hourly' | 'contract',
    base_amount: 0,
    currency: 'USD',
    pay_frequency: 'monthly' as any,
    effective_date: new Date().toISOString().split('T')[0],
    end_date: null as string | null,
    status: 'active' as 'active' | 'pending' | 'expired'
  };

  // Form data - Bonus
  let newBonus = {
    employee_id: '',
    bonus_type: 'monthly' as any,
    amount: 0,
    currency: 'USD',
    description: '',
    payment_date: new Date().toISOString().split('T')[0],
    status: 'pending' as any
  };

  onMount(() => {
    loadData();
  });

  async function loadData() {
    await Promise.all([
      loadCompensationPlans(),
      loadBonuses(),
      loadEmployees()
    ]);
  }

  async function loadCompensationPlans() {
    try {
      loading = true;
      const response = await fetch(`${API_BASE_URL}/compensation/plans`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (!response.ok) throw new Error('Failed to load compensation plans');
      const data = await response.json();
      compensationPlans = Array.isArray(data) ? data : [];
    } catch (err: any) {
      error = err.message;
      compensationPlans = []; // Ensure it's always an array
    } finally {
      loading = false;
    }
  }

  async function loadBonuses() {
    try {
      const response = await fetch(`${API_BASE_URL}/compensation/bonuses`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (!response.ok) throw new Error('Failed to load bonuses');
      const data = await response.json();
      bonuses = Array.isArray(data) ? data : [];
    } catch (err: any) {
      error = err.message;
      bonuses = []; // Ensure it's always an array
    }
  }

  async function loadEmployees() {
    try {
      const response = await fetch(`${API_BASE_URL}/employees`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (!response.ok) throw new Error('Failed to load employees');
      const data = await response.json();
      employees = Array.isArray(data) ? data : [];
    } catch (err: any) {
      error = err.message;
      employees = []; // Ensure it's always an array
    }
  }

  function openAddCompensation() {
    editingPlan = null;
    newCompensation = {
      employee_id: '',
      compensation_type: 'salary',
      base_amount: 0,
      currency: 'USD',
      pay_frequency: 'monthly',
      effective_date: new Date().toISOString().split('T')[0],
      end_date: null,
      status: 'active'
    };
    showAddCompensation = true;
  }

  function openAddBonus() {
    editingBonus = null;
    newBonus = {
      employee_id: '',
      bonus_type: 'monthly',
      amount: 0,
      currency: 'USD',
      description: '',
      payment_date: new Date().toISOString().split('T')[0],
      status: 'pending'
    };
    showAddBonus = true;
  }

  async function saveCompensation() {
    try {
      loading = true;
      error = '';

      // Transform the data to ensure proper date format for API
      const payload = {
        ...newCompensation,
        effective_date: newCompensation.effective_date 
          ? new Date(newCompensation.effective_date).toISOString() 
          : new Date().toISOString(),
        end_date: newCompensation.end_date 
          ? new Date(newCompensation.end_date).toISOString() 
          : null
      };

      const method = editingPlan ? 'PUT' : 'POST';
      const url = editingPlan 
        ? `${API_BASE_URL}/compensation/plans/${editingPlan.id}`
        : `${API_BASE_URL}/compensation/plans`;

      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Failed to save compensation plan' }));
        throw new Error(errorData.error || 'Failed to save compensation plan');
      }

      success = editingPlan ? 'Compensation updated!' : 'Compensation created!';
      setTimeout(() => success = '', 3000);
      
      showAddCompensation = false;
      await loadCompensationPlans();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function saveBonus() {
    try {
      loading = true;
      error = '';

      // Transform the data to ensure proper date format for API
      const payload = {
        ...newBonus,
        payment_date: newBonus.payment_date 
          ? new Date(newBonus.payment_date).toISOString() 
          : new Date().toISOString()
      };

      const method = editingBonus ? 'PUT' : 'POST';
      const url = editingBonus 
        ? `${API_BASE_URL}/compensation/bonuses/${editingBonus.id}`
        : `${API_BASE_URL}/compensation/bonuses`;

      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Failed to save bonus' }));
        throw new Error(errorData.error || 'Failed to save bonus');
      }

      success = editingBonus ? 'Bonus updated!' : 'Bonus created!';
      setTimeout(() => success = '', 3000);
      
      showAddBonus = false;
      await loadBonuses();
    } catch (err: any) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function editCompensation(plan: CompensationPlan) {
    editingPlan = plan;
    newCompensation = {
      employee_id: plan.employee_id,
      compensation_type: plan.compensation_type,
      base_amount: plan.base_amount,
      currency: plan.currency,
      pay_frequency: plan.pay_frequency,
      effective_date: plan.effective_date,
      end_date: plan.end_date || '',
      status: plan.status
    };
    showAddCompensation = true;
  }

  function editBonusItem(bonus: Bonus) {
    editingBonus = bonus;
    newBonus = {
      employee_id: bonus.employee_id,
      bonus_type: bonus.bonus_type,
      amount: bonus.amount,
      currency: bonus.currency,
      description: bonus.description,
      payment_date: bonus.payment_date,
      status: bonus.status
    };
    showAddBonus = true;
  }

  async function deleteCompensation(id: string) {
    if (!confirm('Delete this compensation plan?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/compensation/plans/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });

      if (!response.ok) throw new Error('Failed to delete');
      
      success = 'Compensation plan deleted';
      setTimeout(() => success = '', 3000);
      await loadCompensationPlans();
    } catch (err: any) {
      error = err.message;
    }
  }

  async function deleteBonus(id: string) {
    if (!confirm('Delete this bonus?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/compensation/bonuses/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });

      if (!response.ok) throw new Error('Failed to delete');
      
      success = 'Bonus deleted';
      setTimeout(() => success = '', 3000);
      await loadBonuses();
    } catch (err: any) {
      error = err.message;
    }
  }

  function formatCurrency(amount: number, currency: string = 'USD') {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency
    }).format(amount);
  }

  function formatDate(date: string) {
    return new Date(date).toLocaleDateString();
  }

  function getEmployeeName(employeeId: string) {
    const emp = employees.find(e => e.id === employeeId);
    return emp ? `${emp.first_name} ${emp.last_name}` : 'Unknown';
  }

  function getStatusBadgeClass(status: string) {
    const classes: Record<string, string> = {
      active: 'status-active',
      pending: 'status-pending',
      expired: 'status-expired',
      approved: 'status-approved',
      paid: 'status-paid',
      cancelled: 'status-cancelled'
    };
    return classes[status] || '';
  }
</script>

<div class="compensation-manager">
  <div class="header">
    <h2>💰 Compensation Management</h2>
    <div class="header-actions">
      {#if activeTab === 'plans'}
        <button class="btn btn-primary" onclick={() => openAddCompensation()}>
          + Add Compensation Plan
        </button>
      {:else}
        <button class="btn btn-primary" onclick={() => openAddBonus()}>
          + Add Bonus
        </button>
      {/if}
    </div>
  </div>

  {#if error}
    <div class="alert alert-error">{error}</div>
  {/if}

  {#if success}
    <div class="alert alert-success">{success}</div>
  {/if}

  <!-- Tabs -->
  <div class="tabs">
    <button 
      class="tab {activeTab === 'plans' ? 'active' : ''}"
      onclick={() => activeTab = 'plans'}
    >
      💼 Compensation Plans
    </button>
    <button 
      class="tab {activeTab === 'bonuses' ? 'active' : ''}"
      onclick={() => activeTab = 'bonuses'}
    >
      🎁 Bonuses
    </button>
  </div>

  {#if loading && (compensationPlans?.length ?? 0) === 0 && (bonuses?.length ?? 0) === 0}
    <div class="loading">Loading...</div>
  {:else}
    {#if activeTab === 'plans'}
      <!-- Compensation Plans Table -->
      <div class="table-container">
        <table>
          <thead>
            <tr>
              <th>Employee</th>
              <th>Type</th>
              <th>Base Amount</th>
              <th>Frequency</th>
              <th>Effective Date</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#if !compensationPlans || compensationPlans.length === 0}
              <tr>
                <td colspan="7" class="empty-state">
                  No compensation plans found. Add one above!
                </td>
              </tr>
            {:else}
              {#each compensationPlans as plan}
                <tr>
                  <td class="employee-name">
                    {getEmployeeName(plan.employee_id)}
                  </td>
                  <td>
                    <span class="type-badge type-{plan.compensation_type}">
                      {plan.compensation_type}
                    </span>
                  </td>
                  <td class="amount">
                    {formatCurrency(plan.base_amount, plan.currency)}
                  </td>
                  <td>{plan.pay_frequency}</td>
                  <td>{formatDate(plan.effective_date)}</td>
                  <td>
                    <span class="status-badge {getStatusBadgeClass(plan.status)}">
                      {plan.status}
                    </span>
                  </td>
                  <td class="actions">
                    <button 
                      class="btn-icon" 
                      onclick={() => editCompensation(plan)}
                      title="Edit"
                    >
                      ✏️
                    </button>
                    <button 
                      class="btn-icon" 
                      onclick={() => deleteCompensation(plan.id)}
                      title="Delete"
                    >
                      🗑️
                    </button>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    {:else}
      <!-- Bonuses Table -->
      <div class="table-container">
        <table>
          <thead>
            <tr>
              <th>Employee</th>
              <th>Type</th>
              <th>Amount</th>
              <th>Description</th>
              <th>Payment Date</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#if !bonuses || bonuses.length === 0}
              <tr>
                <td colspan="7" class="empty-state">
                  No bonuses found. Add one above!
                </td>
              </tr>
            {:else}
              {#each bonuses as bonus}
                <tr>
                  <td class="employee-name">
                    {getEmployeeName(bonus.employee_id)}
                  </td>
                  <td>
                    <span class="type-badge type-{bonus.bonus_type}">
                      {bonus.bonus_type}
                    </span>
                  </td>
                  <td class="amount">
                    {formatCurrency(bonus.amount, bonus.currency)}
                  </td>
                  <td>{bonus.description}</td>
                  <td>{formatDate(bonus.payment_date)}</td>
                  <td>
                    <span class="status-badge {getStatusBadgeClass(bonus.status)}">
                      {bonus.status}
                    </span>
                  </td>
                  <td class="actions">
                    <button 
                      class="btn-icon" 
                      onclick={() => editBonusItem(bonus)}
                      title="Edit"
                    >
                      ✏️
                    </button>
                    <button 
                      class="btn-icon" 
                      onclick={() => deleteBonus(bonus.id)}
                      title="Delete"
                    >
                      🗑️
                    </button>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    {/if}
  {/if}
</div>

<!-- Add/Edit Compensation Modal -->
{#if showAddCompensation}
  <div class="modal-overlay" role="button" tabindex="0" onclick={() => showAddCompensation = false} onkeydown={(e) => e.key === 'Escape' && (showAddCompensation = false)}>
    <div class="modal" role="dialog" aria-modal="true" tabindex="-1" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showAddCompensation = false)}>
      <div class="modal-header">
        <h3>{editingPlan ? 'Edit' : 'Add'} Compensation Plan</h3>
        <button class="close-btn" onclick={() => showAddCompensation = false}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-grid">
          <div class="form-group">
            <label for="employee">Employee *</label>
            <select id="employee" bind:value={newCompensation.employee_id} required>
              <option value="">Select Employee</option>
              {#if employees && employees.length > 0}
                {#each employees as emp}
                  <option value={emp.id}>
                    {emp.first_name} {emp.last_name}
                  </option>
                {/each}
              {/if}
            </select>
          </div>

          <div class="form-group">
            <label for="comp-type">Compensation Type *</label>
            <select id="comp-type" bind:value={newCompensation.compensation_type} required>
              <option value="salary">Salary</option>
              <option value="hourly">Hourly</option>
              <option value="contract">Contract</option>
            </select>
          </div>

          <div class="form-group">
            <label for="base-amount">Base Amount *</label>
            <input 
              id="base-amount"
              type="number" 
              bind:value={newCompensation.base_amount}
              min="0"
              step="0.01"
              required
            />
          </div>

          <div class="form-group">
            <label for="currency">Currency</label>
            <select id="currency" bind:value={newCompensation.currency}>
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
              <option value="GBP">GBP</option>
              <option value="CAD">CAD</option>
            </select>
          </div>

          <div class="form-group">
            <label for="pay-freq">Pay Frequency *</label>
            <select id="pay-freq" bind:value={newCompensation.pay_frequency} required>
              <option value="hourly">Hourly</option>
              <option value="weekly">Weekly</option>
              <option value="biweekly">Bi-weekly</option>
              <option value="monthly">Monthly</option>
              <option value="annually">Annually</option>
            </select>
          </div>

          <div class="form-group">
            <label for="effective-date">Effective Date *</label>
            <input 
              id="effective-date"
              type="date" 
              bind:value={newCompensation.effective_date}
              required
            />
          </div>

          <div class="form-group">
            <label for="end-date">End Date (Optional)</label>
            <input 
              id="end-date"
              type="date" 
              bind:value={newCompensation.end_date}
            />
          </div>

          <div class="form-group">
            <label for="status">Status</label>
            <select id="status" bind:value={newCompensation.status}>
              <option value="active">Active</option>
              <option value="pending">Pending</option>
              <option value="expired">Expired</option>
            </select>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showAddCompensation = false}>
          Cancel
        </button>
        <button class="btn btn-primary" onclick={saveCompensation} disabled={loading}>
          {editingPlan ? 'Update' : 'Create'} Plan
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Add/Edit Bonus Modal -->
{#if showAddBonus}
  <div class="modal-overlay" role="button" tabindex="0" onclick={() => showAddBonus = false} onkeydown={(e) => e.key === 'Escape' && (showAddBonus = false)}>
    <div class="modal" role="dialog" aria-modal="true" tabindex="-1" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && (showAddBonus = false)}>
      <div class="modal-header">
        <h3>{editingBonus ? 'Edit' : 'Add'} Bonus</h3>
        <button class="close-btn" onclick={() => showAddBonus = false}>×</button>
      </div>

      <div class="modal-body">
        <div class="form-grid">
          <div class="form-group">
            <label for="bonus-employee">Employee *</label>
            <select id="bonus-employee" bind:value={newBonus.employee_id} required>
              <option value="">Select Employee</option>
              {#if employees && employees.length > 0}
                {#each employees as emp}
                  <option value={emp.id}>
                    {emp.first_name} {emp.last_name}
                  </option>
                {/each}
              {/if}
            </select>
          </div>

          <div class="form-group">
            <label for="bonus-type">Bonus Type *</label>
            <select id="bonus-type" bind:value={newBonus.bonus_type} required>
              <option value="monthly">Monthly</option>
              <option value="quarterly">Quarterly</option>
              <option value="annual">Annual</option>
              <option value="performance">Performance</option>
              <option value="signing">Signing</option>
              <option value="retention">Retention</option>
            </select>
          </div>

          <div class="form-group">
            <label for="bonus-amount">Amount *</label>
            <input 
              id="bonus-amount"
              type="number" 
              bind:value={newBonus.amount}
              min="0"
              step="0.01"
              required
            />
          </div>

          <div class="form-group">
            <label for="bonus-currency">Currency</label>
            <select id="bonus-currency" bind:value={newBonus.currency}>
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
              <option value="GBP">GBP</option>
              <option value="CAD">CAD</option>
            </select>
          </div>

          <div class="form-group full-width">
            <label for="bonus-desc">Description *</label>
            <textarea 
              id="bonus-desc"
              bind:value={newBonus.description}
              rows="3"
              required
            ></textarea>
          </div>

          <div class="form-group">
            <label for="payment-date">Payment Date *</label>
            <input 
              id="payment-date"
              type="date" 
              bind:value={newBonus.payment_date}
              required
            />
          </div>

          <div class="form-group">
            <label for="bonus-status">Status</label>
            <select id="bonus-status" bind:value={newBonus.status}>
              <option value="pending">Pending</option>
              <option value="approved">Approved</option>
              <option value="paid">Paid</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => showAddBonus = false}>
          Cancel
        </button>
        <button class="btn btn-primary" onclick={saveBonus} disabled={loading}>
          {editingBonus ? 'Update' : 'Create'} Bonus
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .compensation-manager {
    padding: 20px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
  }

  .header h2 {
    margin: 0;
    font-size: 24px;
    color: #1f2937;
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }

  .alert {
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 16px;
  }

  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .alert-success {
    background: #d1fae5;
    color: #065f46;
    border: 1px solid #a7f3d0;
  }

  .tabs {
    display: flex;
    gap: 8px;
    margin-bottom: 24px;
    border-bottom: 2px solid #e5e7eb;
  }

  .tab {
    padding: 12px 24px;
    background: none;
    border: none;
    border-bottom: 3px solid transparent;
    color: #6b7280;
    font-size: 15px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: -2px;
  }

  .tab:hover {
    color: #374151;
    background: #f9fafb;
  }

  .tab.active {
    color: #3b82f6;
    border-bottom-color: #3b82f6;
  }

  .table-container {
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
    background: #f9fafb;
    position: sticky;
    top: 0;
    z-index: 10;
  }

  th {
    padding: 16px;
    text-align: left;
    font-size: 12px;
    font-weight: 600;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    border-bottom: 2px solid #e9ecef;
  }

  tbody tr:nth-child(odd) {
    background: #ffffff;
  }

  tbody tr:nth-child(even) {
    background: #f9fafb;
  }

  tbody tr:hover {
    background: #e0f2fe !important;
    cursor: pointer;
  }

  td {
    padding: 16px;
    border-bottom: 1px solid #e9ecef;
    font-size: 14px;
    color: #1f2937;
  }

  .employee-name {
    font-weight: 600;
    color: #111827;
  }

  .amount {
    font-weight: 600;
    color: #059669;
  }

  .empty-state {
    text-align: center;
    color: #9ca3af;
    padding: 40px !important;
  }

  .type-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .type-salary {
    background: #dbeafe;
    color: #1e40af;
  }

  .type-hourly {
    background: #fef3c7;
    color: #92400e;
  }

  .type-contract {
    background: #e0e7ff;
    color: #4338ca;
  }

  .type-monthly, .type-quarterly, .type-annual {
    background: #dcfce7;
    color: #166534;
  }

  .type-performance {
    background: #fce7f3;
    color: #9f1239;
  }

  .type-signing, .type-retention {
    background: #e0e7ff;
    color: #4338ca;
  }

  .status-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .status-active, .status-approved, .status-paid {
    background: #d1fae5;
    color: #065f46;
  }

  .status-pending {
    background: #fef3c7;
    color: #92400e;
  }

  .status-expired, .status-cancelled {
    background: #fee2e2;
    color: #991b1b;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .btn-icon {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 18px;
    padding: 4px 8px;
    transition: transform 0.2s;
  }

  .btn-icon:hover {
    transform: scale(1.2);
  }

  .btn {
    padding: 10px 20px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-primary:hover {
    background: #2563eb;
  }

  .btn-secondary {
    background: #e5e7eb;
    color: #374151;
  }

  .btn-secondary:hover {
    background: #d1d5db;
  }

  .loading {
    text-align: center;
    padding: 40px;
    color: #6b7280;
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
  }

  .modal {
    background: white;
    border-radius: 12px;
    width: 90%;
    max-width: 800px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 20px;
    color: #1f2937;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 28px;
    color: #9ca3af;
    cursor: pointer;
    line-height: 1;
    padding: 0;
    width: 32px;
    height: 32px;
  }

  .close-btn:hover {
    color: #4b5563;
  }

  .modal-body {
    padding: 24px;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
  }

  .form-group.full-width {
    grid-column: span 2;
  }

  .form-group label {
    font-size: 14px;
    font-weight: 500;
    color: #374151;
    margin-bottom: 6px;
  }

  .form-group input,
  .form-group select,
  .form-group textarea {
    padding: 10px 12px;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 14px;
    transition: all 0.2s;
  }

  .form-group input:focus,
  .form-group select:focus,
  .form-group textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .form-group textarea {
    resize: vertical;
    font-family: inherit;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 20px 24px;
    border-top: 1px solid #e5e7eb;
  }
</style>