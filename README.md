# Velure - E-Commerce Microservices Platform

> **Objetivo principal**: Aprender e demonstrar arquitetura de microserviços moderna ✨

Este repositório contém um sistema de e-commerce completo construído com arquitetura de microserviços, seguindo padrões cloud-native e práticas DevSecOps. O projeto implementa funcionalidades essenciais como registro de usuários, autenticação, gestão de produtos e processamento de pedidos.

## 📁 Estrutura do Repositório

```
velure/
├── services/                          # Microserviços
│   ├── auth-service/                 # Autenticação (Go + PostgreSQL + Redis)
│   ├── product-service/              # Catálogo (Go + MongoDB + Redis)
│   ├── publish-order-service/        # Criação de pedidos (Go + PostgreSQL + RabbitMQ)
│   ├── process-order-service/        # Processamento (Go + PostgreSQL + RabbitMQ)
│   └── ui-service/                   # Frontend (React + TypeScript + Vite)
│
├── infrastructure/                    # Infraestrutura como código
│   ├── terraform/                    # AWS EKS (VPC, RDS, EKS cluster)
│   ├── kubernetes/
│   │   ├── charts/                   # Helm charts
│   │   │   ├── velure-datastores/   # MongoDB, Redis, RabbitMQ (unified)
│   │   │   ├── velure-auth/         # Auth service chart
│   │   │   ├── velure-product/      # Product service chart
│   │   │   ├── velure-publish-order/
│   │   │   ├── velure-process-order/
│   │   │   └── velure-ui/
│   │   └── monitoring/              # Prometheus + Grafana (K8s)
│   └── local/                       # Docker Compose
│       ├── docker-compose.yaml      # Aplicação
│       ├── docker-compose.monitoring.yaml  # Grafana + Prometheus
│       └── monitoring/              # Configs Prometheus/Grafana
│
├── docs/                            # Documentação
│   ├── architecture/                # Diagramas AWS + arquitetura
│   ├── DEPLOY_GUIDE.md             # Guia de deploy AWS/EKS
│   ├── MONITORING.md               # Guia de monitoramento K8s
│   ├── PROMETHEUS_METRICS.md       # Referência de métricas
│   └── TROUBLESHOOTING.md          # Solução de problemas
│
├── tests/                          # Testes
│   ├── load/                       # k6 load tests
│   └── integration/                # Testes de integração
│
├── scripts/                        # Scripts de automação
│   └── deploy/                     # Scripts de deploy AWS/EKS
│
├── START_HERE.sh                   # Script interativo para iniciar
├── Makefile                        # Comandos de automação
└── CLAUDE.md                       # Guia completo de desenvolvimento
```

## 🏗️ Arquitetura dos Serviços

### **Auth Service** 🔐
- **Stack**: Go, PostgreSQL, Redis
- **Porta**: 3020
- **Funcionalidades**:
  - Registro e login de usuários
  - Gestão de sessões e JWT tokens
  - Autorização baseada em roles

### **Product Service** 📦
- **Stack**: Go, MongoDB, Redis
- **Porta**: 3010
- **Funcionalidades**:
  - CRUD de produtos
  - Gestão de inventário
  - Cache de produtos frequentes

### **Publish Order Service** 📤
- **Stack**: Go, PostgreSQL, RabbitMQ
- **Porta**: 3030
- **Funcionalidades**:
  - Criação de novos pedidos
  - Validação de dados
  - Publicação em fila para processamento

### **Process Order Service** ⚙️
- **Stack**: Go, PostgreSQL, RabbitMQ
- **Porta**: 3040
- **Funcionalidades**:
  - Processamento assíncrono de pedidos
  - Atualização de status
  - Integração com sistemas externos

### **UI Service** 🎨
- **Stack**: React, TypeScript, Tailwind CSS
- **Porta**: 80 (Nginx)
- **Funcionalidades**:
  - Interface web responsiva
  - Integração com todos os serviços
  - Experiência de usuário moderna

## 🛠️ Tecnologias Utilizadas

### **Backend**
- **Linguagens**: Go, TypeScript
- **Frameworks**: Gin (Go), React
- **Bancos de dados**: PostgreSQL, MongoDB, Redis
- **Mensageria**: RabbitMQ
- **Cache**: Redis

### **DevOps & Infraestrutura**
- **Containers**: Docker, Kubernetes
- **Orquestração**: Helm Charts
- **Cloud**: AWS EKS
- **IaC**: Terraform
- **CI/CD**: GitHub Actions (planejado)
- **Monitoramento**: Prometheus, Grafana

