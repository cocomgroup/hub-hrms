<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  import ContractorView from '../components/ContractorView.svelte';
  import EmployeeView from '../components/EmployeeView.svelte';

  const API_BASE_URL = getApiBaseUrl();

  interface Props {
    navigate: (page: string) => void;
  }
  
  let { navigate }: Props = $props();

  let employee = $state($authStore.employee);
  let loading = $state(true);
  let employmentType = $state<string>('');

  onMount(() => {
    loadEmployeeInfo();
  });

  async function loadEmployeeInfo() {
    loading = true;
    console.log('[EmployeeDashboard] Starting to load employee info');
    console.log('[EmployeeDashboard] authStore.employee:', $authStore.employee);
    try {
      // Get employee info from auth store or API
      if ($authStore.employee) {
        employee = $authStore.employee;
        employmentType = employee.employment_type?.toLowerCase() || 'employee';
        console.log('[EmployeeDashboard] Loaded from store - Type:', employmentType);
      } else {
        // If not in store, load from API
        console.log('[EmployeeDashboard] Employee not in store, fetching from API');
        await fetchEmployeeInfo();
      }
    } catch (err) {
      console.error('[EmployeeDashboard] Error loading employee info:', err);
      employmentType = 'employee'; // Default to employee view
    } finally {
      loading = false;
      console.log('[EmployeeDashboard] Loading complete - isContractor:', isContractor);
    }
  }

  async function fetchEmployeeInfo() {
    try {
      console.log('[EmployeeDashboard] Fetching from API:', `${API_BASE_URL}/employees/me`);
      const response = await fetch(`${API_BASE_URL}/employees/me`, {
        headers: {
          'Authorization': `Bearer ${$authStore.token}`
        }
      });
      
      console.log('[EmployeeDashboard] API response status:', response.status);
      if (response.ok) {
        employee = await response.json();
        employmentType = employee.employment_type?.toLowerCase() || 'employee';
        console.log('[EmployeeDashboard] Fetched employee:', employee);
        console.log('[EmployeeDashboard] Employment type:', employmentType);
      } else {
        console.error('[EmployeeDashboard] API returned error:', response.status);
      }
    } catch (err) {
      console.error('[EmployeeDashboard] Error fetching employee:', err);
    }
  }

  // Determine if user is a contractor (using $derived for Svelte 5)
  let isContractor = $derived(
    employmentType === 'contractor' || 
    employmentType === 'independent contractor' ||
    employmentType === '1099'
  );

  // Reactive: update when employee changes in auth store
  $effect(() => {
    if ($authStore.employee && $authStore.employee.id !== employee?.id) {
      employee = $authStore.employee;
      employmentType = employee.employment_type?.toLowerCase() || 'employee';
    }
  });
</script>

<div class="employee-dashboard-router">
  {#if loading}
    <div class="loading-container">
      <div class="spinner"></div>
      <p>Loading your dashboard...</p>
    </div>
  {:else if isContractor}
    <!-- Contractor View: Limited to Timesheets and Tasks -->
    <ContractorView {employee} />
  {:else}
    <!-- Full Employee View: All features -->
    <EmployeeView {navigate} {employee} />
  {/if}
</div>

<style>
  .employee-dashboard-router {
    min-height: 100vh;
    background: #f9fafb;
  }

  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 400px;
    gap: 16px;
  }

  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading-container p {
    color: #6b7280;
    font-size: 16px;
  }
</style>