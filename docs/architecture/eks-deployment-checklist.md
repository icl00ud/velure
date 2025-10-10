# Checklist: Deploy EKS na AWS com DNS Customizado

## 📋 Visão Geral
Este documento lista todos os componentes e configurações necessárias para subir a aplicação Velure em um cluster EKS na AWS com domínio customizado e seguindo best practices de produção.

---

## ✅ Infraestrutura Base (IaC)

### Terraform/CloudFormation
- [ ] **VPC dedicada** com subnets públicas e privadas em múltiplas AZs
- [ ] **NAT Gateways** para subnets privadas (alta disponibilidade)
- [ ] **Security Groups** otimizados por serviço
- [ ] **EKS Cluster** (v1.28+) com control plane logging habilitado
- [ ] **Node Groups** gerenciados (t3.medium/t3.large recomendado para início)
- [ ] **IAM Roles** para:
  - EKS Cluster Role
  - Node Group Role
  - OIDC Provider para IRSA (IAM Roles for Service Accounts)
  - External Secrets Operator
  - AWS Load Balancer Controller
  - EBS CSI Driver
  - Cluster Autoscaler/Karpenter

### Managed Services AWS
- [ ] **RDS PostgreSQL** (Multi-AZ para produção)
  - Instance: db.t3.medium ou superior
  - Backup automático habilitado
  - Encryption at rest
  - Parameter group otimizado
- [ ] **DocumentDB** ou **MongoDB Atlas** para produtos
  - Cluster com 3 réplicas (Multi-AZ)
  - Encryption habilitada
- [ ] **ElastiCache Redis** (cluster mode habilitado)
  - Cache node: cache.t3.micro para dev, cache.r6g.large para prod
  - Multi-AZ com failover automático
- [ ] **Amazon MQ (RabbitMQ)** ou RabbitMQ em K8s com StatefulSet
  - Broker: mq.t3.micro para dev, mq.m5.large para prod
  - Multi-AZ deployment

---

## 🌐 Networking & DNS

### Route53
- [ ] **Hosted Zone** para domínio (ex: velure.com.br)
- [ ] **A Records** apontando para Load Balancer:
  - `velure.com.br` → ALB
  - `www.velure.com.br` → ALB
  - `api.velure.com.br` → ALB
  - `*.velure.com.br` (wildcard para subdomínios)

### Certificados SSL/TLS
- [ ] **AWS Certificate Manager (ACM)**:
  - Certificado para `*.velure.com.br`
  - Validação via DNS (Route53)
- [ ] **cert-manager** (alternativo/adicional):
  - ClusterIssuer com Let's Encrypt
  - Certificate CRDs para cada serviço

---

## 🔀 Proxy Reverso / Ingress

### ⚠️ **CRÍTICO - IMPLEMENTAR PRIMEIRO**

#### Opção 1: AWS Load Balancer Controller (Recomendado)
- [ ] Instalar **AWS Load Balancer Controller** via Helm
  ```bash
  helm repo add eks https://aws.github.io/eks-charts
  helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
    -n kube-system \
    --set clusterName=velure-eks \
    --set serviceAccount.create=false \
    --set serviceAccount.name=aws-load-balancer-controller
  ```
- [ ] Criar **Application Load Balancer (ALB)** via Ingress annotations
- [ ] Configurar **Target Groups** para cada serviço
- [ ] Habilitar **WAF** (Web Application Firewall) no ALB
- [ ] Configurar **health checks** otimizados

#### Opção 2: Nginx Ingress Controller
- [ ] Instalar **nginx-ingress-controller** via Helm
  ```bash
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
  helm install nginx-ingress ingress-nginx/ingress-nginx \
    -n ingress-nginx --create-namespace \
    --set controller.service.type=LoadBalancer \
    --set controller.service.annotations."service\.beta\.kubernetes\.io/aws-load-balancer-type"=nlb
  ```
