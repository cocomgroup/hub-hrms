# Frontend to Backend Connectivity Troubleshooting

Complete guide to diagnose and fix communication issues between frontend and backend in AWS ECS.

---

## 🔍 Problem Overview

**Issue**: Frontend can't reach backend API endpoints

**Symptoms:**
- Frontend loads but API calls fail
- Console shows network errors (CORS, 404, timeout)
- "Failed to fetch" or "Network Error" in browser
- Backend logs show no incoming requests

---

## Step 1: Verify Both Services Are Healthy

### Check Backend Health

```powershell
# Check backend service
aws ecs describe-services `
  --cluster hub-hrms-cluster `
  --services hub-hrms-backend-service `
  --region us-east-1 `
  --query 'services[0].[runningCount,desiredCount]'

# Check backend target health
aws elbv2 describe-target-health `
  --target-group-arn $(aws elbv2 describe-target-groups --region us-east-1 --query "TargetGroups[?contains(TargetGroupName, 'backend')].TargetGroupArn" --output text) `
  --region us-east-1 `
  --query 'TargetHealthDescriptions[*].[Target.Id,TargetHealth.State]' `
  --output table
```

### Check Frontend Health

```powershell
# Check frontend service
aws ecs describe-services `
  --cluster hub-hrms-cluster `
  --services hub-hrms-frontend-service `
  --region us-east-1 `
  --query 'services[0].[runningCount,desiredCount]'

# Check frontend target health
aws elbv2 describe-target-health `
  --target-group-arn $(aws elbv2 describe-target-groups --region us-east-1 --query "TargetGroups[?contains(TargetGroupName, 'frontend')].TargetGroupArn" --output text) `
  --region us-east-1 `
  --query 'TargetHealthDescriptions[*].[Target.Id,TargetHealth.State]' `
  --output table
```

**Both should show "healthy" status!**

---

## Step 2: Get Backend URL

The frontend needs to know where the backend is:

```powershell
# Get backend ALB DNS name
$BACKEND_URL = aws elbv2 describe-load-balancers `
  --region us-east-1 `
  --query "LoadBalancers[?contains(LoadBalancerName, 'backend')].DNSName" `
  --output text

Write-Host "Backend URL: http://$BACKEND_URL" -ForegroundColor Green

# Test if backend is reachable
curl http://$BACKEND_URL/api/health
# Should return: {"status":"healthy"}
```

---

## Step 3: Check Frontend Environment Variables

### Option A: Frontend Built with Environment Variable

**Frontend needs backend URL at build time:**

```powershell
# Check if VITE_API_URL is set in ECS Task Definition
aws ecs describe-task-definition `
  --task-definition hub-hrms-frontend `
  --region us-east-1 `
  --query 'taskDefinition.containerDefinitions[0].environment' `
  --output table
```

**Should include:**
```json
{
  "name": "VITE_API_URL",
  "value": "http://backend-alb-xxxxx.us-east-1.elb.amazonaws.com"
}
```

**If missing, add it:**

```powershell
# Get current task definition
aws ecs describe-task-definition `
  --task-definition hub-hrms-frontend `
  --region us-east-1 > frontend-task-def.json

# Edit the JSON to add:
# "environment": [
#   {
#     "name": "VITE_API_URL",
#     "value": "http://YOUR_BACKEND_ALB_URL"
#   }
# ]

# Register new task definition
aws ecs register-task-definition --cli-input-json file://frontend-task-def.json

# Update service to use new definition
aws ecs update-service `
  --cluster hub-hrms-cluster `
  --service hub-hrms-frontend-service `
  --force-new-deployment `
  --region us-east-1
