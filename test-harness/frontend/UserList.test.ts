import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import UserList from '$lib/components/UserList.svelte';
import { userStore } from '$stores/userStore';

describe('UserList Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders user list with users', async () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', first_name: 'John', last_name: 'Doe', role: 'employee' },
      { id: '2', email: 'user2@example.com', first_name: 'Jane', last_name: 'Smith', role: 'manager' }
    ];

    render(UserList, { users: mockUsers });

    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    expect(screen.getByText('user1@example.com')).toBeInTheDocument();
    expect(screen.getByText('user2@example.com')).toBeInTheDocument();
  });

  it('displays empty state when no users', () => {
    render(UserList, { users: [] });

    expect(screen.getByText(/no users found/i)).toBeInTheDocument();
  });

  it('calls delete handler when delete button clicked', async () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', first_name: 'John', last_name: 'Doe', role: 'employee' }
    ];
    const onDelete = vi.fn();

    render(UserList, { users: mockUsers, onDelete });

    const deleteButton = screen.getByRole('button', { name: /delete/i });
    await fireEvent.click(deleteButton);

    expect(onDelete).toHaveBeenCalledWith('1');
  });

  it('navigates to edit page when edit button clicked', async () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', first_name: 'John', last_name: 'Doe', role: 'employee' }
    ];
    const onEdit = vi.fn();

    render(UserList, { users: mockUsers, onEdit });

    const editButton = screen.getByRole('button', { name: /edit/i });
    await fireEvent.click(editButton);

    expect(onEdit).toHaveBeenCalledWith('1');
  });

  it('filters users by search term', async () => {
    const mockUsers = [
      { id: '1', email: 'john@example.com', first_name: 'John', last_name: 'Doe', role: 'employee' },
      { id: '2', email: 'jane@example.com', first_name: 'Jane', last_name: 'Smith', role: 'manager' }
    ];

    render(UserList, { users: mockUsers, searchable: true });

    const searchInput = screen.getByPlaceholderText(/search users/i);
    await fireEvent.input(searchInput, { target: { value: 'john' } });

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.queryByText('Jane Smith')).not.toBeInTheDocument();
    });
  });

  it('displays loading state', () => {
    render(UserList, { users: [], loading: true });

    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('displays error message', () => {
    const errorMessage = 'Failed to load users';
    render(UserList, { users: [], error: errorMessage });

    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  it('sorts users by column', async () => {
    const mockUsers = [
      { id: '1', email: 'john@example.com', first_name: 'John', last_name: 'Doe', role: 'employee' },
      { id: '2', email: 'jane@example.com', first_name: 'Jane', last_name: 'Smith', role: 'manager' }
    ];

    render(UserList, { users: mockUsers, sortable: true });

    const nameHeader = screen.getByText(/name/i);
    await fireEvent.click(nameHeader);

    const rows = screen.getAllByRole('row');
    expect(rows[1]).toHaveTextContent('Jane Smith');
    expect(rows[2]).toHaveTextContent('John Doe');
  });

  it('handles pagination', async () => {
    const mockUsers = Array.from({ length: 25 }, (_, i) => ({
      id: `${i + 1}`,
      email: `user${i + 1}@example.com`,
      first_name: `User${i + 1}`,
      last_name: 'Test',
      role: 'employee'
    }));

    render(UserList, { users: mockUsers, pageSize: 10 });

    expect(screen.getByText('User1 Test')).toBeInTheDocument();
    expect(screen.queryByText('User11 Test')).not.toBeInTheDocument();

    const nextButton = screen.getByRole('button', { name: /next/i });
    await fireEvent.click(nextButton);

    await waitFor(() => {
      expect(screen.getByText('User11 Test')).toBeInTheDocument();
      expect(screen.queryByText('User1 Test')).not.toBeInTheDocument();
    });
  });

  it('selects multiple users', async () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', first_name: 'John', last_name: 'Doe', role: 'employee' },
      { id: '2', email: 'user2@example.com', first_name: 'Jane', last_name: 'Smith', role: 'manager' }
    ];

    render(UserList, { users: mockUsers, selectable: true });

    const checkboxes = screen.getAllByRole('checkbox');
    await fireEvent.click(checkboxes[0]);
    await fireEvent.click(checkboxes[1]);

    expect(checkboxes[0]).toBeChecked();
    expect(checkboxes[1]).toBeChecked();
  });

  it('displays role badges with correct styling', () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', first_name: 'John', last_name: 'Doe', role: 'admin' },
      { id: '2', email: 'user2@example.com', first_name: 'Jane', last_name: 'Smith', role: 'employee' }
    ];

    render(UserList, { users: mockUsers });

    const adminBadge = screen.getByText('admin');
    const employeeBadge = screen.getByText('employee');

    expect(adminBadge).toHaveClass('badge-admin');
    expect(employeeBadge).toHaveClass('badge-employee');
  });
});
