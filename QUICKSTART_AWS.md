# Velure - Quickstart AWS EKS

Guia passo a passo completo para subir a infraestrutura na AWS, deployar as aplicações e testar.

## ⏱️ Tempo Total Estimado
- Infraestrutura AWS: **~35 minutos**
- Deploy de aplicações: **~20 minutos**
- Testes e validações: **~10 minutos**
- **TOTAL: ~65 minutos** (1 hora)

## 💰 Custos
- **~$124-150/mês** (detalhes no final deste guia)
- **Importante**: Configure Budget Alerts na AWS!

---

## 📋 FASE 0: Pré-requisitos (15 minutos)

### 0.1. Ferramentas Necessárias

```bash
# Verificar se estão instaladas
aws --version           # >= 2.0.0
terraform --version     # >= 1.6.0
kubectl version --client # >= 1.28.0
helm version            # >= 3.12.0
```

**Instalar** (se necessário):

```bash
# macOS
brew install awscli terraform kubectl helm

# Linux
# Seguir docs oficiais de cada ferramenta
```

### 0.2. Configurar Credenciais AWS

```bash
# Opção 1: Configurar perfil default
aws configure

# Opção 2: Usar perfil específico
export AWS_PROFILE=seu-perfil-aws
```

**Teste a conexão:**
```bash
aws sts get-caller-identity
```

Deve mostrar seu UserId, Account e Arn.

### 0.3. Preparar Terraform

```bash
cd infrastructure/terraform

# Copiar arquivo de exemplo
cp terraform.tfvars.example terraform.tfvars

# Editar variáveis (IMPORTANTE: trocar senhas!)
vim terraform.tfvars
```

**Variáveis principais:**
```hcl
aws_region = "us-east-1"  # Ou sua região preferida

# TROCAR SENHAS!
rds_auth_password   = "SUA_SENHA_FORTE_AQUI"
rds_orders_password = "OUTRA_SENHA_FORTE_AQUI"

# Opcional: ajustar tamanho dos nodes
eks_node_instance_type = "t3.small"  # ou t3.micro para economizar
eks_node_desired_size  = 2
```

---

## 🚀 FASE 1: Infraestrutura AWS (~35 minutos)

### 1.1. Deploy com Terraform

```bash
cd infrastructure/terraform

# Inicializar
terraform init

# Planejar (revisar o que será criado)
terraform plan

# Aplicar (criar recursos)
terraform apply
```

Quando perguntar, digite `yes` para confirmar.

**O que será criado:**
- ✅ VPC com subnets públicas e privadas
- ✅ Cluster EKS (Kubernetes na AWS)
- ✅ 2 nodes EC2 (t3.small)
- ✅ 2 bancos RDS PostgreSQL (auth + orders)
- ✅ Security Groups (firewalls)
- ✅ IAM Roles (permissões)

**Aguarde:** 30-40 minutos ☕

### 1.2. Configurar kubectl para conectar ao EKS

```bash
# Conectar kubectl ao cluster
aws eks update-kubeconfig --region us-east-1 --name velure-prod

# Verificar conexão
kubectl get nodes
```

Deve mostrar 2 nodes com status `Ready`.

---

## ☸️ FASE 2: Deploy das Aplicações (~20 minutos)

### 2.1. Instalar Controllers (ALB + Metrics)

```bash
cd ../../scripts/deploy
./01-install-controllers.sh
```

Aguarde ~5 minutos.

**Verificar:**
```bash
kubectl get deployment -n kube-system aws-load-balancer-controller
kubectl get deployment -n kube-system metrics-server
```

Ambos devem mostrar `READY 1/1` ou `2/2`.

### 2.2. Instalar Datastores (MongoDB, Redis, RabbitMQ)

```bash
./02-install-datastores.sh
```

Aguarde ~10 minutos.

**Verificar:**
```bash
kubectl get pods -n datastores
```

Todos os pods devem estar `Running` (pode demorar alguns minutos).

### 2.3. Instalar Monitoramento (Prometheus + Grafana)

```bash
./03-install-monitoring.sh
```

Aguarde ~5 minutos.

**Verificar:**
```bash
kubectl get pods -n monitoring
```

### 2.4. Deploy dos Microserviços

```bash
./04-deploy-services.sh
```

Aguarde ~5 minutos.

**Verificar:**
```bash
kubectl get pods
```

Deve mostrar:
- `velure-auth-*` - Running
- `velure-product-*` - Running
- `velure-publish-order-*` - Running
- `velure-process-order-*` - Running
- `velure-ui-*` - Running

### 2.5. Aguardar Load Balancer

O AWS ALB (Application Load Balancer) leva ~5 minutos para provisionar:

```bash
kubectl get ingress -w
```

Quando aparecer um endereço em `ADDRESS`, pressione Ctrl+C.

**Copiar a URL:**
```bash
INGRESS_URL=$(kubectl get ingress velure-ui -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
echo "Aplicação disponível em: http://$INGRESS_URL"
```

---

