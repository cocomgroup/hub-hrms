package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"hub-hrms/backend/internal/models"
)

// BackgroundCheckDynamoRepository implements BackgroundCheckRepository using DynamoDB
type BackgroundCheckDynamoRepository struct {
	client    *dynamodb.Client
	tableName string
}

// DynamoDB item structure
type backgroundCheckItem struct {
	PK          string    `dynamodbav:"PK"`          // BGCHECK#{id}
	SK          string    `dynamodbav:"SK"`          // METADATA
	GSI1PK      string    `dynamodbav:"GSI1PK"`      // EMPLOYEE#{employee_id}
	GSI1SK      string    `dynamodbav:"GSI1SK"`      // BGCHECK#{initiated_at}
	EntityType  string    `dynamodbav:"EntityType"`  // BACKGROUND_CHECK
	ID          string    `dynamodbav:"ID"`
	EmployeeID  string    `dynamodbav:"EmployeeID"`
	PackageID   string    `dynamodbav:"PackageID"`
	ProviderID  string    `dynamodbav:"ProviderID"`
	Status      string    `dynamodbav:"Status"`
	Result      string    `dynamodbav:"Result,omitempty"`
	CheckTypes  []string  `dynamodbav:"CheckTypes"`
	CandidateInfo string  `dynamodbav:"CandidateInfo"` // JSON
	ReportURL   string    `dynamodbav:"ReportURL,omitempty"`
	ProviderData string   `dynamodbav:"ProviderData,omitempty"` // JSON
	InitiatedBy string    `dynamodbav:"InitiatedBy"`
	InitiatedAt string    `dynamodbav:"InitiatedAt"`
	CompletedAt string    `dynamodbav:"CompletedAt,omitempty"`
	EstimatedETA string   `dynamodbav:"EstimatedETA,omitempty"`
	Notes       string    `dynamodbav:"Notes,omitempty"`
	ComplianceData string `dynamodbav:"ComplianceData"` // JSON
	CreatedAt   string    `dynamodbav:"CreatedAt"`
	UpdatedAt   string    `dynamodbav:"UpdatedAt"`
}

type packageItem struct {
	PK             string   `dynamodbav:"PK"`  // PACKAGE#{id}
	SK             string   `dynamodbav:"SK"`  // METADATA
	EntityType     string   `dynamodbav:"EntityType"`
	ID             string   `dynamodbav:"ID"`
	Name           string   `dynamodbav:"Name"`
	Description    string   `dynamodbav:"Description"`
	CheckTypes     []string `dynamodbav:"CheckTypes"`
	ProviderID     string   `dynamodbav:"ProviderID"`
	TurnaroundDays int      `dynamodbav:"TurnaroundDays"`
	Cost           float64  `dynamodbav:"Cost"`
	Active         bool     `dynamodbav:"Active"`
	CreatedAt      string   `dynamodbav:"CreatedAt"`
	UpdatedAt      string   `dynamodbav:"UpdatedAt"`
}

// NewBackgroundCheckDynamoRepository creates a new DynamoDB-backed repository
func NewBackgroundCheckDynamoRepository(client *dynamodb.Client, tableName string) *BackgroundCheckDynamoRepository {
	return &BackgroundCheckDynamoRepository{
		client:    client,
		tableName: tableName,
	}
}

// Create stores a new background check
func (r *BackgroundCheckDynamoRepository) Create(ctx context.Context, check *models.BackgroundCheck) error {
	now := time.Now().Format(time.RFC3339)

	candidateJSON, err := json.Marshal(check.CandidateInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal candidate info: %w", err)
	}

	complianceJSON, err := json.Marshal(check.ComplianceData)
	if err != nil {
		return fmt.Errorf("failed to marshal compliance data: %w", err)
	}

	checkTypes := make([]string, len(check.CheckTypes))
	for i, ct := range check.CheckTypes {
		checkTypes[i] = string(ct)
	}

	item := backgroundCheckItem{
		PK:            fmt.Sprintf("BGCHECK#%s", check.ID),
		SK:            "METADATA",
		GSI1PK:        fmt.Sprintf("EMPLOYEE#%s", check.EmployeeID),
		GSI1SK:        fmt.Sprintf("BGCHECK#%s", check.InitiatedAt.Format(time.RFC3339)),
		EntityType:    "BACKGROUND_CHECK",
		ID:            check.ID,
		EmployeeID:    check.EmployeeID,
		PackageID:     check.PackageID,
		ProviderID:    check.ProviderID,
		Status:        string(check.Status),
		Result:        string(check.Result),
		CheckTypes:    checkTypes,
		CandidateInfo: string(candidateJSON),
		ReportURL:     check.ReportURL,
		ProviderData:  string(check.ProviderData),
		InitiatedBy:   check.InitiatedBy,
		InitiatedAt:   check.InitiatedAt.Format(time.RFC3339),
		Notes:         check.Notes,
		ComplianceData: string(complianceJSON),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if check.CompletedAt != nil {
		item.CompletedAt = check.CompletedAt.Format(time.RFC3339)
	}

	if check.EstimatedETA != nil {
		item.EstimatedETA = check.EstimatedETA.Format(time.RFC3339)
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})

	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// GetByID retrieves a background check by ID
func (r *BackgroundCheckDynamoRepository) GetByID(ctx context.Context, id string) (*models.BackgroundCheck, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("BGCHECK#%s", id)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("background check not found")
	}

	var item backgroundCheckItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return r.itemToBackgroundCheck(&item)
}

