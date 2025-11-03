# Guia de Deploy - Velure no EKS

Este guia fornece instruções passo a passo para fazer o deploy completo da aplicação Velure no Amazon EKS.

## 📋 Pré-requisitos

### Ferramentas Necessárias

- **AWS CLI** (v2+): `aws --version`
- **kubectl** (v1.28+): `kubectl version --client`
- **helm** (v3.12+): `helm version`
- **eksctl** (v0.150+): `eksctl version`
- **terraform** (v1.5+): `terraform version`

### Credenciais AWS

```bash
aws configure
# Ou usar AWS_PROFILE
export AWS_PROFILE=your-profile
```

### Conta AWS

- Permissões para criar recursos EKS, VPC, RDS, etc.
- Budget suficiente (~$120-150/mês)

## 🚀 Deploy Rápido (5 Comandos)

```bash
# 1. Deploy infraestrutura Terraform
cd infrastructure/terraform
terraform init && terraform apply

# 2. Configure kubectl
aws eks update-kubeconfig --region us-east-1 --name velure-prod

# 3. Execute scripts de deploy
cd ../../scripts/deploy
./01-install-controllers.sh   # ALB + metrics-server
./02-install-datastores.sh    # MongoDB, Redis, RabbitMQ
./03-install-monitoring.sh    # Prometheus + Grafana
./04-deploy-services.sh       # Microserviços
```

Pronto! Sua aplicação está no ar. 🎉

## 📖 Deploy Detalhado

### Fase 1: Infraestrutura AWS (30-40 minutos)

#### 1.1 Deploy com Terraform

```bash
cd infrastructure/terraform

# Inicializar
terraform init

# Planejar (revisar recursos)
terraform plan

# Aplicar
terraform apply
```

**O que será criado:**
- VPC com subnets públicas e privadas
- Cluster EKS (1.28) com 2 nodes t3.small
- 2x RDS PostgreSQL (db.t4g.micro Free Tier)
- Security Groups
- IAM Roles (EKS, nodes, ALB controller)
- EBS CSI Driver addon

**Tempo estimado:** 30-40 minutos

#### 1.2 Configurar kubectl

```bash
aws eks update-kubeconfig --region us-east-1 --name velure-prod

# Verificar conexão
kubectl cluster-info
kubectl get nodes
```

### Fase 2: Controllers Essenciais (10 minutos)

#### 2.1 Instalar Controllers

```bash
cd scripts/deploy
./01-install-controllers.sh
```

**O que será instalado:**
- **metrics-server**: Para HPA (Horizontal Pod Autoscaler)
- **AWS Load Balancer Controller**: Para Ingress com ALB

**Verificação:**
```bash
kubectl get deployment -n kube-system metrics-server
kubectl get deployment -n kube-system aws-load-balancer-controller
```

### Fase 3: Datastores (15-20 minutos)

#### 3.1 Deploy Datastores

```bash
./02-install-datastores.sh
```

**O que será instalado:**
- **MongoDB** (27017): Product catalog
- **Redis** (6379): Caching
- **RabbitMQ** (5672, 15672): Message queue

**Verificação:**
```bash
kubectl get pods -n datastores
kubectl get pvc -n datastores

# Testar MongoDB
kubectl exec -n datastores velure-datastores-mongodb-0 -- mongosh --eval "db.adminCommand('ping')"

# Testar Redis
kubectl exec -n datastores velure-datastores-redis-master-0 -- redis-cli -a redis_password ping

# Acessar RabbitMQ Management UI
kubectl port-forward -n datastores svc/velure-datastores-rabbitmq 15672:15672
# Abra: http://localhost:15672 (admin/admin_password)
```

### Fase 4: Monitoramento (15 minutos)

#### 4.1 Instalar Stack de Monitoramento

```bash
./03-install-monitoring.sh
```

**O que será instalado:**
- **Prometheus**: Coleta de métricas
- **Grafana**: Visualização
- **Alertmanager**: Gerenciamento de alertas
- **ServiceMonitors**: Scraping dos microserviços

**Verificação:**
```bash
kubectl get pods -n monitoring

# Acessar Prometheus
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
# http://localhost:9090

# Acessar Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
# http://localhost:3000 (admin/admin)
```

### Fase 5: Deploy dos Serviços (10-15 minutos)

#### 5.1 Preparar Secrets

Antes de executar o script, edite os valores dos secrets:

```bash
# O script criará secrets com valores placeholder
# Você precisa editá-los com os valores reais

kubectl edit secret jwt-secret -n default
# Altere jwt-secret e jwt-refresh-secret

kubectl edit secret database-secrets -n default
# Altere URLs do RDS (obtenha do Terraform output)
```

**Obter URLs do RDS:**
```bash
cd infrastructure/terraform
terraform output rds_auth_endpoint
terraform output rds_orders_endpoint
```

#### 5.2 Deploy dos Serviços

```bash
cd scripts/deploy
./04-deploy-services.sh
```

**O que será deployado:**
- **velure-auth**: Autenticação (3020)
- **velure-product**: Catálogo (3010)
- **velure-publish-order**: Criação de pedidos (8080)
- **velure-process-order**: Processamento (8081)
- **velure-ui**: Frontend React (80)

**Verificação:**
```bash
kubectl get pods
kubectl get svc
kubectl get ingress

# Ver logs de um serviço
kubectl logs -f -l app.kubernetes.io/name=velure-auth
```

