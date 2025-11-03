# Guia de Troubleshooting - Velure

Este guia ajuda a diagnosticar e resolver problemas comuns no deploy do Velure.

## 🔍 Diagnóstico Rápido

### Verificar Status Geral

```bash
# Ver todos os pods
kubectl get pods -A

# Ver pods com problemas
kubectl get pods -A | grep -v Running

# Ver eventos recentes
kubectl get events --sort-by='.lastTimestamp' | tail -20

# Ver logs de um pod
kubectl logs <pod-name> -n <namespace>

# Ver logs anteriores (se pod crashou)
kubectl logs <pod-name> --previous
```

## 🚨 Problemas Comuns

### 1. Pod em CrashLoopBackOff

**Sintoma**: Pod reinicia continuamente

```bash
# Ver logs
kubectl logs <pod-name>

# Descrever pod (ver eventos)
kubectl describe pod <pod-name>
```

**Causas Comuns:**

#### A. Falha de Conexão com Banco de Dados

```bash
# Verificar se RDS está acessível
kubectl run test --rm -it --image=postgres:15 -- \
  psql -h <rds-endpoint> -U postgres

# Verificar secret
kubectl get secret database-secrets -o yaml

# Editar secret com valores corretos
kubectl edit secret database-secrets
```

#### B. Variáveis de Ambiente Faltando

```bash
# Ver env vars do pod
kubectl describe pod <pod-name> | grep -A 20 "Environment:"

# Verificar secrets e configmaps
kubectl get secret,configmap
```

#### C. Porta Errada ou Conflito

```bash
# Ver porta configurada
kubectl describe pod <pod-name> | grep Port

# Ver se serviço aponta para porta correta
kubectl describe svc <service-name>
```

**Solução**:

```bash
# Corrigir e fazer redeploy
helm upgrade <release-name> <chart-path> -n <namespace>

# Ou restart do deployment
kubectl rollout restart deployment/<deployment-name>
```

### 2. ImagePullBackOff

**Sintoma**: Não consegue baixar imagem Docker

```bash
# Ver detalhes
kubectl describe pod <pod-name>
```

**Causas:**
- Imagem não existe
- Registry privado sem credenciais
- Typo no nome da imagem

**Solução**:

```bash
# Verificar nome da imagem
kubectl get deployment <name> -o yaml | grep image:

# Se for registry privado, criar secret
kubectl create secret docker-registry regcred \
  --docker-server=<registry> \
  --docker-username=<username> \
  --docker-password=<password>

# Atualizar deployment para usar o secret
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "regcred"}]}'
```

### 3. Pending Pods

**Sintoma**: Pod fica em estado Pending

```bash
kubectl describe pod <pod-name>
```

**Causas Comuns:**

#### A. Recursos Insuficientes

```
Events:
  Warning  FailedScheduling  pod didn't fit node: Insufficient cpu
```

**Solução**:

```bash
# Ver recursos dos nodes
kubectl top nodes

# Ver recursos dos pods
kubectl top pods

# Reduzir requests do pod
kubectl edit deployment <name>
# Ajustar resources.requests.cpu e memory
```

#### B. PVC Não Pode Ser Montado

```
Warning  FailedMount  Unable to mount volumes
```

**Solução**:

```bash
# Ver PVCs
kubectl get pvc

# Ver detalhes
kubectl describe pvc <pvc-name>

# Se storageClass não existe
kubectl get storageclass

# Criar se necessário (para EBS)
kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
EOF
```

### 4. Service Unreachable

**Sintoma**: Não consegue acessar serviço

#### A. Verificar Service e Endpoints

```bash
# Ver services
kubectl get svc

# Ver endpoints (pods por trás do service)
kubectl get endpoints <service-name>

# Se endpoints está vazio, selector está errado
kubectl describe svc <service-name>
kubectl get pods --show-labels
```

#### B. Testar Conectividade Interna

```bash
# Criar pod de teste
kubectl run test --rm -it --image=curlimages/curl -- sh

# Dentro do pod:
curl http://<service-name>:<port>/health
```

### 5. Ingress/LoadBalancer Não Funciona

**Sintoma**: Não consegue acessar aplicação externamente

#### A. LoadBalancer Sem External-IP

```bash
# Ver ingress
kubectl get ingress

# Ver events
kubectl describe ingress <ingress-name>
```