- [ ] Configurar **rate limiting** global
- [ ] Habilitar **ModSecurity WAF** (opcional)
- [ ] SSL Passthrough para serviços que precisam

### Ingress Resources
- [ ] **Ingress principal** com regras de roteamento:
  ```yaml
  velure.com.br/          → ui-service
  velure.com.br/api/auth  → auth-service
  velure.com.br/api/products → product-service
  velure.com.br/api/orders → publish-order-service
  velure.com.br/api/sse   → publish-order-service (SSE)
  ```
- [ ] **Annotations de segurança**:
  - CORS headers
  - Rate limiting
  - Request size limits
  - SSL redirect
  - HSTS headers

### API Gateway (Opcional mas Recomendado)
- [ ] **Kong** ou **Ambassador** para features avançadas:
  - Autenticação centralizada
  - Rate limiting por usuário/API key
  - Request/Response transformation
  - Circuit breaker
  - Retry policies
  - OpenAPI documentation

---

## 🔐 Secrets Management

### External Secrets Operator
- [ ] Instalar **External Secrets Operator**
  ```bash
  helm repo add external-secrets https://charts.external-secrets.io
  helm install external-secrets external-secrets/external-secrets \
    -n external-secrets --create-namespace
  ```
- [ ] Criar **SecretStore** apontando para AWS Secrets Manager
- [ ] Migrar secrets do Kubernetes para AWS Secrets Manager:
  - JWT_SECRET
  - Database credentials
  - RabbitMQ credentials
  - API keys
- [ ] Criar **ExternalSecret** resources para cada serviço
- [ ] Configurar **rotation automática** de secrets

### Alternativa: AWS Systems Manager Parameter Store
- [ ] Usar SSM Parameter Store para configs não-secretas
- [ ] Integrar com External Secrets Operator

---

## 📦 Container Registry

- [ ] **Amazon ECR** (Elastic Container Registry):
  - Criar repositórios para cada serviço:
    - `velure/auth-service`
    - `velure/product-service`
    - `velure/publish-order-service`
    - `velure/process-order-service`
    - `velure/ui-service`
  - Habilitar **scan de vulnerabilidades** automático
  - Configurar **lifecycle policies** (reter últimas 10 imagens)
  - Configurar **image signing** (opcional mas recomendado)

---

## 🔄 CI/CD Pipeline

### GitHub Actions
- [ ] **Workflow de Build**:
  ```yaml
  .github/workflows/build.yml
  ```
  - Trigger: push em branches develop/main
  - Steps: build → test → scan → push para ECR
  - Executar testes unitários (coverage > 80%)
  - Scan de segurança (gosec, govulncheck, trivy)
  - Tag de imagem: `<commit-sha>`, `<branch>`, `latest`

- [ ] **Workflow de Deploy**:
  ```yaml
  .github/workflows/deploy.yml
  ```
  - Trigger: push em main ou release tags
  - Steps: 
    - Update Helm values com nova image tag
    - Helm upgrade em cluster EKS (staging → production)
    - Health check pós-deploy
    - Rollback automático em caso de falha
  - Ambientes: staging, production
  - Approval manual para production

- [ ] **Workflow de Database Migration**:
  ```yaml
  .github/workflows/migrate.yml
  ```
  - Executar migrations antes do deploy
  - Backup automático antes de migrations
  - Rollback de migrations em caso de falha

### Secrets do GitHub
- [ ] Adicionar secrets no repositório:
  - `AWS_ACCESS_KEY_ID`
  - `AWS_SECRET_ACCESS_KEY`
  - `AWS_REGION`
  - `ECR_REPOSITORY_*`
  - `KUBE_CONFIG_DATA` (base64 do kubeconfig)

---

## 📊 Observabilidade

### Prometheus Stack
- [ ] Instalar **kube-prometheus-stack**:
  ```bash
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
    -n monitoring --create-namespace \
    --set prometheus.prometheusSpec.retention=30d \
    --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=50Gi
  ```