### **Desenvolvimento Local**
- **Orquestração**: Docker Compose
- **Proxy reverso**: Caddy (com TLS automático)
- **Testes**: k6 (load testing)

## 🚀 Quick Start

### ⚡ Modo Mais Rápido (Recomendado)

```bash
# 1. Clonar o repositório
git clone https://github.com/icl00ud/velure.git
cd velure

# 2. Configurar /etc/hosts
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts

# 3. Copiar variáveis de ambiente
cp infrastructure/local/.env.example infrastructure/local/.env

# 4. Rodar aplicação completa com monitoramento
./START_HERE.sh
# OU usando Makefile:
make monitoring-setup
```

**Acessos após iniciar:**
- 🌐 Aplicação: https://velure.local
- 📊 Grafana (dashboards): http://localhost:3000 (admin/admin)
- 📈 Prometheus: http://localhost:9090
- 🐰 RabbitMQ Management: http://localhost:15672 (admin/admin_password)

---

## 🛠️ Modos de Execução

### 🐳 Desenvolvimento Local (Docker Compose)

**Opção 1: Aplicação + Monitoramento (Recomendado)**
```bash
cd infrastructure/local
docker-compose -f docker-compose.yaml -f docker-compose.monitoring.yaml up -d
```

**Opção 2: Apenas Aplicação**
```bash
cd infrastructure/local
docker-compose up -d
```

**Opção 3: Serviços Individuais (Hot Reload)**
```bash
# Subir infraestrutura primeiro
make dev

# Em terminais separados, executar cada serviço
cd services/auth-service && go run main.go
cd services/product-service && go run main.go
cd services/publish-order-service && go run main.go
cd services/process-order-service && go run main.go
cd services/ui-service && npm install && npm run dev
```

**Acesso via Proxy Reverso (Caddy):**
- 🌐 **Aplicação**: https://velure.local
- 🔐 **Auth API**: https://velure.local/api/auth/*
- 📦 **Product API**: https://velure.local/api/product/*
- 📤 **Order API**: https://velure.local/api/order/*

> ⚠️ **IMPORTANTE**: Sempre use `https://velure.local` - nunca acesse containers diretamente

---

### ☁️ AWS EKS (Produção)

```bash
# Pré-requisitos: terraform, aws-cli, kubectl, helm
# Ver docs/DEPLOY_GUIDE.md para guia completo

# 1. Deploy da infraestrutura AWS (VPC, EKS, RDS)
cd infrastructure/terraform
terraform init
terraform plan
terraform apply

# 2. Configurar kubectl
aws eks update-kubeconfig --region us-east-1 --name velure-prod

# 3. Deploy completo (controllers + datastores + monitoring + services)
make eks-deploy-full

# OU passo a passo:
make eks-install-controllers    # ALB Controller, metrics-server
make eks-install-datastores     # MongoDB, Redis, RabbitMQ
make eks-install-monitoring     # Prometheus + Grafana
make eks-deploy-services        # Velure microservices
```

**Custo estimado AWS**: ~$100-150/mês (com Free Tier RDS)
**Documentação completa**: Ver [docs/DEPLOY_GUIDE.md](docs/DEPLOY_GUIDE.md)

## 📊 Monitoramento

### **Grafana + Prometheus (Local)**

O stack de monitoramento está integrado no Docker Compose:

```bash
# Iniciar com monitoramento
make monitoring-setup

# Acessar dashboards
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
```

**Dashboard principal**: http://localhost:3000/d/velure-overview

Métricas disponíveis:
- Request rate por serviço
- Response time (p95)
- Error rate (5xx)
- Memory usage
- RabbitMQ queue depth

Ver guia completo: [infrastructure/local/MONITORING.md](infrastructure/local/MONITORING.md)

### **Health Checks**
Todos os serviços expõem `/health` endpoint:
```bash
curl https://velure.local/api/auth/health
curl https://velure.local/api/product/health
curl https://velure.local/api/order/health
```

### **Métricas (Prometheus)**
```bash
# Através do proxy
curl https://velure.local/api/auth/metrics -k
curl https://velure.local/api/product/metrics -k

# Ou diretamente (desenvolvimento)
curl http://localhost:3020/metrics
curl http://localhost:3010/metrics
```

