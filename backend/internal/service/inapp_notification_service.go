package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// DynamoDBInAppNotificationService implements InAppNotificationService using DynamoDB
type DynamoDBInAppNotificationService struct {
	client    *dynamodb.Client
	tableName string
}

// notificationItem represents a DynamoDB item for notifications
type notificationItem struct {
	PK        string                 `dynamodbav:"PK"`        // NOTIFICATION#{id}
	SK        string                 `dynamodbav:"SK"`        // METADATA
	GSI1PK    string                 `dynamodbav:"GSI1PK"`    // USER#{user_id}
	GSI1SK    string                 `dynamodbav:"GSI1SK"`    // CREATED#{timestamp}
	GSI2PK    string                 `dynamodbav:"GSI2PK"`    // USER#{user_id}#UNREAD
	GSI2SK    string                 `dynamodbav:"GSI2SK"`    // CREATED#{timestamp}
	Type      string                 `dynamodbav:"EntityType"` // NOTIFICATION
	ID        string                 `dynamodbav:"ID"`
	UserID    string                 `dynamodbav:"UserID"`
	NotifType string                 `dynamodbav:"NotificationType"`
	Title     string                 `dynamodbav:"Title"`
	Message   string                 `dynamodbav:"Message"`
	Data      map[string]interface{} `dynamodbav:"Data"`
	Read      bool                   `dynamodbav:"Read"`
	ActionURL string                 `dynamodbav:"ActionURL,omitempty"`
	CreatedAt string                 `dynamodbav:"CreatedAt"`
	ExpiresAt string                 `dynamodbav:"ExpiresAt,omitempty"`
	TTL       int64                  `dynamodbav:"TTL,omitempty"` // DynamoDB TTL
}

// NewDynamoDBInAppNotificationService creates a new DynamoDB notification service
func NewDynamoDBInAppNotificationService(client *dynamodb.Client, tableName string) *DynamoDBInAppNotificationService {
	return &DynamoDBInAppNotificationService{
		client:    client,
		tableName: tableName,
	}
}

// CreateNotification creates a single notification
func (s *DynamoDBInAppNotificationService) CreateNotification(
	ctx context.Context,
	notification *InAppNotification,
) error {
	log.Printf("Creating notification for user: %s", notification.UserID)

	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	item := s.notificationToItem(notification)

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      av,
	})

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	log.Printf("Notification created successfully: %s", notification.ID)
	return nil
}

// CreateBulkNotifications creates multiple notifications
func (s *DynamoDBInAppNotificationService) CreateBulkNotifications(
	ctx context.Context,
	notifications []*InAppNotification,
) error {
	log.Printf("Creating %d bulk notifications", len(notifications))

	if len(notifications) == 0 {
		return nil
	}

	// DynamoDB BatchWriteItem supports up to 25 items
	// Split into batches if needed
	batchSize := 25
	for i := 0; i < len(notifications); i += batchSize {
		end := i + batchSize
		if end > len(notifications) {
			end = len(notifications)
		}

		batch := notifications[i:end]
		if err := s.writeBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to write batch %d: %w", i/batchSize, err)
		}
	}

	log.Printf("All bulk notifications created successfully")
	return nil
}

// GetUserNotifications retrieves notifications for a user
func (s *DynamoDBInAppNotificationService) GetUserNotifications(
	ctx context.Context,
	userID string,
	limit int,
	unreadOnly bool,
) ([]*InAppNotification, error) {
	log.Printf("Getting notifications for user: %s (limit: %d, unreadOnly: %v)", userID, limit, unreadOnly)

	var indexName string
	var keyCondition string
	var expressionValues map[string]types.AttributeValue

	if unreadOnly {
		// Use GSI2 for unread notifications
		indexName = "GSI2"
		keyCondition = "GSI2PK = :gsi2pk"
		expressionValues = map[string]types.AttributeValue{
			":gsi2pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s#UNREAD", userID)},
		}
	} else {
		// Use GSI1 for all notifications
		indexName = "GSI1"
		keyCondition = "GSI1PK = :gsi1pk"
		expressionValues = map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", userID)},
		}
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(s.tableName),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: expressionValues,
		ScanIndexForward:          aws.Bool(false), // Most recent first
		Limit:                     aws.Int32(int32(limit)),
	}

	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}

	notifications := make([]*InAppNotification, 0, len(result.Items))
	for _, item := range result.Items {
		var notifItem notificationItem
		if err := attributevalue.UnmarshalMap(item, &notifItem); err != nil {
			log.Printf("WARNING: Failed to unmarshal notification: %v", err)
			continue
		}

		notification := s.itemToNotification(&notifItem)
		notifications = append(notifications, notification)
	}

	log.Printf("Retrieved %d notifications for user: %s", len(notifications), userID)
	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *DynamoDBInAppNotificationService) MarkAsRead(
	ctx context.Context,
	notificationID string,
) error {
	log.Printf("Marking notification as read: %s", notificationID)

	// First, get the notification to get the user ID
	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("NOTIFICATION#%s", notificationID)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	}

	getResult, err := s.client.GetItem(ctx, getInput)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if getResult.Item == nil {
		return fmt.Errorf("notification not found: %s", notificationID)
	}

	var notifItem notificationItem
	if err := attributevalue.UnmarshalMap(getResult.Item, &notifItem); err != nil {
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	// Update the notification
	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("NOTIFICATION#%s", notificationID)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
		UpdateExpression: aws.String("SET #read = :true, GSI2PK = :gsi2pk_read"),
		ExpressionAttributeNames: map[string]string{
			"#read": "Read",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":true":        &types.AttributeValueMemberBOOL{Value: true},
			":gsi2pk_read": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s#READ", notifItem.UserID)},
		},
	}

	_, err = s.client.UpdateItem(ctx, updateInput)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	log.Printf("Notification marked as read: %s", notificationID)
	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (s *DynamoDBInAppNotificationService) MarkAllAsRead(
	ctx context.Context,
	userID string,
) error {
	log.Printf("Marking all notifications as read for user: %s", userID)

	// Get all unread notifications
	notifications, err := s.GetUserNotifications(ctx, userID, 100, true)
	if err != nil {
		return fmt.Errorf("failed to get unread notifications: %w", err)
	}

	// Mark each as read
	for _, notification := range notifications {
		if err := s.MarkAsRead(ctx, notification.ID); err != nil {
			log.Printf("WARNING: Failed to mark notification as read: %v", err)
		}
	}

	log.Printf("Marked %d notifications as read for user: %s", len(notifications), userID)
	return nil
}

