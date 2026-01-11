<script lang="ts">
  import { onMount } from 'svelte';
  
  interface Job {
    id: string;
    title: string;
    department: string;
    location: string;
    status: string;
    applicant_count: number;
  }
  
  interface Candidate {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    phone: string;
    score: number;
    skills: string[];
    strengths: string[];
    weaknesses: string[];
    experience_years: number;
    ai_summary: string;
    applied_date: string;
  }
  
  interface Summary {
    job_title: string;
    total_candidates: number;
    top_candidates: Candidate[];
    comparative_analysis: string;
    recommendations: string[];
    skill_gaps: string[];
    hiring_insights: string;
    generated_at: string;
  }
  
  let jobs: Job[] = [];
  let selectedJobId: string = '';
  let topN: number = 5;
  let loadingJobs = true;
  let generating = false;
  let summary: Summary | null = null;
  let error = '';
  
  $: selectedJob = jobs.find(j => j.id === selectedJobId);
  
  onMount(async () => {
    await loadJobs();
  });
  
  async function loadJobs() {
    try {
      loadingJobs = true;
      const token = localStorage.getItem('token');
      const response = await fetch('/api/recruiting/jobs', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        jobs = await response.json();
        // Auto-select first job with applicants
        const jobWithApplicants = jobs.find(j => j.applicant_count > 0);
        if (jobWithApplicants) {
          selectedJobId = jobWithApplicants.id;
        }
      }
    } catch (err) {
      console.error('Failed to load jobs:', err);
    } finally {
      loadingJobs = false;
    }
  }
  
  async function generateSummary() {
    if (!selectedJobId) {
      error = 'Please select a job posting';
      return;
    }
    
    try {
      generating = true;
      error = '';
      summary = null;
      
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/recruiting/jobs/${selectedJobId}/candidate-summary`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ top_n: topN })
      });
      
      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to generate summary');
      }
      
      summary = await response.json();
    } catch (err: any) {
      error = err.message;
    } finally {
      generating = false;
    }
  }
  
  function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
  
  function exportSummary() {
    if (!summary) return;
    
    const content = `
CANDIDATE SUMMARY REPORT
========================

Job: ${summary.job_title}
Generated: ${formatDate(summary.generated_at)}
Total Candidates Reviewed: ${summary.total_candidates}
Top ${topN} Candidates Selected

TOP CANDIDATES
==============

${summary.top_candidates.map((c, i) => `
${i + 1}. ${c.first_name} ${c.last_name}
   Score: ${c.score}/100
   Experience: ${c.experience_years || 'N/A'} years
   Email: ${c.email}
   Phone: ${c.phone}
   Applied: ${formatDate(c.applied_date)}
   
   Skills: ${c.skills.join(', ')}
   
   Strengths:
   ${c.strengths.map(s => `   - ${s}`).join('\n')}
   
   Concerns:
   ${c.weaknesses.map(w => `   - ${w}`).join('\n')}
   
   Summary: ${c.ai_summary || 'No summary available'}
   
`).join('\n---\n')}

COMPARATIVE ANALYSIS
====================

${summary.comparative_analysis}

RECOMMENDATIONS
===============

${summary.recommendations.map((r, i) => `${i + 1}. ${r}`).join('\n')}

SKILL GAPS IDENTIFIED
=====================

${summary.skill_gaps.map((g, i) => `${i + 1}. ${g}`).join('\n')}

HIRING INSIGHTS
===============

${summary.hiring_insights}
    `.trim();
    
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `candidate-summary-${summary.job_title.replace(/\s+/g, '-')}-${Date.now()}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }
</script>

<div class="candidate-summary">
  <!-- Configuration Section -->
  <div class="config-section">
    <div class="config-card">
      <h3>🎯 Configure Summary</h3>
      <p class="help-text">Select a job posting and specify how many top candidates to analyze</p>
      
      <div class="form-grid">
        <div class="form-group">
          <label for="job-select">Job Posting</label>
          <select 
            id="job-select"
            bind:value={selectedJobId}
            disabled={loadingJobs}
            class="select-input"
          >
            <option value="">Select a job...</option>
            {#each jobs as job}
              <option value={job.id}>
                {job.title} ({job.applicant_count} applicants)
              </option>
            {/each}
          </select>
        </div>
        
        <div class="form-group">
          <label for="top-n">Number of Top Candidates</label>
          <select id="top-n" bind:value={topN} class="select-input">
            <option value={3}>Top 3</option>
            <option value={5}>Top 5</option>
            <option value={10}>Top 10</option>
            <option value={15}>Top 15</option>
          </select>
        </div>
      </div>
      
      {#if selectedJob}
        <div class="job-preview">
          <div class="job-preview-item">
            <span class="label">Position:</span>
            <span class="value">{selectedJob.title}</span>
          </div>
          <div class="job-preview-item">
            <span class="label">Department:</span>
            <span class="value">{selectedJob.department}</span>
          </div>
          <div class="job-preview-item">
            <span class="label">Location:</span>
            <span class="value">{selectedJob.location}</span>
          </div>
          <div class="job-preview-item">
            <span class="label">Applicants:</span>
            <span class="value">{selectedJob.applicant_count}</span>
          </div>
        </div>
      {/if}
      
      <button 
        class="btn-primary"
        on:click={generateSummary}
        disabled={!selectedJobId || generating || (selectedJob && selectedJob.applicant_count === 0)}
      >
        {#if generating}
          <span class="spinner"></span>
          Generating Summary...
        {:else}
          ✨ Generate Candidate Summary
        {/if}
      </button>
      
      {#if selectedJob && selectedJob.applicant_count === 0}
        <p class="warning-text">⚠️ This job has no applicants yet</p>
      {/if}
      
      {#if error}
        <div class="error-message">
          ❌ {error}
        </div>
      {/if}
    </div>
  </div>
  
  <!-- Summary Results -->
  {#if summary}
    <div class="summary-results">
      <!-- Header -->
      <div class="summary-header">
        <div class="header-content">
          <h2>📊 Candidate Summary Report</h2>
          <div class="meta-info">
            <span>📋 {summary.job_title}</span>
            <span>•</span>
            <span>👥 {summary.total_candidates} candidates reviewed</span>
            <span>•</span>
            <span>🕐 {formatDate(summary.generated_at)}</span>
          </div>
        </div>
        <button class="btn-export" on:click={exportSummary}>
          📥 Export Report
        </button>
      </div>
      
      <!-- Top Candidates Grid -->
      <div class="section">
        <h3>🏆 Top {summary.top_candidates.length} Candidates</h3>
        <div class="candidates-grid">
          {#each summary.top_candidates as candidate, index}
            <div class="candidate-card" class:gold={index === 0} class:silver={index === 1} class:bronze={index === 2}>
              <div class="candidate-header">
                <div class="rank-badge">
                  {#if index === 0}🥇
                  {:else if index === 1}🥈
                  {:else if index === 2}🥉
                  {:else}#{index + 1}
                  {/if}
                </div>
                <div class="candidate-name">
                  <h4>{candidate.first_name} {candidate.last_name}</h4>
                  <p class="candidate-contact">{candidate.email}</p>
                </div>
                <div class="candidate-score">
                  <div class="score-circle" class:high={candidate.score >= 80} class:medium={candidate.score >= 60 && candidate.score < 80} class:low={candidate.score < 60}>
                    {candidate.score}
                  </div>
                  <span class="score-label">Score</span>
                </div>
              </div>
              
              <div class="candidate-details">
                <div class="detail-row">
                  <span class="icon">📞</span>
                  <span>{candidate.phone}</span>
                </div>
                <div class="detail-row">
                  <span class="icon">💼</span>
                  <span>{candidate.experience_years || 'N/A'} years experience</span>
                </div>
                <div class="detail-row">
                  <span class="icon">📅</span>
                  <span>Applied {formatDate(candidate.applied_date)}</span>
                </div>
              </div>
              
              {#if candidate.skills && candidate.skills.length > 0}
                <div class="skills-section">
                  <p class="section-label">Skills:</p>
                  <div class="skills-tags">
                    {#each candidate.skills.slice(0, 5) as skill}
                      <span class="skill-tag">{skill}</span>
                    {/each}
                    {#if candidate.skills.length > 5}
                      <span class="skill-tag more">+{candidate.skills.length - 5} more</span>
                    {/if}
                  </div>
                </div>
              {/if}
              
              {#if candidate.strengths && candidate.strengths.length > 0}
                <div class="assessment-section positive">
                  <p class="section-label">✅ Strengths:</p>
                  <ul>
                    {#each candidate.strengths.slice(0, 3) as strength}
                      <li>{strength}</li>
                    {/each}
                  </ul>
                </div>
              {/if}
              
              {#if candidate.weaknesses && candidate.weaknesses.length > 0}
                <div class="assessment-section negative">
                  <p class="section-label">⚠️ Considerations:</p>
                  <ul>
                    {#each candidate.weaknesses.slice(0, 2) as weakness}
                      <li>{weakness}</li>
                    {/each}
                  </ul>
                </div>
              {/if}
              
              {#if candidate.ai_summary}
                <div class="ai-summary-section">
                  <p class="section-label">🤖 AI Analysis:</p>
                  <p class="ai-summary-text">{candidate.ai_summary}</p>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
      
      <!-- Comparative Analysis -->
      <div class="section">
        <h3>🔍 Comparative Analysis</h3>
        <div class="analysis-box">
          <p>{summary.comparative_analysis}</p>
        </div>
      </div>
      
      <!-- Recommendations -->
      <div class="section">
        <h3>💡 Recommendations</h3>
        <div class="recommendations-list">
          {#each summary.recommendations as recommendation, index}
            <div class="recommendation-item">
              <span class="recommendation-number">{index + 1}</span>
              <p>{recommendation}</p>
            </div>
          {/each}
        </div>
      </div>
      
      <!-- Skill Gaps -->
      {#if summary.skill_gaps && summary.skill_gaps.length > 0}
        <div class="section">
          <h3>📊 Skill Gaps Identified</h3>
          <div class="skill-gaps-grid">
            {#each summary.skill_gaps as gap}
              <div class="skill-gap-item">
                <span class="gap-icon">⚠️</span>
                <p>{gap}</p>
              </div>
            {/each}
          </div>
        </div>
      {/if}
      
      <!-- Hiring Insights -->
      <div class="section">
        <h3>🎯 Hiring Insights</h3>
        <div class="insights-box">
          <p>{summary.hiring_insights}</p>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .candidate-summary {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .config-section {
    display: flex;
    justify-content: center;
  }
  
  .config-card {
    background: white;
    border-radius: 12px;
    padding: 32px;
    border: 1px solid #e5e7eb;
    max-width: 800px;
    width: 100%;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  }
  
  .config-card h3 {
    font-size: 24px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .help-text {
    color: #6b7280;
    font-size: 14px;
    margin: 0 0 24px 0;
  }
  
  .form-grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 16px;
    margin-bottom: 24px;
  }
  
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  
  .form-group label {
    font-size: 14px;
    font-weight: 600;
    color: #374151;
  }
  
  .select-input {
    padding: 12px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 14px;
    background: white;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .select-input:hover {
    border-color: #9ca3af;
  }
  
  .select-input:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }
  
  .job-preview {
    background: #f9fafb;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 24px;
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }
  
  .job-preview-item {
    display: flex;
    gap: 8px;
  }
  
  .job-preview-item .label {
    font-size: 13px;
    font-weight: 600;
    color: #6b7280;
  }
  
  .job-preview-item .value {
    font-size: 13px;
    color: #111827;
  }
  
  .btn-primary {
    width: 100%;
    padding: 14px 24px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
  }
  
  .btn-primary:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 8px 16px rgba(102, 126, 234, 0.3);
  }
  
  .btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
  
  .warning-text {
    color: #f59e0b;
    font-size: 13px;
    margin: 8px 0 0 0;
    text-align: center;
  }
  
  .error-message {
    background: #fee2e2;
    color: #991b1b;
    padding: 12px;
    border-radius: 6px;
    font-size: 14px;
    margin-top: 16px;
  }
  
  .summary-results {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  
  .summary-header {
    background: white;
    border-radius: 12px;
    padding: 24px;
    border: 1px solid #e5e7eb;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .header-content h2 {
    font-size: 24px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 8px 0;
  }
  
  .meta-info {
    display: flex;
    gap: 8px;
    font-size: 13px;
    color: #6b7280;
  }
  
  .btn-export {
    padding: 10px 20px;
    background: #f3f4f6;
    color: #374151;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-export:hover {
    background: #e5e7eb;
  }
  
  .section {
    background: white;
    border-radius: 12px;
    padding: 24px;
    border: 1px solid #e5e7eb;
  }
  
  .section h3 {
    font-size: 20px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 20px 0;
  }
  
  .candidates-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
    gap: 20px;
  }
  
  .candidate-card {
    background: #f9fafb;
    border: 2px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    transition: all 0.2s;
  }
  
  .candidate-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }
  
  .candidate-card.gold {
    border-color: #fbbf24;
    background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
  }
  
  .candidate-card.silver {
    border-color: #9ca3af;
    background: linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%);
  }
  
  .candidate-card.bronze {
    border-color: #f59e0b;
    background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
  }
  
  .candidate-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid rgba(0, 0, 0, 0.1);
  }
  
  .rank-badge {
    font-size: 32px;
    flex-shrink: 0;
  }
  
  .candidate-name {
    flex: 1;
  }
  
  .candidate-name h4 {
    font-size: 18px;
    font-weight: 700;
    color: #111827;
    margin: 0 0 4px 0;
  }
  
  .candidate-contact {
    font-size: 13px;
    color: #6b7280;
    margin: 0;
  }
  
  .candidate-score {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }
  
  .score-circle {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    font-weight: 700;
    color: white;
  }
  
  .score-circle.high {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  }
  
  .score-circle.medium {
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  }
  
  .score-circle.low {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  }
  
  .score-label {
    font-size: 11px;
    color: #6b7280;
    text-transform: uppercase;
  }
  
  .candidate-details {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 16px;
  }
  
  .detail-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: #374151;
  }
  
  .detail-row .icon {
    font-size: 16px;
  }
  
  .skills-section,
  .assessment-section,
  .ai-summary-section {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid rgba(0, 0, 0, 0.1);
  }
  
  .section-label {
    font-size: 12px;
    font-weight: 600;
    color: #374151;
    margin: 0 0 8px 0;
  }
  
  .skills-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  
  .skill-tag {
    background: white;
    border: 1px solid #d1d5db;
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 12px;
    color: #374151;
  }
  
  .skill-tag.more {
    background: #e5e7eb;
    border-color: #9ca3af;
  }
  
  .assessment-section ul {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  
  .assessment-section li {
    font-size: 13px;
    padding-left: 16px;
    position: relative;
  }
  
  .assessment-section.positive li {
    color: #065f46;
  }
  
  .assessment-section.positive li::before {
    content: '•';
    position: absolute;
    left: 0;
    color: #10b981;
    font-weight: 700;
  }
  
  .assessment-section.negative li {
    color: #92400e;
  }
  
  .assessment-section.negative li::before {
    content: '•';
    position: absolute;
    left: 0;
    color: #f59e0b;
    font-weight: 700;
  }
  
  .ai-summary-text {
    font-size: 13px;
    color: #374151;
    line-height: 1.6;
    margin: 0;
  }
  
  .analysis-box,
  .insights-box {
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    padding: 20px;
  }
  
  .analysis-box p,
  .insights-box p {
    font-size: 14px;
    color: #374151;
    line-height: 1.7;
    margin: 0;
    white-space: pre-wrap;
  }
  
  .recommendations-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .recommendation-item {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }
  
  .recommendation-number {
    background: #667eea;
    color: white;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 700;
    flex-shrink: 0;
  }
  
  .recommendation-item p {
    font-size: 14px;
    color: #374151;
    line-height: 1.6;
    margin: 4px 0 0 0;
  }
  
  .skill-gaps-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 12px;
  }
  
  .skill-gap-item {
    background: #fef3c7;
    border: 1px solid #fbbf24;
    border-radius: 8px;
    padding: 12px;
    display: flex;
    gap: 10px;
    align-items: flex-start;
  }
  
  .gap-icon {
    font-size: 18px;
    flex-shrink: 0;
  }
  
  .skill-gap-item p {
    font-size: 13px;
    color: #92400e;
    margin: 0;
    line-height: 1.5;
  }
  
  @media (max-width: 768px) {
    .form-grid {
      grid-template-columns: 1fr;
    }
    
    .job-preview {
      grid-template-columns: 1fr;
    }
    
    .candidates-grid {
      grid-template-columns: 1fr;
    }
    
    .summary-header {
      flex-direction: column;
      gap: 16px;
      align-items: flex-start;
    }
  }
</style>
