# HRMS Deployment for Windows (PowerShell)

This guide is specifically for deploying the HRMS application using PowerShell on Windows.

## Prerequisites

1. **AWS CLI** - [Download for Windows](https://awscli.amazonaws.com/AWSCLIV2.msi)
2. **Docker Desktop** - [Download for Windows](https://www.docker.com/products/docker-desktop)
3. **PowerShell 5.1 or later** (comes with Windows)
4. **AWS Account** with configured credentials

## Quick Start

### 1. Configure AWS Credentials

Open PowerShell and run:
```powershell
aws configure
```

Enter your:
- AWS Access Key ID
- AWS Secret Access Key
- Default region (e.g., us-east-1)
- Default output format (json)

### 2. Verify Prerequisites

```powershell
# Check AWS CLI
aws --version

# Check Docker
docker --version

# Check PowerShell version
$PSVersionTable.PSVersion
```

### 3. Deploy the Application

**Option A: Automated Deployment (Recommended)**

```powershell
.\deploy.ps1
```

This will:
- ✅ Check prerequisites
- ✅ Create ECR repositories
- ✅ Build and push Docker images
- ✅ Deploy CloudFormation stack
- ✅ Display application URL

**Option B: Custom Deployment**

```powershell
# Deploy with custom parameters
.\deploy.ps1 -StackName "my-hrms" -Region "us-west-2"

# Skip image building (use existing images)
.\deploy.ps1 -SkipImages

# Use custom JWT secret
.\deploy.ps1 -JWTSecret "your-very-secure-secret-key-here"
```

**Option C: View Help**

```powershell
.\deploy.ps1 -Help
```

### 4. Monitor Deployment

The script will automatically wait for deployment to complete (10-15 minutes).

You can also monitor in the AWS Console:
1. Open [CloudFormation Console](https://console.aws.amazon.com/cloudformation)
2. Find your stack (default: hrms-prod)
3. Click on "Events" tab to see progress

### 5. Initialize Database

After deployment completes:

```powershell
.\init-database.ps1
```

This will:
- Generate SQL initialization file
- Display connection information
- Create helper scripts for connecting

**Then execute the SQL file using one of these methods:**

**Method 1: Using psql (if installed)**
```powershell
psql -h <DB_ENDPOINT> -U postgres -d hrmsdb -f "C:\Users\...\init_database.sql"
```

**Method 2: Using GUI Tool**
- **pgAdmin**: [Download here](https://www.pgadmin.org/download/pgadmin-4-windows/)
- **DBeaver**: [Download here](https://dbeaver.io/download/)
- **Azure Data Studio**: [Download here](https://docs.microsoft.com/en-us/sql/azure-data-studio/download)

Connect using:
- Host: (from script output)
- Port: 5432
- Database: hrmsdb
- Username: postgres
- Password: postgresql123!

Then execute the generated SQL file.

## Common PowerShell Commands

### View Deployment Status

```powershell
aws cloudformation describe-stacks `
    --stack-name hrms-prod `
    --region us-east-1 `
    --query 'Stacks[0].StackStatus' `
    --output text
```

### Get Application URL

```powershell
aws cloudformation describe-stacks `
    --stack-name hrms-prod `
    --region us-east-1 `
    --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' `
    --output text
```

### View Service Logs

```powershell
# Backend logs
aws logs tail /ecs/hrms-prod/backend --follow --region us-east-1

# Frontend logs
aws logs tail /ecs/hrms-prod/frontend --follow --region us-east-1
```

### Update Application (After Code Changes)

```powershell
# Rebuild and push images
cd backend
docker build -t hrms-backend .
$ACCOUNT_ID = aws sts get-caller-identity --query Account --output text
docker tag hrms-backend:latest "$ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/hrms-prod-backend:latest"
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com"
docker push "$ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/hrms-prod-backend:latest"

# Force new deployment
aws ecs update-service `
    --cluster hrms-prod-cluster `
    --service hrms-prod-backend `
    --force-new-deployment `
    --region us-east-1
```

### Scale Services

```powershell
# Scale backend to 4 tasks
aws ecs update-service `
    --cluster hrms-prod-cluster `
    --service hrms-prod-backend `
    --desired-count 4 `
    --region us-east-1

# Scale frontend to 4 tasks
aws ecs update-service `
    --cluster hrms-prod-cluster `
    --service hrms-prod-frontend `
    --desired-count 4 `
    --region us-east-1
```

### Delete Stack (Cleanup)

```powershell
# WARNING: This deletes all resources
aws cloudformation delete-stack `
    --stack-name hrms-prod `
    --region us-east-1

# Monitor deletion
aws cloudformation wait stack-delete-complete `
    --stack-name hrms-prod `
    --region us-east-1
```

## Troubleshooting

### PowerShell Execution Policy Error

If you get an error about script execution being disabled:

```powershell
# Check current policy
Get-ExecutionPolicy

# Allow scripts for current user
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or run script with bypass
PowerShell.exe -ExecutionPolicy Bypass -File .\deploy.ps1
```

### Docker Not Running

Make sure Docker Desktop is running:
1. Start Docker Desktop from Start Menu
2. Wait for Docker to fully start (whale icon in system tray)
3. Run `docker ps` to verify it's working

### AWS CLI Not Found

Add AWS CLI to your PATH:
1. Open System Properties → Environment Variables
2. Edit PATH variable
3. Add: `C:\Program Files\Amazon\AWSCLIV2`
4. Restart PowerShell

### ECR Login Failed

```powershell
# Get new login credentials
$PASSWORD = aws ecr get-login-password --region us-east-1
$ACCOUNT_ID = aws sts get-caller-identity --query Account --output text
echo $PASSWORD | docker login --username AWS --password-stdin "$ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com"
```

### Stack Creation Failed

Check the CloudFormation events:
```powershell
aws cloudformation describe-stack-events `
    --stack-name hrms-prod `
    --region us-east-1 `
    --query 'StackEvents[?ResourceStatus==`CREATE_FAILED`]' `
    --output table
```

Common causes:
- Insufficient IAM permissions
- Service limits reached
- Invalid parameter values
- ECR images not available

### Database Connection Timeout

If you can't connect to the database:
1. Check if RDS instance is running in AWS Console
2. Verify security group allows connections from ECS tasks
3. Check if database is in the same VPC as ECS tasks
4. Try connecting from an ECS task:

```powershell
# Get a task ARN
$TASK_ARN = aws ecs list-tasks `
    --cluster hrms-prod-cluster `
    --service-name hrms-prod-backend `
    --region us-east-1 `
    --query 'taskArns[0]' `
    --output text

# Connect to task (requires ECS Exec enabled)
aws ecs execute-command `
    --cluster hrms-prod-cluster `
    --task $TASK_ARN `
    --container backend `
    --interactive `
    --command "/bin/sh" `
    --region us-east-1
```

## File Structure

```
├── cloudformation-stack.yaml  # Infrastructure definition
├── deploy.ps1                 # PowerShell deployment script
├── init-database.ps1          # PowerShell database init script
├── deploy.sh                  # Bash deployment script (Linux/Mac)
├── init-database.sh           # Bash database init script (Linux/Mac)
├── DEPLOYMENT_GUIDE.md        # Comprehensive deployment guide
├── QUICK_REFERENCE.md         # Quick command reference
└── README_WINDOWS.md          # This file
```

## Best Practices

1. **Change Default Passwords**: Immediately change the default admin password after first login
2. **Use Secrets Manager**: For production, store credentials in AWS Secrets Manager instead of parameters
3. **Enable HTTPS**: Request an ACM certificate and update the stack
4. **Set Up Monitoring**: Configure CloudWatch alarms for critical metrics
5. **Regular Backups**: Enable automated RDS snapshots (already configured)
6. **Update Regularly**: Keep Docker images and dependencies up to date

## Cost Optimization

For development/testing environments:
```powershell
# Scale down to 1 task each
aws ecs update-service --cluster hrms-prod-cluster --service hrms-prod-backend --desired-count 1 --region us-east-1
aws ecs update-service --cluster hrms-prod-cluster --service hrms-prod-frontend --desired-count 1 --region us-east-1

# Or delete the stack when not in use
aws cloudformation delete-stack --stack-name hrms-prod --region us-east-1
```

## Support Resources

- **AWS CloudFormation**: https://docs.aws.amazon.com/cloudformation/
- **AWS ECS**: https://docs.aws.amazon.com/ecs/
- **AWS RDS**: https://docs.aws.amazon.com/rds/
- **Docker**: https://docs.docker.com/
- **PowerShell**: https://docs.microsoft.com/en-us/powershell/

## Next Steps After Deployment

1. ✅ Access your application at the LoadBalancerURL
2. ✅ Initialize the database with migrations
3. ✅ Create an admin user (see backend/scripts/)
4. ✅ Configure custom domain (optional)
5. ✅ Enable HTTPS with ACM certificate (recommended)
6. ✅ Set up CloudWatch alarms
7. ✅ Review and adjust security settings
8. ✅ Test the application thoroughly

## Getting Help

If you encounter issues:
1. Check the troubleshooting section above
2. Review CloudFormation events in AWS Console
3. Check ECS service events and task logs
4. Review the comprehensive DEPLOYMENT_GUIDE.md
5. Use the QUICK_REFERENCE.md for common commands

---

**Note**: All PowerShell scripts are compatible with Windows PowerShell 5.1+ and PowerShell Core 7+.
