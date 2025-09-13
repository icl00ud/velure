# Velure — Guia Local de Kubernetes (não versionado)

Este guia resume o estado atual do projeto e como subir tudo em um cluster local de Kubernetes usando os Helm Charts deste repositório.

## Visão geral

- Serviços (Go):
  - auth-service (HTTP 3020) — autenticação, usa PostgreSQL.
  - product-service (HTTP 3010) — catálogo, usa MongoDB e Redis.
  - publish-order-service (HTTP 3030) — publica ordens em RabbitMQ, usa PostgreSQL.
  - process-order-service (HTTP 3040) — consome ordens do RabbitMQ.
- UI: ui-service (Nginx)
- Mensageria: RabbitMQ (AMQP + painel de gestão)
- Observabilidade: health checks em todos os serviços; Prometheus integrável.
- Kubernetes (por chart): Deployment, Service, Probes, resources, PDB, NetworkPolicy (deny-all). HPA na maioria (min 2 réplicas). Ingress + TLS para auth, ui e product.

## Pré-requisitos

- Docker Desktop com Kubernetes (ou kind/k3d)
- kubectl, Helm v3
- mkcert (para TLS local)
- (opcional) k6 para testes de carga
- Entradas no /etc/hosts apontando para 127.0.0.1:
  - auth.velure.local
  - velure-ui.local
  - velure-product-service.local

## Namespaces

Sugestão de organização por domínio:
- database
- order
- authentication
- frontend

Criação:

```bash
kubectl create namespace database || true
kubectl create namespace order || true
kubectl create namespace authentication || true
kubectl create namespace frontend || true
```

## TLS local com mkcert

Gerar e instalar AC local e certificados:

```bash
mkcert -install
mkcert auth.velure.local
mkcert velure-ui.local
mkcert velure-product-service.local
```

Criar Secrets TLS:

```bash
kubectl -n authentication create secret tls velure-auth-tls \
  --key auth.velure.local-key.pem --cert auth.velure.local.pem --dry-run=client -o yaml | kubectl apply -f -

kubectl -n frontend create secret tls velure-ui-tls \
  --key velure-ui.local-key.pem --cert velure-ui.local.pem --dry-run=client -o yaml | kubectl apply -f -

kubectl -n order create secret tls velure-product-tls \
  --key velure-product-service.local-key.pem --cert velure-product-service.local.pem --dry-run=client -o yaml | kubectl apply -f -
```

## Infra: DBs e Mensageria

### PostgreSQL (ns database)

Instale o chart de PostgreSQL (ajuste conforme seu ambiente):

```bash
helm upgrade --install postgres kubernetes/charts/postgresql -n database
```

Crie URLs de conexão (para auth e publish-order):

```bash
# AUTH (ns authentication)
kubectl -n authentication create secret generic velure-auth-postgres \
  --from-literal=url='postgres://user:pass@postgres.database.svc.cluster.local:5432/auth?sslmode=disable' \
  --dry-run=client -o yaml | kubectl apply -f -

# ORDER (ns order) - usado por publish-order
kubectl -n order create secret generic order-database \
  --from-literal=url='postgres://user:pass@postgres.database.svc.cluster.local:5432/orders?sslmode=disable' \
  --dry-run=client -o yaml | kubectl apply -f -
```

### MongoDB (ns database)

```bash
helm upgrade --install mongodb kubernetes/charts/mongodb -n database
```

O chart do product cria um Secret próprio com username/password/database a partir de values. Configure host/port em `kubernetes/charts/velure-product/values.yaml` (ou via `--set`).

### Redis (ns database)

```bash
helm upgrade --install redis kubernetes/charts/redis -n database
```

Se usar senha, crie o Secret da senha e informe o nome em `velure-product.values.yaml` (chave `redis.passwordSecretName`).

### RabbitMQ (ns order)

```bash
helm upgrade --install rabbitmq kubernetes/charts/velure-rabbitmq -n order
```

- O chart cria `rabbitmq-credentials` com usuários admin/publisher/process conforme values.
- Service: `rabbitmq.order.svc.cluster.local` (amqp 5672, http 15672)

## Deploy dos serviços

### Auth (ns authentication)

Secrets necessários:
- `velure-auth-postgres`: data.url (URL do Postgres)
- `velure-auth-jwt`: data.secret, data.expiresIn, data.refreshSecret, data.refreshExpiresIn
- `velure-auth-session`: data.secret, data.expiresIn

