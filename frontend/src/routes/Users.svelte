<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';

  interface User {
    id: string;
    email: string;
    role: string;
    employee_id?: string;
    created_at: string;
    updated_at: string;
  }

  interface Employee {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
  }

  // State
  let users = $state<User[]>([]);
  let employees = $state<Employee[]>([]);
  let loading = $state(true);
  let error = $state('');
  let success = $state('');
  let showAddModal = $state(false);
  let showEditModal = $state(false);
  let showPasswordModal = $state(false);
  let selectedUser = $state<User | null>(null);
  let searchTerm = $state('');
  let filterRole = $state('all');

  // Form data
  let newUser = $state({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    role: 'employee',
    employee_id: ''
  });

  let editUser = $state({
    email: '',
    role: '',
    employee_id: ''
  });

  let passwordReset = $state({
    new_password: '',
    confirm_password: ''
  });

  onMount(() => {
    loadUsers();
    loadEmployees();
  });

  // API Functions
  async function loadUsers() {
    loading = true;
    error = '';
    
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/users', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!response.ok) throw new Error('Failed to load users');

      const data = await response.json();
      users = data || [];
    } catch (err: any) {
      error = err.message || 'Failed to load users';
      users = [];
    } finally {
      loading = false;
    }
  }

  async function loadEmployees() {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/employees', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const data = await response.json();
        employees = data || [];
      }
    } catch (err) {
      console.error('Failed to load employees:', err);
    }
  }

  async function createUser() {
    error = '';
    success = '';

    // Validation
    if (!newUser.username || !newUser.email || !newUser.password || !newUser.role) {
      error = 'Please fill in all required fields';
      return;
    }

    if (newUser.password !== newUser.confirmPassword) {
      error = 'Passwords do not match';
      return;
    }

    if (newUser.password.length < 8) {
      error = 'Password must be at least 8 characters';
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const userData: any = {
        username: newUser.username,
        email: newUser.email,
        password: newUser.password,
        role: newUser.role
      };

      if (newUser.employee_id) {
        userData.employee_id = newUser.employee_id;
      }

      const response = await fetch('/api/users', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(userData)
      });

      if (!response.ok) {
        const errData = await response.json();
        throw new Error(errData.error || 'Failed to create user');
      }

      success = 'User created successfully';
      showAddModal = false;
      resetNewUserForm();
      await loadUsers();

      setTimeout(() => { success = ''; }, 3000);
    } catch (err: any) {
      error = err.message || 'Failed to create user';
    }
  }

  async function updateUser() {
    error = '';
    success = '';

    if (!selectedUser) return;

    try {
      const token = localStorage.getItem('token');
      const updates: any = {};

      if (editUser.email !== selectedUser.email) {
        updates.email = editUser.email;
      }
      if (editUser.role !== selectedUser.role) {
        updates.role = editUser.role;
      }
      if (editUser.employee_id) {
        updates.employee_id = editUser.employee_id;
      }

      const response = await fetch(`/api/users/${selectedUser.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(updates)
      });

      if (!response.ok) {
        const errData = await response.json();
        throw new Error(errData.error || 'Failed to update user');
      }

      success = 'User updated successfully';
      showEditModal = false;
      selectedUser = null;
      await loadUsers();

      setTimeout(() => { success = ''; }, 3000);
    } catch (err: any) {
      error = err.message || 'Failed to update user';
    }
  }

  async function resetPassword() {
    error = '';
    success = '';

    if (!selectedUser) return;

    if (!passwordReset.new_password || !passwordReset.confirm_password) {
      error = 'Please fill in both password fields';
      return;
    }

    if (passwordReset.new_password !== passwordReset.confirm_password) {
      error = 'Passwords do not match';
      return;
    }

    if (passwordReset.new_password.length < 8) {
      error = 'Password must be at least 8 characters';
      return;
    }

    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/users/${selectedUser.id}/reset-password`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ new_password: passwordReset.new_password })
      });

      if (!response.ok) {
        const errData = await response.json();
        throw new Error(errData.error || 'Failed to reset password');
      }

      success = 'Password reset successfully';
      showPasswordModal = false;
      selectedUser = null;
      passwordReset = { new_password: '', confirm_password: '' };

      setTimeout(() => { success = ''; }, 3000);
    } catch (err: any) {
      error = err.message || 'Failed to reset password';
    }
  }

  // UI Functions
  function openAddModal() {
    resetNewUserForm();
    showAddModal = true;
  }

  function openEditModal(user: User) {
    selectedUser = user;
    editUser = {
      email: user.email,
      role: user.role,
      employee_id: user.employee_id || ''
    };
    showEditModal = true;
  }

  function openPasswordModal(user: User) {
    selectedUser = user;
    passwordReset = { new_password: '', confirm_password: '' };
    showPasswordModal = true;
  }

  function closeAddModal() {
    showAddModal = false;
    resetNewUserForm();
    error = '';
  }

  function closeEditModal() {
    showEditModal = false;
    selectedUser = null;
    error = '';
  }

  function closePasswordModal() {
    showPasswordModal = false;
    selectedUser = null;
    passwordReset = { new_password: '', confirm_password: '' };
    error = '';
  }

  function resetNewUserForm() {
    newUser = {
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
      role: 'employee',
      employee_id: ''
    };
  }

  function getRoleBadgeClass(role: string): string {
    const classes: Record<string, string> = {
      'admin': 'badge-admin',
      'hr-manager': 'badge-hr',
      'manager': 'badge-manager',
      'employee': 'badge-employee'
    };
    return classes[role] || 'badge-employee';
  }

  function getRoleDisplayName(role: string): string {
    const names: Record<string, string> = {
      'admin': 'Admin',
      'hr-manager': 'HR Manager',
      'manager': 'Manager',
      'employee': 'Employee'
    };
    return names[role] || role;
  }

  function getEmployeeName(employeeId: string): string {
    const employee = employees.find(e => e.id === employeeId);
    return employee ? `${employee.first_name} ${employee.last_name}` : 'N/A';
  }

  // Computed/Filtered data
  let filteredUsers = $derived(users.filter(user => {
    const matchesSearch = !searchTerm || 
      user.email.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesRole = filterRole === 'all' || user.role === filterRole;
    return matchesSearch && matchesRole;
  }));

  let uniqueRoles = $derived([...new Set(users.map(u => u.role))]);
