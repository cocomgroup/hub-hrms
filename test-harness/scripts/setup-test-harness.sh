#!/bin/bash

# Test Harness Setup Script for hub-hrms
# Sets up testing infrastructure for both frontend and backend

set -e

echo "🧪 Setting up test harness for hub-hrms..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: go.mod not found. Please run this script from the project root.${NC}"
    exit 1
fi

# Backend Testing Setup
echo -e "${BLUE}Setting up backend testing infrastructure...${NC}"

# Install Go testing dependencies
go get -u github.com/stretchr/testify/assert
go get -u github.com/stretchr/testify/mock
go get -u github.com/stretchr/testify/suite
go get -u github.com/DATA-DOG/go-sqlmock
go get -u github.com/golang/mock/gomock
go get -u github.com/jarcoal/httpmock

# Create test directories
mkdir -p tests/unit
mkdir -p tests/integration
mkdir -p tests/e2e
mkdir -p tests/fixtures
mkdir -p tests/mocks

# Frontend Testing Setup
echo -e "${BLUE}Setting up frontend testing infrastructure...${NC}"

# Check if frontend directory exists
if [ -d "frontend" ]; then
    cd frontend
    
    # Install testing dependencies
    npm install --save-dev \
        @testing-library/svelte \
        @testing-library/jest-dom \
        @testing-library/user-event \
        vitest \
        jsdom \
        @vitest/ui \
        @vitest/coverage-v8 \
        svelte-testing-library \
        msw
    
    cd ..
else
    echo -e "${RED}Warning: frontend directory not found${NC}"
fi

echo -e "${GREEN}✅ Test harness setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Run backend tests: ./scripts/test-backend.sh"
echo "2. Run frontend tests: ./scripts/test-frontend.sh"
echo "3. Run all tests: ./scripts/test-all.sh"