// DeleteNotification deletes a notification
func (s *DynamoDBInAppNotificationService) DeleteNotification(
	ctx context.Context,
	notificationID string,
) error {
	log.Printf("Deleting notification: %s", notificationID)

	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("NOTIFICATION#%s", notificationID)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	log.Printf("Notification deleted successfully: %s", notificationID)
	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *DynamoDBInAppNotificationService) GetUnreadCount(
	ctx context.Context,
	userID string,
) (int, error) {
	log.Printf("Getting unread count for user: %s", userID)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(s.tableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :gsi2pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi2pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s#UNREAD", userID)},
		},
		Select: types.SelectCount,
	}

	result, err := s.client.Query(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	count := int(result.Count)
	log.Printf("User %s has %d unread notifications", userID, count)
	return count, nil
}

// Helper methods

func (s *DynamoDBInAppNotificationService) writeBatch(
	ctx context.Context,
	notifications []*InAppNotification,
) error {
	writeRequests := make([]types.WriteRequest, 0, len(notifications))

	for _, notification := range notifications {
		if notification.ID == "" {
			notification.ID = uuid.New().String()
		}

		item := s.notificationToItem(notification)
		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			return fmt.Errorf("failed to marshal notification: %w", err)
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: av,
			},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			s.tableName: writeRequests,
		},
	}

	_, err := s.client.BatchWriteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to batch write notifications: %w", err)
	}

	return nil
}

func (s *DynamoDBInAppNotificationService) notificationToItem(n *InAppNotification) *notificationItem {
	createdAt := n.CreatedAt.Format(time.RFC3339)
	
	var expiresAt string
	var ttl int64
	if n.ExpiresAt != nil {
		expiresAt = n.ExpiresAt.Format(time.RFC3339)
		ttl = n.ExpiresAt.Unix()
	}

	// GSI2PK changes based on read status for efficient unread queries
	gsi2pk := fmt.Sprintf("USER#%s#UNREAD", n.UserID)
	if n.Read {
		gsi2pk = fmt.Sprintf("USER#%s#READ", n.UserID)
	}

	return &notificationItem{
		PK:        fmt.Sprintf("NOTIFICATION#%s", n.ID),
		SK:        "METADATA",
		GSI1PK:    fmt.Sprintf("USER#%s", n.UserID),
		GSI1SK:    fmt.Sprintf("CREATED#%s", createdAt),
		GSI2PK:    gsi2pk,
		GSI2SK:    fmt.Sprintf("CREATED#%s", createdAt),
		Type:      "NOTIFICATION",
		ID:        n.ID,
		UserID:    n.UserID,
		NotifType: n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Data:      n.Data,
		Read:      n.Read,
		ActionURL: n.ActionURL,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		TTL:       ttl,
	}
}

func (s *DynamoDBInAppNotificationService) itemToNotification(item *notificationItem) *InAppNotification {
	createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
	
	var expiresAt *time.Time
	if item.ExpiresAt != "" {
		parsed, _ := time.Parse(time.RFC3339, item.ExpiresAt)
		expiresAt = &parsed
	}

	return &InAppNotification{
		ID:        item.ID,
		UserID:    item.UserID,
		Type:      item.NotifType,
		Title:     item.Title,
		Message:   item.Message,
		Data:      item.Data,
		Read:      item.Read,
		ActionURL: item.ActionURL,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}
}