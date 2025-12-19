<script lang="ts">
  import { authStore } from './stores/auth';
  import Login from './routes/Login.svelte';
  import Dashboard from './routes/Dashboard.svelte';
  import Employees from './routes/Employees.svelte';
  import Onboarding from './routes/Onboarding.svelte';
  import Workflows from './routes/Workflows.svelte';
  import WorkflowDetail from './routes/WorkflowDetail.svelte';
  import Timesheet from './routes/Timesheet.svelte';
  import PTO from './routes/PTO.svelte';
  import Benefits from './routes/Benefits.svelte';
  import Payroll from './routes/Payroll.svelte';
  import HRDashboard from './routes/HRDashboard.svelte';

  let currentPage = $state('dashboard');
  let workflowId = $state('');
  let isAuthenticated = $state(false);

  // Subscribe to auth store using $effect
  $effect(() => {
    const unsubscribe = authStore.subscribe(value => {
      isAuthenticated = value.isAuthenticated;
    });
    return unsubscribe;
  });

  function navigate(page: string, id?: string) {
    currentPage = page;
    if (id) {
      workflowId = id;
    }
  }

  function logout() {
    authStore.logout();
    currentPage = 'dashboard';
  }

  // Export navigate for child components
  export function navigateToWorkflowDetail(id: string) {
    workflowId = id;
    currentPage = 'workflow-detail';
  }
</script>

