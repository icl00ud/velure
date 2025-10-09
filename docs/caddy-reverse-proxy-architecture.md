# Caddy Reverse Proxy - Arquitetura

## Visão Geral

Este documento descreve a arquitetura do proxy reverso Caddy implementado no projeto Velure, centralizando o acesso a todos os microserviços através de um único ponto de entrada com TLS automático.

## Diagrama de Arquitetura

```mermaid
graph TB
    subgraph "Cliente"
        Browser["🌐 Browser<br/>https://velure.local"]
    end

    subgraph "Caddy Reverse Proxy<br/>:80, :443"
        Caddy["🔒 Caddy 2.8<br/>TLS Automático<br/>Let's Encrypt"]
        
        subgraph "Routing Rules"
            Router["Router<br/>Path Matchers"]
        end
        
        subgraph "Middleware"
            TLS["TLS Termination"]
            CORS["CORS Headers"]
            Security["Security Headers<br/>HSTS, CSP, XSS"]
            Logging["Structured Logs<br/>JSON"]
            Health["Health Checks<br/>10s interval"]
        end
    end

    subgraph "Backend Services"
        subgraph "Auth Service :3020"
            Auth["🔐 auth-service<br/>Gin + PostgreSQL<br/>/authentication/*"]
        end
        
        subgraph "Product Service :3010"
            Product["🛍️ product-service<br/>Fiber + MongoDB + Redis<br/>/product/*"]
        end
        
        subgraph "Order Service :3030"
            Order["📦 publish-order-service<br/>net/http + PostgreSQL + RabbitMQ<br/>/* (REST + SSE)"]
        end
        
        subgraph "UI Service :80"
            UI["⚛️ ui-service<br/>React + Vite + Nginx<br/>SPA"]
        end
    end

    Browser -->|"HTTPS"| TLS
    TLS --> CORS
    CORS --> Security
    Security --> Logging
    Logging --> Health
    Health --> Router

    Router -->|"/api/auth/*<br/>strip /api/auth<br/>rewrite /authentication"| Auth
    Router -->|"/api/product/*<br/>strip /api"| Product
    Router -->|"/api/order/*<br/>strip /api/order<br/>SSE ready"| Order
    Router -->|"/*<br/>fallback"| UI

    Auth -.->|"Health Check<br/>/health"| Health
    Product -.->|"Health Check<br/>/health"| Health
    Order -.->|"Health Check<br/>/health"| Health
    UI -.->|"Health Check<br/>/"| Health

    classDef caddy fill:#00ADD8,stroke:#00758F,color:#fff
    classDef service fill:#FF6B6B,stroke:#C92A2A,color:#fff
    classDef ui fill:#61DAFB,stroke:#21A1C4,color:#000
    classDef browser fill:#4CAF50,stroke:#2E7D32,color:#fff
    
    class Caddy,Router,TLS,CORS,Security,Logging,Health caddy
    class Auth,Product,Order service
    class UI ui
    class Browser browser
```

## Fluxo de Requisições

### 1. Autenticação (POST /api/auth/register)

```mermaid
sequenceDiagram
    participant C as Cliente
    participant Caddy as Caddy Proxy
    participant Auth as auth-service:3020

    C->>Caddy: POST https://velure.local/api/auth/register
    Note over Caddy: TLS Termination<br/>CORS + Security Headers
    Caddy->>Caddy: Match @auth_routes<br/>strip_prefix /api/auth
    Caddy->>Caddy: rewrite /authentication{uri}
    Caddy->>Auth: POST http://auth-service:3020/authentication/register
    Auth-->>Caddy: 200 OK {user}
    Caddy-->>C: 200 OK {user}
```

### 2. Listagem de Produtos (GET /api/product/getProductsByPage)

```mermaid
sequenceDiagram
    participant C as Cliente
    participant Caddy as Caddy Proxy
    participant Product as product-service:3010

    C->>Caddy: GET https://velure.local/api/product/getProductsByPage?page=1
    Note over Caddy: TLS + CORS + Security
    Caddy->>Caddy: Match @product_routes<br/>strip_prefix /api
    Caddy->>Product: GET http://product-service:3010/product/getProductsByPage?page=1
    Note over Product: MongoDB + Redis Cache
    Product-->>Caddy: 200 OK {products[]}
    Note over Caddy: Cache-Control: max-age=300
    Caddy-->>C: 200 OK {products[]}
```

### 3. Criação de Pedido (POST /api/order/create-order)

```mermaid
sequenceDiagram
    participant C as Cliente
    participant Caddy as Caddy Proxy
    participant Order as publish-order-service:3030
    participant RabbitMQ as RabbitMQ

    C->>Caddy: POST https://velure.local/api/order/create-order<br/>Authorization: Bearer {token}
    Note over Caddy: TLS + CORS + Auth Validation
    Caddy->>Caddy: Match @order_routes<br/>strip_prefix /api/order
    Caddy->>Order: POST http://publish-order-service:3030/create-order<br/>Authorization: Bearer {token}
    Order->>Order: Validate JWT
    Order->>Order: Save to PostgreSQL
    Order->>RabbitMQ: Publish order event
    Order-->>Caddy: 200 OK {order_id, status}
    Caddy-->>C: 200 OK {order_id, status}
```

### 4. Server-Sent Events (GET /api/order/user/order/status)

