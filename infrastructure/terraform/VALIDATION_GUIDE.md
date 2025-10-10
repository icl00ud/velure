# Guia de Validação - Terraform Velure

## ✅ Checklist antes de aplicar

### 1. Pré-requisitos instalados

```bash
# Terraform
terraform version  # Deve ser >= 1.6.0
# Se não tiver, instalar: https://www.terraform.io/downloads

# AWS CLI v2
aws --version  # Deve ser >= 2.0.0
# Se não tiver, instalar: https://aws.amazon.com/cli/

# kubectl
kubectl version --client  # Deve ser >= 1.28.0
# Se não tiver, instalar: https://kubernetes.io/docs/tasks/tools/

# Helm
helm version  # Deve ser >= 3.0.0
# Se não tiver, instalar: https://helm.sh/docs/intro/install/
```

### 2. Configurar credenciais AWS

```bash
# Configurar AWS CLI
aws configure
# AWS Access Key ID: [sua-key]
# AWS Secret Access Key: [seu-secret]
# Default region name: us-east-1
# Default output format: json

# Verificar credenciais
aws sts get-caller-identity
```

### 3. Configurar variáveis do Terraform

```bash
cd terraform/

# Copiar exemplo
cp terraform.tfvars.example terraform.tfvars

# IMPORTANTE: Editar e alterar senhas!
vim terraform.tfvars

# Itens OBRIGATÓRIOS para alterar:
# - rds_auth_password
# - rds_orders_password
# - tags.Owner (seu nome/email)
```

### 4. Validar sintaxe Terraform

```bash
# Formatar arquivos
terraform fmt -recursive

# Inicializar
terraform init

# Validar configuração
terraform validate

# Deve retornar: Success! The configuration is valid.
```

### 5. Revisar plano de execução

```bash
# Gerar plano
terraform plan -out=tfplan

# Revisar CUIDADOSAMENTE:
# ✅ 50+ recursos serão criados
# ✅ 0 recursos serão destruídos
# ✅ 0 recursos serão modificados
# ✅ Nenhum erro de validação
# ✅ Todas as variáveis resolvidas
```

**Recursos esperados:**
- 1 VPC
- 3 Subnets (1 public, 2 private)
- 1 Internet Gateway
- 1 NAT Gateway
- 1 Elastic IP
- Route tables e associações
- VPC Flow Logs + CloudWatch Log Group + IAM Role
- 3 Security Groups (EKS nodes, RDS, ALB)
- 1 EKS Cluster
- 1 EKS Node Group
- 1 Launch Template
- IAM Roles e Policies (cluster, nodes, ALB controller, EBS CSI)
- OIDC Provider
- 4 EKS Addons (VPC CNI, CoreDNS, kube-proxy, EBS CSI)
- 2 RDS Instances
- 2 DB Subnet Groups
- 2 DB Parameter Groups
- 4 CloudWatch Log Groups para RDS

### 6. Estimativa de custos

```bash
# Revisar COST_ESTIMATION.md antes de aplicar!
cat COST_ESTIMATION.md

# Custo esperado: ~$143/mês (com Spot) ou ~$164/mês (on-demand)
```

### 7. Aplicar infraestrutura

```bash
# ATENÇÃO: Isso VAI CRIAR recursos cobráveis na AWS!
terraform apply tfplan

# Aguardar ~15-20 minutos
# EKS cluster é o recurso mais demorado
```

### 8. Verificar outputs

```bash
# Ver todos os outputs
terraform output

# Outputs específicos
terraform output eks_cluster_endpoint
terraform output rds_auth_address
terraform output kubeconfig_command
```

### 9. Configurar kubectl

```bash
# Usar comando do output
terraform output -raw kubeconfig_command | bash

# Ou manualmente
aws eks update-kubeconfig --region us-east-1 --name velure-prod

# Verificar
kubectl get nodes
# Deve mostrar 2 nodes em status Ready

kubectl get pods -A
# Deve mostrar pods do sistema (coredns, kube-proxy, etc)
```

### 10. Verificar recursos criados

```bash
# EKS Cluster
aws eks describe-cluster --name velure-prod --region us-east-1

# RDS Instances
aws rds describe-db-instances --region us-east-1 | grep DBInstanceIdentifier

# VPC
aws ec2 describe-vpcs --region us-east-1 | grep velure

# NAT Gateway
aws ec2 describe-nat-gateways --region us-east-1 | grep velure
```

---

## 🔍 Troubleshooting

### Erro: "Error creating EKS Cluster"

**Possíveis causas:**
1. IAM permissions insuficientes
2. Service limits atingidos
3. Subnet sem espaço IP suficiente

