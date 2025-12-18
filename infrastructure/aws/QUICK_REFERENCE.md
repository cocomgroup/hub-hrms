# HRMS CloudFormation Quick Reference

## Quick Deploy (One Command)

```bash
./deploy.sh
```

This will:
1. Check prerequisites (AWS CLI, Docker)
2. Create ECR repositories
3. Build and push Docker images
4. Deploy CloudFormation stack
5. Display application URL

## Manual Deployment Steps

### 1. Create ECR Repositories (First Time Only)

```bash
aws ecr create-repository --repository-name hrms-prod-backend --region us-east-1
aws ecr create-repository --repository-name hrms-prod-frontend --region us-east-1
```

### 2. Build and Push Images

```bash
# Get account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com

# Build and push backend
cd backend
docker build -t hrms-backend .
docker tag hrms-backend:latest ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/hrms-prod-backend:latest
docker push ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/hrms-prod-backend:latest

# Build and push frontend
cd ../frontend
docker build -t hrms-frontend .
docker tag hrms-frontend:latest ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/hrms-prod-frontend:latest
docker push ${ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/hrms-prod-frontend:latest
```

### 3. Deploy Stack

```bash
aws cloudformation create-stack \
  --stack-name hrms-prod \
  --template-body file://cloudformation-stack.yaml \
  --parameters \
    ParameterKey=Environment,ParameterValue=production \
    ParameterKey=DBUsername,ParameterValue=postgres \
    ParameterKey=DBPassword,ParameterValue=postgresql123! \
    ParameterKey=DBName,ParameterValue=hrmsdb \
    ParameterKey=JWTSecret,ParameterValue=$(openssl rand -base64 32) \
  --capabilities CAPABILITY_NAMED_IAM \
  --region us-east-1
```

### 4. Monitor Deployment

```bash
# Watch progress
aws cloudformation describe-stack-events \
  --stack-name hrms-prod \
  --region us-east-1 \
  --max-items 20 \
  --query 'StackEvents[*].[Timestamp,ResourceStatus,ResourceType,LogicalResourceId]' \
  --output table

# Wait for completion
aws cloudformation wait stack-create-complete --stack-name hrms-prod --region us-east-1
```

### 5. Get Application URL

```bash
aws cloudformation describe-stacks \
  --stack-name hrms-prod \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' \
  --output text
```

### 6. Initialize Database

```bash
./init-database.sh
# Then manually run the generated SQL file
```

## Common Operations

### View Logs

```bash
# Backend logs
aws logs tail /ecs/hrms-prod/backend --follow

# Frontend logs
aws logs tail /ecs/hrms-prod/frontend --follow
```

### Update Application

```bash
# Rebuild and push images, then:
aws ecs update-service \
  --cluster hrms-prod-cluster \
  --service hrms-prod-backend \
  --force-new-deployment

aws ecs update-service \
  --cluster hrms-prod-cluster \
  --service hrms-prod-frontend \
  --force-new-deployment
```

### Scale Services

```bash
# Scale backend
aws ecs update-service \
  --cluster hrms-prod-cluster \
  --service hrms-prod-backend \
  --desired-count 4

# Scale frontend
aws ecs update-service \
  --cluster hrms-prod-cluster \
  --service hrms-prod-frontend \
  --desired-count 4
```

### Database Connection

```bash
# Get database endpoint
DB_ENDPOINT=$(aws cloudformation describe-stacks \
  --stack-name hrms-prod \
  --query 'Stacks[0].Outputs[?OutputKey==`DatabaseEndpoint`].OutputValue' \
  --output text)

# Connect
psql -h $DB_ENDPOINT -U postgres -d hrmsdb
```

### Delete Stack

```bash
# WARNING: This deletes all resources
aws cloudformation delete-stack --stack-name hrms-prod --region us-east-1
```

## Stack Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| Environment | production | Environment name |
| DBUsername | postgres | Database master username |
| DBPassword | postgresql123! | Database master password |
| DBName | hrmsdb | Database name |
| JWTSecret | (auto-generated) | JWT secret for authentication |
| BackendImageURI | (from ECR) | Backend Docker image URI |
| FrontendImageURI | (from ECR) | Frontend Docker image URI |
| CertificateArn | (optional) | ACM Certificate for HTTPS |

## Stack Outputs

