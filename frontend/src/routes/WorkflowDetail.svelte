<script lang="ts">
  import { onMount } from 'svelte';

  export let id: string;
  export let navigate: (page: string, id?: string) => void;

  interface WorkflowDetails {
    workflow: any;
    steps: any[];
    exceptions: any[];
    documents: any[];
    employee: any;
  }

  let workflowData: WorkflowDetails | null = null;
  let progress: any = null;
  let loading = true;
  let error = '';
  let activeTab = 'steps';
  
  // For step actions
  let showSkipModal = false;
  let skipStepId = '';
  let skipReason = '';

  onMount(() => {
    loadWorkflowDetails();
    loadProgress();
  });

  async function loadWorkflowDetails() {
    loading = true;
    error = '';
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/workflows/${id}`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!response.ok) throw new Error('Failed to load workflow');
      
      workflowData = await response.json();
    } catch (err: any) {
      error = err.message || 'Failed to load workflow';
    } finally {
      loading = false;
    }
  }

  async function loadProgress() {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/workflows/${id}/progress`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (response.ok) {
        progress = await response.json();
      }
    } catch (err) {
      console.error('Failed to load progress:', err);
    }
  }

  async function completeStep(stepId: string) {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(
        `${import.meta.env.VITE_API_URL}/api/workflows/${id}/steps/${stepId}/complete`,
        {
          method: 'PUT',
          headers: { 'Authorization': `Bearer ${token}` }
        }
      );

      if (!response.ok) throw new Error('Failed to complete step');
      
      await loadWorkflowDetails();
      await loadProgress();
    } catch (err: any) {
      error = err.message;
    }
  }

  async function startStep(stepId: string) {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(
        `${import.meta.env.VITE_API_URL}/api/workflows/${id}/steps/${stepId}/start`,
        {
          method: 'PUT',
          headers: { 'Authorization': `Bearer ${token}` }
        }
      );

      if (!response.ok) throw new Error('Failed to start step');
      
      await loadWorkflowDetails();
    } catch (err: any) {
      error = err.message;
    }
  }

  async function skipStep() {
    if (!skipReason.trim()) {
      error = 'Please provide a reason for skipping';
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(
        `${import.meta.env.VITE_API_URL}/api/workflows/${id}/steps/${skipStepId}/skip`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ reason: skipReason })
        }
      );

      if (!response.ok) throw new Error('Failed to skip step');
      
      showSkipModal = false;
      skipReason = '';
      skipStepId = '';
      await loadWorkflowDetails();
      await loadProgress();
    } catch (err: any) {
      error = err.message;
    }
  }

  function getStatusColor(status: string): string {
    const colors: Record<string, string> = {
      'pending': 'bg-gray-100 text-gray-800',
      'in-progress': 'bg-blue-100 text-blue-800',
      'completed': 'bg-green-100 text-green-800',
      'failed': 'bg-red-100 text-red-800',
      'skipped': 'bg-yellow-100 text-yellow-800',
      'blocked': 'bg-orange-100 text-orange-800'
    };
    return colors[status] || 'bg-gray-100 text-gray-800';
  }

  function getStageColor(stage: string): string {
    const colors: Record<string, string> = {
      'pre-boarding': 'bg-purple-500',
      'day-1': 'bg-blue-500',
      'week-1': 'bg-green-500',
      'month-1': 'bg-orange-500'
    };
    return colors[stage] || 'bg-gray-500';
  }

  function formatDate(dateString: string): string {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function getSeverityColor(severity: string): string {
    const colors: Record<string, string> = {
      'low': 'bg-blue-100 text-blue-800',
      'medium': 'bg-yellow-100 text-yellow-800',
      'high': 'bg-orange-100 text-orange-800',
      'critical': 'bg-red-100 text-red-800'
    };
    return colors[severity] || 'bg-gray-100 text-gray-800';
  }

  // Declare variable with type
  let stageSteps: Record<string, any[]>;

  // Reactive assignment
  $: stageSteps = workflowData ? {
    'pre-boarding': workflowData.steps.filter(s => s.stage === 'pre-boarding'),
    'day-1': workflowData.steps.filter(s => s.stage === 'day-1'),
    'week-1': workflowData.steps.filter(s => s.stage === 'week-1'),
    'month-1': workflowData.steps.filter(s => s.stage === 'month-1')
  } : {
    'pre-boarding': [],
    'day-1': [],
    'week-1': [],
    'month-1': []
  };
</script>

{#if loading}
  <div class="flex justify-center items-center h-screen">
    <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
  </div>
{:else if !workflowData}
  <div class="max-w-7xl mx-auto px-4 py-8">
    <div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
      {error || 'Workflow not found'}
    </div>
  </div>
{:else}
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <!-- Header -->
    <div class="mb-8">
      <div class="flex items-center gap-4 mb-4">
        <button 
          on:click={() => navigate('workflows')}
          class="text-blue-600 hover:text-blue-800"
          aria-label="Go back to workflows list"
        >
          <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 class="text-3xl font-bold text-gray-900">
          {workflowData.employee.first_name} {workflowData.employee.last_name}'s Onboarding
        </h1>
      </div>
      <p class="text-sm text-gray-600">
        {workflowData.employee.position} • {workflowData.employee.department}
      </p>
    </div>

    {#if error}
      <div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6">
        {error}
      </div>
    {/if}

    <!-- Progress Overview -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
      <div class="bg-white rounded-lg shadow p-6">
        <p class="text-sm font-medium text-gray-600 mb-2">Overall Progress</p>
        <p class="text-3xl font-bold text-gray-900">{workflowData.workflow.progress_percentage}%</p>
        <div class="mt-4 w-full bg-gray-200 rounded-full h-2">
          <div
            class="bg-blue-600 h-2 rounded-full transition-all duration-300"
            style="width: {workflowData.workflow.progress_percentage}%"
          ></div>
        </div>
      </div>

      {#if progress}
        <div class="bg-white rounded-lg shadow p-6">
          <p class="text-sm font-medium text-gray-600 mb-2">Steps Status</p>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Completed:</span>
              <span class="font-semibold text-green-600">{progress.completed_steps}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">In Progress:</span>
              <span class="font-semibold text-blue-600">{progress.in_progress_steps}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Pending:</span>
              <span class="font-semibold text-gray-600">{progress.pending_steps}</span>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow p-6">
          <p class="text-sm font-medium text-gray-600 mb-2">Timeline</p>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Days Elapsed:</span>
              <span class="font-semibold">{progress.days_elapsed}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Expected Days:</span>
              <span class="font-semibold">{progress.expected_days}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-600">Status:</span>
              <span class="font-semibold {progress.is_on_track ? 'text-green-600' : 'text-red-600'}">
                {progress.is_on_track ? 'On Track' : 'Behind'}
              </span>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow p-6">
          <p class="text-sm font-medium text-gray-600 mb-2">Issues</p>
          <div class="flex items-center gap-2">
            <svg class="h-8 w-8 {progress.open_exceptions > 0 ? 'text-red-600' : 'text-green-600'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              {#if progress.open_exceptions > 0}
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              {:else}
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              {/if}
            </svg>
            <div>
              <p class="text-2xl font-bold">{progress.open_exceptions}</p>
              <p class="text-xs text-gray-600">Open Exceptions</p>
            </div>
          </div>
        </div>
      {/if}
    </div>

    <!-- Stage Progress Visual -->
    <div class="bg-white rounded-lg shadow p-6 mb-8">
      <h2 class="text-lg font-semibold text-gray-900 mb-6">Workflow Stages</h2>
      <div class="flex items-center justify-between">
        {#each ['pre-boarding', 'day-1', 'week-1', 'month-1'] as stage, index}
          <div class="flex-1 flex items-center">
            <!-- Stage Circle -->
            <div class="flex flex-col items-center">
              <div class="relative">
                <div class="h-12 w-12 rounded-full {getStageColor(stage)} flex items-center justify-center {stage === workflowData.workflow.current_stage ? 'ring-4 ring-offset-2 ring-blue-300' : ''} transition-all">
                  {#if stageSteps[stage]?.every(s => s.status === 'completed' || s.status === 'skipped')}
                    <svg class="h-6 w-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                  {:else}
                    <span class="text-white font-semibold text-sm">{stageSteps[stage]?.filter(s => s.status === 'completed').length || 0}/{stageSteps[stage]?.length || 0}</span>
                  {/if}
                </div>
              </div>
              <p class="mt-2 text-xs font-medium text-gray-700 capitalize text-center">{stage.replace(/-/g, ' ')}</p>
            </div>

            <!-- Connector Line -->
            {#if index < 3}
              <div class="flex-1 h-1 bg-gray-300 mx-2"></div>
            {/if}
          </div>
        {/each}
      </div>
    </div>

    <!-- Tabs -->
    <div class="bg-white rounded-lg shadow mb-6">
      <div class="border-b border-gray-200">
        <nav class="flex space-x-8 px-6" aria-label="Tabs">
          <button
            on:click={() => activeTab = 'steps'}
            class="py-4 px-1 border-b-2 font-medium text-sm {activeTab === 'steps' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
          >
            Steps ({workflowData.steps.length})
          </button>
          <button
            on:click={() => activeTab = 'exceptions'}
            class="py-4 px-1 border-b-2 font-medium text-sm {activeTab === 'exceptions' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
          >
            Exceptions ({workflowData.exceptions.length})
          </button>
          <button
            on:click={() => activeTab = 'documents'}
            class="py-4 px-1 border-b-2 font-medium text-sm {activeTab === 'documents' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
          >
            Documents ({workflowData.documents.length})
          </button>
        </nav>
      </div>

      <!-- Tab Content -->
      <div class="p-6">
        {#if activeTab === 'steps'}
          <!-- Steps by Stage -->
          {#each Object.entries(stageSteps) as [stage, steps]}
            {#if steps.length > 0}
              <div class="mb-8">
                <h3 class="text-lg font-semibold text-gray-900 mb-4 capitalize flex items-center gap-2">
                  <div class="h-3 w-3 rounded-full {getStageColor(stage)}"></div>
                  {stage.replace(/-/g, ' ')}
                  <span class="text-sm font-normal text-gray-500">
                    ({steps.filter(s => s.status === 'completed').length}/{steps.length} completed)
                  </span>
                </h3>

                <div class="space-y-4">
                  {#each steps as step}
                    <div class="border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition-colors">
                      <div class="flex items-start justify-between">
                        <div class="flex-1">
                          <div class="flex items-center gap-3 mb-2">
                            <h4 class="font-medium text-gray-900">{step.step_name}</h4>
                            <span class="px-2 py-1 rounded-full text-xs font-medium {getStatusColor(step.status)} capitalize">
                              {step.status.replace(/-/g, ' ')}
                            </span>
                            {#if step.integration_type}
                              <span class="px-2 py-1 rounded-full text-xs font-medium bg-purple-100 text-purple-800 capitalize">
                                {step.integration_type}
                              </span>
                            {/if}
                          </div>

                          {#if step.description}
                            <p class="text-sm text-gray-600 mb-2">{step.description}</p>
                          {/if}

                          <div class="flex gap-4 text-xs text-gray-500">
                            {#if step.due_date}
                              <span>Due: {formatDate(step.due_date)}</span>
                            {/if}
                            {#if step.completed_at}
                              <span>Completed: {formatDate(step.completed_at)}</span>
                            {/if}
                          </div>
                        </div>

                        <!-- Actions -->
                        <div class="flex gap-2 ml-4">
                          {#if step.status === 'pending'}
                            <button
                              on:click={() => startStep(step.id)}
                              class="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                            >
                              Start
                            </button>
                          {/if}
                          
                          {#if step.status === 'in-progress'}
                            <button
                              on:click={() => completeStep(step.id)}
                              class="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700"
                            >
                              Complete
                            </button>
                          {/if}

                          {#if step.status !== 'completed' && step.status !== 'skipped'}
                            <button
                              on:click={() => { skipStepId = step.id; showSkipModal = true; }}
                              class="px-3 py-1 text-sm border border-gray-300 text-gray-700 rounded hover:bg-gray-50"
                            >
                              Skip
                            </button>
                          {/if}
                        </div>
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          {/each}

        {:else if activeTab === 'exceptions'}
          <!-- Exceptions -->
          {#if workflowData.exceptions.length === 0}
            <div class="text-center py-12">
              <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <h3 class="mt-2 text-sm font-medium text-gray-900">No exceptions</h3>
              <p class="mt-1 text-sm text-gray-500">All steps are running smoothly.</p>
            </div>
          {:else}
            <div class="space-y-4">
              {#each workflowData.exceptions as exception}
                <div class="border border-gray-200 rounded-lg p-4">
                  <div class="flex items-start justify-between mb-2">
                    <div class="flex-1">
                      <div class="flex items-center gap-2 mb-2">
                        <h4 class="font-medium text-gray-900">{exception.title}</h4>
                        <span class="px-2 py-1 rounded-full text-xs font-medium {getSeverityColor(exception.severity)} uppercase">
                          {exception.severity}
                        </span>
                      </div>
                      {#if exception.description}
                        <p class="text-sm text-gray-600 mb-2">{exception.description}</p>
                      {/if}
                      <p class="text-xs text-gray-500">
                        Created: {formatDate(exception.created_at)}
                      </p>
                    </div>
                    <span class="px-2 py-1 rounded-full text-xs font-medium {exception.resolution_status === 'resolved' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'} capitalize">
                      {exception.resolution_status}
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {/if}

        {:else if activeTab === 'documents'}
          <!-- Documents -->
          {#if workflowData.documents.length === 0}
            <div class="text-center py-12">
              <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
              <h3 class="mt-2 text-sm font-medium text-gray-900">No documents</h3>
              <p class="mt-1 text-sm text-gray-500">Documents will appear here as workflow progresses.</p>
            </div>
          {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              {#each workflowData.documents as doc}
                <div class="border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition-colors">
                  <div class="flex items-start gap-3">
                    <div class="flex-shrink-0">
                      <svg class="h-10 w-10 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                      </svg>
                    </div>
                    <div class="flex-1">
                      <h4 class="font-medium text-gray-900">{doc.document_name}</h4>
                      <p class="text-sm text-gray-600 capitalize">{doc.document_type.replace(/-/g, ' ')}</p>
                      <p class="text-xs text-gray-500 mt-1">
                        {doc.file_type.toUpperCase()} • {(doc.file_size / 1024).toFixed(0)} KB
                      </p>
                    </div>
                    <span class="px-2 py-1 rounded-full text-xs font-medium {doc.status === 'signed' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'} capitalize">
                      {doc.status}
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Skip Step Modal -->
{#if showSkipModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
    <div class="bg-white rounded-lg shadow-xl max-w-md w-full">
      <div class="p-6">
        <h2 class="text-xl font-bold text-gray-900 mb-4">Skip Step</h2>
        
        <div class="mb-4">
          <label for="skip-reason" class="block text-sm font-medium text-gray-700 mb-2">
            Reason for skipping
          </label>
          <textarea
            id="skip-reason"
            bind:value={skipReason}
            rows="3"
            placeholder="Provide a reason..."
            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          ></textarea>
        </div>

        <div class="flex gap-3">
          <button
            on:click={() => { showSkipModal = false; skipReason = ''; skipStepId = ''; }}
            class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            on:click={skipStep}
            class="flex-1 px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700"
          >
            Skip Step
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
