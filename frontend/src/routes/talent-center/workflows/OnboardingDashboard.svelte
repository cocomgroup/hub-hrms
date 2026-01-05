<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '../../../stores/auth';
	import { getApiBaseUrl } from '../../../lib/api';
	
	const API_BASE_URL = getApiBaseUrl();
	
	interface OnboardingWorkflow {
		id: string;
		employee_id: string;
		employee_name: string;
		employee_email: string;
		start_date: string;
		status: string;
		overall_progress: number;
		buddy_name: string;
		manager_name: string;
	}
	
	interface DashboardData {
		total_onboardings: number;
		active_onboardings: number;
		completed_onboardings: number;
		overdue_onboardings: number;
		recent_onboardings: OnboardingWorkflow[];
		upcoming_tasks: any[];
		ai_insights: string[];
	}
	
	let dashboardData: DashboardData = {
		total_onboardings: 0,
		active_onboardings: 0,
		completed_onboardings: 0,
		overdue_onboardings: 0,
		recent_onboardings: [],
		upcoming_tasks: [],
		ai_insights: []
	};
	
	let loading = true;
	let error = '';
	let selectedStatus = 'all';
	let selectedWorkflowId: string | null = null;
	let showWorkflowDetail = false;
	let showCreateModal = false;
	
	onMount(async () => {
		await loadDashboard();
	});
	
	async function loadDashboard() {
		try {
			loading = true;
			const token = $authStore.token || localStorage.getItem('token');
			const response = await fetch(`${API_BASE_URL}/onboarding/dashboard`, {
				headers: {
					'Authorization': `Bearer ${token}`
				}
			});
			
			if (!response.ok) throw new Error('Failed to load dashboard');
			
			dashboardData = await response.json();
		} catch (err: any) {
			error = err.message;
		} finally {
			loading = false;
		}
	}
	
	function getStatusColor(status: string): string {
		const colors = {
			'not_started': 'bg-gray-100 text-gray-800',
			'in_progress': 'bg-blue-100 text-blue-800',
			'completed': 'bg-green-100 text-green-800',
			'overdue': 'bg-red-100 text-red-800'
		};
		return colors[status] || 'bg-gray-100 text-gray-800';
	}
	
	function viewOnboarding(id: string) {
		selectedWorkflowId = id;
		showWorkflowDetail = true;
	}
	
	function closeWorkflowDetail() {
		showWorkflowDetail = false;
		selectedWorkflowId = null;
		loadDashboard(); // Refresh data
	}
	
	function createNewOnboarding() {
		showCreateModal = true;
	}
	
	function closeCreateModal() {
		showCreateModal = false;
		loadDashboard(); // Refresh data
	}
</script>

