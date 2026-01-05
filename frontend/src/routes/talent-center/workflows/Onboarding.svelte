<script lang="ts">
  import { onMount } from 'svelte';
  //import { goto } from '$app/navigation';
  import { authStore } from '../../../stores/auth';
  import { getApiBaseUrl } from '../../../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();
  
  // Make navigate optional with default fallback
  //let { navigate = goto }: { navigate?: (page: string) => void } = $props();
  let { navigate }: { navigate: (page: string) => void } = $props();
  
  let employee = $state($authStore.employee);
  let allTasksCompleted = $state(false);
  
  onMount(() => {
    // Check if employee should be here
    if (employee?.status !== 'onboarding') {
      navigate('employee-dashboard');
    }
  });
  
  async function completeOnboarding() {
    if (!allTasksCompleted) {
      alert('Please complete all onboarding tasks first');
      return;
    }
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${API_BASE_URL}/employees/${employee.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          status: 'active'  // Change from 'onboarding' to 'active'
        })
      });
      
      if (response.ok) {
        const updatedEmployee = await response.json();
        
        // Update auth store
        authStore.update(state => ({
          ...state,
          employee: updatedEmployee
        }));
        
        // Now navigate to employee dashboard
        navigate('employee-dashboard');
      }
    } catch (err) {
      console.error('Error completing onboarding:', err);
      alert('Error completing onboarding. Please try again.');
    }
  }
  
  // Check if all tasks are completed
  function checkAllTasksCompleted() {
    // Your task completion logic
    // allTasksCompleted = tasks.every(t => t.completed);
  }
</script>

<div class="onboarding-container">
  <div class="onboarding-header">
    <h1>Welcome to the Team! 🎉</h1>
    <p>Please complete your onboarding tasks to get started</p>
    <div class="status-badge">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <polyline points="12 6 12 12 16 14"></polyline>
      </svg>
      Status: Onboarding
    </div>
  </div>
  
  <!-- Your existing onboarding tasks component -->
  <div class="tasks-section">
    <!-- Tasks go here -->
  </div>
  
  <!-- Complete button -->
  <div class="completion-section">
    {#if allTasksCompleted}
      <button class="btn-complete" onclick={completeOnboarding}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
        Complete Onboarding & Access Dashboard
      </button>
    {:else}
      <button class="btn-disabled" disabled>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        Complete All Tasks to Proceed
      </button>
    {/if}
  </div>
</div>

<style>
  .onboarding-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
  }
  
  .onboarding-header {
    text-align: center;
    margin-bottom: 3rem;
  }
  
  .onboarding-header h1 {
    font-size: 2.5rem;
    font-weight: 700;
    color: #111827;
    margin-bottom: 0.5rem;
  }
  
  .onboarding-header p {
    font-size: 1.125rem;
    color: #6b7280;
  }
  
  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background: #fef3c7;
    color: #92400e;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.875rem;
    margin-top: 1rem;
  }
  
  .status-badge svg {
    width: 16px;
    height: 16px;
  }
  
  .tasks-section {
    margin-bottom: 2rem;
  }
  
  .completion-section {
    margin-top: 2rem;
    padding: 2rem;
    background: #f9fafb;
    border-radius: 12px;
    text-align: center;
  }
  
  .btn-complete {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem 2rem;
    background: #10b981;
    color: white;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    font-size: 1rem;
    cursor: pointer;
    transition: background 0.2s;
  }
  
  .btn-complete:hover {
    background: #059669;
  }
  
  .btn-complete svg {
    width: 20px;
    height: 20px;
  }
  
  .btn-disabled {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem 2rem;
    background: #e5e7eb;
    color: #9ca3af;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    font-size: 1rem;
    cursor: not-allowed;
  }
  
  .btn-disabled svg {
    width: 20px;
    height: 20px;
  }
</style>