<main class="app">
  {#if !isAuthenticated}
    <Login />
  {:else}
    <div class="app-container">
      <aside class="sidebar">
        <div class="logo">
          <div class="logo-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
          </div>
          <span class="logo-text">PeopleHub</span>
        </div>

        <nav class="nav">
          <button class="nav-item" class:active={currentPage === 'dashboard'} onclick={() => navigate('dashboard')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7"></rect>
              <rect x="14" y="3" width="7" height="7"></rect>
              <rect x="14" y="14" width="7" height="7"></rect>
              <rect x="3" y="14" width="7" height="7"></rect>
            </svg>
            <span>Dashboard</span>
          </button>

          <button class="nav-item" class:active={currentPage === 'employees'} onclick={() => navigate('employees')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
            <span>Employees</span>
          </button>

          <button class="nav-item" class:active={currentPage === 'onboarding'} onclick={() => navigate('onboarding')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
              <polyline points="14 2 14 8 20 8"></polyline>
              <line x1="12" y1="18" x2="12" y2="12"></line>
              <line x1="9" y1="15" x2="15" y2="15"></line>
            </svg>
            <span>Onboarding</span>
          </button>

          <button class="nav-item" class:active={currentPage === 'workflows' || currentPage === 'workflow-detail'} onclick={() => navigate('workflows')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
              <polyline points="7.5 4.21 12 6.81 16.5 4.21"></polyline>
              <polyline points="7.5 19.79 7.5 14.6 3 12"></polyline>
              <polyline points="21 12 16.5 14.6 16.5 19.79"></polyline>
              <polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline>
              <line x1="12" y1="22.08" x2="12" y2="12"></line>
            </svg>
            <span>Workflows</span>
          </button>

          <button class="nav-item" class:active={currentPage === 'timesheet'} onclick={() => navigate('timesheet')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
            <span>Timesheet</span>
          </button>

          <button class="nav-item" class:active={currentPage === 'pto'} onclick={() => navigate('pto')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
              <line x1="16" y1="2" x2="16" y2="6"></line>
              <line x1="8" y1="2" x2="8" y2="6"></line>
              <line x1="3" y1="10" x2="21" y2="10"></line>
            </svg>
            <span>Time Off</span>
          </button>

          <button class="nav-item" class:active={currentPage === 'benefits'} onclick={() => navigate('benefits')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
              <polyline points="22 4 12 14.01 9 11.01"></polyline>
            </svg>
            <span>Benefits</span>
          </button>

          <button class="nav-item" class:active={currentPage === 'payroll'} onclick={() => navigate('payroll')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="1" x2="12" y2="23"></line>
              <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
            </svg>
            <span>Payroll</span>
          </button>
        </nav>

        <button class="logout-btn" onclick={logout}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
            <polyline points="16 17 21 12 16 7"></polyline>
            <line x1="21" y1="12" x2="9" y2="12"></line>
          </svg>
          <span>Logout</span>
        </button>
      </aside>

      <div class="content">
        {#if currentPage === 'dashboard'}
          <Dashboard />
        {:else if currentPage === 'employees'}
          <Employees />
        {:else if currentPage === 'onboarding'}
          <Onboarding />
        {:else if currentPage === 'workflows'}
          <Workflows {navigate} />
        {:else if currentPage === 'workflow-detail'}
          <WorkflowDetail id={workflowId} {navigate} />
        {:else if currentPage === 'timesheet'}
          <Timesheet />
        {:else if currentPage === 'pto'}
          <PTO />
        {:else if currentPage === 'benefits'}
          <Benefits />
        {:else if currentPage === 'payroll'}
          <Payroll />
        {:else if currentPage === 'hr-dashboard'}
          <HRDashboard navigate={navigate} />
        {/if}
      </div>
    </div>
  {/if}
</main>

<button class="nav-item" class:active={currentPage === 'hr-dashboard'} 
        onclick={() => navigate('hr-dashboard')}>
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
  </svg>
  <span>HR Dashboard</span>
</button>

<style>
  :global(*) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  :global(body) {
    font-family: 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    background: #0a0e1a;
    color: #e4e7eb;
    overflow-x: hidden;
  }

  .app {
    min-height: 100vh;
    background: 
      radial-gradient(circle at 20% 10%, rgba(99, 102, 241, 0.08) 0%, transparent 50%),
      radial-gradient(circle at 80% 90%, rgba(139, 92, 246, 0.06) 0%, transparent 50%),
      linear-gradient(180deg, #0a0e1a 0%, #0f1629 100%);
  }

  .app-container {
    display: flex;
    min-height: 100vh;
  }

  .sidebar {
    width: 280px;
    background: linear-gradient(180deg, rgba(17, 24, 39, 0.95) 0%, rgba(17, 24, 39, 0.85) 100%);
    backdrop-filter: blur(20px);
    border-right: 1px solid rgba(99, 102, 241, 0.1);
    padding: 2rem 1.5rem;
    display: flex;
    flex-direction: column;
    position: sticky;
    top: 0;
    height: 100vh;
    overflow-y: auto;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 3rem;
    padding: 0 0.5rem;
  }

  .logo-icon {
    width: 44px;
    height: 44px;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
  }

  .logo-icon svg {
    width: 24px;
    height: 24px;
    color: white;
  }

  .logo-text {
    font-size: 1.5rem;
    font-weight: 700;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .nav {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 0.875rem;
    padding: 0.875rem 1rem;
    background: transparent;
    border: none;
    border-radius: 10px;
    color: #94a3b8;
    font-size: 0.9375rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    overflow: hidden;
  }

  .nav-item::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.1) 0%, rgba(139, 92, 246, 0.08) 100%);
    opacity: 0;
    transition: opacity 0.2s;
  }

  .nav-item:hover::before {
    opacity: 1;
  }

  .nav-item:hover {
    color: #e4e7eb;
    transform: translateX(2px);
  }

  .nav-item.active {
    background: linear-gradient(135deg, rgba(99, 102, 241, 0.15) 0%, rgba(139, 92, 246, 0.1) 100%);
    color: #6366f1;
    box-shadow: 0 2px 8px rgba(99, 102, 241, 0.2);
  }

  .nav-item svg {
    width: 20px;
    height: 20px;
    min-width: 20px;
    position: relative;
    z-index: 1;
  }

  .nav-item span {
    position: relative;
    z-index: 1;
  }

  .logout-btn {
    display: flex;
    align-items: center;
    gap: 0.875rem;
    padding: 0.875rem 1rem;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: 10px;
    color: #ef4444;
    font-size: 0.9375rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    margin-top: 1rem;
  }

  .logout-btn:hover {
    background: rgba(239, 68, 68, 0.15);
    border-color: rgba(239, 68, 68, 0.3);
    transform: translateX(2px);
  }

  .logout-btn svg {
    width: 20px;
    height: 20px;
  }

  .content {
    flex: 1;
    padding: 2rem;
    overflow-y: auto;
  }

  @media (max-width: 768px) {
    .sidebar {
      width: 100%;
      height: auto;
      position: relative;
    }

    .app-container {
      flex-direction: column;
    }
  }
</style>