<div class="onboarding-dashboard">
	<!-- Header -->
	<div class="dashboard-header">
		<div>
			<h1>Onboarding Dashboard</h1>
			<p class="subtitle">Track and manage new hire onboarding</p>
		</div>
		<button 
			class="btn-primary"
			on:click={createNewOnboarding}
		>
			+ New Onboarding
		</button>
	</div>
	
	{#if loading}
		<div class="loading">Loading dashboard...</div>
	{:else if error}
		<div class="error">
			<p>Error: {error}</p>
			<button on:click={loadDashboard}>Retry</button>
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="stats-grid">
			<div class="stat-card">
				<div class="stat-icon">👥</div>
				<div class="stat-content">
					<div class="stat-value">{dashboardData.total_onboardings}</div>
					<div class="stat-label">Total</div>
				</div>
			</div>
			
			<div class="stat-card">
				<div class="stat-icon">🚀</div>
				<div class="stat-content">
					<div class="stat-value">{dashboardData.active_onboardings}</div>
					<div class="stat-label">Active</div>
				</div>
			</div>
			
			<div class="stat-card">
				<div class="stat-icon">✓</div>
				<div class="stat-content">
					<div class="stat-value">{dashboardData.completed_onboardings}</div>
					<div class="stat-label">Completed</div>
				</div>
			</div>
			
			<div class="stat-card">
				<div class="stat-icon">⏰</div>
				<div class="stat-content">
					<div class="stat-value">{dashboardData.overdue_onboardings}</div>
					<div class="stat-label">Overdue</div>
				</div>
			</div>
		</div>
		
		<!-- AI Insights -->
		{#if dashboardData.ai_insights && dashboardData.ai_insights.length > 0}
			<div class="ai-insights">
				<div class="insights-header">
					<span class="insights-icon">💡</span>
					<h3>AI Insights</h3>
				</div>
				<div class="insights-list">
					{#each dashboardData.ai_insights as insight}
						<p class="insight-item">
							<span class="bullet">•</span>
							{insight}
						</p>
					{/each}
				</div>
			</div>
		{/if}
		
		<!-- Recent Onboardings Table -->
		<div class="section-card">
			<div class="section-header">
				<h2>Recent Onboardings</h2>
			</div>
			
			<div class="table-container">
				<table class="onboarding-table">
					<thead>
						<tr>
							<th>Employee</th>
							<th>Start Date</th>
							<th>Status</th>
							<th>Progress</th>
							<th>Manager</th>
							<th>Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each dashboardData.recent_onboardings as workflow}
							<tr class="table-row" on:click={() => viewOnboarding(workflow.id)}>
								<td>
									<div class="employee-cell">
										<div class="employee-avatar">
											{workflow.employee_name.split(' ').map(n => n[0]).join('')}
										</div>
										<div class="employee-info">
											<div class="employee-name">{workflow.employee_name}</div>
											<div class="employee-email">{workflow.employee_email}</div>
										</div>
									</div>
								</td>
								<td>
									<span class="date-text">
										{new Date(workflow.start_date).toLocaleDateString()}
									</span>
								</td>
								<td>
									<span class="status-badge {getStatusColor(workflow.status)}">
										{workflow.status.replace('_', ' ')}
									</span>
								</td>
								<td>
									<div class="progress-cell">
										<div class="progress-bar-container">
											<div class="progress-bar-fill" style="width: {workflow.overall_progress}%"></div>
										</div>
										<span class="progress-text">{workflow.overall_progress}%</span>
									</div>
								</td>
								<td>
									<span class="manager-text">{workflow.manager_name || '-'}</span>
								</td>
								<td>
									<button 
										class="action-btn"
										on:click|stopPropagation={() => viewOnboarding(workflow.id)}
									>
										View Details
									</button>
								</td>
							</tr>
						{:else}
							<tr>
								<td colspan="6" class="empty-row">
									No onboarding workflows found
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

<!-- Workflow Detail Modal -->
{#if showWorkflowDetail && selectedWorkflowId}
	<div class="modal-overlay" on:click={closeWorkflowDetail}>
		<div class="modal large" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Onboarding Workflow Details</h2>
				<button class="close-btn" on:click={closeWorkflowDetail}>×</button>
			</div>
			<div class="modal-body">
				<p>Workflow ID: {selectedWorkflowId}</p>
				<p class="text-muted">Detailed workflow view coming soon...</p>
				<!-- TODO: Import and use OnboardingWorkflowView component here -->
			</div>
		</div>
	</div>
{/if}

<!-- Create Onboarding Modal -->
{#if showCreateModal}
	<div class="modal-overlay" on:click={closeCreateModal}>
		<div class="modal" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Create New Onboarding</h2>
				<button class="close-btn" on:click={closeCreateModal}>×</button>
			</div>
			<div class="modal-body">
				<p class="text-muted">Create onboarding form coming soon...</p>
				<!-- TODO: Add create onboarding form here -->
			</div>
		</div>
	</div>
{/if}

<style>
	.onboarding-dashboard {
		max-width: 1400px;
		margin: 0 auto;
		padding: 24px;
	}
	
	.dashboard-header {
		display: flex;
		justify-content: space-between;
		align-items: start;
		margin-bottom: 32px;
	}
	
	.dashboard-header h1 {
		font-size: 32px;
		font-weight: 700;
		color: #111827;
		margin: 0 0 8px 0;
	}
	
	.subtitle {
		font-size: 14px;
		color: #6b7280;
		margin: 0;
	}
	
	.btn-primary {
		padding: 12px 24px;
		background: #3b82f6;
		color: white;
		border: none;
		border-radius: 8px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: background 0.2s;
	}
	
	.btn-primary:hover {
		background: #2563eb;
	}
	
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 20px;
		margin-bottom: 32px;
	}
	
	.stat-card {
		background: white;
		padding: 24px;
		border-radius: 12px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		display: flex;
		align-items: center;
		gap: 16px;
	}
	
	.stat-icon {
		font-size: 36px;
		width: 60px;
		height: 60px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #f3f4f6;
		border-radius: 12px;
	}
	
	.stat-content {
		flex: 1;
	}
	
	.stat-value {
		font-size: 28px;
		font-weight: 600;
		color: #111827;
		margin-bottom: 4px;
	}
	
	.stat-label {
		font-size: 14px;
		color: #6b7280;
	}
	
	.ai-insights {
		background: linear-gradient(135deg, #ede9fe 0%, #ddd6fe 100%);
		padding: 24px;
		border-radius: 12px;
		margin-bottom: 32px;
		border-left: 4px solid #8b5cf6;
	}
	
	.insights-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
	}
	
	.insights-icon {
		font-size: 24px;
	}
	
	.insights-header h3 {
		font-size: 18px;
		font-weight: 600;
		color: #5b21b6;
		margin: 0;
	}
	
	.insights-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	
	.insight-item {
		display: flex;
		align-items: start;
		gap: 8px;
		color: #6b21a8;
		font-size: 14px;
		line-height: 1.5;
		margin: 0;
	}
	
	.bullet {
		color: #8b5cf6;
		font-weight: 700;
	}
	
	.section-card {
		background: white;
		border-radius: 12px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		overflow: hidden;
	}
	
	.section-header {
		padding: 24px;
		border-bottom: 1px solid #e5e7eb;
	}
	
	.section-header h2 {
		font-size: 20px;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}
	
	.table-container {
		overflow-x: auto;
	}
	
	.onboarding-table {
		width: 100%;
		border-collapse: collapse;
	}
	
	.onboarding-table thead {
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
	}
	
	.onboarding-table th {
		padding: 12px 24px;
		text-align: left;
		font-size: 12px;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}
	
	.table-row {
		border-bottom: 1px solid #e5e7eb;
		cursor: pointer;
		transition: background 0.2s;
	}
	
	.table-row:hover {
		background: #f9fafb;
	}
	
	.onboarding-table td {
		padding: 16px 24px;
	}
	
	.employee-cell {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	
	.employee-avatar {
		width: 40px;
		height: 40px;
		background: #3b82f6;
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: 600;
	}
	
	.employee-info {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	
	.employee-name {
		font-size: 14px;
		font-weight: 500;
		color: #111827;
	}
	
	.employee-email {
		font-size: 13px;
		color: #6b7280;
	}
	
	.date-text {
		font-size: 14px;
		color: #374151;
	}
	
	.status-badge {
		display: inline-block;
		padding: 4px 12px;
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
	}
	
	.bg-gray-100 {
		background: #f3f4f6;
	}
	
	.text-gray-800 {
		color: #1f2937;
	}
	
	.bg-blue-100 {
		background: #dbeafe;
	}
	
	.text-blue-800 {
		color: #1e40af;
	}
	
	.bg-green-100 {
		background: #d1fae5;
	}
	
	.text-green-800 {
		color: #065f46;
	}
	
	.bg-red-100 {
		background: #fee2e2;
	}
	
	.text-red-800 {
		color: #991b1b;
	}
	
	.progress-cell {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	
	.progress-bar-container {
		flex: 1;
		height: 8px;
		background: #e5e7eb;
		border-radius: 4px;
		overflow: hidden;
	}
	
	.progress-bar-fill {
		height: 100%;
		background: #3b82f6;
		border-radius: 4px;
		transition: width 0.3s;
	}
	
	.progress-text {
		font-size: 13px;
		font-weight: 500;
		color: #374151;
		min-width: 40px;
	}
	
	.manager-text {
		font-size: 14px;
		color: #374151;
	}
	
	.action-btn {
		padding: 6px 16px;
		background: white;
		color: #3b82f6;
		border: 1px solid #3b82f6;
		border-radius: 6px;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.action-btn:hover {
		background: #3b82f6;
		color: white;
	}
	
	.empty-row {
		text-align: center;
		padding: 48px 24px;
		color: #6b7280;
	}
	
	.loading,
	.error {
		text-align: center;
		padding: 48px;
		color: #666;
	}
	
	.error {
		color: #dc2626;
	}
	
	.error button {
		margin-top: 16px;
		padding: 10px 20px;
		background: #3b82f6;
		color: white;
		border: none;
		border-radius: 6px;
		cursor: pointer;
	}
	
	.text-muted {
		color: #6b7280;
		font-style: italic;
	}
	
	/* Modal Styles */
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
	
	.modal.large {
		max-width: 1000px;
	}
	
	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 24px;
		border-bottom: 1px solid #e5e7eb;
	}
	
	.modal-header h2 {
		margin: 0;
		font-size: 24px;
		font-weight: 600;
		color: #111827;
	}
	
	.close-btn {
		background: none;
		border: none;
		font-size: 32px;
		cursor: pointer;
		color: #6b7280;
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 6px;
		transition: background 0.2s;
	}
	
	.close-btn:hover {
		background: #f3f4f6;
	}
	
	.modal-body {
		padding: 24px;
	}
	
	@media (max-width: 768px) {
		.onboarding-dashboard {
			padding: 16px;
		}
		
		.dashboard-header {
			flex-direction: column;
			gap: 16px;
		}
		
		.dashboard-header h1 {
			font-size: 24px;
		}
		
		.stats-grid {
			grid-template-columns: 1fr;
		}
		
		.table-container {
			overflow-x: scroll;
		}
		
		.modal {
			max-width: 100%;
			max-height: 100vh;
			border-radius: 0;
		}
	}
</style>