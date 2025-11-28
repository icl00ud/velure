# Troubleshooting PostgreSQL Connection Exhaustion

## 🔴 Problem Identified

Your load test revealed a **critical production issue**: PostgreSQL connection pool exhaustion.

### Error Message
```
FATAL: remaining connection slots are reserved for roles with the SUPERUSER attribute (SQLSTATE 53300)
```

### What This Means
- PostgreSQL RDS has reached its **maximum connection limit**
- All available connections are in use
- New connection attempts are rejected
- This causes **500 Internal Server Error** for all auth operations

### Impact on Load Test Results
```
✗ http_req_duration p(95)=17.9s (threshold: <3s)
✗ end_to_end_duration p(95)=36831.9ms (threshold: <15s)
✓ registration_failures rate=0.55% (threshold: <15%)
```

**Good news:** Your system handled this gracefully with only 0.55% registration failures!
**Bad news:** Response times degraded significantly due to connection retry logic.

---

## 🔍 Root Cause Analysis

### PostgreSQL Connection Limits

RDS PostgreSQL connection limit is calculated as:
```
max_connections = {DBInstanceClassMemory/9531392}
```

**Common RDS instance limits:**
| Instance Type | Memory | max_connections |
|---------------|--------|-----------------|
| db.t3.micro | 1 GB | ~100 |
| db.t3.small | 2 GB | ~200 |
| db.t4g.micro | 1 GB | ~100 |
| db.t4g.small | 2 GB | ~200 |
| db.m5.large | 8 GB | ~800 |

**Connection breakdown:**
- Reserved for SUPERUSER: 3 connections
- Available for regular users: `max_connections - 3`
- Your error occurred when hitting this limit

### Why Load Tests Exhaust Connections

Each Virtual User (VU) in k6:
1. Registers → Opens 2-3 DB connections (check existing, insert user, create session)
2. Logs in → Opens 2-3 DB connections (find user, check session, create session)
3. These connections may not close immediately due to connection pooling

**At 150 VUs:**
- Concurrent operations: 150 registrations + 150 logins
- Peak connections needed: ~900 (150 VUs × 3 ops × 2 conn each)
- Your RDS limit: Likely 100-200 connections
- **Result:** Exhaustion at 50-70 VUs

---

## ✅ Solutions

### Solution 1: Increase RDS max_connections (Quick Fix)

#### Check Current Limit
```bash
# Via AWS CLI
aws rds describe-db-parameters \
  --db-parameter-group-name <your-parameter-group> \
  --query "Parameters[?ParameterName=='max_connections']"

# Or via SQL (requires SUPERUSER or rds_superuser role)
kubectl exec -it <auth-service-pod> -- \
  psql -h velure-production-auth.cw9gu66melkv.us-east-1.rds.amazonaws.com \
  -U postgres -d velure_auth -c "SHOW max_connections;"
```

#### Increase Limit via Parameter Group

**Option A: Use RDS Console**
1. Go to RDS → Parameter Groups
2. Find your parameter group (or create custom one)
3. Edit `max_connections` parameter
4. Set value: Recommended `{LEAST(DBInstanceClassMemory/9531392,5000)}`
5. Apply changes (requires DB reboot)

**Option B: Use AWS CLI**
```bash
# Create custom parameter group if needed
aws rds create-db-parameter-group \
  --db-parameter-group-name velure-custom-pg \
  --db-parameter-group-family postgres15 \
  --description "Custom parameters for Velure"

# Modify max_connections
aws rds modify-db-parameter-group \
  --db-parameter-group-name velure-custom-pg \
  --parameters "ParameterName=max_connections,ParameterValue=500,ApplyMethod=pending-reboot"

# Apply parameter group to your instance
aws rds modify-db-instance \
  --db-instance-identifier velure-production-auth \
  --db-parameter-group-name velure-custom-pg \
  --apply-immediately

# Reboot to apply changes
aws rds reboot-db-instance \
  --db-instance-identifier velure-production-auth
```

**⚠️ Warning:** Increasing `max_connections` increases memory usage. Ensure your RDS instance has sufficient memory.

**Recommended values:**
- Development: 200
- Staging: 300-500
- Production: 500-1000 (depending on instance size)

---

### Solution 2: Optimize Connection Pooling (Recommended)

