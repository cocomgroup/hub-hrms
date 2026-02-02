package graphql

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	graphql1 "hub-hrms/backend/internal/graphql"
)

type Resolver struct{}

// Job is the resolver for the job field.
func (r *applicationResolver) Job(ctx context.Context, obj *graphql1.Application) (*graphql1.Job, error) {
	panic("not implemented")
}

// Candidate is the resolver for the candidate field.
func (r *applicationResolver) Candidate(ctx context.Context, obj *graphql1.Application) (*graphql1.Candidate, error) {
	panic("not implemented")
}

// SubmitApplication is the resolver for the submitApplication field.
func (r *mutationResolver) SubmitApplication(ctx context.Context, input graphql1.ApplicationInput) (*graphql1.Application, error) {
	panic("not implemented")
}

// IncrementJobView is the resolver for the incrementJobView field.
func (r *mutationResolver) IncrementJobView(ctx context.Context, id string) (*graphql1.Job, error) {
	panic("not implemented")
}

// UpdateApplicationStatus is the resolver for the updateApplicationStatus field.
func (r *mutationResolver) UpdateApplicationStatus(ctx context.Context, id string, status graphql1.ApplicationStatus, note *string) (*graphql1.Application, error) {
	panic("not implemented")
}

// BulkUpdateApplicationStatus is the resolver for the bulkUpdateApplicationStatus field.
func (r *mutationResolver) BulkUpdateApplicationStatus(ctx context.Context, ids []string, status graphql1.ApplicationStatus) ([]*graphql1.Application, error) {
	panic("not implemented")
}

// AddApplicationNote is the resolver for the addApplicationNote field.
func (r *mutationResolver) AddApplicationNote(ctx context.Context, applicationID string, content string, isInternal *bool) (*graphql1.ApplicationNote, error) {
	panic("not implemented")
}

// ScoreApplication is the resolver for the scoreApplication field.
func (r *mutationResolver) ScoreApplication(ctx context.Context, applicationID string) (*graphql1.AIScore, error) {
	panic("not implemented")
}

// CreateJob is the resolver for the createJob field.
func (r *mutationResolver) CreateJob(ctx context.Context, input graphql1.JobInput) (*graphql1.Job, error) {
	panic("not implemented")
}

// UpdateJob is the resolver for the updateJob field.
func (r *mutationResolver) UpdateJob(ctx context.Context, id string, input graphql1.JobInput) (*graphql1.Job, error) {
	panic("not implemented")
}

// PublishJob is the resolver for the publishJob field.
func (r *mutationResolver) PublishJob(ctx context.Context, id string) (*graphql1.Job, error) {
	panic("not implemented")
}

// CloseJob is the resolver for the closeJob field.
func (r *mutationResolver) CloseJob(ctx context.Context, id string) (*graphql1.Job, error) {
	panic("not implemented")
}

// DeleteJob is the resolver for the deleteJob field.
func (r *mutationResolver) DeleteJob(ctx context.Context, id string) (bool, error) {
	panic("not implemented")
}

// GenerateJobDescription is the resolver for the generateJobDescription field.
func (r *mutationResolver) GenerateJobDescription(ctx context.Context, input graphql1.JobDescriptionInput) (*graphql1.JobDescriptionOutput, error) {
	panic("not implemented")
}

// UpdateCandidateProfile is the resolver for the updateCandidateProfile field.
func (r *mutationResolver) UpdateCandidateProfile(ctx context.Context, id string, input graphql1.CandidateProfileInput) (*graphql1.Candidate, error) {
	panic("not implemented")
}

// Jobs is the resolver for the jobs field.
func (r *queryResolver) Jobs(ctx context.Context, filters *graphql1.JobFilters, limit *int, offset *int) ([]*graphql1.Job, error) {
	panic("not implemented")
}

// Job is the resolver for the job field.
func (r *queryResolver) Job(ctx context.Context, id string) (*graphql1.Job, error) {
	panic("not implemented")
}

// Applications is the resolver for the applications field.
func (r *queryResolver) Applications(ctx context.Context, filters *graphql1.ApplicationFilters, limit *int, offset *int) ([]*graphql1.Application, error) {
	panic("not implemented")
}

// Application is the resolver for the application field.
func (r *queryResolver) Application(ctx context.Context, id string) (*graphql1.Application, error) {
	panic("not implemented")
}

// Candidate is the resolver for the candidate field.
func (r *queryResolver) Candidate(ctx context.Context, id string) (*graphql1.Candidate, error) {
	panic("not implemented")
}

// RecruitmentMetrics is the resolver for the recruitmentMetrics field.
func (r *queryResolver) RecruitmentMetrics(ctx context.Context, dateRange graphql1.DateRangeInput) (*graphql1.RecruitmentMetrics, error) {
	panic("not implemented")
}

// JobPerformance is the resolver for the jobPerformance field.
func (r *queryResolver) JobPerformance(ctx context.Context, jobID string) (*graphql1.JobPerformance, error) {
	panic("not implemented")
}

// ApplicationPipeline is the resolver for the applicationPipeline field.
func (r *queryResolver) ApplicationPipeline(ctx context.Context, jobID *string) ([]*graphql1.PipelineStage, error) {
	panic("not implemented")
}

// Application returns graphql1.ApplicationResolver implementation.
func (r *Resolver) Application() graphql1.ApplicationResolver { return &applicationResolver{r} }

// Mutation returns graphql1.MutationResolver implementation.
func (r *Resolver) Mutation() graphql1.MutationResolver { return &mutationResolver{r} }

// Query returns graphql1.QueryResolver implementation.
func (r *Resolver) Query() graphql1.QueryResolver { return &queryResolver{r} }

type applicationResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
*/
