#!/bin/bash

# HR Application AWS Deployment Script

set -e

ENVIRONMENT=${1:-dev}
REGION=${AWS_REGION:-us-east-1}

echo "=== Deploying HR Application to AWS ==="
echo "Environment: $ENVIRONMENT"
echo "Region: $REGION"
echo ""

# Function to wait for stack completion
wait_for_stack() {
    local stack_name=$1
    echo "Waiting for stack $stack_name to complete..."
    aws cloudformation wait stack-create-complete \
        --stack-name "$stack_name" \
        --region "$REGION" 2>/dev/null || \
    aws cloudformation wait stack-update-complete \
        --stack-name "$stack_name" \
        --region "$REGION" 2>/dev/null
    echo "Stack $stack_name completed successfully"
}

# Deploy network infrastructure
echo "Deploying network infrastructure..."
aws cloudformation deploy \
    --template-file network.yaml \
    --stack-name "${ENVIRONMENT}-hr-network" \
    --parameter-overrides Environment="$ENVIRONMENT" \
    --region "$REGION" \
    --no-fail-on-empty-changeset

wait_for_stack "${ENVIRONMENT}-hr-network"

# Deploy storage
echo "Deploying S3 storage..."
aws cloudformation deploy \
    --template-file storage.yaml \
    --stack-name "${ENVIRONMENT}-hr-storage" \
    --parameter-overrides Environment="$ENVIRONMENT" \
    --capabilities CAPABILITY_IAM \
    --region "$REGION" \
    --no-fail-on-empty-changeset

wait_for_stack "${ENVIRONMENT}-hr-storage"

# Deploy database
echo "Deploying database..."
read -sp "Enter database password: " DB_PASSWORD
echo ""

aws cloudformation deploy \
    --template-file database.yaml \
    --stack-name "${ENVIRONMENT}-hr-database" \
    --parameter-overrides \
        Environment="$ENVIRONMENT" \
        DBPassword="$DB_PASSWORD" \
    --region "$REGION" \
    --no-fail-on-empty-changeset

wait_for_stack "${ENVIRONMENT}-hr-database"

# Get outputs
DB_ENDPOINT=$(aws cloudformation describe-stacks \
    --stack-name "${ENVIRONMENT}-hr-database" \
    --query 'Stacks[0].Outputs[?OutputKey==`DBEndpoint`].OutputValue' \
    --output text \
    --region "$REGION")

S3_BUCKET=$(aws cloudformation describe-stacks \
    --stack-name "${ENVIRONMENT}-hr-storage" \
    --query 'Stacks[0].Outputs[?OutputKey==`DocumentsBucketName`].OutputValue' \
    --output text \
    --region "$REGION")

echo ""
echo "=== Deployment Complete ==="
echo "Database Endpoint: $DB_ENDPOINT"
echo "S3 Bucket: $S3_BUCKET"
echo ""
echo "Next steps:"
echo "1. Build and push Docker images to ECR"
echo "2. Deploy ECS services"
echo "3. Run database migrations"
