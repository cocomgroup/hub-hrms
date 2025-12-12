import { writable, derived } from 'svelte/store';

// Auth user interface
export interface AuthUser {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  role: string;
  employee_id?: string;
}

// Auth state interface
export interface AuthState {
  token: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
}

// Initial state
const initialState: AuthState = {
  token: null,
  user: null,
  isAuthenticated: false,
  loading: false,
  error: null
};

// Create the writable store
function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  // Try to load auth from localStorage on initialization
  if (typeof window !== 'undefined') {
    const storedToken = localStorage.getItem('auth_token');
    const storedUser = localStorage.getItem('auth_user');
    
    if (storedToken && storedUser) {
      try {
        const user = JSON.parse(storedUser);
        set({
          token: storedToken,
          user,
          isAuthenticated: true,
          loading: false,
          error: null
        });
      } catch (err) {
        // Invalid stored data, clear it
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth_user');
      }
    }
  }

  return {
    subscribe,
    
    // Login method
    login: async (email: string, password: string) => {
      update(state => ({ ...state, loading: true, error: null }));
      
      try {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ email, password })
        });
        
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.message || 'Login failed');
        }
        
        const data = await response.json();
        
        // Store token and user
        if (typeof window !== 'undefined') {
          localStorage.setItem('auth_token', data.token);
          localStorage.setItem('auth_user', JSON.stringify(data.user));
        }
        
        update(state => ({
          ...state,
          token: data.token,
          user: data.user,
          isAuthenticated: true,
          loading: false,
          error: null
        }));
        
        return data;
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Login failed';
        update(state => ({
          ...state,
          loading: false,
          error: errorMessage
        }));
        throw err;
      }
    },
    
    // Logout method
    logout: () => {
      // Clear localStorage
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth_user');
      }
      
      // Reset state
      set(initialState);
    },
    
    // Register method
    register: async (userData: {
      email: string;
      password: string;
      first_name: string;
      last_name: string;
    }) => {
      update(state => ({ ...state, loading: true, error: null }));
      
      try {
        const response = await fetch('/api/auth/register', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(userData)
        });
        
        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.message || 'Registration failed');
        }
        
        const data = await response.json();
        
        // Store token and user
        if (typeof window !== 'undefined') {
          localStorage.setItem('auth_token', data.token);
          localStorage.setItem('auth_user', JSON.stringify(data.user));
        }
        
        update(state => ({
          ...state,
          token: data.token,
          user: data.user,
          isAuthenticated: true,
          loading: false,
          error: null
        }));
        
        return data;
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Registration failed';
        update(state => ({
          ...state,
          loading: false,
          error: errorMessage
        }));
        throw err;
      }
    },
    
    // Update user profile
    updateUser: (user: AuthUser) => {
      if (typeof window !== 'undefined') {
        localStorage.setItem('auth_user', JSON.stringify(user));
      }
      
      update(state => ({
        ...state,
        user
      }));
    },
    
    // Clear error
    clearError: () => {
      update(state => ({ ...state, error: null }));
    },
    
    // Set loading state
    setLoading: (loading: boolean) => {
      update(state => ({ ...state, loading }));
    },
    
    // Check if token is valid (you can enhance this with actual token validation)
    checkAuth: async () => {
      update(state => ({ ...state, loading: true }));
      
      try {
        const token = typeof window !== 'undefined' 
          ? localStorage.getItem('auth_token') 
          : null;
        
        if (!token) {
          update(state => ({
            ...initialState,
            loading: false
          }));
          return false;
        }
        
        // Validate token with backend
        const response = await fetch('/api/auth/me', {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });
        
        if (!response.ok) {
          throw new Error('Invalid token');
        }
        
        const user = await response.json();
        
        update(state => ({
          ...state,
          user,
          isAuthenticated: true,
          loading: false
        }));
        
        return true;
      } catch (err) {
        // Token is invalid, clear everything
        if (typeof window !== 'undefined') {
          localStorage.removeItem('auth_token');
          localStorage.removeItem('auth_user');
        }
        
        update(state => ({
          ...initialState,
          loading: false
        }));
        
        return false;
      }
    }
  };
}

// Create and export the store
export const authStore = createAuthStore();

// Derived stores for convenience
export const isAuthenticated = derived(
  authStore,
  $authStore => $authStore.isAuthenticated
);

export const currentUser = derived(
  authStore,
  $authStore => $authStore.user
);

export const authToken = derived(
  authStore,
  $authStore => $authStore.token
);

export const isLoading = derived(
  authStore,
  $authStore => $authStore.loading
);

export const authError = derived(
  authStore,
  $authStore => $authStore.error
);

// Helper to get auth header
export function getAuthHeader(): { Authorization: string } | {} {
  let token: string | null = null;
  
  authStore.subscribe(state => {
    token = state.token;
  })();
  
  return token ? { Authorization: `Bearer ${token}` } : {};
}