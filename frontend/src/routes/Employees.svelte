<script lang="ts">
  import { authStore } from '../stores/auth';

  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    phone: string;
    date_of_birth: string | null;
    hire_date: string;
    department: string;
    position: string;
    manager_id: string | null;
    employment_type: string;
    status: string;
    street_address: string | null;
    city: string | null;
    state: string | null;
    zip_code: string | null;
    country: string | null;
    emergency_contact_name: string | null;
    emergency_contact_phone: string | null;
  }

  let employees = $state<Employee[]>([]);
  let loading = $state(true);
  let error = $state('');
  let showAddModal = $state(false);
  let showEditModal = $state(false);
  let selectedEmployee = $state<Employee | null>(null);
  let searchTerm = $state('');
  let filterDepartment = $state('all');
  let filterStatus = $state('active');

  // New employee form
  let newEmployee = $state({
    first_name: '',
    last_name: '',
    email: '',
    phone: '',
    date_of_birth: '',
    hire_date: new Date().toISOString().split('T')[0],
    department: '',
    position: '',
    employment_type: 'full-time',
    status: 'active',
    street_address: '',
    city: '',
    state: '',
    zip_code: '',
    country: 'USA',
    emergency_contact_name: '',
    emergency_contact_phone: ''
  });

  $effect(() => {
    loadEmployees();
  });

  async function loadEmployees() {
    loading = true;
    error = '';
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/employees`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to load employees');
      }

      const data = await response.json();
      // Filter out system administrator
      employees = (data || []).filter((emp: Employee) => 
        emp.email !== 'admin@hub-hrms.local'
      );
    } catch (err: any) {
      error = err.message || 'Failed to load employees';
      employees = [];
    } finally {
      loading = false;
    }
  }

  async function createEmployee() {
    try {
      const token = localStorage.getItem('token');
      
      // Prepare employee data, removing empty date_of_birth
      const employeeData = { ...newEmployee };
      if (!employeeData.date_of_birth) {
        delete employeeData.date_of_birth;
      }
      
      console.log('Creating employee with data:', employeeData);
      
      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/employees`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(employeeData)
      });

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Create employee error:', errorText);
        throw new Error(errorText || 'Failed to create employee');
      }

      await loadEmployees();
      closeAddModal();
    } catch (err: any) {
      console.error('Create employee exception:', err);
      error = err.message || 'Failed to create employee';
    }
  }

  async function updateEmployee() {
    if (!selectedEmployee) return;

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/employees/${selectedEmployee.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(selectedEmployee)
      });

      if (!response.ok) {
        throw new Error('Failed to update employee');
      }

      await loadEmployees();
      closeEditModal();
    } catch (err: any) {
      error = err.message || 'Failed to update employee';
    }
  }

  function openAddModal() {
    showAddModal = true;
    resetNewEmployee();
  }

  function closeAddModal() {
    showAddModal = false;
    resetNewEmployee();
  }

  function openEditModal(employee: Employee) {
    selectedEmployee = { ...employee };
    showEditModal = true;
  }

  function closeEditModal() {
    showEditModal = false;
    selectedEmployee = null;
  }

  function resetNewEmployee() {
    newEmployee = {
      first_name: '',
      last_name: '',
      email: '',
      phone: '',
      date_of_birth: '',
      hire_date: new Date().toISOString().split('T')[0],
      department: '',
      position: '',
      employment_type: 'full-time',
      status: 'active',
      street_address: '',
      city: '',
      state: '',
      zip_code: '',
      country: 'USA',
      emergency_contact_name: '',
      emergency_contact_phone: ''
    };
  }

  // Computed filtered employees
  let filteredEmployees = $derived(() => {
    let result = employees;

    // Filter by search term
    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      result = result.filter(emp =>
        emp.first_name.toLowerCase().includes(term) ||
        emp.last_name.toLowerCase().includes(term) ||
        emp.email.toLowerCase().includes(term) ||
        emp.department.toLowerCase().includes(term) ||
        emp.position.toLowerCase().includes(term)
      );
    }

    // Filter by department
    if (filterDepartment !== 'all') {
      result = result.filter(emp => emp.department === filterDepartment);
    }

    // Filter by status
    if (filterStatus !== 'all') {
      result = result.filter(emp => emp.status === filterStatus);
    }

    return result;
  });

  // Get unique departments
  let departments = $derived(() => {
    const depts = new Set(employees.map(e => e.department));
    return Array.from(depts).sort();
  });

  function formatDate(dateString: string | null) {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString();
  }

  function getStatusColor(status: string) {
    switch (status) {
      case 'active': return 'status-active';
      case 'terminated': return 'status-terminated';
      case 'on-leave': return 'status-leave';
      default: return '';
    }
  }