Your Go services likely use GORM with default connection pool settings. These need tuning.

#### Check Current Settings

Look for this in your auth-service code:
```go
// services/auth-service/internal/database/database.go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(???)
sqlDB.SetMaxIdleConns(???)
sqlDB.SetConnMaxLifetime(???)
```

#### Recommended Configuration

**For auth-service:**
```go
// After initializing GORM
sqlDB, err := db.DB()
if err != nil {
    return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
}

// Connection pool configuration
sqlDB.SetMaxOpenConns(25)           // Max 25 connections per pod
sqlDB.SetMaxIdleConns(10)           // Keep 10 idle connections
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Recycle connections every 5 min
sqlDB.SetConnMaxIdleTime(2 * time.Minute)  // Close idle connections after 2 min

log.Printf("Database connection pool configured: MaxOpen=%d, MaxIdle=%d", 25, 10)
```

**For publish-order-service:**
```go
sqlDB.SetMaxOpenConns(15)           // Fewer connections (less traffic)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
sqlDB.SetConnMaxIdleTime(2 * time.Minute)
```

**Calculation:**
```
Total max connections = (Pods × MaxOpenConns) + Buffer
Example: (3 auth pods × 25 conns) + (2 order pods × 15 conns) + 20 buffer = 125 connections

Ensure: Total < RDS max_connections - 3 (reserved)
```

---

### Solution 3: Implement Connection Retry Logic

Add exponential backoff for connection failures:

```go
// internal/database/database.go
func ConnectWithRetry(dsn string, maxRetries int) (*gorm.DB, error) {
    var db *gorm.DB
    var err error

    for i := 0; i < maxRetries; i++ {
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
            Logger: logger.Default.LogMode(logger.Info),
        })

        if err == nil {
            // Test connection
            sqlDB, _ := db.DB()
            if err := sqlDB.Ping(); err == nil {
                return db, nil
            }
        }

        // Exponential backoff: 1s, 2s, 4s, 8s, 16s
        waitTime := time.Duration(1<<uint(i)) * time.Second
        log.Printf("Connection attempt %d/%d failed: %v. Retrying in %v...",
            i+1, maxRetries, err, waitTime)
        time.Sleep(waitTime)
    }

    return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}
```

---

### Solution 4: Use PgBouncer (Advanced)

PgBouncer is a lightweight connection pooler for PostgreSQL.

#### Architecture
```
[K8s Pods] → [PgBouncer] → [RDS PostgreSQL]
  100 pods    10 pooled       RDS
  × 25 conns  connections     instance
  = 2500
```

#### Deploy PgBouncer in Kubernetes

**pgbouncer-config.yaml:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pgbouncer-config
data:
  pgbouncer.ini: |
    [databases]
    velure_auth = host=velure-production-auth.cw9gu66melkv.us-east-1.rds.amazonaws.com port=5432 dbname=velure_auth

    [pgbouncer]
    listen_addr = 0.0.0.0
    listen_port = 5432
    auth_type = md5
    auth_file = /etc/pgbouncer/userlist.txt
    pool_mode = transaction
    max_client_conn = 1000
    default_pool_size = 25
    reserve_pool_size = 5
    reserve_pool_timeout = 3
    server_lifetime = 3600
    server_idle_timeout = 600
    log_connections = 1
    log_disconnections = 1

  userlist.txt: |
    "postgres" "md5<password_hash>"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pgbouncer
spec:
  replicas: 2
  selector:
    matchLabels:
      app: pgbouncer
  template:
    metadata:
      labels:
        app: pgbouncer
    spec:
      containers:
      - name: pgbouncer
        image: edoburu/pgbouncer:1.21.0
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: config
          mountPath: /etc/pgbouncer
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
      volumes:
      - name: config
        configMap:
          name: pgbouncer-config
---
apiVersion: v1
kind: Service
metadata:
  name: pgbouncer
spec:
  selector:
    app: pgbouncer
  ports:
  - port: 5432
    targetPort: 5432
```

**Update service connection strings:**
```bash
# Before
POSTGRES_HOST=velure-production-auth.cw9gu66melkv.us-east-1.rds.amazonaws.com

