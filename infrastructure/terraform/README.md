# Velure - Terraform Infrastructure

Infraestrutura como código para deploy do Velure na AWS usando EKS.

## 📋 Pré-requisitos

```bash
# AWS CLI v2
aws --version  # >= 2.0.0
aws configure  # Configurar credenciais

# Terraform
terraform --version  # >= 1.6.0

# kubectl
kubectl version --client  # >= 1.28.0

# Helm
helm version  # >= 3.0.0
```

## 💰 Estimativa de Custos

**AVISO**: Este é um setup otimizado para projetos pessoais, mas ainda gera custos.

| Recurso | Especificação | Custo Mensal (us-east-1) |
|---------|--------------|--------------------------|
| EKS Cluster | 1 cluster | $72.00 |
| EC2 Nodes | 2x t3.small (on-demand) | ~$30.00 |
| NAT Gateway | 1x + data transfer | ~$32.00 + transfer |
| RDS Auth | db.t4g.micro (Free Tier) | $0.00 (750h/mês) |
| RDS Orders | db.t4g.micro (Free Tier) | $0.00 (750h/mês) |
| EBS Volumes | 2x 20GB gp3 (nodes) | ~$3.20 |
| VPC | 1 VPC + subnets | $0.00 |
| CloudWatch Logs | ~5GB/mês | ~$2.50 |
| **TOTAL** | | **~$140-150/mês** |

### ⚠️ Free Tier (primeiro ano AWS)
- RDS: 750h/mês de db.t4g.micro (suficiente para 1 instância 24/7)
- EBS: 30GB de armazenamento gp3

### 💡 Dicas para Redução de Custos
1. **Pare os nodes quando não estiver usando**: `kubectl scale deployment --all --replicas=0`
2. **Delete a infra nos finais de semana**: `terraform destroy`
3. **Use Spot Instances**: Trocar node_instance_type para spot (economia de ~70%)
4. **Monitore custos**: AWS Cost Explorer + Budget Alerts

## 🚀 Deploy

### 1. Clonar e Configurar

```bash
cd terraform/

# Copiar exemplo de variáveis
cp terraform.tfvars.example terraform.tfvars

# Editar variáveis (principalmente senhas!)
vim terraform.tfvars
```

### 2. Inicializar Terraform

```bash
terraform init
```

### 3. Validar Configuração

```bash
terraform validate
terraform fmt -recursive
```

### 4. Revisar Plano

```bash
terraform plan -out=tfplan

# Revisar cuidadosamente:
# - Recursos que serão criados
# - Custos estimados
# - Security groups
```

### 5. Aplicar Infraestrutura

```bash
terraform apply tfplan

# Aguardar ~15-20 minutos
# EKS cluster demora para criar
```

### 6. Configurar kubectl

```bash
# Obter comando para configurar kubectl
terraform output -raw kubeconfig_command | bash

# Testar conectividade
kubectl get nodes
kubectl get pods -A
```

## 🔧 Pós-Deploy

### 1. Instalar AWS Load Balancer Controller

```bash
# Criar ServiceAccount com IRSA
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-load-balancer-controller
  namespace: kube-system
  annotations:
    eks.amazonaws.com/role-arn: $(terraform output -raw alb_controller_role_arn)
EOF

# Instalar via Helm
helm repo add eks https://aws.github.io/eks-charts
helm repo update

helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=$(terraform output -raw eks_cluster_name) \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller

# Verificar
kubectl get deployment -n kube-system aws-load-balancer-controller
```

### 2. Instalar Redis (In-Cluster)

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami

helm install redis bitnami/redis \
  --set architecture=standalone \
  --set auth.password="$(openssl rand -base64 32)" \
  --set master.persistence.size=1Gi \
  --set master.resources.requests.memory=256Mi \
  --set master.resources.requests.cpu=100m \
  --set master.resources.limits.memory=512Mi \
  --set master.resources.limits.cpu=200m

# Obter senha
export REDIS_PASSWORD=$(kubectl get secret redis -o jsonpath="{.data.redis-password}" | base64 -d)
echo "Redis Password: $REDIS_PASSWORD"
```

### 3. Instalar RabbitMQ (In-Cluster)

```bash
helm install rabbitmq bitnami/rabbitmq \
  --set auth.username=admin \
  --set auth.password="$(openssl rand -base64 32)" \
  --set persistence.size=2Gi \
  --set resources.requests.memory=256Mi \
  --set resources.requests.cpu=100m \
  --set resources.limits.memory=512Mi \
  --set resources.limits.cpu=200m

