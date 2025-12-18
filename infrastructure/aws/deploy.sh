#!/bin/bash

# HRMS CloudFormation Deployment Script
# This script automates the deployment of the HRMS application to AWS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
STACK_NAME="hrms-prod"
REGION="us-east-1"
DB_USERNAME="postgres"
DB_PASSWORD="postgresql123!"
DB_NAME="hrmsdb"
ENVIRONMENT="production"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install it first."
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install it first."
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        print_error "AWS credentials are not configured. Run 'aws configure' first."
        exit 1
    fi
    
    print_success "All prerequisites met"
}

# Function to get AWS account ID
get_account_id() {
    aws sts get-caller-identity --query Account --output text
}

# Function to create ECR repositories if they don't exist
create_ecr_repos() {
    print_info "Creating ECR repositories..."
    
    ACCOUNT_ID=$(get_account_id)
    
    # Create backend repository
    if ! aws ecr describe-repositories --repository-names "${STACK_NAME}-backend" --region "$REGION" &> /dev/null; then
        aws ecr create-repository \
            --repository-name "${STACK_NAME}-backend" \
            --region "$REGION" \
            --image-scanning-configuration scanOnPush=true
        print_success "Created backend ECR repository"
    else
        print_info "Backend ECR repository already exists"
    fi
    
    # Create frontend repository
    if ! aws ecr describe-repositories --repository-names "${STACK_NAME}-frontend" --region "$REGION" &> /dev/null; then
        aws ecr create-repository \
            --repository-name "${STACK_NAME}-frontend" \
            --region "$REGION" \
            --image-scanning-configuration scanOnPush=true
        print_success "Created frontend ECR repository"
    else
        print_info "Frontend ECR repository already exists"
    fi
}

# Function to build and push Docker images
build_and_push_images() {
    print_info "Building and pushing Docker images..."
    
    ACCOUNT_ID=$(get_account_id)
    
    # Login to ECR
    print_info "Logging in to ECR..."
    aws ecr get-login-password --region "$REGION" | \
        docker login --username AWS --password-stdin "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com"
    
    # Build and push backend
    print_info "Building backend image..."
    cd backend
    docker build -t "${STACK_NAME}-backend:latest" .
    docker tag "${STACK_NAME}-backend:latest" \
        "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${STACK_NAME}-backend:latest"
    
    print_info "Pushing backend image..."
    docker push "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${STACK_NAME}-backend:latest"
    print_success "Backend image pushed successfully"
    
    # Build and push frontend
    print_info "Building frontend image..."
    cd ../frontend
    docker build -t "${STACK_NAME}-frontend:latest" .
    docker tag "${STACK_NAME}-frontend:latest" \
        "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${STACK_NAME}-frontend:latest"
    
    print_info "Pushing frontend image..."
    docker push "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${STACK_NAME}-frontend:latest"
    print_success "Frontend image pushed successfully"
    
    cd ..
}

# Function to generate JWT secret
generate_jwt_secret() {
    openssl rand -base64 32 | tr -d '\n'
}

# Function to deploy CloudFormation stack
deploy_stack() {
    print_info "Deploying CloudFormation stack..."
    
    # Generate JWT secret if not set
    if [ -z "$JWT_SECRET" ]; then
        JWT_SECRET=$(generate_jwt_secret)
        print_info "Generated JWT secret"
    fi
    
    # Check if stack exists
    if aws cloudformation describe-stacks --stack-name "$STACK_NAME" --region "$REGION" &> /dev/null; then
        print_warning "Stack already exists. Updating..."
        OPERATION="update-stack"
    else
        print_info "Creating new stack..."
        OPERATION="create-stack"
    fi
    
    # Deploy stack
    aws cloudformation "$OPERATION" \
        --stack-name "$STACK_NAME" \
        --template-body file://cloudformation-stack.yaml \
        --parameters \
            ParameterKey=Environment,ParameterValue="$ENVIRONMENT" \
            ParameterKey=DBUsername,ParameterValue="$DB_USERNAME" \
            ParameterKey=DBPassword,ParameterValue="$DB_PASSWORD" \
            ParameterKey=DBName,ParameterValue="$DB_NAME" \
            ParameterKey=JWTSecret,ParameterValue="$JWT_SECRET" \
        --capabilities CAPABILITY_NAMED_IAM \
        --region "$REGION"
    
    print_success "CloudFormation stack deployment initiated"
}

# Function to wait for stack completion
wait_for_stack() {
    print_info "Waiting for stack to complete..."
    
    if aws cloudformation wait stack-create-complete --stack-name "$STACK_NAME" --region "$REGION" 2>/dev/null; then
        print_success "Stack created successfully"
    elif aws cloudformation wait stack-update-complete --stack-name "$STACK_NAME" --region "$REGION" 2>/dev/null; then
        print_success "Stack updated successfully"
    else
        print_error "Stack operation failed or timed out"
        print_info "Check CloudFormation console for details"
        exit 1
    fi
}

# Function to get stack outputs
get_stack_outputs() {
    print_info "Retrieving stack outputs..."
    
    OUTPUTS=$(aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --region "$REGION" \
        --query 'Stacks[0].Outputs[*].[OutputKey,OutputValue]' \
        --output text)
    
    echo ""
    echo "======================================"
    echo "Stack Outputs"
    echo "======================================"
    echo "$OUTPUTS"
    echo ""
    
    # Get and display the application URL
    APP_URL=$(aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --region "$REGION" \
        --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' \
        --output text)
    
    print_success "Application URL: $APP_URL"
}

# Function to display next steps
display_next_steps() {
    echo ""
    echo "======================================"
    echo "Deployment Complete!"
    echo "======================================"
    echo ""
    echo "Next steps:"
    echo "1. Wait a few minutes for ECS services to start"
    echo "2. Access your application at the URL above"
    echo "3. Initialize the database with migrations"
    echo "4. Create an admin user using scripts in backend/scripts/"
    echo ""
    echo "For database initialization, see: DEPLOYMENT_GUIDE.md"
    echo ""
}

# Main deployment flow
main() {
    print_info "Starting HRMS deployment..."
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --stack-name)
                STACK_NAME="$2"
                shift 2
                ;;
            --region)
                REGION="$2"
                shift 2
                ;;
            --skip-images)
                SKIP_IMAGES=true
                shift
                ;;
            --jwt-secret)
                JWT_SECRET="$2"
                shift 2
                ;;
            --help)
                echo "Usage: ./deploy.sh [options]"
                echo ""
                echo "Options:"
                echo "  --stack-name NAME    CloudFormation stack name (default: hrms-prod)"
                echo "  --region REGION      AWS region (default: us-east-1)"
                echo "  --skip-images        Skip building and pushing Docker images"
                echo "  --jwt-secret SECRET  Use specific JWT secret"
                echo "  --help               Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    check_prerequisites
    
    if [ "$SKIP_IMAGES" != true ]; then
        create_ecr_repos
        build_and_push_images
    else
        print_warning "Skipping image build and push"
    fi
    
    deploy_stack
    wait_for_stack
    get_stack_outputs
    display_next_steps
    
    print_success "Deployment completed successfully!"
}

# Run main function
main "$@"
