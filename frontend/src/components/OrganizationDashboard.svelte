<script lang="ts">
  import { onMount } from 'svelte';
  import EmployeeDetail from './EmployeeDetail.svelte';
  
  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    department: string;
    position: string;
    hire_date: string;
    status: string;
    manager_id: string | null;
    manager_name?: string;
  }
  
  interface OrgNode {
    employee: Employee;
    directReports: OrgNode[];
  }
  
  let activeTab: 'view' | 'manage' | 'hierarchy' = 'view';
  let employees: Employee[] = [];
  let managers: Employee[] = [];
  let loading = true;
  let error = '';
  
  // View All tab
  let searchQuery = '';
  let filterDepartment = '';
  let filterStatus = '';
  let sortBy: 'name' | 'department' | 'hire_date' = 'name';
  let sortDirection: 'asc' | 'desc' = 'asc';
  
  // Pagination - matching EmployeeList.svelte style
  let currentPage = 1;
  let itemsPerPage = 12;
  
  // Employee Detail Modal
  let showEmployeeDetail = false;
  let selectedEmployeeId: string | null = null;
  
  // Employee Management tab (for manager assignment)
  let selectedEmployee: Employee | null = null;
  let selectedManagerId = '';
  let showAssignModal = false;
  let assignmentSuccess = '';
  let assignmentError = '';
  
  // Organization Hierarchy tab
  let orgTree: OrgNode[] = [];
  let expandedNodes = new Set<string>();
  let searchHierarchy = '';
  
  onMount(async () => {
    await loadEmployees();
  });
  
  async function loadEmployees() {
    loading = true;
    error = '';
    
    try {
      const token = localStorage.getItem('token');
      
      if (!token) {
        error = 'Not authenticated. Please log in.';
        loading = false;
        return;
      }
      
      const response = await fetch('/api/employees', {
        headers: { 
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      
      if (response.status === 401) {
        error = 'Authentication failed. Please log in again.';
        loading = false;
        localStorage.removeItem('token');
        return;
      }
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to load employees: ${response.status} ${errorText}`);
      }
      
      const data = await response.json();
      employees = data;
      
      // Get managers (employees who have direct reports or specific positions)
      managers = employees.filter(emp => 
        emp.position?.toLowerCase().includes('manager') ||
        emp.position?.toLowerCase().includes('director') ||
        emp.position?.toLowerCase().includes('lead') ||
        employees.some(e => e.manager_id === emp.id)
      );
      
      // Build organization hierarchy
      buildOrgTree();
      
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load data';
      console.error('Error loading employees:', err);
    } finally {
      loading = false;
    }
  }
  
  function buildOrgTree() {
    // Find top-level employees (no manager)
    const topLevel = employees.filter(emp => !emp.manager_id);
    
    // Recursively build tree
    function buildNode(employee: Employee): OrgNode {
      const directReports = employees
        .filter(emp => emp.manager_id === employee.id)
        .map(emp => buildNode(emp));
      
      return {
        employee,
        directReports: directReports.sort((a, b) => 
          `${a.employee.first_name} ${a.employee.last_name}`.localeCompare(
            `${b.employee.first_name} ${b.employee.last_name}`
          )
        )
      };
    }
    
    orgTree = topLevel.map(emp => buildNode(emp));
  }
  
  // View All - Filtering and Sorting
  $: departments = [...new Set(employees.map(e => e.department).filter(Boolean))];
  
  $: filteredEmployees = employees
    .filter(emp => {
      const matchesSearch = searchQuery === '' || 
        `${emp.first_name} ${emp.last_name}`.toLowerCase().includes(searchQuery.toLowerCase()) ||
        emp.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
        emp.position?.toLowerCase().includes(searchQuery.toLowerCase());
      
      const matchesDepartment = !filterDepartment || emp.department === filterDepartment;
      const matchesStatus = !filterStatus || emp.status === filterStatus;
      
      return matchesSearch && matchesDepartment && matchesStatus;
    })
    .sort((a, b) => {
      let comparison = 0;
      
      if (sortBy === 'name') {
        const nameA = `${a.first_name} ${a.last_name}`;
        const nameB = `${b.first_name} ${b.last_name}`;
        comparison = nameA.localeCompare(nameB);
      } else if (sortBy === 'department') {
        comparison = (a.department || '').localeCompare(b.department || '');
      } else if (sortBy === 'hire_date') {
        comparison = new Date(a.hire_date).getTime() - new Date(b.hire_date).getTime();
      }
      
      return sortDirection === 'asc' ? comparison : -comparison;
    });
  
  // Pagination calculations
  $: totalPages = Math.ceil(filteredEmployees.length / itemsPerPage);
  $: paginatedEmployees = filteredEmployees.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );
  
  // Reset to page 1 when filters change
  $: {
    searchQuery;
    filterDepartment;
    filterStatus;
    currentPage = 1;
  }
  
  function goToPage(page: number) {
    if (page >= 1 && page <= totalPages) {
      currentPage = page;
      // Scroll to top of employee grid
      const grid = document.querySelector('.employee-grid');
      if (grid) {
        grid.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }
    }
  }
  
  // Employee Management - Assignment
  function openAssignModal(employee: Employee) {
    selectedEmployee = employee;
    selectedManagerId = employee.manager_id || '';
    showAssignModal = true;
    assignmentSuccess = '';
    assignmentError = '';
  }
  
  // Open Employee Detail Modal
  function openEmployeeDetail(employeeId: string) {
    selectedEmployeeId = employeeId;
    showEmployeeDetail = true;
  }
  
  function closeEmployeeDetail() {
    showEmployeeDetail = false;
    selectedEmployeeId = null;
    // Reload employees in case anything changed
    loadEmployees();
  }
  
  async function assignManager() {
    if (!selectedEmployee) return;
    
    assignmentError = '';
    assignmentSuccess = '';
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/employees/${selectedEmployee.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          ...selectedEmployee,
          manager_id: selectedManagerId || null
        })
      });
      
      if (!response.ok) throw new Error('Failed to assign manager');
      
      assignmentSuccess = `Successfully ${selectedManagerId ? 'assigned' : 'removed'} manager for ${selectedEmployee.first_name} ${selectedEmployee.last_name}`;
      
      // Reload data
      await loadEmployees();
      
      // Close modal after 2 seconds
      setTimeout(() => {
        showAssignModal = false;
        assignmentSuccess = '';
      }, 2000);
      
    } catch (err) {
      assignmentError = err instanceof Error ? err.message : 'Failed to assign manager';
    }
  }
  
  // Organization Hierarchy - Expand/Collapse
  function toggleNode(nodeId: string) {
    if (!expandedNodes) {
      expandedNodes = new Set();
    }
    
    if (expandedNodes.has(nodeId)) {
      expandedNodes.delete(nodeId);
    } else {
      expandedNodes.add(nodeId);
    }
    expandedNodes = expandedNodes; // Trigger reactivity
  }
  
  function expandAll() {
    const allIds = new Set<string>();
    
    function collectIds(node: OrgNode) {
      allIds.add(node.employee.id);
      node.directReports.forEach(child => collectIds(child));
    }
    
    orgTree.forEach(node => collectIds(node));
    expandedNodes = allIds;
  }
  
  function collapseAll() {
    expandedNodes = new Set();
  }
  
  $: filteredOrgTree = searchHierarchy ? filterHierarchy(orgTree, searchHierarchy) : orgTree;
  
  function filterHierarchy(nodes: OrgNode[], query: string): OrgNode[] {
    const lowerQuery = query.toLowerCase();
    
    return nodes
      .map(node => {
        const matchesSearch = 
          `${node.employee.first_name} ${node.employee.last_name}`.toLowerCase().includes(lowerQuery) ||
          node.employee.position?.toLowerCase().includes(lowerQuery) ||
          node.employee.department?.toLowerCase().includes(lowerQuery);
        
        const filteredChildren = filterHierarchy(node.directReports, query);
        
        if (matchesSearch || filteredChildren.length > 0) {
          return {
            ...node,
            directReports: filteredChildren
          };
        }
        
        return null;
      })
      .filter(Boolean) as OrgNode[];
  }
  
  // Recursive hierarchy rendering function
  function renderHierarchyNode(node: OrgNode, level: number = 0): string {
    if (!node || !expandedNodes) return '';
    
    const isExpanded = expandedNodes.has(node.employee.id);
    const hasChildren = node.directReports && node.directReports.length > 0;
    const reportCount = employees.filter(e => e.manager_id === node.employee.id).length;
    
    let html = `
      <div class="hierarchy-node" style="margin-left: ${level * 32}px">
        <div class="hierarchy-node-content" data-node-id="${node.employee.id}">
          ${hasChildren ? `
            <button class="expand-btn" data-toggle="${node.employee.id}">
              ${isExpanded ? '▼' : '▶'}
            </button>
          ` : '<span class="expand-spacer"></span>'}
          
          <div class="hierarchy-employee">
            <div class="employee-avatar small">
              ${node.employee.first_name[0]}${node.employee.last_name[0]}
            </div>
            <div class="hierarchy-info">
              <div class="hierarchy-name">
                ${node.employee.first_name} ${node.employee.last_name}
              </div>
              <div class="hierarchy-details">
                ${node.employee.position || 'N/A'} • ${node.employee.department || 'N/A'}
                ${hasChildren ? ` • ${reportCount} reports` : ''}
              </div>
            </div>
          </div>
        </div>
    `;
    
    if (isExpanded && hasChildren) {
      html += '<div class="hierarchy-children">';
      node.directReports.forEach(child => {
        html += renderHierarchyNode(child, level + 1);
      });
      html += '</div>';
    }
    
    html += '</div>';
    
    return html;
  }
  
  // Handler for hierarchy node clicks
  function handleHierarchyClick(event: MouseEvent) {
    const target = event.target as HTMLElement;
    const toggleBtn = target.closest('.expand-btn') as HTMLElement;
    
    if (toggleBtn) {
      const nodeId = toggleBtn.getAttribute('data-toggle');
      if (nodeId) {
        toggleNode(nodeId);
      }
    }
  }
  
  // Helper functions
  function formatDate(date: string) {
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }
  
  function getStatusColor(status: string) {
    const colors: Record<string, string> = {
      'active': 'green',
      'pending': 'yellow',
      'inactive': 'gray'
    };
    return colors[status.toLowerCase()] || 'gray';
  }
  
  function getEmployeeName(emp: Employee) {
    return `${emp.first_name} ${emp.last_name}`;
  }
  
  function getManagerName(managerId: string | null) {
    if (!managerId) return 'None';
    const manager = employees.find(e => e.id === managerId);
    return manager ? getEmployeeName(manager) : 'Unknown';
  }
  
  function getDirectReportCount(employeeId: string) {
    return employees.filter(e => e.manager_id === employeeId).length;
  }
</script>

<div class="org-dashboard">
  <!-- Header -->
  <div class="dashboard-header">
    <div>
      <h1>Organization Management</h1>
      <p>Manage employees, reporting structures, and organizational hierarchy</p>
    </div>
    <button class="btn-refresh" on:click={loadEmployees}>
      🔄 Refresh
    </button>
  </div>

  <!-- Tabs -->
  <div class="tabs">
    <button 
      class="tab"
      class:active={activeTab === 'view'}
      on:click={() => activeTab = 'view'}
    >
      <span class="tab-icon">👥</span>
      View All Employees
      <span class="tab-badge">{employees.length}</span>
    </button>
    <button 
      class="tab"
      class:active={activeTab === 'manage'}
      on:click={() => activeTab = 'manage'}
    >
      <span class="tab-icon">⚙️</span>
      Employee Management
    </button>
    <button 
      class="tab"
      class:active={activeTab === 'hierarchy'}
      on:click={() => activeTab = 'hierarchy'}
    >
      <span class="tab-icon">🏢</span>
      Organization Hierarchy
    </button>
  </div>

  <!-- Loading State -->
  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Loading organization data...</p>
    </div>
  {:else if error}
    <div class="error-state">
      <span class="error-icon">⚠️</span>
      <h3>
        {#if error.includes('Authentication') || error.includes('authenticated')}
          Authentication Required
        {:else}
          Error Loading Data
        {/if}
      </h3>
      <p>{error}</p>
      
      {#if error.includes('Authentication') || error.includes('authenticated')}
        <div class="error-help">
          <p><strong>Why am I seeing this?</strong></p>
          <ul>
            <li>Your session may have expired</li>
            <li>You need to log in to access employee data</li>
            <li>Your authentication token is missing or invalid</li>
          </ul>
          <p><strong>What should I do?</strong></p>
          <ul>
            <li>Return to the login page and sign in again</li>
            <li>Check that you're accessing the correct URL</li>
            <li>Clear your browser cache and cookies if the problem persists</li>
          </ul>
        </div>
        <div class="error-actions">
          <button class="btn-primary" on:click={() => window.location.href = '/login'}>
            Go to Login
          </button>
          <button class="btn-secondary" on:click={loadEmployees}>
            Try Again
          </button>
        </div>
      {:else}
        <button class="btn-primary" on:click={loadEmployees}>Try Again</button>
      {/if}
    </div>
  {:else}
    <!-- Tab Content -->
    <div class="tab-content">
      <!-- TAB 1: View All Employees -->
      {#if activeTab === 'view'}
        <div class="view-all-tab">
          <!-- Filters and Search -->
          <div class="filters-section">
            <div class="search-box">
              <span class="search-icon">🔍</span>
              <input
                type="text"
                class="search-input"
                placeholder="Search by name, email, or position..."
                bind:value={searchQuery}
              />
            </div>
            
            <select class="filter-select" bind:value={filterDepartment}>
              <option value="">All Departments</option>
              {#each departments as dept}
                <option value={dept}>{dept}</option>
              {/each}
            </select>
            
            <select class="filter-select" bind:value={filterStatus}>
              <option value="">All Statuses</option>
              <option value="active">Active</option>
              <option value="pending">Pending</option>
              <option value="inactive">Inactive</option>
            </select>
            
            <select class="filter-select" bind:value={sortBy}>
              <option value="name">Sort by Name</option>
              <option value="department">Sort by Department</option>
              <option value="hire_date">Sort by Hire Date</option>
            </select>
            
            <button 
              class="btn-sort-direction"
              on:click={() => sortDirection = sortDirection === 'asc' ? 'desc' : 'asc'}
            >
              {sortDirection === 'asc' ? '↑' : '↓'}
            </button>
          </div>

          <!-- Results Count -->
          <div class="results-info">
            Showing <strong>{(currentPage - 1) * itemsPerPage + 1}</strong> to <strong>{Math.min(currentPage * itemsPerPage, filteredEmployees.length)}</strong> of <strong>{filteredEmployees.length}</strong> employees
          </div>

          <!-- Employee Grid -->
          <div class="employee-grid">
            {#each paginatedEmployees as employee}
              <div class="employee-card">
                <div class="employee-card-header">
                  <div class="employee-avatar">
                    {employee.first_name[0]}{employee.last_name[0]}
                  </div>
                  <div class="employee-info">
                    <h3>{getEmployeeName(employee)}</h3>
                    <p class="employee-email">{employee.email}</p>
                  </div>
                  <span class="status-badge {getStatusColor(employee.status)}">
                    {employee.status}
                  </span>
                </div>
                
                <div class="employee-card-body">
                  <div class="info-row">
                    <span class="info-label">Position:</span>
                    <span class="info-value">{employee.position || 'N/A'}</span>
                  </div>
                  <div class="info-row">
                    <span class="info-label">Department:</span>
                    <span class="info-value">{employee.department || 'N/A'}</span>
                  </div>
                  <div class="info-row">
                    <span class="info-label">Manager:</span>
                    <span class="info-value">{getManagerName(employee.manager_id)}</span>
                  </div>
                  <div class="info-row">
                    <span class="info-label">Hire Date:</span>
                    <span class="info-value">{formatDate(employee.hire_date)}</span>
                  </div>
                  <div class="info-row">
                    <span class="info-label">Direct Reports:</span>
                    <span class="info-value">{getDirectReportCount(employee.id)}</span>
                  </div>
                </div>
                
                <div class="employee-card-footer">
                  <button 
                    class="btn-small"
                    on:click={() => openEmployeeDetail(employee.id)}
                    type="button"
                  >
                    Employee Details
                  </button>
                </div>
              </div>
            {/each}
          </div>

          <!-- Pagination Controls -->
          {#if filteredEmployees.length > 0}
            <div class="pagination-container">
              <div class="pagination-info">
                <span class="pagination-text">
                  Showing {(currentPage - 1) * itemsPerPage + 1} to {Math.min(currentPage * itemsPerPage, filteredEmployees.length)} of {filteredEmployees.length}
                </span>
                
                <div class="pagination-per-page">
                  <label for="items-per-page">Per page:</label>
                  <select 
                    id="items-per-page"
                    class="per-page-select"
                    bind:value={itemsPerPage}
                    on:change={() => currentPage = 1}
                  >
                    <option value={6}>6</option>
                    <option value={12}>12</option>
                    <option value={24}>24</option>
                    <option value={50}>50</option>
                  </select>
                </div>
              </div>

              {#if totalPages > 1}
                <div class="pagination-controls">
                  <button 
                    class="pagination-btn"
                    disabled={currentPage === 1}
                    on:click={() => goToPage(1)}
                    title="First page"
                  >
                    ««
                  </button>
                  <button 
                    class="pagination-btn"
                    disabled={currentPage === 1}
                    on:click={() => goToPage(currentPage - 1)}
                    title="Previous page"
                  >
                    ‹
                  </button>
                  
                  <div class="pagination-center">
                    <span class="pagination-page-info">
                      Page
                    </span>
                    <select 
                      class="page-select"
                      bind:value={currentPage}
                    >
                      {#each Array.from({length: totalPages}, (_, i) => i + 1) as page}
                        <option value={page}>{page}</option>
                      {/each}
                    </select>
                    <span class="pagination-page-info">
                      of {totalPages}
                    </span>
                  </div>
                  
                  <button 
                    class="pagination-btn"
                    disabled={currentPage === totalPages}
                    on:click={() => goToPage(currentPage + 1)}
                    title="Next page"
                  >
                    ›
                  </button>
                  <button 
                    class="pagination-btn"
                    disabled={currentPage === totalPages}
                    on:click={() => goToPage(totalPages)}
                    title="Last page"
                  >
                    »»
                  </button>
                </div>
              {/if}
            </div>
          {/if}

          {#if filteredEmployees.length === 0}
            <div class="empty-state">
              <span class="empty-icon">🔍</span>
              <h3>No Employees Found</h3>
              <p>Try adjusting your search or filters</p>
            </div>
          {/if}
        </div>

      <!-- TAB 2: Employee Management -->
      {:else if activeTab === 'manage'}
        <div class="manage-tab">
          <div class="manage-header">
            <h2>Employee-Manager Assignments</h2>
            <p>Assign or change reporting relationships for employees</p>
          </div>

          <!-- Employee List for Management -->
          <div class="manage-search">
            <input
              type="text"
              class="search-input"
              placeholder="🔍 Search employees..."
              bind:value={searchQuery}
            />
          </div>

          <div class="manage-table-container">
            <table class="manage-table">
              <thead>
                <tr>
                  <th>Employee</th>
                  <th>Position</th>
                  <th>Department</th>
                  <th>Current Manager</th>
                  <th>Direct Reports</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredEmployees as employee}
                  <tr>
                    <td>
                      <div class="table-employee">
                        <div class="employee-avatar small">
                          {employee.first_name[0]}{employee.last_name[0]}
                        </div>
                        <div>
                          <div class="employee-name">{getEmployeeName(employee)}</div>
                          <div class="employee-email">{employee.email}</div>
                        </div>
                      </div>
                    </td>
                    <td>{employee.position || 'N/A'}</td>
                    <td>{employee.department || 'N/A'}</td>
                    <td>
                      <span class="manager-badge">
                        {getManagerName(employee.manager_id)}
                      </span>
                    </td>
                    <td>
                      <span class="report-count">
                        {getDirectReportCount(employee.id)}
                      </span>
                    </td>
                    <td>
                      <button 
                        class="btn-action"
                        on:click={() => openAssignModal(employee)}
                      >
                        {employee.manager_id ? 'Change' : 'Assign'} Manager
                      </button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>

      <!-- TAB 3: Organization Hierarchy -->
      {:else if activeTab === 'hierarchy'}
        <div class="hierarchy-tab">
          <div class="hierarchy-header">
            <div>
              <h2>Organization Hierarchy</h2>
              <p>Visual representation of reporting structure</p>
            </div>
            <div class="hierarchy-actions">
              <input
                type="text"
                class="search-input-small"
                placeholder="🔍 Search..."
                bind:value={searchHierarchy}
              />
              <button class="btn-secondary-small" on:click={expandAll}>
                Expand All
              </button>
              <button class="btn-secondary-small" on:click={collapseAll}>
                Collapse All
              </button>
            </div>
          </div>

          <div class="hierarchy-tree" on:click={handleHierarchyClick}>
            {#if filteredOrgTree.length === 0}
              <div class="empty-state">
                <span class="empty-icon">🏢</span>
                <h3>No Results Found</h3>
                <p>Try adjusting your search</p>
              </div>
            {:else}
              {@html filteredOrgTree.map(node => renderHierarchyNode(node, 0)).join('')}
            {/if}
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Assign Manager Modal -->
{#if showAssignModal && selectedEmployee}
  <div class="modal-overlay" on:click={() => showAssignModal = false}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h2>Assign Manager</h2>
        <button class="close-btn" on:click={() => showAssignModal = false}>✕</button>
      </div>
      
      <div class="modal-body">
        <div class="employee-summary">
          <div class="employee-avatar large">
            {selectedEmployee.first_name[0]}{selectedEmployee.last_name[0]}
          </div>
          <div>
            <h3>{getEmployeeName(selectedEmployee)}</h3>
            <p>{selectedEmployee.position} • {selectedEmployee.department}</p>
          </div>
        </div>

        <div class="form-group">
          <label for="manager-select">Select Manager</label>
          <select 
            id="manager-select"
            class="form-select"
            bind:value={selectedManagerId}
          >
            <option value="">No Manager (Top Level)</option>
            {#each managers.filter(m => m.id !== selectedEmployee.id) as manager}
              <option value={manager.id}>
                {getEmployeeName(manager)} - {manager.position}
              </option>
            {/each}
          </select>
          <p class="form-help">
            Select a manager for this employee or leave blank for top-level position
          </p>
        </div>

        {#if assignmentSuccess}
          <div class="alert alert-success">
            ✓ {assignmentSuccess}
          </div>
        {/if}

        {#if assignmentError}
          <div class="alert alert-error">
            ⚠ {assignmentError}
          </div>
        {/if}
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" on:click={() => showAssignModal = false}>
          Cancel
        </button>
        <button class="btn-primary" on:click={assignManager}>
          {selectedManagerId ? 'Assign Manager' : 'Remove Manager'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Employee Detail Modal -->
{#if showEmployeeDetail && selectedEmployeeId}
  <div class="modal-overlay" on:click={closeEmployeeDetail}>
    <div class="modal-wrapper" on:click|stopPropagation>
      <button class="modal-close-btn" on:click={closeEmployeeDetail} title="Close">
        ✕
      </button>
      <EmployeeDetail 
        employeeId={selectedEmployeeId}
      />
    </div>
  </div>
{/if}

<style>
  .org-dashboard {
    padding: 24px;
    max-width: 1600px;
    margin: 0 auto;
  }

  /* Header */
  .dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 32px;
  }

  .dashboard-header h1 {
    font-size: 32px;
    font-weight: 700;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .dashboard-header p {
    font-size: 16px;
    color: #64748b;
    margin: 0;
  }

  .btn-refresh {
    padding: 10px 20px;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    color: #475569;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-refresh:hover {
    background: #f8fafc;
    border-color: #cbd5e0;
  }

  /* Tabs */
  .tabs {
    display: flex;
    gap: 8px;
    border-bottom: 2px solid #e2e8f0;
    margin-bottom: 32px;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 14px 24px;
    background: none;
    border: none;
    border-bottom: 3px solid transparent;
    font-size: 15px;
    font-weight: 600;
    color: #64748b;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
    bottom: -2px;
  }

  .tab:hover {
    color: #4f46e5;
    background: #f8fafc;
  }

  .tab.active {
    color: #4f46e5;
    border-bottom-color: #4f46e5;
    background: #f8fafc;
  }

  .tab-icon {
    font-size: 18px;
  }

  .tab-badge {
    display: inline-block;
    padding: 2px 8px;
    background: #e0e7ff;
    color: #4f46e5;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 700;
  }

  /* Loading & Error States */
  .loading-state,
  .error-state {
    text-align: center;
    padding: 80px 20px;
    max-width: 800px;
    margin: 0 auto;
  }

  .error-help {
    text-align: left;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 20px;
    margin: 24px 0;
  }

  .error-help p {
    margin: 16px 0 8px 0;
    font-size: 14px;
    font-weight: 600;
    color: #1a202c;
  }

  .error-help ul {
    margin: 8px 0;
    padding-left: 24px;
  }

  .error-help li {
    margin: 6px 0;
    font-size: 14px;
    color: #64748b;
    line-height: 1.6;
  }

  .error-actions {
    display: flex;
    gap: 12px;
    justify-content: center;
    flex-wrap: wrap;
  }

  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid #e2e8f0;
    border-top-color: #4f46e5;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 24px;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error-icon {
    font-size: 64px;
    display: block;
    margin-bottom: 16px;
  }

  .error-state h3 {
    font-size: 24px;
    font-weight: 600;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .error-state p {
    font-size: 16px;
    color: #64748b;
    margin: 0 0 24px 0;
  }

  /* Tab Content */
  .tab-content {
    animation: fadeIn 0.3s ease-in;
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  /* View All Tab */
  .filters-section {
    display: flex;
    gap: 12px;
    margin-bottom: 24px;
    flex-wrap: wrap;
  }

  .search-box {
    position: relative;
    flex: 1;
    min-width: 300px;
  }

  .search-icon {
    position: absolute;
    left: 14px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 16px;
    color: #94a3b8;
  }

  .search-input {
    width: 100%;
    padding: 12px 14px 12px 42px;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    font-size: 14px;
    color: #1e293b;
    transition: all 0.2s;
  }

  .search-input:focus {
    outline: none;
    border-color: #4f46e5;
    box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
  }

  .filter-select {
    padding: 12px 14px;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    font-size: 14px;
    color: #1e293b;
    background: white;
    cursor: pointer;
    transition: all 0.2s;
  }

  .filter-select:focus {
    outline: none;
    border-color: #4f46e5;
  }

  .btn-sort-direction {
    width: 48px;
    padding: 12px;
    background: white;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    font-size: 18px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-sort-direction:hover {
    background: #f8fafc;
    border-color: #4f46e5;
  }

  .results-info {
    font-size: 14px;
    color: #64748b;
    margin-bottom: 20px;
  }

  /* Employee Grid */
  .employee-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 20px;
  }

  .employee-card {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 12px;
    padding: 20px;
    transition: all 0.2s;
  }

  .employee-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
    border-color: #cbd5e0;
  }

  .employee-card-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid #e2e8f0;
  }

  .employee-avatar {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    font-size: 16px;
    flex-shrink: 0;
  }

  .employee-avatar.small {
    width: 36px;
    height: 36px;
    font-size: 13px;
  }

  .employee-avatar.large {
    width: 64px;
    height: 64px;
    font-size: 24px;
  }

  .employee-info {
    flex: 1;
  }

  .employee-info h3 {
    font-size: 16px;
    font-weight: 600;
    color: #1a202c;
    margin: 0 0 4px 0;
  }

  .employee-email {
    font-size: 13px;
    color: #64748b;
  }

  .status-badge {
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 600;
    text-transform: capitalize;
  }

  .status-badge.green {
    background: #d1fae5;
    color: #065f46;
  }

  .status-badge.yellow {
    background: #fef3c7;
    color: #92400e;
  }

  .status-badge.gray {
    background: #f1f5f9;
    color: #475569;
  }

  .employee-card-body {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 16px;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 14px;
  }

  .info-label {
    color: #64748b;
    font-weight: 500;
  }

  .info-value {
    color: #1e293b;
    font-weight: 600;
    text-align: right;
  }

  .employee-card-footer {
    display: flex;
    gap: 8px;
    padding-top: 16px;
    border-top: 1px solid #e2e8f0;
  }

  .btn-small {
    flex: 1;
    padding: 8px 16px;
    background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-small:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
  }

  /* Employee Management Tab */
  .manage-tab {
    background: white;
    border-radius: 12px;
    padding: 24px;
  }

  .manage-header h2 {
    font-size: 24px;
    font-weight: 700;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .manage-header p {
    font-size: 14px;
    color: #64748b;
    margin: 0 0 24px 0;
  }

  .manage-search {
    margin-bottom: 24px;
  }

  .manage-table-container {
    overflow-x: auto;
  }

  .manage-table {
    width: 100%;
    border-collapse: collapse;
  }

  .manage-table thead {
    background: #f8fafc;
  }

  .manage-table th {
    padding: 12px 16px;
    text-align: left;
    font-size: 12px;
    font-weight: 600;
    color: #475569;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    border-bottom: 2px solid #e2e8f0;
  }

  .manage-table td {
    padding: 16px;
    border-bottom: 1px solid #e2e8f0;
    font-size: 14px;
    color: #1e293b;
  }

  .manage-table tbody tr:hover {
    background: #f8fafc;
  }

  .table-employee {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .employee-name {
    font-weight: 600;
    color: #1a202c;
  }

  .manager-badge {
    display: inline-block;
    padding: 4px 12px;
    background: #e0e7ff;
    color: #4f46e5;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
  }

  .report-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    background: #f1f5f9;
    color: #475569;
    border-radius: 50%;
    font-weight: 600;
  }

  .btn-action {
    padding: 8px 16px;
    background: white;
    color: #4f46e5;
    border: 1px solid #4f46e5;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-action:hover {
    background: #4f46e5;
    color: white;
  }

  /* Hierarchy Tab */
  .hierarchy-tab {
    background: white;
    border-radius: 12px;
    padding: 24px;
  }

  .hierarchy-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 32px;
  }

  .hierarchy-header h2 {
    font-size: 24px;
    font-weight: 700;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .hierarchy-header p {
    font-size: 14px;
    color: #64748b;
    margin: 0;
  }

  .hierarchy-actions {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .search-input-small {
    padding: 8px 12px;
    border: 1px solid #cbd5e0;
    border-radius: 6px;
    font-size: 13px;
    width: 200px;
  }

  .btn-secondary-small {
    padding: 8px 16px;
    background: white;
    border: 1px solid #cbd5e0;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    color: #475569;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-secondary-small:hover {
    background: #f8fafc;
    border-color: #4f46e5;
    color: #4f46e5;
  }

  .hierarchy-tree {
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 20px;
  }

  .hierarchy-node {
    margin-bottom: 8px;
  }

  .hierarchy-node-content {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    transition: all 0.2s;
  }

  .hierarchy-node-content:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    border-color: #cbd5e0;
  }

  :global(.expand-btn) {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: none;
    border: none;
    color: #64748b;
    cursor: pointer;
    font-size: 12px;
    transition: all 0.2s;
  }

  :global(.expand-btn:hover) {
    color: #4f46e5;
  }

  :global(.expand-spacer) {
    width: 24px;
    display: inline-block;
  }

  :global(.hierarchy-employee) {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
  }

  :global(.hierarchy-info) {
    flex: 1;
  }

  :global(.hierarchy-name) {
    font-size: 15px;
    font-weight: 600;
    color: #1a202c;
    margin-bottom: 4px;
  }

  :global(.hierarchy-details) {
    font-size: 13px;
    color: #64748b;
  }

  :global(.hierarchy-children) {
    margin-top: 8px;
  }

  /* Modal */
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
    padding: 20px;
    overflow-y: auto;
  }

  .modal-wrapper {
    position: relative;
    background: white;
    border-radius: 12px;
    max-width: 1200px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  }

  .modal-close-btn {
    position: absolute;
    top: 16px;
    right: 16px;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: white;
    border: 1px solid #e2e8f0;
    font-size: 24px;
    color: #64748b;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10;
    transition: all 0.2s;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .modal-close-btn:hover {
    background: #f8fafc;
    color: #1e293b;
    border-color: #cbd5e0;
    transform: scale(1.05);
  }

  .modal {
    background: white;
    border-radius: 12px;
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e2e8f0;
  }

  .modal-header h2 {
    font-size: 20px;
    font-weight: 700;
    color: #1a202c;
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 24px;
    color: #94a3b8;
    cursor: pointer;
    transition: color 0.2s;
  }

  .close-btn:hover {
    color: #1e293b;
  }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .employee-summary {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    background: #f8fafc;
    border-radius: 12px;
    margin-bottom: 24px;
  }

  .employee-summary h3 {
    margin: 0 0 4px 0;
    font-size: 18px;
    font-weight: 600;
    color: #1a202c;
  }

  .employee-summary p {
    margin: 0;
    color: #64748b;
    font-size: 14px;
  }

  .form-group {
    margin-bottom: 20px;
  }

  .form-group label {
    display: block;
    font-size: 14px;
    font-weight: 600;
    color: #1a202c;
    margin-bottom: 8px;
  }

  .form-select {
    width: 100%;
    padding: 12px;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    font-size: 14px;
    color: #1e293b;
    background: white;
    cursor: pointer;
  }

  .form-select:focus {
    outline: none;
    border-color: #4f46e5;
    box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
  }

  .form-help {
    margin: 8px 0 0 0;
    font-size: 13px;
    color: #64748b;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 24px;
    border-top: 1px solid #e2e8f0;
  }

  .btn-primary {
    padding: 12px 24px;
    background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(79, 70, 229, 0.3);
  }

  .btn-secondary {
    padding: 12px 24px;
    background: white;
    color: #475569;
    border: 1px solid #cbd5e0;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-secondary:hover {
    background: #f8fafc;
    border-color: #94a3b8;
  }

  /* Alerts */
  .alert {
    padding: 12px 16px;
    border-radius: 8px;
    font-size: 14px;
    margin-top: 16px;
  }

  .alert-success {
    background: #d1fae5;
    color: #065f46;
    border: 1px solid #6ee7b7;
  }

  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fca5a5;
  }

  .alert-info {
    background: #dbeafe;
    color: #1e40af;
    border: 1px solid #93c5fd;
  }

  /* Employee Detail Fallback Modal */
  /* Empty State */
  .empty-state {
    text-align: center;
    padding: 60px 20px;
  }

  .empty-icon {
    font-size: 64px;
    display: block;
    margin-bottom: 16px;
  }

  .empty-state h3 {
    font-size: 20px;
    font-weight: 600;
    color: #1a202c;
    margin: 0 0 8px 0;
  }

  .empty-state p {
    font-size: 14px;
    color: #64748b;
    margin: 0;
  }

  /* Pagination */
  .pagination-container {
    margin-top: 24px;
    padding: 20px 0;
    border-top: 1px solid #e2e8f0;
  }

  .pagination-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    flex-wrap: wrap;
    gap: 12px;
  }

  .pagination-text {
    font-size: 14px;
    color: #64748b;
  }

  .pagination-per-page {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: #64748b;
  }

  .per-page-select {
    padding: 6px 12px;
    border: 1px solid #cbd5e0;
    border-radius: 6px;
    font-size: 14px;
    color: #1e293b;
    background: white;
    cursor: pointer;
  }

  .per-page-select:focus {
    outline: none;
    border-color: #4f46e5;
  }

  .pagination-controls {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 8px;
  }

  .pagination-btn {
    min-width: 36px;
    height: 36px;
    padding: 6px 12px;
    background: white;
    border: 1px solid #cbd5e0;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 600;
    color: #475569;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .pagination-btn:hover:not(:disabled) {
    background: #f8fafc;
    border-color: #4f46e5;
    color: #4f46e5;
  }

  .pagination-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .pagination-center {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 12px;
  }

  .pagination-page-info {
    font-size: 14px;
    color: #64748b;
    font-weight: 500;
  }

  .page-select {
    padding: 6px 12px;
    border: 1px solid #cbd5e0;
    border-radius: 6px;
    font-size: 14px;
    color: #1e293b;
    background: white;
    cursor: pointer;
    min-width: 60px;
  }

  .page-select:focus {
    outline: none;
    border-color: #4f46e5;
    box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
  }

  /* Responsive */
  @media (max-width: 768px) {
    .org-dashboard {
      padding: 16px;
    }

    .dashboard-header {
      flex-direction: column;
      gap: 16px;
    }

    .tabs {
      flex-direction: column;
    }

    .tab {
      justify-content: center;
    }

    .filters-section {
      flex-direction: column;
    }

    .search-box {
      min-width: 100%;
    }

    .employee-grid {
      grid-template-columns: 1fr;
    }

    .hierarchy-actions {
      flex-direction: column;
      align-items: stretch;
    }

    .search-input-small {
      width: 100%;
    }

    .manage-table-container {
      overflow-x: scroll;
    }

    .manage-table {
      min-width: 800px;
    }

    .pagination-info {
      flex-direction: column;
      align-items: flex-start;
    }

    .pagination-controls {
      flex-wrap: wrap;
      gap: 6px;
    }

    .pagination-btn {
      min-width: 32px;
      height: 32px;
      padding: 4px 8px;
      font-size: 13px;
    }

    .pagination-center {
      padding: 0 8px;
    }
  }
</style>