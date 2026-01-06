<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../../stores/auth';
  import { getApiBaseUrl } from '../../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();
  
  let activeTab: 'manager' | 'enrollment' = 'manager';
  
  // Benefits Manager State
  interface BenefitsProvider {
    id: string;
    name: string;
    type: 'healthcare' | 'retirement';
    status: 'active' | 'inactive';
    enrolled_employees: number;
    total_cost: number;
    employer_contribution: number;
    employee_contribution: number;
    provider_details: any;
  }
  
  interface HealthcarePlan {
    id: string;
    provider_id: string;
    provider_name: string;
    plan_name: string;
    plan_type: 'HMO' | 'PPO' | 'EPO' | 'HDHP';
    monthly_premium: number;
    deductible: number;
    out_of_pocket_max: number;
    copay_primary: number;
    copay_specialist: number;
    employer_subsidy_percentage: number;
    enrolled_count: number;
    coverage_tiers: {
      employee_only: number;
      employee_spouse: number;
      employee_children: number;
      family: number;
    };
  }
  
  interface RetirementPlan {
    id: string;
    provider_id: string;
    provider_name: string;
    plan_type: '401k' | '403b' | 'Roth401k' | 'SEP';
    employer_match_percentage: number;
    employer_match_cap: number;
    vesting_schedule: string;
    enrolled_count: number;
    total_assets: number;
    average_contribution_rate: number;
    available_funds: {
      ticker: string;
      name: string;
      asset_class: string;
      expense_ratio: number;
    }[];
  }
  
  let benefitsProviders: BenefitsProvider[] = [];
  let healthcarePlans: HealthcarePlan[] = [];
  let retirementPlans: RetirementPlan[] = [];
  let selectedProvider: BenefitsProvider | null = null;
  let showProviderModal = false;
  
  // Enrollment State
  interface Employee {
    id: string;
    name: string;
    email: string;
    department: string;
    hire_date: string;
    enrollment_status: 'pending' | 'in-progress' | 'completed';
  }
  
  interface EnrollmentSession {
    id: string;
    employee_id: string;
    employee_name: string;
    status: 'pending' | 'in-progress' | 'completed';
    healthcare_selected?: string;
    retirement_selected?: string;
    healthcare_contribution?: number;
    retirement_contribution_percentage?: number;
    ai_recommendations: {
      healthcare_plan_id?: string;
      retirement_allocation?: any;
      reasoning?: string;
    };
    started_at?: string;
    completed_at?: string;
  }
  
  let pendingEmployees: Employee[] = [];
  let enrollmentSessions: EnrollmentSession[] = [];
  let selectedEmployee: Employee | null = null;
  let selectedEnrollment: EnrollmentSession | null = null;
  let showEnrollmentModal = false;
  let showAIAssistModal = false;
  
  // AI Assistant State
  let aiQuery = '';
  let aiResponse = '';
  let aiLoading = false;
  let riskProfile: 'conservative' | 'moderate' | 'aggressive' = 'moderate';
  let preferredAssetClasses: string[] = [];
  
  let loading = true;
  
  onMount(async () => {
    await loadBenefitsData();
  });
  
  async function loadBenefitsData() {
    try {
      loading = true;
      const token = $authStore.token || localStorage.getItem('token');
      
      const [providersRes, healthcareRes, retirementRes, employeesRes, enrollmentsRes] = await Promise.all([
        fetch(`${API_BASE_URL}/benefits/providers`, {
          headers: { 'Authorization': `Bearer ${token}` }
        }),
        fetch(`${API_BASE_URL}/benefits/healthcare/plans`, {
          headers: { 'Authorization': `Bearer ${token}` }
        }),
        fetch(`${API_BASE_URL}/benefits/retirement/plans`, {
          headers: { 'Authorization': `Bearer ${token}` }
        }),
        fetch(`${API_BASE_URL}/benefits/enrollment/pending`, {
          headers: { 'Authorization': `Bearer ${token}` }
        }),
        fetch(`${API_BASE_URL}/benefits/enrollment/sessions`, {
          headers: { 'Authorization': `Bearer ${token}` }
        })
      ]);
      
      if (providersRes.ok) benefitsProviders = await providersRes.json();
      if (healthcareRes.ok) healthcarePlans = await healthcareRes.json();
      if (retirementRes.ok) retirementPlans = await retirementRes.json();
      if (employeesRes.ok) pendingEmployees = await employeesRes.json();
      if (enrollmentsRes.ok) enrollmentSessions = await enrollmentsRes.json();
      
    } catch (err) {
      console.error('Failed to load benefits data:', err);
    } finally {
      loading = false;
    }
  }
  
  function viewProvider(provider: BenefitsProvider) {
    selectedProvider = provider;
    showProviderModal = true;
  }
  
  function closeProviderModal() {
    showProviderModal = false;
    selectedProvider = null;
  }
  
  function startEnrollment(employee: Employee) {
    selectedEmployee = employee;
    selectedEnrollment = enrollmentSessions.find(e => e.employee_id === employee.id) || {
      id: '',
      employee_id: employee.id,
      employee_name: employee.name,
      status: 'pending',
      ai_recommendations: {}
    };
    showEnrollmentModal = true;
  }
  
  function closeEnrollmentModal() {
    showEnrollmentModal = false;
    selectedEmployee = null;
    selectedEnrollment = null;
  }
  
  async function getAIRecommendations() {
    if (!selectedEnrollment) return;
    
    try {
      aiLoading = true;
      const token = $authStore.token || localStorage.getItem('token');
      
      const response = await fetch(`${API_BASE_URL}/benefits/ai/recommend`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          employee_id: selectedEnrollment.employee_id,
          risk_profile: riskProfile,
          preferred_asset_classes: preferredAssetClasses,
          query: aiQuery
        })
      });
      
      if (response.ok) {
        const data = await response.json();
        selectedEnrollment.ai_recommendations = data.recommendations;
        aiResponse = data.explanation || '';
      }
    } catch (err) {
      console.error('Failed to get AI recommendations:', err);
      aiResponse = 'Failed to get recommendations. Please try again.';
    } finally {
      aiLoading = false;
    }
  }
  
  async function saveEnrollment() {
    if (!selectedEnrollment) return;
    
    try {
      const token = $authStore.token || localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/benefits/enrollment`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(selectedEnrollment)
      });
      
      if (response.ok) {
        await loadBenefitsData();
        closeEnrollmentModal();
        alert('Enrollment saved successfully!');
      }
    } catch (err) {
      console.error('Failed to save enrollment:', err);
      alert('Failed to save enrollment: ' + err.message);
    }
  }
  
  function getStatusColor(status: string): string {
    const colors = {
      'active': 'bg-green-100 text-green-800',
      'inactive': 'bg-gray-100 text-gray-800',
      'pending': 'bg-yellow-100 text-yellow-800',
      'in-progress': 'bg-blue-100 text-blue-800',
      'completed': 'bg-green-100 text-green-800'
    };
    return colors[status] || 'bg-gray-100 text-gray-800';
  }
  
  $: healthcareProviders = benefitsProviders.filter(p => p.type === 'healthcare');
  $: retirementProviders = benefitsProviders.filter(p => p.type === 'retirement');
  $: totalEmployerCosts = benefitsProviders.reduce((sum, p) => sum + p.employer_contribution, 0);
</script>

<div class="benefits-dashboard">
  <!-- Header -->
  <div class="header">
    <div>
      <h1>🏥 Benefits Dashboard</h1>
      <p class="subtitle">Manage healthcare and retirement benefits with AI-powered enrollment assistance</p>
    </div>
  </div>
  
  <!-- Tab Navigation -->
  <div class="tabs">
    <button 
      class="tab"
      class:active={activeTab === 'manager'}
      on:click={() => activeTab = 'manager'}
    >
      <span class="tab-icon">⚙️</span>
      <span class="tab-text">Benefits Manager</span>
    </button>
    
    <button 
      class="tab"
      class:active={activeTab === 'enrollment'}
      on:click={() => activeTab = 'enrollment'}
    >
      <span class="tab-icon">👥</span>
      <span class="tab-text">Employee Enrollment</span>
      {#if pendingEmployees.length > 0}
        <span class="tab-badge">{pendingEmployees.length}</span>
      {/if}
    </button>
  </div>
  
  {#if loading}
    <div class="loading">Loading benefits data...</div>
  {:else}
    
    {#if activeTab === 'manager'}
      <!-- Benefits Manager Tab -->
      <div class="tab-content">
        <!-- Summary Stats -->
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-icon">🏥</div>
            <div class="stat-content">
              <div class="stat-value">{healthcareProviders.length}</div>
              <div class="stat-label">Healthcare Providers</div>
              <div class="stat-sublabel">{healthcarePlans.length} plans available</div>
            </div>
          </div>
          
          <div class="stat-card">
            <div class="stat-icon">💰</div>
            <div class="stat-content">
              <div class="stat-value">{retirementProviders.length}</div>
              <div class="stat-label">Retirement Plans</div>
              <div class="stat-sublabel">{retirementPlans.reduce((sum, p) => sum + p.available_funds.length, 0)} investment options</div>
            </div>
          </div>
          
          <div class="stat-card highlight">
            <div class="stat-icon">💵</div>
            <div class="stat-content">
              <div class="stat-value">${totalEmployerCosts.toLocaleString()}</div>
              <div class="stat-label">Monthly Employer Cost</div>
              <div class="stat-sublabel">Total contributions & subsidies</div>
            </div>
          </div>
        </div>
        
        <!-- Healthcare Section -->
        <div class="section-card">
          <div class="section-header">
            <h2>🏥 Healthcare Plans</h2>
            <button class="btn-primary">+ Add Plan</button>
          </div>
          
          {#if healthcarePlans.length === 0}
            <div class="empty-state">
              <span class="empty-icon">🏥</span>
              <p>No healthcare plans configured</p>
            </div>
          {:else}
            <div class="plans-grid">
              {#each healthcarePlans as plan}
                <div class="plan-card" on:click={() => viewProvider(benefitsProviders.find(p => p.id === plan.provider_id))}>
                  <div class="plan-header">
                    <div>
                      <h3 class="plan-name">{plan.plan_name}</h3>
                      <p class="plan-provider">{plan.provider_name}</p>
                    </div>
                    <span class="plan-type-badge">{plan.plan_type}</span>
                  </div>
                  
                  <div class="plan-details">
                    <div class="detail-row">
                      <span class="detail-label">Monthly Premium</span>
                      <span class="detail-value">${plan.monthly_premium.toLocaleString()}</span>
                    </div>
                    <div class="detail-row">
                      <span class="detail-label">Deductible</span>
                      <span class="detail-value">${plan.deductible.toLocaleString()}</span>
                    </div>
                    <div class="detail-row">
                      <span class="detail-label">Out-of-Pocket Max</span>
                      <span class="detail-value">${plan.out_of_pocket_max.toLocaleString()}</span>
                    </div>
                  </div>
                  
                  <div class="plan-subsidy">
                    <div class="subsidy-bar">
                      <div class="subsidy-fill" style="width: {plan.employer_subsidy_percentage}%"></div>
                    </div>
                    <span class="subsidy-text">
                      {plan.employer_subsidy_percentage}% Employer Subsidy
                    </span>
                  </div>
                  
                  <div class="plan-footer">
                    <span class="enrolled-count">
                      <span class="count-icon">👥</span>
                      {plan.enrolled_count} enrolled
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
        
        <!-- Retirement Section -->
        <div class="section-card">
          <div class="section-header">
            <h2>💰 Retirement Plans</h2>
            <button class="btn-primary">+ Add Plan</button>
          </div>
          
          {#if retirementPlans.length === 0}
            <div class="empty-state">
              <span class="empty-icon">💰</span>
              <p>No retirement plans configured</p>
            </div>
          {:else}
            <div class="retirement-list">
              {#each retirementPlans as plan}
                <div class="retirement-card" on:click={() => viewProvider(benefitsProviders.find(p => p.id === plan.provider_id))}>
                  <div class="retirement-header">
                    <div class="retirement-info">
                      <h3 class="retirement-name">{plan.plan_type}</h3>
                      <p class="retirement-provider">{plan.provider_name}</p>
                    </div>
                    <div class="retirement-stats">
                      <div class="stat-item">
                        <span class="stat-label">Total Assets</span>
                        <span class="stat-value">${(plan.total_assets / 1000000).toFixed(1)}M</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-label">Enrolled</span>
                        <span class="stat-value">{plan.enrolled_count}</span>
                      </div>
                    </div>
                  </div>
                  
                  <div class="retirement-body">
                    <div class="match-info">
                      <span class="match-icon">🎁</span>
                      <div class="match-details">
                        <strong>Employer Match:</strong> {plan.employer_match_percentage}% up to {plan.employer_match_cap}% of salary
                      </div>
                    </div>
                    
                    <div class="contribution-rate">
                      <span class="rate-label">Average Contribution Rate:</span>
                      <span class="rate-value">{plan.average_contribution_rate}%</span>
                      <div class="rate-bar">
                        <div class="rate-fill" style="width: {(plan.average_contribution_rate / 15) * 100}%"></div>
                      </div>
                    </div>
                    
                    <div class="funds-info">
                      <span class="funds-label">
                        <span class="funds-icon">📊</span>
                        {plan.available_funds.length} Investment Options
                      </span>
                      <div class="asset-classes">
                        {#each [...new Set(plan.available_funds.map(f => f.asset_class))] as assetClass}
                          <span class="asset-class-badge">{assetClass}</span>
                        {/each}
                      </div>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
      
    {:else}
      <!-- Employee Enrollment Tab -->
      <div class="tab-content">
        <!-- AI Assistant Banner -->
        <div class="ai-banner">
          <div class="ai-icon">🤖</div>
          <div class="ai-content">
            <h3>AI Benefits Assistant</h3>
            <p>Get personalized recommendations for healthcare plans and retirement allocations based on employee profiles, risk tolerance, and financial goals.</p>
          </div>
        </div>
        
        <!-- Pending Enrollments -->
        <div class="section-card">
          <div class="section-header">
            <h2>Pending Enrollments</h2>
            <span class="count-badge">{pendingEmployees.length} pending</span>
          </div>
          
          {#if pendingEmployees.length === 0}
            <div class="empty-state">
              <span class="empty-icon">✓</span>
              <p>All employees are enrolled in benefits</p>
            </div>
          {:else}
            <div class="enrollment-list">
              {#each pendingEmployees as employee}
                <div class="enrollment-card">
                  <div class="employee-info">
                    <div class="employee-avatar">
                      {employee.name.split(' ').map(n => n[0]).join('')}
                    </div>
                    <div class="employee-details">
                      <h3 class="employee-name">{employee.name}</h3>
                      <p class="employee-meta">
                        <span>{employee.department}</span>
                        <span class="divider">•</span>
                        <span>Hired: {new Date(employee.hire_date).toLocaleDateString()}</span>
                      </p>
                    </div>
                  </div>
                  
                  <span class="status-badge {getStatusColor(employee.enrollment_status)}">
                    {employee.enrollment_status}
                  </span>
                  
                  <button class="btn-primary" on:click={() => startEnrollment(employee)}>
                    {employee.enrollment_status === 'pending' ? 'Start Enrollment' : 'Continue Enrollment'}
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
        
        <!-- Recent Enrollments -->
        {#if enrollmentSessions.filter(e => e.status === 'completed').length > 0}
          <div class="section-card">
            <div class="section-header">
              <h2>Recently Completed</h2>
            </div>
            
            <div class="completed-list">
              {#each enrollmentSessions.filter(e => e.status === 'completed').slice(0, 5) as session}
                <div class="completed-item">
                  <span class="completed-icon">✓</span>
                  <div class="completed-info">
                    <strong>{session.employee_name}</strong>
                    <span class="completed-date">
                      Completed {session.completed_at ? new Date(session.completed_at).toLocaleDateString() : 'Recently'}
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}
    
  {/if}
</div>

<!-- Provider Detail Modal -->
{#if showProviderModal && selectedProvider}
  <div class="modal-overlay" on:click={closeProviderModal}>
    <div class="modal large" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{selectedProvider.name}</h2>
          <p class="modal-subtitle">{selectedProvider.type === 'healthcare' ? 'Healthcare Provider' : 'Retirement Plan Provider'}</p>
        </div>
        <button class="close-btn" on:click={closeProviderModal}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="provider-stats-grid">
          <div class="provider-stat">
            <span class="provider-stat-label">Enrolled Employees</span>
            <span class="provider-stat-value">{selectedProvider.enrolled_employees}</span>
          </div>
          <div class="provider-stat">
            <span class="provider-stat-label">Total Monthly Cost</span>
            <span class="provider-stat-value">${selectedProvider.total_cost.toLocaleString()}</span>
          </div>
          <div class="provider-stat">
            <span class="provider-stat-label">Employer Contribution</span>
            <span class="provider-stat-value">${selectedProvider.employer_contribution.toLocaleString()}</span>
          </div>
          <div class="provider-stat">
            <span class="provider-stat-label">Employee Contribution</span>
            <span class="provider-stat-value">${selectedProvider.employee_contribution.toLocaleString()}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Enrollment Modal -->
{#if showEnrollmentModal && selectedEmployee && selectedEnrollment}
  <div class="modal-overlay" on:click={closeEnrollmentModal}>
    <div class="modal xlarge" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>Benefits Enrollment: {selectedEmployee.name}</h2>
          <p class="modal-subtitle">{selectedEmployee.department}</p>
        </div>
        <button class="close-btn" on:click={closeEnrollmentModal}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="enrollment-wizard">
          <!-- AI Assistant Section -->
          <div class="wizard-section">
            <div class="section-title">
              <span class="title-icon">🤖</span>
              <h3>AI Benefits Assistant</h3>
            </div>
            
            <div class="ai-controls">
              <div class="control-group">
                <label>Risk Profile</label>
                <select bind:value={riskProfile} class="control-select">
                  <option value="conservative">Conservative</option>
                  <option value="moderate">Moderate</option>
                  <option value="aggressive">Aggressive</option>
                </select>
              </div>
              
              <div class="control-group">
                <label>Preferred Asset Classes</label>
                <div class="checkbox-group">
                  <label class="checkbox-label">
                    <input type="checkbox" value="stocks" bind:group={preferredAssetClasses} />
                    <span>Stocks</span>
                  </label>
                  <label class="checkbox-label">
                    <input type="checkbox" value="bonds" bind:group={preferredAssetClasses} />
                    <span>Bonds</span>
                  </label>
                  <label class="checkbox-label">
                    <input type="checkbox" value="international" bind:group={preferredAssetClasses} />
                    <span>International</span>
                  </label>
                  <label class="checkbox-label">
                    <input type="checkbox" value="real-estate" bind:group={preferredAssetClasses} />
                    <span>Real Estate</span>
                  </label>
                </div>
              </div>
              
              <div class="control-group">
                <label>Ask AI Assistant</label>
                <textarea 
                  bind:value={aiQuery} 
                  placeholder="e.g., 'Which healthcare plan is best for a family of 4?' or 'How should I allocate retirement funds for long-term growth?'"
                  rows="3"
                  class="ai-query-input"
                ></textarea>
              </div>
              
              <button 
                class="btn-ai" 
                on:click={getAIRecommendations}
                disabled={aiLoading}
              >
                {aiLoading ? '🔄 Getting Recommendations...' : '✨ Get AI Recommendations'}
              </button>
              
              {#if aiResponse}
                <div class="ai-response">
                  <div class="response-header">
                    <span class="response-icon">💡</span>
                    <strong>AI Recommendation:</strong>
                  </div>
                  <p class="response-text">{aiResponse}</p>
                </div>
              {/if}
            </div>
          </div>
          
          <!-- Healthcare Selection -->
          <div class="wizard-section">
            <div class="section-title">
              <span class="title-icon">🏥</span>
              <h3>Healthcare Plan Selection</h3>
            </div>
            
            <div class="plan-selection">
              {#each healthcarePlans as plan}
                <label class="selection-card" class:selected={selectedEnrollment.healthcare_selected === plan.id}>
                  <input 
                    type="radio" 
                    name="healthcare"
                    value={plan.id}
                    bind:group={selectedEnrollment.healthcare_selected}
                  />
                  <div class="selection-content">
                    <div class="selection-header">
                      <strong>{plan.plan_name}</strong>
                      <span class="plan-type-badge small">{plan.plan_type}</span>
                    </div>
                    <div class="selection-details">
                      <span>Premium: ${plan.monthly_premium}/mo</span>
                      <span>Deductible: ${plan.deductible.toLocaleString()}</span>
                      <span>Employer pays: {plan.employer_subsidy_percentage}%</span>
                    </div>
                    {#if selectedEnrollment.ai_recommendations.healthcare_plan_id === plan.id}
                      <div class="ai-recommended">
                        ✨ AI Recommended
                      </div>
                    {/if}
                  </div>
                </label>
              {/each}
            </div>
          </div>
          
          <!-- Retirement Selection -->
          <div class="wizard-section">
            <div class="section-title">
              <span class="title-icon">💰</span>
              <h3>Retirement Plan & Contribution</h3>
            </div>
            
            <div class="retirement-selection">
              {#each retirementPlans as plan}
                <label class="selection-card" class:selected={selectedEnrollment.retirement_selected === plan.id}>
                  <input 
                    type="radio" 
                    name="retirement"
                    value={plan.id}
                    bind:group={selectedEnrollment.retirement_selected}
                  />
                  <div class="selection-content">
                    <div class="selection-header">
                      <strong>{plan.plan_type}</strong>
                    </div>
                    <div class="selection-details">
                      <span>Match: {plan.employer_match_percentage}% up to {plan.employer_match_cap}%</span>
                      <span>Vesting: {plan.vesting_schedule}</span>
                    </div>
                  </div>
                </label>
              {/each}
              
              {#if selectedEnrollment.retirement_selected}
                <div class="contribution-selector">
                  <label>Contribution Percentage</label>
                  <div class="slider-container">
                    <input 
                      type="range" 
                      min="0" 
                      max="25" 
                      step="0.5"
                      bind:value={selectedEnrollment.retirement_contribution_percentage}
                      class="contribution-slider"
                    />
                    <span class="slider-value">{selectedEnrollment.retirement_contribution_percentage || 0}%</span>
                  </div>
                  {#if selectedEnrollment.retirement_contribution_percentage}
                    {@const selectedPlan = retirementPlans.find(p => p.id === selectedEnrollment.retirement_selected)}
                    {#if selectedPlan}
                      <div class="match-display">
                        {#if selectedEnrollment.retirement_contribution_percentage >= selectedPlan.employer_match_cap}
                          <span class="match-full">✓ Receiving full employer match ({selectedPlan.employer_match_percentage}%)</span>
                        {:else}
                          <span class="match-partial">⚠️ Contributing below match cap - consider increasing to {selectedPlan.employer_match_cap}%</span>
                        {/if}
                      </div>
                    {/if}
                  {/if}
                </div>
              {/if}
            </div>
          </div>
          
          <!-- Asset Allocation (if AI recommendations available) -->
          {#if selectedEnrollment.ai_recommendations.retirement_allocation}
            <div class="wizard-section">
              <div class="section-title">
                <span class="title-icon">📊</span>
                <h3>Recommended Asset Allocation</h3>
              </div>
              
              <div class="allocation-display">
                <p class="allocation-note">Based on {riskProfile} risk profile</p>
                {#each Object.entries(selectedEnrollment.ai_recommendations.retirement_allocation) as [assetClass, percentage]}
                  <div class="allocation-row">
                    <span class="allocation-label">{assetClass}</span>
                    <div class="allocation-bar">
                      <div class="allocation-fill" style="width: {percentage}%"></div>
                    </div>
                    <span class="allocation-percentage">{percentage}%</span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
        
        <div class="modal-actions">
          <button class="btn secondary" on:click={closeEnrollmentModal}>
            Cancel
          </button>
          <button 
            class="btn primary" 
            on:click={saveEnrollment}
            disabled={!selectedEnrollment.healthcare_selected || !selectedEnrollment.retirement_selected}
          >
            <span class="btn-icon">💾</span>
            Save Enrollment
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .benefits-dashboard {
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
    background: #10b981;
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
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
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
  
  .stat-card.highlight {
    background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
    border-left: 4px solid #10b981;
  }
  
  .stat-icon {
    font-size: 36px;
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
    font-weight: 500;
    color: #111827;
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
    background: #10b981;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
  }
  
  .btn-primary:hover {
    background: #059669;
  }
  
  .btn-primary:disabled {
    background: #9ca3af;
    cursor: not-allowed;
  }
  
  .plans-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 16px;
    padding: 24px;
  }
  
  .plan-card {
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .plan-card:hover {
    border-color: #10b981;
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.1);
    transform: translateY(-2px);
  }
  
  .plan-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 16px;
  }
  
  .plan-name {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .plan-provider {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .plan-type-badge {
    padding: 4px 12px;
    background: #dbeafe;
    color: #1e40af;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }
  
  .plan-type-badge.small {
    padding: 2px 8px;
    font-size: 10px;
  }
  
  .plan-details {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 16px;
  }
  
  .detail-row {
    display: flex;
    justify-content: space-between;
    font-size: 13px;
  }
  
  .detail-label {
    color: #6b7280;
  }
  
  .detail-value {
    color: #111827;
    font-weight: 500;
  }
  
  .plan-subsidy {
    margin-bottom: 12px;
  }
  
  .subsidy-bar {
    width: 100%;
    height: 6px;
    background: #e5e7eb;
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 4px;
  }
  
  .subsidy-fill {
    height: 100%;
    background: #10b981;
    border-radius: 3px;
  }
  
  .subsidy-text {
    font-size: 12px;
    color: #059669;
    font-weight: 500;
  }
  
  .plan-footer {
    padding-top: 12px;
    border-top: 1px solid #e5e7eb;
  }
  
  .enrolled-count {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    color: #6b7280;
  }
  
  .count-icon {
    font-size: 14px;
  }
  
  .retirement-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 24px;
  }
  
  .retirement-card {
    padding: 24px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .retirement-card:hover {
    border-color: #10b981;
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.1);
  }
  
  .retirement-header {
    display: flex;
    justify-content: space-between;
    align-items: start;
    margin-bottom: 20px;
  }
  
  .retirement-name {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .retirement-provider {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .retirement-stats {
    display: flex;
    gap: 24px;
  }
  
  .stat-item {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
  }
  
  .stat-item .stat-label {
    font-size: 12px;
    color: #6b7280;
  }
  
  .stat-item .stat-value {
    font-size: 20px;
    font-weight: 700;
    color: #111827;
  }
  
  .retirement-body {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  
  .match-info {
    display: flex;
    align-items: start;
    gap: 12px;
    padding: 16px;
    background: #f0fdf4;
    border-left: 4px solid #10b981;
    border-radius: 8px;
  }
  
  .match-icon {
    font-size: 24px;
  }
  
  .match-details {
    font-size: 14px;
    color: #065f46;
  }
  
  .contribution-rate {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  
  .rate-label {
    font-size: 13px;
    color: #6b7280;
  }
  
  .rate-value {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
  }
  
  .rate-bar {
    width: 100%;
    height: 6px;
    background: #e5e7eb;
    border-radius: 3px;
    overflow: hidden;
  }
  
  .rate-fill {
    height: 100%;
    background: #10b981;
    border-radius: 3px;
  }
  
  .funds-info {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  
  .funds-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;
    color: #374151;
  }
  
  .funds-icon {
    font-size: 16px;
  }
  
  .asset-classes {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  
  .asset-class-badge {
    padding: 4px 10px;
    background: #ede9fe;
    color: #5b21b6;
    border-radius: 6px;
    font-size: 11px;
    font-weight: 500;
  }
  
  .ai-banner {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 24px;
    background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
    border-left: 4px solid #3b82f6;
    border-radius: 12px;
  }
  
  .ai-icon {
    font-size: 48px;
  }
  
  .ai-content h3 {
    font-size: 18px;
    font-weight: 600;
    color: #1e40af;
    margin: 0 0 8px 0;
  }
  
  .ai-content p {
    font-size: 14px;
    color: #1e3a8a;
    margin: 0;
    line-height: 1.5;
  }
  
  .count-badge {
    padding: 4px 12px;
    background: #fef3c7;
    color: #92400e;
    border-radius: 12px;
    font-size: 13px;
    font-weight: 600;
  }
  
  .enrollment-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 24px;
  }
  
  .enrollment-card {
    display: flex;
    align-items: center;
    gap: 20px;
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    transition: all 0.2s;
  }
  
  .enrollment-card:hover {
    border-color: #10b981;
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.1);
  }
  
  .employee-info {
    display: flex;
    align-items: center;
    gap: 16px;
    flex: 1;
  }
  
  .employee-avatar {
    width: 48px;
    height: 48px;
    background: #10b981;
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-weight: 600;
  }
  
  .employee-name {
    font-size: 16px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .employee-meta {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
    display: flex;
    gap: 8px;
  }
  
  .divider {
    color: #d1d5db;
  }
  
  .status-badge {
    padding: 6px 12px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
  }
  
  .bg-green-100 { background: #d1fae5; }
  .text-green-800 { color: #065f46; }
  .bg-yellow-100 { background: #fef3c7; }
  .text-yellow-800 { color: #92400e; }
  .bg-blue-100 { background: #dbeafe; }
  .text-blue-800 { color: #1e40af; }
  .bg-gray-100 { background: #f3f4f6; }
  .text-gray-800 { color: #1f2937; }
  
  .completed-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 24px;
  }
  
  .completed-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: #f9fafb;
    border-radius: 8px;
  }
  
  .completed-icon {
    width: 24px;
    height: 24px;
    background: #10b981;
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
  }
  
  .completed-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  
  .completed-date {
    font-size: 12px;
    color: #6b7280;
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
    max-width: 700px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
  }
  
  .modal.large {
    max-width: 900px;
  }
  
  .modal.xlarge {
    max-width: 1200px;
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
  
  .provider-stats-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }
  
  .provider-stat {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 16px;
    background: #f9fafb;
    border-radius: 8px;
  }
  
  .provider-stat-label {
    font-size: 13px;
    color: #6b7280;
  }
  
  .provider-stat-value {
    font-size: 24px;
    font-weight: 700;
    color: #111827;
  }
  
  .enrollment-wizard {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .wizard-section {
    padding: 20px;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
  }
  
  .section-title {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;
  }
  
  .title-icon {
    font-size: 24px;
  }
  
  .section-title h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }
  
  .ai-controls {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  
  .control-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  
  .control-group label {
    font-size: 14px;
    font-weight: 500;
    color: #374151;
  }
  
  .control-select {
    padding: 10px 12px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
  }
  
  .checkbox-group {
    display: flex;
    gap: 16px;
    flex-wrap: wrap;
  }
  
  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }
  
  .checkbox-label input[type="checkbox"] {
    width: 18px;
    height: 18px;
    cursor: pointer;
  }
  
  .ai-query-input {
    width: 100%;
    padding: 12px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    font-family: inherit;
    resize: vertical;
  }
  
  .ai-query-input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }
  
  .btn-ai {
    padding: 12px 24px;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-ai:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  }
  
  .btn-ai:disabled {
    background: #9ca3af;
    cursor: not-allowed;
    transform: none;
  }
  
  .ai-response {
    padding: 16px;
    background: #eff6ff;
    border-left: 4px solid #3b82f6;
    border-radius: 8px;
  }
  
  .response-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    color: #1e40af;
    font-weight: 600;
  }
  
  .response-icon {
    font-size: 20px;
  }
  
  .response-text {
    font-size: 14px;
    color: #1e3a8a;
    line-height: 1.6;
    margin: 0;
  }
  
  .plan-selection,
  .retirement-selection {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .selection-card {
    position: relative;
    padding: 16px;
    border: 2px solid #e5e7eb;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: start;
    gap: 12px;
  }
  
  .selection-card:hover {
    border-color: #10b981;
    background: #f0fdf4;
  }
  
  .selection-card.selected {
    border-color: #10b981;
    background: #d1fae5;
  }
  
  .selection-card input[type="radio"] {
    width: 20px;
    height: 20px;
    cursor: pointer;
    margin-top: 2px;
  }
  
  .selection-content {
    flex: 1;
  }
  
  .selection-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }
  
  .selection-details {
    display: flex;
    gap: 16px;
    flex-wrap: wrap;
    font-size: 13px;
    color: #6b7280;
  }
  
  .ai-recommended {
    margin-top: 8px;
    padding: 4px 12px;
    background: #dbeafe;
    color: #1e40af;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 600;
    display: inline-block;
  }
  
  .contribution-selector {
    margin-top: 16px;
    padding: 16px;
    background: #f9fafb;
    border-radius: 8px;
  }
  
  .contribution-selector label {
    display: block;
    font-size: 14px;
    font-weight: 500;
    color: #374151;
    margin-bottom: 12px;
  }
  
  .slider-container {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  
  .contribution-slider {
    flex: 1;
    height: 8px;
    -webkit-appearance: none;
    appearance: none;
    background: #e5e7eb;
    border-radius: 4px;
    outline: none;
  }
  
  .contribution-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 20px;
    height: 20px;
    background: #10b981;
    border-radius: 50%;
    cursor: pointer;
  }
  
  .contribution-slider::-moz-range-thumb {
    width: 20px;
    height: 20px;
    background: #10b981;
    border-radius: 50%;
    cursor: pointer;
    border: none;
  }
  
  .slider-value {
    font-size: 18px;
    font-weight: 700;
    color: #10b981;
    min-width: 50px;
  }
  
  .match-display {
    margin-top: 12px;
    font-size: 13px;
  }
  
  .match-full {
    color: #059669;
    font-weight: 500;
  }
  
  .match-partial {
    color: #f59e0b;
    font-weight: 500;
  }
  
  .allocation-display {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .allocation-note {
    font-size: 13px;
    color: #6b7280;
    font-style: italic;
    margin: 0;
  }
  
  .allocation-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  
  .allocation-label {
    font-size: 14px;
    font-weight: 500;
    color: #374151;
    min-width: 120px;
  }
  
  .allocation-bar {
    flex: 1;
    height: 24px;
    background: #e5e7eb;
    border-radius: 4px;
    overflow: hidden;
  }
  
  .allocation-fill {
    height: 100%;
    background: linear-gradient(90deg, #10b981 0%, #059669 100%);
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding-right: 8px;
    color: white;
    font-size: 12px;
    font-weight: 600;
  }
  
  .allocation-percentage {
    font-size: 14px;
    font-weight: 600;
    color: #111827;
    min-width: 50px;
  }
  
  .modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid #e5e7eb;
  }
  
  .btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn.primary {
    background: #10b981;
    color: white;
  }
  
  .btn.primary:hover {
    background: #059669;
  }
  
  .btn.secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }
  
  .btn.secondary:hover {
    background: #f9fafb;
  }
  
  .btn-icon {
    font-size: 16px;
  }
  
  @media (max-width: 1024px) {
    .plans-grid {
      grid-template-columns: 1fr;
    }
    
    .provider-stats-grid {
      grid-template-columns: 1fr;
    }
  }
  
  @media (max-width: 768px) {
    .stats-grid {
      grid-template-columns: 1fr;
    }
    
    .retirement-stats {
      flex-direction: column;
      gap: 12px;
    }
    
    .stat-item {
      align-items: flex-start;
    }
    
    .tabs {
      flex-direction: column;
    }
    
    .ai-banner {
      flex-direction: column;
    }
    
    .modal {
      max-width: 100%;
      max-height: 100vh;
      border-radius: 0;
    }
  }
</style>