<script lang="ts">
  import { onMount } from 'svelte';
  
  interface Applicant {
    id: string;
    name: string;
    email: string;
    phone: string;
    position: string;
    source: string;
    applied_date: string;
    resume_url: string;
    ai_score: number;
    ai_analysis: {
      skills_match: number;
      experience_score: number;
      education_score: number;
      culture_fit: number;
      strengths: string[];
      concerns: string[];
      summary: string;
    };
    status: 'new' | 'reviewing' | 'interview' | 'offer' | 'rejected';
    notes: string;
  }
  
  let applicants: Applicant[] = [];
  let loading = true;
  let selectedApplicant: Applicant | null = null;
  let sortBy: 'score' | 'date' | 'name' = 'score';
  let filterStatus = 'all';
  let analyzing = false;
  
  // Upload Resume state
  let showUploadModal = false;
  let uploadForm = {
    name: '',
    email: '',
    phone: '',
    position: '',
    source: 'manual'
  };
  let selectedFile: File | null = null;
  let uploading = false;
  let uploadError = '';
  let uploadSuccess = '';
  
  $: sortedApplicants = applicants
    .filter(a => filterStatus === 'all' || a.status === filterStatus)
    .sort((a, b) => {
      if (sortBy === 'score') return b.ai_score - a.ai_score;
      if (sortBy === 'date') return new Date(b.applied_date).getTime() - new Date(a.applied_date).getTime();
      return a.name.localeCompare(b.name);
    });
  
  onMount(async () => {
    await loadApplicants();
  });
  
  async function loadApplicants() {
    try {
      loading = true;
      const token = localStorage.getItem('token');
      const response = await fetch('/api/recruiting/applicants', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        applicants = await response.json();
      }
    } catch (err) {
      console.error('Failed to load applicants:', err);
    } finally {
      loading = false;
    }
  }
  
  async function analyzeApplicant(applicantId: string) {
    try {
      analyzing = true;
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/recruiting/applicants/${applicantId}/analyze`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        await loadApplicants();
        const updated = applicants.find(a => a.id === applicantId);
        if (updated) selectedApplicant = updated;
      }
    } catch (err) {
      console.error('Failed to analyze applicant:', err);
    } finally {
      analyzing = false;
    }
  }
  
  function handleFileSelect(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files[0]) {
      selectedFile = input.files[0];
      
      // Validate file type
      const allowedTypes = ['application/pdf', 'application/msword', 
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document'];
      if (!allowedTypes.includes(selectedFile.type)) {
        uploadError = 'Please upload a PDF or Word document';
        selectedFile = null;
        return;
      }
      
      // Validate file size (5MB max)
      if (selectedFile.size > 5 * 1024 * 1024) {
        uploadError = 'File size must be less than 5MB';
        selectedFile = null;
        return;
      }
      
      uploadError = '';
    }
  }
  
  async function uploadResume() {
    if (!selectedFile) {
      uploadError = 'Please select a resume file';
      return;
    }
    
    if (!uploadForm.name || !uploadForm.email || !uploadForm.position) {
      uploadError = 'Please fill in all required fields';
      return;
    }
    
    try {
      uploading = true;
      uploadError = '';
      
      const formData = new FormData();
      formData.append('resume', selectedFile);
      formData.append('name', uploadForm.name);
      formData.append('email', uploadForm.email);
      formData.append('phone', uploadForm.phone);
      formData.append('position', uploadForm.position);
      formData.append('source', uploadForm.source);
      
      const token = localStorage.getItem('token');
      const response = await fetch('/api/recruiting/applicants/upload', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        },
        body: formData
      });
      
      if (response.ok) {
        uploadSuccess = 'Resume uploaded successfully!';
        // Reset form
        uploadForm = {
          name: '',
          email: '',
          phone: '',
          position: '',
          source: 'manual'
        };
        selectedFile = null;
        
        // Reload applicants
        await loadApplicants();
        
        // Close modal after 2 seconds
        setTimeout(() => {
          showUploadModal = false;
          uploadSuccess = '';
        }, 2000);
      } else {
        const error = await response.text();
        uploadError = error || 'Failed to upload resume';
      }
    } catch (err) {
      uploadError = 'Failed to upload resume: ' + (err as Error).message;
    } finally {
      uploading = false;
    }
  }
  
  function closeUploadModal() {
    showUploadModal = false;
    uploadError = '';
    uploadSuccess = '';
    uploadForm = {
      name: '',
      email: '',
      phone: '',
      position: '',
      source: 'manual'
    };
    selectedFile = null;
  }
  
  async function updateApplicantStatus(applicantId: string, status: string) {
    try {
      const token = localStorage.getItem('token');
      await fetch(`/api/recruiting/applicants/${applicantId}`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ status })
      });
      await loadApplicants();
      if (selectedApplicant?.id === applicantId) {
        selectedApplicant = null;
      }
    } catch (err) {
      console.error('Failed to update status:', err);
    }
  }
  
  function getRankIcon(index: number) {
    if (index === 0) return '🥇';
    if (index === 1) return '🥈';
    if (index === 2) return '🥉';
    return `#${index + 1}`;
  }
  
  function getScoreColor(score: number) {
    if (score >= 80) return '#10b981';
    if (score >= 60) return '#3b82f6';
    if (score >= 40) return '#f59e0b';
    return '#ef4444';
  }
</script>

<div class="leaderboard-container">
  <div class="leaderboard-header">
    <div>
      <h2>🏆 Applicant Leaderboard</h2>
      <p class="subtitle">AI-powered ranking and evaluation</p>
    </div>
    
    <div class="controls">
      <button 
        class="btn-upload"
        on:click={() => showUploadModal = true}
      >
        📎 Upload Resume
      </button>
      
      <select bind:value={filterStatus} class="control-select">
        <option value="all">All Status</option>
        <option value="new">New</option>
        <option value="reviewing">Reviewing</option>
        <option value="interview">Interview</option>
        <option value="offer">Offer</option>
        <option value="rejected">Rejected</option>
      </select>
      
      <select bind:value={sortBy} class="control-select">
        <option value="score">Sort by Score</option>
        <option value="date">Sort by Date</option>
        <option value="name">Sort by Name</option>
      </select>
    </div>
  </div>
  
  <div class="leaderboard-grid">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        Loading applicants...
      </div>
    {:else if sortedApplicants.length === 0}
      <div class="empty-state">
        <span class="empty-icon">👥</span>
        <p>No applicants yet</p>
      </div>
    {:else}
      {#each sortedApplicants as applicant, index}
        <div 
          class="applicant-card"
          class:selected={selectedApplicant?.id === applicant.id}
          class:top-three={index < 3}
          on:click={() => selectedApplicant = applicant}
        >
          <div class="card-header">
            <div class="rank-badge" class:top-three={index < 3}>
              {getRankIcon(index)}
            </div>
            <div class="applicant-info">
              <h3 class="applicant-name">{applicant.name}</h3>
              <p class="applicant-position">{applicant.position}</p>
            </div>
            <div class="score-badge" style="background: {getScoreColor(applicant.ai_score)}20; color: {getScoreColor(applicant.ai_score)}">
              {applicant.ai_score}/100
            </div>
          </div>
          
          <div class="card-body">
            <div class="meta-row">
              <span class="meta-item">📍 {applicant.source}</span>
              <span class="meta-item">📅 {new Date(applicant.applied_date).toLocaleDateString()}</span>
            </div>
            
            <div class="score-breakdown">
              <div class="breakdown-item">
                <div class="breakdown-header">
                  <span>Skills</span>
                  <span>{applicant.ai_analysis.skills_match}%</span>
                </div>
                <div class="breakdown-bar">
                  <div class="breakdown-fill" style="width: {applicant.ai_analysis.skills_match}%; background: {getScoreColor(applicant.ai_analysis.skills_match)}"></div>
                </div>
              </div>
              
              <div class="breakdown-item">
                <div class="breakdown-header">
                  <span>Experience</span>
                  <span>{applicant.ai_analysis.experience_score}%</span>
                </div>
                <div class="breakdown-bar">
                  <div class="breakdown-fill" style="width: {applicant.ai_analysis.experience_score}%; background: {getScoreColor(applicant.ai_analysis.experience_score)}"></div>
                </div>
              </div>
              
              <div class="breakdown-item">
                <div class="breakdown-header">
                  <span>Culture Fit</span>
                  <span>{applicant.ai_analysis.culture_fit}%</span>
                </div>
                <div class="breakdown-bar">
                  <div class="breakdown-fill" style="width: {applicant.ai_analysis.culture_fit}%; background: {getScoreColor(applicant.ai_analysis.culture_fit)}"></div>
                </div>
              </div>
            </div>
          </div>
          
          <div class="card-footer">
            <span class="status-badge status-{applicant.status}">
              {applicant.status}
            </span>
            <button class="view-btn" on:click|stopPropagation={() => selectedApplicant = applicant}>
              View Details →
            </button>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

{#if selectedApplicant}
  <div class="modal-overlay" on:click={() => selectedApplicant = null}>
    <div class="modal-large" on:click|stopPropagation>
      <div class="modal-header">
        <div>
          <h2>{selectedApplicant.name}</h2>
          <p class="modal-subtitle">{selectedApplicant.position}</p>
        </div>
        <button class="close-btn" on:click={() => selectedApplicant = null}>×</button>
      </div>
      
      <div class="modal-body">
        <div class="detail-grid">
          <div class="detail-section">
            <div class="section-header">
              <h3>🤖 AI Analysis</h3>
              <button class="refresh-btn" on:click={() => analyzeApplicant(selectedApplicant.id)} disabled={analyzing}>
                {analyzing ? '⟳ Analyzing...' : '↻ Re-analyze'}
              </button>
            </div>
            
            <div class="overall-score">
              <div class="score-circle" style="border-color: {getScoreColor(selectedApplicant.ai_score)}">
                <div class="score-number" style="color: {getScoreColor(selectedApplicant.ai_score)}">
                  {selectedApplicant.ai_score}
                </div>
                <div class="score-label">Overall Score</div>
              </div>
            </div>
            
            <div class="score-details">
              {#each [
                { label: 'Skills Match', score: selectedApplicant.ai_analysis.skills_match },
                { label: 'Experience', score: selectedApplicant.ai_analysis.experience_score },
                { label: 'Education', score: selectedApplicant.ai_analysis.education_score },
                { label: 'Culture Fit', score: selectedApplicant.ai_analysis.culture_fit }
              ] as item}
                <div class="score-detail-item">
                  <div class="detail-header">
                    <span>{item.label}</span>
                    <span class="detail-score">{item.score}%</span>
                  </div>
                  <div class="detail-bar">
                    <div class="detail-fill" style="width: {item.score}%; background: {getScoreColor(item.score)}"></div>
                  </div>
                </div>
              {/each}
            </div>
            
            <div class="analysis-section">
              <h4>Strengths</h4>
              <ul class="strength-list">
                {#each selectedApplicant.ai_analysis.strengths as strength}
                  <li>✓ {strength}</li>
                {/each}
              </ul>
            </div>
            
            {#if selectedApplicant.ai_analysis.concerns.length > 0}
              <div class="analysis-section">
                <h4>Areas of Concern</h4>
                <ul class="concern-list">
                  {#each selectedApplicant.ai_analysis.concerns as concern}
                    <li>⚠ {concern}</li>
                  {/each}
                </ul>
              </div>
            {/if}
            
            <div class="analysis-summary">
              <h4>Summary</h4>
              <p>{selectedApplicant.ai_analysis.summary}</p>
            </div>
          </div>
          
          <div class="detail-section">
            <div class="section-header">
              <h3>📋 Applicant Details</h3>
            </div>
            
            <div class="info-grid">
              <div class="info-field">
                <label>Email</label>
                <div class="info-value">
                  <a href="mailto:{selectedApplicant.email}">{selectedApplicant.email}</a>
                </div>
              </div>
              
              <div class="info-field">
                <label>Phone</label>
                <div class="info-value">{selectedApplicant.phone}</div>
              </div>
              
              <div class="info-field">
                <label>Source</label>
                <div class="info-value">{selectedApplicant.source}</div>
              </div>
              
              <div class="info-field">
                <label>Applied Date</label>
                <div class="info-value">{new Date(selectedApplicant.applied_date).toLocaleDateString()}</div>
              </div>
            </div>
            
            <div class="actions-section">
              <h4>Quick Actions</h4>
              <div class="action-buttons">
                <button class="action-button primary" on:click={() => updateApplicantStatus(selectedApplicant.id, 'interview')}>
                  📅 Schedule Interview
                </button>
                <button class="action-button success" on:click={() => updateApplicantStatus(selectedApplicant.id, 'offer')}>
                  ✓ Send Offer
                </button>
                <button class="action-button" on:click={() => window.open(selectedApplicant.resume_url, '_blank')}>
                  📄 View Resume
                </button>
                <button class="action-button danger" on:click={() => updateApplicantStatus(selectedApplicant.id, 'rejected')}>
                  ✗ Reject
                </button>
              </div>
            </div>
            
            <div class="notes-section">
              <h4>Notes</h4>
              <textarea 
                class="notes-textarea" 
                placeholder="Add notes about this applicant..."
                value={selectedApplicant.notes || ''}
              ></textarea>
              <button class="save-notes-btn">Save Notes</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Upload Resume Modal -->
{#if showUploadModal}
  <div class="modal-overlay" on:click={closeUploadModal}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h3>📎 Upload Resume</h3>
        <button class="close-btn" on:click={closeUploadModal}>✕</button>
      </div>
      
      <div class="modal-body">
        {#if uploadSuccess}
          <div class="alert alert-success">
            ✓ {uploadSuccess}
          </div>
        {/if}
        
        {#if uploadError}
          <div class="alert alert-error">
            ⚠️ {uploadError}
          </div>
        {/if}
        
        <form on:submit|preventDefault={uploadResume}>
          <div class="form-group">
            <label for="name">Full Name *</label>
            <input 
              type="text" 
              id="name"
              bind:value={uploadForm.name}
              required
              placeholder="John Doe"
            />
          </div>
          
          <div class="form-group">
            <label for="email">Email *</label>
            <input 
              type="email" 
              id="email"
              bind:value={uploadForm.email}
              required
              placeholder="john@example.com"
            />
          </div>
          
          <div class="form-group">
            <label for="phone">Phone</label>
            <input 
              type="tel" 
              id="phone"
              bind:value={uploadForm.phone}
              placeholder="+1 (555) 123-4567"
            />
          </div>
          
          <div class="form-group">
            <label for="position">Position *</label>
            <input 
              type="text" 
              id="position"
              bind:value={uploadForm.position}
              required
              placeholder="Software Engineer"
            />
          </div>
          
          <div class="form-group">
            <label for="source">Source</label>
            <select id="source" bind:value={uploadForm.source}>
              <option value="manual">Manual Upload</option>
              <option value="linkedin">LinkedIn</option>
              <option value="indeed">Indeed</option>
              <option value="referral">Referral</option>
              <option value="website">Company Website</option>
              <option value="other">Other</option>
            </select>
          </div>
          
          <div class="form-group">
            <label for="resume">Resume File * (PDF or Word)</label>
            <input 
              type="file" 
              id="resume"
              accept=".pdf,.doc,.docx"
              on:change={handleFileSelect}
              required
            />
            {#if selectedFile}
              <p class="file-info">
                📄 {selectedFile.name} ({(selectedFile.size / 1024).toFixed(1)} KB)
              </p>
            {/if}
          </div>
          
          <div class="modal-footer">
            <button 
              type="button"
              class="btn-secondary" 
              on:click={closeUploadModal}
              disabled={uploading}
            >
              Cancel
            </button>
            <button 
              type="submit"
              class="btn-primary" 
              disabled={uploading || !selectedFile}
            >
              {uploading ? 'Uploading...' : 'Upload Resume'}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}

<style>
  .leaderboard-container { display: flex; flex-direction: column; gap: 24px; }
  .leaderboard-header { display: flex; justify-content: space-between; align-items: start; flex-wrap: wrap; gap: 16px; }
  .leaderboard-header h2 { font-size: 24px; font-weight: 700; color: #111827; margin: 0 0 8px 0; }
  .subtitle { font-size: 14px; color: #6b7280; margin: 0; }
  .controls { display: flex; gap: 12px; }
  .control-select { padding: 8px 16px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 14px; background: white; cursor: pointer; }
  .control-select:focus { outline: none; border-color: #3b82f6; }
  
  .btn-upload {
    padding: 10px 20px;
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    transition: all 0.2s;
    white-space: nowrap;
  }
  
  .btn-upload:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
  
  .leaderboard-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); gap: 20px; }
  .applicant-card { background: white; border-radius: 12px; padding: 20px; box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1); cursor: pointer; transition: all 0.2s; border: 2px solid transparent; }
  .applicant-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15); }
  .applicant-card.selected { border-color: #3b82f6; }
  .applicant-card.top-three { background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%); }
  
  .card-header { display: flex; align-items: start; gap: 12px; margin-bottom: 16px; }
  .rank-badge { font-size: 24px; min-width: 40px; text-align: center; font-weight: 700; color: #6b7280; }
  .rank-badge.top-three { color: #f59e0b; }
  .applicant-info { flex: 1; }
  .applicant-name { font-size: 18px; font-weight: 600; color: #111827; margin: 0 0 4px 0; }
  .applicant-position { font-size: 13px; color: #6b7280; margin: 0; }
  .score-badge { padding: 6px 12px; border-radius: 8px; font-size: 14px; font-weight: 700; }
  
  .card-body { margin-bottom: 16px; }
  .meta-row { display: flex; gap: 12px; margin-bottom: 12px; font-size: 13px; color: #6b7280; }
  .meta-item { display: flex; align-items: center; gap: 4px; }
  
  .score-breakdown { display: flex; flex-direction: column; gap: 8px; }
  .breakdown-item { }
  .breakdown-header { display: flex; justify-content: space-between; font-size: 12px; color: #6b7280; margin-bottom: 4px; }
  .breakdown-bar { width: 100%; height: 6px; background: #e5e7eb; border-radius: 3px; overflow: hidden; }
  .breakdown-fill { height: 100%; border-radius: 3px; transition: width 0.3s; }
  
  .card-footer { display: flex; justify-content: space-between; align-items: center; padding-top: 16px; border-top: 1px solid #e5e7eb; }
  .status-badge { padding: 4px 12px; border-radius: 6px; font-size: 12px; font-weight: 500; }
  .status-badge.status-new { background: #dbeafe; color: #1e40af; }
  .status-badge.status-reviewing { background: #fef3c7; color: #92400e; }
  .status-badge.status-interview { background: #e0e7ff; color: #3730a3; }
  .status-badge.status-offer { background: #d1fae5; color: #065f46; }
  .status-badge.status-rejected { background: #fee2e2; color: #991b1b; }
  .view-btn { padding: 6px 12px; background: none; border: none; color: #3b82f6; font-size: 13px; font-weight: 500; cursor: pointer; }
  .view-btn:hover { text-decoration: underline; }
  
  .loading-state, .empty-state { text-align: center; padding: 60px 20px; background: white; border-radius: 12px; }
  .empty-icon { font-size: 64px; display: block; margin-bottom: 16px; }
  .spinner { width: 40px; height: 40px; border: 4px solid #e5e7eb; border-top-color: #3b82f6; border-radius: 50%; animation: spin 1s linear infinite; margin: 0 auto 16px; }
  @keyframes spin { to { transform: rotate(360deg); } }
  
  .modal-overlay { position: fixed; inset: 0; background: rgba(0, 0, 0, 0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; padding: 20px; }
  .modal-large { background: white; border-radius: 12px; max-width: 1200px; width: 100%; max-height: 90vh; overflow-y: auto; }
  .modal-header { display: flex; justify-content: space-between; align-items: start; padding: 24px; border-bottom: 1px solid #e5e7eb; }
  .modal-header h2 { font-size: 24px; font-weight: 600; margin: 0 0 4px 0; }
  .modal-subtitle { font-size: 14px; color: #6b7280; margin: 0; }
  .close-btn { background: none; border: none; font-size: 28px; cursor: pointer; color: #6b7280; width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: 6px; }
  .close-btn:hover { background: #f3f4f6; }
  
  .modal-body { padding: 24px; }
  .detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
  .detail-section { display: flex; flex-direction: column; gap: 20px; }
  .section-header { display: flex; justify-content: space-between; align-items: center; }
  .section-header h3 { font-size: 18px; font-weight: 600; margin: 0; }
  .refresh-btn { padding: 6px 12px; background: #eff6ff; color: #3b82f6; border: 1px solid #3b82f6; border-radius: 6px; font-size: 13px; cursor: pointer; }
  .refresh-btn:hover { background: #3b82f6; color: white; }
  .refresh-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  
  .overall-score { display: flex; justify-content: center; padding: 20px; }
  .score-circle { width: 150px; height: 150px; border: 8px solid; border-radius: 50%; display: flex; flex-direction: column; align-items: center; justify-content: center; }
  .score-number { font-size: 48px; font-weight: 700; }
  .score-label { font-size: 13px; color: #6b7280; margin-top: 4px; }
  
  .score-details { display: flex; flex-direction: column; gap: 16px; }
  .score-detail-item { }
  .detail-header { display: flex; justify-content: space-between; font-size: 14px; margin-bottom: 6px; }
  .detail-score { font-weight: 600; }
  .detail-bar { width: 100%; height: 8px; background: #e5e7eb; border-radius: 4px; overflow: hidden; }
  .detail-fill { height: 100%; border-radius: 4px; }
  
  .analysis-section { }
  .analysis-section h4 { font-size: 14px; font-weight: 600; margin: 0 0 12px 0; }
  .strength-list, .concern-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 8px; }
  .strength-list li { color: #059669; font-size: 14px; }
  .concern-list li { color: #dc2626; font-size: 14px; }
  .analysis-summary { }
  .analysis-summary p { font-size: 14px; line-height: 1.6; color: #374151; margin: 0; }
  
  .info-grid { display: grid; gap: 16px; }
  .info-field { }
  .info-field label { display: block; font-size: 12px; font-weight: 500; color: #6b7280; margin-bottom: 4px; }
  .info-value { font-size: 14px; color: #111827; }
  .info-value a { color: #3b82f6; text-decoration: none; }
  .info-value a:hover { text-decoration: underline; }
  
  .actions-section { }
  .actions-section h4 { font-size: 14px; font-weight: 600; margin: 0 0 12px 0; }
  .action-buttons { display: grid; gap: 8px; }
  .action-button { padding: 10px 16px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 14px; font-weight: 500; cursor: pointer; transition: all 0.2s; }
  .action-button.primary { background: #3b82f6; border-color: #3b82f6; color: white; }
  .action-button.success { background: #10b981; border-color: #10b981; color: white; }
  .action-button.danger { background: #ef4444; border-color: #ef4444; color: white; }
  .action-button:hover { transform: translateY(-1px); box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15); }
  
  .notes-section { }
  .notes-section h4 { font-size: 14px; font-weight: 600; margin: 0 0 12px 0; }
  .notes-textarea { width: 100%; min-height: 120px; padding: 12px; border: 1px solid #d1d5db; border-radius: 8px; font-size: 14px; font-family: inherit; resize: vertical; }
  .notes-textarea:focus { outline: none; border-color: #3b82f6; }
  .save-notes-btn { margin-top: 8px; padding: 8px 16px; background: #3b82f6; color: white; border: none; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; }
  .save-notes-btn:hover { background: #2563eb; }
  
  /* Upload Modal */
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
    padding: 20px;
  }
  
  .modal {
    background: white;
    border-radius: 12px;
    max-width: 600px;
    width: 100%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 24px;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .modal-header h3 {
    font-size: 20px;
    font-weight: 700;
    color: #111827;
    margin: 0;
  }
  
  .close-btn {
    background: none;
    border: none;
    font-size: 24px;
    color: #9ca3af;
    cursor: pointer;
    transition: color 0.2s;
  }
  
  .close-btn:hover {
    color: #374151;
  }
  
  .modal-body {
    padding: 24px;
  }
  
  .form-group {
    margin-bottom: 20px;
  }
  
  .form-group label {
    display: block;
    font-size: 14px;
    font-weight: 600;
    color: #374151;
    margin-bottom: 8px;
  }
  
  .form-group input[type="text"],
  .form-group input[type="email"],
  .form-group input[type="tel"],
  .form-group select {
    width: 100%;
    padding: 10px 14px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    transition: border-color 0.2s;
  }
  
  .form-group input:focus,
  .form-group select:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }
  
  .form-group input[type="file"] {
    width: 100%;
    padding: 10px;
    border: 2px dashed #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    cursor: pointer;
  }
  
  .file-info {
    margin-top: 8px;
    font-size: 13px;
    color: #10b981;
    font-weight: 500;
  }
  
  .alert {
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 20px;
    font-size: 14px;
    font-weight: 500;
  }
  
  .alert-success {
    background: #d1fae5;
    color: #065f46;
    border: 1px solid #6ee7b7;
  }
  
  .alert-error {
    background: #fee2e2;
    color: #991b1b;
    border: 1px solid #fca5a5;
  }
  
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 24px;
  }
  
  .btn-primary {
    padding: 10px 24px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
    transform: translateY(-1px);
  }
  
  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  .btn-secondary {
    padding: 10px 24px;
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-secondary:hover:not(:disabled) {
    background: #f9fafb;
    border-color: #9ca3af;
  }
  
  .btn-secondary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  @media (max-width: 1024px) {
    .detail-grid { grid-template-columns: 1fr; }
    .leaderboard-grid { grid-template-columns: 1fr; }
  }
</style>