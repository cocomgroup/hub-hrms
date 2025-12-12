<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth-store';
  
  // Types
  interface BenefitPlan {
    id: string;
    name: string;
    category: string;
    plan_type: string;
    provider: string;
    description: string;
    employee_cost: number;
    employer_cost: number;
    deductible_single: number;
    deductible_family: number;
    out_of_pocket_max_single: number;
    out_of_pocket_max_family: number;
    copay_primary_care: number;
    copay_specialist: number;
    copay_emergency: number;
    coinsurance_rate: number;
    active: boolean;
    enrollment_start_date: string;
    enrollment_end_date: string;
    effective_date: string;
  }

  interface Dependent {
    id?: string;
    first_name: string;
    last_name: string;
    relationship: string;
    date_of_birth: string;
    ssn?: string;
  }

  interface Enrollment {
    id: string;
    employee_id: string;
    plan_id: string;
    coverage_level: string;
    status: string;
    enrollment_date: string;
    effective_date: string;
    termination_date?: string;
    employee_cost: number;
    employer_cost: number;
    total_cost: number;
    payroll_deduction: number;
    plan_name: string;
    plan_category: string;
    dependents: Dependent[];
  }

  interface BenefitsSummary {
    employee_id: string;
    active_enrollments: number;
    total_employee_cost: number;
    total_employer_cost: number;
    monthly_deduction: number;
    enrollments: Enrollment[];
    available_plans: BenefitPlan[];
  }

  // State
  let loading = false;
  let error = '';
  let success = '';
  let activeTab: 'plans' | 'enrollments' | 'summary' = 'plans';
  
  let plans: BenefitPlan[] = [];
  let enrollments: Enrollment[] = [];
  let summary: BenefitsSummary | null = null;
  
  let showEnrollModal = false;
  let selectedPlan: BenefitPlan | null = null;
  let coverageLevel: string = 'employee';
  let effectiveDate: string = '';
  let dependents: Dependent[] = [];
  
  let showPlanDetails = false;
  let selectedPlanDetails: BenefitPlan | null = null;
  
  let selectedCategory = 'all';
  
  // Coverage level costs multiplier
  const coverageMultipliers = {
    employee: 1.0,
    employee_spouse: 1.5,
    employee_child: 1.3,
    family: 2.0
  };

  // Computed
  $: filteredPlans = selectedCategory === 'all' 
    ? plans 
    : plans.filter(p => p.category === selectedCategory);
  
  $: categories = [...new Set(plans.map(p => p.category))];
  
  $: calculatedCosts = selectedPlan ? {
    employeeCost: selectedPlan.employee_cost * (coverageMultipliers[coverageLevel] || 1.0),
    employerCost: selectedPlan.employer_cost * (coverageMultipliers[coverageLevel] || 1.0),
    totalCost: (selectedPlan.employee_cost + selectedPlan.employer_cost) * (coverageMultipliers[coverageLevel] || 1.0),
    monthlyDeduction: (selectedPlan.employee_cost * (coverageMultipliers[coverageLevel] || 1.0)) / 12
  } : null;

  // API calls
  async function loadPlans() {
    try {
      loading = true;
      error = '';
      
      const response = await fetch('/api/benefits/plans?active=true', {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to load plans');
      
      plans = await response.json();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function loadEnrollments() {
    try {
      loading = true;
      error = '';
      
      const response = await fetch('/api/benefits/enrollments', {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to load enrollments');
      
      enrollments = await response.json();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function loadSummary() {
    try {
      loading = true;
      error = '';
      
      const response = await fetch('/api/benefits/summary', {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) throw new Error('Failed to load summary');
      
      summary = await response.json();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function createEnrollment() {
    try {
      loading = true;
      error = '';
      success = '';
      
      // Validate
      if (!selectedPlan) {
        error = 'Please select a plan';
        return;
      }
      
      if (!effectiveDate) {
        error = 'Please select an effective date';
        return;
      }
      
      // Validate dependents based on coverage level
      if (coverageLevel === 'employee_spouse' || coverageLevel === 'family') {
        const hasSpouse = dependents.some(d => d.relationship === 'spouse');
        if (!hasSpouse) {
          error = 'Spouse coverage requires at least one spouse dependent';
          return;
        }
      }
      
      if (coverageLevel === 'employee_child' || coverageLevel === 'family') {
        const hasChild = dependents.some(d => d.relationship === 'child');
        if (!hasChild) {
          error = 'Child coverage requires at least one child dependent';
          return;
        }
      }
      
      const response = await fetch('/api/benefits/enrollments', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          plan_id: selectedPlan.id,
          coverage_level: coverageLevel,
          effective_date: effectiveDate,
          dependents: dependents.map(d => ({
            first_name: d.first_name,
            last_name: d.last_name,
            relationship: d.relationship,
            date_of_birth: d.date_of_birth,
            ssn: d.ssn || ''
          }))
        })
      });
      
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to create enrollment');
      }
      
      success = 'Enrollment created successfully!';
      showEnrollModal = false;
      resetEnrollmentForm();
      
      // Reload data
      await loadEnrollments();
      await loadSummary();
      
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  async function cancelEnrollment(enrollmentId: string) {
    if (!confirm('Are you sure you want to cancel this enrollment? This will take effect at the end of the current month.')) {
      return;
    }
    
    try {
      loading = true;
      error = '';
      success = '';
      
      const response = await fetch(`/api/benefits/enrollments/${enrollmentId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to cancel enrollment');
      }
      
      success = 'Enrollment cancelled successfully';
      
      // Reload data
      await loadEnrollments();
      await loadSummary();
      
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function openEnrollModal(plan: BenefitPlan) {
    selectedPlan = plan;
    coverageLevel = 'employee';
    effectiveDate = '';
    dependents = [];
    showEnrollModal = true;
  }

  function resetEnrollmentForm() {
    selectedPlan = null;
    coverageLevel = 'employee';
    effectiveDate = '';
    dependents = [];
  }

  function addDependent() {
    dependents = [...dependents, {
      first_name: '',
      last_name: '',
      relationship: 'spouse',
      date_of_birth: '',
      ssn: ''
    }];
  }

  function removeDependent(index: number) {
    dependents = dependents.filter((_, i) => i !== index);
  }

  function viewPlanDetails(plan: BenefitPlan) {
    selectedPlanDetails = plan;
    showPlanDetails = true;
  }

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

  function getCategoryIcon(category: string): string {
    const icons = {
      health: '🏥',
      dental: '🦷',
      vision: '👁️',
      life: '🛡️',
      disability: '♿',
      retirement: '🏦',
      fsa: '💰',
      hsa: '💳',
      commuter: '🚌',
      wellness: '💪',
      other: '📋'
    };
    return icons[category] || '📋';
  }

  function getStatusBadgeClass(status: string): string {
    const classes = {
      active: 'badge-success',
      pending: 'badge-warning',
      cancelled: 'badge-error',
      expired: 'badge-ghost'
    };
    return classes[status] || 'badge-ghost';
  }

  onMount(() => {
    loadPlans();
    loadEnrollments();
    loadSummary();
  });
</script>

<div class="benefits-container">
  <!-- Header -->
  <div class="benefits-header">
    <h1>Benefits</h1>
    <p class="text-muted">Manage your health and welfare benefits</p>
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
      class="tab {activeTab === 'plans' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'plans'}
    >
      Available Plans
    </button>
    <button 
      class="tab {activeTab === 'enrollments' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'enrollments'}
    >
      My Enrollments
    </button>
    <button 
      class="tab {activeTab === 'summary' ? 'tab-active' : ''}"
      on:click={() => activeTab = 'summary'}
    >
      Summary
    </button>
  </div>

  <!-- Content -->
  <div class="tab-content">
    {#if loading}
      <div class="loading-spinner">
        <div class="spinner"></div>
        <p>Loading benefits...</p>
      </div>
    {:else if activeTab === 'plans'}
      <!-- Available Plans Tab -->
      <div class="plans-section">
        <!-- Category Filter -->
        <div class="filter-bar">
          <label>
            Filter by category:
            <select bind:value={selectedCategory}>
              <option value="all">All Categories</option>
              {#each categories as category}
                <option value={category}>
                  {getCategoryIcon(category)} {category.charAt(0).toUpperCase() + category.slice(1)}
                </option>
              {/each}
            </select>
          </label>
        </div>

        <!-- Plans Grid -->
        <div class="plans-grid">
          {#each filteredPlans as plan}
            <div class="plan-card">
              <div class="plan-header">
                <div class="plan-icon">
                  {getCategoryIcon(plan.category)}
                </div>
                <div class="plan-title">
                  <h3>{plan.name}</h3>
                  <p class="plan-provider">{plan.provider}</p>
                  <span class="badge badge-info">{plan.category}</span>
                  {#if plan.plan_type}
                    <span class="badge badge-outline">{plan.plan_type.toUpperCase()}</span>
                  {/if}
                </div>
              </div>

              <div class="plan-body">
                <p class="plan-description">{plan.description || 'No description available'}</p>

                <div class="plan-costs">
                  <div class="cost-item">
                    <span class="cost-label">Your Cost (Annual)</span>
                    <span class="cost-value">{formatCurrency(plan.employee_cost)}</span>
                  </div>
                  <div class="cost-item">
                    <span class="cost-label">Employer Contribution</span>
                    <span class="cost-value employer">{formatCurrency(plan.employer_cost)}</span>
                  </div>
                  <div class="cost-item total">
                    <span class="cost-label">Monthly Deduction</span>
                    <span class="cost-value">{formatCurrency(plan.employee_cost / 12)}</span>
                  </div>
                </div>

                {#if plan.category === 'health'}
                  <div class="plan-details-quick">
                    <div class="detail-row">
                      <span>Deductible (Single):</span>
                      <strong>{formatCurrency(plan.deductible_single)}</strong>
                    </div>
                    <div class="detail-row">
                      <span>PCP Copay:</span>
                      <strong>{formatCurrency(plan.copay_primary_care)}</strong>
                    </div>
                    <div class="detail-row">
                      <span>Specialist Copay:</span>
                      <strong>{formatCurrency(plan.copay_specialist)}</strong>
                    </div>
                  </div>
                {/if}
              </div>

              <div class="plan-footer">
                <button class="btn btn-outline btn-sm" on:click={() => viewPlanDetails(plan)}>
                  View Details
                </button>
                <button class="btn btn-primary btn-sm" on:click={() => openEnrollModal(plan)}>
                  Enroll
                </button>
              </div>
            </div>
          {/each}

          {#if filteredPlans.length === 0}
            <div class="empty-state">
              <p>No plans available in this category</p>
            </div>
          {/if}
        </div>
      </div>

    {:else if activeTab === 'enrollments'}
      <!-- My Enrollments Tab -->
      <div class="enrollments-section">
        {#if enrollments.length === 0}
          <div class="empty-state">
            <div class="empty-icon">📋</div>
            <h3>No Enrollments</h3>
            <p>You haven't enrolled in any benefits yet</p>
            <button class="btn btn-primary" on:click={() => activeTab = 'plans'}>
              Browse Available Plans
            </button>
          </div>
        {:else}
          <div class="enrollments-list">
            {#each enrollments as enrollment}
              <div class="enrollment-card">
                <div class="enrollment-header">
                  <div>
                    <h3>{enrollment.plan_name}</h3>
                    <p class="enrollment-category">{enrollment.plan_category}</p>
                  </div>
                  <span class="badge {getStatusBadgeClass(enrollment.status)}">
                    {enrollment.status}
                  </span>
                </div>

                <div class="enrollment-body">
                  <div class="enrollment-info-grid">
                    <div class="info-item">
                      <span class="info-label">Coverage Level</span>
                      <span class="info-value">{enrollment.coverage_level.replace('_', ' + ')}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Effective Date</span>
                      <span class="info-value">{formatDate(enrollment.effective_date)}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Your Annual Cost</span>
                      <span class="info-value">{formatCurrency(enrollment.employee_cost)}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Monthly Deduction</span>
                      <span class="info-value">{formatCurrency(enrollment.payroll_deduction)}</span>
                    </div>
                  </div>

                  {#if enrollment.dependents && enrollment.dependents.length > 0}
                    <div class="dependents-section">
                      <h4>Covered Dependents</h4>
                      <div class="dependents-list">
                        {#each enrollment.dependents as dependent}
                          <div class="dependent-card">
                            <div class="dependent-icon">
                              {dependent.relationship === 'spouse' ? '💑' : 
                               dependent.relationship === 'child' ? '👶' : 
                               dependent.relationship === 'domestic_partner' ? '👥' : '👤'}
                            </div>
                            <div class="dependent-info">
                              <strong>{dependent.first_name} {dependent.last_name}</strong>
                              <span class="dependent-relationship">{dependent.relationship}</span>
                              <span class="dependent-dob">DOB: {formatDate(dependent.date_of_birth)}</span>
                            </div>
                          </div>
                        {/each}
                      </div>
                    </div>
                  {/if}
                </div>

                {#if enrollment.status === 'active'}
                  <div class="enrollment-footer">
                    <button 
                      class="btn btn-error btn-sm"
                      on:click={() => cancelEnrollment(enrollment.id)}
                    >
                      Cancel Enrollment
                    </button>
                  </div>
                {/if}

                {#if enrollment.termination_date}
                  <div class="termination-notice">
                    ⚠️ This enrollment will terminate on {formatDate(enrollment.termination_date)}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}
      </div>

    {:else if activeTab === 'summary'}
      <!-- Summary Tab -->
      <div class="summary-section">
        {#if summary}
          <div class="summary-cards">
            <div class="summary-card">
              <div class="summary-icon">📊</div>
              <div class="summary-content">
                <h3>{summary.active_enrollments}</h3>
                <p>Active Enrollments</p>
              </div>
            </div>

            <div class="summary-card">
              <div class="summary-icon">💰</div>
              <div class="summary-content">
                <h3>{formatCurrency(summary.total_employee_cost)}</h3>
                <p>Your Annual Cost</p>
              </div>
            </div>

            <div class="summary-card">
              <div class="summary-icon">🏢</div>
              <div class="summary-content">
                <h3>{formatCurrency(summary.total_employer_cost)}</h3>
                <p>Employer Contribution</p>
              </div>
            </div>

            <div class="summary-card">
              <div class="summary-icon">📅</div>
              <div class="summary-content">
                <h3>{formatCurrency(summary.monthly_deduction)}</h3>
                <p>Monthly Deduction</p>
              </div>
            </div>
          </div>

          <div class="summary-breakdown">
            <h3>Cost Breakdown by Category</h3>
            <div class="breakdown-table">
              {#each [...new Set(summary.enrollments.filter(e => e.status === 'active').map(e => e.plan_category))] as category}
                {@const categoryEnrollments = summary.enrollments.filter(e => e.status === 'active' && e.plan_category === category)}
                {@const categoryTotal = categoryEnrollments.reduce((sum, e) => sum + e.employee_cost, 0)}
                {@const categoryMonthly = categoryEnrollments.reduce((sum, e) => sum + e.payroll_deduction, 0)}
                
                <div class="breakdown-row">
                  <div class="breakdown-category">
                    <span class="category-icon">{getCategoryIcon(category)}</span>
                    <span class="category-name">{category}</span>
                    <span class="category-count">({categoryEnrollments.length} plan{categoryEnrollments.length !== 1 ? 's' : ''})</span>
                  </div>
                  <div class="breakdown-costs">
                    <span class="breakdown-annual">{formatCurrency(categoryTotal)}/year</span>
                    <span class="breakdown-monthly">{formatCurrency(categoryMonthly)}/mo</span>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {:else}
          <div class="empty-state">
            <p>No summary data available</p>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<!-- Enrollment Modal -->
{#if showEnrollModal && selectedPlan}
  <div class="modal modal-open">
    <div class="modal-box max-w-3xl">
      <h3 class="font-bold text-lg mb-4">Enroll in {selectedPlan.name}</h3>
      
      <div class="enrollment-form">
        <!-- Coverage Level -->
        <div class="form-control">
          <label class="label" for="coverage-level">
            <span class="label-text">Coverage Level *</span>
          </label>
          <select id="coverage-level" bind:value={coverageLevel} class="select select-bordered w-full">
            <option value="employee">Employee Only</option>
            <option value="employee_spouse">Employee + Spouse</option>
            <option value="employee_child">Employee + Child(ren)</option>
            <option value="family">Family</option>
          </select>
        </div>

        <!-- Effective Date -->
        <div class="form-control">
          <label class="label" for="effective-date">
            <span class="label-text">Effective Date *</span>
          </label>
          <input 
            id="effective-date"
            type="date" 
            bind:value={effectiveDate}
            min={new Date().toISOString().split('T')[0]}
            class="input input-bordered w-full"
          />
        </div>

        <!-- Cost Calculation -->
        {#if calculatedCosts}
          <div class="cost-calculation">
            <h4>Cost Calculation</h4>
            <div class="cost-grid">
              <div class="cost-row">
                <span>Your Annual Cost:</span>
                <strong>{formatCurrency(calculatedCosts.employeeCost)}</strong>
              </div>
              <div class="cost-row">
                <span>Employer Contribution:</span>
                <strong class="text-success">{formatCurrency(calculatedCosts.employerCost)}</strong>
              </div>
              <div class="cost-row">
                <span>Total Annual Cost:</span>
                <strong>{formatCurrency(calculatedCosts.totalCost)}</strong>
              </div>
              <div class="cost-row highlight">
                <span>Monthly Payroll Deduction:</span>
                <strong>{formatCurrency(calculatedCosts.monthlyDeduction)}</strong>
              </div>
            </div>
          </div>
        {/if}

        <!-- Dependents -->
        {#if coverageLevel !== 'employee'}
          <div class="dependents-form">
            <div class="dependents-header">
              <h4>Dependents</h4>
              <button class="btn btn-sm btn-outline" on:click={addDependent}>
                + Add Dependent
              </button>
            </div>

            {#each dependents as dependent, index}
              <div class="dependent-form-card">
                <div class="dependent-form-header">
                  <h5>Dependent {index + 1}</h5>
                  <button 
                    class="btn btn-ghost btn-xs"
                    on:click={() => removeDependent(index)}
                  >
                    Remove
                  </button>
                </div>

                <div class="form-grid">
                  <div class="form-control">
                    <label class="label" for="dependent-first-name-{index}">
                      <span class="label-text">First Name *</span>
                    </label>
                    <input 
                      id="dependent-first-name-{index}"
                      type="text" 
                      bind:value={dependent.first_name}
                      class="input input-bordered input-sm"
                      required
                    />
                  </div>

                  <div class="form-control">
                    <label class="label" for="dependent-last-name-{index}">
                      <span class="label-text">Last Name *</span>
                    </label>
                    <input 
                      id="dependent-last-name-{index}"
                      type="text" 
                      bind:value={dependent.last_name}
                      class="input input-bordered input-sm"
                      required
                    />
                  </div>

                  <div class="form-control">
                    <label class="label" for="dependent-relationship-{index}">
                      <span class="label-text">Relationship *</span>
                    </label>
                    <select id="dependent-relationship-{index}" bind:value={dependent.relationship} class="select select-bordered select-sm">
                      <option value="spouse">Spouse</option>
                      <option value="child">Child</option>
                      <option value="domestic_partner">Domestic Partner</option>
                      <option value="parent">Parent</option>
                    </select>
                  </div>

                  <div class="form-control">
                    <label class="label" for="dependent-dob-{index}">
                      <span class="label-text">Date of Birth *</span>
                    </label>
                    <input 
                      id="dependent-dob-{index}"
                      type="date" 
                      bind:value={dependent.date_of_birth}
                      class="input input-bordered input-sm"
                      required
                    />
                  </div>

                  <div class="form-control col-span-2">
                    <label class="label" for="dependent-ssn-{index}">
                      <span class="label-text">SSN (Optional)</span>
                    </label>
                    <input 
                      id="dependent-ssn-{index}"
                      type="text" 
                      bind:value={dependent.ssn}
                      placeholder="XXX-XX-XXXX"
                      class="input input-bordered input-sm"
                      maxlength="11"
                    />
                  </div>
                </div>
              </div>
            {/each}

            {#if dependents.length === 0}
              <p class="text-muted text-sm">Please add at least one dependent for this coverage level</p>
            {/if}
          </div>
        {/if}
      </div>

      <div class="modal-action">
        <button 
          class="btn" 
          on:click={() => { showEnrollModal = false; resetEnrollmentForm(); }}
          disabled={loading}
        >
          Cancel
        </button>
        <button 
          class="btn btn-primary" 
          on:click={createEnrollment}
          disabled={loading || !effectiveDate}
        >
          {loading ? 'Enrolling...' : 'Confirm Enrollment'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Plan Details Modal -->
{#if showPlanDetails && selectedPlanDetails}
  <div class="modal modal-open">
    <div class="modal-box max-w-2xl">
      <h3 class="font-bold text-lg mb-4">{selectedPlanDetails.name}</h3>
      
      <div class="plan-details-full">
        <div class="detail-section">
          <h4>Provider Information</h4>
          <div class="detail-grid">
            <div><strong>Provider:</strong> {selectedPlanDetails.provider}</div>
            <div><strong>Plan Type:</strong> {selectedPlanDetails.plan_type?.toUpperCase() || 'N/A'}</div>
            <div><strong>Category:</strong> {selectedPlanDetails.category}</div>
          </div>
        </div>

        <div class="detail-section">
          <h4>Description</h4>
          <p>{selectedPlanDetails.description || 'No description available'}</p>
        </div>

        <div class="detail-section">
          <h4>Cost Information</h4>
          <div class="detail-grid">
            <div><strong>Employee Annual Cost:</strong> {formatCurrency(selectedPlanDetails.employee_cost)}</div>
            <div><strong>Employer Contribution:</strong> {formatCurrency(selectedPlanDetails.employer_cost)}</div>
            <div><strong>Monthly Deduction:</strong> {formatCurrency(selectedPlanDetails.employee_cost / 12)}</div>
            <div><strong>Total Annual Cost:</strong> {formatCurrency(selectedPlanDetails.employee_cost + selectedPlanDetails.employer_cost)}</div>
          </div>
        </div>

        {#if selectedPlanDetails.category === 'health'}
          <div class="detail-section">
            <h4>Coverage Details</h4>
            <div class="detail-grid">
              <div><strong>Deductible (Single):</strong> {formatCurrency(selectedPlanDetails.deductible_single)}</div>
              <div><strong>Deductible (Family):</strong> {formatCurrency(selectedPlanDetails.deductible_family)}</div>
              <div><strong>Out-of-Pocket Max (Single):</strong> {formatCurrency(selectedPlanDetails.out_of_pocket_max_single)}</div>
              <div><strong>Out-of-Pocket Max (Family):</strong> {formatCurrency(selectedPlanDetails.out_of_pocket_max_family)}</div>
            </div>
          </div>

          <div class="detail-section">
            <h4>Copays</h4>
            <div class="detail-grid">
              <div><strong>Primary Care:</strong> {formatCurrency(selectedPlanDetails.copay_primary_care)}</div>
              <div><strong>Specialist:</strong> {formatCurrency(selectedPlanDetails.copay_specialist)}</div>
              <div><strong>Emergency Room:</strong> {formatCurrency(selectedPlanDetails.copay_emergency)}</div>
              <div><strong>Coinsurance Rate:</strong> {selectedPlanDetails.coinsurance_rate}%</div>
            </div>
          </div>
        {/if}

        <div class="detail-section">
          <h4>Enrollment Period</h4>
          <div class="detail-grid">
            <div><strong>Enrollment Start:</strong> {formatDate(selectedPlanDetails.enrollment_start_date)}</div>
            <div><strong>Enrollment End:</strong> {formatDate(selectedPlanDetails.enrollment_end_date)}</div>
            <div><strong>Coverage Effective:</strong> {formatDate(selectedPlanDetails.effective_date)}</div>
          </div>
        </div>
      </div>

      <div class="modal-action">
        <button class="btn" on:click={() => showPlanDetails = false}>Close</button>
        <button class="btn btn-primary" on:click={() => { showPlanDetails = false; openEnrollModal(selectedPlanDetails); }}>
          Enroll in This Plan
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .benefits-container {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .benefits-header {
    margin-bottom: 2rem;
  }

  .benefits-header h1 {
    font-size: 2rem;
    font-weight: 700;
    margin-bottom: 0.5rem;
  }

  .text-muted {
    color: #6b7280;
  }

  .alert {
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .alert-error {
    background-color: #fef2f2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .alert-success {
    background-color: #f0fdf4;
    color: #166534;
    border: 1px solid #bbf7d0;
  }

  .alert button {
    background: none;
    border: none;
    font-size: 1.25rem;
    cursor: pointer;
    opacity: 0.7;
  }

  .alert button:hover {
    opacity: 1;
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
    cursor: pointer;
    font-weight: 500;
    color: #6b7280;
    border-bottom: 3px solid transparent;
    transition: all 0.2s;
  }

  .tab:hover {
    color: #111827;
  }

  .tab-active {
    color: #3b82f6;
    border-bottom-color: #3b82f6;
  }

  .loading-spinner {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem;
  }

  .spinner {
    width: 3rem;
    height: 3rem;
    border: 4px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .filter-bar {
    margin-bottom: 1.5rem;
  }

  .filter-bar select {
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    margin-left: 0.5rem;
  }

  .plans-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }

  .plan-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    overflow: hidden;
    transition: all 0.2s;
  }

  .plan-card:hover {
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }

  .plan-header {
    display: flex;
    gap: 1rem;
    padding: 1.5rem;
    background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
  }

  .plan-icon {
    font-size: 2.5rem;
  }

  .plan-title h3 {
    font-size: 1.25rem;
    font-weight: 700;
    margin-bottom: 0.25rem;
  }

  .plan-provider {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
  }

  .badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    margin-right: 0.5rem;
    margin-top: 0.25rem;
  }

  .badge-info {
    background-color: #dbeafe;
    color: #1e40af;
  }

  .badge-outline {
    border: 1px solid #d1d5db;
    color: #6b7280;
  }

  .badge-success {
    background-color: #d1fae5;
    color: #065f46;
  }

  .badge-warning {
    background-color: #fef3c7;
    color: #92400e;
  }

  .badge-error {
    background-color: #fee2e2;
    color: #991b1b;
  }

  .badge-ghost {
    background-color: #f3f4f6;
    color: #6b7280;
  }

  .plan-body {
    padding: 1.5rem;
  }

  .plan-description {
    color: #4b5563;
    font-size: 0.875rem;
    margin-bottom: 1rem;
    line-height: 1.5;
  }

  .plan-costs {
    background: #f9fafb;
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1rem;
  }

  .cost-item {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.5rem;
  }

  .cost-item:last-child {
    margin-bottom: 0;
  }

  .cost-item.total {
    padding-top: 0.5rem;
    border-top: 2px solid #e5e7eb;
    font-weight: 600;
  }

  .cost-label {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .cost-value {
    font-weight: 600;
    color: #111827;
  }

  .cost-value.employer {
    color: #059669;
  }

  .plan-details-quick {
    background: #fff;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    padding: 1rem;
  }

  .detail-row {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
  }

  .detail-row:not(:last-child) {
    border-bottom: 1px solid #f3f4f6;
  }

  .detail-row span {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .plan-footer {
    padding: 1rem 1.5rem;
    background: #f9fafb;
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
  }

  .btn {
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-weight: 500;
    cursor: pointer;
    border: 1px solid transparent;
    transition: all 0.2s;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background-color: #3b82f6;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: #2563eb;
  }

  .btn-outline {
    border-color: #d1d5db;
    color: #374151;
    background: white;
  }

  .btn-outline:hover:not(:disabled) {
    background-color: #f9fafb;
  }

  .btn-error {
    background-color: #ef4444;
    color: white;
  }

  .btn-error:hover:not(:disabled) {
    background-color: #dc2626;
  }

  .btn-sm {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
  }

  .btn-ghost {
    background: transparent;
    color: #6b7280;
  }

  .btn-ghost:hover {
    background-color: #f3f4f6;
  }

  .btn-xs {
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
  }

  .empty-state {
    text-align: center;
    padding: 4rem 2rem;
    color: #6b7280;
  }

  .empty-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
  }

  .empty-state h3 {
    font-size: 1.5rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
    color: #111827;
  }

  .enrollments-list {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .enrollment-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    overflow: hidden;
  }

  .enrollment-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 1.5rem;
    background: linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%);
  }

  .enrollment-header h3 {
    font-size: 1.25rem;
    font-weight: 700;
    margin-bottom: 0.25rem;
  }

  .enrollment-category {
    color: #6b7280;
    font-size: 0.875rem;
    text-transform: capitalize;
  }

  .enrollment-body {
    padding: 1.5rem;
  }

  .enrollment-info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .info-item {
    display: flex;
    flex-direction: column;
  }

  .info-label {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 0.25rem;
  }

  .info-value {
    font-weight: 600;
    color: #111827;
    text-transform: capitalize;
  }

  .dependents-section {
    background: #f9fafb;
    padding: 1rem;
    border-radius: 0.5rem;
    margin-top: 1rem;
  }

  .dependents-section h4 {
    font-size: 0.875rem;
    font-weight: 600;
    color: #6b7280;
    margin-bottom: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .dependents-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 0.75rem;
  }

  .dependent-card {
    display: flex;
    gap: 0.75rem;
    background: white;
    padding: 0.75rem;
    border-radius: 0.375rem;
    border: 1px solid #e5e7eb;
  }

  .dependent-icon {
    font-size: 1.5rem;
  }

  .dependent-info {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .dependent-info strong {
    font-size: 0.875rem;
  }

  .dependent-relationship,
  .dependent-dob {
    font-size: 0.75rem;
    color: #6b7280;
    text-transform: capitalize;
  }

  .enrollment-footer {
    padding: 1rem 1.5rem;
    background: #f9fafb;
    border-top: 1px solid #e5e7eb;
    display: flex;
    justify-content: flex-end;
  }

  .termination-notice {
    padding: 1rem;
    background: #fef3c7;
    color: #92400e;
    border-top: 1px solid #fde68a;
    font-size: 0.875rem;
  }

  .summary-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .summary-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
    display: flex;
    gap: 1rem;
    align-items: center;
  }

  .summary-icon {
    font-size: 2.5rem;
  }

  .summary-content h3 {
    font-size: 1.75rem;
    font-weight: 700;
    margin-bottom: 0.25rem;
  }

  .summary-content p {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .summary-breakdown {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.75rem;
    padding: 1.5rem;
  }

  .summary-breakdown h3 {
    font-size: 1.25rem;
    font-weight: 700;
    margin-bottom: 1.5rem;
  }

  .breakdown-table {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .breakdown-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    background: #f9fafb;
    border-radius: 0.5rem;
  }

  .breakdown-category {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .category-icon {
    font-size: 1.5rem;
  }

  .category-name {
    font-weight: 600;
    text-transform: capitalize;
  }

  .category-count {
    color: #6b7280;
    font-size: 0.875rem;
  }

  .breakdown-costs {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.25rem;
  }

  .breakdown-annual {
    font-weight: 600;
    color: #111827;
  }

  .breakdown-monthly {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .modal {
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-box {
    background: white;
    border-radius: 0.75rem;
    padding: 2rem;
    max-height: 90vh;
    overflow-y: auto;
    width: 90%;
  }

  .max-w-3xl {
    max-width: 48rem;
  }

  .max-w-2xl {
    max-width: 42rem;
  }

  .modal-action {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
    margin-top: 2rem;
  }

  .enrollment-form {
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

  .cost-calculation {
    background: #f0f9ff;
    border: 1px solid #bae6fd;
    border-radius: 0.5rem;
    padding: 1rem;
  }

  .cost-calculation h4 {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 1rem;
    color: #0c4a6e;
  }

  .cost-grid {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .cost-row {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
  }

  .cost-row.highlight {
    background: #075985;
    color: white;
    padding: 0.75rem;
    border-radius: 0.375rem;
    margin-top: 0.5rem;
  }

  .text-success {
    color: #059669;
  }

  .dependents-form {
    background: #f9fafb;
    padding: 1rem;
    border-radius: 0.5rem;
  }

  .dependents-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .dependents-header h4 {
    font-size: 1rem;
    font-weight: 600;
  }

  .dependent-form-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    padding: 1rem;
    margin-bottom: 1rem;
  }

  .dependent-form-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .dependent-form-header h5 {
    font-weight: 600;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .col-span-2 {
    grid-column: span 2;
  }

  .input-sm,
  .select-sm {
    padding: 0.375rem;
    font-size: 0.875rem;
  }

  .plan-details-full {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .detail-section {
    border-bottom: 1px solid #e5e7eb;
    padding-bottom: 1rem;
  }

  .detail-section:last-child {
    border-bottom: none;
  }

  .detail-section h4 {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 0.75rem;
    color: #374151;
  }

  .detail-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 0.75rem;
  }

  .detail-grid div {
    font-size: 0.875rem;
    color: #4b5563;
  }

  .detail-grid strong {
    color: #111827;
  }

  @media (max-width: 768px) {
    .benefits-container {
      padding: 1rem;
    }

    .plans-grid {
      grid-template-columns: 1fr;
    }

    .summary-cards {
      grid-template-columns: 1fr;
    }

    .form-grid {
      grid-template-columns: 1fr;
    }

    .col-span-2 {
      grid-column: span 1;
    }

    .enrollment-info-grid {
      grid-template-columns: 1fr;
    }

    .dependents-list {
      grid-template-columns: 1fr;
    }
  }
</style>