### **Logs**
Todos os serviços usam structured logging (JSON) com:
- `timestamp`, `level`, `message`, `service`
- `trace_id`, `user_id` (quando aplicável)
- Agregação com CloudWatch (AWS) ou stdout (local)

## 🧪 Testes

### **Testes de Carga & Escalonamento Horizontal (k6 + HPA)**

A aplicação está preparada para testes de carga com observação de escalonamento horizontal automático (HPA) no ambiente Kubernetes (AWS EKS).

**Quick Start - Kubernetes:**
```bash
cd tests/load

# 1. Rodar teste integrado
./run-all-tests.sh

# 2. Monitorar escalonamento em tempo real (em outro terminal)
./monitor-scaling.sh
```

**Testes Disponíveis:**
- `auth` - Auth service (200 VUs max)
- `product` - Product service (400 VUs max)
- `order` - Order service (1000 VUs max)
- `ui` - UI service (250 VUs max)
- `integrated` - Todos os serviços (500 VUs max) **← Recomendado**

**Observar Escalonamento:**
```bash
# Terminal 1: Executar teste
./run-all-tests.sh

# Terminal 2: Monitorar pods escalando
./monitor-scaling.sh

# Terminal 3: Watch HPA
kubectl get hpa -w
```

**O que você verá:**
- 🚀 Pods escalando de 2 → 5-10 replicas quando CPU > 80%
- 📈 Métricas em tempo real no dashboard Grafana
- ⏱️  Response time se mantendo estável mesmo com carga alta
- 📉 Scale-down automático após teste (5 min de estabilização)

**Documentação completa:** [docs/LOAD_TESTING.md](docs/LOAD_TESTING.md)

### **Testes Unitários**
```bash
# Cada serviço tem seus próprios testes
cd services/auth-service
go test ./...

cd services/ui-service
npm test
```

## 🔐 Segurança

### **Implementado**
- ✅ JWT tokens com refresh
- ✅ HTTPS em todos os endpoints
- ✅ Rate limiting
- ✅ Input validation
- ✅ CORS configurado
- ✅ Network policies (Kubernetes)
- ✅ Security contexts (containers não-root)
- ✅ Secrets management

### **Planejado**
- 🔄 OAuth2/OpenID Connect
- 🔄 Scanning de vulnerabilidades
- 🔄 WAF (Web Application Firewall)
- 🔄 Pod Security Standards

## 🗺️ Roadmap

### **Versão 2.0** (Em desenvolvimento)
- [ ] Payment Service
- [ ] Notification Service
- [ ] User Service (separar do Auth)
- [ ] API Gateway (Kong/Ambassador)
- [ ] Service Mesh (Istio)

### **Versão 3.0** (Planejado)
- [ ] Event Sourcing
- [ ] CQRS pattern
- [ ] Distributed tracing (Jaeger)
- [ ] Chaos engineering
- [ ] Multi-region deployment

## 📚 Documentação

| Documento | Descrição |
|-----------|-----------|
| [START_HERE.sh](START_HERE.sh) | Script interativo - ponto de entrada único |
| [CLAUDE.md](CLAUDE.md) | Guia completo para desenvolvimento |
| [Arquitetura AWS](docs/architecture/ARCHITECTURE.md) | Diagramas e infraestrutura completa |
| [Deploy AWS/EKS](docs/DEPLOY_GUIDE.md) | Guia passo-a-passo para produção |
| [Monitoramento](docs/MONITORING.md) | Grafana + Prometheus (local e K8s) |
| [Load Testing & HPA](docs/LOAD_TESTING.md) | Testes de carga e escalonamento horizontal |
| [Troubleshooting](docs/TROUBLESHOOTING.md) | Solução de problemas comuns |
| [Prometheus Metrics](docs/PROMETHEUS_METRICS.md) | Referência de métricas |

## 🤝 Contribuindo

1. Fork o projeto
2. Crie sua feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### **Padrões de Código**
- Go: `gofmt`, `golint`, `gosec`
- TypeScript: `prettier`, `eslint`
- Commits: [Conventional Commits](https://www.conventionalcommits.org/)

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo [LICENSE](LICENSE) para detalhes.

## 👨‍💻 Autor

**icl00ud**
- GitHub: [@icl00ud](https://github.com/icl00ud)
- LinkedIn: [Seu LinkedIn]

---

**⭐ Se este projeto te ajudou, considere dar uma estrela!**

> Feito com ❤️ para aprender e compartilhar conhecimento sobre microserviços.