| Output | Description |
|--------|-------------|
| LoadBalancerURL | Application URL |
| LoadBalancerDNS | ALB DNS name |
| DatabaseEndpoint | RDS endpoint |
| DatabaseName | Database name |
| S3BucketName | Documents bucket |
| BackendECRRepository | Backend ECR URI |
| FrontendECRRepository | Frontend ECR URI |
| ECSClusterName | ECS cluster name |

## Architecture

```
Internet
    ↓
Application Load Balancer (Public Subnets)
    ↓
├─→ Frontend Service (Private Subnets)
│   └─→ Nginx:3000 → Svelte App
│
└─→ Backend Service (Private Subnets)
    └─→ Go API:8080
        └─→ RDS PostgreSQL (Private Subnets)
        └─→ S3 Bucket (Documents)
```

## Resources Created

- **VPC**: 10.0.0.0/16
- **Subnets**: 2 public, 2 private (multi-AZ)
- **NAT Gateways**: 2 (high availability)
- **Application Load Balancer**: Internet-facing
- **ECS Cluster**: Fargate
- **ECS Services**: 2 (frontend, backend)
- **RDS PostgreSQL**: db.t3.micro, 20GB
- **S3 Bucket**: Encrypted, versioned
- **ECR Repositories**: 2 (frontend, backend)
- **CloudWatch Logs**: Service logs
- **Auto Scaling**: CPU-based (70% target)

## Cost Estimate (us-east-1)

Monthly costs (approximate):
- **ECS Fargate**: ~$50-100 (2-4 tasks)
- **RDS db.t3.micro**: ~$15
- **NAT Gateways**: ~$65 (2x $32.50)
- **Application Load Balancer**: ~$23
- **Data Transfer**: Variable
- **S3 Storage**: Variable (~$0.023/GB)

**Total**: ~$150-200/month baseline

## Security Features

- ✅ All data encrypted at rest and in transit
- ✅ Private subnets for application and database
- ✅ Security groups with least privilege
- ✅ IAM roles with minimal permissions
- ✅ RDS automated backups (7-day retention)
- ✅ CloudWatch logging enabled
- ✅ Container insights enabled

## Troubleshooting

### Stack Creation Failed
```bash
# Check failed resources
aws cloudformation describe-stack-events \
  --stack-name hrms-prod \
  --query 'StackEvents[?ResourceStatus==`CREATE_FAILED`]'
```

### ECS Tasks Not Starting
```bash
# Check task failures
aws ecs describe-services \
  --cluster hrms-prod-cluster \
  --services hrms-prod-backend hrms-prod-frontend
```

### Health Check Failures
```bash
# Check target group health
aws elbv2 describe-target-health \
  --target-group-arn <TARGET_GROUP_ARN>
```

### Database Connection Issues
```bash
# Test from ECS task
aws ecs execute-command \
  --cluster hrms-prod-cluster \
  --task <TASK_ID> \
  --container backend \
  --interactive \
  --command "/bin/sh"

# Then test connection
wget -O- http://localhost:8080/api/health
```

## Support

For detailed information, see:
- **DEPLOYMENT_GUIDE.md** - Complete deployment documentation
- **cloudformation-stack.yaml** - Full infrastructure definition
- **deploy.sh** - Automated deployment script
- **init-database.sh** - Database initialization script

## Environment Variables (Backend)

The backend container receives these environment variables:

```bash
DB_HOST=<RDS_ENDPOINT>
DB_PORT=5432
DB_NAME=hrmsdb
DB_USER=postgres
DB_PASSWORD=postgresql123!
DATABASE_URL=postgres://...
PORT=8080
FRONTEND_URL=http://<ALB_DNS>
JWT_SECRET=<GENERATED_SECRET>
AWS_REGION=us-east-1
S3_BUCKET=<BUCKET_NAME>
ENVIRONMENT=production
```

## Nginx Configuration (Frontend)

Frontend proxies `/api/*` requests to backend service:
- Frontend: Port 3000
- Backend: Port 8080
- Proxy: `location /api { proxy_pass http://backend:8080; }`

## Next Steps After Deployment

1. **Access Application**: Visit the LoadBalancerURL
2. **Initialize Database**: Run init-database.sh and execute SQL
3. **Create Admin User**: Use backend/scripts/setup-initial-admin.sql
4. **Configure DNS**: Point your domain to ALB (optional)
5. **Enable HTTPS**: Add ACM certificate (recommended)
6. **Set Up Monitoring**: Configure CloudWatch alarms
7. **Review Security**: Security audit and hardening
8. **Backup Strategy**: Configure RDS snapshots
