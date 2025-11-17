#!/bin/bash

# Velure User Journey Load Test Runner
# This script runs a comprehensive load test of the complete user journey

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-https://velure.local}"
WARMUP_DURATION="${WARMUP_DURATION:-10s}"
TEST_DURATION="${TEST_DURATION:-1m}"
COOLDOWN_DURATION="${COOLDOWN_DURATION:-10s}"
TARGET_VUS="${TARGET_VUS:-20}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Velure User Journey Load Test${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${YELLOW}Configuration:${NC}"
echo "  Base URL: $BASE_URL"
echo "  Warmup: $WARMUP_DURATION"
echo "  Test Duration: $TEST_DURATION"
echo "  Cooldown: $COOLDOWN_DURATION"
echo "  Target VUs: $TARGET_VUS"
echo ""
echo -e "${YELLOW}Test Journey:${NC}"
echo "  1. Register New Account"
echo "  2. Login to Account"
echo "  3. Browse Product Catalog"
echo "  4. Search Products"
echo "  5. Create Purchase Order"
echo "  6. View Order History"
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}Error: k6 is not installed${NC}"
    echo "Please install k6 from https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# Check if services are running
echo -e "${YELLOW}Checking if services are accessible...${NC}"
if ! curl -k -s -o /dev/null -w "%{http_code}" "$BASE_URL" | grep -q "200\|301\|302"; then
    echo -e "${RED}Warning: Cannot reach $BASE_URL${NC}"
    echo "Make sure the services are running (make dev-services)"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Run the test
echo -e "${GREEN}Starting load test...${NC}"
echo ""

k6 run \
  --out json=user-journey-results.json \
  -e BASE_URL="$BASE_URL" \
  -e WARMUP_DURATION="$WARMUP_DURATION" \
  -e TEST_DURATION="$TEST_DURATION" \
  -e COOLDOWN_DURATION="$COOLDOWN_DURATION" \
  -e TARGET_VUS="$TARGET_VUS" \
  "$(dirname "$0")/user-journey-test.js"

echo ""
echo -e "${GREEN}Test completed!${NC}"
echo ""
echo -e "${YELLOW}Results saved to: user-journey-results.json${NC}"
echo -e "${YELLOW}View metrics in Grafana: http://localhost:3000${NC}"
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