# After
POSTGRES_HOST=pgbouncer.default.svc.cluster.local
```

**Benefits:**
- Reduce RDS connections from 2500 to 25
- Handle 1000+ client connections with minimal RDS load
- Automatic connection recycling
- Transaction-level pooling

---

### Solution 5: Horizontal Scaling with Connection Awareness

Instead of scaling to more pods, optimize connection distribution:

**Current (Problem):**
```
3 auth-service pods × 100 connections = 300 connections
At 150 VUs → Exhaustion
```

**Optimized:**
```
3 auth-service pods × 20 connections = 60 connections
At 150 VUs → Healthy (with proper pooling)
```

**HPA with connection metrics:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: auth-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: auth-service
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

---

## 📊 Monitoring Connection Usage

### Query Active Connections

```sql
-- Total connections
SELECT count(*) FROM pg_stat_activity;

-- Connections by database
SELECT datname, count(*)
FROM pg_stat_activity
GROUP BY datname;

-- Connections by state
SELECT state, count(*)
FROM pg_stat_activity
GROUP BY state;

-- Identify connection hogs
SELECT
    usename,
    application_name,
    client_addr,
    state,
    count(*)
FROM pg_stat_activity
WHERE state = 'idle'
GROUP BY usename, application_name, client_addr, state
ORDER BY count DESC;
```

### CloudWatch Metrics to Monitor

In AWS CloudWatch, track:
- **DatabaseConnections** - Current connection count
- **CPUUtilization** - Should stay below 70%
- **FreeableMemory** - Should not be critically low
- **ReadLatency / WriteLatency** - Should stay low

**Set alarms:**
```bash
aws cloudwatch put-metric-alarm \
  --alarm-name velure-db-connections-high \
  --alarm-description "Database connections above 80%" \
  --metric-name DatabaseConnections \
  --namespace AWS/RDS \
  --statistic Average \
  --period 300 \
  --evaluation-periods 2 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --dimensions Name=DBInstanceIdentifier,Value=velure-production-auth
```

---

## 🎯 Recommended Action Plan

### Immediate (Today)
1. ✅ **Increase RDS max_connections to 500** (or appropriate for your instance)
2. ✅ **Verify parameter group is applied and reboot RDS**
3. ✅ **Re-run load test to verify fix**

### Short-term (This Week)
4. ✅ **Add connection pooling config to all services**
   - auth-service: MaxOpenConns=25
   - publish-order-service: MaxOpenConns=15
   - process-order-service: MaxOpenConns=10
5. ✅ **Add connection retry logic with exponential backoff**
6. ✅ **Set up CloudWatch alarms for connection count**

### Long-term (Next Sprint)
7. ✅ **Implement PgBouncer for connection pooling**
8. ✅ **Consider read replicas for read-heavy operations**
9. ✅ **Add database connection metrics to Prometheus**
10. ✅ **Regular load testing in staging environment**

---

## 🧪 Verify the Fix

After implementing solutions, re-run the load test:

```bash
cd tests/load
./run-k8s-load-test.sh k8s
```

**Expected results:**
```
✓ http_req_duration p(95) < 3s
✓ registration_failures rate < 5%
✓ No "remaining connection slots" errors
```

**Monitor:**
```bash
# Terminal 1: Run test
./run-k8s-load-test.sh k8s

# Terminal 2: Monitor connections
watch -n 5 "kubectl exec -it <postgres-pod> -- \
  psql -h velure-production-auth.cw9gu66melkv.us-east-1.rds.amazonaws.com \
  -U postgres -d velure_auth \
  -c 'SELECT count(*) as connections FROM pg_stat_activity;'"
```

---

## 📚 Additional Resources

- [AWS RDS PostgreSQL Connection Limits](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html)
- [GORM Connection Pool Configuration](https://gorm.io/docs/generic_interface.html)
- [PgBouncer Documentation](https://www.pgbouncer.org/config.html)
- [PostgreSQL Connection Pooling Best Practices](https://www.postgresql.org/docs/current/runtime-config-connection.html)

---

## 🎉 Conclusion

**This load test was successful** because it revealed a production-critical issue **before it caused an outage**.

The 0.55% failure rate and 99.45% success rate show your system degrades gracefully under connection exhaustion, but the p95 latency of 17.9s indicates users would experience severe slowdowns.

By implementing the solutions above, you should be able to handle 150+ concurrent users with p95 latency under 1 second.
