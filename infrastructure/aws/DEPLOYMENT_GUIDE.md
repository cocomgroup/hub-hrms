# HRMS Application - CloudFormation Deployment Guide

## Overview
This CloudFormation stack deploys a complete HRMS application with:
- **Frontend**: Svelte/TypeScript application on ECS Fargate (Port 3000)
- **Backend**: Go API server on ECS Fargate (Port 8080)
- **Database**: RDS PostgreSQL 15.4 (Multi-AZ capable)
- **Storage**: S3 bucket for document management
- **Networking**: VPC with public/private subnets across 2 AZs
- **Load Balancing**: Application Load Balancer with health checks
- **Auto Scaling**: Automatic scaling for both frontend and backend services

## Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Docker** installed for building container images
3. **AWS Account** with permissions to create VPC, ECS, RDS, S3, ECR, IAM resources

## Quick Start Deployment

### Step 1: Build and Push Docker Images

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com

# Build and push backend
cd backend
docker build -t hrms-backend:latest .
docker tag hrms-backend:latest <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/<STACK_NAME>-backend:latest
docker push <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/<STACK_NAME>-backend:latest

# Build and push frontend
cd ../frontend
docker build -t hrms-frontend:latest .
docker tag hrms-frontend:latest <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/<STACK_NAME>-frontend:latest
docker push <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/<STACK_NAME>-frontend:latest
```

### Step 2: Deploy CloudFormation Stack

```bash
aws cloudformation create-stack \
  --stack-name hrms-prod \
  --template-body file://cloudformation-stack.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=production \
    ParameterKey=DBUsername,ParameterValue=postgres \
    ParameterKey=DBPassword,ParameterValue=postgresql123! \
    ParameterKey=DBName,ParameterValue=hrmsdb \
    ParameterKey=JWTSecret,ParameterValue=<YOUR_32_CHAR_RANDOM_SECRET> \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

### Step 3: Monitor Deployment

```bash
# Watch stack creation progress
aws cloudformation describe-stack-events \
  --stack-name hrms-prod \
  --region us-east-1 \
  --query 'StackEvents[*].[Timestamp,ResourceStatus,ResourceType,LogicalResourceId]' \
  --output table

# Get stack outputs once complete
aws cloudformation describe-stacks \
  --stack-name hrms-prod \
  --region us-east-1 \
  --query 'Stacks[0].Outputs'
```

## Alternative: Deploy with Pre-built Images

If you already have images in ECR or Docker Hub:

```bash
aws cloudformation create-stack \
  --stack-name hrms-prod \
  --template-body file://cloudformation-stack.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=production \
    ParameterKey=DBUsername,ParameterValue=postgres \
    ParameterKey=DBPassword,ParameterValue=postgresql123! \
    ParameterKey=DBName,ParameterValue=hrmsdb \
    ParameterKey=JWTSecret,ParameterValue=<YOUR_SECRET> \
    ParameterKey=BackendImageURI,ParameterValue=<BACKEND_IMAGE_URI> \
    ParameterKey=FrontendImageURI,ParameterValue=<FRONTEND_IMAGE_URI> \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

## Post-Deployment Tasks

### 1. Get Application URL

```bash
aws cloudformation describe-stacks \
  --stack-name hrms-prod \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' \
  --output text
```

### 2. Initialize Database (First Time Only)

Connect to the backend container and run migrations:

```bash
# Get the backend task ID
TASK_ID=$(aws ecs list-tasks \
  --cluster hrms-prod-cluster \
  --service-name hrms-prod-backend \
  --query 'taskArns[0]' \
  --output text | rev | cut -d'/' -f1 | rev)

# Execute migration commands
aws ecs execute-command \
  --cluster hrms-prod-cluster \
  --task $TASK_ID \
  --container backend \
  --interactive \
  --command "/bin/sh"
