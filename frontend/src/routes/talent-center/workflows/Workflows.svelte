<script lang="ts">
  import { onMount } from 'svelte';

  export let navigate: (page: string, id?: string) => void;

  interface Workflow {
    id: string;
    employee_id: string;
    template_name: string;
    status: string;
    current_stage: string;
    progress_percentage: number;
    started_at: string;
    expected_completion: string;
    actual_completion?: string;
    created_at: string;
  }

  let workflows: Workflow[] = [];
  let loading = true;
  let error = '';
  let statusFilter = 'all';
  let searchTerm = '';
  let showCreateModal = false;

  // For creating new workflow
  let employees: any[] = [];
  let selectedEmployeeId = '';
  let selectedTemplate = 'software-engineer';

  const templates = [
    { value: 'software-engineer', label: 'Software Engineer (17 steps)' },
    { value: 'generic', label: 'Generic Employee (4 steps)' },
    { value: 'sales-representative', label: 'Sales Representative' },
    { value: 'manager', label: 'Manager' }
  ];

  onMount(() => {
    loadWorkflows();
    loadEmployees();
  });

  async function loadWorkflows() {
    loading = true;
    error = '';
    
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        error = 'Not authenticated. Please login again.';
        loading = false;
        return;
      }

      let url = `/api/workflows`;
      
      if (statusFilter !== 'all') {
        url += `?status=${statusFilter}`;
      }
      
      console.log('Fetching workflows from:', url);
      
      const response = await fetch(url, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      console.log('Response status:', response.status);

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Error response:', errorText);
        throw new Error(`Failed to load workflows: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.json();
      console.log('Workflows data:', data);
      workflows = Array.isArray(data) ? data : [];
    } catch (err: any) {
      console.error('Load workflows error:', err);
      error = err.message || 'Failed to load workflows';
      workflows = [];
    } finally {
      loading = false;
    }
  }

  async function loadEmployees() {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/employees`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!response.ok) throw new Error('Failed to load employees');
      
      const data = await response.json();
      employees = (data || []).filter((emp: any) => 
        emp.status === 'active' && emp.email !== 'admin@cocomgroup.com'
      );
    } catch (err) {
      console.error('Failed to load employees:', err);
    }
  }

  async function createWorkflow() {
    if (!selectedEmployeeId) {
      error = 'Please select an employee';
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/workflows`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          employee_id: selectedEmployeeId,
          template_name: selectedTemplate
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create workflow');
      }

      showCreateModal = false;
      selectedEmployeeId = '';
      selectedTemplate = 'software-engineer';
      await loadWorkflows();
    } catch (err: any) {
      error = err.message || 'Failed to create workflow';
    }
  }

  function getStageColor(stage: string): string {
    const colors: Record<string, string> = {
      'pre-boarding': 'bg-purple-100 text-purple-800',
      'day-1': 'bg-blue-100 text-blue-800',
      'week-1': 'bg-green-100 text-green-800',
      'month-1': 'bg-orange-100 text-orange-800',
      'completed': 'bg-gray-100 text-gray-800'
    };
    return colors[stage] || 'bg-gray-100 text-gray-800';
  }

  function getStatusColor(status: string): string {
    const colors: Record<string, string> = {
      'active': 'bg-green-100 text-green-800',
      'completed': 'bg-blue-100 text-blue-800',
      'cancelled': 'bg-red-100 text-red-800',
      'on-hold': 'bg-yellow-100 text-yellow-800'
    };
    return colors[status] || 'bg-gray-100 text-gray-800';
  }

  function formatDate(dateString: string): string {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  $: filteredWorkflows = workflows.filter(workflow => {
    const matchesSearch = searchTerm === '' || 
      workflow.template_name.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesSearch;
  });

  $: activeCount = workflows.filter(w => w.status === 'active').length;
  $: completedCount = workflows.filter(w => w.status === 'completed').length;
  $: onHoldCount = workflows.filter(w => w.status === 'on-hold').length;
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
  <!-- Header -->
  <div class="mb-8">
    <h1 class="text-3xl font-bold text-gray-900">Onboarding Workflows</h1>
    <p class="mt-2 text-sm text-gray-600">
      Manage and monitor employee onboarding workflows with intelligent automation
    </p>
  </div>

  <!-- Stats Cards -->
  <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center">
        <div class="flex-1">
          <p class="text-sm font-medium text-gray-600">Total Workflows</p>
          <p class="text-3xl font-bold text-gray-900">{workflows.length}</p>
        </div>
        <div style="width: 48px; height: 48px; min-width: 48px; min-height: 48px;" class="bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
          <svg style="width: 24px; height: 24px;" class="text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center">
        <div class="flex-1">
          <p class="text-sm font-medium text-gray-600">Active</p>
          <p class="text-3xl font-bold text-green-600">{activeCount}</p>
        </div>
        <div style="width: 48px; height: 48px; min-width: 48px; min-height: 48px;" class="bg-green-100 rounded-full flex items-center justify-center flex-shrink-0">
          <svg style="width: 24px; height: 24px;" class="text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center">
        <div class="flex-1">
          <p class="text-sm font-medium text-gray-600">Completed</p>
          <p class="text-3xl font-bold text-blue-600">{completedCount}</p>
        </div>
        <div style="width: 48px; height: 48px; min-width: 48px; min-height: 48px;" class="bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
          <svg style="width: 24px; height: 24px;" class="text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center">
        <div class="flex-1">
          <p class="text-sm font-medium text-gray-600">On Hold</p>
          <p class="text-3xl font-bold text-yellow-600">{onHoldCount}</p>
        </div>
        <div style="width: 48px; height: 48px; min-width: 48px; min-height: 48px;" class="bg-yellow-100 rounded-full flex items-center justify-center flex-shrink-0">
          <svg style="width: 24px; height: 24px;" class="text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
      </div>
    </div>
  </div>

  <!-- Filters and Actions -->
  <div class="bg-white rounded-lg shadow mb-6">
    <div class="p-6">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div class="flex flex-col sm:flex-row gap-4 flex-1">
          <!-- Search -->
          <div class="flex-1">
            <input
              type="text"
              bind:value={searchTerm}
              placeholder="Search workflows..."
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <!-- Status Filter -->
          <select
            bind:value={statusFilter}
            on:change={loadWorkflows}
            class="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
            <option value="on-hold">On Hold</option>
          </select>
        </div>

        <!-- Create Button -->
        <button
          on:click={() => showCreateModal = true}
          class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2 whitespace-nowrap"
        >
          <svg style="width: 20px; height: 20px;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          New Workflow
        </button>
      </div>
    </div>
  </div>

  <!-- Error Message -->
  {#if error}
    <div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6">
      {error}
    </div>
  {/if}

  <!-- Loading State -->
  {#if loading}
    <div class="flex justify-center items-center h-64">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
    </div>
  {:else if filteredWorkflows.length === 0}
    <div class="bg-white rounded-lg shadow p-12 text-center">
      <svg style="width: 48px; height: 48px;" class="mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">No workflows found</h3>
      <p class="mt-1 text-sm text-gray-500">Get started by creating a new workflow.</p>
      <div class="mt-6">
        <button
          on:click={() => showCreateModal = true}
          class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
        >
          <svg style="width: 20px; height: 20px; margin-left: -4px; margin-right: 8px;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          New Workflow
        </button>
      </div>
    </div>
  {:else}
    <!-- Workflows Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each filteredWorkflows as workflow (workflow.id)}
        <div 
          on:click={() => navigate('workflow-detail', workflow.id)}
          on:keydown={(e) => e.key === 'Enter' && navigate('workflow-detail', workflow.id)}
          role="button"
          tabindex="0"
          class="bg-white rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer"
        >
          <div class="p-6">
            <!-- Header -->
            <div class="flex items-start justify-between mb-4">
              <div class="flex-1">
                <h3 class="text-lg font-semibold text-gray-900 capitalize">
                  {workflow.template_name.replace(/-/g, ' ')}
                </h3>
                <p class="text-sm text-gray-500 mt-1">
                  Started {formatDate(workflow.started_at)}
                </p>
              </div>
              <span class="px-3 py-1 rounded-full text-xs font-medium {getStatusColor(workflow.status)} capitalize">
                {workflow.status}
              </span>
            </div>

            <!-- Progress Bar -->
            <div class="mb-4">
              <div class="flex justify-between items-center mb-2">
                <span class="text-sm font-medium text-gray-700">Progress</span>
                <span class="text-sm font-bold text-gray-900">{workflow.progress_percentage}%</span>
              </div>
              <div class="w-full bg-gray-200 rounded-full h-2">
                <div
                  class="bg-blue-600 h-2 rounded-full transition-all duration-300"
                  style="width: {workflow.progress_percentage}%"
                ></div>
              </div>
            </div>

            <!-- Stage Badge -->
            <div class="flex items-center justify-between">
              <span class="px-3 py-1 rounded-full text-xs font-medium {getStageColor(workflow.current_stage)} capitalize">
                {workflow.current_stage.replace(/-/g, ' ')}
              </span>
              
              {#if workflow.expected_completion}
                <span class="text-xs text-gray-500">
                  Due {formatDate(workflow.expected_completion)}
                </span>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Create Workflow Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
    <div class="bg-white rounded-lg shadow-xl max-w-md w-full">
      <div class="p-6">
        <h2 class="text-xl font-bold text-gray-900 mb-4">Create New Workflow</h2>
        
        <div class="space-y-4">
          <!-- Employee Select -->
          <div>
            <label for="employee-select" class="block text-sm font-medium text-gray-700 mb-2">
              Select Employee
            </label>
            <select
              id="employee-select"
              bind:value={selectedEmployeeId}
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Choose an employee...</option>
              {#each employees as employee}
                <option value={employee.id}>
                  {employee.first_name} {employee.last_name} - {employee.position || 'No position'}
                </option>
              {/each}
            </select>
          </div>

          <!-- Template Select -->
          <div>
            <label for="template-select" class="block text-sm font-medium text-gray-700 mb-2">
              Workflow Template
            </label>
            <select
              id="template-select"
              bind:value={selectedTemplate}
              class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              {#each templates as template}
                <option value={template.value}>{template.label}</option>
              {/each}
            </select>
          </div>

          <!-- Template Description -->
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p class="text-sm text-blue-800">
              {#if selectedTemplate === 'software-engineer'}
                Creates a comprehensive 17-step workflow including pre-boarding, day 1 setup, week 1 training, and month 1 review.
              {:else if selectedTemplate === 'generic'}
                Creates a basic 4-step workflow for standard employee onboarding.
              {:else}
                Creates a customized workflow for this role.
              {/if}
            </p>
          </div>
        </div>

        {#if error}
          <div class="mt-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        {/if}

        <!-- Actions -->
        <div class="flex gap-3 mt-6">
          <button
            on:click={() => { showCreateModal = false; error = ''; }}
            class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            on:click={createWorkflow}
            disabled={!selectedEmployeeId}
            class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
          >
            Create Workflow
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}