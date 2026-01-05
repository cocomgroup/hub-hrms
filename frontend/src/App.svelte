<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from './stores/auth';
  import { getApiBaseUrl } from './lib/api';
  import Login from './routes/Login.svelte';
  import Dashboard from './routes/Dashboard.svelte';
  import Employees from './routes/Employees.svelte';
  import Onboarding from './routes/talent-center/workflows/Onboarding.svelte';
  import Workflows from './routes/talent-center/workflows/Workflows.svelte';
  import WorkflowDetail from './routes/talent-center/workflows/WorkflowDetail.svelte';
  import Timesheet from './routes/Timesheet.svelte';
  import PTO from './routes/PTO.svelte';
  import Benefits from './routes/Benefits.svelte';
  import Payroll from './routes/Payroll.svelte';
  import HRDashboard from './routes/HRDashboard.svelte';
  import ManagerDashboard from './routes/ManagerDashboard.svelte';
  import EmployeeDashboard from './routes/EmployeeDashboard.svelte'; 
  import Users from './routes/Users.svelte';  
  import Recruiting from './routes/talent-center/recruiting/Recruiting.svelte';  
  import ProjectManagement from './routes/ProjectManagement.svelte';  

  const API_BASE_URL = getApiBaseUrl();

  let currentPage = $state('dashboard');
  let workflowId = $state('');
  let isAuthenticated = $state(false);
  let userRole = $state<string | null>(null);

  onMount(() => {
    const token = localStorage.getItem('token');
    const userStr = localStorage.getItem('user');
    const employeeStr = localStorage.getItem('employee');

    if (token && userStr) {
      const user = JSON.parse(userStr);
      const employee = employeeStr ? JSON.parse(employeeStr) : null;

      authStore.set({
        token,
        user,
        employee,
        isAuthenticated: true
      });

      // ✅ ADD THIS: Check if employee needs onboarding
      if (employee && employee.status === 'Onboarding') {
        currentPage = 'onboarding';
        return;
      }

      // Set current page based on role
      if (user.role === 'admin' || user.role === 'hr-manager') {
        currentPage = 'hr-dashboard';
      } else if (user.role === 'manager') {
        currentPage = 'manager-dashboard';
      } else {
        currentPage = 'employee-dashboard';
      }
    }
  });

  // Subscribe to auth store using $effect
  $effect(() => {
    const unsubscribe = authStore.subscribe(value => {
      isAuthenticated = value.isAuthenticated;
      userRole = value.user?.role || null;
      
        // Role-based routing after login
      if (isAuthenticated && value.user) {
        // Redirect HR managers to HR Manager Dashboard
        if (value.user.role === 'hr-manager' || value.user.role === 'admin') {
          if (currentPage === 'dashboard') {
            currentPage = 'hr-dashboard';
          }
        }
        // Redirect managers to Manager Dashboard
        else if (value.user.role === 'manager') {
          if (currentPage === 'dashboard') {
            currentPage = 'manager-dashboard';
          }
        }
        // Redirect employees to Employee Dashboard
        else {
          if (currentPage === 'dashboard') {
            currentPage = 'employee-dashboard';
          }
        }
      } 
     });
    return unsubscribe;
  });

  function navigate(page: string, id?: string) {
    const employee = $authStore.employee;
    // ✅ ADD THIS: Block navigation if employee is onboarding
    if (employee?.status === 'onboarding' && page !== 'onboarding') {
      console.log('Cannot navigate: Employee must complete onboarding first');
      return;
    }
  
    currentPage = page;
    if (id) {
      workflowId = id;
    }
  }

  function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('employee');
    
    authStore.set({
      token: null,
      user: null,
      employee: null,
      isAuthenticated: false
    });
    
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
          <span class="logo-text">CoCom PeopleHub</span>
        </div>

        <nav class="nav">
          <!-- HR Dashboard - Only visible to HR managers and admins -->
          {#if userRole === 'hr-manager' || userRole === 'admin'}
            <button class="nav-item" class:active={currentPage === 'hr-dashboard'} onclick={() => navigate('hr-dashboard')}>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
              </svg>
              <span>HR Dashboard</span>
            </button>

          {/if}

          <!-- For Managers -->
          {#if userRole === 'manager' || userRole === 'admin'}
            <button 
              class="nav-item" 
              class:active={currentPage === 'manager-dashboard'} 
              onclick={() => navigate('manager-dashboard')}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                <line x1="9" y1="9" x2="15" y2="9"></line>
                <line x1="9" y1="15" x2="15" y2="15"></line>
              </svg>
              <span>Manager Dashboard</span>
            </button>
          {/if}

          <button class="nav-item" class:active={currentPage === 'employee-dashboard'} onclick={() => navigate('employee-dashboard')}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
            <span>Employee Dashboard</span>
          </button>

          <!-- ADD THIS: Users - Only visible to admins -->
          {#if userRole === 'admin'}
            <button class="nav-item" class:active={currentPage === 'users'} onclick={() => navigate('users')}>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                <circle cx="8.5" cy="7" r="4"></circle>
                <line x1="20" y1="8" x2="20" y2="14"></line>
                <line x1="23" y1="11" x2="17" y2="11"></line>
              </svg>
              <span>Users</span>
            </button>
          {/if}
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
          <Onboarding navigate={navigate} />
        {:else if currentPage === 'workflows'}
          <Workflows {navigate} />
        {:else if currentPage === 'hr-dashboard'}
          <HRDashboard navigate={navigate} />
        {:else if currentPage === 'manager-dashboard'}
          <ManagerDashboard navigate={navigate} />
        {:else if currentPage === 'employee-dashboard'}
          <EmployeeDashboard navigate={navigate} />
        {:else if currentPage === 'users'}
          <Users />  
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
        {:else if currentPage === 'projects'}
          <ProjectManagement {navigate} />
        {:else if currentPage === 'recruiting'}
          <Recruiting />
        {:else if currentPage === 'hr-dashboard'}
          <HRDashboard navigate={navigate} />
        {:else if currentPage === 'users'}
          <Users />
        {/if}
      </div>
    </div>
  {/if}
</main>

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