Criar JWT e Session (exemplo):

```bash
kubectl -n authentication create secret generic velure-auth-jwt \
  --from-literal=secret='changeme' \
  --from-literal=expiresIn='1h' \
  --from-literal=refreshSecret='changeme2' \
  --from-literal=refreshExpiresIn='7d' \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n authentication create secret generic velure-auth-session \
  --from-literal=secret='changeme' \
  --from-literal=expiresIn='86400000' \
  --dry-run=client -o yaml | kubectl apply -f -
```

Instalação:

```bash
helm upgrade --install velure-auth kubernetes/charts/velure-auth -n authentication
```

- Ingress: https://auth.velure.local (TLS `velure-auth-tls`)
- Porta: 3020
- Health: `/health`

### Product (ns order)

Ajuste `kubernetes/charts/velure-product/values.yaml` para:
- `mongodb.host`, `mongodb.port`
- `mongodb.username`, `mongodb.password`, `mongodb.database` (o chart cria `{{ .Release.Name }}-secret` com estes valores)
- `redis.host`, `redis.port`, e `redis.passwordSecretName` (se houver senha)

Instalação:

```bash
helm upgrade --install velure-product kubernetes/charts/velure-product -n order
```

- Ingress: https://velure-product-service.local (TLS `velure-product-tls`)
- Porta: 3010
- Health: `/health`

### Publish Order (ns order)

Secrets:
- `rabbitmq-conn`: data.url (AMQP). Pode ser criado pelo chart se `secrets.create=true` nos values.
- `order-database`: data.url (Postgres). Pode ser criado pelo chart idem.

Instalação:

```bash
helm upgrade --install velure-publish-order kubernetes/charts/velure-publish-order -n order
```

- Porta: 3030
- Health: `/health`

### Process Order (ns order)

Espera o Secret `rabbitmq-conn` com `data.url` no namespace `order`.

Instalação:

```bash
helm upgrade --install velure-process-order kubernetes/charts/velure-process-order -n order
```

- Porta: 3040
- Health: `/health`

### UI (ns frontend)

Instalação:

```bash
helm upgrade --install velure-ui kubernetes/charts/velure-ui -n frontend
```

- Ingress: https://velure-ui.local (TLS `velure-ui-tls`)
- Service: NodePort 80

## Verificação

```bash
kubectl get pods -A

# Auth
kubectl -n authentication port-forward svc/velure-auth 3020:3020 >/dev/null 2>&1 &
curl -k https://auth.velure.local/health

# Product
curl -k https://velure-product-service.local/health

# UI
curl -k https://velure-ui.local/

# RabbitMQ (painel)
kubectl -n order port-forward svc/rabbitmq 15672:15672
# acesse http://localhost:15672
```

## Políticas de rede

- Todos os charts trazem uma NetworkPolicy `deny-all` por serviço.
- O chart do product inclui uma policy de allow com:
  - Ingress do `ingress-nginx` para porta 3010
  - Egress para DNS (53), MongoDB (27017) e Redis (6379)
- Próximos passos recomendados: adicionar allow-lists equivalentes para auth (Postgres), publish-order (Postgres + RabbitMQ) e process-order (RabbitMQ).

## Troubleshooting

- CrashLoopBackOff: verifique Secrets e envs; veja `kubectl logs` e `kubectl describe pod`.
- Ingress 404/TLS: confira classe do ingress, Secrets TLS no namespace certo e /etc/hosts.
- Conexões bloqueadas: ajuste/adicione NetworkPolicies de allow específicas.
- Imagens: confirme `image.repository` e `image.tag` nos values. Evite `latest`.

## Desinstalação rápida

```bash
helm uninstall velure-ui -n frontend || true
helm uninstall velure-auth -n authentication || true
helm uninstall velure-product -n order || true
helm uninstall velure-publish-order -n order || true
helm uninstall velure-process-order -n order || true
helm uninstall rabbitmq -n order || true
helm uninstall postgres -n database || true
helm uninstall mongodb -n database || true
helm uninstall redis -n database || true
```

---

Observação: este arquivo é um guia local e não precisa ser commitado no repositório. Ajuste nomes de hosts, namespaces e URLs conforme o seu ambiente.