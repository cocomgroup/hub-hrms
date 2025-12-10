#!/bin/bash

# Fix PostgreSQL Permissions Script
# This grants necessary permissions to your database user

echo "=========================================="
echo "PostgreSQL Permissions Fix"
echo "=========================================="
echo ""

# Get database details
read -p "Database host [localhost]: " DB_HOST
DB_HOST=${DB_HOST:-localhost}

read -p "Database port [5432]: " DB_PORT
DB_PORT=${DB_PORT:-5432}

read -p "Database name [hrapp]: " DB_NAME
DB_NAME=${DB_NAME:-hrapp}

read -p "Your database user (the one having permission issues): " DB_USER

echo ""
echo "This script will grant ALL permissions to user '$DB_USER' on database '$DB_NAME'"
echo "You'll need to connect as a superuser (usually 'postgres')"
echo ""

read -p "Superuser name [postgres]: " SUPER_USER
SUPER_USER=${SUPER_USER:-postgres}

read -sp "Superuser password: " SUPER_PASSWORD
echo ""
echo ""

export PGPASSWORD="$SUPER_PASSWORD"

echo "Granting permissions..."

psql -h "$DB_HOST" -p "$DB_PORT" -U "$SUPER_USER" -d "$DB_NAME" << EOF

-- Grant all privileges on database
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;

-- Connect to the database
\c $DB_NAME

-- Grant all privileges on schema public
GRANT ALL ON SCHEMA public TO $DB_USER;

-- Grant all privileges on all tables in public schema
GRANT ALL PRIVILEGES ON ALL TABLES IN public SCHEMA TO $DB_USER;

-- Grant all privileges on all sequences in public schema
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN public SCHEMA TO $DB_USER;

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

EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Permissions granted successfully!"
    echo ""
    echo "User '$DB_USER' now has full access to:"
    echo "  - Database: $DB_NAME"
    echo "  - Schema: public"
    echo "  - All tables and sequences"
    echo ""
    echo "You can now run migrations:"
    echo "  cd backend"
    echo "  go run cmd/main.go migrate"
else
    echo ""
    echo "❌ Error granting permissions"
    echo "Please check the error messages above"
fi

unset PGPASSWORD
