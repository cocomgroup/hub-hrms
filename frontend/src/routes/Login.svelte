<script lang="ts">
  import { authStore } from '../stores/auth';
  
  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleLogin() {
    error = '';
    loading = true;

    try {
      const response = await fetch(`${import.meta.env.VITE_API_URL}/api/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        throw new Error('Invalid credentials');
      }

      const data = await response.json();
      authStore.login(data.token, data.user, data.employee);
    } catch (err: any) {
      error = err.message || 'Login failed. Please try again.';
    } finally {
      loading = false;
    }
  }
</script>

<div class="login-container">
  <div class="login-card">
    <div class="login-header">
      <div class="logo">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
          <circle cx="9" cy="7" r="4"></circle>
          <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
          <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
        </svg>
      </div>
      <h1>Welcome to PeopleHub</h1>
      <p>Sign in to manage your HR operations</p>
    </div>

    <form onsubmit={(e) => { e.preventDefault(); handleLogin(); }}>
      {#if error}
        <div class="error-message">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          <span>{error}</span>
        </div>
      {/if}

      <div class="form-group">
        <label for="email">Email</label>
        <div class="input-wrapper">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
            <polyline points="22,6 12,13 2,6"></polyline>
          </svg>
          <input
            type="email"
            id="email"
            bind:value={email}
            placeholder="you@company.com"
            required
          />
        </div>
      </div>

      <div class="form-group">
        <label for="password">Password</label>
        <div class="input-wrapper">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
            <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
          </svg>
          <input
            type="password"
            id="password"
            bind:value={password}
            placeholder="••••••••"
            required
          />
        </div>
      </div>

      <button type="submit" class="submit-btn" disabled={loading}>
        {#if loading}
          <div class="spinner"></div>
          <span>Signing in...</span>
        {:else}
          <span>Sign In</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="5" y1="12" x2="19" y2="12"></line>
            <polyline points="12 5 19 12 12 19"></polyline>
          </svg>
        {/if}
      </button>
    </form>

    <div class="login-footer">
      <p>Demo credentials: <a href="/cdn-cgi/l/email-protection" class="__cf_email__" data-cfemail="58393c353136183b373528393621763b3735">[email&#160;protected]</a> / password</p>
    </div>
  </div>

  <div class="decorative-bg">
    <div class="circle circle-1"></div>
    <div class="circle circle-2"></div>
    <div class="circle circle-3"></div>
  </div>
</div>

<style>
  .login-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    position: relative;
    overflow: hidden;
  }

  .decorative-bg {
    position: absolute;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
  }

  .circle {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    animation: float 20s infinite;
  }

  .circle-1 {
    width: 400px;
    height: 400px;
    background: radial-gradient(circle, rgba(99, 102, 241, 0.3) 0%, transparent 70%);
    top: -200px;
    left: -100px;
    animation-delay: 0s;
  }

  .circle-2 {
    width: 500px;
    height: 500px;
    background: radial-gradient(circle, rgba(139, 92, 246, 0.25) 0%, transparent 70%);
    bottom: -250px;
    right: -150px;
    animation-delay: -7s;
  }

  .circle-3 {
    width: 300px;
    height: 300px;
    background: radial-gradient(circle, rgba(99, 102, 241, 0.2) 0%, transparent 70%);
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    animation-delay: -14s;
  }

  @keyframes float {
    0%, 100% {
      transform: translate(0, 0) scale(1);
    }
    33% {
      transform: translate(50px, -50px) scale(1.1);
    }
    66% {
      transform: translate(-50px, 50px) scale(0.9);
    }
  }

  .login-card {
    width: 100%;
    max-width: 440px;
    background: rgba(17, 24, 39, 0.9);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(99, 102, 241, 0.2);
    border-radius: 24px;
    padding: 3rem;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
    position: relative;
    z-index: 1;
    animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(30px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .login-header {
    text-align: center;
    margin-bottom: 2.5rem;
  }

  .logo {
    width: 64px;
    height: 64px;
    margin: 0 auto 1.5rem;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 8px 24px rgba(99, 102, 241, 0.3);
  }

  .logo svg {
    width: 32px;
    height: 32px;
    color: white;
  }

  h1 {
    font-size: 1.875rem;
    font-weight: 700;
    margin-bottom: 0.5rem;
    background: linear-gradient(135deg, #e4e7eb 0%, #94a3b8 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }

  .login-header p {
    color: #94a3b8;
    font-size: 0.9375rem;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem 1.25rem;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: 12px;
    color: #ef4444;
    font-size: 0.875rem;
    margin-bottom: 1.5rem;
  }

  .error-message svg {
    width: 20px;
    height: 20px;
    min-width: 20px;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  label {
    display: block;
    font-size: 0.875rem;
    font-weight: 600;
    color: #e4e7eb;
    margin-bottom: 0.5rem;
  }

  .input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .input-wrapper svg {
    position: absolute;
    left: 1rem;
    width: 20px;
    height: 20px;
    color: #64748b;
    pointer-events: none;
  }

  input {
    width: 100%;
    padding: 0.875rem 1rem 0.875rem 3rem;
    background: rgba(15, 23, 42, 0.6);
    border: 1px solid rgba(99, 102, 241, 0.2);
    border-radius: 12px;
    color: #e4e7eb;
    font-size: 0.9375rem;
    transition: all 0.2s;
  }

  input::placeholder {
    color: #64748b;
  }

  input:focus {
    outline: none;
    border-color: #6366f1;
    background: rgba(15, 23, 42, 0.8);
    box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
  }

  .submit-btn {
    width: 100%;
    padding: 1rem;
    background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
    border: none;
    border-radius: 12px;
    color: white;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    box-shadow: 0 4px 16px rgba(99, 102, 241, 0.4);
  }

  .submit-btn:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(99, 102, 241, 0.5);
  }

  .submit-btn:active:not(:disabled) {
    transform: translateY(0);
  }

  .submit-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .submit-btn svg {
    width: 20px;
    height: 20px;
  }

  .spinner {
    width: 18px;
    height: 18px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .login-footer {
    margin-top: 2rem;
    padding-top: 2rem;
    border-top: 1px solid rgba(99, 102, 241, 0.1);
    text-align: center;
  }

  .login-footer p {
    color: #64748b;
    font-size: 0.8125rem;
  }
</style>