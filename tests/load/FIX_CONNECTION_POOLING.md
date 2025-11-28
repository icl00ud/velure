# Fix: Implementing Connection Pooling in Go Services

## 🎯 Quick Reference

**Problem:** PostgreSQL connection exhaustion under load
**Solution:** Configure GORM connection pooling with appropriate limits
**Impact:** Reduces RDS connections from ~900 to ~100, eliminates 500 errors
**Difficulty:** Easy (5-10 minutes per service)

---

## 📝 Step-by-Step Implementation

### 1. auth-service Connection Pooling

**File:** `services/auth-service/internal/database/database.go`

Find the database initialization code and add connection pool configuration:

```go
package database

import (
    "fmt"
    "log"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func NewConnection(dsn string) (*gorm.DB, error) {
    // Open database connection
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        PrepareStmt: true, // Cache prepared statements
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Get underlying SQL DB to configure connection pool
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
    }

    // ====== CONNECTION POOL CONFIGURATION ======

    // SetMaxOpenConns sets the maximum number of open connections to the database
    // Calculation: (Expected Pods × MaxOpenConns) should be < RDS max_connections
    // Example: 3 auth-service pods × 25 = 75 connections (well under 500 RDS limit)
    sqlDB.SetMaxOpenConns(25)

    // SetMaxIdleConns sets the maximum number of connections in the idle connection pool
    // Rule of thumb: 30-50% of MaxOpenConns
    sqlDB.SetMaxIdleConns(10)

    // SetConnMaxLifetime sets the maximum amount of time a connection may be reused
    // Prevents stale connections and ensures fresh connections periodically
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    // SetConnMaxIdleTime sets the maximum amount of time a connection may be idle
    // Closes idle connections faster to free up resources
    sqlDB.SetConnMaxIdleTime(2 * time.Minute)

    // ====== VERIFY CONNECTION ======

    if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Printf("✅ Database connected successfully")
    log.Printf("📊 Connection pool: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%s, MaxIdleTime=%s",
        25, 10, 5*time.Minute, 2*time.Minute)

    return db, nil
}

// Optional: Add connection pool metrics for monitoring
func GetConnectionStats(db *gorm.DB) map[string]interface{} {
    sqlDB, _ := db.DB()
    stats := sqlDB.Stats()

    return map[string]interface{}{
        "max_open_connections":   stats.MaxOpenConnections,
        "open_connections":       stats.OpenConnections,
        "in_use":                 stats.InUse,
        "idle":                   stats.Idle,
        "wait_count":             stats.WaitCount,
        "wait_duration_ms":       stats.WaitDuration.Milliseconds(),
        "max_idle_closed":        stats.MaxIdleClosed,
        "max_lifetime_closed":    stats.MaxLifetimeClosed,
    }
}
```

---

### 2. publish-order-service Connection Pooling

**File:** `services/publish-order-service/internal/repository/order_repository.go`

The publish-order-service likely initializes its DB connection in the repository. Update it:

```go
package repository

import (
    "database/sql"
    "fmt"
    "time"

    _ "github.com/lib/pq"
)

type PostgresOrderRepository struct {
    db *sql.DB
}

func NewOrderRepository(connStr string) (OrderRepository, error) {
    // Open connection
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // ====== CONNECTION POOL CONFIGURATION ======

    // publish-order-service handles less traffic than auth-service
    // so we use fewer connections
    db.SetMaxOpenConns(15)           // Max 15 connections per pod
    db.SetMaxIdleConns(5)            // Keep 5 idle connections
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(2 * time.Minute)

    // ====== VERIFY CONNECTION ======

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    zap.L().Info("Database connected",
        zap.Int("max_open_conns", 15),
        zap.Int("max_idle_conns", 5),
        zap.Duration("conn_max_lifetime", 5*time.Minute))

    return &PostgresOrderRepository{db: db}, nil
}

// Add method to expose DB for migrations
func (r *PostgresOrderRepository) DB() *sql.DB {
    return r.db
}
```

---

### 3. Environment-Based Configuration (Optional but Recommended)

Make connection pool settings configurable via environment variables:

**File:** `services/auth-service/internal/config/config.go`

```go
package config

import (
    "os"
    "strconv"
    "time"
)

type DatabaseConfig struct {
    DSN                  string
    MaxOpenConns         int
    MaxIdleConns         int
    ConnMaxLifetime      time.Duration
    ConnMaxIdleTime      time.Duration
}

func LoadDatabaseConfig() DatabaseConfig {
    return DatabaseConfig{
        DSN:                  os.Getenv("DATABASE_URL"),
        MaxOpenConns:         getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
        MaxIdleConns:         getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
        ConnMaxLifetime:      getEnvAsDuration("DB_CONN_MAX_LIFETIME", "5m"),
        ConnMaxIdleTime:      getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", "2m"),
    }
}

func getEnvAsInt(key string, defaultVal int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultVal
}

func getEnvAsDuration(key string, defaultVal string) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    duration, _ := time.ParseDuration(defaultVal)
    return duration
}
```

**File:** `services/auth-service/internal/database/database.go` (updated)

