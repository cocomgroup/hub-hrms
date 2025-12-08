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
  department?: string;
  position?: string;
}

interface AuthState {
  isAuthenticated: boolean;
  token: string | null;
  user: User | null;
  employee: Employee | null;
}

const initialState: AuthState = {
  isAuthenticated: false,
  token: null,
  user: null,
  employee: null,
};

function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  // Check for saved auth state on init
  if (typeof window !== 'undefined') {
    const savedToken = localStorage.getItem('token');
    const savedUser = localStorage.getItem('user');
    const savedEmployee = localStorage.getItem('employee');
    
    if (savedToken && savedUser) {
      set({
        isAuthenticated: true,
        token: savedToken,
        user: JSON.parse(savedUser),
        employee: savedEmployee ? JSON.parse(savedEmployee) : null,
      });
    }
  }

  return {
    subscribe,
    login: (token: string, user: User, employee?: Employee) => {
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      if (employee) {
        localStorage.setItem('employee', JSON.stringify(employee));
      }
      
      set({
        isAuthenticated: true,
        token,
        user,
        employee: employee || null,
      });
    },
    logout: () => {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('employee');
      set(initialState);
    },
    updateEmployee: (employee: Employee) => {
      localStorage.setItem('employee', JSON.stringify(employee));
      update(state => ({
        ...state,
        employee,
      }));
    },
  };
}

export const authStore = createAuthStore();