## ✅ FASE 3: Testes e Validações (~10 minutos)

### 3.1. Testar Health Checks dos Serviços

```bash
# Obter URL do Load Balancer
INGRESS_URL=$(kubectl get ingress velure-ui -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

# Testar cada serviço
curl http://$INGRESS_URL/api/auth/health
curl http://$INGRESS_URL/api/product/health
curl http://$INGRESS_URL/api/order/health
```

Todos devem retornar: `{"status":"healthy"}` ou similar.

### 3.2. Testar UI no Navegador

```bash
# Abrir no navegador
echo "Abra no navegador: http://$INGRESS_URL"
```

Ou simplesmente copie a URL e cole no navegador.

Você deve ver a interface do Velure funcionando!

### 3.3. Testar Fluxo Completo (Opcional)

**Registro de usuário:**
```bash
curl -X POST http://$INGRESS_URL/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "senha123"
  }'
```

**Login:**
```bash
curl -X POST http://$INGRESS_URL/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "senha123"
  }'
```

Deve retornar um token JWT.

**Listar produtos:**
```bash
curl http://$INGRESS_URL/api/product/products
```

### 3.4. Acessar Grafana (Monitoramento)

```bash
# Port-forward do Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
```

Abra no navegador: **http://localhost:3000**
- Usuário: `admin`
- Senha: `admin` (vai pedir para trocar)

Explore os dashboards:
- **Velure Overview** - Visão geral dos serviços
- **Kubernetes / Compute Resources / Cluster** - Uso do cluster

### 3.5. Acessar RabbitMQ Management (Opcional)

```bash
kubectl port-forward -n datastores svc/velure-datastores-rabbitmq 15672:15672
```

Abra: **http://localhost:15672**
- Usuário: `admin`
- Senha: `admin_password` (ou a que foi configurada)

### 3.6. Ver Logs dos Serviços

```bash
# Logs do auth-service
kubectl logs -f -l app.kubernetes.io/name=velure-auth

# Logs do product-service
kubectl logs -f -l app.kubernetes.io/name=velure-product

# Logs de um pod específico
kubectl logs -f velure-auth-<tab-para-completar>
```

---

## 🧪 FASE 4: Testes Locais (Opcional)

### 4.1. Testar com k6 (Load Testing)

```bash
cd tests/load

# Configurar URL base
export BASE_URL=http://$INGRESS_URL

# Rodar teste integrado
./run-all-tests.sh
```

### 4.2. Monitorar HPA (Autoscaling)

```bash
# Em um terminal, rodar teste de carga
cd tests/load
./run-all-tests.sh

# Em outro terminal, monitorar scaling
watch kubectl get hpa
```

Você verá os pods escalando automaticamente quando a carga aumentar!

---

## 🗑️ Destruir Tudo (Quando Terminar)

**⚠️ ATENÇÃO: Isso deletará TUDO, incluindo dados!**

```bash
# 1. Deletar ingresses e load balancers
kubectl delete ingress --all -A
kubectl delete svc --field-selector spec.type=LoadBalancer -A

# Aguarde ~3 minutos para ALBs serem deletados
sleep 180

# 2. Deletar aplicações
helm uninstall velure-auth velure-product velure-publish-order velure-process-order velure-ui -n default

# 3. Deletar monitoramento
helm uninstall kube-prometheus-stack -n monitoring
kubectl delete namespace monitoring

# 4. Deletar datastores
helm uninstall velure-datastores -n datastores
kubectl delete pvc --all -n datastores
kubectl delete namespace datastores

# 5. Deletar controllers
helm uninstall aws-load-balancer-controller -n kube-system
kubectl delete -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# 6. Destruir infraestrutura AWS
cd infrastructure/terraform
terraform destroy
```

Digite `yes` quando perguntar.

---

## 🛠️ Troubleshooting Comum

### Problema: Pods em CrashLoopBackOff

```bash
# Ver logs do pod
kubectl logs <nome-do-pod>

# Descrever o pod (ver eventos)
kubectl describe pod <nome-do-pod>
```

**Causas comuns:**
- Secrets não configurados corretamente
- Banco de dados inacessível
- Variáveis de ambiente faltando

### Problema: Ingress sem ADDRESS

```bash
# Ver logs do ALB controller
kubectl logs -n kube-system -l app.kubernetes.io/name=aws-load-balancer-controller
```

**Causas comuns:**
- IAM role do ALB controller sem permissões
- Subnets sem tags corretas
- Security groups bloqueando

### Problema: Não consigo conectar ao RDS

```bash
# Testar de dentro de um pod
kubectl run test --rm -it --image=postgres:15 -- \
  psql -h <rds-endpoint> -U postgres -d velure_auth
```

**Causas comuns:**
- Security group do RDS não permite conexões do EKS
- Endpoint RDS incorreto
- Senha incorreta

### Problema: 502 Bad Gateway no ALB

```bash
# Verificar se os pods estão rodando
kubectl get pods

# Verificar health checks
kubectl describe svc velure-ui
```

