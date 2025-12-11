# Cleanup Failed Database Creation Resources
# Run this if create-database-ssm.ps1 fails

param(
    [Parameter(Mandatory=$false)]
    [string]$Region = 'us-east-1'
)

Write-Host "Cleaning up failed database creation resources..." -ForegroundColor Yellow
Write-Host ""

# Delete security group
$sg = aws ec2 describe-security-groups --filters "Name=group-name,Values=hrms-db-creator-sg" --region $Region --query 'SecurityGroups[0].GroupId' --output text 2>$null
if ($sg -and $sg -ne "None") {
    Write-Host "Deleting security group: $sg" -ForegroundColor Cyan
    aws ec2 delete-security-group --group-id $sg --region $Region 2>&1 | Out-Null
    Write-Host "OK" -ForegroundColor Green
}

# Find and delete IAM role
$roles = aws iam list-roles --query 'Roles[?starts_with(RoleName, `hrms-db-creator-role`)].RoleName' --output text
if ($roles) {
    foreach ($role in $roles -split '\s+') {
        if ($role) {
            Write-Host "Deleting IAM role: $role" -ForegroundColor Cyan
            
            # Remove from instance profile
            $null = aws iam remove-role-from-instance-profile --instance-profile-name $role --role-name $role --region $Region 2>&1
            
            # Delete instance profile
            $null = aws iam delete-instance-profile --instance-profile-name $role --region $Region 2>&1
            
            # Detach policies
            $null = aws iam detach-role-policy --role-name $role --policy-arn arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore --region $Region 2>&1
            
            # Delete role
            $null = aws iam delete-role --role-name $role --region $Region 2>&1
            
            Write-Host "OK" -ForegroundColor Green
        }
    }
}

# Terminate any running instances
$instances = aws ec2 describe-instances --filters "Name=tag:Name,Values=hrms-db-creator" "Name=instance-state-name,Values=running,pending" --region $Region --query 'Reservations[*].Instances[*].InstanceId' --output text
if ($instances) {
    foreach ($instance in $instances -split '\s+') {
        if ($instance) {
            Write-Host "Terminating instance: $instance" -ForegroundColor Cyan
            aws ec2 terminate-instances --instance-ids $instance --region $Region | Out-Null
            Write-Host "OK" -ForegroundColor Green
        }
    }
}

Write-Host ""
Write-Host "Cleanup complete!" -ForegroundColor Green