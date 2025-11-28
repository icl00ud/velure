# PgBouncer Implementation Guide

Este diretório contém configurações para implementar PgBouncer no Kubernetes.

## 🎯 Opções de Implementação

Você tem **3 opções** para implementar PgBouncer:

### Opção 1: AWS RDS Proxy (Mais Fácil) ⭐ RECOMENDADO PARA COMEÇAR
### Opção 2: PgBouncer Centralizado no Kubernetes (Melhor Long-term)
### Opção 3: PgBouncer Sidecar (Por Pod)

---

## 📌 Opção 1: AWS RDS Proxy (Gerenciado pela AWS)

### O Que É?
AWS RDS Proxy é um **PgBouncer gerenciado pela AWS**. Você não precisa gerenciar nada.

### Vantagens
- ✅ **Zero manutenção** - AWS gerencia tudo
- ✅ **Alta disponibilidade** automática
- ✅ **Failover automático** entre RDS instances
- ✅ **Compatível com IAM authentication**
- ✅ **Logs no CloudWatch** automáticos
- ✅ **Não precisa mudar código Kubernetes**

### Desvantagens
- ⚠️ **Custo adicional** (~$0.015/hora por vCPU)
- ⚠️ **Somente AWS** (vendor lock-in)
- ⚠️ **Menos controle** sobre configurações

### Como Implementar

#### Via Terraform

**Arquivo:** `infrastructure/terraform/rds-proxy.tf` (criar novo arquivo)

```hcl
# RDS Proxy para auth-service database
resource "aws_db_proxy" "velure_auth_proxy" {
  name                   = "velure-auth-proxy"
  engine_family          = "POSTGRESQL"
  auth {
    auth_scheme = "SECRETS"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.rds_auth_credentials.arn
  }

  role_arn               = aws_iam_role.rds_proxy_role.arn
  vpc_subnet_ids         = module.vpc.private_subnets
  require_tls            = true

  tags = {
    Name        = "velure-auth-proxy"
    Environment = var.environment
  }
}

# Target group apontando para RDS
resource "aws_db_proxy_default_target_group" "velure_auth_proxy_tg" {
  db_proxy_name = aws_db_proxy.velure_auth_proxy.name

  connection_pool_config {
    max_connections_percent      = 90
    max_idle_connections_percent = 50
    connection_borrow_timeout    = 120
  }
}

# Associar RDS instance ao proxy
resource "aws_db_proxy_target" "velure_auth_proxy_target" {
  db_proxy_name         = aws_db_proxy.velure_auth_proxy.name
  target_group_name     = aws_db_proxy_default_target_group.velure_auth_proxy_tg.name
  db_instance_identifier = aws_db_instance.velure_auth.id
}

# IAM role para o proxy
resource "aws_iam_role" "rds_proxy_role" {
  name = "velure-rds-proxy-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "rds.amazonaws.com"
      }
    }]
  })
}

# Policy para acessar secrets
resource "aws_iam_role_policy" "rds_proxy_secrets" {
  role = aws_iam_role.rds_proxy_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "secretsmanager:GetSecretValue"
      ]
      Resource = [
        aws_secretsmanager_secret.rds_auth_credentials.arn
      ]
    }]
  })
}

# Output do endpoint do proxy
output "rds_proxy_endpoint" {
  description = "RDS Proxy endpoint"
  value       = aws_db_proxy.velure_auth_proxy.endpoint
}
```

#### Depois de Aplicar Terraform

```bash
cd infrastructure/terraform
terraform apply

# Pegar endpoint do proxy
terraform output rds_proxy_endpoint
# Exemplo: velure-auth-proxy.proxy-xxx.us-east-1.rds.amazonaws.com
```

#### Atualizar Connection String no Kubernetes Secret

```bash
# Antes
POSTGRES_HOST=velure-production-auth.cw9gu66melkv.us-east-1.rds.amazonaws.com

# Depois (usar endpoint do proxy)
POSTGRES_HOST=velure-auth-proxy.proxy-xxx.us-east-1.rds.amazonaws.com
```

**Pronto!** Não precisa mudar nada no código Go. ✅

---

## 📌 Opção 2: PgBouncer Centralizado no Kubernetes ⭐ RECOMENDADO

### O Que É?
Um deployment de PgBouncer que fica entre **todos** os seus services e o RDS.

### Vantagens
- ✅ **Totalmente open-source** (sem custo extra)
- ✅ **Controle total** sobre configuração
- ✅ **Portável** (funciona em qualquer K8s, não só AWS)
- ✅ **Lightweight** (20MB RAM por pod)
- ✅ **Fácil de debugar**

### Desvantagens
- ⚠️ **Você gerencia** (upgrades, monitoring, etc.)
- ⚠️ **Single point of failure** (mitigado com 2+ replicas)

### Arquitetura

```
┌─────────────────────────────────────────────┐
│  Kubernetes Cluster                         │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ auth-svc │  │order-svc │  │other-svc │ │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘ │
│        │             │              │      │
│        └─────────────┼──────────────┘      │
│                      ▼                      │
│            ┌──────────────────┐            │
│            │  pgbouncer-svc   │            │
│            │  (LoadBalancer)  │            │
│            └────────┬─────────┘            │
│                     │                       │
│      ┌──────────────┼──────────────┐       │
│      ▼              ▼              ▼       │
│  ┌────────┐    ┌────────┐    ┌────────┐  │
│  │PgBouncer│   │PgBouncer│   │PgBouncer│  │
│  │ Pod 1  │   │ Pod 2  │   │ Pod 3  │  │
│  └───┬────┘    └───┬────┘    └───┬────┘  │
└──────┼─────────────┼─────────────┼────────┘
       │             │             │
       └─────────────┼─────────────┘
                     ▼
         ┌────────────────────────┐
         │   AWS RDS PostgreSQL   │
         │   (20 connections)     │
         └────────────────────────┘
```