- [ ] Configurar **ServiceMonitors** para cada serviço
- [ ] Criar **PrometheusRules** para alertas customizados
- [ ] Configurar **persistent volume** para Prometheus (EBS gp3)

### Grafana
- [ ] Configurar **datasources**:
  - Prometheus
  - Loki
  - Tempo (traces)
- [ ] Importar **dashboards**:
  - Kubernetes cluster overview
  - Node exporter
  - PostgreSQL
  - MongoDB
  - Redis
  - RabbitMQ
  - Go application metrics
  - Nginx ingress
- [ ] Criar dashboards customizados:
  - Orders funnel (por status)
  - Revenue metrics
  - Error rates por serviço
  - Latency percentiles (p50, p95, p99)
- [ ] Configurar **OAuth** para autenticação (Google/GitHub)

### Logging (Loki)
- [ ] Instalar **Loki** e **Promtail**:
  ```bash
  helm repo add grafana https://grafana.github.io/helm-charts
  helm install loki grafana/loki-stack \
    -n monitoring \
    --set loki.persistence.enabled=true \
    --set loki.persistence.size=50Gi \
    --set promtail.enabled=true
  ```
- [ ] Configurar **retention** de logs (30 dias)
- [ ] Criar **LogQL queries** para erros críticos
- [ ] Integrar com Grafana para visualização

### Distributed Tracing (Tempo/Jaeger)
- [ ] Instalar **Grafana Tempo**:
  ```bash
  helm install tempo grafana/tempo -n monitoring
  ```
- [ ] Instrumentar aplicações Go com **OpenTelemetry**
- [ ] Configurar **sampling** (1% em produção, 100% em staging)
- [ ] Criar queries para trace de requests entre serviços

### Alertmanager
- [ ] Configurar **receivers**:
  - Slack webhook
  - PagerDuty (para produção)
  - Email para alertas não-críticos
- [ ] Criar **routing rules** por severidade
- [ ] Configurar **inhibition rules** para evitar alert storm

---

## 📈 Autoscaling

### Horizontal Pod Autoscaler (HPA)
- [ ] Instalar **metrics-server**:
  ```bash
  kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
  ```
- [ ] Configurar **HPA** para cada serviço:
  - auth-service: min=2, max=10, targetCPU=70%
  - product-service: min=2, max=10, targetCPU=70%
  - publish-order-service: min=2, max=10, targetCPU=70%
  - process-order-service: min=3, max=15, targetCPU=70% (consumer)
  - ui-service: min=2, max=10, targetCPU=70%
- [ ] Considerar **custom metrics** (RabbitMQ queue length para process-order)

