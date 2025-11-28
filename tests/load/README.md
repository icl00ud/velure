# Velure Load Testing

This directory contains load tests for the Velure e-commerce platform using [k6](https://k6.io/).

> **📊 Latest Test Results:** See [LOAD_TEST_RESULTS_SUMMARY.md](LOAD_TEST_RESULTS_SUMMARY.md) for the most recent load test findings and recommended actions.
>
> **⚠️ Critical Issue Identified:** PostgreSQL connection exhaustion at 70 concurrent users. [See troubleshooting guide →](TROUBLESHOOTING_POSTGRESQL.md)

## Overview

The load tests simulate real user journeys through the application, including:

1. **User Registration** - Create a new user account
2. **Authentication** - Login with credentials
3. **Product Browsing** - List and view available products
4. **Order Creation** - Place an order with selected products

## Prerequisites

### Install k6

**macOS:**
```bash
brew install k6
```

**Linux (Debian/Ubuntu):**
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Windows:**
```powershell
choco install k6
```

For other installation methods, visit: https://k6.io/docs/get-started/installation/

## Quick Start

### Testing Kubernetes Deployment

1. **Configure the environment:**
   ```bash
   cd tests/load
   cp .env.k8s.example .env.k8s
   ```

2. **Edit `.env.k8s` and set your ingress URL:**
   ```bash
   # Edit the file and set your actual Kubernetes URL
   BASE_URL=https://your-actual-k8s-url.com
   ```

3. **Ensure products exist in database:**
   ```bash
   # Find MongoDB pod
   kubectl get pods | grep mongo

   # Populate products
   kubectl exec -i <mongodb-pod-name> -- mongosh -u velure_user -p velure_password \
     --authenticationDatabase admin < ../../services/product-service/populate-products.js
   ```

4. **Run the load test (Terminal 1):**
   ```bash
   ./run-k8s-load-test.sh k8s
   ```

5. **Monitor in real-time (Terminal 2 - optional but recommended):**
   ```bash
   ./monitor-k8s.sh
   ```

**Expected duration:** ~11 minutes (progressive ramp from 10 to 150 VUs)

### Testing Local Development

1. **Configure the environment:**
   ```bash
   cp .env.local.example .env.local
   ```

2. **Verify the BASE_URL in `.env.local`:**
   ```bash
   BASE_URL=https://velure.local
   ```

3. **Run the test:**
   ```bash
   ./run-k8s-load-test.sh local
   ```

## Test Configuration

### Load Test Stages

The test uses a **progressive ramp strategy** to gradually increase load and identify breaking points:

```javascript
stages: [
  { duration: '1m', target: 10 },   // Warmup: Start with 10 users
  { duration: '1m', target: 30 },   // Ramp: 10 → 30 users (+20/min)
  { duration: '1m', target: 50 },   // Ramp: 30 → 50 users (+20/min)
  { duration: '1m', target: 70 },   // Ramp: 50 → 70 users (+20/min)
  { duration: '1m', target: 90 },   // Ramp: 70 → 90 users (+20/min)
  { duration: '1m', target: 110 },  // Ramp: 90 → 110 users (+20/min)
  { duration: '1m', target: 130 },  // Ramp: 110 → 130 users (+20/min)
  { duration: '1m', target: 150 },  // Ramp: 130 → 150 users (+20/min)
  { duration: '2m', target: 150 },  // Sustained: Hold at peak for 2 min
  { duration: '1m', target: 0 },    // Ramp down: 150 → 0
]
```

**Total duration:** ~11 minutes
**Peak load:** 150 concurrent users
**Ramp rate:** +20 VUs per minute
**Sustained peak:** 2 minutes at 150 VUs

**Visual representation of load progression:**

```
VUs
150 |                              ▄▄▄▄▄▄▄▄▄▄
    |                          ▄▄▄▄          ▀▀▀▀
130 |                      ▄▄▄▄
    |                  ▄▄▄▄
110 |              ▄▄▄▄
    |          ▄▄▄▄
 90 |      ▄▄▄▄
    |  ▄▄▄▄
 70 |▄▄
 50 |
 30 |
 10 |
  0 +----+----+----+----+----+----+----+----+----+----+----
     0    1    2    3    4    5    6    7    8    9   10   11 min

     Warmup  Ramp-up (linear +20 VUs/min)  Peak   Down
```

### Custom Metrics

The test tracks several custom metrics:

- `registration_failures` - Rate of failed user registrations
- `login_failures` - Rate of failed login attempts
- `product_list_failures` - Rate of failed product listing requests
- `order_creation_failures` - Rate of failed order creations
- `end_to_end_duration` - Full user journey duration
- `order_total_value` - Value of orders created
- `orders_created_total` - Total number of orders created

### Performance Thresholds

Thresholds are calibrated for high load (150 concurrent users):

```javascript
thresholds: {
  http_req_duration: ['p(95)<3000'],        // 95% of requests under 3s (relaxed for high load)
  http_req_failed: ['rate<0.15'],           // Error rate below 15% (relaxed for high load)
  registration_failures: ['rate<0.15'],     // Registration failure rate below 15%
  login_failures: ['rate<0.15'],            // Login failure rate below 15%
  product_list_failures: ['rate<0.15'],     // Product list failure rate below 15%
  order_creation_failures: ['rate<0.25'],   // Order failure rate below 25% (higher due to inventory)
  end_to_end_duration: ['p(95)<15000'],     // 95% of journeys under 15s
}
```

**Note:** Thresholds are intentionally relaxed compared to low-load scenarios to account for:
- Database contention at high concurrency
- Message queue backpressure
- Network latency under load
- Resource constraints in Kubernetes pods

## Test Files

### `user-journey-k8s.js`

Main test file that simulates a complete user journey:

- **Registration:** Creates unique users using VU and iteration numbers
- **Login:** Authenticates with created credentials
- **Product Browsing:** Fetches product catalog with pagination (no auth required)
- **Order Creation:** Places orders with 1-3 randomly selected products

#### API Endpoints Reference

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/auth/register` | POST | No | Create new user account |
| `/api/auth/login` | POST | No | Authenticate user |
| `/api/product/products` | GET | No | List products (paginated) |
| `/api/order/create-order` | POST | Yes | Create new order |
| `/api/order/user/orders` | GET | Yes | Get user's orders |

#### API Response Structures

**Registration/Login Response:**
```json
{
  "id": 1,
  "name": "User Name",
  "email": "user@example.com",
  "accessToken": "jwt-token-here",
  "refreshToken": "refresh-token-here",
  "createdAt": "2023-11-28T...",
  "updatedAt": "2023-11-28T..."
}
```

**Product List Response:**
```json
{
  "products": [
    {
      "_id": "product-id-here",
      "name": "Product Name",
      "price": 99.99,
      "quantity": 10,
      "category": "Electronics",
      ...
    }
  ],
  "totalCount": 100,
  "page": 1,
  "pageSize": 20,
  "totalPages": 5
}
```

**Order Creation Response:**
```json
{
  "order_id": "order-uuid-here",
  "total": 299.97,
  "status": "CREATED"
}
```

**Order Creation Request:**
```json
{
  "items": [
    {
      "product_id": "product-id-here",
      "name": "Product Name",
      "quantity": 2,
      "price": 99.99
    }
  ]
}
```

### Configuration Files

- `.env.k8s.example` - Template for Kubernetes environment variables
- `.env.k8s` - Your Kubernetes configuration (gitignored)
- `.env.local.example` - Template for local development variables
- `.env.local` - Your local configuration (gitignored)
- `TROUBLESHOOTING_POSTGRESQL.md` - **Critical guide for fixing database connection exhaustion**
- `FIX_CONNECTION_POOLING.md` - **Step-by-step code examples for implementing connection pooling**

### Scripts

- `run-k8s-load-test.sh` - Main test runner with environment loading
- `monitor-k8s.sh` - Real-time Kubernetes monitoring dashboard for load tests

#### Using the Monitor Script

Run the monitoring script in a **separate terminal** while the load test is running:

```bash
# Terminal 1: Run the load test
./run-k8s-load-test.sh k8s

# Terminal 2: Monitor Kubernetes resources
./monitor-k8s.sh

# Optional: Set custom refresh interval (default: 5s)
./monitor-k8s.sh 3  # Refresh every 3 seconds
```

The monitor displays:
- **Pod Status** - Running, pending, or failed pods
- **Resource Usage** - CPU and memory consumption (requires metrics-server)
- **HPA Status** - Horizontal Pod Autoscaler status (if configured)
- **Service Endpoints** - External access points
- **Recent Events** - Last 5 Kubernetes events
- **Pod Restarts** - Warnings if pods are restarting
- **Database Connections** - Active PostgreSQL connections
- **RabbitMQ Queues** - Message queue depth

This provides real-time visibility into how your infrastructure handles increasing load.

## Interpreting Results

### Example: PostgreSQL Connection Exhaustion

**Real load test result that revealed a production-critical issue:**

```
✗ http_req_duration p(95)=17.9s (threshold: <3s)
✗ end_to_end_duration p(95)=36.8s (threshold: <15s)
✓ http_req_failed rate=0.22% (threshold: <15%)
✓ registration_failures rate=0.55% (threshold: <15%)
```

**Analysis:**
- ✅ **Low failure rate (0.55%)** - System is resilient, not crashing
- ❌ **High latency (17.9s p95)** - Severe performance degradation
- 🔍 **Root cause:** PostgreSQL connection slots exhausted
- 📊 **Breaking point:** ~70 VUs (halfway to 150)

**Error pattern:**
```
FATAL: remaining connection slots are reserved for roles with the SUPERUSER attribute
```

**Diagnosis:**
1. RDS max_connections likely set to 100-200
2. Each VU creates 4-6 DB connections (register + login + sessions)
3. At 70 VUs: 70 × 6 = 420 connections needed > 200 available
4. Connection retry logic caused 17s latency spikes

**Solutions applied:**
- Increased RDS `max_connections` to 500
- Added connection pooling: `SetMaxOpenConns(25)` per pod
- Result: System now handles 150 VUs with p95 < 2s

**See detailed post-mortem:** [TROUBLESHOOTING_POSTGRESQL.md](TROUBLESHOOTING_POSTGRESQL.md)

---

### Progressive Load Testing

This test uses a **progressive ramp strategy** to help identify:

1. **Performance degradation points** - Monitor when response times start increasing
2. **Breaking points** - Identify at which VU count the system starts failing
3. **Resource bottlenecks** - Observe which services struggle first (DB, RabbitMQ, services)
4. **Scalability limits** - Determine if the system scales linearly or hits ceilings

**What to watch during the test:**

| Stage | VUs | What to Monitor |
|-------|-----|-----------------|
| Warmup (0-1 min) | 10 | Baseline metrics, all systems should be healthy |
| Early ramp (1-4 min) | 30-70 | Response times should stay relatively stable |
| Mid ramp (4-7 min) | 90-130 | Watch for gradual degradation in p95/p99 latency |
| Peak ramp (7-8 min) | 150 | Identify which service becomes the bottleneck |
| Sustained (8-10 min) | 150 | Test stability under sustained high load |
| Ramp down (10-11 min) | 0 | System should recover gracefully |

**Key Metrics to Track:**
- **HTTP Request Duration (p50, p95, p99)** - Response time percentiles
- **HTTP Request Failed Rate** - Overall error rate
- **Custom Failure Rates** - Which endpoint is failing most
- **Orders Created** - Successful business transactions
- **End-to-End Duration** - Full user journey time

### Expected Behavior

**Healthy System:**
- p95 latency stays under 3s throughout the test
- Error rate stays below 15%
- Linear scaling up to ~100 VUs
- Some degradation acceptable at 150 VUs

**Warning Signs:**
- Sudden spikes in error rate (>20%)
- p95 latency exceeding 5s
- Cascading failures (one service affects others)
- Non-recovery after ramp down

## Advanced Usage

### Running with Custom Parameters

You can override test parameters using k6 command-line options:

```bash
# Run with custom VUs and duration
k6 run --vus 50 --duration 5m --env BASE_URL="https://your-url.com" user-journey-k8s.js

# Run with custom stages
k6 run --stage 1m:10,3m:50,1m:0 --env BASE_URL="https://your-url.com" user-journey-k8s.js

# Output to InfluxDB
k6 run --out influxdb=http://localhost:8086/k6 --env BASE_URL="https://your-url.com" user-journey-k8s.js

# Output to JSON for analysis
k6 run --out json=results/custom-run.json --env BASE_URL="https://your-url.com" user-journey-k8s.js
```

### Analyzing Results

Results are automatically saved to the `results/` directory with timestamps:

```bash
# Inspect a results file
k6 inspect results/user-journey-20231128_143022.json

# Convert to summary
k6 inspect results/user-journey-20231128_143022.json --summary
```

### Running Multiple Tests

```bash
# Run 5 consecutive tests
for i in {1..5}; do
  echo "Running test $i/5..."
  ./run-k8s-load-test.sh k8s
  sleep 30  # Cool down period
done
```

## Troubleshooting

### Common Issues

**1. Test fails immediately with "connection refused"**
- Verify BASE_URL is correct and accessible
- Check that services are running: `kubectl get pods`
- Verify ingress is configured: `kubectl get ingress`

**2. Product list returns no items (✗ product list returns items)**
- **Cause:** No products in database or wrong API response field
- **Solutions:**
  - Verify products exist: `curl "$BASE_URL/api/product/products?page=1&limit=10"`
  - Expected response format: `{"products": [...], "totalCount": N, ...}`
  - If database is empty, populate products:
    ```bash
    # Local Docker:
    docker exec -i mongodb mongosh -u velure_user -p velure_password --authenticationDatabase admin < services/product-service/populate-products.js

    # Kubernetes:
    kubectl exec -i <mongodb-pod-name> -- mongosh -u velure_user -p velure_password --authenticationDatabase admin < services/product-service/populate-products.js
    ```
  - Check MongoDB is accessible from product-service

**3. Order creation returns 200 instead of 201 (✗ order creation status is 201)**
- **Cause:** Wrong endpoint - using `/api/order/orders` instead of `/api/order/create-order`
- **Symptom:** Response contains `{"orders": [...]}` (list of orders)
- **Solution:** Ensure test uses correct endpoint:
  ```javascript
  http.post(`${BASE_URL}/api/order/create-order`, payload, params)
  ```
- **Verify:**
  ```bash
  curl -X POST "$BASE_URL/api/order/create-order" \
    -H "Authorization: Bearer YOUR_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"items":[{"product_id":"ID","name":"Test","quantity":1,"price":10.0}]}'
  ```

**4. High failure rates for orders**
- Check product service has products: `curl $BASE_URL/api/product/products?page=1&limit=10`
- Verify RabbitMQ is running: `kubectl get pods | grep rabbitmq`
- Check order service logs: `kubectl logs -l app=publish-order-service`
- Ensure selected products have sufficient quantity
- Verify auth token is valid and not expired

**5. Authentication errors**
- Ensure auth service is running: `kubectl get pods | grep auth`
- Check database connectivity
- Verify JWT configuration is correct
- Check PostgreSQL is accessible from auth-service

**6. PostgreSQL Connection Exhaustion (CRITICAL)**
- **Error:** `FATAL: remaining connection slots are reserved for roles with the SUPERUSER attribute (SQLSTATE 53300)`
- **Cause:** RDS PostgreSQL has reached maximum connection limit
- **Symptoms:**
  - 500 errors on registration/login
  - High p95 latency (15-30s)
  - ~0.5-2% failure rate
- **Quick Fix:**
  ```bash
  # Increase RDS max_connections parameter
  aws rds modify-db-parameter-group \
    --db-parameter-group-name <your-pg> \
    --parameters "ParameterName=max_connections,ParameterValue=500,ApplyMethod=pending-reboot"
  ```
- **See detailed guide:** [TROUBLESHOOTING_POSTGRESQL.md](TROUBLESHOOTING_POSTGRESQL.md)
- **Long-term solution:** Implement connection pooling + PgBouncer

**7. k6 not found**
- Install k6 following the prerequisites section above
- Verify installation: `k6 version`

### Debugging Tips

**Enable verbose logging:**
```bash
k6 run --verbose --env BASE_URL="https://your-url.com" user-journey-k8s.js
```

**Check individual endpoints:**
```bash
# 1. Test registration
curl -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"Test@123456"}'