```

Then inside the container, run your SQL migrations located in `/app/cmd/migrations/`.

### 3. Create Initial Admin User

Use the SQL scripts in `backend/scripts/setup-initial-admin.sql`:

```sql
-- Connect to your RDS instance and run:
-- See backend/scripts/setup-initial-admin.sql
```

### 4. Configure DNS (Optional)

Point your domain to the Load Balancer DNS:

```bash
# Get ALB DNS name
aws cloudformation describe-stacks \
  --stack-name hrms-prod \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerDNS`].OutputValue' \
  --output text

# Create CNAME record in Route53 or your DNS provider
# Example: hrms.yourdomain.com -> hrms-prod-alb-123456789.us-east-1.elb.amazonaws.com
```

### 5. Enable HTTPS (Recommended)

1. Request an ACM certificate for your domain
2. Update the stack with the certificate ARN:

```bash
aws cloudformation update-stack \
  --stack-name hrms-prod \
  --use-previous-template \
  --parameters \
    ParameterKey=Environment,UsePreviousValue=true \
    ParameterKey=DBUsername,UsePreviousValue=true \
    ParameterKey=DBPassword,UsePreviousValue=true \
    ParameterKey=DBName,UsePreviousValue=true \
    ParameterKey=JWTSecret,UsePreviousValue=true \
    ParameterKey=CertificateArn,ParameterValue=<YOUR_ACM_CERT_ARN> \
  --capabilities CAPABILITY_NAMED_IAM
```

## Stack Components

### Network Architecture
- **VPC**: 10.0.0.0/16
- **Public Subnets**: 10.0.1.0/24, 10.0.2.0/24 (ALB)
- **Private Subnets**: 10.0.11.0/24, 10.0.12.0/24 (ECS, RDS)
- **NAT Gateways**: 2 (one per AZ for high availability)

### Database Configuration
- **Engine**: PostgreSQL 15.4
- **Instance Class**: db.t3.micro (adjustable)
- **Storage**: 20GB GP3 (encrypted)
- **Backups**: 7-day retention
- **Multi-AZ**: Can be enabled for production

### ECS Services
- **Backend**: 2-4 tasks (CPU: 512, Memory: 1024MB)
- **Frontend**: 2-4 tasks (CPU: 256, Memory: 512MB)
- **Auto Scaling**: Based on CPU utilization (70% target)

### Security
- **Encryption**: All data encrypted at rest and in transit
- **Security Groups**: Strict ingress/egress rules
- **IAM Roles**: Least privilege access
- **Private Subnets**: Application and database isolated

## Updating the Stack

### Update Container Images

```bash
# Build and push new images
docker build -t hrms-backend:v2 ./backend
docker tag hrms-backend:v2 <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/hrms-prod-backend:latest
docker push <ACCOUNT_ID>.dkr.ecr.us-east-1.amazonaws.com/hrms-prod-backend:latest

# Force new deployment
aws ecs update-service \
  --cluster hrms-prod-cluster \
  --service hrms-prod-backend \
  --force-new-deployment
```

### Update Stack Parameters

```bash
aws cloudformation update-stack \
  --stack-name hrms-prod \
  --use-previous-template \
  --parameters \
    ParameterKey=Environment,ParameterValue=production \
    ParameterKey=DBUsername,UsePreviousValue=true \
    ParameterKey=DBPassword,UsePreviousValue=true \
    ParameterKey=DBName,UsePreviousValue=true \
    ParameterKey=JWTSecret,ParameterValue=<NEW_SECRET> \
  --capabilities CAPABILITY_NAMED_IAM
```

## Monitoring and Logs

### View ECS Service Logs

```bash
# Backend logs
aws logs tail /ecs/hrms-prod/backend --follow

# Frontend logs
aws logs tail /ecs/hrms-prod/frontend --follow
```

### View RDS Logs

```bash
# PostgreSQL logs
aws rds download-db-log-file-portion \
  --db-instance-identifier hrms-prod-postgres \
  --log-file-name error/postgresql.log
```

### CloudWatch Metrics

Access CloudWatch Console to view:
- ECS service CPU/Memory utilization
- ALB request count and response times
- RDS performance metrics
- Target group health

