# PostgreSQL Permissions Fix Script (PowerShell)
# This grants necessary permissions to your database user

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "PostgreSQL Permissions Fix" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Check if psql is available
try {
    $null = Get-Command psql -ErrorAction Stop
} catch {
    Write-Host "Error: psql command not found. Please install PostgreSQL client." -ForegroundColor Red
    Write-Host "Download from: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    exit 1
}

# Get database details
$DB_HOST = Read-Host "Database host [localhost]"
if ([string]::IsNullOrWhiteSpace($DB_HOST)) { $DB_HOST = "localhost" }

$DB_PORT = Read-Host "Database port [5432]"
if ([string]::IsNullOrWhiteSpace($DB_PORT)) { $DB_PORT = "5432" }

$DB_NAME = Read-Host "Database name [hrapp]"
if ([string]::IsNullOrWhiteSpace($DB_NAME)) { $DB_NAME = "hrapp" }

$DB_USER = Read-Host "Your database user (the one having permission issues)"

Write-Host ""
Write-Host "This script will grant ALL permissions to user '$DB_USER' on database '$DB_NAME'" -ForegroundColor Yellow
Write-Host "You'll need to connect as a superuser (usually 'postgres')" -ForegroundColor Yellow
Write-Host ""

$SUPER_USER = Read-Host "Superuser name [postgres]"
if ([string]::IsNullOrWhiteSpace($SUPER_USER)) { $SUPER_USER = "postgres" }

$SecurePassword = Read-Host "Superuser password" -AsSecureString
$SUPER_PASSWORD = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecurePassword)
)

Write-Host ""
Write-Host "Granting permissions..." -ForegroundColor Yellow

$env:PGPASSWORD = $SUPER_PASSWORD

$permissionsSQL = @"
-- Grant all privileges on database
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

-- Grant all privileges on schema public
GRANT ALL ON SCHEMA public TO $DB_USER;

-- Grant all privileges on all tables in public schema
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;

-- Grant all privileges on all sequences in public schema
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;

-- Grant default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;

-- Make the user owner of the public schema (most permissive)
ALTER SCHEMA public OWNER TO $DB_USER;

-- Verify permissions
SELECT 
    grantee, 
    string_agg(privilege_type, ', ') as privileges
FROM information_schema.schema_privileges 
WHERE schema_name = 'public' AND grantee = '$DB_USER'
GROUP BY grantee;
"@

$result = $permissionsSQL | & psql -h $DB_HOST -p $DB_PORT -U $SUPER_USER -d $DB_NAME 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "[SUCCESS] Permissions granted successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "User '$DB_USER' now has full access to:" -ForegroundColor White
    Write-Host "  - Database: $DB_NAME" -ForegroundColor Gray
    Write-Host "  - Schema: public" -ForegroundColor Gray
    Write-Host "  - All tables and sequences" -ForegroundColor Gray
    Write-Host ""
    Write-Host "You can now run migrations:" -ForegroundColor Yellow
    Write-Host "  cd backend" -ForegroundColor Gray
    Write-Host "  go run cmd/main.go migrate" -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "[ERROR] Error granting permissions" -ForegroundColor Red
    Write-Host "Please check the error messages above" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Common issues:" -ForegroundColor Yellow
    Write-Host "  1. Wrong superuser password" -ForegroundColor Gray
    Write-Host "  2. Database doesn't exist" -ForegroundColor Gray
    Write-Host "  3. User doesn't exist yet" -ForegroundColor Gray
    Write-Host ""
    Write-Host "To create the user first:" -ForegroundColor Yellow
    Write-Host "  psql -h $DB_HOST -U $SUPER_USER -d $DB_NAME" -ForegroundColor Gray
    Write-Host "  CREATE USER $DB_USER WITH PASSWORD 'your_password';" -ForegroundColor Gray
}

# Clear password from environment
$env:PGPASSWORD = $null
$SUPER_PASSWORD = $null