### Cluster Autoscaler ou Karpenter
- [ ] **Opção 1 - Cluster Autoscaler**:
  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml
  ```
  - Configurar IAM role com políticas necessárias
  - Ajustar min/max nodes por node group

- [ ] **Opção 2 - Karpenter** (Recomendado):
  ```bash
  helm repo add karpenter https://charts.karpenter.sh
  helm install karpenter karpenter/karpenter -n karpenter --create-namespace
  ```
  - Configurar **Provisioners** com instance types variados
  - Habilitar **spot instances** para workloads tolerantes
  - Configurar **disruption budgets**

---

## 🛡️ Security

### Network Policies
- [ ] Criar **NetworkPolicy** para cada namespace:
  - Default deny all ingress/egress
  - Allow apenas comunicação necessária entre serviços
  - Allow DNS queries
  - Allow métricas do Prometheus
- [ ] Exemplo para `order` namespace:
  ```yaml
  # Allow publish-order → PostgreSQL
  # Allow publish-order → RabbitMQ
  # Allow process-order → RabbitMQ
  # Deny resto
  ```

### Pod Security Standards
- [ ] Aplicar **Pod Security Admission**:
  - `authentication` namespace: restricted
  - `order` namespace: restricted
  - `frontend` namespace: restricted
  - `database` namespace: baseline
- [ ] Garantir que todos os pods rodam como **non-root**
- [ ] Garantir **readOnlyRootFilesystem: true** onde possível
- [ ] Drop capabilities desnecessárias

### RBAC
- [ ] Criar **ServiceAccounts** específicos por serviço
- [ ] Configurar **Roles/ClusterRoles** com princípio de menor privilégio
- [ ] Nunca usar `cluster-admin` em produção
- [ ] Auditar permissões regularmente

### Image Security
- [ ] Escanear imagens com **Trivy/Snyk** no CI
- [ ] Assinar imagens com **Cosign** (sigstore)
- [ ] Configurar **ImagePolicyWebhook** para aceitar apenas imagens assinadas
- [ ] Usar **distroless images** no final stage do Dockerfile

### AWS Security
- [ ] Habilitar **GuardDuty** para detecção de ameaças
- [ ] Habilitar **AWS Config** para compliance
- [ ] Configurar **CloudTrail** para auditoria
- [ ] Habilitar **VPC Flow Logs**
- [ ] Configurar **Security Hub** para visão centralizada

---

## 💾 Storage & Backup

### Persistent Volumes
- [ ] Instalar **EBS CSI Driver**:
  ```bash
  helm repo add aws-ebs-csi-driver https://kubernetes-sigs.github.io/aws-ebs-csi-driver
  helm install aws-ebs-csi-driver aws-ebs-csi-driver/aws-ebs-csi-driver \
    -n kube-system
  ```
- [ ] Criar **StorageClass** otimizada (gp3):
  ```yaml
  kind: StorageClass
  apiVersion: storage.k8s.io/v1
  metadata:
    name: ebs-gp3
  provisioner: ebs.csi.aws.com
  parameters:
    type: gp3
    iops: "3000"
    throughput: "125"
  volumeBindingMode: WaitForFirstConsumer
  ```

### Backup Strategy
- [ ] **Velero** para backup de recursos Kubernetes:
  ```bash
  helm repo add vmware-tanzu https://vmware-tanzu.github.io/helm-charts
  helm install velero vmware-tanzu/velero \
    -n velero --create-namespace \
    --set-file credentials.secretContents.cloud=./credentials-velero \
    --set configuration.backupStorageLocation[0].bucket=velure-k8s-backups \
    --set configuration.backupStorageLocation[0].provider=aws \
    --set snapshotsEnabled=true
  ```
- [ ] Configurar **backup schedule** diário
- [ ] Testar **restore procedures** mensalmente

- [ ] **Database backups**:
  - RDS: automated backups + manual snapshots antes de migrations
  - MongoDB: backup automático com retention de 30 dias
  - Redis: snapshot diário (se persistência habilitada)

---

## 💰 Cost Optimization

### Kubecost
- [ ] Instalar **Kubecost**:
  ```bash
  helm repo add kubecost https://kubecost.github.io/cost-analyzer/
  helm install kubecost kubecost/cost-analyzer \
    -n kubecost --create-namespace \
    --set prometheus.server.global.external_labels.cluster_id=velure-eks
  ```
- [ ] Configurar **alerts** para anomalias de custo
- [ ] Revisar **recommendations** semanalmente

### Savings Plans
- [ ] Avaliar **Compute Savings Plans** após 1 mês de uso
- [ ] Considerar **Reserved Instances** para databases
- [ ] Usar **Spot Instances** para workloads não-críticos (até 90% economia)

---

## 🧪 Ambientes

### Staging
- [ ] Cluster EKS separado ou namespace isolado
- [ ] Databases menores (db.t3.small, cache.t3.micro)
- [ ] Subdomínio: `staging.velure.com.br`
- [ ] Sync automático com branch `develop`
- [ ] Retention menor de logs/backups

### Production
- [ ] Cluster EKS dedicado
- [ ] Multi-AZ deployment obrigatório
- [ ] Domínio principal: `velure.com.br`
- [ ] Manual approval para deploy
- [ ] Backup e retention completos

---

## 📚 Documentação

- [ ] **Runbooks** para incidentes comuns
- [ ] **Architecture Decision Records (ADRs)**
- [ ] **API documentation** (Swagger/OpenAPI)
- [ ] **Diagrama de arquitetura** atualizado
- [ ] **Disaster Recovery Plan** documentado
- [ ] **Onboarding guide** para novos desenvolvedores

---

## 🚀 Próximos Passos Recomendados

### Fase 1: Proxy Reverso Local (Agora)
1. Implementar Nginx como proxy reverso no Docker Compose
2. Centralizar roteamento: `/api/auth`, `/api/products`, `/api/orders`
3. Configurar CORS, rate limiting, SSL local
4. Testar fluxo completo

### Fase 2: IaC & EKS Base (Semana 1)
1. Criar repositório Terraform para infraestrutura
2. Provisionar VPC, EKS, Node Groups
3. Configurar kubectl e Helm
4. Deploy de serviços básicos (metrics-server, aws-load-balancer-controller)

### Fase 3: Ingress & DNS (Semana 1-2)
1. Instalar Ingress Controller
2. Configurar Route53 e ACM
3. Deploy dos microserviços com Ingress
4. Validar SSL e DNS

### Fase 4: Observabilidade (Semana 2)
1. Deploy do stack Prometheus/Grafana/Loki
2. Configurar dashboards e alertas
3. Instrumentar aplicações com métricas customizadas

### Fase 5: CI/CD (Semana 2-3)
1. Criar workflows GitHub Actions
2. Configurar ECR
3. Automatizar build e deploy
4. Implementar estratégia de rollback

### Fase 6: Security Hardening (Semana 3-4)
1. External Secrets Operator
2. Network Policies
3. Pod Security Standards
4. Image scanning e signing

### Fase 7: Production Ready (Semana 4+)
1. Backups automatizados
2. Disaster recovery testing
3. Load testing e tunning
4. Documentation completa

---

## 🎯 Prioridade ALTA - Implementar AGORA

### 1. Proxy Reverso Local (Nginx)
**Por que:** Padronizar acesso aos serviços, preparar para Ingress K8s

**Implementação:**
```nginx
# nginx.conf
location /api/auth {
  proxy_pass http://auth-service:3001;
}
location /api/products {
  proxy_pass http://product-service:3010;
}
location /api/orders {
  proxy_pass http://publish-order-service:3002;
}
location / {
  proxy_pass http://ui-service:3000;
}
```

Benefícios:
- ✅ Single entry point
- ✅ CORS centralizado
- ✅ Rate limiting
- ✅ SSL termination
- ✅ Request logging
- ✅ Health check aggregation

---

## 📊 Estimativa de Custos Mensais (AWS)

### Ambiente Staging (Mínimo)
- EKS Control Plane: $73/mês
- 2x t3.medium nodes: ~$60/mês
- RDS db.t3.small: ~$35/mês
- DocumentDB t3.medium (1 node): ~$70/mês
- ElastiCache cache.t3.micro: ~$15/mês
- ALB: ~$20/mês
- NAT Gateway: ~$35/mês
- **Total: ~$308/mês**

### Ambiente Production (Recomendado)
- EKS Control Plane: $73/mês
- 3x t3.large nodes (initial): ~$210/mês
- RDS db.t3.medium Multi-AZ: ~$150/mês
- DocumentDB m5.large (3 nodes): ~$450/mês
- ElastiCache cache.r6g.large: ~$130/mês
- ALB: ~$30/mês
- NAT Gateway (Multi-AZ): ~$70/mês
- S3 (backups/logs): ~$20/mês
- CloudWatch/Logs: ~$30/mês
- **Total: ~$1,163/mês**

**Nota:** Custos podem variar com traffic/uso. Considerar Savings Plans após estabilização.

---

**Deseja que eu implemente o proxy reverso Nginx agora?**
