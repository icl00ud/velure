# Velure - Arquitetura Completa do Sistema

## Diagrama de Arquitetura Completa

```mermaid
graph TB
    subgraph "Cliente / Browser"
        Browser["🌐 Browser<br/>Chrome/Firefox/Safari"]
    end

    subgraph "DNS & Load Balancer"
        DNS["🌍 DNS<br/>velure.local<br/>→ 127.0.0.1"]
    end

    subgraph "Caddy Reverse Proxy :80, :443"
        direction TB
        Caddy["🔒 Caddy 2.8-alpine<br/>TLS Termination<br/>Let's Encrypt / Local Certs"]
        
        subgraph "Middleware Pipeline"
            M1["① TLS Termination"]
            M2["② CORS Headers"]
            M3["③ Security Headers<br/>HSTS, CSP, XSS"]
            M4["④ Structured Logging"]
            M5["⑤ Health Checks<br/>10s interval"]
            M6["⑥ Compression<br/>gzip, zstd, brotli"]
        end
        
        Router["🔀 Path Router<br/>Matchers + Strip Prefix"]
    end

    subgraph "Application Layer - Docker Network: velure-network"
        direction TB
        
        subgraph "Frontend - UI Service :80"
            UI["⚛️ ui-service<br/>━━━━━━━━━━━<br/>React 18 + TypeScript<br/>Vite 5 + Bun<br/>TailwindCSS + shadcn/ui<br/>Nginx Alpine<br/>━━━━━━━━━━━<br/>SPA with Client-Side Routing"]
        end
        
        subgraph "Auth Service :3020 - Network: auth, order, frontend"
            Auth["🔐 auth-service<br/>━━━━━━━━━━━<br/>Golang 1.24<br/>Gin Framework<br/>JWT + Sessions<br/>━━━━━━━━━━━<br/>Endpoints:<br/>/authentication/register<br/>/authentication/login<br/>/authentication/validateToken<br/>/authentication/users"]
        end
        
        subgraph "Product Service :3010 - Network: order, frontend"
            Product["🛍️ product-service<br/>━━━━━━━━━━━<br/>Golang 1.24<br/>Fiber Framework<br/>Clean Architecture<br/>━━━━━━━━━━━<br/>Endpoints:<br/>/product/getProductsByPage<br/>/product/getById<br/>/product/search"]
        end
        
        subgraph "Order Services"
            direction LR
            
            PublishOrder["📦 publish-order-service :3030<br/>━━━━━━━━━━━<br/>Golang 1.24<br/>net/http<br/>Clean Architecture<br/>━━━━━━━━━━━<br/>REST Endpoints:<br/>/create-order<br/>/update-order-status<br/>/user/orders<br/>/user/order<br/>━━━━━━━━━━━<br/>SSE Endpoint:<br/>/user/order/status"]
            
            ProcessOrder["⚙️ process-order-service :3040<br/>━━━━━━━━━━━<br/>Golang 1.24<br/>RabbitMQ Consumer<br/>Worker Pool (3 workers)<br/>━━━━━━━━━━━<br/>Processes:<br/>- Order validation<br/>- Payment simulation<br/>- Status updates"]
        end
    end

    subgraph "Data Layer"
        direction TB
        
        subgraph "PostgreSQL Cluster"
            Postgres["🐘 PostgreSQL 17<br/>━━━━━━━━━━━<br/>Database: velure_db<br/>Port: 5432<br/>━━━━━━━━━━━<br/>Tables:<br/>- users (auth)<br/>- sessions (auth)<br/>- orders (orders)<br/>- order_items (orders)<br/>━━━━━━━━━━━<br/>Migrations: golang-migrate"]
        end
        
        subgraph "MongoDB Cluster"
            Mongo["🍃 MongoDB 8<br/>━━━━━━━━━━━<br/>Database: velure_database<br/>Port: 27017<br/>━━━━━━━━━━━<br/>Collections:<br/>- products (34 docs)<br/>- categories<br/>━━━━━━━━━━━<br/>Realistic pet shop data<br/>Images from Unsplash"]
        end
        
        subgraph "Redis Cache"
            Redis["🔴 Redis 7-alpine<br/>━━━━━━━━━━━<br/>Port: 6379<br/>━━━━━━━━━━━<br/>Caching:<br/>- Product queries<br/>- Session data<br/>- Rate limiting<br/>━━━━━━━━━━━<br/>TTL: 5-15 minutes"]
        end
    end

    subgraph "Message Queue"
        RabbitMQ["🐰 RabbitMQ 4-management<br/>━━━━━━━━━━━<br/>Port: 5672 (AMQP)<br/>Port: 15672 (Management UI)<br/>━━━━━━━━━━━<br/>Exchange: orders (topic)<br/>━━━━━━━━━━━<br/>Queues:<br/>- publish-order-status-updates<br/>- process-order-queue<br/>━━━━━━━━━━━<br/>Virtual Host: rabbitmq<br/>Users: admin, publisher-order, process-payment"]
    end

    subgraph "Observability Stack"
        direction LR
        
        Prometheus["📊 Prometheus<br/>(Future)<br/>━━━━━━━━━━━<br/>Metrics Collection<br/>Service Discovery<br/>AlertManager"]
        
        Grafana["📈 Grafana<br/>(Future)<br/>━━━━━━━━━━━<br/>Dashboards<br/>Alerts<br/>Logs + Traces"]
        
        Loki["📝 Loki<br/>(Future)<br/>━━━━━━━━━━━<br/>Log Aggregation<br/>JSON Logs"]
        
        Jaeger["🔍 Jaeger<br/>(Future)<br/>━━━━━━━━━━━<br/>Distributed Tracing<br/>OpenTelemetry"]
    end

    %% Client Flow
    Browser -->|"HTTPS Request"| DNS
    DNS -->|"Resolve to 127.0.0.1"| Caddy
    
    %% Caddy Middleware Pipeline
    Caddy --> M1
    M1 --> M2
    M2 --> M3
    M3 --> M4
    M4 --> M5
    M5 --> M6
    M6 --> Router

    %% Routing
    Router -->|"Path: /*<br/>Fallback"| UI
    Router -->|"Path: /api/auth/*<br/>Strip: /api/auth<br/>Rewrite: /authentication"| Auth
    Router -->|"Path: /api/product/*<br/>Strip: /api"| Product
    Router -->|"Path: /api/order/*<br/>Strip: /api/order"| PublishOrder

    %% Frontend to Backend (via Caddy)
    UI -.->|"API Calls<br/>VITE_*_SERVICE_URL=/api/*"| Router

    %% Auth Service Connections
    Auth -->|"SQL Queries<br/>Users, Sessions"| Postgres
    
    %% Product Service Connections
    Product -->|"Read Products<br/>Aggregations"| Mongo
    Product -->|"Cache GET/SET<br/>TTL: 5min"| Redis

    %% Order Services Connections
    PublishOrder -->|"Save Orders<br/>CRUD Operations"| Postgres
    PublishOrder -->|"Publish Events<br/>order.created<br/>order.updated"| RabbitMQ
    ProcessOrder -->|"Subscribe to Queue<br/>process-order-queue"| RabbitMQ
    ProcessOrder -->|"Update Order Status"| RabbitMQ
    RabbitMQ -->|"Status Updates<br/>via SSE"| PublishOrder

    %% Health Checks (dotted lines)
    M5 -.->|"HTTP GET /health<br/>Every 10s"| Auth
    M5 -.->|"HTTP GET /health<br/>Every 10s"| Product
    M5 -.->|"HTTP GET /health<br/>Every 10s"| PublishOrder
    M5 -.->|"HTTP GET /<br/>Every 30s"| UI

    %% Observability (Future - dotted)
    Caddy -.->|"Metrics Endpoint<br/>/metrics"| Prometheus
    Auth -.->|"App Metrics"| Prometheus
    Product -.->|"App Metrics"| Prometheus
    PublishOrder -.->|"App Metrics"| Prometheus
    ProcessOrder -.->|"App Metrics"| Prometheus
    
    Prometheus -.->|"Query + Alerts"| Grafana
    Loki -.->|"Logs Query"| Grafana
    Jaeger -.->|"Trace Visualization"| Grafana
    
    Caddy -.->|"JSON Logs<br/>/var/log/caddy/"| Loki
    Auth -.->|"Structured Logs"| Loki
    Product -.->|"Structured Logs"| Loki
    PublishOrder -.->|"Structured Logs"| Loki
    ProcessOrder -.->|"Structured Logs"| Loki

    %% Styling
    classDef proxy fill:#00ADD8,stroke:#00758F,stroke-width:3px,color:#fff
    classDef frontend fill:#61DAFB,stroke:#21A1C4,stroke-width:2px,color:#000
    classDef backend fill:#FF6B6B,stroke:#C92A2A,stroke-width:2px,color:#fff
    classDef database fill:#4DB33D,stroke:#2E7D32,stroke-width:2px,color:#fff
    classDef queue fill:#FF9933,stroke:#CC6600,stroke-width:2px,color:#000
    classDef monitor fill:#9B59B6,stroke:#6C3483,stroke-width:2px,color:#fff
    classDef middleware fill:#3498DB,stroke:#2471A3,stroke-width:1px,color:#fff
    classDef client fill:#2ECC71,stroke:#229954,stroke-width:2px,color:#fff
    
    class Browser,DNS client
    class Caddy,Router proxy
    class M1,M2,M3,M4,M5,M6 middleware
    class UI frontend
    class Auth,Product,PublishOrder,ProcessOrder backend
    class Postgres,Mongo,Redis database
    class RabbitMQ queue
    class Prometheus,Grafana,Loki,Jaeger monitor
```