#### 5.3 Aguardar LoadBalancer

O ALB leva alguns minutos para provisionar:

```bash
kubectl get ingress -w

# Quando aparecer ADDRESS, a aplicação está acessível
```

### Fase 6: Verificação Final

#### 6.1 Testar Aplicação

```bash
# Obter URL
INGRESS_URL=$(kubectl get ingress velure-ui -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

echo "Aplicação disponível em: http://$INGRESS_URL"

# Testar endpoints
curl http://$INGRESS_URL/api/auth/health
curl http://$INGRESS_URL/api/product/health
curl http://$INGRESS_URL/api/order/health
```

#### 6.2 Verificar Métricas no Prometheus

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
```

Acesse http://localhost:9090/targets e verifique se todos os targets estão UP:
- velure-auth
- velure-product
- velure-publish-order
- velure-process-order

#### 6.3 Verificar Dashboards no Grafana

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
```

Acesse http://localhost:3000 (admin/admin) e verifique os dashboards.

## 🔧 Troubleshooting

### Pods em CrashLoopBackOff

```bash
# Ver logs
kubectl logs <pod-name>

# Descrever pod
kubectl describe pod <pod-name>

# Verificar eventos
kubectl get events --sort-by='.lastTimestamp'
```

### LoadBalancer sem External-IP

```bash
# Verificar events do service
kubectl describe svc velure-ui

# Verificar logs do ALB controller
kubectl logs -n kube-system -l app.kubernetes.io/name=aws-load-balancer-controller
```

### Métricas não aparecem no Prometheus

```bash
# Verificar ServiceMonitors
kubectl get servicemonitor -A

# Verificar se o pod expõe /metrics
kubectl port-forward <pod-name> 8080:8080
curl http://localhost:8080/metrics

# Ver configuração do Prometheus
kubectl get prometheus -n monitoring -o yaml
```

### RDS Connection Refused

```bash
# Verificar Security Groups
aws ec2 describe-security-groups --group-ids <rds-sg-id>

# Verificar se nodes podem acessar RDS
kubectl run test --rm -it --image=postgres:15 -- psql -h <rds-endpoint> -U postgres
```

## 📊 Monitoramento

### Acessar Grafana

```bash
# Via port-forward
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80

# Ou via LoadBalancer (se configurado)
kubectl get svc -n monitoring kube-prometheus-stack-grafana
```

**Dashboards disponíveis:**
- Velure Overview: Visão geral de todos os serviços
- Auth Service: Métricas de autenticação
- Product Service: Catálogo e cache
- Orders: Pedidos e pagamentos
- Infrastructure: Cluster Kubernetes

### Queries PromQL Úteis

```promql
# Taxa de requisições por serviço
sum(rate(auth_http_requests_total[5m])) by (status)

# Latência p95
histogram_quantile(0.95, rate(auth_http_request_duration_seconds_bucket[5m]))

# Taxa de erros
rate(auth_errors_total[5m])

# Cache hit rate
rate(product_cache_hits_total[5m]) /
(rate(product_cache_hits_total[5m]) + rate(product_cache_misses_total[5m]))
```

## 🗑️ Cleanup (Destruir Tudo)

```bash
# 1. Deletar serviços
helm uninstall velure-auth velure-product velure-publish-order velure-process-order velure-ui -n default

# 2. Deletar monitoramento
helm uninstall kube-prometheus-stack -n monitoring
kubectl delete namespace monitoring

# 3. Deletar datastores
helm uninstall velure-datastores -n datastores
kubectl delete pvc --all -n datastores
kubectl delete namespace datastores

# 4. Deletar controllers
helm uninstall aws-load-balancer-controller -n kube-system
kubectl delete -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# 5. Destruir infraestrutura
cd infrastructure/terraform
terraform destroy
```

**⚠️ ATENÇÃO:** Isso deletará TODOS os recursos e dados!

## 💰 Estimativa de Custos

### Custo Mensal Estimado (~$124/mês)

| Recurso | Quantidade | Custo/mês |
|---------|-----------|-----------|
| EKS Cluster | 1 | $72.00 |
| EC2 t3.small (nodes) | 2 | ~$30.00 |
| RDS db.t4g.micro | 2 | $0.00 (Free Tier) |
| EBS gp3 (50GB) | - | ~$4.00 |
| ALB | 1 | ~$16.00 |
| Data Transfer | <1GB/dia | ~$2.00 |
| **TOTAL** | | **~$124/mês** |

### Reduzir Custos

**Opção 1: Usar t3.micro nodes** (~$15/mês economia)
```hcl
# terraform.tfvars
eks_node_instance_type = "t3.micro"
```

**Opção 2: 1 node apenas** (~$15/mês economia)
```hcl
eks_node_desired_size = 1
eks_node_min_size = 1
eks_node_max_size = 1
```

**Opção 3: Usar Fargate** (pague por uso)
- Requer modificações no Terraform
- ~$30-50/mês para workload small

## 📚 Referências

- [Documentação do Projeto](../README.md)
- [Métricas Prometheus](./PROMETHEUS_METRICS.md)
- [Monitoramento](./MONITORING.md)
- [Troubleshooting](./TROUBLESHOOTING.md)
- [Terraform AWS EKS](https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [Helm Documentation](https://helm.sh/docs/)