</script>

<div class="users-container">
  <!-- Header -->
  <div class="page-header">
    <div class="header-left">
      <h1>User Management</h1>
      <p>Manage system users and access control</p>
    </div>
    <button class="btn-primary" onclick={openAddModal}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
        <circle cx="8.5" cy="7" r="4"></circle>
        <line x1="20" y1="8" x2="20" y2="14"></line>
        <line x1="23" y1="11" x2="17" y2="11"></line>
      </svg>
      Add User
    </button>
  </div>

  <!-- Alerts -->
  {#if error}
    <div class="alert alert-error">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"></circle>
        <line x1="12" y1="8" x2="12" y2="12"></line>
        <line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      <span>{error}</span>
    </div>
  {/if}

  {#if success}
    <div class="alert alert-success">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
        <polyline points="22 4 12 14.01 9 11.01"></polyline>
      </svg>
      <span>{success}</span>
    </div>
  {/if}

  <!-- Filters -->
  <div class="filters-bar">
    <div class="search-box">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"></circle>
        <path d="m21 21-4.35-4.35"></path>
      </svg>
      <input
        type="text"
        placeholder="Search by email..."
        bind:value={searchTerm}
      />
    </div>

    <select bind:value={filterRole}>
      <option value="all">All Roles</option>
      {#each uniqueRoles as role}
        <option value={role}>{getRoleDisplayName(role)}</option>
      {/each}
    </select>
  </div>

  <!-- Users Table -->
  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Loading users...</p>
    </div>
  {:else if filteredUsers.length === 0}
    <div class="empty-state">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
        <circle cx="9" cy="7" r="4"></circle>
        <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
        <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
      </svg>
      <h3>No users found</h3>
      <p>Try adjusting your search or filters</p>
    </div>
  {:else}
    <div class="table-container">
      <table class="users-table">
        <thead>
          <tr>
            <th>Email</th>
            <th>Role</th>
            <th>Linked Employee</th>
            <th>Created</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each filteredUsers as user}
            <tr>
              <td>
                <div class="user-email">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                    <polyline points="22,6 12,13 2,6"></polyline>
                  </svg>
                  {user.email}
                </div>
              </td>
              <td>
                <span class="role-badge {getRoleBadgeClass(user.role)}">
                  {getRoleDisplayName(user.role)}
                </span>
              </td>
              <td>
                {#if user.employee_id}
                  <span class="employee-link">{getEmployeeName(user.employee_id)}</span>
                {:else}
                  <span class="text-muted">Not linked</span>
                {/if}
              </td>
              <td>{new Date(user.created_at).toLocaleDateString()}</td>
              <td>
                <div class="action-buttons">
                  <button class="btn-icon" title="Edit user" onclick={() => openEditModal(user)}>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                      <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                    </svg>
                  </button>
                  <button class="btn-icon" title="Reset password" onclick={() => openPasswordModal(user)}>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                      <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<!-- Add User Modal -->
{#if showAddModal}
  <div class="modal-overlay" onclick={closeAddModal} onkeydown={(e) => e.key === 'Escape' && closeAddModal()} role="button" tabindex="0">
    <div class="modal" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Add New User</h2>
        <button class="btn-close" onclick={closeAddModal}>×</button>
      </div>

      <form onsubmit={(e) => { e.preventDefault(); createUser(); }}>
        <div class="modal-body">
          <div class="form-group">
          <div class="form-group">
            <label for="username">Username *</label>
            <input
              type="text"
              id="username"
              bind:value={newUser.username}
              placeholder="johndoe"
              required
            />
          </div>

            <label for="email">Email *</label>
            <input
              type="email"
              id="email"
              bind:value={newUser.email}
              placeholder="user@company.com"
              required
            />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="password">Password *</label>
              <input
                type="password"
                id="password"
                bind:value={newUser.password}
                placeholder="Min 8 characters"
                required
                minlength="8"
              />
            </div>

            <div class="form-group">
              <label for="confirmPassword">Confirm Password *</label>
              <input
                type="password"
                id="confirmPassword"
                bind:value={newUser.confirmPassword}
                placeholder="Re-enter password"
                required
                minlength="8"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="role">Role *</label>
              <select id="role" bind:value={newUser.role} required>
                <option value="employee">Employee</option>
                <option value="manager">Manager</option>
                <option value="hr-manager">HR Manager</option>
                <option value="admin">Admin</option>
              </select>
            </div>

            <div class="form-group">
              <label for="employee">Link to Employee (Optional)</label>
              <select id="employee" bind:value={newUser.employee_id}>
                <option value="">None</option>
                {#each employees as employee}
                  <option value={employee.id}>
                    {employee.first_name} {employee.last_name} - {employee.email}
                  </option>
                {/each}
              </select>
            </div>
          </div>

          <div class="form-help">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
              <line x1="12" y1="17" x2="12.01" y2="17"></line>
            </svg>
            <span>Password must be at least 8 characters. Linking to an employee is optional.</span>
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn-secondary" onclick={closeAddModal}>Cancel</button>
          <button type="submit" class="btn-primary">Create User</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Edit User Modal -->
{#if showEditModal && selectedUser}
  <div class="modal-overlay" onclick={closeEditModal} onkeydown={(e) => e.key === 'Escape' && closeEditModal()} role="button" tabindex="0">
    <div class="modal" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
      <div class="modal-header">
        <h2>Edit User</h2>
        <button class="btn-close" onclick={closeEditModal}>×</button>
      </div>

      <form onsubmit={(e) => { e.preventDefault(); updateUser(); }}>
        <div class="modal-body">
          <div class="form-group">
            <label for="edit-email">Email</label>
            <input
              type="email"
              id="edit-email"
              bind:value={editUser.email}
              required
            />
          </div>

          <div class="form-group">
            <label for="edit-role">Role</label>
            <select id="edit-role" bind:value={editUser.role} required>
              <option value="employee">Employee</option>
              <option value="manager">Manager</option>
              <option value="hr-manager">HR Manager</option>
              <option value="admin">Admin</option>
            </select>
          </div>

          <div class="form-group">
            <label for="edit-employee">Linked Employee</label>
            <select id="edit-employee" bind:value={editUser.employee_id}>
              <option value="">None</option>
              {#each employees as employee}
                <option value={employee.id}>
                  {employee.first_name} {employee.last_name} - {employee.email}
                </option>
              {/each}
            </select>
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn-secondary" onclick={closeEditModal}>Cancel</button>
          <button type="submit" class="btn-primary">Update User</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- Reset Password Modal -->
{#if showPasswordModal && selectedUser}
  <div class="modal-overlay" onclick={closePasswordModal}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2>Reset Password</h2>
        <button class="btn-close" onclick={closePasswordModal}>×</button>
      </div>

      <form onsubmit={(e) => { e.preventDefault(); resetPassword(); }}>
        <div class="modal-body">
          <div class="info-box">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="16" x2="12" y2="12"></line>
              <line x1="12" y1="8" x2="12.01" y2="8"></line>
            </svg>
            <span>Resetting password for: <strong>{selectedUser.email}</strong></span>
          </div>

          <div class="form-group">
            <label for="new-password">New Password *</label>
            <input
              type="password"
              id="new-password"
              bind:value={passwordReset.new_password}
              placeholder="Min 8 characters"
              required
              minlength="8"
            />
          </div>

          <div class="form-group">
            <label for="confirm-password">Confirm New Password *</label>
            <input
              type="password"
              id="confirm-password"
              bind:value={passwordReset.confirm_password}
              placeholder="Re-enter password"
              required
              minlength="8"
            />
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn-secondary" onclick={closePasswordModal}>Cancel</button>
          <button type="submit" class="btn-primary">Reset Password</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<style>
  .users-container {
    padding: 2rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
  }

  .header-left h1 {
    font-size: 2rem;
    font-weight: 600;
    margin: 0 0 0.5rem 0;
    color: #1a1a1a;
  }

  .header-left p {
    color: #6b7280;
    margin: 0;
  }

  .btn-primary {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 8px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary:hover {
    background: #2563eb;
    transform: translateY(-1px);
  }

  .btn-primary svg {
    width: 20px;
    height: 20px;
  }

  .alert {
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .alert svg {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
  }

  .alert-error {
    background: #fef2f2;
    color: #991b1b;
    border: 1px solid #fecaca;
  }

  .alert-success {
    background: #f0fdf4;
    color: #166534;
    border: 1px solid #bbf7d0;
  }

  .filters-bar {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .search-box {
    flex: 1;
    position: relative;
  }

  .search-box svg {
    position: absolute;
    left: 1rem;
    top: 50%;
    transform: translateY(-50%);
    width: 20px;
    height: 20px;
    color: #9ca3af;
  }

  .search-box input {
    width: 100%;
    padding: 0.75rem 1rem 0.75rem 3rem;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 0.95rem;
  }

  select {
    padding: 0.75rem 1rem;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    background: white;
    cursor: pointer;
  }

  .table-container {
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    overflow: hidden;
  }

  .users-table {
    width: 100%;
    border-collapse: collapse;
  }

  .users-table thead {
    background: #f9fafb;
  }

  .users-table th {
    padding: 1rem;
    text-align: left;
    font-weight: 600;
    color: #374151;
    border-bottom: 2px solid #e5e7eb;
  }

  .users-table td {
    padding: 1rem;
    border-bottom: 1px solid #f3f4f6;
  }

  .users-table tbody tr:hover {
    background: #f9fafb;
  }

  .user-email {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .user-email svg {
    width: 16px;
    height: 16px;
    color: #6b7280;
  }

  .role-badge {
    display: inline-block;
    padding: 0.375rem 0.75rem;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 500;
  }

  .badge-admin {
    background: #fef2f2;
    color: #991b1b;
  }

  .badge-hr {
    background: #f0f9ff;
    color: #075985;
  }

  .badge-manager {
    background: #fefce8;
    color: #854d0e;
  }

  .badge-employee {
    background: #f0fdf4;
    color: #166534;
  }

  .employee-link {
    color: #3b82f6;
    font-weight: 500;
  }

  .text-muted {
    color: #9ca3af;
  }

  .action-buttons {
    display: flex;
    gap: 0.5rem;
  }

  .btn-icon {
    padding: 0.5rem;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-icon:hover {
    background: #f9fafb;
    border-color: #3b82f6;
  }

  .btn-icon svg {
    width: 18px;
    height: 18px;
    color: #6b7280;
  }

  .loading-state, .empty-state {
    text-align: center;
    padding: 4rem 2rem;
  }

  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid #f3f4f6;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 1rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .empty-state svg {
    width: 64px;
    height: 64px;
    color: #d1d5db;
    margin-bottom: 1rem;
  }

  .empty-state h3 {
    margin: 0 0 0.5rem 0;
    color: #374151;
  }

  .empty-state p {
    color: #6b7280;
    margin: 0;
  }

  /* Modal Styles */
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
  }

  .modal {
    background: white;
    border-radius: 12px;
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.5rem;
    color: #1a1a1a;
  }

  .btn-close {
    background: none;
    border: none;
    font-size: 2rem;
    color: #6b7280;
    cursor: pointer;
    line-height: 1;
    padding: 0;
    width: 32px;
    height: 32px;
  }

  .btn-close:hover {
    color: #374151;
  }

  .modal-body {
    padding: 1.5rem;
  }

  .form-group {
    margin-bottom: 1.25rem;
  }

  .form-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: #374151;
  }

  .form-group input,
  .form-group select {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 0.95rem;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }

  .form-help {
    display: flex;
    gap: 0.5rem;
    padding: 0.75rem;
    background: #f0f9ff;
    border-radius: 8px;
    font-size: 0.875rem;
    color: #075985;
  }

  .form-help svg {
    width: 18px;
    height: 18px;
    flex-shrink: 0;
  }

  .info-box {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    background: #fefce8;
    border-radius: 8px;
    margin-bottom: 1.5rem;
  }

  .info-box svg {
    width: 20px;
    height: 20px;
    color: #854d0e;
    flex-shrink: 0;
  }

  .info-box span {
    color: #854d0e;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    padding: 1.5rem;
    border-top: 1px solid #e5e7eb;
  }

  .btn-secondary {
    padding: 0.75rem 1.5rem;
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-secondary:hover {
    background: #f9fafb;
  }

  @media (max-width: 768px) {
    .users-container {
      padding: 1rem;
    }

    .page-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 1rem;
    }

    .filters-bar {
      flex-direction: column;
    }

    .form-row {
      grid-template-columns: 1fr;
    }

    .table-container {
      overflow-x: auto;
    }

    .users-table {
      min-width: 800px;
    }
  }
</style>