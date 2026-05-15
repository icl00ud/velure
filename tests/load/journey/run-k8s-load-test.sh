#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Velure K8s User Journey Load Test                  ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}❌ Error: k6 is not installed${NC}"
    echo ""
    echo "Please install k6:"
    echo "  macOS:   brew install k6"
    echo "  Linux:   sudo gpg -k"
    echo "           sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69"
    echo "           echo \"deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main\" | sudo tee /etc/apt/sources.list.d/k6.list"
    echo "           sudo apt-get update"
    echo "           sudo apt-get install k6"
    echo "  Windows: choco install k6"
    echo ""
    echo "Or visit: https://k6.io/docs/get-started/installation/"
    exit 1
fi

# Determine which environment file to use
ENV_FILE=""
if [ "$1" == "local" ]; then
    ENV_FILE="${SCRIPT_DIR}/.env.local"
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${YELLOW}⚠️  Warning: .env.local not found${NC}"
        echo -e "Creating from .env.local.example..."
        if [ -f "${SCRIPT_DIR}/.env.local.example" ]; then
            cp "${SCRIPT_DIR}/.env.local.example" "$ENV_FILE"
            echo -e "${GREEN}✓ Created .env.local${NC}"
        else
            echo -e "${RED}❌ Error: .env.local.example not found${NC}"
            exit 1
        fi
    fi
    echo -e "${GREEN}📍 Using local environment configuration${NC}"
elif [ "$1" == "k8s" ] || [ "$1" == "kubernetes" ] || [ -z "$1" ]; then
    ENV_FILE="${SCRIPT_DIR}/.env.k8s"
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${YELLOW}⚠️  Warning: .env.k8s not found${NC}"
        echo -e "Creating from .env.k8s.example..."
        if [ -f "${SCRIPT_DIR}/.env.k8s.example" ]; then
            cp "${SCRIPT_DIR}/.env.k8s.example" "$ENV_FILE"
            echo -e "${GREEN}✓ Created .env.k8s${NC}"
            echo ""
            echo -e "${YELLOW}⚠️  IMPORTANT: Please edit .env.k8s and set your BASE_URL${NC}"
            echo ""
            exit 1
        else
            echo -e "${RED}❌ Error: .env.k8s.example not found${NC}"
            exit 1
        fi
    fi
    echo -e "${GREEN}📍 Using Kubernetes environment configuration${NC}"
else
    echo -e "${RED}❌ Error: Invalid argument${NC}"
    echo ""
    echo "Usage: $0 [local|k8s|kubernetes]"
    echo ""
    echo "Examples:"
    echo "  $0           # Run against Kubernetes (default)"
    echo "  $0 k8s       # Run against Kubernetes"
    echo "  $0 local     # Run against local development"
    exit 1
fi

# Load environment variables
echo -e "${BLUE}📂 Loading environment from: $ENV_FILE${NC}"
export $(grep -v '^#' "$ENV_FILE" | xargs)

# Validate BASE_URL is set
if [ -z "$BASE_URL" ] || [ "$BASE_URL" == "https://your-k8s-ingress-url.com" ]; then
    echo -e "${RED}❌ Error: BASE_URL is not configured${NC}"
    echo ""
    echo "Please edit $ENV_FILE and set the BASE_URL variable"
    echo "Example: BASE_URL=https://velure.yourdomain.com"
    exit 1
fi

echo -e "${GREEN}🎯 Target URL: $BASE_URL${NC}"
echo ""

# Create results directory
RESULTS_DIR="${SCRIPT_DIR}/results"
mkdir -p "$RESULTS_DIR"

# Generate timestamp for results file
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_FILE="${RESULTS_DIR}/user-journey-${TIMESTAMP}.json"

echo -e "${BLUE}🚀 Starting load test...${NC}"
echo ""

# Run k6 test
k6 run \
    --env BASE_URL="$BASE_URL" \
    --out json="$RESULTS_FILE" \
    "${SCRIPT_DIR}/user-journey-k8s.js"

TEST_EXIT_CODE=$?

echo ""
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                 ✅ Load Test Completed Successfully           ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}📊 Results saved to: $RESULTS_FILE${NC}"
else
    echo -e "${RED}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                   ❌ Load Test Failed                         ║${NC}"
    echo -e "${RED}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}📊 Partial results may be available at: $RESULTS_FILE${NC}"
fi

echo ""
echo -e "${BLUE}💡 Tips:${NC}"
echo "  - View detailed metrics in the k6 output above"
echo "  - Analyze results with: k6 inspect $RESULTS_FILE"
echo "  - For cloud testing, consider k6 Cloud: https://k6.io/cloud"
echo ""

exit $TEST_EXIT_CODE