## Descrição dos Componentes

### 🌐 Camada de Cliente
- **Browser**: Interface do usuário acessando via HTTPS
- **DNS**: Resolução local (velure.local → 127.0.0.1)

### 🔒 Camada de Proxy (Caddy)
- **TLS Termination**: Certificados auto-assinados (dev) ou Let's Encrypt (prod)
- **Middleware Pipeline**: 6 etapas de processamento
- **Path Router**: Distribuição inteligente de requisições

### ⚛️ Camada de Aplicação

#### Frontend (ui-service)
- React 18 com TypeScript
- Build: Vite 5 + Bun
- Styling: TailwindCSS + shadcn/ui
- Server: Nginx Alpine
- Container: ~50MB (multi-stage build)

#### Backend Services

**auth-service** (Port 3020)
- Framework: Gin (Golang)
- Autenticação: JWT + Session
- Database: PostgreSQL
- Networks: auth, order, frontend

**product-service** (Port 3010)
- Framework: Fiber (Golang)
- Cache: Redis (TTL 5min)
- Database: MongoDB
- Networks: order, frontend

**publish-order-service** (Port 3030)
- Framework: net/http (Golang)
- REST API + SSE (Server-Sent Events)
- Database: PostgreSQL
- Message Bus: RabbitMQ (Publisher)
- Networks: order