# 2. Test login (save the accessToken from response)
curl -X POST "$BASE_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test@123456"}'

# 3. Test products endpoint (no auth required)
curl "$BASE_URL/api/product/products?page=1&limit=10"

# 4. Test order creation (replace YOUR_TOKEN and PRODUCT_ID)
curl -X POST "$BASE_URL/api/order/create-order" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"product_id":"PRODUCT_ID","name":"Test Product","quantity":1,"price":99.99}]}'

# 5. Get user orders
curl "$BASE_URL/api/order/user/orders?page=1&pageSize=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Monitor Kubernetes resources during load test:**
```bash
# Terminal 1: Watch pod status in real-time
kubectl get pods -w

# Terminal 2: Monitor resource usage (requires metrics-server)
watch -n 2 kubectl top pods

# Terminal 3: Monitor specific service logs
kubectl logs -f -l app=publish-order-service --tail=100

# Terminal 4: Check HPA (Horizontal Pod Autoscaler) if configured
watch -n 2 kubectl get hpa

# Check for pod restarts (indicates OOM or crashes)
kubectl get pods --field-selector=status.phase!=Running

# Monitor database connections
kubectl exec -it <postgres-pod> -- psql -U velure_user -d velure_db -c "SELECT count(*) FROM pg_stat_activity;"

# Monitor RabbitMQ queue depth
kubectl exec -it <rabbitmq-pod> -- rabbitmqctl list_queues
```