**Solução:**
```bash
# Verificar permissões IAM
aws iam get-user

# Verificar service limits
aws service-quotas list-service-quotas --service-code eks

# Verificar subnets
aws ec2 describe-subnets --filters "Name=tag:Name,Values=*velure*"
```

### Erro: "Error creating RDS Cluster: DBSubnetGroupDoesNotCoverEnoughAZs"

**Causa:** RDS precisa de pelo menos 2 AZs diferentes

**Solução:** Já está implementado! Temos 2 subnets privadas em us-east-1a e us-east-1b

### Erro: "InvalidParameterException: Node IAM role cannot be assumed"

**Causa:** IAM role não tem trust relationship correto

**Solução:**
```bash
# Aguardar alguns minutos para propagação IAM
# Ou verificar assume role policy
aws iam get-role --role-name velure-prod-eks-node-role
```

### Erro: "Error creating Security Group Rule: duplicate rule"

**Causa:** Security group rule já existe

**Solução:**
```bash
# Limpar state e reimportar
terraform state rm module.security_groups.aws_security_group_rule.xxx
terraform import module.security_groups.aws_security_group_rule.xxx sgr-xxxxx
```

### Nodes não aparecem no cluster

**Causa:** User data script falhou ou security groups incorretos

**Solução:**
```bash
# Verificar logs do node
# 1. Pegar instance ID
aws ec2 describe-instances --filters "Name=tag:Name,Values=*velure*" \
  --query 'Reservations[].Instances[].InstanceId'

# 2. Ver console output
aws ec2 get-console-output --instance-id i-xxxxx

# 3. Verificar security groups
aws eks describe-cluster --name velure-prod \
  --query 'cluster.resourcesVpcConfig.securityGroupIds'
```

### RDS inacessível dos pods

**Causa:** Security group não permite tráfego do EKS

**Solução:**
```bash
# Testar conectividade de um pod
kubectl run -it --rm psql-test --image=postgres:16-alpine --restart=Never -- \
  psql -h $(terraform output -raw rds_auth_address) \
       -U postgres \
       -d velure_auth

# Se falhar, verificar security groups
aws ec2 describe-security-group-rules \
  --filters "Name=group-id,Values=$(terraform output -raw rds_security_group_id)"
```

---

## 🧹 Limpeza / Destroy

### ATENÇÃO: Isso DELETARÁ TUDO, incluindo dados!

```bash
# 1. Deletar resources criados pelo Kubernetes primeiro
kubectl delete ingress --all -A
kubectl delete service --field-selector spec.type=LoadBalancer -A
kubectl delete pvc --all -A

# 2. Aguardar LoadBalancers serem deletados
aws elbv2 describe-load-balancers | grep velure
# Deve retornar vazio

# 3. Deletar EBS volumes órfãos (se houver)
aws ec2 describe-volumes \
  --filters "Name=tag:kubernetes.io/cluster/velure-prod,Values=owned" \
  --query 'Volumes[].VolumeId'

# 4. Destruir com Terraform
terraform destroy

# Confirmar digitando: yes

# 5. Verificar que tudo foi deletado
aws eks list-clusters --region us-east-1
aws rds describe-db-instances --region us-east-1
aws ec2 describe-vpcs --region us-east-1 --filters "Name=tag:Name,Values=*velure*"
```

---

## 📋 Checklist de Validação Completa

- [ ] Terraform >= 1.6 instalado
- [ ] AWS CLI v2 instalado e configurado
- [ ] kubectl instalado
- [ ] Helm instalado
- [ ] Credenciais AWS configuradas (`aws sts get-caller-identity`)
- [ ] terraform.tfvars criado e senhas alteradas
- [ ] `terraform init` executado com sucesso
- [ ] `terraform validate` retornou sucesso
- [ ] `terraform plan` revisado (50+ recursos)
- [ ] COST_ESTIMATION.md lido e custos entendidos
- [ ] `terraform apply` executado
- [ ] EKS cluster criado (15-20 min)
- [ ] Nodes aparecem com `kubectl get nodes`
- [ ] RDS instances criadas
- [ ] Security groups corretos
- [ ] NAT Gateway funcionando

## 🎯 Próximos Passos

Após validação bem-sucedida:

1. ✅ Instalar AWS Load Balancer Controller (veja README.md)
2. ✅ Instalar Redis via Helm
3. ✅ Instalar RabbitMQ via Helm
4. ✅ Criar Kubernetes Secrets para databases
5. ✅ Deployar microserviços
6. ✅ Configurar Ingress para ALB
7. ✅ Testar aplicação end-to-end

---

**Boa sorte! 🚀**