// GetByEmployeeID retrieves all background checks for an employee
func (r *BackgroundCheckDynamoRepository) GetByEmployeeID(ctx context.Context, employeeID string) ([]*models.BackgroundCheck, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMPLOYEE#%s", employeeID)},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	checks := make([]*models.BackgroundCheck, 0, len(result.Items))
	for _, itemMap := range result.Items {
		var item backgroundCheckItem
		if err := attributevalue.UnmarshalMap(itemMap, &item); err != nil {
			continue
		}

		check, err := r.itemToBackgroundCheck(&item)
		if err != nil {
			continue
		}

		checks = append(checks, check)
	}

	return checks, nil
}

// Update updates an existing background check
func (r *BackgroundCheckDynamoRepository) Update(ctx context.Context, check *models.BackgroundCheck) error {
	now := time.Now().Format(time.RFC3339)

	candidateJSON, err := json.Marshal(check.CandidateInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal candidate info: %w", err)
	}

	complianceJSON, err := json.Marshal(check.ComplianceData)
	if err != nil {
		return fmt.Errorf("failed to marshal compliance data: %w", err)
	}

	updateExpr := "SET #status = :status, #result = :result, #reportUrl = :reportUrl, " +
		"#providerData = :providerData, #notes = :notes, #candidateInfo = :candidateInfo, " +
		"#complianceData = :complianceData, #updatedAt = :updatedAt"

	exprAttrNames := map[string]string{
		"#status":         "Status",
		"#result":         "Result",
		"#reportUrl":      "ReportURL",
		"#providerData":   "ProviderData",
		"#notes":          "Notes",
		"#candidateInfo":  "CandidateInfo",
		"#complianceData": "ComplianceData",
		"#updatedAt":      "UpdatedAt",
	}

	exprAttrValues := map[string]types.AttributeValue{
		":status":         &types.AttributeValueMemberS{Value: string(check.Status)},
		":result":         &types.AttributeValueMemberS{Value: string(check.Result)},
		":reportUrl":      &types.AttributeValueMemberS{Value: check.ReportURL},
		":providerData":   &types.AttributeValueMemberS{Value: string(check.ProviderData)},
		":notes":          &types.AttributeValueMemberS{Value: check.Notes},
		":candidateInfo":  &types.AttributeValueMemberS{Value: string(candidateJSON)},
		":complianceData": &types.AttributeValueMemberS{Value: string(complianceJSON)},
		":updatedAt":      &types.AttributeValueMemberS{Value: now},
	}

	if check.CompletedAt != nil {
		updateExpr += ", #completedAt = :completedAt"
		exprAttrNames["#completedAt"] = "CompletedAt"
		exprAttrValues[":completedAt"] = &types.AttributeValueMemberS{
			Value: check.CompletedAt.Format(time.RFC3339),
		}
	}

	if check.EstimatedETA != nil {
		updateExpr += ", #estimatedETA = :estimatedETA"
		exprAttrNames["#estimatedETA"] = "EstimatedETA"
		exprAttrValues[":estimatedETA"] = &types.AttributeValueMemberS{
			Value: check.EstimatedETA.Format(time.RFC3339),
		}
	}

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("BGCHECK#%s", check.ID)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
	})

	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}