# Obter senha
export RABBITMQ_PASSWORD=$(kubectl get secret rabbitmq -o jsonpath="{.data.rabbitmq-password}" | base64 -d)
echo "RabbitMQ Password: $RABBITMQ_PASSWORD"
```

### 4. Configurar Secrets das Aplicações

```bash
# RDS Auth Service
kubectl create secret generic auth-db-secret \
  --from-literal=username=postgres \
  --from-literal=password="$(terraform output -raw rds_auth_password)" \
  --from-literal=host="$(terraform output -raw rds_auth_address)" \
  --from-literal=port=5432 \
  --from-literal=database=velure_auth

# RDS Orders Service
kubectl create secret generic orders-db-secret \
  --from-literal=username=postgres \
  --from-literal=password="$(terraform output -raw rds_orders_password)" \
  --from-literal=host="$(terraform output -raw rds_orders_address)" \
  --from-literal=port=5432 \
  --from-literal=database=velure_orders
```

## 📊 Monitoramento

### CloudWatch Logs

```bash
# Verificar logs do EKS
aws logs tail /aws/eks/velure-cluster/cluster --follow

# Verificar logs do RDS
aws logs tail /aws/rds/instance/velure-auth-db/postgresql --follow
```

### Kubernetes

```bash
# Nodes
kubectl top nodes

# Pods
kubectl top pods -A

# Events
kubectl get events -A --sort-by='.lastTimestamp'
```

## 🛠️ Troubleshooting

### Nodes não conectam ao cluster

```bash
# Verificar security groups
aws ec2 describe-security-groups \
  --group-ids $(terraform output -raw eks_node_security_group_id)

# Verificar logs do node
aws ec2 get-console-output --instance-id <instance-id>
```

### RDS inacessível

```bash
# Testar conectividade de um pod
kubectl run -it --rm debug --image=postgres:16-alpine --restart=Never -- \
  psql -h $(terraform output -raw rds_auth_address) -U postgres -d velure_auth

# Verificar security group
aws ec2 describe-security-groups \
  --group-ids $(terraform output -raw rds_security_group_id)
```

### ALB não cria

```bash
# Verificar logs do controller
kubectl logs -n kube-system deployment/aws-load-balancer-controller

# Verificar IAM role
aws iam get-role --role-name $(terraform output -raw alb_controller_role_name)
```

## 🗑️ Destruir Infraestrutura

```bash
# AVISO: Isso deletará TUDO, incluindo dados do RDS!

# 1. Deletar LoadBalancers criados pelo controller
kubectl delete ingress --all -A
kubectl delete service --field-selector spec.type=LoadBalancer -A

# 2. Aguardar ALBs serem deletados (~2 minutos)
aws elbv2 describe-load-balancers --query 'LoadBalancers[].LoadBalancerName'

# 3. Destruir com Terraform
terraform destroy

# Confirmar com "yes"
```

## 📁 Estrutura

```
terraform/
├── main.tf                    # Root module
├── variables.tf               # Input variables
├── outputs.tf                 # Outputs
├── versions.tf                # Provider versions
├── terraform.tfvars.example   # Example configuration
└── modules/
    ├── vpc/                   # VPC, subnets, NAT
    ├── security-groups/       # Security groups
    ├── eks/                   # EKS cluster + nodes
    └── rds/                   # PostgreSQL databases
```

## 🔐 Security Best Practices

- ✅ IMDSv2 enforced nos nodes
- ✅ Security groups com least privilege
- ✅ RDS em private subnet
- ✅ Encryption at rest (EBS + RDS)
- ✅ CloudWatch logs habilitados
- ✅ IRSA (IAM Roles for Service Accounts)
- ✅ Secrets via Kubernetes Secrets (considere External Secrets Operator)
- ✅ Network policies (a implementar)

## 📚 Próximos Passos

1. **External Secrets Operator**: Integrar com AWS Secrets Manager
2. **Network Policies**: Isolar comunicação entre pods
3. **Pod Security Standards**: Enforçar PSS restricted
4. **Prometheus + Grafana**: Monitoramento avançado
5. **ArgoCD**: GitOps para deployments
6. **Cert-Manager**: TLS automático
7. **Karpenter**: Autoscaling mais eficiente que Cluster Autoscaler

## 🔗 Referências

- [EKS Best Practices](https://aws.github.io/aws-eks-best-practices/)
- [AWS Load Balancer Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
