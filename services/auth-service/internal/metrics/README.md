# Auth Service - Prometheus Metrics

This document describes every custom metric exposed by auth-service on the `/metrics` endpoint.

## Login Metrics

### `auth_login_attempts_total` (Counter)
Total login attempts.
- **Labels**: `status` (success, failure)
- **Use**: track authentication success/failure rate

### `auth_login_duration_seconds` (Histogram)
Login request duration in seconds.
- **Labels**: `status` (success, failure)
- **Buckets**: Prometheus defaults
- **Use**: detect slowdowns in the authentication flow

## Registration Metrics

### `auth_registration_attempts_total` (Counter)
Total registration attempts.
- **Labels**: `status` (success, failure, conflict)
- **Use**: track new-registration success rate

### `auth_registration_duration_seconds` (Histogram)
Registration request duration in seconds.
- **Buckets**: Prometheus defaults
- **Use**: spot slow user creation

## Token Metrics

### `auth_token_validations_total` (Counter)
Total token validations.
- **Labels**: `result` (valid, invalid)
- **Use**: detect a spike of invalid tokens (possible attack)

### `auth_token_generations_total` (Counter)
Total JWTs issued.
- **Use**: measure volume of successful authentications

### `auth_token_generation_duration_seconds` (Histogram)
JWT generation duration in seconds.
- **Buckets**: [.001, .005, .01, .025, .05, .1]
- **Use**: monitor JWT generation performance

## Session Metrics

### `auth_active_sessions` (Gauge)
Currently active sessions.
- **Use**: track concurrent connected users

### `auth_logout_requests_total` (Counter)
Total logout requests.
- **Use**: measure disconnection volume

## User Metrics

### `auth_total_users` (Gauge)
Total registered users.
- **Use**: follow user-base growth

### `auth_user_queries_total` (Counter)
Total user lookups.
- **Labels**: `type` (by_id, by_email, list)
- **Use**: spot API usage patterns

## Database Metrics

### `auth_database_queries_total` (Counter)
Total DB queries executed.
- **Labels**: `operation` (select, insert, update, delete)
- **Use**: monitor DB load

### `auth_database_query_duration_seconds` (Histogram)
DB query duration.
- **Labels**: `operation` (select, insert, update, delete)
- **Buckets**: [.001, .005, .01, .025, .05, .1, .25]
- **Use**: catch slow queries

## Error Metrics

### `auth_errors_total` (Counter)
Total errors.
- **Labels**: `type` (validation, database, auth, internal)
- **Use**: alert on service issues

## HTTP Metrics (Middleware)

### `auth_http_requests_total` (Counter)
Total HTTP requests.
- **Labels**: `method`, `endpoint`, `status`
- **Use**: track traffic and status codes

### `auth_http_request_duration_seconds` (Histogram)
HTTP request duration.
- **Labels**: `method`, `endpoint`
- **Buckets**: Prometheus defaults
- **Use**: spot slow endpoints

## Useful PromQL Queries

```promql
# Login success rate (last 5 minutes)
rate(auth_login_attempts_total{status="success"}[5m]) / rate(auth_login_attempts_total[5m])

# Average login time
histogram_quantile(0.95, rate(auth_login_duration_seconds_bucket[5m]))

# Error rate
rate(auth_errors_total[5m])

# Requests per second
rate(auth_http_requests_total[1m])

# Endpoint p95 latency
histogram_quantile(0.95, rate(auth_http_request_duration_seconds_bucket[5m]))

# Invalid tokens per minute
rate(auth_token_validations_total{result="invalid"}[1m]) * 60
```

## Suggested Alerts

1. **High login failure rate**: `rate(auth_login_attempts_total{status="failure"}[5m]) > 10`
2. **Too many errors**: `rate(auth_errors_total[5m]) > 5`
3. **High latency**: `histogram_quantile(0.95, rate(auth_login_duration_seconds_bucket[5m])) > 1`
4. **Invalid tokens (possible attack)**: `rate(auth_token_validations_total{result="invalid"}[1m]) > 50`
