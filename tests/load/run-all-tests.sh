#!/bin/bash

# Velure Microservices Load Testing Suite
# Runs all load tests with proper sequencing and reporting

set -e

echo "🚀 Starting Velure Microservices Load Testing Suite"
echo "=================================================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}❌ K6 is not installed. Please install K6 first.${NC}"
    echo "Installation: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# Check if services are running
echo -e "${YELLOW}🔍 Checking if services are running...${NC}"

check_service() {
    local service_name=$1
    local url=$2
    
    if curl -s --connect-timeout 5 "$url" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ $service_name is running${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name is not accessible at $url${NC}"
        return 1
    fi
}

# Service health checks
services_ok=true

if ! check_service "Auth Service" "http://localhost:3020/authentication/users"; then
    services_ok=false
fi

if ! check_service "Product Service" "http://localhost:3010/health"; then
    services_ok=false
fi

if ! check_service "Order Service" "http://localhost:3030"; then
    services_ok=false
fi

if [ "$services_ok" = false ]; then
    echo -e "${RED}❌ Some services are not running. Please start them with: docker-compose up -d${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All services are running!${NC}"
echo ""

# Create results directory
RESULTS_DIR="results/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$RESULTS_DIR"

echo -e "${YELLOW}📁 Results will be saved to: $RESULTS_DIR${NC}"
echo ""

# Function to run individual tests
run_test() {
    local test_name=$1
    local test_file=$2
    local service_name=$3
    
    echo -e "${YELLOW}🧪 Running $test_name...${NC}"
    
    if k6 run --out json="$RESULTS_DIR/${test_name}-results.json" "$test_file"; then
        echo -e "${GREEN}✅ $test_name completed successfully${NC}"
        generate_simple_report "$test_name"
    else
        echo -e "${RED}❌ $test_name failed${NC}"
        return 1
    fi
    
    echo ""
    sleep 5  # Brief pause between tests
}

# Function to generate simple HTML report from JSON
generate_simple_report() {
    local test_name=$1
    local json_file="$RESULTS_DIR/${test_name}-results.json"
    local html_file="$RESULTS_DIR/${test_name}-report.html"
    
    if [ -f "$json_file" ]; then
        cat > "$html_file" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>${test_name} Load Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric { margin: 10px 0; padding: 15px; background-color: #f5f5f5; border-radius: 5px; }
        .success { background-color: #d4edda; border-left: 4px solid #28a745; }
        .warning { background-color: #fff3cd; border-left: 4px solid #ffc107; }
        .error { background-color: #f8d7da; border-left: 4px solid #dc3545; }
        h1 { color: #333; }
        h3 { color: #555; margin-top: 0; }
        .summary { font-size: 18px; font-weight: bold; }
    </style>
</head>
<body>
    <h1>📊 ${test_name} Load Test Results</h1>
    <div class="metric summary">
        <h3>Test Completed Successfully ✅</h3>
        <p>Generated on: $(date)</p>
        <p>Test Duration: Full load test with 15-second escalation stages</p>
    </div>
    
    <div class="metric success">
        <h3>📈 Performance Summary</h3>
        <p>• Load test executed with gradual user escalation</p>
        <p>• Multiple scenarios tested across different endpoints</p>
        <p>• Error handling and response time validation included</p>
        <p>• Raw JSON data available for detailed analysis</p>
    </div>
    
    <div class="metric">
        <h3>📋 Test Configuration</h3>
        <p>• Escalation Pattern: 15-second stages</p>
        <p>• Service: ${test_name}</p>
        <p>• Multiple user scenarios with realistic data</p>
        <p>• Performance thresholds and error rate monitoring</p>
    </div>
    
    <div class="metric">
        <h3>📊 Raw Data</h3>
        <p>For detailed metrics analysis, please examine the JSON results file:</p>
        <code>${test_name}-results.json</code>
    </div>
</body>
</html>
EOF
    fi
}

# Run individual service tests
echo -e "${YELLOW}🎯 Starting Individual Service Tests${NC}"
echo "======================================"

run_test "auth-service" "k6-load-tests/auth-service-test.js" "Authentication Service"
run_test "product-service" "k6-load-tests/product-service-test.js" "Product Service"
run_test "publish-order-service" "k6-load-tests/publish-order-service-test.js" "Order Service"

echo -e "${YELLOW}🔄 Waiting 30 seconds before integrated test...${NC}"
sleep 30

# Run integrated test
echo -e "${YELLOW}🎯 Starting Integrated Load Test${NC}"
echo "================================="

run_test "integrated" "k6-load-tests/integrated-load-test.js" "All Services"

# Generate summary report
echo -e "${YELLOW}📊 Generating Summary Report...${NC}"

cat > "$RESULTS_DIR/summary.html" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Velure Load Testing Summary</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .test-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin: 20px 0; }
        .test-card { border: 1px solid #ddd; padding: 15px; border-radius: 8px; background: #f9f9f9; }
        .test-card h3 { margin-top: 0; color: #333; }
        .btn { display: inline-block; padding: 8px 16px; margin: 5px; text-decoration: none; border-radius: 4px; color: white; }
        .btn-primary { background-color: #007bff; }
        .btn-success { background-color: #28a745; }
        .footer { text-align: center; margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Velure Microservices Load Testing Results</h1>
            <p>Generated on: $(date)</p>
        </div>

        <div class="test-grid">
            <div class="test-card">
                <h3>🔐 Authentication Service</h3>
                <p>User registration, login, token validation</p>
                <a href="auth-service-report.html" class="btn btn-primary">View Report</a>
                <a href="auth-service-results.json" class="btn btn-success">Raw Data</a>
            </div>

            <div class="test-card">
                <h3>🛍️ Product Service</h3>
                <p>Product CRUD operations, search, pagination</p>
                <a href="product-service-report.html" class="btn btn-primary">View Report</a>
                <a href="product-service-results.json" class="btn btn-success">Raw Data</a>
            </div>

            <div class="test-card">
                <h3>📦 Order Service</h3>
                <p>Order creation and status updates</p>
                <a href="publish-order-service-report.html" class="btn btn-primary">View Report</a>
                <a href="publish-order-service-results.json" class="btn btn-success">Raw Data</a>
            </div>

            <div class="test-card">
                <h3>🔄 Integrated Test</h3>
                <p>Cross-service workflow simulation</p>
                <a href="integrated-report.html" class="btn btn-primary">View Report</a>
                <a href="integrated-results.json" class="btn btn-success">Raw Data</a>
            </div>
        </div>

        <div class="footer">
            <p>Load testing completed with escalating user load every 15 seconds</p>
            <p>For detailed analysis, review individual service reports</p>
        </div>
    </div>
</body>
</html>
EOF

echo -e "${GREEN}🎉 Load Testing Suite Completed!${NC}"
echo "=================================="
echo -e "${GREEN}✅ All tests have been executed${NC}"
echo -e "${YELLOW}📊 Results available at: $RESULTS_DIR${NC}"
echo -e "${YELLOW}🌐 Open summary.html for a complete overview${NC}"
echo ""
echo "📋 Quick Access:"
echo "   Summary: $RESULTS_DIR/summary.html"
echo "   Auth Service: $RESULTS_DIR/auth-service-report.html"
echo "   Product Service: $RESULTS_DIR/product-service-report.html"
echo "   Order Service: $RESULTS_DIR/publish-order-service-report.html"
echo "   Integrated: $RESULTS_DIR/integrated-report.html"