```

### Option B: Frontend Uses Relative URLs

**If frontend uses relative URLs like `/api/...`, you need an ALB rule to proxy to backend:**

This requires a **single ALB** with path-based routing:
- `/` → Frontend target group
- `/api/*` → Backend target group

---

## Step 4: Verify CORS Configuration

### Check Backend CORS Settings

**Your backend MUST allow requests from the frontend domain:**

```go
// In main.go or router setup
cors.Handler(cors.Options{
    AllowedOrigins:   []string{
        "http://localhost:5173",           // Dev
        "http://YOUR_FRONTEND_ALB_URL",    // Production
        "*",                               // Allow all (less secure)
    },
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
    AllowCredentials: true,
    MaxAge:           300,
})
```

**Test CORS:**

```powershell
# From your local machine, test CORS preflight
curl -X OPTIONS http://YOUR_BACKEND_ALB/api/health `
  -H "Origin: http://YOUR_FRONTEND_ALB" `
  -H "Access-Control-Request-Method: GET" `
  -v

# Should see:
# < Access-Control-Allow-Origin: http://YOUR_FRONTEND_ALB
# < Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

---

## Step 5: Test Backend API Directly

```powershell
# Get backend URL
$BACKEND_URL = aws elbv2 describe-load-balancers `
  --region us-east-1 `
  --query "LoadBalancers[?contains(LoadBalancerName, 'backend')].DNSName" `
  --output text

# Test health endpoint
curl http://$BACKEND_URL/api/health

# Test authentication
curl -X POST http://$BACKEND_URL/api/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"admin@cocomgroup.com","password":"admin123"}'

# Should return JWT token
```

If these work, backend is accessible. Issue is in frontend configuration.

---

## Step 6: Check Frontend Code Configuration

### Check where API calls are made

**Look for API base URL configuration:**

**Common patterns:**

**Pattern 1: Environment variable (Vite)**
```typescript
// In your frontend code
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

fetch(`${API_BASE_URL}/api/pto/balance`, {
  headers: { 'Authorization': `Bearer ${token}` }
});
```

**Pattern 2: Relative URLs**
```typescript
// Uses same domain as frontend
fetch('/api/pto/balance', {
  headers: { 'Authorization': `Bearer ${token}` }
});
```

**Pattern 3: Hardcoded URL**
```typescript
fetch('http://localhost:8080/api/pto/balance', {
  headers: { 'Authorization': `Bearer ${token}` }
});
```

---

## Step 7: Frontend Configuration Solutions

### Solution A: Environment Variable (Recommended)

**1. Create `.env` file for local development:**

```bash
# frontend/.env
VITE_API_URL=http://localhost:8080
```

**2. Add to CloudFormation/ECS Task Definition:**

```yaml
Environment:
  - Name: VITE_API_URL
    Value: !Sub 'http://${BackendALB.DNSName}'
  - Name: NODE_ENV
    Value: production
```

**3. Use in code:**

```typescript
// src/config.ts
export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// src/stores/auth.ts or API calls
import { API_BASE_URL } from './config';

fetch(`${API_BASE_URL}/api/auth/login`, {...});
```

**4. Rebuild frontend with environment variable:**

```powershell
# Build with backend URL
$env:VITE_API_URL = "http://YOUR_BACKEND_ALB_URL"
npm run build

# Or build Docker image with build arg
docker build --build-arg VITE_API_URL=http://YOUR_BACKEND_ALB_URL -t frontend .
```

### Solution B: Shared ALB with Path-Based Routing

**Use one ALB for both frontend and backend:**

```yaml
# CloudFormation
ALBListenerRule:
  Type: AWS::ElasticLoadBalancingV2::ListenerRule
  Properties:
    ListenerArn: !Ref ALBListener
    Priority: 1
    Conditions:
      - Field: path-pattern
        Values:
          - /api/*
    Actions:
      - Type: forward
        TargetGroupArn: !Ref BackendTargetGroup

ALBListenerRuleDefault:
  Type: AWS::ElasticLoadBalancingV2::ListenerRule
  Properties:
    ListenerArn: !Ref ALBListener
    Priority: 2
    Conditions:
      - Field: path-pattern
        Values:
          - /*
    Actions:
      - Type: forward
        TargetGroupArn: !Ref FrontendTargetGroup
```

**Then frontend uses relative URLs:**
```typescript
fetch('/api/pto/balance', {...});  // No base URL needed!
```

---

## Step 8: Check Security Groups

### Verify Backend Security Group

```powershell
# Get backend task security group
$BACKEND_SG = aws ecs describe-services `
  --cluster hub-hrms-cluster `
  --services hub-hrms-backend-service `
  --region us-east-1 `
  --query 'services[0].networkConfiguration.awsvpcConfiguration.securityGroups[0]' `
  --output text

# Check inbound rules
aws ec2 describe-security-groups `
  --group-ids $BACKEND_SG `
  --region us-east-1 `
  --query 'SecurityGroups[0].IpPermissions' `
  --output table
```

**Backend MUST allow:**
- Port 8080 from ALB Security Group
- Port 8080 from Frontend Security Group (if direct communication)

### Verify Frontend Security Group

```powershell
# Get frontend task security group
$FRONTEND_SG = aws ecs describe-services `
  --cluster hub-hrms-cluster `
  --services hub-hrms-frontend-service `
  --region us-east-1 `
  --query 'services[0].networkConfiguration.awsvpcConfiguration.securityGroups[0]' `
  --output text

# Check outbound rules
aws ec2 describe-security-groups `
  --group-ids $FRONTEND_SG `
  --region us-east-1 `
  --query 'SecurityGroups[0].IpPermissionsEgress' `
  --output table
```

**Frontend MUST allow:**
- Outbound to all (0.0.0.0/0) on all ports
- Or outbound to Backend SG on port 8080

---

## Step 9: Check from Inside Frontend Container

**SSH into a running frontend task:**

```powershell
# Get task ID
$TASK_ID = aws ecs list-tasks `
  --cluster hub-hrms-cluster `
  --service-name hub-hrms-frontend-service `
  --region us-east-1 `
  --query 'taskArns[0]' `
  --output text

# Enable ECS Exec (if not already enabled)
aws ecs update-service `
  --cluster hub-hrms-cluster `
  --service hub-hrms-frontend-service `
  --enable-execute-command `
  --region us-east-1

# Execute command in container
aws ecs execute-command `
  --cluster hub-hrms-cluster `
  --task $TASK_ID `
  --container frontend `
  --command "/bin/sh" `
  --interactive `
  --region us-east-1
```

**Once inside container:**

```bash
# Test DNS resolution
nslookup backend-alb-xxxxx.us-east-1.elb.amazonaws.com

# Test connectivity
wget -O- http://backend-alb-xxxxx.us-east-1.elb.amazonaws.com/api/health

# Check environment variables
env | grep VITE
env | grep API

# Exit
exit
```

---

## Step 10: Check Browser Console

**Open your frontend in a browser and check Developer Tools:**

**Chrome/Edge:**
- Press F12
- Go to "Network" tab
- Try to use the app
- Look for failed API requests

**Common errors:**

### Error: "CORS policy: No 'Access-Control-Allow-Origin' header"
**Fix**: Update backend CORS to allow frontend origin

### Error: "Failed to fetch" or "net::ERR_NAME_NOT_RESOLVED"
**Fix**: Frontend has wrong backend URL

### Error: "404 Not Found"
**Fix**: Backend endpoint doesn't exist or path is wrong

### Error: "502 Bad Gateway"
**Fix**: Backend is down or not responding

### Error: "Timeout"
**Fix**: Security group blocking connection or backend is slow

---

## Complete Diagnostic Script

```powershell
# Save as check-connectivity.ps1

param(
    [string]$Region = "us-east-1",
    [string]$Cluster = "hub-hrms-cluster"
)

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Frontend-Backend Connectivity Check" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# 1. Get Backend ALB URL
Write-Host "[1/6] Getting Backend ALB URL..." -ForegroundColor Yellow
$backendALB = aws elbv2 describe-load-balancers `
  --region $Region `
  --query "LoadBalancers[?contains(LoadBalancerName, 'backend')].DNSName" `
  --output text

if ($backendALB) {
    Write-Host "   Backend URL: http://$backendALB" -ForegroundColor Green
} else {
    Write-Host "   ✗ Backend ALB not found!" -ForegroundColor Red
}

# 2. Test Backend Health
Write-Host "`n[2/6] Testing Backend Health..." -ForegroundColor Yellow
try {
    $response = curl -s "http://$backendALB/api/health" 2>&1
    if ($response -match "healthy") {
        Write-Host "   ✓ Backend is responding" -ForegroundColor Green
        Write-Host "   Response: $response" -ForegroundColor Gray
    } else {
        Write-Host "   ✗ Backend returned unexpected response" -ForegroundColor Red
        Write-Host "   Response: $response" -ForegroundColor Gray
    }
} catch {
    Write-Host "   ✗ Cannot reach backend" -ForegroundColor Red
}

# 3. Get Frontend ALB URL
Write-Host "`n[3/6] Getting Frontend ALB URL..." -ForegroundColor Yellow
$frontendALB = aws elbv2 describe-load-balancers `
  --region $Region `
  --query "LoadBalancers[?contains(LoadBalancerName, 'frontend')].DNSName" `
  --output text

if ($frontendALB) {
    Write-Host "   Frontend URL: http://$frontendALB" -ForegroundColor Green
} else {
    Write-Host "   ✗ Frontend ALB not found!" -ForegroundColor Red
}

# 4. Check Frontend Environment Variables
Write-Host "`n[4/6] Checking Frontend Environment..." -ForegroundColor Yellow
$frontendEnv = aws ecs describe-task-definition `
  --task-definition hub-hrms-frontend `
  --region $Region `
  --query 'taskDefinition.containerDefinitions[0].environment' `
  --output json | ConvertFrom-Json

$apiUrl = $frontendEnv | Where-Object { $_.name -eq "VITE_API_URL" } | Select-Object -ExpandProperty value

if ($apiUrl) {
    Write-Host "   VITE_API_URL: $apiUrl" -ForegroundColor Green
    if ($apiUrl -ne "http://$backendALB") {
        Write-Host "   ⚠ Warning: URL doesn't match backend ALB!" -ForegroundColor Yellow
        Write-Host "   Expected: http://$backendALB" -ForegroundColor Gray
    }
} else {
    Write-Host "   ✗ VITE_API_URL not set!" -ForegroundColor Red
    Write-Host "   Frontend won't know where to find backend" -ForegroundColor Red
}

# 5. Check CORS
Write-Host "`n[5/6] Testing CORS..." -ForegroundColor Yellow
try {
    $corsTest = curl -X OPTIONS "http://$backendALB/api/health" `
      -H "Origin: http://$frontendALB" `
      -H "Access-Control-Request-Method: GET" `
      -v 2>&1
    
    if ($corsTest -match "Access-Control-Allow-Origin") {
        Write-Host "   ✓ CORS is configured" -ForegroundColor Green
    } else {
        Write-Host "   ✗ CORS may not be configured properly" -ForegroundColor Red
    }
} catch {
    Write-Host "   ⚠ Could not test CORS" -ForegroundColor Yellow
}

# 6. Check Security Groups
Write-Host "`n[6/6] Checking Security Groups..." -ForegroundColor Yellow

# Get backend SG
$backendSG = aws ecs describe-services `
  --cluster $Cluster `
  --services hub-hrms-backend-service `
  --region $Region `
  --query 'services[0].networkConfiguration.awsvpcConfiguration.securityGroups[0]' `
  --output text 2>$null

if ($backendSG) {
    Write-Host "   Backend SG: $backendSG" -ForegroundColor Gray
    
    $backendRules = aws ec2 describe-security-groups `
      --group-ids $backendSG `
      --region $Region `
      --query 'SecurityGroups[0].IpPermissions[?FromPort==`8080`]' `
      --output json | ConvertFrom-Json
    
    if ($backendRules) {
        Write-Host "   ✓ Backend allows inbound on port 8080" -ForegroundColor Green
    } else {
        Write-Host "   ✗ Backend may not allow inbound on port 8080" -ForegroundColor Red
    }
}

# Summary
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Summary" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "Backend URL:  http://$backendALB" -ForegroundColor White
Write-Host "Frontend URL: http://$frontendALB" -ForegroundColor White

if ($apiUrl) {
    Write-Host "Frontend configured to use: $apiUrl" -ForegroundColor White
} else {
    Write-Host "Frontend API URL: NOT CONFIGURED" -ForegroundColor Red
}

Write-Host "`nNext Steps:" -ForegroundColor Yellow
if (-not $apiUrl -or $apiUrl -eq "http://localhost:8080") {
    Write-Host "1. Set VITE_API_URL environment variable in ECS task definition" -ForegroundColor White
    Write-Host "2. Rebuild frontend with: VITE_API_URL=http://$backendALB npm run build" -ForegroundColor White
    Write-Host "3. Update ECS service to use new task definition" -ForegroundColor White
} else {
    Write-Host "1. Test frontend in browser: http://$frontendALB" -ForegroundColor White
    Write-Host "2. Open browser console (F12) and check for errors" -ForegroundColor White
    Write-Host "3. Verify backend CORS allows frontend origin" -ForegroundColor White
}

Write-Host ""