## Scaling

### Manual Scaling

```bash
# Scale backend service
aws ecs update-service \
  --cluster hrms-prod-cluster \
  --service hrms-prod-backend \
  --desired-count 4

# Scale frontend service
aws ecs update-service \
  --cluster hrms-prod-cluster \
  --service hrms-prod-frontend \
  --desired-count 4
```

### Upgrade Database Instance

```bash
aws cloudformation update-stack \
  --stack-name hrms-prod \
  --use-previous-template \
  --parameters \
    ParameterKey=Environment,UsePreviousValue=true \
    ParameterKey=DBUsername,UsePreviousValue=true \
    ParameterKey=DBPassword,UsePreviousValue=true \
    ParameterKey=DBName,UsePreviousValue=true \
    ParameterKey=JWTSecret,UsePreviousValue=true \
  --capabilities CAPABILITY_NAMED_IAM

# Note: Modify RDSInstance.DBInstanceClass in the template first
```

## Troubleshooting

### Stack Creation Fails

1. Check CloudFormation events:
```bash
aws cloudformation describe-stack-events \
  --stack-name hrms-prod \
  --query 'StackEvents[?ResourceStatus==`CREATE_FAILED`]'
```

2. Common issues:
   - Insufficient IAM permissions
   - ECR images not available
   - Invalid parameter values
   - Resource limits exceeded

### ECS Tasks Not Starting

1. Check task definition:
```bash
aws ecs describe-tasks \
  --cluster hrms-prod-cluster \
  --tasks <TASK_ID>
```

2. Check service events:
```bash
aws ecs describe-services \
  --cluster hrms-prod-cluster \
  --services hrms-prod-backend
```

3. Common issues:
   - Image pull errors (ECR permissions)
   - Health check failures
   - Insufficient memory/CPU
   - Database connection issues

### Health Check Failures

1. Verify backend health endpoint:
```bash
# Get task IP
TASK_IP=$(aws ecs describe-tasks \
  --cluster hrms-prod-cluster \
  --tasks <TASK_ID> \
  --query 'tasks[0].containers[0].networkInterfaces[0].privateIpv4Address' \
  --output text)

# Test health endpoint from bastion/VPN
curl http://$TASK_IP:8080/api/health
```

2. Check RDS connectivity:
```bash
# From ECS task
psql -h <RDS_ENDPOINT> -U postgres -d hrmsdb
```

## Cost Optimization

### Development Environment

For a cheaper dev stack:
1. Use single NAT Gateway
2. Reduce task count to 1
3. Use smaller RDS instance (db.t3.micro)
4. Disable RDS Multi-AZ
5. Reduce backup retention

### Production Recommendations

1. Enable RDS Multi-AZ for high availability
2. Use larger instance types (db.t3.small or higher)
3. Consider Reserved Instances for cost savings
4. Enable CloudWatch Container Insights selectively
5. Implement S3 lifecycle policies for document storage

## Cleanup

To delete the entire stack:

```bash
# Warning: This will delete all resources including the database
aws cloudformation delete-stack --stack-name hrms-prod

# Monitor deletion
aws cloudformation wait stack-delete-complete --stack-name hrms-prod
```

**Note**: The RDS instance will create a final snapshot before deletion due to `DeletionPolicy: Snapshot`.

## Support and Resources

- [AWS ECS Documentation](https://docs.aws.amazon.com/ecs/)
- [AWS RDS PostgreSQL](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html)
- [CloudFormation Best Practices](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/best-practices.html)

## Security Best Practices

1. **Change default passwords** immediately after deployment
2. **Rotate JWT secrets** regularly
3. **Enable AWS WAF** on the ALB for DDoS protection
4. **Configure CloudWatch alarms** for unusual activity
5. **Enable VPC Flow Logs** for network monitoring
6. **Use AWS Secrets Manager** for sensitive credentials (instead of parameters)
7. **Implement database encryption** with KMS
8. **Regular security audits** using AWS Security Hub
