#!/bin/bash

# Frontend Test Runner for hub-hrms
# Runs all frontend tests with coverage and watch mode

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🧪 Running frontend test suite...${NC}"

# Check if frontend directory exists
if [ ! -d "frontend" ]; then
    echo -e "${RED}Error: frontend directory not found${NC}"
    exit 1
fi

cd frontend

# Parse arguments
COVERAGE=false
WATCH=false
UI=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -w|--watch)
            WATCH=true
            shift
            ;;
        -u|--ui)
            UI=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Build test command
TEST_CMD="npm run test"

if [ "$WATCH" = true ]; then
    echo -e "${YELLOW}Running tests in watch mode...${NC}"
    npm run test:watch
elif [ "$UI" = true ]; then
    echo -e "${YELLOW}Starting test UI...${NC}"
    npm run test:ui
elif [ "$COVERAGE" = true ]; then
    echo -e "${YELLOW}Running tests with coverage...${NC}"
    npm run test:coverage
else
    echo -e "${YELLOW}Running all tests...${NC}"
    npm run test
fi

cd ..

echo -e "${GREEN}✅ Frontend tests complete!${NC}"
