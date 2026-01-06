import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { userStore } from '$stores/userStore';
import { authStore } from '$stores/authStore';
import { timesheetStore } from '$stores/timesheetStore';

describe('userStore', () => {
  beforeEach(() => {
    userStore.reset();
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('initializes with empty state', () => {
    const state = get(userStore);
    expect(state.users).toEqual([]);
    expect(state.loading).toBe(false);
    expect(state.error).toBeNull();
  });

  it('loads users successfully', async () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', first_name: 'John', last_name: 'Doe' },
      { id: '2', email: 'user2@example.com', first_name: 'Jane', last_name: 'Smith' }
    ];

    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => ({ users: mockUsers, total: 2 })
    });

    await userStore.loadUsers();

    const state = get(userStore);
    expect(state.users).toEqual(mockUsers);
    expect(state.loading).toBe(false);
    expect(state.error).toBeNull();
  });

  it('handles load users error', async () => {
    global.fetch = vi.fn().mockRejectedValueOnce(new Error('Network error'));

    await userStore.loadUsers();

    const state = get(userStore);
    expect(state.users).toEqual([]);
    expect(state.loading).toBe(false);
    expect(state.error).toBe('Failed to load users');
  });

  it('creates user successfully', async () => {
    const newUser = {
      email: 'new@example.com',
      first_name: 'New',
      last_name: 'User',
      password: 'Pass123!'
    };

    const createdUser = { id: '3', ...newUser };

    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => createdUser
    });

    await userStore.createUser(newUser);

    const state = get(userStore);
    expect(state.users).toContainEqual(createdUser);
  });

  it('updates user successfully', async () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', first_name: 'John', last_name: 'Doe' }
    ];

    userStore.set({ users: mockUsers, loading: false, error: null });

    const updatedData = { first_name: 'Updated', last_name: 'Name' };
    const updatedUser = { id: '1', email: 'user1@example.com', ...updatedData };

    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => updatedUser
    });

    await userStore.updateUser('1', updatedData);

    const state = get(userStore);
    expect(state.users[0].first_name).toBe('Updated');
  });

  it('deletes user successfully', async () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', first_name: 'John', last_name: 'Doe' },
      { id: '2', email: 'user2@example.com', first_name: 'Jane', last_name: 'Smith' }
    ];

    userStore.set({ users: mockUsers, loading: false, error: null });

    global.fetch = vi.fn().mockResolvedValueOnce({ ok: true });

    await userStore.deleteUser('1');

    const state = get(userStore);
    expect(state.users).toHaveLength(1);
    expect(state.users[0].id).toBe('2');
  });

  it('filters users by search term', () => {
    const mockUsers = [
      { id: '1', email: 'john@example.com', first_name: 'John', last_name: 'Doe' },
      { id: '2', email: 'jane@example.com', first_name: 'Jane', last_name: 'Smith' }
    ];

    userStore.set({ users: mockUsers, loading: false, error: null });

    const filtered = userStore.filterUsers('john');
    expect(filtered).toHaveLength(1);
    expect(filtered[0].first_name).toBe('John');
  });
});

describe('authStore', () => {
  beforeEach(() => {
    authStore.reset();
    localStorage.clear();
    global.fetch = vi.fn();
  });

  it('initializes with unauthenticated state', () => {
    const state = get(authStore);
    expect(state.user).toBeNull();
    expect(state.token).toBeNull();
    expect(state.isAuthenticated).toBe(false);
  });

  it('logs in successfully', async () => {
    const mockResponse = {
      token: 'test-token',
      user: { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User' }
    };

    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse
    });

    await authStore.login('test@example.com', 'password');

    const state = get(authStore);
    expect(state.user).toEqual(mockResponse.user);
    expect(state.token).toBe('test-token');
    expect(state.isAuthenticated).toBe(true);
    expect(localStorage.getItem('auth_token')).toBe('test-token');
  });

  it('handles login error', async () => {
    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: async () => ({ error: 'Invalid credentials' })
    });

    await expect(authStore.login('test@example.com', 'wrong')).rejects.toThrow('Invalid credentials');

    const state = get(authStore);
    expect(state.isAuthenticated).toBe(false);
  });

  it('logs out successfully', () => {
    authStore.set({
      user: { id: '1', email: 'test@example.com' },
      token: 'test-token',
      isAuthenticated: true,
      loading: false,
      error: null
    });

    authStore.logout();

    const state = get(authStore);
    expect(state.user).toBeNull();
    expect(state.token).toBeNull();
    expect(state.isAuthenticated).toBe(false);
    expect(localStorage.getItem('auth_token')).toBeNull();
  });

  it('restores session from localStorage', async () => {
    localStorage.setItem('auth_token', 'stored-token');

    const mockUser = { id: '1', email: 'test@example.com', first_name: 'Test', last_name: 'User' };

    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => mockUser
    });

    await authStore.restoreSession();

    const state = get(authStore);
    expect(state.user).toEqual(mockUser);
    expect(state.token).toBe('stored-token');
    expect(state.isAuthenticated).toBe(true);
  });

  it('clears invalid session', async () => {
    localStorage.setItem('auth_token', 'invalid-token');

    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: false,
      status: 401
    });

    await authStore.restoreSession();

    const state = get(authStore);
    expect(state.isAuthenticated).toBe(false);
    expect(localStorage.getItem('auth_token')).toBeNull();
  });
});

describe('timesheetStore', () => {
  beforeEach(() => {
    timesheetStore.reset();
    global.fetch = vi.fn();
  });

  it('loads timesheet entries', async () => {
    const mockEntries = [
      { id: '1', date: '2024-01-15', hours: 8, description: 'Work', project_id: 'proj-1' },
      { id: '2', date: '2024-01-16', hours: 7.5, description: 'Meeting', project_id: 'proj-2' }
    ];

    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => mockEntries
    });

    await timesheetStore.loadEntries('2024-01-01', '2024-01-31');

    const state = get(timesheetStore);
    expect(state.entries).toEqual(mockEntries);
  });

  it('creates timesheet entry', async () => {
    const newEntry = {
      date: '2024-01-17',
      hours: 8,
      description: 'Development',
      project_id: 'proj-1'
    };

    const createdEntry = { id: '3', ...newEntry };

    global.fetch = vi.fn().mockResolvedValueOnce({
      ok: true,
      json: async () => createdEntry
    });

    await timesheetStore.createEntry(newEntry);

    const state = get(timesheetStore);
    expect(state.entries).toContainEqual(createdEntry);
  });

  it('calculates total hours', () => {
    const mockEntries = [
      { id: '1', date: '2024-01-15', hours: 8, description: 'Work' },
      { id: '2', date: '2024-01-16', hours: 7.5, description: 'Work' }
    ];

    timesheetStore.set({ entries: mockEntries, loading: false, error: null });

    const total = timesheetStore.getTotalHours();
    expect(total).toBe(15.5);
  });

  it('groups entries by week', () => {
    const mockEntries = [
      { id: '1', date: '2024-01-15', hours: 8, description: 'Work' },
      { id: '2', date: '2024-01-16', hours: 7.5, description: 'Work' },
      { id: '3', date: '2024-01-22', hours: 8, description: 'Work' }
    ];

    timesheetStore.set({ entries: mockEntries, loading: false, error: null });

    const grouped = timesheetStore.groupByWeek();
    expect(Object.keys(grouped)).toHaveLength(2);
  });
});
