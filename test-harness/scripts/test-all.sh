#!/bin/bash

# Master Test Runner for hub-hrms
# Runs all test types: unit, integration, E2E, frontend

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Parse arguments
RUN_UNIT=true
RUN_INTEGRATION=false
RUN_E2E=false
RUN_FRONTEND=true
COVERAGE=false
PARALLEL=false

usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -a, --all           Run all tests (unit, integration, e2e, frontend)"
    echo "  -u, --unit          Run unit tests only (default)"
    echo "  -i, --integration   Run integration tests"
    echo "  -e, --e2e           Run E2E tests"
    echo "  -f, --frontend      Run frontend tests (default)"
    echo "  -c, --coverage      Generate coverage reports"
    echo "  -p, --parallel      Run tests in parallel where possible"
    echo "  -h, --help          Show this help message"
    exit 1
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--all)
            RUN_UNIT=true
            RUN_INTEGRATION=true
            RUN_E2E=true
            RUN_FRONTEND=true
            shift
            ;;
        -u|--unit)
            RUN_UNIT=true
            RUN_INTEGRATION=false
            RUN_E2E=false
            RUN_FRONTEND=false
            shift
            ;;
        -i|--integration)
            RUN_INTEGRATION=true
            shift
            ;;
        -e|--e2e)
            RUN_E2E=true
            shift
            ;;
        -f|--frontend)
            RUN_FRONTEND=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     hub-hrms Test Suite                    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Track start time
START_TIME=$(date +%s)

# Backend Unit Tests
if [ "$RUN_UNIT" = true ]; then
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}  Running Backend Unit Tests${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    if [ "$COVERAGE" = true ]; then
        ./scripts/test-backend.sh --coverage
    else
        ./scripts/test-backend.sh
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Backend unit tests passed${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ Backend unit tests failed${NC}"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    echo ""
fi

# Backend Integration Tests
if [ "$RUN_INTEGRATION" = true ]; then
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}  Running Backend Integration Tests${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # Start test database
    echo "Starting test database..."
    docker-compose -f docker-compose.test.yml up -d db
    sleep 5
    
    if [ "$COVERAGE" = true ]; then
        ./scripts/test-backend.sh --integration --coverage
    else
        ./scripts/test-backend.sh --integration
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Integration tests passed${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ Integration tests failed${NC}"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # Cleanup test database
    docker-compose -f docker-compose.test.yml down
    echo ""
fi

# Frontend Tests
if [ "$RUN_FRONTEND" = true ]; then
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}  Running Frontend Tests${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    if [ "$COVERAGE" = true ]; then
        ./scripts/test-frontend.sh --coverage
    else
        ./scripts/test-frontend.sh
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Frontend tests passed${NC}"
        ((PASSED_TESTS++))
    else
        echo -e "${RED}❌ Frontend tests failed${NC}"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    echo ""
fi

# E2E Tests
if [ "$RUN_E2E" = true ]; then
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}  Running E2E Tests${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # Start the application
    echo "Starting application for E2E tests..."
    docker-compose -f docker-compose.test.yml up -d
    sleep 10
    
    # Wait for application to be ready
    echo "Waiting for application to be ready..."
    timeout 60 bash -c 'until curl -f http://localhost:5173 > /dev/null 2>&1; do sleep 2; done'
    
    if [ $? -eq 0 ]; then
        # Run E2E tests
        ./scripts/test-backend.sh --e2e
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ E2E tests passed${NC}"
            ((PASSED_TESTS++))
        else
            echo -e "${RED}❌ E2E tests failed${NC}"
            ((FAILED_TESTS++))
        fi
    else
        echo -e "${RED}❌ Application failed to start${NC}"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # Cleanup
    docker-compose -f docker-compose.test.yml down
    echo ""
fi

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Print summary
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Test Summary                           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Total test suites: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"
echo -e "Duration: ${DURATION}s"
echo ""

# Generate coverage report if requested
if [ "$COVERAGE" = true ]; then
    echo -e "${YELLOW}Generating combined coverage report...${NC}"
    
    if [ -d "coverage" ]; then
        echo "Backend coverage reports:"
        if [ -f "coverage/unit.out" ]; then
            go tool cover -func=coverage/unit.out | tail -1
        fi
        
        if [ -f "coverage/integration.out" ]; then
            go tool cover -func=coverage/integration.out | tail -1
        fi
    fi
    
    if [ -d "frontend/coverage" ]; then
        echo ""
        echo "Frontend coverage reports:"
        echo "See frontend/coverage/index.html"
    fi
    
    echo ""
fi

# Exit with appropriate code
if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     All Tests Passed! ✅                   ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════╝${NC}"
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║     Some Tests Failed ❌                   ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════╝${NC}"
    exit 1
fi
