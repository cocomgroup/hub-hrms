// src/lib/api.ts
// Centralized API configuration with environment support

// Get API base URL from environment variable or use default
// In development: defaults to '/api' which is proxied by Vite to localhost:8080
// In production: uses VITE_API_BASE_URL environment variable (set during build)
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

// Log the API URL being used (helpful for debugging)
if (import.meta.env.DEV) {
  console.log('🔗 API Base URL:', API_BASE_URL);
}

/**
 * Makes an authenticated API request
 * @param endpoint - API endpoint (e.g., '/users', '/employees')
 * @param options - Fetch options (method, body, headers, etc.)
 * @returns Response object
 */
export async function apiRequest(
  endpoint: string,
  options: RequestInit = {}
): Promise<Response> {
  const token = localStorage.getItem('token');
  
  // Use Record type for proper header typing
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  // Merge existing headers if provided
  if (options.headers) {
    const existingHeaders = new Headers(options.headers);
    existingHeaders.forEach((value, key) => {
      headers[key] = value;
    });
  }

  // Add authorization token if available
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  // Ensure endpoint starts with /
  const normalizedEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
  const url = `${API_BASE_URL}${normalizedEndpoint}`;

  try {
    const response = await fetch(url, {
      ...options,
      headers,
    });

    return response;
  } catch (error) {
    console.error('API Request failed:', error);
    throw error;
  }
}

/**
 * Helper methods for common HTTP operations
 */
export const api = {
  /**
   * GET request
   */
  get: (endpoint: string) => 
    apiRequest(endpoint, { method: 'GET' }),
  
  /**
   * POST request with JSON body
   */
  post: (endpoint: string, data: any) => 
    apiRequest(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  
  /**
   * PUT request with JSON body
   */
  put: (endpoint: string, data: any) =>
    apiRequest(endpoint, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  
  /**
   * PATCH request with JSON body
   */
  patch: (endpoint: string, data: any) =>
    apiRequest(endpoint, {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),
  
  /**
   * DELETE request
   */
  delete: (endpoint: string) =>
    apiRequest(endpoint, { method: 'DELETE' }),
};

/**
 * Get the configured API base URL
 * Useful for direct fetch calls or debugging
 */
export function getApiBaseUrl(): string {
  return API_BASE_URL;
}

/**
 * Build a full API URL from an endpoint
 * @param endpoint - API endpoint
 * @returns Full URL string
 */
export function buildApiUrl(endpoint: string): string {
  const normalizedEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
  return `${API_BASE_URL}${normalizedEndpoint}`;
}