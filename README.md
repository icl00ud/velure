# Velure - E-Commerce Microservices Platform

> **Objetivo principal**: Aprender e demonstrar arquitetura de microserviços moderna ✨

Este repositório contém um sistema de e-commerce completo construído com arquitetura de microserviços, seguindo padrões cloud-native e práticas DevSecOps. O projeto implementa funcionalidades essenciais como registro de usuários, autenticação, gestão de produtos e processamento de pedidos.

## 📁 Estrutura do Repositório

```
velure/
├── services/                    # Microserviços
│   ├── auth-service/           # Autenticação (Go)
│   ├── product-service/        # Catálogo (Go + MongoDB)
│   ├── publish-order-service/  # Criação de pedidos (Go)
│   ├── process-order-service/  # Processamento (Go)
│   └── ui-service/            # Frontend (React)
├── 
├── infrastructure/             # Toda infraestrutura como código
│   ├── terraform/             # AWS EKS deployment
│   ├── kubernetes/            # Helm charts e manifests
│   └── local/                 # Docker Compose local
├── 
├── shared/                    # Código compartilhado
│   └── models/               # Modelos de dados
├── 
├── docs/                     # Documentação
│   ├── architecture/         # Diagramas e arquitetura
│   ├── api/                  # Documentação das APIs
│   └── deployment/           # Guias de deploy
├── 
├── tests/                    # Testes integrados
│   ├── load/                 # Testes de carga (k6)
│   └── integration/          # Testes de integração
├── 
├── tools/                    # Ferramentas e utilitários
│   └── monitoring/           # Prometheus, Grafana
└── 
└── scripts/                  # Scripts de automação
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

## 🚀 Como Executar

### 🐳 Desenvolvimento Local (Docker Compose)

```bash
# 1. Clonar o repositório
git clone https://github.com/icl00ud/velure.git
cd velure

# 2. Subir dependências (bancos, cache, filas)
cd infrastructure/local
docker-compose up -d

# 3. Executar cada serviço individualmente para desenvolvimento
# Auth Service
cd services/auth-service
go run main.go

# Product Service
cd services/product-service
go run main.go

# Publish Order Service
cd services/publish-order-service
go run main.go

# Process Order Service
cd services/process-order-service
go run main.go

# UI Service
cd services/ui-service
npm install && npm run dev
```

**URLs Locais**:
- Frontend: https://localhost:3000
- Auth API: https://localhost:3020
- Product API: https://localhost:3010
- Order APIs: https://localhost:3030, https://localhost:3040

### ☸️ Kubernetes Local

```bash
# Pré-requisitos: kubectl, helm, mkcert
# Ver docs/deployment/kubernetes-local-guide.md para setup completo

# 1. Criar namespaces
kubectl create namespace database
kubectl create namespace order
kubectl create namespace authentication
kubectl create namespace frontend

# 2. Deploy databases
helm upgrade --install postgres infrastructure/kubernetes/charts/postgresql -n database
helm upgrade --install mongodb infrastructure/kubernetes/charts/mongodb -n database
helm upgrade --install redis infrastructure/kubernetes/charts/redis -n database

# 3. Deploy serviços
helm upgrade --install velure-auth infrastructure/kubernetes/charts/velure-auth -n authentication
helm upgrade --install velure-product infrastructure/kubernetes/charts/velure-product -n order
# ... outros serviços
```

### ☁️ AWS EKS (Produção)

```bash
# Pré-requisitos: terraform, aws-cli, kubectl
# Ver docs/deployment/terraform-guide.md para setup completo

cd infrastructure/terraform

# 1. Configurar variáveis
cp terraform.tfvars.example terraform.tfvars
# Editar senhas e configurações

# 2. Deploy da infraestrutura
terraform init
terraform plan
terraform apply

# 3. Configurar kubectl
aws eks update-kubeconfig --region us-east-1 --name velure-prod

# 4. Deploy dos serviços via Helm
# Ver docs/deployment/terraform-guide.md
```

**Custo estimado AWS**: ~$143/mês (com Spot instances e Free Tier RDS)

## 📊 Monitoramento

### **Health Checks**
Todos os serviços expõem `/health` endpoint:
```bash
curl http://localhost:3020/health  # Auth
curl http://localhost:3010/health  # Product
curl http://localhost:3030/health  # Publish Order
curl http://localhost:3040/health  # Process Order
```

### **Métricas (Prometheus)**
```bash
curl http://localhost:3020/metrics  # Métricas do Auth Service
# Grafana dashboard disponível em tools/monitoring/
```

### **Logs**
Todos os serviços usam structured logging (JSON) com:
- `timestamp`, `level`, `message`
- `trace_id`, `user_id` (quando aplicável)
- Integração com ELK Stack (planejado)

## 🧪 Testes

### **Testes de Carga (k6)**
```bash
cd tests/load

# Teste individual de um serviço
k6 run auth-service-test.js

# Teste integrado de todo o fluxo
k6 run integrated-load-test.js

# Todos os testes
./run-all-tests.sh
```

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
| [Arquitetura](docs/architecture/ARCHITECTURE_DIAGRAM.md) | Diagramas e fluxos do sistema |
| [Deploy AWS](docs/deployment/terraform-guide.md) | Guia completo para AWS EKS |
| [Deploy Local](docs/deployment/kubernetes-local-guide.md) | Kubernetes local com Helm |
| [Estimativa de Custos](docs/deployment/COST_ESTIMATION.md) | Análise detalhada de custos AWS |
| [API Reference](docs/api/) | Documentação das APIs |

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
