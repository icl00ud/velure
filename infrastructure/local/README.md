# Local Development Environment

Este diretório contém a configuração para executar toda a plataforma Velure localmente usando Docker Compose.

## 📋 Pré-requisitos

- Docker Desktop (ou Docker Engine + Docker Compose)
- 8GB+ de RAM disponível
- 20GB+ de espaço em disco

## 🚀 Como Usar

### 1. Configurar Variáveis de Ambiente

Copie o arquivo de exemplo e ajuste conforme necessário:

```bash
cp .env.example .env
```

**Importante:** O arquivo `.env` já contém valores funcionais para desenvolvimento local. Você só precisa alterá-los se quiser customizar portas, credenciais ou outros parâmetros.

### 2. Iniciar os Serviços

```bash
# Iniciar todos os serviços
docker compose up -d

# Ou com rebuild (útil após mudanças no código)
docker compose up -d --build

# Ou com rebuild forçado (limpa cache)
docker compose up -d --build --force-recreate
```

### 3. Verificar Status

```bash
# Ver logs de todos os serviços
docker compose logs -f

# Ver logs de um serviço específico
docker compose logs -f auth-service

# Ver status dos containers
docker compose ps
```

### 4. Acessar os Serviços

Com o Caddy reverse proxy, todos os serviços estão disponíveis através de um único ponto de entrada:

- **UI (Frontend)**: https://velure.local
- **Auth API**: https://auth.velure.local/api/auth
- **Product API**: https://product.velure.local/api/product
- **Order API**: https://order.velure.local/api/order
- **RabbitMQ Management**: http://localhost:15672 (admin/admin_password)

### 5. Parar os Serviços

```bash
# Parar e remover containers
docker compose down

# Parar, remover containers E volumes (limpa dados)
docker compose down -v
```

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────┐
│              Caddy Reverse Proxy                │
│         (TLS Termination & Routing)             │
└────────┬─────────────┬─────────────┬────────────┘
         │             │             │
    ┌────▼────┐   ┌────▼────┐   ┌───▼──────┐
    │   UI    │   │  Auth   │   │ Product  │
    │ Service │   │ Service │   │ Service  │
    └─────────┘   └────┬────┘   └────┬─────┘
                       │             │
                  ┌────▼─────┐  ┌───▼──────┐
                  │PostgreSQL│  │ MongoDB  │
                  └──────────┘  │  +Redis  │
                                └──────────┘
    ┌──────────────────────────────────────┐
    │         RabbitMQ Message Queue       │
    └────┬─────────────────────────┬───────┘
         │                         │
    ┌────▼────────┐      ┌─────────▼─────┐
    │   Publish   │      │    Process    │
    │Order Service│      │ Order Service │
    └─────────────┘      └───────────────┘
```

## 📦 Serviços Incluídos

### Aplicação
- **ui-service**: Frontend React (Vite + TypeScript)
- **auth-service**: Autenticação e autorização (Go)
- **product-service**: Catálogo de produtos (Go)
- **publish-order-service**: Publicação de pedidos (Go)
- **process-order-service**: Processamento de pedidos (Go)

### Infraestrutura
- **caddy**: Reverse proxy com TLS automático
- **postgres**: Banco de dados relacional (Auth + Orders)
- **mongodb**: Banco NoSQL (Produtos)
- **redis**: Cache em memória
- **rabbitmq**: Message broker para processamento assíncrono

## 🔧 Configuração de Recursos

Os recursos estão otimizados para desenvolvimento local:

| Serviço | CPU Limit | Memory Limit |
|---------|-----------|--------------|
| caddy | 0.5 cores | 256 MB |
| postgres | 1.0 cores | 512 MB |
| mongodb | 0.5 cores | 256 MB |
| rabbitmq | 0.5 cores | 512 MB |
| auth-service | 0.25 cores | 128 MB |
| product-service | 0.25 cores | 128 MB |
| publish-order | 0.5 cores | 128 MB |
| process-order | 0.25 cores | 128 MB |
| ui-service | 0.25 cores | 128 MB |

**Total estimado:** ~2-3 GB RAM, ~2-3 CPU cores

## 🔍 Troubleshooting

### Porta já em uso
```bash
# Verificar o que está usando a porta
lsof -i :80
lsof -i :443

# Parar o processo ou alterar a porta no .env
```

### Container não inicia
```bash
# Ver logs detalhados
docker compose logs <service-name>

# Rebuild do container específico
docker compose up -d --build --force-recreate <service-name>
```

### Erro de permissão em volumes
```bash
# Remover volumes e recriar
docker compose down -v
docker compose up -d
```

### Problemas de DNS (*.velure.local)
```bash
# Adicionar ao /etc/hosts (macOS/Linux) ou C:\Windows\System32\drivers\etc\hosts (Windows)
127.0.0.1 velure.local
127.0.0.1 auth.velure.local
127.0.0.1 product.velure.local
127.0.0.1 order.velure.local
```

## 📝 Variáveis de Ambiente Importantes

### Segurança
- `JWT_SECRET`: Chave secreta para tokens JWT
- `JWT_REFRESH_TOKEN_SECRET`: Chave para refresh tokens
- `SESSION_SECRET`: Chave para sessões

**⚠️ IMPORTANTE:** Altere estes valores em produção!

### Banco de Dados
- `POSTGRES_*`: Configurações do PostgreSQL
- `MONGODB_*`: Configurações do MongoDB
- `REDIS_*`: Configurações do Redis

### RabbitMQ
- `RABBITMQ_*`: Configurações do message broker
- Usuários separados por serviço para melhor isolamento

## 🔄 Atualização dos Serviços

Após fazer mudanças no código:

```bash
# Rebuild apenas o serviço alterado
docker compose up -d --build auth-service

# Ou rebuild de todos
docker compose up -d --build
```

## 🧹 Limpeza

```bash
# Parar e remover tudo
docker compose down -v

# Remover imagens não usadas
docker image prune -a

# Limpeza completa do Docker
docker system prune -a --volumes
```

## 📚 Documentação Adicional

- [Documentação Principal](../../README.md)
- [Guia de Arquitetura](../../docs/architecture/ARCHITECTURE.md)
- [Deployment no Kubernetes](../../docs/deployment/kubernetes-local-guide.md)