</script>

<div class="employees-page">
  <header class="page-header">
    <div>
      <h1>Employees</h1>
      <p>Manage your organization's employees</p>
    </div>
    <button class="btn-primary" onclick={openAddModal}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M12 5v14M5 12h14"></path>
      </svg>
      Add Employee
    </button>
  </header>

  {#if error}
    <div class="error-banner">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="8" x2="12" y2="12"></line>
        <line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      {error}
    </div>
  {/if}

  <!-- Filters -->
  <div class="filters">
    <div class="search-box">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"></circle>
        <path d="m21 21-4.35-4.35"></path>
      </svg>
      <input
        type="text"
        bind:value={searchTerm}
        placeholder="Search employees..."
      />
    </div>

    <select bind:value={filterDepartment} class="filter-select">
      <option value="all">All Departments</option>
      {#each departments() as dept}
        <option value={dept}>{dept}</option>
      {/each}
    </select>

    <select bind:value={filterStatus} class="filter-select">
      <option value="all">All Status</option>
      <option value="active">Active</option>
      <option value="on-leave">On Leave</option>
      <option value="terminated">Terminated</option>
    </select>
  </div>

  <!-- Employee List -->
  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading employees...</p>
    </div>
  {:else if filteredEmployees().length === 0}
    <div class="empty-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
        <circle cx="9" cy="7" r="4"></circle>
        <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
        <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
      </svg>
      <p>No employees found</p>
      <button class="btn-secondary" onclick={openAddModal}>Add First Employee</button>
    </div>
  {:else}
    <div class="employee-grid">
      {#each filteredEmployees() as employee}
        <div class="employee-card">
          <div class="employee-header">
            <div class="employee-avatar">
              {employee.first_name[0]}{employee.last_name[0]}
            </div>
            <div class="employee-info">
              <h3>{employee.first_name} {employee.last_name}</h3>
              <p>{employee.position}</p>
              <span class="status-badge {getStatusColor(employee.status)}">
                {employee.status}
              </span>
            </div>
          </div>

          <div class="employee-details">
            <div class="detail-row">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                <polyline points="22,6 12,13 2,6"></polyline>
              </svg>
              <span>{employee.email}</span>
            </div>

            <div class="detail-row">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"></path>
              </svg>
              <span>{employee.phone}</span>
            </div>

            <div class="detail-row">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                <circle cx="12" cy="7" r="4"></circle>
              </svg>
              <span>{employee.department}</span>
            </div>

            <div class="detail-row">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                <line x1="16" y1="2" x2="16" y2="6"></line>
                <line x1="8" y1="2" x2="8" y2="6"></line>
                <line x1="3" y1="10" x2="21" y2="10"></line>
              </svg>
              <span>Hired: {formatDate(employee.hire_date)}</span>
            </div>
          </div>

          <button class="btn-edit" onclick={() => openEditModal(employee)}>
            Edit Details
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Add Employee Modal -->
{#if showAddModal}
  <div 
    class="modal-overlay" 
    onclick={(e) => e.target === e.currentTarget && closeAddModal()}
    onkeydown={(e) => e.key === 'Escape' && closeAddModal()}
    role="button"
    tabindex="0"
    aria-label="Close modal"
  >
    <div 
      class="modal"
      role="dialog"
      aria-modal="true"
      aria-labelledby="add-employee-title"
      tabindex="-1"
    >
      <div class="modal-header">
        <h2 id="add-employee-title">Add New Employee</h2>
        <button class="close-btn" onclick={closeAddModal}>×</button>
      </div>

      <form onsubmit={(e) => { e.preventDefault(); createEmployee(); }} class="modal-body">
        <div class="form-grid">
          <!-- Personal Information -->
          <div class="form-section">
            <h3>Personal Information</h3>
            
            <div class="form-row">
              <div class="form-group">
                <label>First Name *<input type="text" bind:value={newEmployee.first_name} required /></label>
              </div>
              <div class="form-group">
                <label>Last Name *<input type="text" bind:value={newEmployee.last_name} required /></label>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>Email *<input type="email" bind:value={newEmployee.email} required /></label>
              </div>
              <div class="form-group">
                <label>Phone *<input type="tel" bind:value={newEmployee.phone} required /></label>
              </div>
            </div>

            <div class="form-group">
              <label>Date of Birth<input type="date" bind:value={newEmployee.date_of_birth} /></label>
            </div>
          </div>

          <!-- Employment Information -->
          <div class="form-section">
            <h3>Employment Details</h3>
            
            <div class="form-row">
              <div class="form-group">
                <label>Department *<input type="text" bind:value={newEmployee.department} required /></label>
              </div>
              <div class="form-group">
                <label>Position *<input type="text" bind:value={newEmployee.position} required /></label>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>Employment Type *<select bind:value={newEmployee.employment_type} required>
                  <option value="full-time">Full-time</option>
                  <option value="part-time">Part-time</option>
                  <option value="contract">Contract</option>
                  <option value="intern">Intern</option>
                </select></label>
              </div>
              <div class="form-group">
                <label>Hire Date *<input type="date" bind:value={newEmployee.hire_date} required /></label>
              </div>
            </div>

            <div class="form-group">
              <label>Status *<select bind:value={newEmployee.status} required>
                <option value="active">Active</option>
                <option value="on-leave">On Leave</option>
                <option value="terminated">Terminated</option>
              </select></label>
            </div>
          </div>

          <!-- Address Information -->
          <div class="form-section">
            <h3>Address</h3>
            
            <div class="form-group">
              <label>Street Address<input type="text" bind:value={newEmployee.street_address} /></label>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>City<input type="text" bind:value={newEmployee.city} /></label>
              </div>
              <div class="form-group">
                <label>State<input type="text" bind:value={newEmployee.state} /></label>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>ZIP Code<input type="text" bind:value={newEmployee.zip_code} /></label>
              </div>
              <div class="form-group">
                <label>Country<input type="text" bind:value={newEmployee.country} /></label>
              </div>
            </div>
          </div>

          <!-- Emergency Contact -->
          <div class="form-section">
            <h3>Emergency Contact</h3>
            
            <div class="form-group">
              <label>Contact Name<input type="text" bind:value={newEmployee.emergency_contact_name} /></label>
            </div>

            <div class="form-group">
              <label>Contact Phone<input type="tel" bind:value={newEmployee.emergency_contact_phone} /></label>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn-secondary" onclick={closeAddModal}>Cancel</button>
          <button type="submit" class="btn-primary">Create Employee</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Edit Employee Modal -->
{#if showEditModal && selectedEmployee}
  <div 
    class="modal-overlay" 
    onclick={(e) => e.target === e.currentTarget && closeEditModal()}
    onkeydown={(e) => e.key === 'Escape' && closeEditModal()}
    role="button"
    tabindex="0"
    aria-label="Close modal"
  >
    <div 
      class="modal"
      role="dialog"
      aria-modal="true"
      aria-labelledby="edit-employee-title"
      tabindex="-1"
    >
      <div class="modal-header">
        <h2 id="edit-employee-title">Edit Employee</h2>
        <button class="close-btn" onclick={closeEditModal}>×</button>
      </div>

      <form onsubmit={(e) => { e.preventDefault(); updateEmployee(); }} class="modal-body">
        <div class="form-grid">
          <div class="form-section">
            <h3>Personal Information</h3>
            
            <div class="form-row">
              <div class="form-group">
                <label>First Name *<input type="text" bind:value={selectedEmployee.first_name} required /></label>
              </div>
              <div class="form-group">
                <label>Last Name *<input type="text" bind:value={selectedEmployee.last_name} required /></label>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label>Email *<input type="email" bind:value={selectedEmployee.email} required /></label>
              </div>
              <div class="form-group">
                <label>Phone *<input type="tel" bind:value={selectedEmployee.phone} required /></label>
              </div>
            </div>
          </div>

          <div class="form-section">
            <h3>Employment Details</h3>
            
            <div class="form-row">
              <div class="form-group">
                <label>Department *<input type="text" bind:value={selectedEmployee.department} required /></label>
              </div>
              <div class="form-group">
                <label>Position *<input type="text" bind:value={selectedEmployee.position} required /></label>
              </div>
            </div>

            <div class="form-group">
              <label>Status *<select bind:value={selectedEmployee.status} required>
                <option value="active">Active</option>
                <option value="on-leave">On Leave</option>
                <option value="terminated">Terminated</option>
              </select></label>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn-secondary" onclick={closeEditModal}>Cancel</button>
          <button type="submit" class="btn-primary">Update Employee</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<style>
  .employees-page {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 2rem;
  }

  .page-header h1 {
    font-size: 2rem;
    font-weight: 700;
    color: #e4e7eb;
    margin-bottom: 0.5rem;
  }

  .page-header p {
    color: #94a3b8;
    font-size: 0.9375rem;
  }

  .btn-primary {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    color: white;
    border: none;
    border-radius: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(99, 102, 241, 0.4);
  }

  .btn-primary svg {
    width: 20px;
    height: 20px;
  }

  .btn-secondary {
    padding: 0.75rem 1.5rem;
    background: rgba(99, 102, 241, 0.1);
    color: #6366f1;
    border: 1px solid rgba(99, 102, 241, 0.3);
    border-radius: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-secondary:hover {
    background: rgba(99, 102, 241, 0.2);
  }

  .error-banner {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem 1.5rem;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: 12px;
    color: #ef4444;
    margin-bottom: 1.5rem;
  }

  .error-banner svg {
    width: 24px;
    height: 24px;
    min-width: 24px;
  }

  .filters {
    display: flex;
    gap: 1rem;
    margin-bottom: 2rem;
    flex-wrap: wrap;
  }

  .search-box {
    flex: 1;
    min-width: 300px;
    position: relative;
    display: flex;
    align-items: center;
  }

  .search-box svg {
    position: absolute;
    left: 1rem;
    width: 20px;
    height: 20px;
    color: #64748b;
  }

  .search-box input {
    width: 100%;
    padding: 0.875rem 1rem 0.875rem 3rem;
    background: rgba(15, 23, 42, 0.6);
    border: 1px solid rgba(99, 102, 241, 0.2);
    border-radius: 12px;
    color: #e4e7eb;
    font-size: 0.9375rem;
    transition: all 0.2s;
  }

  .search-box input:focus {
    outline: none;
    border-color: #6366f1;
    background: rgba(15, 23, 42, 0.8);
  }

  .filter-select {
    padding: 0.875rem 1rem;
    background: rgba(15, 23, 42, 0.6);
    border: 1px solid rgba(99, 102, 241, 0.2);
    border-radius: 12px;
    color: #e4e7eb;
    font-size: 0.9375rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .filter-select:focus {
    outline: none;
    border-color: #6366f1;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    gap: 1rem;
  }

  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid rgba(99, 102, 241, 0.1);
    border-top-color: #6366f1;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading p {
    color: #94a3b8;
    font-size: 0.9375rem;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    gap: 1rem;
  }

  .empty-state svg {
    width: 64px;
    height: 64px;
    color: #64748b;
    margin-bottom: 1rem;
  }

  .empty-state p {
    color: #94a3b8;
    font-size: 1.125rem;
    margin-bottom: 1rem;
  }

  .employee-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }

  .employee-card {
    background: rgba(15, 23, 42, 0.6);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(99, 102, 241, 0.2);
    border-radius: 16px;
    padding: 1.5rem;
    transition: all 0.3s;
  }

  .employee-card:hover {
    transform: translateY(-4px);
    border-color: rgba(99, 102, 241, 0.4);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  }

  .employee-header {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .employee-avatar {
    width: 56px;
    height: 56px;
    min-width: 56px;
    border-radius: 12px;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.25rem;
    font-weight: 700;
    color: white;
  }

  .employee-info h3 {
    font-size: 1.125rem;
    font-weight: 600;
    color: #e4e7eb;
    margin-bottom: 0.25rem;
  }

  .employee-info p {
    color: #94a3b8;
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
  }

  .status-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: capitalize;
  }

  .status-active {
    background: rgba(34, 197, 94, 0.1);
    color: #22c55e;
  }

  .status-terminated {
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
  }

  .status-leave {
    background: rgba(251, 191, 36, 0.1);
    color: #fbbf24;
  }

  .employee-details {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 1.5rem;
  }

  .detail-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: #94a3b8;
    font-size: 0.875rem;
  }

  .detail-row svg {
    width: 16px;
    height: 16px;
    min-width: 16px;
    color: #6366f1;
  }

  .btn-edit {
    width: 100%;
    padding: 0.75rem;
    background: rgba(99, 102, 241, 0.1);
    color: #6366f1;
    border: 1px solid rgba(99, 102, 241, 0.3);
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-edit:hover {
    background: rgba(99, 102, 241, 0.2);
  }

  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 2rem;
    overflow-y: auto;
  }

  .modal {
    background: rgba(17, 24, 39, 0.95);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(99, 102, 241, 0.2);
    border-radius: 24px;
    max-width: 900px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem 2rem;
    border-bottom: 1px solid rgba(99, 102, 241, 0.2);
  }

  .modal-header h2 {
    font-size: 1.5rem;
    font-weight: 700;
    color: #e4e7eb;
  }

  .close-btn {
    width: 32px;
    height: 32px;
    border: none;
    background: rgba(99, 102, 241, 0.1);
    color: #e4e7eb;
    border-radius: 8px;
    font-size: 1.5rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: rgba(99, 102, 241, 0.2);
  }

  .modal-body {
    padding: 2rem;
  }

  .form-grid {
    display: grid;
    gap: 2rem;
  }

  .form-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-section h3 {
    font-size: 1.125rem;
    font-weight: 600;
    color: #e4e7eb;
    margin-bottom: 0.5rem;
  }

  .form-row {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form-group label {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: #e4e7eb;
  }

  .form-group input,
  .form-group select {
    padding: 0.875rem 1rem;
    background: rgba(15, 23, 42, 0.6);
    border: 1px solid rgba(99, 102, 241, 0.2);
    border-radius: 8px;
    color: #e4e7eb;
    font-size: 0.9375rem;
    transition: all 0.2s;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: #6366f1;
    background: rgba(15, 23, 42, 0.8);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding: 1.5rem 2rem;
    border-top: 1px solid rgba(99, 102, 241, 0.2);
  }

  @media (max-width: 768px) {
    .form-row {
      grid-template-columns: 1fr;
    }

    .employee-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
