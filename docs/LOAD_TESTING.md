# Load Testing & Horizontal Pod Autoscaling Guide

Este guia explica como executar testes de carga k6 na aplicação Velure e observar o escalonamento horizontal (HPA) em ação.

## 📋 Índice

- [Pré-requisitos](#pré-requisitos)
- [Arquitetura de Escalonamento](#arquitetura-de-escalonamento)
- [Testes Disponíveis](#testes-disponíveis)
- [Quick Start](#quick-start)
- [Kubernetes Local](#kubernetes-local)
- [AWS EKS](#aws-eks)
- [Monitoramento](#monitoramento)
- [Troubleshooting](#troubleshooting)

---

## 🎯 Pré-requisitos

### Ferramentas Necessárias

1. **k6** - Ferramenta de teste de carga
   ```bash
   # macOS
   brew install k6

   # Linux
   sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
     --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
   echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | \
     sudo tee /etc/apt/sources.list.d/k6.list
   sudo apt-get update
   sudo apt-get install k6

   # Windows
   choco install k6
   ```

2. **kubectl** - Cliente Kubernetes
3. **Cluster Kubernetes** rodando (minikube, kind, Docker Desktop, ou EKS)

### Verificar Instalação

```bash
k6 version
kubectl version --client
kubectl cluster-info
```

---

## 🏗️ Arquitetura de Escalonamento

### HorizontalPodAutoscaler (HPA)

Todos os serviços estão configurados com HPA baseado em métricas duplas:

| Serviço | Min Replicas | Max Replicas | CPU Target | Memory Target |
|---------|-------------|--------------|------------|---------------|
| auth-service | 2 | 10 | 80% | 65% |
| product-service | 2 | 10 | 80% | 65% |
| publish-order | 2 | 5 | 80% | 65% |
| process-order | 2 | 5 | 80% | 65% |
| ui-service | 1 | 100 | 80% | - |

### Como Funciona

1. **Warmup (30s)**: Pods começam a receber tráfego gradual
2. **Ramp-up**: Carga aumenta progressivamente, CPU/memória sobem
3. **Trigger**: Quando CPU > 80% ou Memory > 65%, HPA cria novos pods
4. **Stabilização**: Novos pods distribuem a carga
5. **Ramp-down**: Carga diminui, HPA aguarda 5 min e remove pods extras

---

## 🧪 Testes Disponíveis

### 1. auth-service-test.js
- **Carga máxima**: 200 usuários virtuais
- **Endpoints**: register, login, validateToken, getUsers
- **Threshold p95**: < 500ms
- **Taxa de erro**: < 10%

### 2. product-service-test.js
- **Carga máxima**: 400 usuários virtuais
- **Endpoints**: listProducts, search, pagination, createProduct
- **Threshold p95**: < 1000ms
- **Taxa de erro**: < 5%

### 3. publish-order-service-test.js
- **Carga máxima**: 1000 usuários virtuais
- **Endpoints**: createOrder (70%), getOrders (30%)
- **Threshold p95**: < 2000ms
- **Taxa de erro**: < 10%

### 4. ui-service-test.js
- **Carga máxima**: 250 usuários virtuais
- **Endpoints**: homepage, static assets, navigation
- **Threshold p95**: < 3000ms
- **Taxa de erro**: < 15%

### 5. integrated-load-test.js
- **Carga máxima**: 500 usuários virtuais
- **Testa todos os serviços** em proporção realista
- **Threshold p95**: < 2000ms
- **Taxa de erro**: < 10%

---

## 🚀 Quick Start

### Kubernetes Local (Minikube/kind/Docker Desktop)

```bash
cd tests/load

# 1. Instalar metrics-server (necessário para HPA)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Para cluster local, pode precisar de flag insecure:
kubectl patch deployment metrics-server -n kube-system --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'

# 2. Deploy da aplicação (se ainda não fez)
cd ../../
make k8s-deploy

# 3. Verificar HPA está ativo
kubectl get hpa

# 4. Rodar teste integrado
cd tests/load
./run-k8s-local.sh integrated

# 5. Em outro terminal, monitorar escalonamento em tempo real
./monitor-scaling.sh
```

---

## ☸️ Kubernetes Local - Detalhado

### Passo 1: Setup do Cluster

```bash
# Opção A: Minikube
minikube start --cpus=4 --memory=8192
minikube addons enable metrics-server

# Opção B: kind
kind create cluster --config infrastructure/kubernetes/kind-config.yaml

# Opção C: Docker Desktop
# Habilitar Kubernetes nas configurações
```

### Passo 2: Deploy da Aplicação

```bash
# Deploy completo
make k8s-deploy

# Verificar pods
kubectl get pods -A

# Verificar HPA
kubectl get hpa
```

### Passo 3: Rodar Testes

```bash
cd tests/load

# Teste individual
./run-k8s-local.sh auth      # Auth service
./run-k8s-local.sh product   # Product service
./run-k8s-local.sh order     # Order service
./run-k8s-local.sh integrated # Todos os serviços

# Todos os testes em sequência
./run-k8s-local.sh all
```

### Passo 4: Monitorar Escalonamento

**Terminal 1 - Executar teste:**
```bash
./run-k8s-local.sh integrated
```

**Terminal 2 - Monitorar em tempo real:**
```bash
./monitor-scaling.sh
```

**Terminal 3 - Watch HPA:**
```bash
kubectl get hpa -w
```

---

## ☁️ AWS EKS

### Pré-requisitos

```bash
# 1. Cluster EKS rodando
terraform apply  # na pasta infrastructure/terraform

# 2. kubectl configurado
aws eks update-kubeconfig --region us-east-1 --name velure-prod

# 3. Aplicação deployada
make eks-deploy-full
```

### Rodar Testes

```bash
cd tests/load

# Obter URL do Load Balancer
ALB_URL=$(kubectl get ingress velure-ui -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

echo "Load Balancer URL: $ALB_URL"

# Rodar teste manualmente com URL do ALB
k6 run \
  -e AUTH_URL="http://$ALB_URL/api/auth" \
  -e PRODUCT_URL="http://$ALB_URL/api/product" \
  -e ORDER_URL="http://$ALB_URL/api/order" \
  -e UI_URL="http://$ALB_URL" \
  integrated-load-test.js
```

### Monitorar no EKS

```bash
# Terminal 1 - Teste
k6 run -e BASE_URL="http://$ALB_URL" integrated-load-test.js

# Terminal 2 - Monitoramento
./monitor-scaling.sh

# Verificar logs do CloudWatch
aws logs tail /aws/eks/velure-prod/cluster --follow
```

---

## 📊 Monitoramento

### Grafana Dashboard

1. **Acessar Grafana:**
   ```bash
   # Local
   open http://localhost:3000

   # EKS (port-forward)
   kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80
   open http://localhost:3000
   ```

2. **Login**: `admin` / `admin` (ou senha configurada)

3. **Dashboard**: Procure por "Velure - K6 Load Testing & HPA Scaling"

### Métricas Importantes

**HPA Metrics:**
- Current replicas vs Desired replicas
- CPU utilization vs Target (80%)
- Memory utilization vs Target (65%)
- Scaling events timeline

**K6 Metrics:**
- Virtual Users (VUs) over time
- Request rate (requests/sec)
- Response time percentiles (p50, p95, p99)
- Error rate
- HTTP status codes distribution

**Pod Metrics:**
- Pod count per service over time
- CPU/Memory per pod
- Pod creation/termination events

### CLI Monitoring

```bash
# Watch HPA
kubectl get hpa -w

# Watch pods
kubectl get pods -w -l app.kubernetes.io/part-of=velure

# Top pods (CPU/Memory)
kubectl top pods

# Events
kubectl get events --sort-by='.lastTimestamp' | grep -i "scaled\|horizontal"
```

---

## 🔧 Customização dos Testes

### Variáveis de Ambiente

Todos os testes suportam as seguintes variáveis:

```bash
# URLs dos serviços
AUTH_URL=https://velure.local/api/auth
PRODUCT_URL=https://velure.local/api/product
ORDER_URL=https://velure.local/api/order
UI_URL=https://velure.local

# Duração dos testes
WARMUP_DURATION=30s      # Fase de aquecimento
TEST_DURATION=15s        # Duração de cada estágio

# Exemplo de uso
k6 run \
  -e AUTH_URL=https://velure.local/api/auth \
  -e WARMUP_DURATION=60s \
  -e TEST_DURATION=30s \
  auth-service-test.js
```

### Ajustar Intensidade

Edite os arquivos `.js` para modificar:

```javascript
export const options = {
  stages: [
    { duration: '30s', target: 100 },  // Aumente target para mais carga
    { duration: '1m', target: 500 },   // Aumente duration para teste mais longo
    // ...
  ],
};
```

---

## 🐛 Troubleshooting

### HPA Não Está Escalando

**Problema**: Pods não aumentam mesmo com carga alta.

**Soluções:**

1. **Verificar metrics-server:**
   ```bash
   kubectl get deployment metrics-server -n kube-system

   # Se não existir:
   kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
   ```

2. **Verificar se HPA está ativo:**
   ```bash
   kubectl get hpa

   # Se mostrar "<unknown>" em TARGETS:
   kubectl describe hpa velure-auth
   # Procure por erros nas "Conditions"
   ```

3. **Verificar resource requests/limits:**
   ```bash
   kubectl get pod <pod-name> -o yaml | grep -A 5 resources
   ```

   HPA requer que os pods tenham `resources.requests` definidos.

### Pods Não Alcançam CPU/Memory Target

**Problema**: CPU fica abaixo de 80% mesmo com muita carga.

**Causas possíveis:**
- Pods têm resources muito altos
- Serviço é muito eficiente
- Teste não gera carga suficiente

**Soluções:**
- Diminuir `resources.requests.cpu` nos valores do Helm
- Aumentar número de VUs no teste
- Aumentar duração do teste

### Erro "Connection Refused"

**Problema**: k6 não consegue conectar aos serviços.

**Soluções:**

1. **Verificar serviços estão rodando:**
   ```bash
   kubectl get pods
   kubectl get svc
   ```

2. **Testar conectividade manual:**
   ```bash
   kubectl port-forward svc/velure-auth 3020:3020
   curl http://localhost:3020/health
   ```

3. **Verificar Ingress (se usando):**
   ```bash
   kubectl get ingress
   kubectl describe ingress velure-ui
   ```

### Métricas Não Aparecem no Grafana

**Problema**: Dashboard vazio ou sem dados.

**Soluções:**

1. **Verificar Prometheus está coletando:**
   ```bash
   # Port-forward Prometheus
   kubectl port-forward -n monitoring svc/prometheus-k8s 9090:9090

   # Acessar: http://localhost:9090
   # Query: up{job="velure-auth"}
   ```

2. **Verificar ServiceMonitor:**
   ```bash
   kubectl get servicemonitor -n monitoring
   ```

3. **Verificar labels dos pods:**
   ```bash
   kubectl get pods --show-labels
   ```

---

## 📈 Resultados Esperados

### Escalonamento Normal

Com o teste `integrated` em condições normais:

| Fase | Tempo | VUs | Pods Auth | Pods Product | CPU Médio |
|------|-------|-----|-----------|--------------|-----------|
| Inicial | 0s | 0 | 2 | 2 | ~10% |
| Warmup | 30s | 10 | 2 | 2 | ~30% |
| Ramp-up | 2min | 150 | 3-4 | 3-4 | ~70% |
| Peak | 3min | 500 | 5-7 | 6-8 | ~85% |
| Cool-down | 5min | 0 | 2 | 2 | ~10% |

### Gráficos

Você deverá observar no Grafana:
- 📈 CPU subindo gradualmente até ~80-85%
- 🚀 Número de pods aumentando em steps
- ⏱️ Response time estável (não aumenta muito)
- ✅ Error rate < 10%

---

## 🎓 Melhores Práticas

1. **Sempre use warmup** de pelo menos 30 segundos
2. **Monitore em tempo real** com `./monitor-scaling.sh`
3. **Espere 5-10 minutos** após o teste para observar scale-down
4. **Rode testes em horários de baixo uso** (se em produção)
5. **Documente baseline** de performance antes de mudanças
6. **Use thresholds** do k6 para validar SLOs automaticamente

---

## 📚 Referências

- [k6 Documentation](https://k6.io/docs/)
- [Kubernetes HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [Metrics Server](https://github.com/kubernetes-sigs/metrics-server)
- [Prometheus Operator](https://prometheus-operator.dev/)

---

**Desenvolvido com ❤️ pela equipe Velure**
