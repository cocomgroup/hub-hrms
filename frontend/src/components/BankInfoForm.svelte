<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth';
  import { getApiBaseUrl } from '../lib/api';
  
  const API_BASE_URL = getApiBaseUrl();

  interface BankInfo {
    id?: string;
    employee_id?: string;
    account_holder_name: string;
    bank_name: string;
    account_type: string;
    account_number: string;
    routing_number: string;
    swift_code?: string;
    bank_address?: string;
    bank_city?: string;
    bank_state?: string;
    bank_zip?: string;
    bank_country?: string;
    is_primary: boolean;
    status: string;
    verified: boolean;
  }

  // Props
  let { 
    employeeId = null,
    onComplete = null,
    isOnboarding = true,
    showHeader = true
  }: { 
    employeeId?: string | null;
    onComplete?: ((bankInfo: BankInfo) => void) | null;
    isOnboarding?: boolean;
    showHeader?: boolean;
  } = $props();

  // State
  let loading = $state(false);
  let saving = $state(false);
  let error = $state('');
  let success = $state('');
  let existingBankInfo = $state<BankInfo | null>(null);
  let showAccountNumber = $state(false);
  let showRoutingNumber = $state(false);

  // Form data
  let bankInfo = $state<BankInfo>({
    account_holder_name: '',
    bank_name: '',
    account_type: 'checking',
    account_number: '',
    routing_number: '',
    swift_code: '',
    bank_address: '',
    bank_city: '',
    bank_state: '',
    bank_zip: '',
    bank_country: 'US',
    is_primary: true,
    status: 'pending',
    verified: false
  });

  onMount(() => {
    if (employeeId) {
      loadExistingBankInfo();
    }
    
    // Pre-fill account holder name if employee info available
    if ($authStore.employee) {
      bankInfo.account_holder_name = `${$authStore.employee.first_name} ${$authStore.employee.last_name}`;
    }
  });

  async function loadExistingBankInfo() {
    try {
      loading = true;
      const response = await fetch(`${API_BASE_URL}/bank-info/${employeeId}`, {
        headers: { 'Authorization': `Bearer ${$authStore.token}` }
      });
      
      if (response.ok) {
        existingBankInfo = await response.json();
        if (existingBankInfo) {
          bankInfo = { ...existingBankInfo };
        }
      }
    } catch (err: any) {
      console.error('Error loading bank info:', err);
    } finally {
      loading = false;
    }
  }

  async function handleSubmit() {
    error = '';
    success = '';

    // Validation
    if (!bankInfo.account_holder_name.trim()) {
      error = 'Account holder name is required';
      return;
    }
    if (!bankInfo.bank_name.trim()) {
      error = 'Bank name is required';
      return;
    }
    if (!bankInfo.account_number.trim()) {
      error = 'Account number is required';
      return;
    }
    if (!bankInfo.routing_number.trim()) {
      error = 'Routing number is required';
      return;
    }
    
    // Validate routing number (9 digits)
    if (!/^\d{9}$/.test(bankInfo.routing_number)) {
      error = 'Routing number must be exactly 9 digits';
      return;
    }
    
    // Validate account number (typically 8-17 digits)
    if (!/^\d{8,17}$/.test(bankInfo.account_number)) {
      error = 'Account number must be between 8 and 17 digits';
      return;
    }

    try {
      saving = true;
      
      const payload = {
        ...bankInfo,
        employee_id: employeeId || $authStore.employee?.id
      };

      const method = existingBankInfo ? 'PUT' : 'POST';
      const url = existingBankInfo 
        ? `${API_BASE_URL}/bank-info/${existingBankInfo.id}`
        : `${API_BASE_URL}/bank-info`;

      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${$authStore.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Failed to save bank information' }));
        throw new Error(errorData.error || 'Failed to save bank information');
      }

      const result = await response.json();
      success = existingBankInfo ? 'Bank information updated successfully!' : 'Bank information saved successfully!';
      
      // Call onComplete callback if provided
      if (onComplete) {
        onComplete(result);
      }

      // Clear success message after 3 seconds
      setTimeout(() => {
        success = '';
      }, 3000);

    } catch (err: any) {
      error = err.message;
    } finally {
      saving = false;
    }
  }

  function handleSkip() {
    if (onComplete) {
      onComplete(null);
    }
  }

  function formatAccountNumber(value: string): string {
    // Only allow digits
    return value.replace(/\D/g, '');
  }

  function formatRoutingNumber(value: string): string {
    // Only allow digits, max 9
    return value.replace(/\D/g, '').slice(0, 9);
  }

  // US states for dropdown
  const US_STATES = [
    'AL', 'AK', 'AZ', 'AR', 'CA', 'CO', 'CT', 'DE', 'FL', 'GA',
    'HI', 'ID', 'IL', 'IN', 'IA', 'KS', 'KY', 'LA', 'ME', 'MD',
    'MA', 'MI', 'MN', 'MS', 'MO', 'MT', 'NE', 'NV', 'NH', 'NJ',
    'NM', 'NY', 'NC', 'ND', 'OH', 'OK', 'OR', 'PA', 'RI', 'SC',
    'SD', 'TN', 'TX', 'UT', 'VT', 'VA', 'WA', 'WV', 'WI', 'WY'
  ];
