# Auth Service - Prometheus Metrics

Este documento descreve todas as métricas customizadas expostas pelo auth-service no endpoint `/metrics`.

## Métricas de Login

### `auth_login_attempts_total` (Counter)
Número total de tentativas de login.
- **Labels**: `status` (success, failure)
- **Uso**: Monitorar taxa de sucesso/falha de autenticação

### `auth_login_duration_seconds` (Histogram)
Duração das requisições de login em segundos.
- **Labels**: `status` (success, failure)
- **Buckets**: Default Prometheus
- **Uso**: Detectar lentidão no processo de autenticação

## Métricas de Registro

### `auth_registration_attempts_total` (Counter)
Número total de tentativas de registro.
- **Labels**: `status` (success, failure, conflict)
- **Uso**: Monitorar taxa de sucesso de novos registros

### `auth_registration_duration_seconds` (Histogram)
Duração das requisições de registro em segundos.
- **Buckets**: Default Prometheus
- **Uso**: Identificar lentidão na criação de usuários

## Métricas de Tokens

### `auth_token_validations_total` (Counter)
Número total de validações de token.
- **Labels**: `result` (valid, invalid)
- **Uso**: Detectar alto volume de tokens inválidos (possível ataque)

### `auth_token_generations_total` (Counter)
Número total de tokens JWT gerados.
- **Uso**: Rastrear volume de autenticações bem-sucedidas

### `auth_token_generation_duration_seconds` (Histogram)
Duração da geração de tokens em segundos.
- **Buckets**: [.001, .005, .01, .025, .05, .1]
- **Uso**: Monitorar performance da geração de JWT

## Métricas de Sessões

### `auth_active_sessions` (Gauge)
Número atual de sessões ativas.
- **Uso**: Monitorar carga de usuários conectados

### `auth_logout_requests_total` (Counter)
Número total de requisições de logout.
- **Uso**: Rastrear volume de desconexões

## Métricas de Usuários

### `auth_total_users` (Gauge)
Número total de usuários registrados.
- **Uso**: Acompanhar crescimento da base de usuários

### `auth_user_queries_total` (Counter)
Número total de consultas de usuário.
- **Labels**: `type` (by_id, by_email, list)
- **Uso**: Identificar padrões de uso da API

## Métricas de Banco de Dados

### `auth_database_queries_total` (Counter)
Número total de queries executadas no banco.
- **Labels**: `operation` (select, insert, update, delete)
- **Uso**: Monitorar volume de operações no banco

### `auth_database_query_duration_seconds` (Histogram)
Duração das queries no banco de dados.
- **Labels**: `operation` (select, insert, update, delete)
- **Buckets**: [.001, .005, .01, .025, .05, .1, .25]
- **Uso**: Detectar queries lentas

## Métricas de Erros

### `auth_errors_total` (Counter)
Número total de erros.
- **Labels**: `type` (validation, database, auth, internal)
- **Uso**: Alertar sobre problemas no serviço

## Métricas HTTP (Middleware)

### `auth_http_requests_total` (Counter)
Número total de requisições HTTP.
- **Labels**: `method`, `endpoint`, `status`
- **Uso**: Monitorar tráfego e status codes

### `auth_http_request_duration_seconds` (Histogram)
Duração das requisições HTTP.
- **Labels**: `method`, `endpoint`
- **Buckets**: Default Prometheus
- **Uso**: Identificar endpoints lentos

## Queries PromQL Úteis

```promql
# Taxa de sucesso de login (últimos 5 minutos)
rate(auth_login_attempts_total{status="success"}[5m]) / rate(auth_login_attempts_total[5m])

# Tempo médio de login
histogram_quantile(0.95, rate(auth_login_duration_seconds_bucket[5m]))

# Taxa de erros
rate(auth_errors_total[5m])

# Requests por segundo
rate(auth_http_requests_total[1m])

# Latência p95 dos endpoints
histogram_quantile(0.95, rate(auth_http_request_duration_seconds_bucket[5m]))

# Tokens inválidos por minuto
rate(auth_token_validations_total{result="invalid"}[1m]) * 60
```

## Alertas Sugeridos

1. **Alta taxa de falhas de login**: `rate(auth_login_attempts_total{status="failure"}[5m]) > 10`
2. **Muitos erros**: `rate(auth_errors_total[5m]) > 5`
3. **Latência alta**: `histogram_quantile(0.95, rate(auth_login_duration_seconds_bucket[5m])) > 1`
4. **Tokens inválidos (possível ataque)**: `rate(auth_token_validations_total{result="invalid"}[1m]) > 50`