**process-order-service** (Port 3040)
- Framework: Consumer worker pool
- Workers: 3 concurrent goroutines
- Database: PostgreSQL
- Message Bus: RabbitMQ (Consumer)
- Networks: order

### 💾 Camada de Dados

**PostgreSQL 17**
- Databases: velure_db
- Used by: auth-service, publish-order-service, process-order-service
- Migrations: golang-migrate/migrate
- Connection Pool: max 25 connections

**MongoDB 8**
- Database: velure_database
- Used by: product-service
- Collections: products (34 documents), categories
- Data: Realistic pet shop products

**Redis 7**
- Used by: product-service
- Purpose: Query caching, session storage
- TTL: 5-15 minutes
- Eviction: LRU

### 📨 Camada de Mensageria

**RabbitMQ 4**
- Port: 5672 (AMQP), 15672 (Management UI)
- Exchange: orders (topic exchange)
- Queues:
  - `publish-order-status-updates`: Status updates para SSE
  - `process-order-queue`: Pedidos para processamento
- Users:
  - `admin`: Gerenciamento
  - `publisher-order`: Publicação de eventos
  - `process-payment`: Consumo de eventos

### 📊 Camada de Observabilidade (Planejado)

**Prometheus**
- Coleta de métricas de todos os serviços
- Service discovery automático
- AlertManager para notificações

