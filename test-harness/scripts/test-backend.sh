#!/bin/bash

# Backend Test Runner for hub-hrms
# Runs all backend tests with coverage reporting

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🧪 Running backend test suite...${NC}"

# Parse arguments
COVERAGE=false
VERBOSE=false
INTEGRATION=false
E2E=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -i|--integration)
            INTEGRATION=true
            shift
            ;;
        -e|--e2e)
            E2E=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create coverage directory
mkdir -p coverage

# Unit Tests
echo -e "${YELLOW}Running unit tests...${NC}"
if [ "$COVERAGE" = true ]; then
    if [ "$VERBOSE" = true ]; then
        go test -v -race -coverprofile=coverage/unit.out -covermode=atomic ./... -short
    else
        go test -race -coverprofile=coverage/unit.out -covermode=atomic ./... -short
    fi
else
    if [ "$VERBOSE" = true ]; then
        go test -v -race ./... -short
    else
        go test -race ./... -short
    fi
fi

# Integration Tests
if [ "$INTEGRATION" = true ]; then
    echo -e "${YELLOW}Running integration tests...${NC}"
    if [ "$COVERAGE" = true ]; then
        go test -v -race -coverprofile=coverage/integration.out -covermode=atomic ./tests/integration/...
    else
        go test -v -race ./tests/integration/...
    fi
fi

# E2E Tests
if [ "$E2E" = true ]; then
    echo -e "${YELLOW}Running E2E tests...${NC}"
    go test -v -race ./tests/e2e/...
fi

# Generate coverage report
if [ "$COVERAGE" = true ]; then
    echo -e "${YELLOW}Generating coverage report...${NC}"
    go tool cover -html=coverage/unit.out -o coverage/coverage.html
    go tool cover -func=coverage/unit.out | grep total:
    echo -e "${GREEN}Coverage report generated: coverage/coverage.html${NC}"
fi

echo -e "${GREEN}✅ Backend tests complete!${NC}"
