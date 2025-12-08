#!/bin/bash

# Build and push Docker images to AWS ECR

set -e

ENVIRONMENT=${1:-dev}
REGION=${AWS_REGION:-us-east-1}
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "=== Building and Pushing Docker Images ==="
echo "Environment: $ENVIRONMENT"
echo "Region: $REGION"
echo "Account: $AWS_ACCOUNT_ID"
echo ""

# Login to ECR
echo "Logging in to ECR..."
aws ecr get-login-password --region "$REGION" | \
    docker login --username AWS --password-stdin "${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com"

# Create ECR repositories if they don't exist
create_repo() {
    local repo_name=$1
    echo "Creating ECR repository: $repo_name"
    aws ecr create-repository \
        --repository-name "$repo_name" \
        --region "$REGION" \
        --image-scanning-configuration scanOnPush=true \
        --encryption-configuration encryptionType=AES256 \
        2>/dev/null || echo "Repository $repo_name already exists"
}

create_repo "${ENVIRONMENT}-hr-backend"
create_repo "${ENVIRONMENT}-hr-frontend"

# Build and push backend
echo ""
echo "Building backend image..."
cd ../backend
docker build -t "${ENVIRONMENT}-hr-backend" .
docker tag "${ENVIRONMENT}-hr-backend:latest" \
    "${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${ENVIRONMENT}-hr-backend:latest"

echo "Pushing backend image..."
docker push "${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${ENVIRONMENT}-hr-backend:latest"

# Build and push frontend
echo ""
echo "Building frontend image..."
cd ../frontend
docker build -t "${ENVIRONMENT}-hr-frontend" .
docker tag "${ENVIRONMENT}-hr-frontend:latest" \
    "${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${ENVIRONMENT}-hr-frontend:latest"

echo "Pushing frontend image..."
docker push "${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${ENVIRONMENT}-hr-frontend:latest"

cd ../infrastructure

echo ""
echo "=== Docker Images Pushed Successfully ==="
echo "Backend: ${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${ENVIRONMENT}-hr-backend:latest"
echo "Frontend: ${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${ENVIRONMENT}-hr-frontend:latest"
