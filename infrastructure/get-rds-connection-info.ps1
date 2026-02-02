# Get RDS Connection Details for PgAdmin4
# This script retrieves all necessary information to connect PgAdmin to AWS RDS

param(
    [Parameter(Mandatory=$false)]
    [string]$DBInstanceIdentifier = "",
    
    [string]$Region = "us-east-1",
    
    [string]$SecretName = ""
)

$ErrorActionPreference = "Stop"

Write-Host "`n=== RDS Connection Details for PgAdmin4 ===" -ForegroundColor Cyan
Write-Host ""

# List all RDS instances if none specified
if (-not $DBInstanceIdentifier) {
    Write-Host "Available RDS instances:" -ForegroundColor Yellow
    
    $instances = aws rds describe-db-instances `
        --region $Region `
        --query 'DBInstances[*].[DBInstanceIdentifier,Engine,DBInstanceStatus]' `
        --output json | ConvertFrom-Json
    
    if ($instances.Count -eq 0) {
        Write-Host "No RDS instances found in region $Region" -ForegroundColor Red
        exit 1
    }
    
    for ($i = 0; $i -lt $instances.Count; $i++) {
        Write-Host "  [$i] $($instances[$i][0]) ($($instances[$i][1])) - $($instances[$i][2])"
    }
    
    Write-Host ""
    $selection = Read-Host "Select instance number"
    $DBInstanceIdentifier = $instances[$selection][0]
}

Write-Host ""
Write-Host "Getting details for: $DBInstanceIdentifier" -ForegroundColor Yellow
Write-Host ""

# Get RDS instance details
$dbInfo = aws rds describe-db-instances `
    --db-instance-identifier $DBInstanceIdentifier `
    --region $Region `
    --query 'DBInstances[0]' `
    --output json | ConvertFrom-Json

if (-not $dbInfo) {
    Write-Host "Failed to get RDS instance details" -ForegroundColor Red
    exit 1
}

# Display connection information
Write-Host "========================================" -ForegroundColor Green
Write-Host "CONNECTION INFORMATION" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Endpoint:          $($dbInfo.Endpoint.Address)" -ForegroundColor Cyan
Write-Host "Port:              $($dbInfo.Endpoint.Port)" -ForegroundColor Cyan
Write-Host "Database Name:     $($dbInfo.DBName)" -ForegroundColor Cyan
Write-Host "Master Username:   $($dbInfo.MasterUsername)" -ForegroundColor Cyan
Write-Host "Engine:            $($dbInfo.Engine) $($dbInfo.EngineVersion)" -ForegroundColor Cyan
Write-Host "Status:            $($dbInfo.DBInstanceStatus)" -ForegroundColor Cyan
Write-Host ""

# Network information
Write-Host "========================================" -ForegroundColor Green
Write-Host "NETWORK INFORMATION" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "VPC:               $($dbInfo.DBSubnetGroup.VpcId)" -ForegroundColor Cyan
Write-Host "Publicly Accessible: $($dbInfo.PubliclyAccessible)" -ForegroundColor $(if ($dbInfo.PubliclyAccessible) { "Yellow" } else { "Red" })
Write-Host "Multi-AZ:          $($dbInfo.MultiAZ)" -ForegroundColor Cyan
Write-Host ""

# Security groups
Write-Host "Security Groups:" -ForegroundColor Yellow
foreach ($sg in $dbInfo.VpcSecurityGroups) {
    Write-Host "  - $($sg.VpcSecurityGroupId) ($($sg.Status))"
    
    # Get security group details
    $sgDetails = aws ec2 describe-security-groups `
        --group-ids $sg.VpcSecurityGroupId `
        --region $Region `
        --query 'SecurityGroups[0].IpPermissions' `
        --output json | ConvertFrom-Json
    
    Write-Host "    Inbound Rules:"
    foreach ($rule in $sgDetails) {
        if ($rule.FromPort -eq 5432) {
            foreach ($ipRange in $rule.IpRanges) {
                Write-Host "      Port 5432 from $($ipRange.CidrIp)" -ForegroundColor Green
            }
            foreach ($sgRef in $rule.UserIdGroupPairs) {
                Write-Host "      Port 5432 from $($sgRef.GroupId)" -ForegroundColor Green
            }
        }
    }
}
Write-Host ""

# Get password from Secrets Manager if secret name provided
if ($SecretName) {
    Write-Host "Retrieving password from Secrets Manager..." -ForegroundColor Yellow
    
    try {
        $secret = aws secretsmanager get-secret-value `
            --secret-id $SecretName `
            --region $Region `
            --query SecretString `
            --output text | ConvertFrom-Json
        
        Write-Host "Password:          $($secret.password)" -ForegroundColor Cyan
    } catch {
        Write-Host "Could not retrieve password from Secrets Manager" -ForegroundColor Red
    }
    Write-Host ""
}

# Check if your IP can connect
Write-Host "========================================" -ForegroundColor Green
Write-Host "CONNECTIVITY CHECK" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

$myIP = (Invoke-WebRequest -Uri "https://api.ipify.org" -UseBasicParsing).Content
Write-Host "Your Public IP:    $myIP" -ForegroundColor Cyan

if ($dbInfo.PubliclyAccessible) {
    Write-Host ""
    Write-Host "Testing connection..." -ForegroundColor Yellow
    
    try {
        $testResult = Test-NetConnection -ComputerName $dbInfo.Endpoint.Address -Port $dbInfo.Endpoint.Port -WarningAction SilentlyContinue
        
        if ($testResult.TcpTestSucceeded) {
            Write-Host "[OK] Port is reachable!" -ForegroundColor Green
        } else {
            Write-Host "[FAILED] Cannot reach port $($dbInfo.Endpoint.Port)" -ForegroundColor Red
            Write-Host ""
            Write-Host "Possible reasons:" -ForegroundColor Yellow
            Write-Host "  1. Your IP ($myIP) is not in the security group"
            Write-Host "  2. Network firewall blocking outbound connections"
            Write-Host "  3. RDS instance is not running"
        }
    } catch {
        Write-Host "[ERROR] Connection test failed: $_" -ForegroundColor Red
    }
} else {
    Write-Host "[INFO] RDS is not publicly accessible" -ForegroundColor Yellow
    Write-Host "You need to use one of these methods:" -ForegroundColor Yellow
    Write-Host "  1. SSH tunnel through bastion host"
    Write-Host "  2. AWS Systems Manager Session Manager"
    Write-Host "  3. AWS Client VPN"
}

Write-Host ""

# PgAdmin configuration
Write-Host "========================================" -ForegroundColor Green
Write-Host "PGADMIN4 CONFIGURATION" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

if ($dbInfo.PubliclyAccessible) {
    Write-Host "Connection Type:   Direct Connection" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "In PgAdmin4:" -ForegroundColor Yellow
    Write-Host "  1. Right-click 'Servers' -> Create -> Server"
    Write-Host "  2. General Tab:"
    Write-Host "       Name: Hub-HRMS $($dbInfo.DBInstanceIdentifier)"
    Write-Host "  3. Connection Tab:"
    Write-Host "       Host: $($dbInfo.Endpoint.Address)"
    Write-Host "       Port: $($dbInfo.Endpoint.Port)"
    Write-Host "       Maintenance database: $($dbInfo.DBName)"
    Write-Host "       Username: $($dbInfo.MasterUsername)"
    Write-Host "       Password: [Your password]"
    Write-Host "       Save password: [Check]"
    Write-Host "  4. SSL Tab:"
    Write-Host "       SSL mode: Require"
    Write-Host "  5. Click 'Save'"
} else {
    Write-Host "Connection Type:   SSH Tunnel Required" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "In PgAdmin4:" -ForegroundColor Yellow
    Write-Host "  1. First, start SSH tunnel (see tunnel command below)"
    Write-Host "  2. Right-click 'Servers' -> Create -> Server"
    Write-Host "  3. General Tab:"
    Write-Host "       Name: Hub-HRMS $($dbInfo.DBInstanceIdentifier) (via SSH)"
    Write-Host "  4. Connection Tab:"
    Write-Host "       Host: localhost"
    Write-Host "       Port: 5432"
    Write-Host "       Maintenance database: $($dbInfo.DBName)"
    Write-Host "       Username: $($dbInfo.MasterUsername)"
    Write-Host "       Password: [Your password]"
    Write-Host "  5. Click 'Save'"
}

Write-Host ""

# Save configuration to file
$configFile = "pgadmin-connection-$DBInstanceIdentifier.txt"
$configContent = @"
PgAdmin4 Connection Configuration
Generated: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

RDS Instance: $DBInstanceIdentifier
Region: $Region

CONNECTION DETAILS:
  Host: $($dbInfo.Endpoint.Address)
  Port: $($dbInfo.Endpoint.Port)
  Database: $($dbInfo.DBName)
  Username: $($dbInfo.MasterUsername)
  Engine: $($dbInfo.Engine) $($dbInfo.EngineVersion)
  Publicly Accessible: $($dbInfo.PubliclyAccessible)

NETWORK:
  VPC: $($dbInfo.DBSubnetGroup.VpcId)
  Security Groups: $(($dbInfo.VpcSecurityGroups | ForEach-Object { $_.VpcSecurityGroupId }) -join ', ')
  Your IP: $myIP

PGADMIN SETTINGS:
  SSL Mode: Require
  Save Password: Yes

"@

if (-not $dbInfo.PubliclyAccessible) {
    # Find bastion host
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "SSH TUNNEL SETUP" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    
    $bastions = aws ec2 describe-instances `
        --filters "Name=vpc-id,Values=$($dbInfo.DBSubnetGroup.VpcId)" "Name=instance-state-name,Values=running" `
        --region $Region `
        --query 'Reservations[*].Instances[*].[InstanceId,PublicIpAddress,PrivateIpAddress,Tags[?Key==`Name`].Value|[0]]' `
        --output json | ConvertFrom-Json
    
    if ($bastions.Count -gt 0) {
        Write-Host "Available bastion hosts / EC2 instances:" -ForegroundColor Yellow
        
        for ($i = 0; $i -lt $bastions.Count; $i++) {
            $instance = $bastions[$i][0]
            Write-Host "  [$i] $($instance[3]) - $($instance[0])"
            Write-Host "      Public IP: $($instance[1])"
            Write-Host "      Private IP: $($instance[2])"
        }
        
        Write-Host ""
        Write-Host "SSH Tunnel Command (replace with your key file):" -ForegroundColor Yellow
        Write-Host ""
        $bastionIP = $bastions[0][0][1]
        Write-Host "ssh -i your-key.pem -L 5432:$($dbInfo.Endpoint.Address):5432 ec2-user@$bastionIP -N" -ForegroundColor Cyan
        
        $configContent += @"
SSH TUNNEL:
  Bastion Host: $bastionIP
  Command: ssh -i your-key.pem -L 5432:$($dbInfo.Endpoint.Address):5432 ec2-user@$bastionIP -N

"@
    } else {
        Write-Host "No EC2 instances found in VPC. You need to create a bastion host." -ForegroundColor Red
    }
}

# Add security group command if needed
if ($dbInfo.PubliclyAccessible) {
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "SECURITY GROUP UPDATE" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "If you cannot connect, add your IP to the security group:" -ForegroundColor Yellow
    Write-Host ""
    
    $sgId = $dbInfo.VpcSecurityGroups[0].VpcSecurityGroupId
    Write-Host "aws ec2 authorize-security-group-ingress ``" -ForegroundColor Cyan
    Write-Host "  --group-id $sgId ``" -ForegroundColor Cyan
    Write-Host "  --protocol tcp ``" -ForegroundColor Cyan
    Write-Host "  --port 5432 ``" -ForegroundColor Cyan
    Write-Host "  --cidr $myIP/32 ``" -ForegroundColor Cyan
    Write-Host "  --region $Region" -ForegroundColor Cyan
    
    $configContent += @"
SECURITY GROUP UPDATE:
  aws ec2 authorize-security-group-ingress \
    --group-id $sgId \
    --protocol tcp \
    --port 5432 \
    --cidr $myIP/32 \
    --region $Region

"@
}

# Save to file
$configContent | Out-File -FilePath $configFile -Encoding utf8

Write-Host ""
Write-Host "Configuration saved to: $configFile" -ForegroundColor Green
Write-Host ""

# Option to add security group rule
if ($dbInfo.PubliclyAccessible) {
    $addRule = Read-Host "Would you like to add your IP to the security group now? (y/n)"
    
    if ($addRule -eq 'y') {
        Write-Host ""
        Write-Host "Adding security group rule..." -ForegroundColor Yellow
        
        $sgId = $dbInfo.VpcSecurityGroups[0].VpcSecurityGroupId
        
        try {
            aws ec2 authorize-security-group-ingress `
                --group-id $sgId `
                --protocol tcp `
                --port 5432 `
                --cidr "$myIP/32" `
                --region $Region 2>&1 | Out-Null
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "[OK] Security group rule added successfully!" -ForegroundColor Green
                Write-Host "You can now connect from PgAdmin4" -ForegroundColor Green
            }
        } catch {
            Write-Host "[ERROR] Failed to add security group rule" -ForegroundColor Red
            Write-Host "Run the command manually or check IAM permissions" -ForegroundColor Yellow
        }
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green