// List retrieves background checks with filters
func (r *BackgroundCheckDynamoRepository) List(
	ctx context.Context,
	filters map[string]interface{},
) ([]*models.BackgroundCheck, error) {
	// Implementation would depend on your specific filter requirements
	// For now, return all checks (with pagination in production)
	
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("EntityType = :entityType"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":entityType": &types.AttributeValueMemberS{Value: "BACKGROUND_CHECK"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan items: %w", err)
	}

	checks := make([]*models.BackgroundCheck, 0, len(result.Items))
	for _, itemMap := range result.Items {
		var item backgroundCheckItem
		if err := attributevalue.UnmarshalMap(itemMap, &item); err != nil {
			continue
		}

		check, err := r.itemToBackgroundCheck(&item)
		if err != nil {
			continue
		}

		checks = append(checks, check)
	}

	return checks, nil
}

// CreatePackage stores a new background check package
func (r *BackgroundCheckDynamoRepository) CreatePackage(ctx context.Context, pkg *models.BackgroundCheckPackage) error {
	now := time.Now().Format(time.RFC3339)

	checkTypes := make([]string, len(pkg.CheckTypes))
	for i, ct := range pkg.CheckTypes {
		checkTypes[i] = string(ct)
	}

	item := packageItem{
		PK:             fmt.Sprintf("PACKAGE#%s", pkg.ID),
		SK:             "METADATA",
		EntityType:     "BGCHECK_PACKAGE",
		ID:             pkg.ID,
		Name:           pkg.Name,
		Description:    pkg.Description,
		CheckTypes:     checkTypes,
		ProviderID:     pkg.ProviderID,
		TurnaroundDays: pkg.TurnaroundDays,
		Cost:           pkg.Cost,
		Active:         pkg.Active,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal package: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})

	if err != nil {
		return fmt.Errorf("failed to put package: %w", err)
	}

	return nil
}

// GetPackage retrieves a package by ID
func (r *BackgroundCheckDynamoRepository) GetPackage(ctx context.Context, id string) (*models.BackgroundCheckPackage, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("PACKAGE#%s", id)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("package not found")
	}

	var item packageItem
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	return r.itemToPackage(&item), nil
}

// ListPackages retrieves all packages
func (r *BackgroundCheckDynamoRepository) ListPackages(ctx context.Context) ([]*models.BackgroundCheckPackage, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("EntityType = :entityType"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":entityType": &types.AttributeValueMemberS{Value: "BGCHECK_PACKAGE"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan packages: %w", err)
	}

	packages := make([]*models.BackgroundCheckPackage, 0, len(result.Items))
	for _, itemMap := range result.Items {
		var item packageItem
		if err := attributevalue.UnmarshalMap(itemMap, &item); err != nil {
			continue
		}

		packages = append(packages, r.itemToPackage(&item))
	}

	return packages, nil
}

// Helper functions

func (r *BackgroundCheckDynamoRepository) itemToBackgroundCheck(item *backgroundCheckItem) (*models.BackgroundCheck, error) {
	var candidateInfo models.CandidateInfo
	if err := json.Unmarshal([]byte(item.CandidateInfo), &candidateInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal candidate info: %w", err)
	}

	var complianceData models.ComplianceData
	if err := json.Unmarshal([]byte(item.ComplianceData), &complianceData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal compliance data: %w", err)
	}

	checkTypes := make([]models.BackgroundCheckType, len(item.CheckTypes))
	for i, ct := range item.CheckTypes {
		checkTypes[i] = models.BackgroundCheckType(ct)
	}

	initiatedAt, _ := time.Parse(time.RFC3339, item.InitiatedAt)

	check := &models.BackgroundCheck{
		ID:             item.ID,
		EmployeeID:     item.EmployeeID,
		PackageID:      item.PackageID,
		ProviderID:     item.ProviderID,
		Status:         models.BackgroundCheckStatus(item.Status),
		Result:         models.BackgroundCheckResult(item.Result),
		CheckTypes:     checkTypes,
		CandidateInfo:  candidateInfo,
		ReportURL:      item.ReportURL,
		ProviderData:   []byte(item.ProviderData),
		InitiatedBy:    item.InitiatedBy,
		InitiatedAt:    initiatedAt,
		Notes:          item.Notes,
		ComplianceData: complianceData,
	}

	if item.CompletedAt != "" {
		completedAt, _ := time.Parse(time.RFC3339, item.CompletedAt)
		check.CompletedAt = &completedAt
	}

	if item.EstimatedETA != "" {
		estimatedETA, _ := time.Parse(time.RFC3339, item.EstimatedETA)
		check.EstimatedETA = &estimatedETA
	}

	return check, nil
}

func (r *BackgroundCheckDynamoRepository) itemToPackage(item *packageItem) *models.BackgroundCheckPackage {
	checkTypes := make([]models.BackgroundCheckType, len(item.CheckTypes))
	for i, ct := range item.CheckTypes {
		checkTypes[i] = models.BackgroundCheckType(ct)
	}

	return &models.BackgroundCheckPackage{
		ID:             item.ID,
		Name:           item.Name,
		Description:    item.Description,
		CheckTypes:     checkTypes,
		ProviderID:     item.ProviderID,
		TurnaroundDays: item.TurnaroundDays,
		Cost:           item.Cost,
		Active:         item.Active,
	}
}
