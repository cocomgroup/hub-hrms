// Svelte 5 Auth Store using writable for compatibility
import { writable } from 'svelte/store';

interface User {
  id: string;
  email: string;
  role: string;
}

interface Employee {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  position: string;
  department: string;
  status?: string;
  manager_id?: string;
}

interface AuthState {
  token: string | null;
  user: User | null;
  employee: Employee | null;
  isAuthenticated: boolean;
}

const initialState: AuthState = {
  token: null,
  user: null,
  employee: null,
  isAuthenticated: false
};

// Create writable store for compatibility
export const authStore = writable<AuthState>(initialState);

// Helper functions
export function login(token: string, user: User, employee: Employee | null) {
  localStorage.setItem('token', token);
  localStorage.setItem('user', JSON.stringify(user));
  if (employee) {
    localStorage.setItem('employee', JSON.stringify(employee));
  }
  
  authStore.set({
    token,
    user,
    employee,
    isAuthenticated: true
  });
}

export function logout() {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  localStorage.removeItem('employee');
  
  authStore.set(initialState);
}

export function updateEmployee(employee: Employee) {
  localStorage.setItem('employee', JSON.stringify(employee));
  
  authStore.update(state => ({
    ...state,
    employee
  }));
}

export function updateUser(user: User) {
  localStorage.setItem('user', JSON.stringify(user));
  
  authStore.update(state => ({
    ...state,
    user
  }));
}