**Grafana**
- Dashboards unificados
- Visualização de métricas, logs e traces
- Alertas configuráveis

**Loki**
- Agregação de logs estruturados (JSON)
- Integração com Grafana

**Jaeger**
- Distributed tracing
- OpenTelemetry integration

## Fluxo de Dados

### 1️⃣ Registro de Usuário
```
Browser → Caddy (TLS) → auth-service → PostgreSQL
                                      ↓
                                    JWT Token
                                      ↓
                                    Browser
```

### 2️⃣ Busca de Produtos
```
Browser → Caddy → product-service → Redis (cache miss)
                                   ↓
                                 MongoDB
                                   ↓
                                 Redis (cache set)
                                   ↓
                                 Browser
```

### 3️⃣ Criação de Pedido
```
Browser → Caddy → publish-order-service → PostgreSQL (save order)
                                        ↓
                                      RabbitMQ (publish event)
                                        ↓
                                  process-order-service
                                        ↓
                                  PostgreSQL (update status)
                                        ↓
                                  RabbitMQ (publish status)
                                        ↓
                                  publish-order-service
                                        ↓
                                  Browser (SSE)
```

## Portas e Exposição

| Serviço | Porta Interna | Porta Externa | Exposta? | Acesso |
|---------|---------------|---------------|----------|--------|
| Caddy | 80, 443 | 80, 443 | ✅ Sim | 0.0.0.0 |
| Caddy Admin | 2019 | - | ❌ Não | localhost |
| auth-service | 3020 | - | ❌ Não | Via Caddy |
| product-service | 3010 | - | ❌ Não | Via Caddy |
| publish-order-service | 3030 | - | ❌ Não | Via Caddy |
| process-order-service | 3040 | - | ❌ Não | Interno |
| ui-service | 80 | - | ❌ Não | Via Caddy |
| PostgreSQL | 5432 | - | ❌ Não | Interno |
| MongoDB | 27017 | - | ❌ Não | Interno |
| Redis | 6379 | - | ❌ Não | Interno |
| RabbitMQ AMQP | 5672 | - | ❌ Não | Interno |
| RabbitMQ Management | 15672 | 15672 | ✅ Sim | 0.0.0.0 |

## Redes Docker

| Rede | Serviços | Propósito |
|------|----------|-----------|
| `velure-network` | Todos | Rede principal |
| `auth` | auth-service, postgres | Isolamento auth |
| `order` | publish-order, process-order, auth, product, postgres, rabbitmq | Isolamento orders |
| `frontend` | ui-service, auth, product | Isolamento frontend |

## Volumes Persistentes

| Volume | Tipo | Serviço | Dados |
|--------|------|---------|-------|
| `postgres_data` | Volume | PostgreSQL | Databases, WAL logs |
| `mongodb_data` | Volume | MongoDB | Collections, indexes |
| `rabbitmq_data` | Volume | RabbitMQ | Queues, messages |
| `redis_data` | Volume | Redis | Cache, snapshots |
| `caddy_data` | Volume | Caddy | Certificados TLS |
| `caddy_config` | Volume | Caddy | Configurações dinâmicas |
| `caddy_logs` | Bind mount | Caddy | Access logs, error logs |

