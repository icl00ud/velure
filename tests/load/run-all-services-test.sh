#!/bin/bash

# Velure - Run All Services Load Test
# Executes comprehensive load test against all microservices

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load environment variables
if [ -f "$SCRIPT_DIR/.env.local" ]; then
    export $(cat "$SCRIPT_DIR/.env.local" | grep -v '^#' | xargs)
    echo -e "${GREEN}✓${NC} Loaded environment from .env.local"
else
    echo -e "${RED}✗${NC} .env.local not found, using defaults"
fi

# Function to check if services are healthy
check_services() {
    echo -e "\n${BLUE}Checking service health...${NC}"

    local services=("auth:3020" "product:3010" "order:3030")
    local all_healthy=true

    for service in "${services[@]}"; do
        IFS=':' read -r name port <<< "$service"
        if curl -s -f "http://localhost:${port}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} ${name}-service is healthy"
        else
            echo -e "${RED}✗${NC} ${name}-service is not responding"
            all_healthy=false
        fi
    done

    if [ "$all_healthy" = false ]; then
        echo -e "\n${YELLOW}Warning:${NC} Some services are not healthy. Continue anyway? (y/n)"
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            echo -e "${RED}Test aborted${NC}"
            exit 1
        fi
    fi

    echo ""
}

# Function to check if k6 is installed
check_k6() {
    if ! command -v k6 &> /dev/null; then
        echo -e "${RED}✗${NC} k6 is not installed"
        echo -e "\n${YELLOW}Install k6:${NC}"
        echo "  macOS:   brew install k6"
        echo "  Linux:   See https://k6.io/docs/getting-started/installation/"
        echo "  Windows: choco install k6"
        exit 1
    fi

    echo -e "${GREEN}✓${NC} k6 version: $(k6 version)"
}

# Function to display help
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Run comprehensive load test against all Velure microservices"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -s, --skip-checks       Skip service health checks"
    echo "  -d, --duration DURATION Set test duration (default: from .env.local)"
    echo "  -w, --warmup DURATION   Set warmup duration (default: from .env.local)"
    echo "  -o, --out FILE          Save results to file"
    echo "  -q, --quiet             Minimal output"
    echo "  --no-summary            Don't show summary at the end"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run with default settings"
    echo "  $0 -d 5m -w 1m                       # 5 min test with 1 min warmup"
    echo "  $0 -o results.json                   # Save results to file"
    echo "  $0 -s -q                             # Skip checks, quiet mode"
    echo ""
}

# Parse command line arguments
SKIP_CHECKS=false
TEST_DURATION=""
WARMUP_DURATION=""
OUTPUT_FILE=""
QUIET=false
NO_SUMMARY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -s|--skip-checks)
            SKIP_CHECKS=true
            shift
            ;;
        -d|--duration)
            TEST_DURATION="$2"
            shift 2
            ;;
        -w|--warmup)
            WARMUP_DURATION="$2"
            shift 2
            ;;
        -o|--out)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        --no-summary)
            NO_SUMMARY=true
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Velure - All Services Load Test${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Check k6
check_k6

# Check services health
if [ "$SKIP_CHECKS" = false ]; then
    check_services
fi

# Build k6 command
K6_CMD="k6 run"

# Add environment variables
if [ -n "$TEST_DURATION" ]; then
    K6_CMD="$K6_CMD -e TEST_DURATION=$TEST_DURATION"
fi

if [ -n "$WARMUP_DURATION" ]; then
    K6_CMD="$K6_CMD -e WARMUP_DURATION=$WARMUP_DURATION"
fi

# Add output file
if [ -n "$OUTPUT_FILE" ]; then
    K6_CMD="$K6_CMD --out json=$OUTPUT_FILE"
fi

# Add quiet mode
if [ "$QUIET" = true ]; then
    K6_CMD="$K6_CMD --quiet"
fi

# Add no summary
if [ "$NO_SUMMARY" = true ]; then
    K6_CMD="$K6_CMD --no-summary"
fi

# Add test file
K6_CMD="$K6_CMD $SCRIPT_DIR/all-services-load-test.js"

# Display test configuration
echo -e "${BLUE}Test Configuration:${NC}"
echo -e "  Base URL:     ${BASE_URL:-https://velure.local}"
echo -e "  Auth URL:     ${AUTH_URL:-$BASE_URL/api/auth}"
echo -e "  Product URL:  ${PRODUCT_URL:-$BASE_URL/api/product}"
echo -e "  Order URL:    ${ORDER_URL:-$BASE_URL/api/order}"
echo -e "  Warmup:       ${WARMUP_DURATION:-${WARMUP_DURATION:-30s}}"
echo -e "  Test Time:    ${TEST_DURATION:-${TEST_DURATION:-2m}}"
echo -e "  Cooldown:     ${COOLDOWN_DURATION:-30s}"
if [ -n "$OUTPUT_FILE" ]; then
    echo -e "  Output:       $OUTPUT_FILE"
fi
echo ""

# Confirm before running
echo -e "${YELLOW}Press Enter to start the load test, or Ctrl+C to cancel${NC}"
read -r

# Run the test
echo -e "\n${GREEN}Starting load test...${NC}\n"
echo -e "${BLUE}Command:${NC} $K6_CMD\n"

eval $K6_CMD
EXIT_CODE=$?

# Summary
echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Load Test Completed Successfully${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}Next Steps:${NC}"
    echo -e "  1. Check Grafana dashboards: ${GREEN}http://localhost:3000${NC}"
    echo -e "  2. Review Prometheus metrics: ${GREEN}http://localhost:9090${NC}"
    if [ -n "$OUTPUT_FILE" ]; then
        echo -e "  3. Analyze results: ${GREEN}$OUTPUT_FILE${NC}"
    fi
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}  Load Test Failed${NC}"
    echo -e "${RED}========================================${NC}"
    echo -e "\n${YELLOW}Check the logs above for errors${NC}"
fi

exit $EXIT_CODE