```go
func NewConnection(cfg config.DatabaseConfig) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        PrepareStmt: true,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
    }

    // Use config values
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

    if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Printf("✅ Database connected with pool: MaxOpen=%d, MaxIdle=%d",
        cfg.MaxOpenConns, cfg.MaxIdleConns)

    return db, nil
}
```

---

### 4. Kubernetes ConfigMap/Env Updates

**File:** `infrastructure/kubernetes/deployments/auth-service.yaml`

Add environment variables for connection pooling:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: auth-service
        image: icl00ud/velure-auth:latest
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: velure-secrets
              key: auth-db-url
        # Connection pool configuration
        - name: DB_MAX_OPEN_CONNS
          value: "25"
        - name: DB_MAX_IDLE_CONNS
          value: "10"
        - name: DB_CONN_MAX_LIFETIME
          value: "5m"
        - name: DB_CONN_MAX_IDLE_TIME
          value: "2m"
```

---

## 🧮 Connection Pool Sizing Calculator

Use this formula to calculate appropriate pool sizes:

```
Total Connections = Σ(Service Pods × MaxOpenConns) + Buffer

Must satisfy: Total Connections < (RDS max_connections - 3)
```

**Example for your setup:**

| Service | Pods | MaxOpenConns | Total |
|---------|------|--------------|-------|
| auth-service | 3 | 25 | 75 |
| publish-order | 2 | 15 | 30 |
| process-order | 2 | 10 | 20 |
| **Subtotal** | | | **125** |
| Buffer (20%) | | | 25 |
| **Grand Total** | | | **150** |

**Recommendation:**
- RDS max_connections: 500 (allows 3.3x headroom for spikes)
- If using HPA to scale to 10 pods each: 10 × (25+15+10) = 500 connections

---

## 📊 Monitoring Connection Pool Health

### Add Prometheus Metrics

**File:** `services/auth-service/internal/metrics/database.go` (new file)

```go
package metrics

import (
    "database/sql"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "gorm.io/gorm"
)

var (
    dbConnectionsOpen = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "db_connections_open",
        Help: "Number of open database connections",
    })

    dbConnectionsInUse = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "db_connections_in_use",
        Help: "Number of database connections in use",
    })

    dbConnectionsIdle = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "db_connections_idle",
        Help: "Number of idle database connections",
    })

    dbConnectionsWaitCount = promauto.NewCounter(prometheus.CounterOpts{
        Name: "db_connections_wait_count_total",
        Help: "Total number of times waited for a connection",
    })

    dbConnectionsWaitDuration = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "db_connections_wait_duration_seconds",
        Help: "Total time waited for database connections",
    })
)

// StartConnectionPoolMonitor starts a goroutine that periodically updates connection pool metrics
func StartConnectionPoolMonitor(db *gorm.DB, interval time.Duration) {
    sqlDB, _ := db.DB()

    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        for range ticker.C {
            updateConnectionPoolMetrics(sqlDB)
        }
    }()
}

func updateConnectionPoolMetrics(db *sql.DB) {
    stats := db.Stats()

    dbConnectionsOpen.Set(float64(stats.OpenConnections))
    dbConnectionsInUse.Set(float64(stats.InUse))
    dbConnectionsIdle.Set(float64(stats.Idle))
    dbConnectionsWaitCount.Add(float64(stats.WaitCount))
    dbConnectionsWaitDuration.Set(stats.WaitDuration.Seconds())
}
```

**Usage in main.go:**

```go
import "your-module/internal/metrics"

func main() {
    // ... after DB connection
    db := database.NewConnection(cfg.Database)

    // Start monitoring connection pool every 10 seconds
    metrics.StartConnectionPoolMonitor(db, 10*time.Second)

    // ... rest of your code
}
```

### Grafana Dashboard Query

Add these queries to your Grafana dashboard:

```promql
# Connection pool utilization (%)
(db_connections_in_use / db_connections_open) * 100

# Idle connections
db_connections_idle

# Connection wait rate (per second)
rate(db_connections_wait_count_total[1m])

# Alert: Connection pool saturated
(db_connections_in_use / db_connections_open) > 0.9
```

---

## ✅ Verification Checklist

After implementing connection pooling:

- [ ] Code changes committed to all services
- [ ] Environment variables added to Kubernetes deployments
- [ ] Services redeployed with new configuration
- [ ] Logs show "Connection pool configured" message
- [ ] Prometheus metrics showing db_connections_* data
- [ ] Load test re-run with successful results
- [ ] CloudWatch shows RDS connections < 200 at peak load
- [ ] No more "connection slots reserved" errors

---

## 🎯 Expected Improvement

**Before (Connection Exhaustion):**
```
Load: 70 VUs
RDS Connections: 420 (exceeded limit of 200)
p95 Latency: 17.9s
Error Rate: 0.55%
```

**After (Proper Pooling):**
```
Load: 150 VUs
RDS Connections: 150 (well under limit of 500)
p95 Latency: <2s
Error Rate: <0.1%
```

---

## 📚 Learn More

- [GORM Generic Database Interface](https://gorm.io/docs/generic_interface.html)
- [Go database/sql Package](https://pkg.go.dev/database/sql)
- [PostgreSQL Connection Pool Tuning](https://wiki.postgresql.org/wiki/Number_Of_Database_Connections)
- [AWS RDS Best Practices](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_BestPractices.html)