**Causas:**
- ALB Controller não instalado
- IAM role incorreta
- Subnets sem tags corretas

**Solução**:

```bash
# Verificar ALB controller
kubectl get deployment -n kube-system aws-load-balancer-controller

# Ver logs
kubectl logs -n kube-system -l app.kubernetes.io/name=aws-load-balancer-controller

# Reinstalar se necessário
cd scripts/deploy
./01-install-controllers.sh
```

#### B. Verificar Tags das Subnets

Subnets públicas devem ter:
```
kubernetes.io/role/elb = 1
```

Subnets privadas devem ter:
```
kubernetes.io/role/internal-elb = 1
```

```bash
# Ver via AWS CLI
aws ec2 describe-subnets --filters "Name=vpc-id,Values=<vpc-id>"
```

### 6. Prometheus Não Scrape Métricas

**Sintoma**: Métricas não aparecem no Grafana

#### A. Verificar Targets no Prometheus

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
```

Acesse: http://localhost:9090/targets

**Status deve ser UP** para todos os targets velure-*

#### B. Se Target está DOWN

```bash
# Verificar ServiceMonitor
kubectl get servicemonitor -A

# Verificar se service tem labels corretas
kubectl get svc <service-name> --show-labels

# ServiceMonitor deve ter selector que match com service labels
kubectl describe servicemonitor <name>
```

#### C. Testar Endpoint de Métricas

```bash
# Port-forward para o pod
kubectl port-forward <pod-name> 8080:8080

# Testar
curl http://localhost:8080/metrics

# Deve retornar métricas Prometheus
```

**Se não retorna métricas:**
- Verificar se `/metrics` endpoint existe no código
- Verificar se porta está correta

#### D. Verificar Prometheus Logs

```bash
kubectl logs -n monitoring prometheus-kube-prometheus-stack-prometheus-0
```

### 7. RabbitMQ Connection Refused

**Sintoma**: process-order-service não consegue conectar ao RabbitMQ

```bash
# Ver logs
kubectl logs -l app.kubernetes.io/name=velure-process-order

# Erro: "connection refused" ou "dial tcp: connect refused"
```

**Solução**:

```bash
# Verificar se RabbitMQ está rodando
kubectl get pods -n datastores -l app.kubernetes.io/name=rabbitmq

# Testar conectividade
kubectl run test --rm -it --image=curlimages/curl -- sh
curl http://velure-datastores-rabbitmq.datastores:15672

# Verificar usuário e senha
kubectl exec -n datastores velure-datastores-rabbitmq-0 -- \
  rabbitmqctl list_users

# Verificar queues
kubectl exec -n datastores velure-datastores-rabbitmq-0 -- \
  rabbitmqctl list_queues
```

### 8. MongoDB/Redis Connection Issues

#### MongoDB

```bash
# Verificar se está rodando
kubectl get pods -n datastores -l app.kubernetes.io/name=mongodb

# Testar conexão
kubectl exec -n datastores velure-datastores-mongodb-0 -- \
  mongosh --eval "db.adminCommand('ping')"

# Ver logs
kubectl logs -n datastores -l app.kubernetes.io/name=mongodb
```

#### Redis

```bash
# Verificar se está rodando
kubectl get pods -n datastores -l app.kubernetes.io/name=redis

# Testar conexão
kubectl exec -n datastores velure-datastores-redis-master-0 -- \
  redis-cli -a redis_password ping

# Deve retornar: PONG
```

### 9. High Memory/CPU Usage

**Sintoma**: Pods sendo killed (OOMKilled) ou throttled

```bash
# Ver recursos
kubectl top pods
kubectl top nodes

# Ver limites configurados
kubectl describe pod <pod-name> | grep -A 5 "Limits:"

# Ver se pod foi killed por OOM
kubectl describe pod <pod-name> | grep -i oom
```

**Solução**:

```bash
# Aumentar limits
kubectl edit deployment <name>

# Exemplo:
resources:
  limits:
    memory: "1Gi"
    cpu: "1000m"
  requests:
    memory: "512Mi"
    cpu: "250m"
```

### 10. Disk Pressure / PVC Full

**Sintoma**: Pods evicted, disk pressure

```bash
# Ver PVCs
kubectl get pvc -A

# Ver uso de disco nos pods
kubectl exec <pod-name> -- df -h

