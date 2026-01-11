<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { slide } from 'svelte/transition';
  
  const dispatch = createEventDispatcher();

  export let employeeId: string;
  export let employeeName: string;

  interface Package {
    id: string;
    name: string;
    description: string;
    check_types: string[];
    turnaround_days: number;
    cost: number;
  }

  interface CandidateInfo {
    first_name: string;
    middle_name?: string;
    last_name: string;
    email: string;
    phone: string;
    date_of_birth: string;
    ssn?: string;
    driver_license?: string;
    address: {
      street1: string;
      street2?: string;
      city: string;
      state: string;
      postal_code: string;
      country: string;
    };
  }

  let packages: Package[] = [];
  let loading = false;
  let error = '';
  let currentStep = 1;
  
  // Form data
  let selectedPackage = '';
  let candidateInfo: CandidateInfo = {
    first_name: '',
    last_name: '',
    email: '',
    phone: '',
    date_of_birth: '',
    address: {
      street1: '',
      city: '',
      state: '',
      postal_code: '',
      country: 'US'
    }
  };
  
  let fcraConsent = false;
  let fcraDisclosureRead = false;
  let signatureData = '';

  // Load available packages
  async function loadPackages() {
    try {
      loading = true;
      const response = await fetch('/api/v1/background-checks/packages');
      if (!response.ok) throw new Error('Failed to load packages');
      packages = await response.json();
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  // Initialize
  loadPackages();

  function nextStep() {
    if (currentStep === 1 && !selectedPackage) {
      error = 'Please select a background check package';
      return;
    }
    
    if (currentStep === 2 && !validateCandidateInfo()) {
      return;
    }
    
    if (currentStep === 3 && (!fcraConsent || !fcraDisclosureRead)) {
      error = 'You must read the disclosure and provide consent';
      return;
    }
    
    error = '';
    currentStep++;
  }

  function previousStep() {
    error = '';
    currentStep--;
  }

  function validateCandidateInfo(): boolean {
    if (!candidateInfo.first_name || !candidateInfo.last_name) {
      error = 'First name and last name are required';
      return false;
    }
    
    if (!candidateInfo.email || !candidateInfo.email.includes('@')) {
      error = 'Valid email is required';
      return false;
    }
    
    if (!candidateInfo.date_of_birth) {
      error = 'Date of birth is required';
      return false;
    }
    
    if (!candidateInfo.address.street1 || !candidateInfo.address.city || 
        !candidateInfo.address.state || !candidateInfo.address.postal_code) {
      error = 'Complete address is required';
      return false;
    }
    
    return true;
  }

  async function submitBackgroundCheck() {
    if (!fcraConsent || !fcraDisclosureRead) {
      error = 'FCRA consent is required';
      return;
    }

    try {
      loading = true;
      error = '';

      const response = await fetch('/api/v1/background-checks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          employee_id: employeeId,
          package_id: selectedPackage,
          candidate: candidateInfo,
          consent: {
            fcra_disclosure_provided: true,
            fcra_consent_obtained: fcraConsent,
            fcra_consent_date: new Date().toISOString(),
            signature_data: signatureData
          }
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to initiate background check');
      }

      const result = await response.json();
      dispatch('success', result);
      
    } catch (err) {
      error = err.message;
    } finally {
      loading = false;
    }
  }

  function getSelectedPackage(): Package | undefined {
    return packages.find(p => p.id === selectedPackage);
  }

  const states = [
    'AL', 'AK', 'AZ', 'AR', 'CA', 'CO', 'CT', 'DE', 'FL', 'GA',
    'HI', 'ID', 'IL', 'IN', 'IA', 'KS', 'KY', 'LA', 'ME', 'MD',
    'MA', 'MI', 'MN', 'MS', 'MO', 'MT', 'NE', 'NV', 'NH', 'NJ',
    'NM', 'NY', 'NC', 'ND', 'OH', 'OK', 'OR', 'PA', 'RI', 'SC',
    'SD', 'TN', 'TX', 'UT', 'VT', 'VA', 'WA', 'WV', 'WI', 'WY'
  ];
</script>

<div class="background-check-wizard">
  <div class="wizard-header">
    <h2>Initiate Background Check</h2>
    <p class="employee-name">for {employeeName}</p>
  </div>

  <!-- Progress Steps -->
  <div class="progress-steps">
    <div class="step" class:active={currentStep === 1} class:completed={currentStep > 1}>
      <div class="step-number">1</div>
      <div class="step-label">Select Package</div>
    </div>
    <div class="step-divider"></div>
    <div class="step" class:active={currentStep === 2} class:completed={currentStep > 2}>
      <div class="step-number">2</div>
      <div class="step-label">Candidate Info</div>
    </div>
    <div class="step-divider"></div>
    <div class="step" class:active={currentStep === 3} class:completed={currentStep > 3}>
      <div class="step-number">3</div>
      <div class="step-label">FCRA Consent</div>
    </div>
    <div class="step-divider"></div>
    <div class="step" class:active={currentStep === 4}>
      <div class="step-number">4</div>
      <div class="step-label">Review & Submit</div>
    </div>
  </div>

  {#if error}
    <div class="error-message" transition:slide>
      {error}
    </div>
  {/if}

  <!-- Step 1: Select Package -->
  {#if currentStep === 1}
    <div class="step-content" transition:slide>
      <h3>Select Background Check Package</h3>
      
      {#if loading}
        <div class="loading">Loading packages...</div>
      {:else if packages.length === 0}
        <div class="empty-state">No packages available</div>
      {:else}
        <div class="package-grid">
          {#each packages as pkg}
            <label class="package-card" class:selected={selectedPackage === pkg.id}>
              <input 
                type="radio" 
                name="package" 
                value={pkg.id} 
                bind:group={selectedPackage}
              />
              <div class="package-content">
                <h4>{pkg.name}</h4>
                <p class="package-description">{pkg.description}</p>
                <div class="package-details">
                  <div class="detail">
                    <span class="label">Includes:</span>
                    <ul class="check-types">
                      {#each pkg.check_types as checkType}
                        <li>{checkType.replace(/_/g, ' ')}</li>
                      {/each}
                    </ul>
                  </div>
                  <div class="package-meta">
                    <span class="turnaround">~{pkg.turnaround_days} days</span>
                    <span class="cost">${pkg.cost.toFixed(2)}</span>
                  </div>
                </div>
              </div>
            </label>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <!-- Step 2: Candidate Information -->
  {#if currentStep === 2}
    <div class="step-content" transition:slide>
      <h3>Candidate Information</h3>
      
      <div class="form-grid">
        <div class="form-group">
          <label for="first_name">First Name *</label>
          <input 
            type="text" 
            id="first_name" 
            bind:value={candidateInfo.first_name}
            required
          />
        </div>

        <div class="form-group">
          <label for="middle_name">Middle Name</label>
          <input 
            type="text" 
            id="middle_name" 
            bind:value={candidateInfo.middle_name}
          />
        </div>

        <div class="form-group">
          <label for="last_name">Last Name *</label>
          <input 
            type="text" 
            id="last_name" 
            bind:value={candidateInfo.last_name}
            required
          />
        </div>

        <div class="form-group">
          <label for="email">Email *</label>
          <input 
            type="email" 
            id="email" 
            bind:value={candidateInfo.email}
            required
          />
        </div>

        <div class="form-group">
          <label for="phone">Phone *</label>
          <input 
            type="tel" 
            id="phone" 
            bind:value={candidateInfo.phone}
            placeholder="(555) 123-4567"
            required
          />
        </div>

        <div class="form-group">
          <label for="dob">Date of Birth *</label>
          <input 
            type="date" 
            id="dob" 
            bind:value={candidateInfo.date_of_birth}
            required
          />
        </div>

        <div class="form-group">
          <label for="ssn">SSN (Optional)</label>
          <input 
            type="password" 
            id="ssn" 
            bind:value={candidateInfo.ssn}
            placeholder="XXX-XX-XXXX"
            autocomplete="off"
          />
          <small>Encrypted and stored securely</small>
        </div>

        <div class="form-group">
          <label for="dl">Driver License (Optional)</label>
          <input 
            type="text" 
            id="dl" 
            bind:value={candidateInfo.driver_license}
          />
        </div>
      </div>

      <h4>Address</h4>
      <div class="form-grid">
        <div class="form-group full-width">
          <label for="street1">Street Address *</label>
          <input 
            type="text" 
            id="street1" 
            bind:value={candidateInfo.address.street1}
            required
          />
        </div>

        <div class="form-group full-width">
          <label for="street2">Apt/Suite (Optional)</label>
          <input 
            type="text" 
            id="street2" 
            bind:value={candidateInfo.address.street2}
          />
        </div>

        <div class="form-group">
          <label for="city">City *</label>
          <input 
            type="text" 
            id="city" 
            bind:value={candidateInfo.address.city}
            required
          />
        </div>

        <div class="form-group">
          <label for="state">State *</label>
          <select id="state" bind:value={candidateInfo.address.state} required>
            <option value="">Select State</option>
            {#each states as state}
              <option value={state}>{state}</option>
            {/each}
          </select>
        </div>

        <div class="form-group">
          <label for="postal">ZIP Code *</label>
          <input 
            type="text" 
            id="postal" 
            bind:value={candidateInfo.address.postal_code}
            placeholder="12345"
            required
          />
        </div>
      </div>
    </div>
  {/if}

  <!-- Step 3: FCRA Consent -->
  {#if currentStep === 3}
    <div class="step-content" transition:slide>
      <h3>FCRA Disclosure and Authorization</h3>
      
      <div class="fcra-disclosure">
        <h4>Important Information About Background Reports</h4>
        <div class="disclosure-text">
          <p>In connection with your employment application, we may obtain a "consumer report" 
          as defined in the Fair Credit Reporting Act (FCRA) from a consumer reporting agency. 
          The consumer report may include information about your character, general reputation, 
          personal characteristics, and mode of living.</p>

          <p>The types of information that may be obtained include, but are not limited to:</p>
          <ul>
            <li>Social Security number verification</li>
            <li>Criminal records check</li>
            <li>Employment history verification</li>
            <li>Education verification</li>
            <li>Motor vehicle records</li>
            <li>Credit history (if applicable)</li>
          </ul>

          <p>You have the right under the FCRA to:</p>
          <ul>
            <li>Request more information about the nature and scope of the report</li>
            <li>Know whether a consumer report was requested</li>
            <li>Request information from the consumer reporting agency regarding the report</li>
            <li>Dispute the accuracy or completeness of information in the report</li>
          </ul>

          <p>The consumer reporting agency is Checkr, Inc., located at 1 Montgomery St., 
          Suite 2400, San Francisco, CA 94104. Phone: 1-844-824-3257. 
          Website: https://checkr.com</p>
        </div>

        <label class="consent-checkbox">
          <input type="checkbox" bind:checked={fcraDisclosureRead} />
          <span>I have read and understand the FCRA disclosure</span>
        </label>

        <label class="consent-checkbox">
          <input type="checkbox" bind:checked={fcraConsent} />
          <span>I authorize the company to obtain consumer reports and investigative 
          consumer reports for employment purposes</span>
        </label>

        <div class="signature-section">
          <label for="signature">Electronic Signature *</label>
          <input 
            type="text" 
            id="signature" 
            bind:value={signatureData}
            placeholder="Type your full name"
            required
          />
          <small>By typing your name, you are providing an electronic signature that is 
          legally binding</small>
        </div>

        <div class="consent-date">
          <strong>Date:</strong> {new Date().toLocaleDateString()}
        </div>
      </div>
    </div>
  {/if}

  <!-- Step 4: Review and Submit -->
  {#if currentStep === 4}
    <div class="step-content" transition:slide>
      <h3>Review and Submit</h3>
      
      <div class="review-section">
        <h4>Selected Package</h4>
        {#if getSelectedPackage()}
          <div class="review-item">
            <span class="label">{getSelectedPackage().name}</span>
            <span class="value">${getSelectedPackage().cost.toFixed(2)}</span>
          </div>
        {/if}

        <h4>Candidate Information</h4>
        <div class="review-item">
          <span class="label">Name:</span>
          <span class="value">
            {candidateInfo.first_name} 
            {candidateInfo.middle_name || ''} 
            {candidateInfo.last_name}
          </span>
        </div>
        <div class="review-item">
          <span class="label">Email:</span>
          <span class="value">{candidateInfo.email}</span>
        </div>
        <div class="review-item">
          <span class="label">Phone:</span>
          <span class="value">{candidateInfo.phone}</span>
        </div>
        <div class="review-item">
          <span class="label">Address:</span>
          <span class="value">
            {candidateInfo.address.street1}
            {candidateInfo.address.street2 ? `, ${candidateInfo.address.street2}` : ''}
            <br />
            {candidateInfo.address.city}, {candidateInfo.address.state} 
            {candidateInfo.address.postal_code}
          </span>
        </div>

        <h4>Consent Status</h4>
        <div class="review-item">
          <span class="label">FCRA Consent:</span>
          <span class="value consent-status">✓ Obtained</span>
        </div>
        <div class="review-item">
          <span class="label">Signature:</span>
          <span class="value">{signatureData}</span>
        </div>
      </div>
    </div>
  {/if}

  <!-- Navigation Buttons -->
  <div class="wizard-footer">
    {#if currentStep > 1}
      <button class="btn btn-secondary" on:click={previousStep} disabled={loading}>
        Previous
      </button>
    {/if}

    {#if currentStep < 4}
      <button class="btn btn-primary" on:click={nextStep} disabled={loading}>
        Next
      </button>
    {:else}
      <button class="btn btn-primary" on:click={submitBackgroundCheck} disabled={loading}>
        {loading ? 'Submitting...' : 'Submit Background Check'}
      </button>
    {/if}
  </div>
</div>

<style>
  .background-check-wizard {
    max-width: 900px;
    margin: 0 auto;
    padding: 2rem;
  }

  .wizard-header {
    text-align: center;
    margin-bottom: 2rem;
  }

  .wizard-header h2 {
    margin: 0;
    font-size: 1.75rem;
    color: #1a1a1a;
  }

  .employee-name {
    color: #666;
    margin-top: 0.5rem;
  }

  .progress-steps {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 3rem;
  }

  .step {
    display: flex;
    flex-direction: column;
    align-items: center;
    flex: 1;
  }

  .step-number {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: #e0e0e0;
    color: #999;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    margin-bottom: 0.5rem;
    transition: all 0.3s;
  }

  .step.active .step-number {
    background: #3b82f6;
    color: white;
  }

  .step.completed .step-number {
    background: #10b981;
    color: white;
  }

  .step-label {
    font-size: 0.875rem;
    color: #666;
    text-align: center;
  }

  .step-divider {
    flex: 1;
    height: 2px;
    background: #e0e0e0;
    margin: 0 1rem;
    margin-bottom: 1.5rem;
  }

  .error-message {
    background: #fee;
    color: #c33;
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
  }

  .step-content {
    background: white;
    padding: 2rem;
    border-radius: 0.5rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    margin-bottom: 2rem;
  }

  .step-content h3 {
    margin-top: 0;
    color: #1a1a1a;
  }

  .package-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-top: 1.5rem;
  }

  .package-card {
    border: 2px solid #e0e0e0;
    border-radius: 0.5rem;
    padding: 1.5rem;
    cursor: pointer;
    transition: all 0.3s;
  }

  .package-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 8px rgba(59, 130, 246, 0.2);
  }

  .package-card.selected {
    border-color: #3b82f6;
    background: #eff6ff;
  }

  .package-card input[type="radio"] {
    display: none;
  }

  .package-content h4 {
    margin-top: 0;
    color: #1a1a1a;
  }

  .package-description {
    color: #666;
    font-size: 0.875rem;
    margin: 0.5rem 0 1rem;
  }

  .check-types {
    list-style: none;
    padding: 0;
    margin: 0.5rem 0;
  }

  .check-types li {
    padding: 0.25rem 0;
    font-size: 0.875rem;
    color: #666;
    text-transform: capitalize;
  }

  .check-types li:before {
    content: "✓ ";
    color: #10b981;
    font-weight: bold;
  }

  .package-meta {
    display: flex;
    justify-content: space-between;
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #e0e0e0;
  }

  .turnaround, .cost {
    font-size: 0.875rem;
    font-weight: 600;
  }

  .cost {
    color: #3b82f6;
    font-size: 1.125rem;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
    margin-top: 1.5rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
  }

  .form-group.full-width {
    grid-column: 1 / -1;
  }

  .form-group label {
    font-weight: 500;
    margin-bottom: 0.5rem;
    color: #1a1a1a;
  }

  .form-group input,
  .form-group select {
    padding: 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    font-size: 1rem;
  }

  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .form-group small {
    font-size: 0.75rem;
    color: #666;
    margin-top: 0.25rem;
  }

  .fcra-disclosure {
    background: #f9fafb;
    padding: 2rem;
    border-radius: 0.5rem;
    border: 1px solid #e0e0e0;
  }

  .disclosure-text {
    max-height: 400px;
    overflow-y: auto;
    padding: 1rem;
    background: white;
    border-radius: 0.375rem;
    margin-bottom: 1.5rem;
  }

  .disclosure-text p {
    margin-bottom: 1rem;
    line-height: 1.6;
    color: #374151;
  }

  .disclosure-text ul {
    margin-left: 1.5rem;
    margin-bottom: 1rem;
  }

  .disclosure-text li {
    margin-bottom: 0.5rem;
    color: #374151;
  }

  .consent-checkbox {
    display: flex;
    align-items: center;
    margin: 1rem 0;
    cursor: pointer;
  }

  .consent-checkbox input[type="checkbox"] {
    width: 20px;
    height: 20px;
    margin-right: 0.75rem;
    cursor: pointer;
  }

  .signature-section {
    margin-top: 2rem;
  }

  .signature-section label {
    display: block;
    font-weight: 600;
    margin-bottom: 0.5rem;
  }

  .signature-section input {
    width: 100%;
    padding: 0.75rem;
    border: 2px solid #3b82f6;
    border-radius: 0.375rem;
    font-family: cursive;
    font-size: 1.125rem;
  }

  .consent-date {
    margin-top: 1.5rem;
    padding: 1rem;
    background: white;
    border-radius: 0.375rem;
    text-align: center;
  }

  .review-section h4 {
    margin-top: 1.5rem;
    margin-bottom: 1rem;
    color: #1a1a1a;
  }

  .review-section h4:first-child {
    margin-top: 0;
  }

  .review-item {
    display: flex;
    justify-content: space-between;
    padding: 0.75rem 0;
    border-bottom: 1px solid #e0e0e0;
  }

  .review-item .label {
    font-weight: 500;
    color: #666;
  }

  .review-item .value {
    text-align: right;
    color: #1a1a1a;
  }

  .consent-status {
    color: #10b981;
    font-weight: 600;
  }

  .wizard-footer {
    display: flex;
    justify-content: space-between;
    margin-top: 2rem;
  }

  .btn {
    padding: 0.75rem 2rem;
    border: none;
    border-radius: 0.375rem;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn:disabled {
    opacity: 0.5;
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
    background: #6b7280;
    color: white;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #4b5563;
  }

  .loading,
  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #666;
  }
</style>