</script>

<div class="bank-info-form">
  {#if showHeader}
    <div class="form-header">
      <h2>🏦 Bank Information</h2>
      <p class="subtitle">
        {#if isOnboarding}
          Enter your banking details for direct deposit payments. Your information is encrypted and secure.
        {:else}
          Update your banking information for direct deposit.
        {/if}
      </p>
    </div>
  {/if}

  {#if loading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading bank information...</p>
    </div>
  {:else}
    <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      {#if error}
        <div class="alert alert-error">
          ⚠️ {error}
        </div>
      {/if}

      {#if success}
        <div class="alert alert-success">
          ✓ {success}
        </div>
      {/if}

      <!-- Account Holder Information -->
      <div class="form-section">
        <h3>Account Holder Information</h3>
        
        <div class="form-group">
          <label for="account-holder-name">Account Holder Name *</label>
          <input
            id="account-holder-name"
            type="text"
            bind:value={bankInfo.account_holder_name}
            placeholder="Full name as it appears on the account"
            required
          />
          <span class="help-text">Must match the name on your bank account</span>
        </div>
      </div>

      <!-- Bank Details -->
      <div class="form-section">
        <h3>Bank Details</h3>
        
        <div class="form-grid">
          <div class="form-group">
            <label for="bank-name">Bank Name *</label>
            <input
              id="bank-name"
              type="text"
              bind:value={bankInfo.bank_name}
              placeholder="e.g., Chase, Bank of America"
              required
            />
          </div>

          <div class="form-group">
            <label for="account-type">Account Type *</label>
            <select id="account-type" bind:value={bankInfo.account_type} required>
              <option value="checking">Checking</option>
              <option value="savings">Savings</option>
            </select>
          </div>
        </div>

        <div class="form-grid">
          <div class="form-group">
            <label for="routing-number">Routing Number *</label>
            <div class="input-with-toggle">
              <input
                id="routing-number"
                type={showRoutingNumber ? 'text' : 'password'}
                bind:value={bankInfo.routing_number}
                oninput={(e) => bankInfo.routing_number = formatRoutingNumber(e.target.value)}
                placeholder="9 digits"
                maxlength="9"
                required
              />
              <button
                type="button"
                class="toggle-visibility"
                onclick={() => showRoutingNumber = !showRoutingNumber}
              >
                {showRoutingNumber ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
            <span class="help-text">9-digit ABA routing number</span>
          </div>

          <div class="form-group">
            <label for="account-number">Account Number *</label>
            <div class="input-with-toggle">
              <input
                id="account-number"
                type={showAccountNumber ? 'text' : 'password'}
                bind:value={bankInfo.account_number}
                oninput={(e) => bankInfo.account_number = formatAccountNumber(e.target.value)}
                placeholder="8-17 digits"
                maxlength="17"
                required
              />
              <button
                type="button"
                class="toggle-visibility"
                onclick={() => showAccountNumber = !showAccountNumber}
              >
                {showAccountNumber ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
            <span class="help-text">8-17 digit account number</span>
          </div>
        </div>

        <div class="form-group">
          <label for="swift-code">SWIFT/BIC Code (Optional)</label>
          <input
            id="swift-code"
            type="text"
            bind:value={bankInfo.swift_code}
            placeholder="For international transfers"
            maxlength="11"
          />
          <span class="help-text">Required only for international payments</span>
        </div>
      </div>

      <!-- Bank Address (Optional) -->
      <div class="form-section collapsible">
        <h3>Bank Address (Optional)</h3>
        
        <div class="form-group">
          <label for="bank-address">Street Address</label>
          <input
            id="bank-address"
            type="text"
            bind:value={bankInfo.bank_address}
            placeholder="Bank branch address"
          />
        </div>

        <div class="form-grid">
          <div class="form-group">
            <label for="bank-city">City</label>
            <input
              id="bank-city"
              type="text"
              bind:value={bankInfo.bank_city}
              placeholder="City"
            />
          </div>

          <div class="form-group">
            <label for="bank-state">State</label>
            <select id="bank-state" bind:value={bankInfo.bank_state}>
              <option value="">Select State</option>
              {#each US_STATES as state}
                <option value={state}>{state}</option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label for="bank-zip">ZIP Code</label>
            <input
              id="bank-zip"
              type="text"
              bind:value={bankInfo.bank_zip}
              placeholder="12345"
              maxlength="10"
            />
          </div>
        </div>

        <div class="form-group">
          <label for="bank-country">Country</label>
          <select id="bank-country" bind:value={bankInfo.bank_country}>
            <option value="US">United States</option>
            <option value="CA">Canada</option>
            <option value="GB">United Kingdom</option>
            <option value="AU">Australia</option>
            <option value="NZ">New Zealand</option>
            <option value="IE">Ireland</option>
          </select>
        </div>
      </div>

      <!-- Security Notice -->
      <div class="security-notice">
        <div class="security-icon">🔒</div>
        <div class="security-text">
          <strong>Your information is secure</strong>
          <p>Bank details are encrypted and stored securely. We never share your information with third parties.</p>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="form-actions">
        {#if isOnboarding}
          <button type="button" class="btn btn-secondary" onclick={handleSkip}>
            Skip for Now
          </button>
        {/if}
        
        <button type="submit" class="btn btn-primary" disabled={saving}>
          {#if saving}
            <span class="spinner-small"></span> Saving...
          {:else if existingBankInfo}
            Update Bank Information
          {:else}
            Save Bank Information
          {/if}
        </button>
      </div>

      <!-- Help Text -->
      <div class="help-section">
        <h4>Where to find this information</h4>
        <ul>
          <li><strong>Routing Number:</strong> 9-digit number found on the bottom left of your checks</li>
          <li><strong>Account Number:</strong> Found on the bottom center of your checks (8-17 digits)</li>
          <li><strong>Account Type:</strong> Check with your bank if unsure</li>
        </ul>
      </div>
    </form>
  {/if}
</div>

<style>
  .bank-info-form {
    max-width: 800px;
    margin: 0 auto;
    padding: 24px;
    background: white;
    border-radius: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .form-header {
    margin-bottom: 32px;
    text-align: center;
  }

  .form-header h2 {
    font-size: 28px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }

  .subtitle {
    font-size: 16px;
    color: #6b7280;
    margin: 0;
    line-height: 1.5;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px;
    gap: 16px;
  }

  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading p {
    color: #6b7280;
    margin: 0;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .alert {
    padding: 12px 16px;
    border-radius: 8px;
    font-size: 14px;
    margin-bottom: 16px;
  }

  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fca5a5;
  }

  .alert-success {
    background: #d1fae5;
    color: #065f46;
    border: 1px solid #6ee7b7;
  }

  .form-section {
    padding: 20px;
    background: #f9fafb;
    border-radius: 8px;
    border: 1px solid #e5e7eb;
  }

  .form-section h3 {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    margin: 0 0 16px 0;
  }

  .form-section.collapsible {
    background: #ffffff;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 16px;
  }

  .form-group:last-child {
    margin-bottom: 0;
  }

  label {
    font-size: 14px;
    font-weight: 500;
    color: #374151;
  }

  input, select {
    padding: 10px 12px;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 14px;
    transition: border-color 0.2s;
  }

  input:focus, select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  input[type="password"] {
    font-family: 'Courier New', monospace;
    letter-spacing: 2px;
  }

  .input-with-toggle {
    position: relative;
  }

  .input-with-toggle input {
    padding-right: 40px;
  }

  .toggle-visibility {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    cursor: pointer;
    font-size: 18px;
    padding: 4px 8px;
    opacity: 0.6;
    transition: opacity 0.2s;
  }

  .toggle-visibility:hover {
    opacity: 1;
  }

  .help-text {
    font-size: 12px;
    color: #6b7280;
    font-style: italic;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
  }

  .security-notice {
    display: flex;
    gap: 12px;
    padding: 16px;
    background: #eff6ff;
    border: 1px solid #bfdbfe;
    border-radius: 8px;
    margin-top: 8px;
  }

  .security-icon {
    font-size: 24px;
  }

  .security-text {
    flex: 1;
  }

  .security-text strong {
    display: block;
    color: #1e40af;
    margin-bottom: 4px;
  }

  .security-text p {
    font-size: 13px;
    color: #1e40af;
    margin: 0;
    line-height: 1.5;
  }

  .form-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    padding-top: 16px;
    border-top: 1px solid #e5e7eb;
  }

  .btn {
    padding: 10px 24px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #f9fafb;
  }

  .spinner-small {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  .help-section {
    background: #f9fafb;
    padding: 16px;
    border-radius: 8px;
    margin-top: 8px;
  }

  .help-section h4 {
    font-size: 14px;
    font-weight: 600;
    color: #374151;
    margin: 0 0 8px 0;
  }

  .help-section ul {
    margin: 0;
    padding-left: 20px;
  }

  .help-section li {
    font-size: 13px;
    color: #6b7280;
    margin-bottom: 6px;
    line-height: 1.5;
  }

  .help-section li:last-child {
    margin-bottom: 0;
  }

  @media (max-width: 768px) {
    .bank-info-form {
      padding: 16px;
    }

    .form-header h2 {
      font-size: 24px;
    }

    .form-grid {
      grid-template-columns: 1fr;
    }

    .form-actions {
      flex-direction: column-reverse;
    }

    .btn {
      width: 100%;
      justify-content: center;
    }
  }
</style>