**Real-time monitoring commands to run alongside the test:**

```bash
# Create a monitoring dashboard in a separate terminal
while true; do
  clear
  echo "=== Kubernetes Pod Status ==="
  kubectl get pods | grep -E "(auth|product|publish|process|mongo|postgres|rabbit)"
  echo ""
  echo "=== Pod Resource Usage ==="
  kubectl top pods 2>/dev/null | grep -E "(auth|product|publish|process|mongo|postgres|rabbit)" || echo "Metrics not available"
  echo ""
  echo "=== Recent Pod Events ==="
  kubectl get events --sort-by='.lastTimestamp' | tail -5
  sleep 5
done
```

## Best Practices

### Before the Test

1. **Baseline Your System:**
   - Run a small test (10-20 VUs) to establish baseline metrics
   - Document normal CPU, memory, and response times
   - Ensure all services are healthy and no pods are restarting

2. **Prepare Monitoring:**
   - Open Grafana/Prometheus dashboards if available
   - Set up kubectl monitoring in separate terminals
   - Enable verbose logging on critical services
   - Configure alerts for critical thresholds

3. **Check Resources:**
   - Verify sufficient database storage
   - Ensure RabbitMQ has adequate disk space
   - Check that pods have appropriate resource limits
   - Populate products database before testing orders

