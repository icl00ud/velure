# Velure - Guia Completo: Kubernetes Local com kind

Este guia detalha como rodar a aplicação Velure completa em um cluster Kubernetes local usando **kind** (Kubernetes in Docker).

## 📋 Índice

- [Por que kind?](#por-que-kind)
- [Pré-requisitos](#pré-requisitos)
- [Quick Start](#quick-start)
- [Instalação Detalhada](#instalação-detalhada)
- [Uso Diário](#uso-diário)
- [Troubleshooting](#troubleshooting)
- [Comparação com outras soluções](#comparação-com-outras-soluções)

---

## Por que kind?

**kind** (Kubernetes in Docker) é a melhor escolha para desenvolvimento local Kubernetes porque:

✅ **Rápido**: Cluster criado em ~20 segundos
✅ **Leve**: Usa containers Docker em vez de VMs (economiza RAM)
✅ **Compatível**: 95% compatível com clusters de produção (EKS/GKE/AKS)
✅ **Ingress nativo**: Acesso direto via localhost sem port-forward
✅ **Fácil**: Automação completa via scripts
✅ **Oficial**: Mantido pelo time do Kubernetes

---

## Pré-requisitos

### 1. Docker Desktop

```bash
# Verificar se Docker está rodando
docker info

# Se não estiver instalado:
# macOS: brew install --cask docker
# Ou baixe: https://www.docker.com/products/docker-desktop
```

### 2. kind

```bash
# macOS
brew install kind

# Linux
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Verificar instalação
kind version
```

### 3. kubectl

```bash
# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Verificar
kubectl version --client
```

### 4. helm

```bash
# macOS
brew install helm

# Linux
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Verificar
helm version
```

---

## Quick Start

### Opção 1: Usando START_HERE.sh (Recomendado)

```bash
# 1. Executar script
./START_HERE.sh

# 2. Escolher opção 10
# "🚀 Criar cluster Kubernetes local (kind) + Deploy completo"

# 3. Aguardar ~5-7 minutos (build + deploy)

# 4. Acessar aplicação
open http://velure.local
```

### Opção 2: Usando Makefile

```bash
# Deploy completo (cria cluster + build + deploy)
make kind-deploy

# Acessar
open http://velure.local
```

### Opção 3: Usando script diretamente

```bash
# Deploy completo
./scripts/k8s/setup-kind-cluster.sh
```

---

## Instalação Detalhada

### 1. Criar Cluster kind

```bash
# Usando configuração customizada
kind create cluster --config=infrastructure/kubernetes/kind-config.yaml

# O que isso faz:
# - Cria cluster chamado "velure"
# - Expõe portas 80 e 443 para ingress
# - Configura node labels para ingress controller
```

### 2. Instalar NGINX Ingress Controller

```bash
# Adicionar repositório
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Instalar
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=NodePort \
  --set controller.hostPort.enabled=true \
  --set controller.hostPort.ports.http=80 \
  --set controller.hostPort.ports.https=443 \
  --wait
```

### 3. Build e Carregar Imagens

```bash
# Build todas as imagens
make docker-build

# Carregar no kind
kind load docker-image velure-auth-service:latest --name velure
kind load docker-image velure-product-service:latest --name velure
kind load docker-image velure-publish-order-service:latest --name velure
kind load docker-image velure-process-order-service:latest --name velure
kind load docker-image velure-ui-service:latest --name velure

# Ou usar script:
./scripts/k8s/load-images-to-kind.sh
```

### 4. Deploy Datastores

```bash
# Criar namespaces
kubectl create namespace datastores
kubectl create namespace velure

# Adicionar repo Bitnami
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Deploy datastores (MongoDB, Redis, RabbitMQ)
helm upgrade --install velure-datastores \
  infrastructure/kubernetes/charts/velure-datastores \
  -n datastores \
  --wait
```

### 5. Deploy Microserviços

```bash
# Auth Service
helm upgrade --install velure-auth \
  infrastructure/kubernetes/charts/velure-auth \
  -n velure \
  --set image.pullPolicy=Never \
  --set ingress.className=nginx

# Product Service
helm upgrade --install velure-product \
  infrastructure/kubernetes/charts/velure-product \
  -n velure \
  --set image.pullPolicy=Never \
  --set ingress.className=nginx

# Order Services
helm upgrade --install velure-publish-order \
  infrastructure/kubernetes/charts/velure-publish-order \
  -n velure \
  --set image.pullPolicy=Never

helm upgrade --install velure-process-order \
  infrastructure/kubernetes/charts/velure-process-order \
  -n velure \
  --set image.pullPolicy=Never

# UI Service
helm upgrade --install velure-ui \
  infrastructure/kubernetes/charts/velure-ui \
  -n velure \
  --set image.pullPolicy=Never \
  --set ingress.className=nginx
```

### 6. Configurar /etc/hosts

```bash
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts
echo "127.0.0.1 auth.velure.local" | sudo tee -a /etc/hosts
echo "127.0.0.1 product.velure.local" | sudo tee -a /etc/hosts
```

### 7. Verificar Deploy

```bash
# Ver pods
kubectl get pods -n velure
kubectl get pods -n datastores

# Ver ingress
kubectl get ingress -n velure

# Ver logs
kubectl logs -f deployment/velure-ui -n velure
```

---

## Uso Diário

### Ver Status

```bash
# Usando Makefile
make kind-status

# Ou diretamente
kubectl get pods -n velure
kubectl get pods -n datastores
kubectl get ingress -n velure
kubectl get hpa -n velure  # HorizontalPodAutoscaler
```

### Ver Logs

```bash
# Logs de um serviço específico
kubectl logs -f deployment/velure-auth -n velure

# Logs de todos serviços
make kind-logs

# Logs de um pod específico
kubectl logs -f <pod-name> -n velure
```

### Acessar Aplicação

```bash
# Frontend
open http://velure.local

# APIs
curl http://auth.velure.local/health
curl http://product.velure.local/health
curl http://velure.local/api/order/health
```

### Executar Comandos em Pods

```bash
# Shell em um pod
kubectl exec -it <pod-name> -n velure -- /bin/sh

# Exemplo: auth-service
kubectl exec -it deployment/velure-auth -n velure -- /bin/sh
```

### Port-Forward (se ingress não funcionar)

```bash
# UI
kubectl port-forward -n velure svc/velure-ui 8080:80

# Grafana (se instalado)
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

# RabbitMQ Management
kubectl port-forward -n datastores svc/velure-datastores-rabbitmq 15672:15672
```

### Deletar e Recriar

```bash
# Deletar tudo
make kind-delete

# Recriar do zero
make kind-deploy
```

### Atualizar Código

Quando você modifica código de um serviço:

```bash
# 1. Rebuild a imagem
cd services/auth-service
docker build -t velure-auth-service:latest .

# 2. Carregar no kind
kind load docker-image velure-auth-service:latest --name velure

# 3. Reiniciar deployment
kubectl rollout restart deployment/velure-auth -n velure

# 4. Acompanhar rollout
kubectl rollout status deployment/velure-auth -n velure
```

---

## Troubleshooting

### Problema: Cluster não cria

```bash
# Verificar Docker
docker info

# Deletar cluster existente
kind delete cluster --name velure

# Recriar
kind create cluster --config=infrastructure/kubernetes/kind-config.yaml
```

### Problema: Pods ficam em "Pending"

```bash
# Ver eventos
kubectl describe pod <pod-name> -n velure

# Ver recursos do node
kubectl top nodes

# Verificar se imagens foram carregadas
docker exec -it velure-control-plane crictl images | grep velure
```

### Problema: Ingress não funciona (404)

```bash
# Verificar ingress controller
kubectl get pods -n ingress-nginx

# Ver logs do ingress
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller

# Verificar ingress resources
kubectl get ingress -n velure -o yaml

# Testar diretamente o service
kubectl port-forward -n velure svc/velure-ui 8080:80
curl http://localhost:8080
```

### Problema: Imagens não encontradas (ImagePullBackOff)

```bash
# Verificar se imagem foi carregada
docker exec -it velure-control-plane crictl images

# Carregar novamente
kind load docker-image velure-ui-service:latest --name velure

# Verificar pullPolicy
kubectl get deployment velure-ui -n velure -o yaml | grep imagePullPolicy
# Deve ser "Never" para imagens locais
```

### Problema: /etc/hosts não funciona

```bash
# Verificar se entrada existe
grep velure.local /etc/hosts

# Adicionar manualmente
sudo nano /etc/hosts
# Adicionar: 127.0.0.1 velure.local

# No macOS, flush DNS cache
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

### Problema: Datastores não iniciam

```bash
# Ver logs
kubectl logs -n datastores -l app.kubernetes.io/instance=velure-datastores

# Ver PVCs (Persistent Volume Claims)
kubectl get pvc -n datastores

# Deletar e recriar
helm uninstall velure-datastores -n datastores
kubectl delete pvc --all -n datastores
helm install velure-datastores infrastructure/kubernetes/charts/velure-datastores -n datastores
```

### Problema: Conflito de porta 80

```bash
# Ver o que está usando porta 80
sudo lsof -i :80

# Parar Docker Compose se estiver rodando
cd infrastructure/local
docker-compose down

# Recriar cluster kind
kind delete cluster --name velure
kind create cluster --config=infrastructure/kubernetes/kind-config.yaml
```

---

## Comparação com outras soluções

| Feature | kind | minikube | k3d | Docker Desktop K8s |
|---------|------|----------|-----|--------------------|
| **Startup time** | ~20s | ~60s | ~5s | ~30s |
| **RAM usage** | 450MB | 650MB | 420MB | 1GB+ |
| **Ingress** | Native | Tunnel | Native | Port-forward |
| **Multi-node** | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No |
| **Production-like** | ✅ 95% | ✅ 90% | ⚠️ 80% | ⚠️ 70% |
| **Setup complexity** | Low | Medium | Low | Very Low |
| **Documentation** | Excellent | Excellent | Good | Good |
| **Velure compatibility** | ✅ 95% | ✅ 90% | ⚠️ 85% | ⚠️ 80% |

**Recomendação**: Use **kind** para Velure. É o melhor equilíbrio entre simplicidade, performance e compatibilidade.

---

## Recursos Adicionais

### Documentação Oficial

- kind: https://kind.sigs.k8s.io/
- kubectl: https://kubernetes.io/docs/reference/kubectl/
- helm: https://helm.sh/docs/

### Arquivos de Configuração

- `infrastructure/kubernetes/kind-config.yaml` - Configuração do cluster
- `scripts/k8s/setup-kind-cluster.sh` - Script de setup completo
- `scripts/k8s/load-images-to-kind.sh` - Script para carregar imagens

### Comandos Makefile

```bash
make kind-create       # Criar cluster kind
make kind-setup        # Criar cluster + ingress
make kind-build-load   # Build + carregar imagens
make kind-deploy       # Deploy completo
make kind-delete       # Deletar cluster
make kind-status       # Ver status
make kind-logs         # Ver logs
```

### Dashboards e Monitoramento

Para instalar Grafana + Prometheus no kind:

```bash
# Instalar kube-prometheus-stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace

# Port-forward Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
# Acesse: http://localhost:3000 (admin/prom-operator)
```

---

**Próximos passos**:
- ✅ Cluster kind criado e rodando
- 📊 Ver [LOAD_TESTING.md](../LOAD_TESTING.md) para testes de carga
- 📈 Ver [MONITORING.md](../MONITORING.md) para configurar monitoramento
- ☁️ Ver [DEPLOY_GUIDE.md](../DEPLOY_GUIDE.md) para deploy em produção (AWS EKS)