```mermaid
sequenceDiagram
    participant C as Cliente
    participant Caddy as Caddy Proxy
    participant Order as publish-order-service:3030
    participant RabbitMQ as RabbitMQ

    C->>Caddy: GET https://velure.local/api/order/user/order/status?id={order_id}
    Note over Caddy: SSE Config:<br/>flush_interval -1<br/>read_timeout 7200s
    Caddy->>Caddy: Match @order_routes<br/>strip_prefix /api/order
    Caddy->>Order: GET http://publish-order-service:3030/user/order/status?id={order_id}
    
    Note over Order,RabbitMQ: Long-lived connection
    
    loop Status Updates
        RabbitMQ->>Order: Order status changed
        Order-->>Caddy: event: status<br/>data: {status: "PROCESSING"}
        Caddy-->>C: event: status<br/>data: {status: "PROCESSING"}
    end
```

## Configuração de Rotas

| Path Frontend | Caddy Matcher | Strip Prefix | Upstream Service | Endpoint Final |
|--------------|---------------|--------------|------------------|----------------|
| `/api/auth/register` | `@auth_routes path /api/auth/*` | `/api/auth` + rewrite `/authentication` | `auth-service:3020` | `/authentication/register` |
| `/api/auth/login` | `@auth_routes path /api/auth/*` | `/api/auth` + rewrite `/authentication` | `auth-service:3020` | `/authentication/login` |
| `/api/product/getProductsByPage` | `@product_routes path /api/product/*` | `/api` | `product-service:3010` | `/product/getProductsByPage` |
| `/api/order/create-order` | `@order_routes path /api/order/*` | `/api/order` | `publish-order-service:3030` | `/create-order` |
| `/api/order/user/orders` | `@order_routes path /api/order/*` | `/api/order` | `publish-order-service:3030` | `/user/orders` |
| `/api/order/user/order/status` | `@order_routes path /api/order/*` | `/api/order` | `publish-order-service:3030` | `/user/order/status` (SSE) |
| `/` | Fallback | - | `ui-service:80` | `/` (SPA) |

## Funcionalidades Implementadas

### 🔒 Segurança

- **TLS Automático**: Let's Encrypt para produção, certificados locais para desenvolvimento
- **Security Headers**:
  - `Strict-Transport-Security`: HSTS com 1 ano
  - `X-Content-Type-Options`: nosniff
  - `X-Frame-Options`: SAMEORIGIN
  - `X-XSS-Protection`: 1; mode=block
  - `Content-Security-Policy`: CSP configurado
  - `Referrer-Policy`: strict-origin-when-cross-origin
- **CORS**: Headers configurados para permitir cross-origin de forma controlada
- **JWT Validation**: Middleware de autenticação nos serviços

### 📊 Observabilidade

- **Structured Logging**: Logs em JSON com campos estruturados
- **Log Rotation**: 
  - Global: 100MB por arquivo, mantém 5 arquivos
  - Por domínio: 50MB por arquivo, mantém 3 arquivos
- **Health Checks**: 
  - Intervalo: 10 segundos
  - Timeout: 5 segundos
  - Status esperado: 2xx

### ⚡ Performance

- **Compression**: gzip, zstd, brotli automático
- **Cache Headers**: Cache de 5 minutos para produtos
- **Connection Pooling**: Mantém conexões com upstreams
- **SSE Optimization**: 
  - `flush_interval -1`: Flush imediato
  - Long timeouts: 2 horas para conexões SSE
  - `X-Accel-Buffering: no`: Desabilita buffering

### 🎯 Resiliência

- **Health Checks Ativos**: Monitora saúde dos upstreams
- **Graceful Degradation**: Remove upstreams não saudáveis do pool
- **Timeouts Configurados**: 
  - Dial: 10s
  - Response: 30s (auth), 7200s (SSE)
- **Error Pages**: Páginas de erro customizadas em JSON

## Variáveis de Ambiente (Frontend)

```bash
# docker-compose.yaml
VITE_PRODUCT_SERVICE_URL=/api/product
VITE_AUTHENTICATION_SERVICE_URL=/api/auth
VITE_ORDER_SERVICE_URL=/api/order
```

Todas as URLs são relativas, permitindo que o frontend use automaticamente o mesmo domínio (velure.local) para chamadas de API.

## Portas

| Serviço | Porta Interna | Porta Externa | Exposta? |
|---------|---------------|---------------|----------|
| Caddy | 80, 443 | 80, 443 | ✅ Sim |
| auth-service | 3020 | - | ❌ Não |
| product-service | 3010 | - | ❌ Não |
| publish-order-service | 3030 | - | ❌ Não |
| ui-service | 80 | - | ❌ Não |

**Benefício**: Apenas Caddy expõe portas, todos os outros serviços são acessíveis apenas via proxy reverso, aumentando a segurança.

## Deployment

### Desenvolvimento

```bash
# Inicia todos os serviços incluindo Caddy
docker compose up -d

# Acessa via HTTPS (certificados auto-assinados)
https://velure.local
```

### Produção (EKS)

1. **Domínio Real**: Configurar DNS apontando para Load Balancer
2. **Let's Encrypt**: Remover `local_certs` do Caddyfile
3. **Ingress Controller**: Caddy como Ingress ou usar AWS ALB
4. **Secrets**: Externalizar via External Secrets Operator
5. **Observability**: Integrar com Prometheus/Grafana/Loki

## Próximos Passos

- [ ] Rate limiting por IP/endpoint
- [ ] WAF (Web Application Firewall)
- [ ] Prometheus metrics via `/metrics`
- [ ] Distributed tracing (OpenTelemetry)
- [ ] A/B testing via path rewrite
- [ ] Canary deployments
- [ ] Blue/Green deployments via weighted load balancing