**Causas comuns:**
- Pods não estão prontos (ainda iniciando)
- Health checks falhando
- Porta incorreta no Service

---

## 💰 Detalhamento de Custos

### Custo Mensal (us-east-1)

| Recurso | Specs | Custo/mês |
|---------|-------|-----------|
| **EKS Cluster** | 1 cluster | $72.00 |
| **EC2 Nodes** | 2x t3.small | $30.00 |
| **RDS PostgreSQL** | 2x db.t4g.micro | $0.00 (Free Tier*) |
| **EBS Volumes** | ~50GB gp3 | $4.00 |
| **ALB** | 1 Application LB | $16.00 |
| **NAT Gateway** | 1 NAT + data | $32.00 |
| **Data Transfer** | <1GB/dia | $2.00 |
| **CloudWatch Logs** | ~5GB/mês | $2.50 |
| **TOTAL** | | **~$158/mês** |

\* *Free Tier válido por 12 meses, 750h/mês. Após isso: ~$25/mês por instância.*

### Como Economizar

**1. Parar nodes quando não usar** (economia: ~$30/mês)
```bash
# Escalar para 0
kubectl scale deployment --all --replicas=0

# Ou parar os nodes EC2 no console AWS
```

**2. Usar 1 node apenas** (economia: ~$15/mês)
```hcl
# terraform.tfvars
eks_node_desired_size = 1
eks_node_min_size     = 1
eks_node_max_size     = 1
```

**3. Usar t3.micro nos nodes** (economia: ~$15/mês)
```hcl
eks_node_instance_type = "t3.micro"
```

**4. Deletar completamente nos fins de semana**
```bash
# Sexta à noite
terraform destroy

# Segunda de manhã
terraform apply
```

**5. Configurar Budget Alert**
```bash
# No console AWS, ir em Billing > Budgets
# Criar alerta para $150/mês
```

---

## 📊 Comandos Úteis

### Status Geral
```bash
# Ver todos os pods
kubectl get pods -A

# Ver nodes
kubectl get nodes

# Ver serviços
kubectl get svc -A

# Ver ingresses
kubectl get ingress -A

# Ver uso de recursos
kubectl top nodes
kubectl top pods -A
```

### Logs
```bash
# Logs de um deployment
kubectl logs -f deployment/velure-auth

# Logs de todos os pods de um serviço
kubectl logs -f -l app.kubernetes.io/name=velure-auth

# Logs de um pod específico
kubectl logs -f velure-auth-xxxxxxxxx-xxxxx
```

### Exec em Pods
```bash
# Shell em um pod
kubectl exec -it velure-auth-xxxxxxxxx-xxxxx -- /bin/sh

# Rodar comando único
kubectl exec velure-product-xxxxxxxxx-xxxxx -- env
```

### Port-Forward
```bash
# Acessar Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80

# Acessar Prometheus
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090

# Acessar RabbitMQ
kubectl port-forward -n datastores svc/velure-datastores-rabbitmq 15672:15672

# Acessar MongoDB
kubectl port-forward -n datastores svc/velure-datastores-mongodb 27017:27017

# Acessar Redis
kubectl port-forward -n datastores svc/velure-datastores-redis-master 6379:6379
```

### Restart de Serviços
```bash
# Restart de um deployment
kubectl rollout restart deployment/velure-auth

# Ver status do rollout
kubectl rollout status deployment/velure-auth
```

---

## 📚 Próximos Passos

Depois que tudo estiver funcionando:

1. **Configurar DNS customizado**
   - Comprar domínio
   - Apontar para o ALB
   - Configurar TLS/HTTPS com cert-manager

2. **Configurar CI/CD**
   - GitHub Actions para build automático
   - Deploy automático no EKS

3. **Implementar backups**
   - RDS automated backups
   - Snapshots dos volumes EBS

4. **Segurança avançada**
   - Network Policies
   - Pod Security Standards
   - External Secrets Operator (AWS Secrets Manager)

5. **Observabilidade**
   - Alertas no Grafana
   - Integração com PagerDuty/Slack
   - Distributed tracing (Jaeger)

---

## 📖 Documentação Adicional

- **Deploy Detalhado**: [docs/DEPLOY_GUIDE.md](docs/DEPLOY_GUIDE.md)
- **Métricas Prometheus**: [docs/PROMETHEUS_METRICS.md](docs/PROMETHEUS_METRICS.md)
- **Monitoramento**: [docs/MONITORING.md](docs/MONITORING.md)
- **Troubleshooting**: [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
- **Terraform**: [infrastructure/terraform/README.md](infrastructure/terraform/README.md)

---

## 🆘 Suporte

Se encontrar problemas:

1. Verifique os logs: `kubectl logs -f <pod-name>`
2. Verifique eventos: `kubectl get events --sort-by='.lastTimestamp'`
3. Consulte [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
4. Abra uma issue no GitHub

---

**Feito com ❤️ para aprender e compartilhar conhecimento sobre cloud-native!**
