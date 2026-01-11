package repository

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	User            UserRepository
	Employee        EmployeeRepository
	Onboarding      OnboardingRepository
	Workflow        WorkflowRepository
	Timesheet       TimesheetRepository
	PTO             PTORepository
	Benefits        BenefitsRepository
	Payroll         PayrollRepository
	Recruiting      RecruitingRepository
	Organization    OrganizationRepository
	Project         ProjectRepository
	Compensation    CompensationRepository
	BankInfo        BankInfoRepository
	BackgroundCheck *BackgroundCheckDynamoRepository
	Applicant		ApplicantRepository
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	// Initialize DynamoDB client for background checks
	dynamoClient, tableName := initializeDynamoDBClient()
	
	return &Repositories{
		User:            NewUserRepository(db),
		Employee:        NewEmployeeRepository(db),
		Onboarding:      NewOnboardingRepository(db),
		Workflow:        NewWorkflowRepository(db),
		Timesheet:       NewTimesheetRepository(db),
		PTO:             NewPTORepository(db),
		Benefits:        NewBenefitsRepository(db),
		Payroll:         NewPayrollRepository(db),
		Recruiting:      NewRecruitingRepository(db),
		Organization:    NewOrganizationRepository(db),
		Project:         NewProjectRepository(db),
		Compensation:    NewCompensationRepository(db),
		BankInfo:        NewBankInfoRepository(db),
		BackgroundCheck: NewBackgroundCheckDynamoRepository(dynamoClient, tableName),
		Applicant:       NewApplicantRepository(db),
	}
}

// initializeDynamoDBClient creates and configures the DynamoDB client
func initializeDynamoDBClient() (*dynamodb.Client, string) {
	ctx := context.Background()
	
	// Get table name from environment
	tableName := os.Getenv("DYNAMODB_BGCHECK_TABLE_NAME")
	if tableName == "" {
		tableName = "hrms-background-checks" // default table name
		log.Printf("DYNAMODB_BGCHECK_TABLE_NAME not set, using default: %s", tableName)
	}
	
	// Check if we should use local DynamoDB
	useLocalDynamoDB := os.Getenv("USE_LOCAL_DYNAMODB") == "true"
	
	var cfg aws.Config
	var err error
	
	if useLocalDynamoDB {
		// Local DynamoDB setup (for development)
		localEndpoint := os.Getenv("DYNAMODB_ENDPOINT")
		if localEndpoint == "" {
			localEndpoint = "http://localhost:8000"
		}
		
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion("us-east-1"), // Region doesn't matter for local
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					if service == dynamodb.ServiceID {
						return aws.Endpoint{
							URL:           localEndpoint,
							SigningRegion: region,
						}, nil
					}
					return aws.Endpoint{}, &aws.EndpointNotFoundError{}
				},
			)),
		)
		
		if err != nil {
			log.Fatalf("Failed to load local DynamoDB config: %v", err)
		}
		
		log.Printf("DynamoDB: Using local endpoint at %s (table: %s)", localEndpoint, tableName)
	} else {
		// Production AWS DynamoDB
		cfg, err = config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("Failed to load AWS config: %v", err)
		}
		
		region := cfg.Region
		if region == "" {
			region = "us-east-1" // fallback
		}
		
		log.Printf("DynamoDB: Using AWS DynamoDB in region %s (table: %s)", region, tableName)
	}
	
	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)
	
	// Optionally verify table exists (only in development)
	if os.Getenv("VERIFY_DYNAMODB_TABLE") == "true" {
		verifyTable(ctx, client, tableName)
	}
	
	return client, tableName
}

// verifyTable checks if the DynamoDB table exists
func verifyTable(ctx context.Context, client *dynamodb.Client, tableName string) {
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	
	if err != nil {
		log.Printf("WARNING: Could not verify DynamoDB table '%s': %v", tableName, err)
		log.Printf("The table will be created when first used, or you may need to create it manually.")
	} else {
		log.Printf("DynamoDB table '%s' verified successfully", tableName)
	}
}