## Recursos e Limites

| Serviço | CPU Limit | Memory Limit | CPU Reserved | Memory Reserved |
|---------|-----------|--------------|--------------|-----------------|
| auth-service | 0.25 | 128M | 0.1 | 64M |
| product-service | 0.25 | 128M | 0.1 | 64M |
| publish-order-service | 0.5 | 256M | 0.25 | 128M |
| process-order-service | 0.5 | 512M | 0.25 | 256M |
| ui-service | - | - | - | - |
| PostgreSQL | 1.0 | 1G | 0.5 | 512M |
| MongoDB | 1.0 | 1G | 0.5 | 512M |
| Redis | 0.25 | 256M | 0.1 | 128M |
| RabbitMQ | 0.5 | 512M | 0.25 | 256M |

## Variáveis de Ambiente Críticas

```bash
# Auth Service
JWT_SECRET=***
JWT_REFRESH_TOKEN_SECRET=***
SESSION_SECRET=***

# Databases
POSTGRES_URL=postgres://velure_user:***@postgres:5432/velure_db
MONGODB_HOST=mongodb
REDIS_URL=redis://redis:6379

# RabbitMQ
PUBLISHER_RABBITMQ_URL=amqp://publisher-order:***@rabbitmq:5672/
PROCESS_RABBITMQ_URL=amqp://process-payment:***@rabbitmq:5672/

# Frontend (Build time)
VITE_PRODUCT_SERVICE_URL=/api/product
VITE_AUTHENTICATION_SERVICE_URL=/api/auth
VITE_ORDER_SERVICE_URL=/api/order
```

## Segurança

### ✅ Implementado
- TLS 1.3 obrigatório
- Security Headers (HSTS, CSP, XSS Protection)
- JWT com refresh tokens
- CORS configurado
- Health checks ativos
- Containers não-root
- Secrets via environment variables
- Redes Docker isoladas

### 🔄 Planejado
- Rate limiting (Caddy)
- WAF (Web Application Firewall)
- Secrets management (External Secrets Operator)
- Network policies (Kubernetes)
- Pod security policies
- Image scanning (Trivy)
- RBAC (Kubernetes)

## Performance

### Cache Strategy
- **Redis**: Product queries (5min TTL)
- **Caddy**: Static assets (browser cache)
- **CDN**: Future - Cloudflare/CloudFront

### Connection Pooling
- **PostgreSQL**: Max 25 connections por serviço
- **MongoDB**: Max 100 connections
- **Redis**: Pool de 10 conexões

### Compression
- **Caddy**: gzip, zstd, brotli automático
- **Nginx**: gzip level 6

## Escalabilidade

### Horizontal Scaling Ready
- ✅ auth-service: Stateless
- ✅ product-service: Stateless + Redis cache
- ✅ publish-order-service: Stateless
- ✅ process-order-service: Worker pool (pode escalar workers)
- ✅ ui-service: Stateless

### Database Scaling
- **PostgreSQL**: Read replicas (futuro)
- **MongoDB**: Sharding (futuro)
- **Redis**: Redis Cluster (futuro)
- **RabbitMQ**: Cluster (futuro)

## Deployment

### Desenvolvimento
```bash
# Iniciar stack completo
docker compose up -d

# Verificar saúde
curl -k https://velure.local/health

# Acessar
open https://velure.local
```

### Produção (AWS EKS)
1. Cluster EKS com node groups
2. Helm charts para cada serviço
3. AWS RDS PostgreSQL (Multi-AZ)
4. DocumentDB (MongoDB compatible)
5. ElastiCache Redis (Cluster mode)
6. Amazon MQ (RabbitMQ)
7. ALB Ingress Controller + Caddy
8. External Secrets Operator
9. Prometheus + Grafana + Loki stack
10. CloudWatch Logs integration

---

**Última atualização**: 2025-10-09  
**Versão**: 1.0.0  
**Status**: ✅ Em desenvolvimento