# Ver eventos de disk pressure
kubectl get events | grep -i "disk"
```

**Solução**:

```bash
# Expandir PVC (requer storageClass com allowVolumeExpansion: true)
kubectl edit pvc <pvc-name>
# Aumentar spec.resources.requests.storage

# Limpar dados antigos (Prometheus)
kubectl exec -n monitoring prometheus-kube-prometheus-stack-prometheus-0 -- \
  promtool tsdb clean-tombstones /prometheus
```

## 🛠️ Ferramentas Úteis

### Listar Todos os Recursos

```bash
kubectl get all -A
```

### Ver Consumo de Recursos

```bash
# Por node
kubectl top nodes

# Por pod
kubectl top pods -A

# Por namespace
kubectl top pods -n <namespace>
```

### Debug de Rede

```bash
# Criar pod de debug
kubectl run debug --rm -it --image=nicolaka/netshoot -- bash

# Dentro do pod:
# Testar DNS
nslookup velure-datastores-mongodb.datastores.svc.cluster.local

# Testar conectividade
curl http://velure-auth:3020/health
telnet velure-datastores-mongodb 27017
```

### Ver Configuração Aplicada

```bash
# Ver YAML completo do recurso
kubectl get deployment <name> -o yaml

# Ver apenas spec
kubectl get deployment <name> -o jsonpath='{.spec}'

# Ver env vars
kubectl get deployment <name> -o jsonpath='{.spec.template.spec.containers[0].env}'
```

## 📊 Logs Centralizados

### Ver Logs de Múltiplos Pods

```bash
# Todos os pods de um deployment
kubectl logs -l app.kubernetes.io/name=velure-auth --tail=100

# Follow logs
kubectl logs -f -l app.kubernetes.io/name=velure-auth

# Logs de todos os containers em um pod
kubectl logs <pod-name> --all-containers=true
```

### Filtrar Logs

```bash
# Grep em logs
kubectl logs <pod-name> | grep ERROR

# Últimas 100 linhas
kubectl logs <pod-name> --tail=100

# Logs desde timestamp
kubectl logs <pod-name> --since-time=2024-01-01T00:00:00Z
```

## 🔄 Recovery Procedures

### Restart de Serviço

```bash
# Restart graceful
kubectl rollout restart deployment/<name>

# Deletar pod (será recriado)
kubectl delete pod <pod-name>

# Scale down e up
kubectl scale deployment/<name> --replicas=0
kubectl scale deployment/<name> --replicas=2
```

### Rollback de Deploy

```bash
# Ver histórico
kubectl rollout history deployment/<name>

# Rollback para versão anterior
kubectl rollout undo deployment/<name>

# Rollback para revisão específica
kubectl rollout undo deployment/<name> --to-revision=2
```

### Limpar Recursos Travados

```bash
# Forçar deleção de pod
kubectl delete pod <pod-name> --grace-period=0 --force

# Remover finalizers de recurso travado
kubectl patch <resource> <name> -p '{"metadata":{"finalizers":[]}}' --type=merge
```

## 📞 Suporte

### Coletar Informações para Debug

```bash
# Criar diretório de debug
mkdir velure-debug
cd velure-debug

# Coletar informações
kubectl get all -A > all-resources.txt
kubectl get events --sort-by='.lastTimestamp' -A > events.txt
kubectl describe nodes > nodes.txt
kubectl top nodes > node-resources.txt
kubectl top pods -A > pod-resources.txt

# Logs de serviços problemáticos
kubectl logs -l app.kubernetes.io/name=velure-auth > auth-logs.txt
kubectl logs -n monitoring -l app.kubernetes.io/name=prometheus > prometheus-logs.txt

# Compactar
tar czf velure-debug.tar.gz *
```

### Checklist de Debug

- [ ] Pods estão Running?
- [ ] Services têm endpoints?
- [ ] Ingress tem External-IP?
- [ ] Secrets estão configurados?
- [ ] Prometheus está scrapando métricas?
- [ ] Logs mostram erros?
- [ ] Recursos (CPU/Memory) suficientes?
- [ ] Conectividade de rede OK?
- [ ] PVCs montados corretamente?

## 📚 Referências

- [Kubernetes Troubleshooting](https://kubernetes.io/docs/tasks/debug/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Deploy Guide](./DEPLOY_GUIDE.md)
- [Monitoring Guide](./MONITORING.md)