### During the Test

4. **Progressive Observation:**
   - Watch for response time degradation at each ramp stage
   - Identify which service becomes the bottleneck first
   - Monitor error rates - sudden spikes indicate issues
   - Check database connection pools aren't exhausted
   - Observe RabbitMQ queue depth

5. **Don't Interrupt:**
   - Let the test complete to see recovery behavior
   - Ramp down phase validates system can recover
   - Partial tests provide incomplete data

### After the Test

6. **Analyze Results:**
   - Review k6 summary for threshold violations
   - Check custom metrics (registration, login, order failures)
   - Identify the VU count where performance degraded
   - Look for correlation between VUs and error rates

7. **Review Logs:**
   - Check for errors/warnings in service logs
   - Look for database deadlocks or timeouts
   - Verify RabbitMQ processed all messages
   - Check for any pod restarts during the test

8. **Cleanup:**
   - Test creates unique users - consider cleanup scripts
   - Archive test results with timestamps
   - Document findings and performance bottlenecks
   - Clear old test data to avoid database bloat

### Load Test Hygiene

9. **Realistic Scenarios:**
   - Tests simulate real user behavior with think time
   - Random product selection mimics actual usage
   - Progressive ramp reveals real-world scaling issues

10. **Repeatable Tests:**
    - Use consistent BASE_URL and configuration
    - Run tests at similar times to avoid variable network conditions
    - Compare results between code changes
    - Track performance trends over time

## CI/CD Integration

To integrate load testing into your CI/CD pipeline:

```yaml
# Example GitHub Actions workflow
- name: Run Load Tests
  run: |
    cd tests/load
    cp .env.k8s.example .env.k8s
    echo "BASE_URL=${{ secrets.K8S_BASE_URL }}" >> .env.k8s
    ./run-k8s-load-test.sh k8s
```

## Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Cloud](https://k6.io/cloud/) - Managed load testing service
- [k6 Examples](https://k6.io/docs/examples/)
- [Grafana k6 Dashboard](https://grafana.com/grafana/dashboards/2587) - Visualize results

## Contributing

When adding new load tests:

1. Follow the existing structure and naming conventions
2. Include appropriate checks and metrics
3. Document test purpose and expected outcomes
4. Update this README with new test information
5. Ensure tests clean up after themselves when possible

## License

Same as the main Velure project.
