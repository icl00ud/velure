# Velure Microservices Load Testing with K6

This directory contains comprehensive load testing scripts for all Velure microservices using K6.

## Test Scripts

### Individual Service Tests
- **`auth-service-test.js`** - Load test for authentication service
- **`product-service-test.js`** - Load test for product service  
- **`publish-order-service-test.js`** - Load test for order publishing service
- **`ui-service-test.js`** - Load test for UI service

### Integrated Test
- **`integrated-load-test.js`** - Comprehensive test across all services

## Load Testing Pattern

All tests follow a **15-second escalation pattern**:

```
Stage 1: 15s → Ramp up to initial load
Stage 2: 15s → Increase load 
Stage 3: 15s → Continue scaling
Stage 4: 15s → Reach higher load
Stage 5: 15s → Peak load preparation
Stage 6: 15s → PEAK LOAD
Stage 7: 15s → Ramp down
Stage 8: 15s → Continue ramp down
Stage 9: 15s → Return to 0 users
```

## Prerequisites

1. **Install K6**:
   ```bash
   # macOS
   brew install k6
   
   # Ubuntu/Debian
   sudo apt update && sudo apt install k6
   
   # Windows
   choco install k6
   ```

2. **Start all services**:
   ```bash
   # From project root
   docker-compose up -d
   ```

3. **Verify services are running**:
   - Auth Service: http://localhost:3020
   - Product Service: http://localhost:3010  
   - Publish Order Service: http://localhost:3030
   - UI Service: http://localhost:80

## Running Tests

### Individual Service Tests

```bash
# Test auth service
k6 run auth-service-test.js

# Test product service
k6 run product-service-test.js

# Test order service
k6 run publish-order-service-test.js

# Test UI service
k6 run ui-service-test.js
```

### Integrated Test (All Services)

```bash
# Test all services simultaneously
k6 run integrated-load-test.js
```

### Run with Custom Options

```bash
# Run with different user count
k6 run --stage 15s:50,15s:100,15s:150 auth-service-test.js

# Run with specific duration
k6 run --duration 60s --vus 50 product-service-test.js

# Generate detailed HTML report
k6 run --out html=results.html integrated-load-test.js
```

## Test Configuration

### Performance Thresholds

Each test includes performance thresholds:
- **Response Time**: 95% of requests < specific threshold per service
- **Error Rate**: < 5-15% depending on service complexity
- **Request Volume**: Minimum request counts per service

### Service-Specific Load Patterns

| Service | Peak Users | Response Threshold | Error Threshold |
|---------|------------|-------------------|-----------------|
| Auth | 200 | 500ms | 10% |
| Product | 400 | 1000ms | 5% |
| Order | 200 | 2000ms | 10% |
| UI | 250 | 3000ms | 15% |
| Integrated | 500 | 2000ms | 10% |

## Test Scenarios

### Auth Service
- User registration
- User login/logout  
- Token validation
- User profile retrieval

### Product Service  
- Product listing (all, by page, by category)
- Product search by name
- Product creation
- Product count queries
- Health checks

### Order Service
- Order creation with random data
- Order status updates
- Order processing workflow

### UI Service
- Homepage loading
- Static asset delivery
- Navigation between pages
- API endpoint testing

### Integrated Test
- Cross-service workflow simulation
- Realistic user journey patterns
- Service interaction testing
- End-to-end performance validation

## Monitoring During Tests

### Real-time Monitoring
```bash
# Monitor system resources
htop

# Monitor Docker containers
docker stats

# Monitor specific service logs
docker-compose logs -f auth-service
docker-compose logs -f product-service
```

### K6 Metrics
K6 provides real-time metrics including:
- HTTP request rate
- Response times (avg, p95, p99)
- Error rates
- Active virtual users
- Data transfer rates

## Interpreting Results

### Success Criteria
- ✅ All thresholds passed
- ✅ Error rate below threshold
- ✅ Response times within limits
- ✅ No service crashes

### Common Issues
- **High error rates**: Check service logs for connection issues
- **Slow response times**: Monitor CPU/memory usage
- **Connection refused**: Verify services are running and ports accessible
- **Database timeouts**: Check database connection limits

## Scaling Recommendations

Based on test results, consider:
- **Horizontal scaling**: Add more service instances
- **Database optimization**: Connection pooling, query optimization
- **Caching**: Implement Redis caching for frequently accessed data
- **Load balancing**: Distribute traffic across multiple instances

## Custom Test Development

To create custom tests:

1. Copy an existing test file
2. Modify the `BASE_URL` and endpoints
3. Adjust the load stages in `options.stages`
4. Update test scenarios and checks
5. Set appropriate thresholds

Example custom stage configuration:
```javascript
export const options = {
  stages: [
    { duration: '15s', target: 10 },
    { duration: '15s', target: 25 },
    { duration: '15s', target: 50 },
    // ... continue pattern
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    errors: ['rate<0.1'],
  },
};
```