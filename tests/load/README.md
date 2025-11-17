# Velure Load Testing

Comprehensive load testing suite for all Velure microservices using k6.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Test Scripts](#test-scripts)
- [Configuration](#configuration)
- [Running Tests](#running-tests)
- [Monitoring](#monitoring)
- [Interpreting Results](#interpreting-results)

## 🎯 Prerequisites

### 1. Install k6

```bash
# macOS
brew install k6

# Linux (Debian/Ubuntu)
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
  --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Windows
choco install k6
```

### 2. Verify Installation

```bash
k6 version
```

### 3. Start Velure Services

```bash
# From repository root
make dev              # Start infrastructure
make dev-services     # Start all services (or use docker-compose)
```

## 🚀 Quick Start

```bash
# Navigate to load tests directory
cd tests/load

# Run all services load test
./run-all-services-test.sh

# Or run directly with k6
k6 run all-services-load-test.js
```

## 📝 Test Scripts

### all-services-load-test.js

Comprehensive load test that tests all microservices simultaneously:

- **Auth Service**: Registration, login, token validation
- **Product Service**: Product listing, search, details
- **Publish Order Service**: Order creation, order listing
- **Process Order Service**: Tested indirectly via RabbitMQ message processing

#### Test Scenarios

1. **auth_load**: Tests authentication endpoints
   - 0 → 20 VUs (warmup) → 100 VUs → 200 VUs (peak) → 0 VUs (cooldown)
   - Target: p95 < 500ms, error rate < 10%

2. **product_load**: Tests product catalog
   - 0 → 30 VUs (warmup) → 200 VUs → 400 VUs (peak) → 0 VUs (cooldown)
   - Target: p95 < 1000ms, error rate < 5%

3. **order_publish_load**: Tests order creation
   - 0 → 50 VUs (warmup) → 300 VUs → 500 VUs (peak) → 0 VUs (cooldown)
   - Target: p95 < 2000ms, error rate < 10%

4. **user_journey**: Complete user flow
   - 0 → 10 VUs (warmup) → 50 VUs → 100 VUs (peak) → 0 VUs (cooldown)
   - Tests: Register → Browse → Search → Order → View Orders

#### Peak Load

- **Total Virtual Users**: ~800 VUs at peak
- **Duration**: ~5-6 minutes (configurable)
- **Requests/sec**: ~2000-3000 rps (varies by scenario)

## ⚙️ Configuration

### Environment Variables

Create or modify `.env.local`:

```bash
# Service URLs (via Caddy reverse proxy)
BASE_URL=https://velure.local
AUTH_URL=https://velure.local/api/auth
PRODUCT_URL=https://velure.local/api/product
ORDER_URL=https://velure.local/api/order

# Test duration
WARMUP_DURATION=30s
TEST_DURATION=2m
COOLDOWN_DURATION=30s
```

### For Local Docker Compose

```bash
BASE_URL=http://localhost
AUTH_URL=http://localhost/api/auth
PRODUCT_URL=http://localhost/api/product
ORDER_URL=http://localhost/api/order
```

### For Kubernetes/EKS

```bash
# Use .env.eks.example as template
BASE_URL=https://your-domain.com
AUTH_URL=https://your-domain.com/api/auth
PRODUCT_URL=https://your-domain.com/api/product
ORDER_URL=https://your-domain.com/api/order
```

## 🏃 Running Tests

### Using the Helper Script (Recommended)

```bash
# Basic run
./run-all-services-test.sh

# Custom duration
./run-all-services-test.sh -d 5m -w 1m

# Save results to file
./run-all-services-test.sh -o results.json

# Skip health checks
./run-all-services-test.sh -s

# Quiet mode
./run-all-services-test.sh -q

# Get help
./run-all-services-test.sh -h
```

### Using k6 Directly

```bash
# Load from .env.local
export $(cat .env.local | grep -v '^#' | xargs)

# Run test
k6 run all-services-load-test.js

# With custom parameters
k6 run \
  -e BASE_URL=https://velure.local \
  -e TEST_DURATION=5m \
  -e WARMUP_DURATION=1m \
  all-services-load-test.js

# Output results to file
k6 run --out json=results.json all-services-load-test.js

# With InfluxDB (if configured)
k6 run --out influxdb=http://localhost:8086/k6 all-services-load-test.js
```

## 📊 Monitoring

### Real-time Monitoring

**Terminal 1 - Run test:**
```bash
./run-all-services-test.sh
```

**Terminal 2 - Watch containers:**
```bash
watch -n 1 'docker ps --format "table {{.Names}}\t{{.Status}}\t{{.CPUPerc}}\t{{.MemPerc}}"'
```

**Terminal 3 - Watch logs:**
```bash
# Watch all services
docker-compose -f infrastructure/local/docker-compose.yaml logs -f

# Watch specific service
docker logs -f <service-name>
```

### Grafana Dashboards

1. Open Grafana: http://localhost:3000 (admin/admin)
2. Navigate to:
   - **Velure - Microservices Overview**: Overall system health
   - **Auth Service**: Auth-specific metrics
   - **Product Service**: Product service metrics
   - **Publish Order Service**: Order service metrics
   - **Process Order Service**: Background processing metrics

### Prometheus Queries

Access Prometheus at http://localhost:9090

Useful queries:
```promql
# Request rate per service
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m])

# Response time percentiles
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Active connections
rabbitmq_connections
```

## 📈 Interpreting Results

### k6 Output

```
scenarios: (100.00%) 4 scenarios, 800 max VUs, 6m30s max duration
✓ http_req_duration..............: avg=245ms  min=12ms   med=180ms  max=2.1s   p(95)=650ms  p(99)=1.2s
✓ http_req_failed................: 2.31%  ✓ 234   ✗ 9866
  http_reqs......................: 10100  167/s
```

#### Key Metrics

- **http_req_duration**: Response time distribution
  - p(95) < 2000ms: ✅ 95% of requests under 2 seconds
  - p(99) < 5000ms: ✅ 99% of requests under 5 seconds

- **http_req_failed**: Error rate
  - < 10%: ✅ Less than 10% errors

- **http_reqs**: Total requests and rate
  - Shows throughput (requests/second)

#### Custom Metrics

- **auth_requests_total**: Total auth requests
- **auth_error_rate**: Auth-specific error rate
- **auth_request_duration**: Auth response time

- **product_requests_total**: Total product requests
- **product_error_rate**: Product-specific error rate
- **product_request_duration**: Product response time

- **order_requests_total**: Total order requests
- **order_error_rate**: Order-specific error rate
- **order_request_duration**: Order response time

### Success Criteria

✅ **PASS** if:
- p95 response time < 2000ms
- p99 response time < 5000ms
- Error rate < 10%
- No service crashes
- All scenarios complete

❌ **FAIL** if:
- Response times exceed thresholds
- Error rate > 10%
- Services crash or become unresponsive
- Database connections exhausted

### Common Issues

1. **High error rate**
   - Check service logs: `docker logs <service>`
   - Verify database connectivity
   - Check RabbitMQ status

2. **Slow response times**
   - Monitor CPU/Memory: `docker stats`
   - Check database query performance
   - Review service logs for bottlenecks

3. **Connection refused**
   - Verify services are running: `docker ps`
   - Check network connectivity
   - Verify URLs in `.env.local`

## 🎯 Best Practices

1. **Start with warmup**: Always use warmup period to avoid cold start issues
2. **Monitor in real-time**: Watch dashboards during test execution
3. **Baseline first**: Run test without load to establish baseline
4. **Gradual ramp-up**: Use stages to gradually increase load
5. **Cool-down period**: Allow services to stabilize after peak load
6. **Clean data**: Clear test data between runs if needed
7. **Document results**: Save results and compare over time

## 📚 Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Best Practices](https://k6.io/docs/testing-guides/test-types/)
- [Grafana Dashboards](http://localhost:3000)
- [Prometheus](http://localhost:9090)
- [RabbitMQ Management](http://localhost:15672) (admin/admin_password)

## 🔍 Troubleshooting

### k6 not found
```bash
# Install k6
brew install k6  # macOS
```

### Services not responding
```bash
# Check service health
make health

# Restart services
make dev-stop && make dev
```

### SSL certificate errors
```bash
# Add to /etc/hosts
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts

# Accept browser certificate warning first
open https://velure.local
```

### RabbitMQ connection issues
```bash
# Check RabbitMQ
docker logs rabbitmq

# Restart RabbitMQ
docker restart rabbitmq
```

---

**Happy Load Testing! 🚀**