### Como Implementar

Veja os arquivos:
- [`deployment.yaml`](./deployment.yaml) - PgBouncer deployment
- [`configmap.yaml`](./configmap.yaml) - Configuração do PgBouncer
- [`service.yaml`](./service.yaml) - Service interno do K8s
- [`secret.yaml.example`](./secret.yaml.example) - Credenciais RDS

### Quick Start

```bash
# 1. Criar namespace (opcional)
kubectl create namespace velure-db

# 2. Criar secret com credenciais RDS
kubectl create secret generic pgbouncer-secret \
  --from-literal=db-host='velure-production-auth.cw9gu66melkv.us-east-1.rds.amazonaws.com' \
  --from-literal=db-user='postgres' \
  --from-literal=db-password='sua-senha-aqui' \
  --namespace velure-db

# 3. Aplicar configurações
kubectl apply -f infrastructure/kubernetes/pgbouncer/

# 4. Verificar status
kubectl get pods -n velure-db
kubectl logs -f deployment/pgbouncer -n velure-db

# 5. Testar conectividade
kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \
  psql -h pgbouncer.velure-db.svc.cluster.local -U postgres -d velure_auth
```

### Atualizar Services para Usar PgBouncer

**auth-service:**
```yaml
# infrastructure/kubernetes/charts/velure-auth-service/values.yaml
env:
  POSTGRES_HOST: "pgbouncer.velure-db.svc.cluster.local"
  POSTGRES_PORT: "5432"
  # Resto permanece igual
```

**publish-order-service:**
```yaml
# infrastructure/kubernetes/charts/velure-publish-order-service/values.yaml
env:
  POSTGRES_HOST: "pgbouncer.velure-db.svc.cluster.local"
  POSTGRES_PORT: "5432"
```

**Deploy:**
```bash
helm upgrade velure-auth-service ./infrastructure/kubernetes/charts/velure-auth-service
helm upgrade velure-publish-order-service ./infrastructure/kubernetes/charts/velure-publish-order-service
```

---

## 📌 Opção 3: PgBouncer Sidecar (Por Pod)

### O Que É?
Cada pod da sua aplicação roda um container PgBouncer ao lado.

### Vantagens
- ✅ **Isolamento total** entre services
- ✅ **Latência ultra-baixa** (localhost)
- ✅ **Sem single point of failure**

### Desvantagens
- ⚠️ **Mais recursos** (1 PgBouncer por pod)
- ⚠️ **Mais complexo** de gerenciar
- ⚠️ **Mais conexões RDS** (N pods × pool_size)

### Como Implementar

Veja exemplo em: [`sidecar-example.yaml`](./sidecar-example.yaml)

---

## 🎯 Qual Opção Escolher?

### Para Começar Rápido (Hoje/Amanhã)
➡️ **Opção 1: AWS RDS Proxy**
- Terraform apply
- Trocar endpoint
- Pronto!

### Para Produção Long-term (Próxima Sprint)
➡️ **Opção 2: PgBouncer Centralizado no K8s**
- Mais controle
- Sem custo extra
- Portável

### Para Casos Especiais
➡️ **Opção 3: Sidecar**
- Múltiplos bancos diferentes
- Isolamento crítico
- Latência < 1ms necessária

---

## 📊 Comparação

| Feature | RDS Proxy | PgBouncer K8s | Sidecar |
|---------|-----------|---------------|---------|
| Setup | ⚡ Rápido | 🔧 Moderado | 🛠️ Complexo |
| Custo | 💰 $40/mês | ✅ Grátis | ✅ Grátis |
| Manutenção | ✅ Zero | 🔧 Baixa | 🔧 Alta |
| Portabilidade | ❌ AWS only | ✅ Qualquer K8s | ✅ Qualquer K8s |
| Controle | ⚠️ Limitado | ✅ Total | ✅ Total |
| Latência | ~2ms | ~0.5ms | ~0.1ms |
| HA | ✅ Auto | ✅ Replicas | ⚠️ Por pod |

---

## 📚 Recursos

- [AWS RDS Proxy Docs](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy.html)
- [PgBouncer Official Docs](https://www.pgbouncer.org/config.html)
- [Kubernetes PgBouncer Examples](https://github.com/kubernetes/examples/tree/master/staging/pgbouncer)

---

## 🚀 Próximos Passos

1. **Escolher opção** (recomendo começar com RDS Proxy)
2. **Implementar** seguindo guia acima
3. **Testar** com load test
4. **Monitorar** métricas de conexão
5. **Ajustar** pool sizes conforme necessário

---

## ❓ FAQ

**P: Posso usar RDS Proxy E PgBouncer?**
R: Sim, mas não faz sentido. São redundantes.

**P: PgBouncer funciona com read replicas?**
R: Sim! Configure múltiplos databases no pgbouncer.ini.

**P: Preciso mudar código Go?**
R: Não! Apenas trocar connection string (POSTGRES_HOST).

**P: Quanto PgBouncer melhora performance?**
R: Reduz conexões em 90%+, melhora latência em ~30-50%.

**P: É seguro?**
R: Sim. Usado por GitHub, Instagram, Discord, etc.
