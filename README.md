# Velure - Cloud-Native E-Commerce Platform

<div align="center">

![Velure Architecture](https://img.shields.io/badge/Architecture-Microservices-blue)
![Infrastructure](https://img.shields.io/badge/Infrastructure-AWS_EKS-orange)
![IaC](https://img.shields.io/badge/IaC-Terraform-purple)
![Orchestration](https://img.shields.io/badge/Orchestration-Kubernetes-326CE5)
![CI/CD](https://img.shields.io/badge/CI%2FCD-GitHub_Actions-2088FF)
![Monitoring](https://img.shields.io/badge/Monitoring-Prometheus%20%2B%20Grafana-E6522C)

**Plataforma de e-commerce construída como projeto de aprendizado para demonstrar práticas modernas de DevOps, Cloud-Native Architecture e Site Reliability Engineering (SRE)**

[Arquitetura](#-arquitetura-de-microserviços) • [Infraestrutura](#%EF%B8%8F-infraestrutura-como-código-iac) • [CI/CD](#-cicd-pipeline) • [Monitoramento](#-observabilidade--monitoramento) • [Quick Start](#-quick-start)

</div>

---

## 📋 Índice

- [Visão Geral do Projeto](#-visão-geral-do-projeto)
- [Arquitetura de Microserviços](#-arquitetura-de-microserviços)
- [Stack Tecnológica](#-stack-tecnológica)
- [Infraestrutura como Código (IaC)](#%EF%B8%8F-infraestrutura-como-código-iac)
- [CI/CD Pipeline](#-cicd-pipeline)
- [Observabilidade & Monitoramento](#-observabilidade--monitoramento)
- [Segurança & DevSecOps](#-segurança--devsecops)
- [Padrões de Comunicação](#-padrões-de-comunicação)
- [Quick Start](#-quick-start)
- [Deployment](#-deployment)
- [Load Testing](#-load-testing)
- [Automação com Makefile](#-automação-com-makefile)

---

## 🎯 Visão Geral do Projeto

Velure é uma **plataforma de e-commerce cloud-native** desenvolvida para demonstrar as melhores práticas de:

- **DevOps**: Automação completa do ciclo de vida de desenvolvimento, testes e deployment
- **Cloud-Native Architecture**: Aplicação projetada desde o início para rodar em ambientes cloud
- **Infrastructure as Code (IaC)**: Toda infraestrutura versionada e reproduzível via Terraform
- **Microservices**: Arquitetura de serviços independentes, escaláveis e resilientes
- **GitOps**: Deploy automatizado via Git com workflows declarativos
- **Observabilidade**: Monitoramento, logs e métricas com stack Prometheus/Grafana
- **Site Reliability Engineering (SRE)**: Alta disponibilidade, auto-scaling e disaster recovery

> **Objetivo Principal**: Este não é apenas um e-commerce, mas uma **plataforma de referência** para práticas modernas de engenharia de software e operações cloud.

---

## 🏗️ Arquitetura de Microserviços

### Visão Arquitetural

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          AWS Cloud / EKS Cluster                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐         ┌──────────────────────────────────────┐    │
│  │   Route53    │────────▶│     ALB (Ingress Controller)         │    │
│  │ DNS + Health │         │   - TLS Termination                  │    │
│  └──────────────┘         │   - Path-based Routing               │    │
│                           │   - Health Checks                     │    │
│                           └─────────────┬────────────────────────┘    │
│                                         │                              │
│  ┌──────────────────────────────────────┼────────────────────────┐    │
│  │              Microservices Layer     │                        │    │
│  │                                      │                        │    │
│  │  ┌─────────────┐   ┌────────────┐   │   ┌──────────────┐     │    │
│  │  │ Auth        │   │ Product    │   │   │ Publish-Order│     │    │
│  │  │ Service     │   │ Service    │   │   │ Service      │     │    │
│  │  │ Go + Gin    │   │ Go + Fiber │   │   │ Go + SSE     │     │    │
│  │  │ JWT + OAuth │   │ MongoDB    │   │   │ PostgreSQL   │     │    │
│  │  └──────┬──────┘   └──────┬─────┘   │   └──────┬───────┘     │    │
│  │         │                  │         │          │             │    │
│  │         ▼                  ▼         │          ▼             │    │
│  │  ┌────────────┐   ┌────────────┐    │   ┌──────────────┐     │    │
│  │  │ PostgreSQL │   │   Redis    │    │   │  RabbitMQ    │     │    │
│  │  │  (RDS)     │   │  (Cache)   │    │   │  (AmazonMQ)  │     │    │
│  │  └────────────┘   └────────────┘    │   └──────┬───────┘     │    │
│  │                                      │          │             │    │
│  │                                      │          │ Queue       │    │
│  │                                      │          │ "orders"    │    │
│  │                                      │          ▼             │    │
│  │                                      │   ┌──────────────┐     │    │
│  │                                      │   │ Process-Order│     │    │
│  │                                      │   │ Service      │     │    │
│  │                                      │   │ Async Worker │     │    │
│  │                                      │   └──────────────┘     │    │
│  │                                      │                        │    │
│  │  ┌──────────────────────────────────────────────────────┐     │    │
│  │  │            UI Service (React SPA)                    │     │    │
│  │  │  Vite + TypeScript + TailwindCSS + Radix UI          │     │    │
│  │  └──────────────────────────────────────────────────────┘     │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │              Observability Stack (Monitoring NS)               │    │
│  │                                                                 │    │
│  │  ┌────────────┐   ┌─────────────┐   ┌──────────────┐          │    │
│  │  │ Prometheus │◀──│ ServiceMon  │   │   Grafana    │          │    │
│  │  │  Metrics   │   │ (exporters) │   │  Dashboards  │          │    │
│  │  └────────────┘   └─────────────┘   └──────────────┘          │    │
│  │                                                                 │    │
│  │  ┌────────────┐   ┌─────────────┐                              │    │
│  │  │    Loki    │◀──│  Promtail   │   (Logs aggregation)         │    │
│  │  └────────────┘   └─────────────┘                              │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Características Arquiteturais

#### ✅ **Loose Coupling**
- Serviços independentes com responsabilidades bem definidas
- Comunicação via APIs REST e message queues
- Falha de um serviço não afeta os demais

#### ✅ **High Cohesion**
- Cada serviço gerencia seu próprio banco de dados (Database-per-Service pattern)
- Lógica de negócio concentrada no serviço responsável

#### ✅ **Resilience & Fault Tolerance**
- Health checks configurados (Liveness + Readiness probes)
- Auto-restart de containers falhados
- Circuit breaker patterns para chamadas externas

#### ✅ **Scalability**
- Horizontal Pod Autoscaler (HPA) configurado
- Stateless services (exceto datastores)
- Cache distribuído com Redis

#### ✅ **Event-Driven Architecture**
- RabbitMQ para processamento assíncrono de pedidos
- Desacoplamento entre criação e processamento de orders
- Server-Sent Events (SSE) para updates em tempo real

---

## 🛠️ Stack Tecnológica

### Backend Services

| Componente | Tecnologia | Justificativa |
|-----------|-----------|--------------|
| **Runtime** | Go 1.23+ | Alto performance, baixo consumo de memória, concorrência nativa (goroutines) |
| **Web Frameworks** | Gin (auth) / Fiber (product) | Gin: robusto para auth complexo; Fiber: ultra-rápido para APIs simples |
| **ORM** | GORM | Migrations automáticas, type-safe queries, suporte a transactions |
| **Databases** | PostgreSQL 17 (relacional)<br>MongoDB 6.0 (NoSQL) | PostgreSQL: dados transacionais (auth, orders)<br>MongoDB: catálogo flexível de produtos |
| **Cache** | Redis 8.0 | Cache distribuído, session storage, rate limiting |
| **Message Queue** | RabbitMQ 4.0 | AMQP protocol, reliable message delivery, dead-letter queues |
| **Auth** | JWT + Refresh Tokens | Stateless authentication, escalável, seguro |

### Frontend

| Componente | Tecnologia | Justificativa |
|-----------|-----------|--------------|
| **Framework** | React 18 | Componentização, ecosystem maduro, virtual DOM |
| **Build Tool** | Vite | HMR ultra-rápido, build otimizado |
| **Language** | TypeScript | Type safety, melhor DX, menos bugs em runtime |
| **Styling** | TailwindCSS | Utility-first, design system consistente |
| **Components** | Radix UI | Acessível, composable, headless components |
| **Routing** | React Router v6 | Client-side routing, code splitting |

### Infrastructure & DevOps

#### Containerization
- **Docker**: Multi-stage builds para otimização de imagens
- **Docker Compose**: Desenvolvimento local com hot-reload
- **Registry**: Docker Hub com multi-arch builds (amd64/arm64)

#### Orchestration
- **Kubernetes**: Cluster gerenciado via AWS EKS
- **Helm Charts**: Packaging e deploy declarativo de aplicações
- **Namespace Isolation**: Segmentação lógica (auth, order, product, datastores, monitoring)

#### Infrastructure as Code (IaC)
- **Terraform**: Provisionamento completo da AWS
- **Modules**: VPC, EKS, RDS, AmazonMQ, Route53, Secrets Manager
- **State Management**: Remote state com locking (S3 + DynamoDB)

#### CI/CD
- **GitHub Actions**: Workflows declarativos
- **Path-based Triggers**: Build apenas serviços alterados
- **Reusable Workflows**: DRY principles para pipelines
- **Multi-stage Pipeline**: Test → Build → Scan → Push → Deploy

#### Observability
- **Prometheus**: Metrics collection e alerting
- **Grafana**: Dashboards customizados com 20+ visualizações
- **Loki**: Log aggregation
- **cAdvisor**: Container metrics
- **Node Exporter**: Host-level metrics

#### Security & Scanning
- **Semgrep**: SAST (Static Application Security Testing)
- **Trivy**: Container vulnerability scanning
- **gosec**: Go security scanner
- **Docker Scout**: Supply chain security
- **SonarCloud**: Code quality & security analysis

#### Reverse Proxy & Load Balancing
- **Caddy 2.8** (Local para desenvolvimento): Automatic HTTPS, reverse proxy
- **AWS ALB** (Ambiente de produção): Load Balancer de camada 7, TLS termination

---

## ⚙️ Infraestrutura como Código (IaC)

### Terraform Architecture

```
infrastructure/terraform/
├── main.tf                 # Root module orchestrator
├── variables.tf            # Input variables
├── outputs.tf              # Output values (endpoints, ARNs)
├── versions.tf             # Provider versions
└── modules/
    ├── vpc/                # Network infrastructure
    │   ├── main.tf         # VPC, Subnets, IGW, NAT Gateway
    │   ├── routes.tf       # Route tables
    │   └── outputs.tf      # VPC ID, Subnet IDs
    ├── security-groups/    # Network security
    │   └── main.tf         # SG para EKS, RDS, AmazonMQ
    ├── eks/                # Kubernetes cluster
    │   ├── main.tf         # EKS cluster + Node groups
    │   ├── iam.tf          # IRSA (IAM Roles for Service Accounts)
    │   └── addons.tf       # VPC-CNI, CoreDNS, kube-proxy
    ├── rds/                # Managed PostgreSQL
    │   ├── main.tf         # RDS instances (auth + orders)
    │   └── backups.tf      # Automated backups
    ├── amazonmq/           # Managed RabbitMQ
    │   └── main.tf         # AmazonMQ broker
    ├── route53/            # DNS management
    │   └── main.tf         # Hosted Zone + Records
    └── secrets-manager/    # Centralized secrets
        └── main.tf         # Secrets para DB, JWT, RabbitMQ
```

### Recursos Provisionados na AWS

#### **Networking** (VPC Module)
- VPC com CIDR /16
- 2 Availability Zones para alta disponibilidade
- 2 Public Subnets (para ALB)
- 2 Private Subnets (para EKS nodes, RDS, AmazonMQ)
- Internet Gateway para acesso público
- NAT Gateway para egress privado
- Route Tables customizadas

#### **Compute** (EKS Module)
- EKS Cluster v1.31
- Managed Node Group (t3.medium)
- Auto-scaling (2-4 nodes)
- IAM Roles for Service Accounts (IRSA)
  - ALB Controller: gerenciar Application Load Balancers
  - External Secrets Operator: integrar com AWS Secrets Manager
- Add-ons: VPC-CNI, CoreDNS, kube-proxy

#### **Databases** (RDS Module)
- **RDS Auth**: PostgreSQL 17 para auth-service
- **RDS Orders**: PostgreSQL 17 compartilhado por publish-order e process-order
- Multi-AZ para alta disponibilidade
- Automated backups (7 dias de retenção)
- Encryption at rest

#### **Message Queue** (AmazonMQ Module)
- RabbitMQ gerenciado
- Single-instance (dev) ou Cluster (prod)
- Automatic failover em cluster mode
- CloudWatch logs habilitados

#### **DNS** (Route53 Module)
- Hosted Zone para domínio customizado
- Health checks configuráveis
- Automatic DNS record para ALB

#### **Secrets Management** (Secrets Manager Module)
- Credenciais de banco de dados
- JWT secrets
- RabbitMQ credentials
- Rotation automática habilitável

### Terraform Best Practices Implementadas

> Baseado em: https://www.terraform-best-practices.com/

- **Modularização**: Módulos reutilizáveis e testáveis
- **Remote State**: State armazenado remotamente (configurável)
- **Variable Validation**: Validação de inputs com regras customizadas
- **Output Management**: Outputs estruturados para integração
- **Destroy-time Provisioners**: Cleanup automático de recursos K8s antes de destruir VPC
- **Resource Tagging**: Tags consistentes para billing e organização
- **Dependency Management**: `depends_on` explícito para ordem correta

---

## 🔄 CI/CD Pipeline

### Pipeline Architecture

O pipeline é baseado em **monorepo** com **path-based triggers** para otimizar builds.

```
┌─────────────────────────────────────────────────────────────────────┐
│                      GitHub Actions Workflow                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. DETERMINE CHANGES (Dorny Path Filter)                          │
│     ├─ services/auth-service/**        → trigger: auth pipeline    │
│     ├─ services/product-service/**     → trigger: product pipeline │
│     ├─ services/ui-service/**          → trigger: ui pipeline      │
│     └─ shared/**                        → trigger: ALL Go services │
│                                                                     │
│  2. PARALLEL EXECUTION (Matrix Strategy)                           │
│     ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│     │ Go Service   │  │ Go Service   │  │ Node Service │          │
│     │  Workflow    │  │  Workflow    │  │   Workflow   │          │
│     └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│            │                  │                  │                  │
│            ▼                  ▼                  ▼                  │
│     ┌──────────────────────────────────────────────────┐           │
│     │          TEST & COVERAGE                         │           │
│     │  • go test -coverprofile -race                   │           │
│     │  • npm test (vitest)                             │           │
│     │  • Coverage upload to artifacts                  │           │
│     │  • SonarCloud analysis (quality gate)            │           │
│     └──────────────────┬───────────────────────────────┘           │
│                        │ (only if tests pass)                      │
│                        ▼                                            │
│     ┌──────────────────────────────────────────────────┐           │
│     │          BUILD & PUSH                            │           │
│     │  • Docker Buildx (multi-stage builds)            │           │
│     │  • Tag: branch, PR#, SHA, latest                 │           │
│     │  • Push to Docker Hub                            │           │
│     │  • Cache layers (GitHub Actions cache)           │           │
│     └──────────────────┬───────────────────────────────┘           │
│                        │ (only on master branch)                   │
│                        ▼                                            │
│     ┌──────────────────────────────────────────────────┐           │
│     │          DEPLOY TO EKS                           │           │
│     │  • Update kubeconfig (aws eks)                   │           │
│     │  • Helm upgrade --install                        │           │
│     │  • Rollout status check                          │           │
│     │  • Health check validation                       │           │
│     └──────────────────────────────────────────────────┘           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Workflows Detalhados

#### **1. CI/CD Main Pipeline** (`.github/workflows/ci-cd.yml`)

Orquestrador principal que:
- Detecta mudanças em cada serviço via `dorny/paths-filter`
- Invoca workflows reutilizáveis apenas para serviços alterados
- Reduz build time em ~80% (não builda serviços não modificados)
- Suporta PR reviews e merge automático

#### **2. Go Service Workflow** (`.github/workflows/go-service.yml`)

Workflow reutilizável para todos os serviços Go:

```yaml
jobs:
  test-and-coverage:
    - Checkout com fetch-depth: 0 (para SonarCloud)
    - Setup Go 1.23 com cache de dependências
    - go mod download
    - go test ./... -coverprofile -covermode=atomic -race
    - SonarCloud scan (SAST + quality metrics)
    - Upload coverage artifacts (30 dias de retenção)

  build-and-push:
    needs: test-and-coverage
    - Docker Buildx setup (multi-arch support)
    - Login Docker Hub
    - Extract metadata (tags dinâmicos)
    - Build multi-stage Dockerfile
    - Push com tags: branch, pr-X, sha-abc123, latest
    - Cache Docker layers (GitHub Actions cache)
```

#### **3. Node Service Workflow** (`.github/workflows/node-service.yml`)

Similar ao Go, adaptado para React:
- `npm ci` (clean install)
- `npm run lint` (Biome linting)
- `npm run build` (Vite production build)
- Coverage com Vitest
- Docker build com Nginx

#### **4. Deploy Service Workflow** (`.github/workflows/deploy-service.yml`)

Workflow de deployment para EKS:

```yaml
jobs:
  deploy:
    - Configure AWS credentials (OIDC)
    - Update kubeconfig: aws eks update-kubeconfig
    - Helm upgrade --install \
        --set image.tag=${{ inputs.image-tag }} \
        --wait --timeout 5m
    - kubectl rollout status deployment/velure-${{ inputs.service }}
    - Health check: curl http://service/health
```

### Security Scanning Pipeline

Pipeline adicional para security (`.github/workflows/security-quality.yml`):

```yaml
schedule:
  - cron: '0 2 * * *'  # Daily at 2 AM UTC

jobs:
  semgrep-sast:
    - Semgrep scan (40+ security rules)
    - SARIF upload para GitHub Security tab

  trivy-container-scan:
    - Scan de vulnerabilidades em imagens Docker
    - Block on HIGH/CRITICAL CVEs

  gosec:
    - Go security scanner
    - Check for hardcoded secrets, SQL injection, etc.

  docker-scout:
    - Supply chain security
    - SBOM generation
```

---

## 📊 Observabilidade & Monitoramento

### Stack de Monitoramento

#### **Arquitetura**

```
┌─────────────────────────────────────────────────────────────────────┐
│                       Monitoring Stack                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌───────────────────────────────────────────────────────────┐     │
│  │  Metrics Pipeline                                         │     │
│  │                                                            │     │
│  │  Application  ──┐                                          │     │
│  │  (Prometheus    │                                          │     │
│  │   client libs)  │                                          │     │
│  │                 │                                          │     │
│  │  Node Exporter ─┼──▶ ServiceMonitor ──▶ Prometheus ──┐    │     │
│  │  (host metrics) │    (scrape config)    (TSDB)       │    │     │
│  │                 │                                     │    │     │
│  │  cAdvisor ──────┘                                     │    │     │
│  │  (containers)                                         │    │     │
│  │                                                       │    │     │
│  │                                                       ▼    │     │
│  │                                              ┌─────────────┴─┐   │
│  │                                              │   Grafana     │   │
│  │                                              │  Dashboards   │   │
│  │                                              └───────────────┘   │
│  └───────────────────────────────────────────────────────────┘     │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────┐     │
│  │  Logs Pipeline                                            │     │
│  │                                                            │     │
│  │  Application  ──┐                                          │     │
│  │  (stdout/stderr)│                                          │     │
│  │                 │                                          │     │
│  │  Container ─────┼──▶ Promtail ──▶ Loki ──▶ Grafana        │     │
│  │  logs (Docker)  │    (collector)   (store)  (visualization)│     │
│  │                 │                                          │     │
│  └─────────────────┘                                          │     │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────┐     │
│  │  Alerting                                                 │     │
│  │                                                            │     │
│  │  Prometheus ──▶ AlertManager ──▶ Notification Channels    │     │
│  │  (rules)        (routing)         (Slack, Email, etc.)    │     │
│  └───────────────────────────────────────────────────────────┘     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### Dashboards Grafana

Dashboards customizados com 20+ visualizações:

#### **Velure Overview Dashboard**
- **Request Rate**: Requests/sec por serviço
- **Error Rate**: Taxa de erros HTTP 4xx/5xx
- **Latency**: P50, P95, P99 por endpoint
- **Throughput**: Bytes in/out
- **Active Connections**: Conexões ativas por serviço

#### **Database Performance Dashboard**
- **PostgreSQL**: Connections, queries/sec, cache hit ratio
- **MongoDB**: Operations/sec, document counts, replication lag
- **Redis**: Hit rate, evictions, memory usage

#### **RabbitMQ Dashboard**
- **Queue Depth**: Mensagens pendentes
- **Publish Rate**: Msgs/sec publicadas
- **Consume Rate**: Msgs/sec consumidas
- **Consumer Lag**: Delay no processamento

#### **Infrastructure Dashboard**
- **CPU/Memory**: Uso por node e pod
- **Disk I/O**: IOPS, throughput
- **Network**: Packet loss, bandwidth

#### **SLI/SLO Dashboard**
- **Availability**: Uptime % (target: 99.9%)
- **Latency SLO**: % requests < 500ms (target: 95%)
- **Error Budget**: Budget restante para o mês

### Métricas Customizadas

Cada serviço expõe métricas Prometheus:

```go
// auth-service/internal/middleware/prometheus.go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latencies",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)
```

### Alerting Rules

Alertas configurados para:

- **HighErrorRate**: Error rate > 5% por 5 minutos
- **HighLatency**: P95 latency > 1s por 10 minutos
- **PodCrashLooping**: Pod reiniciando > 3x em 5 minutos
- **HighMemoryUsage**: Memory usage > 90%
- **DiskSpaceRunningOut**: Disk usage > 85%
- **DatabaseConnectionPoolExhausted**: Connections > 90% do pool
- **RabbitMQQueueGrowing**: Queue depth crescendo por 15 minutos

### Deployment Local vs AWS EKS

#### **Local (Docker Compose)**

```bash
# Iniciar aplicação completa + monitoramento
make dev-full

# Acessos:
# - Aplicação: https://velure.local
# - Grafana: http://localhost:3000 (admin/admin)
# - Prometheus: http://localhost:9090
# - RabbitMQ: http://localhost:15672 (admin/admin_password)

# Parar tudo
make dev-stop-full
```

#### **AWS EKS (Production)**

**Fluxo de Deploy**:

```
1. Terraform (Infraestrutura AWS)
   ├─ VPC + Subnets
   ├─ EKS Cluster + Node Groups
   ├─ RDS (PostgreSQL)
   ├─ AmazonMQ (RabbitMQ)
   └─ Route53 + Secrets Manager

2. deploy-eks.sh (Kubernetes)
   ├─ AWS Load Balancer Controller
   ├─ Metrics Server + External Secrets
   ├─ Datastores (MongoDB, Redis)
   ├─ Monitoring (Prometheus + Grafana)
   └─ Velure Services (auth, product, orders, UI)
```

**Comandos Essenciais**:

```bash
# 1. Planejar infraestrutura AWS
make aws-plan

# 2. Provisionar infraestrutura (VPC, EKS, RDS, AmazonMQ)
make aws-deploy
# Aguarde ~15 minutos

# 3. Configurar kubectl para o cluster EKS
make aws-kubeconfig

# 4. Deploy completo Kubernetes (ALB Controller + Helm Charts + Monitoring + Services)
./scripts/deploy-eks.sh
# Aguarde ~10 minutos
# Retorna URL do ALB ao final

# 5. Verificar status do deployment
kubectl get pods -A
kubectl get ingress -A

# 6. Port-forward Grafana (opcional)
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
# Acesse: http://localhost:3000

# 7. Destruir toda a infraestrutura (requer confirmação)
make aws-destroy
```

**Notas Importantes**:
- O script `deploy-eks.sh` é **idempotente** (pode ser executado múltiplas vezes)
- Faz **health checks** automáticos antes de prosseguir para próximas etapas
- Adota secrets existentes para evitar conflitos com Helm
- Limpa releases travadas automaticamente

---

## 🔒 Segurança & DevSecOps

### Camadas de Segurança

#### **1. Application Security**

✅ **Authentication & Authorization**
- JWT com refresh tokens (short-lived access tokens)
- Bcrypt hashing para senhas (cost factor: 12)
- Rate limiting (100 req/min por IP)
- CORS configurado por ambiente

✅ **Input Validation**
- Validação de todos os inputs (struct tags)
- Sanitização de SQL queries (prepared statements)
- Content-Type validation

✅ **Secrets Management**
- Nunca hardcoded em código
- AWS Secrets Manager em produção
- Environment variables em dev
- Rotation automática configurável

#### **2. Infrastructure Security**

✅ **Network Security**
- Security Groups restritivos (least privilege)
- Private subnets para bancos de dados
- Public subnets apenas para ALB
- NACLs configuradas

✅ **IAM Policies**
- IRSA (IAM Roles for Service Accounts)
- Princípio do menor privilégio
- Service accounts por namespace
- Policies granulares (não usar `*` permissions)

✅ **Encryption**
- TLS 1.3 em todas as comunicações
- RDS encryption at rest (AES-256)
- Secrets Manager encryption (KMS)
- HTTPS enforcement via ALB

#### **3. Container Security**

✅ **Image Hardening**
- Multi-stage builds (imagens finais < 50MB)
- Distroless base images (Go services)
- Non-root user (UID 1000)
- Vulnerability scanning (Trivy)

✅ **Runtime Security**
- Read-only root filesystem
- Drop all capabilities
- securityContext configurado
- Resource limits (CPU/Memory)

#### **4. Supply Chain Security**

✅ **Dependency Management**
- Dependabot alerts habilitado
- `go mod tidy` em CI
- npm audit em pipelines
- SBOM generation (Docker Scout)

✅ **Code Scanning**
- SAST: Semgrep (40+ rules)
- Go-specific: gosec
- Quality: SonarCloud (code smells, bugs, vulnerabilities)
- Container: Trivy (CVE scanning)

### Security Scanning no CI/CD

```yaml
# .github/workflows/security-quality.yml
security:
  - Semgrep (OWASP Top 10)
  - Trivy (CVE database)
  - gosec (Go security)
  - Docker Scout (supply chain)
  - SonarCloud (quality + security)
```

### Compliance & Best Practices

✅ CIS Kubernetes Benchmark
✅ OWASP Top 10 coverage
✅ NIST Cybersecurity Framework
✅ Principle of Least Privilege
✅ Defense in Depth

---

## 🔗 Padrões de Comunicação

### 1. Synchronous (HTTP/REST)

- **Frontend ↔ Backend**: Chamadas HTTP para APIs REST
- **process-order ↔ product-service**: Verificação de estoque via HTTP

**Vantagens**:
- Simples de implementar
- Request/response imediato
- Fácil debugging

**Desvantagens**:
- Tight coupling
- Timeout issues
- Não resiliente a falhas

### 2. Asynchronous (Message Queue)

- **publish-order → process-order**: RabbitMQ exchange "orders"

**Vantagens**:
- Loose coupling
- Resiliente a falhas (retry automático)
- Backpressure handling
- Event-driven

**Desvantagens**:
- Eventual consistency
- Mais complexo de debugar
- Overhead de infraestrutura

### 3. Real-time (Server-Sent Events)

- **publish-order → Frontend**: Updates de status de pedido via SSE

**Vantagens**:
- Conexão unidirecional (server → client)
- Auto-reconnect
- Compatível com HTTP/2

**Desvantagens**:
- Apenas server → client
- Não suporta binary data

---

## 🚀 Quick Start

### Pré-requisitos

**Para desenvolvimento local:**
- Docker 24+ & Docker Compose v2
- Make

**Para deployment AWS:**
- Todas as ferramentas acima, mais:
- kubectl 1.31+
- Helm 3.16+
- Terraform 1.9+
- AWS CLI v2

### Configuração Inicial

```bash
# 1. Clone o repositório
git clone https://github.com/icl00ud/velure.git
cd velure

# 2. Configure /etc/hosts (necessário para acesso local)
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts
```

---

### Opção 1: Desenvolvimento Local 🏠

```bash
# Subir aplicação COMPLETA (infra + services + monitoring)
make local-up

# Aguarde ~30 segundos
# ✅ Acesse: https://velure.local (aceite certificado self-signed)
```

**URLs disponíveis:**
- **Aplicação:** https://velure.local
- **Grafana:** http://localhost:3000 (admin/admin)
- **Prometheus:** http://localhost:9090
- **RabbitMQ:** http://localhost:15672 (admin/admin_password)

**Quando terminar:**
```bash
make local-down
```

---

### Opção 2: AWS EKS (Production) ☁️

```bash
# 1. Configurar credenciais AWS
aws configure
# AWS Access Key ID: ***
# AWS Secret Access Key: ***
# Default region: us-east-1

# 2. Subir infraestrutura COMPLETA (Terraform + Kubernetes)
make cloud-up
# ⏳ Aguarde ~25 minutos

# 3. Obter URLs de acesso
make cloud-urls
```

**O que será criado:**
- ✅ VPC + Subnets (multi-AZ)
- ✅ EKS Cluster (2-4 nodes t3.medium)
- ✅ RDS PostgreSQL x2
- ✅ AmazonMQ (RabbitMQ)
- ✅ Datastores (MongoDB, Redis)
- ✅ Monitoring (Prometheus + Grafana)
- ✅ Velure Services (auth, product, orders, UI)

**Quando terminar:**
```bash
make cloud-down
# Digite: DESTROY (confirmação obrigatória)
```

---

## 🌐 Deployment

### Opções de Deployment

| Ambiente | Comando | Tempo Estimado | Custo |
|----------|---------|----------------|-------|
| **Local** | `make local-up` | ~30 seg | $0 |
| **AWS EKS** | `make cloud-up` | ~25 min | ~$150/mês |

---

### Deployment Local

```bash
# Subir aplicação completa
make local-up

# Acesse: https://velure.local
```

**Componentes:**
- PostgreSQL, MongoDB, Redis, RabbitMQ
- Serviços: auth, product, orders, UI
- Monitoring: Prometheus, Grafana, cAdvisor
- Reverse Proxy: Caddy (HTTPS automático)

**Derrubar:**
```bash
make local-down
```

---

### Deployment AWS

#### Um Único Comando

```bash
# Deploy completo automatizado
make cloud-up
```

**O que acontece:**

**Fase 1 - Terraform (~15 min):**
1. Provisiona VPC + Subnets (2 AZs)
2. Cria EKS Cluster + Node Groups
3. Provisiona RDS PostgreSQL x2 (auth + orders)
4. Provisiona AmazonMQ (RabbitMQ)
5. Configura Route53 + Secrets Manager

**Fase 2 - Kubernetes (~10 min):**
1. Instala AWS Load Balancer Controller
2. Instala Metrics Server + External Secrets
3. Deploy datastores via Helm (MongoDB, Redis)
4. Deploy monitoring (Prometheus + Grafana)
5. Deploy Velure services (auth, product, orders, UI)

#### Obter URLs

```bash
make cloud-urls
```

**Exemplo de output:**
```
═══════════════════════════════════════════════════════════════
                    URLs DE ACESSO (AWS)
═══════════════════════════════════════════════════════════════

🌐 Frontend (UI):
   http://k8s-frontend-velureui-xxx.us-east-1.elb.amazonaws.com

📊 Grafana (Observabilidade):
   kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
   Depois acesse: http://localhost:3000 (admin/admin)
```

#### Destruir Tudo

```bash
make cloud-down
# Digite: DESTROY
```

**O que será removido:**
- Todos os recursos Kubernetes (Helm releases, PVCs)
- Secrets Manager (forçado, mesmo pendentes)
- EKS Cluster + Node Groups
- RDS Databases
- AmazonMQ Broker
- VPC + Subnets + NAT Gateway

**Tempo estimado:** ~10 minutos

### Continuous Deployment (GitOps)

O deployment é **automatizado via GitHub Actions**:

```
1. Developer push para master
2. GitHub Actions detecta mudanças (path-based)
3. Pipeline executa:
   ├─ Tests + Coverage
   ├─ Build Docker image
   ├─ Push para Docker Hub
   └─ Deploy para EKS (Helm upgrade)
4. Helm faz rolling update (zero-downtime)
5. Health checks validam deployment
```

**Zero-downtime deployment garantido por**:
- Rolling update strategy (maxUnavailable: 0)
- Readiness probes (serviço só recebe tráfego quando saudável)
- PodDisruptionBudget (mínimo de pods sempre disponíveis)

---

## 📈 Load Testing

### Ferramentas

- **k6**: Ferramenta de load testing moderna (Go-based)
- **Scripts customizados**: Cenários realistas de e-commerce

### Testes Disponíveis

```
tests/load/
├── auth-service-test.js           # Login, registro, token refresh
├── product-service-test.js        # Listagem, busca, detalhes de produtos
├── publish-order-service-test.js  # Criação de pedidos, SSE
├── integrated-load-test.js        # Jornada completa do usuário
└── run-all-tests.sh               # Executa todos os testes sequencialmente
```

### Executando Load Tests

#### Teste Individual

```bash
cd tests/load

# Teste de autenticação
k6 run auth-service-test.js

# Teste de produtos (com cache)
k6 run product-service-test.js

# Teste integrado (user journey completo)
k6 run integrated-load-test.js
```

#### Suite Completa

```bash
# Da raiz do projeto
make test-load

# Ou manualmente
cd tests/load
./run-all-tests.sh
```

### Cenários de Teste

#### **integrated-load-test.js** (User Journey)

Simula jornada completa de compra:

```
1. Registro de usuário
2. Login (obtenção de JWT)
3. Listagem de produtos
4. Busca de produto específico
5. Criação de pedido
6. Monitoramento de status via SSE
```

**Métricas coletadas**:
- Request rate: ~500 RPS
- Error rate: < 1%
- P95 latency: < 800ms
- P99 latency: < 1.5s

**Configuração**:
```javascript
export const options = {
  stages: [
    { duration: '30s', target: 50 },  // Ramp-up
    { duration: '1m', target: 100 },  // Sustained load
    { duration: '30s', target: 0 },   // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<800', 'p(99)<1500'],
    http_req_failed: ['rate<0.01'],  // < 1% errors
  },
};
```

### Resultados Esperados

| Métrica | Target | Atual |
|---------|--------|-------|
| **Throughput** | > 400 RPS | ~500 RPS |
| **Error Rate** | < 1% | 0.3% |
| **P95 Latency** | < 800ms | 650ms |
| **P99 Latency** | < 1.5s | 1.2s |
| **CPU Usage** | < 70% | 55% |
| **Memory** | < 80% | 65% |

### Otimizações Implementadas

✅ **Connection Pooling**: PostgreSQL (max 100 conn), MongoDB (max 50 conn)
✅ **Redis Caching**: Cache de produtos com TTL de 5 minutos
✅ **Bcrypt Workers**: Pool de workers para hash paralelo
✅ **Token Caching**: Cache de tokens JWT válidos
✅ **Database Indexing**: Indexes em colunas frequentemente consultadas
✅ **Compression**: Gzip compression em responses HTTP

---

## 🛠️ Automação com Makefile

O projeto possui **5 comandos essenciais** via Makefile para gerenciar todo o ciclo de vida da aplicação.

### Comandos Disponíveis

```bash
make help              # Mostrar todos os comandos disponíveis
```

#### 🏠 Desenvolvimento Local

```bash
# Subir aplicação COMPLETA localmente (infra + services + monitoring)
make local-up

# Derrubar aplicação local completa (remove containers + volumes)
make local-down
```

**O que `local-up` faz:**
- Cria redes Docker (auth, order, frontend)
- Inicia Docker Compose com:
  - Infraestrutura: PostgreSQL, MongoDB, Redis, RabbitMQ
  - Serviços: auth, product, publish-order, process-order, UI
  - Monitoramento: Prometheus, Grafana, cAdvisor
  - Reverse Proxy: Caddy (HTTPS automático)
- Aguarda inicialização (20 segundos)
- Mostra URLs de acesso

**Acessos após `local-up`:**
- **Aplicação:** https://velure.local
- **Grafana:** http://localhost:3000 (admin/admin)
- **Prometheus:** http://localhost:9090
- **RabbitMQ:** http://localhost:15672 (admin/admin_password)
- **cAdvisor:** http://localhost:8080

---

#### ☁️ Cloud (AWS EKS)

```bash
# Subir infraestrutura COMPLETA na AWS (Terraform + Kubernetes + Monitoring)
make cloud-up

# Destruir TODA infraestrutura AWS + deletar secrets forçadamente
make cloud-down

# Mostrar URLs de acesso da aplicação na AWS
make cloud-urls
```

**O que `cloud-up` faz:**

**Fase 1 - Terraform (~15 minutos):**
- VPC + Subnets (public/private em 2 AZs)
- EKS Cluster + Node Groups (t3.medium, auto-scaling 2-4 nodes)
- RDS PostgreSQL x2 (auth + orders)
- AmazonMQ (RabbitMQ gerenciado)
- Route53 Hosted Zone
- Secrets Manager

**Fase 2 - Kubernetes via deploy-eks.sh (~10 minutos):**
- AWS Load Balancer Controller
- Metrics Server + External Secrets Operator
- Datastores Helm Charts: MongoDB, Redis, RabbitMQ
- Monitoring Stack: Prometheus + Grafana + Alertmanager
- Velure Services: auth, product, publish-order, process-order, UI

**O que `cloud-down` faz:**
- **Fase 1:** Deleta todos os secrets do Secrets Manager (forçado, mesmo pendentes)
- **Fase 2:** Remove recursos Kubernetes (Helm releases, PVCs, namespaces)
- **Fase 3:** Destrói infraestrutura Terraform (VPC, EKS, RDS, AmazonMQ)
- **Confirmação obrigatória:** Requer digitar "DESTROY"

**O que `cloud-urls` faz:**
- Busca URL do ALB do Frontend (Ingress)
- Busca URL do Grafana (ou mostra comando port-forward)
- Lista todos os Ingresses ativos
- Mostra credenciais de acesso

---

### Fluxo de Trabalho Típico

#### Local Development
```bash
# Dia 1: Subir ambiente
make local-up

# Desenvolver, testar, debugar...
# Acessar: https://velure.local

# Fim do dia: Derrubar
make local-down
```

#### Cloud Deployment
```bash
# Deploy completo
make cloud-up
# Aguardar ~25 minutos (Terraform + Kubernetes)

# Obter URLs
make cloud-urls

# Testar produção...

# Destruir (quando terminar testes)
make cloud-down
# Digite: DESTROY
```

---

### Características do Makefile

- **Simplificado:** Apenas comandos essenciais, sem complexidade desnecessária
- **Verboso:** Feedback claro sobre cada etapa do processo
- **Seguro:** Confirmação obrigatória para comandos destrutivos
- **Idempotente:** Comandos podem ser executados múltiplas vezes
- **Self-documented:** `make help` mostra todos os comandos

---

## 📚 Documentação Adicional

- [**CLAUDE.md**](CLAUDE.md) - Guia para Claude Code (desenvolvimento assistido)
- [**infrastructure/terraform/README.md**](infrastructure/terraform/README.md) - Detalhes do Terraform
- [**infrastructure/local/README.md**](infrastructure/local/README.md) - Setup local detalhado
- [**infrastructure/kubernetes/README.md**](infrastructure/kubernetes/README.md) - Helm charts e Kubernetes

---

## 🎓 Aprendizados e Best Practices

Este projeto demonstra:

### DevOps
- ✅ CI/CD completo com GitHub Actions
- ✅ Infrastructure as Code (Terraform modular)
- ✅ GitOps (deployment via Git)
- ✅ Automated testing (unit + integration + load)
- ✅ Security scanning integrado ao pipeline

### Cloud-Native
- ✅ Containerização com multi-stage builds
- ✅ Orquestração Kubernetes (EKS)
- ✅ Service mesh ready (Istio/Linkerd)
- ✅ 12-Factor App principles
- ✅ Stateless services (exceto datastores)

### SRE (Site Reliability Engineering)
- ✅ Observabilidade completa (metrics + logs + traces)
- ✅ SLI/SLO tracking
- ✅ Error budgets
- ✅ Incident response playbooks
- ✅ Chaos engineering ready

### Architecture
- ✅ Microservices com loose coupling
- ✅ Database-per-service pattern
- ✅ Event-driven architecture (RabbitMQ)
- ✅ CQRS pattern (publish vs process orders)
- ✅ API Gateway pattern (Caddy/ALB)

### Security
- ✅ Defense in depth
- ✅ Least privilege IAM
- ✅ Secrets management (não hardcoded)
- ✅ Network segmentation
- ✅ Automated vulnerability scanning

---

## 🤝 Contribuindo

Contribuições são bem-vindas! Este é um projeto educacional.

```bash
# 1. Fork o projeto
# 2. Crie sua feature branch
git checkout -b feature/nova-feature

# 3. Commit suas mudanças
git commit -m "feat: adiciona nova feature"

# 4. Push para o branch
git push origin feature/nova-feature

# 5. Abra um Pull Request
```

### Convenções

- **Commits**: Seguir [Conventional Commits](https://www.conventionalcommits.org/)
- **Code Style**: `make format` antes de commit
- **Tests**: Adicionar testes para novas features
- **Docs**: Atualizar documentação relevante

---

## 📄 Licença

Este projeto é licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes.

---

## 👨‍💻 Autor

**iCl00ud**

- GitHub: [@icl00ud](https://github.com/icl00ud)
- LinkedIn: [iCl00ud](https://linkedin.com/in/icl00ud)

---

## 🙏 Agradecimentos

- **Hashicorp** - Terraform
- **Kubernetes** - Orquestração de containers
- **Prometheus** - Monitoring
- **Grafana** - Visualização
- **AWS** - Cloud infrastructure
- **Docker** - Containerização
- **RabbitMQ** - Message queue
- **Caddy** - Reverse proxy

---

<div align="center">

**⭐ Se este projeto foi útil, considere dar uma estrela!**

**Made with ❤️ for learning DevOps & Cloud-Native technologies**

</div>
