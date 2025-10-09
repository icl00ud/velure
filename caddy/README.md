# Caddy Reverse Proxy - Velure

## 🎯 Visão Geral

Caddy é o proxy reverso da aplicação Velure, fornecendo:
- ✅ **TLS automático** com Let's Encrypt (produção)
- ✅ **Roteamento centralizado** para todos os microserviços
- ✅ **CORS** configurado
- ✅ **Security headers** (HSTS, CSP, X-Frame-Options, etc)
- ✅ **Health checks** automáticos
- ✅ **Compressão** (gzip, zstd)
- ✅ **Logs estruturados** em JSON
- ✅ **Error handling** com mensagens JSON

---

## 📁 Estrutura

```
caddy/
├── Caddyfile          # Configuração do Caddy
└── README.md          # Este arquivo
```

---

## 🚀 Como Usar

### Desenvolvimento Local

1. **Adicionar domínio ao /etc/hosts:**
   ```bash
   sudo echo "127.0.0.1 velure.local" >> /etc/hosts
   ```

2. **Subir stack completa:**
   ```bash
   docker compose up -d
   ```

3. **Acessar aplicação:**
   - Frontend: https://velure.local
   - API Auth: https://velure.local/api/auth
   - API Products: https://velure.local/api/product
   - API Orders: https://velure.local/api/order
   - SSE: https://velure.local/api/sse
   - Health Check: https://velure.local/health

4. **Aceitar certificado auto-assinado:**
   - Chrome: Clique em "Advanced" → "Proceed to velure.local (unsafe)"
   - Firefox: "Advanced" → "Accept the Risk and Continue"

### Verificar Status

```bash
# Logs do Caddy
docker logs caddy-proxy -f

# Verificar configuração
docker exec caddy-proxy caddy validate --config /etc/caddy/Caddyfile

# Recarregar configuração sem downtime
docker exec caddy-proxy caddy reload --config /etc/caddy/Caddyfile
```

---

## 🔧 Configuração

### Caddyfile - Estrutura

```caddyfile
# Configuração Global
{
    email admin@velure.local
    local_certs              # Apenas desenvolvimento
    log { ... }
}

# Domínio (velure.local ou velure.com.br)
velure.local {
    # Security Headers
    header { ... }
    
    # CORS
    @options { method OPTIONS }
    handle @options { ... }
    
    # Rotas
    handle_path /api/auth* { ... }
    handle_path /api/product* { ... }
    handle_path /api/order* { ... }
    handle_path /api/sse* { ... }    # SSE streaming
    handle /* { ... }                 # Frontend SPA
    
    # Compressão e Error Handling
    encode gzip zstd
    handle_errors { ... }
}
```

### Rotas Configuradas

| Rota | Backend | Porta | Observações |
|------|---------|-------|-------------|
| `/` | ui-service | 80 | SPA React |
| `/api/auth/*` | auth-service | 3001 | Autenticação/Sessão |
| `/api/product/*` | product-service | 3010 | Catálogo de produtos |
| `/api/order/*` | publish-order-service | 3002 | Orders REST API |
| `/api/sse/*` | publish-order-service | 3002 | Server-Sent Events |
| `/health` | caddy | - | Health check agregado |

### Security Headers

- **HSTS**: `max-age=31536000; includeSubDomains; preload`
- **X-Content-Type-Options**: `nosniff`
- **X-Frame-Options**: `SAMEORIGIN`
- **X-XSS-Protection**: `1; mode=block`
- **Referrer-Policy**: `strict-origin-when-cross-origin`
- **Permissions-Policy**: Bloqueia geolocation, microphone, camera
- **CSP**: Permite `'self'`, `'unsafe-inline'`, `'unsafe-eval'`, HTTPS, data:, blob:

---

## 🌐 Produção

### 1. Configurar Domínio Real

No `Caddyfile`, descomentar seção de produção e remover `local_certs`:

```caddyfile
{
    email admin@velure.com.br  # Email para Let's Encrypt
    # Remover: local_certs
}

# Frontend
velure.com.br, www.velure.com.br {
    # Redirect www → non-www
    @www host www.velure.com.br
    redir @www https://velure.com.br{uri} permanent
    
    # Mesma configuração de rotas
    handle_path /api/auth* { ... }
    # ... etc
}

# API subdomain (opcional)
api.velure.com.br {
    # Rotas sem prefixo /api
    handle_path /auth* { ... }
    handle_path /product* { ... }
}
```

### 2. DNS Records (Route53)

```
A    velure.com.br     → <IP do Load Balancer>
A    www.velure.com.br → <IP do Load Balancer>
A    api.velure.com.br → <IP do Load Balancer>
```

### 3. TLS Automático

Caddy automaticamente:
1. Solicita certificados Let's Encrypt
2. Renova antes de expirar (< 30 dias)
3. Redireciona HTTP → HTTPS
4. Habilita HTTP/2 e HTTP/3

**Nenhuma configuração manual necessária!**

---

## 📊 Monitoring & Logs

### Logs

Logs em JSON para fácil parsing:

```bash
# Access logs
docker exec caddy-proxy tail -f /var/log/caddy/access.log | jq

# Velure domain logs
docker exec caddy-proxy tail -f /var/log/caddy/velure.log | jq

# Filtrar por status code
docker exec caddy-proxy cat /var/log/caddy/velure.log | jq 'select(.status >= 500)'

# Filtrar por endpoint
docker exec caddy-proxy cat /var/log/caddy/velure.log | jq 'select(.request.uri | contains("/api/order"))'
```

### Métricas (Prometheus)

Caddy expõe métricas Prometheus em `/metrics` (necessário habilitar):

```caddyfile
:2019 {
    metrics /metrics
}
```

Integrar com Prometheus:

```yaml
scrape_configs:
  - job_name: 'caddy'
    static_configs:
      - targets: ['caddy-proxy:2019']
```

---

## 🔍 Health Checks

### Health Check Agregado

```bash
curl https://velure.local/health
```

Resposta:
```json
{
  "status": "healthy",
  "timestamp": "1728432000",
  "proxy": "caddy v2.8",
  "services": {
    "auth": "/api/auth/health",
    "product": "/api/product/health",
    "orders": "/api/order/health"
  }
}
```

### Health Checks Individuais

Caddy verifica health de cada backend:
- **auth-service**: `GET /health` a cada 10s
- **product-service**: `GET /health` a cada 10s
- **publish-order-service**: `GET /health` a cada 10s
- **ui-service**: `GET /` a cada 30s

Se backend falhar, Caddy automaticamente remove do pool.

---

## 🚨 Troubleshooting

### Erro: "certificate is not trusted"

**Causa:** Certificado auto-assinado em desenvolvimento.

**Solução:**
1. Aceitar certificado no navegador
2. Ou instalar CA root do Caddy:
   ```bash
   docker exec caddy-proxy caddy trust
   ```

### Erro: "upstream connect error or disconnect/reset before headers"

**Causa:** Backend não está respondendo.

**Debug:**
```bash
# Verificar se backend está up
docker ps | grep -E "(auth-service|product-service|publish-order-service|ui-service)"

# Verificar logs do Caddy
docker logs caddy-proxy --tail 100

# Verificar logs do backend
docker logs auth-service --tail 50
```

### Erro: "connection refused"

**Causa:** Backend não está na mesma network do Caddy.

**Solução:** Verificar networks no `docker-compose.yaml`:
```yaml
caddy:
  networks:
    - frontend
    - auth
    - order

auth-service:
  networks:
    - auth
    - frontend
```

### SSE não funciona

**Causa:** Configurações de buffering.

**Verificação:**
```bash
curl -N -H "Accept: text/event-stream" https://velure.local/api/sse/orders/<order_id>
```

**Solução:** Garantir que Caddyfile tem:
```caddyfile
handle_path /api/sse* {
    reverse_proxy ... {
        flush_interval -1      # Flush imediato
        transport http {
            read_timeout 7200s  # 2 horas
        }
    }
}
```

---

## 🔐 Security Best Practices

### Desenvolvimento
- ✅ Certificados auto-assinados (`local_certs`)
- ✅ Headers de segurança habilitados
- ✅ CORS permissivo (`*`)

### Produção
- ✅ Let's Encrypt certificados válidos
- ✅ HSTS preload habilitado
- ✅ CORS restrito a domínios específicos
- ✅ Rate limiting (adicionar módulo)
- ✅ WAF (Web Application Firewall) - considerar Cloudflare
- ✅ DDoS protection - AWS Shield

### Rate Limiting (Adicionar)

```caddyfile
velure.com.br {
    # Rate limit: 100 req/min por IP
    rate_limit {
        zone dynamic {
            key {remote_host}
            events 100
            window 1m
        }
    }
    
    # ... resto da config
}
```

**Nota:** Requer módulo `caddy-ratelimit`.

---

## 🎨 Customização

### Adicionar Novo Serviço

1. **Adicionar rota no Caddyfile:**
   ```caddyfile
   handle_path /api/novo-servico* {
       reverse_proxy novo-servico:8080 {
           health_uri /health
           health_interval 10s
           health_timeout 5s
           health_status 2xx
       }
   }
   ```

2. **Adicionar network no docker-compose:**
   ```yaml
   caddy:
     depends_on:
       - novo-servico
   
   novo-servico:
     networks:
       - frontend  # Ou auth, order conforme necessário
   ```

3. **Recarregar config:**
   ```bash
   docker exec caddy-proxy caddy reload --config /etc/caddy/Caddyfile
   ```

### Custom Error Pages

```caddyfile
handle_errors {
    @404 expression `{http.error.status_code} == 404`
    handle @404 {
        root * /var/www/errors
        rewrite * /404.html
        file_server
    }
}
```

---

## 📚 Recursos

- [Caddy Documentation](https://caddyserver.com/docs/)
- [Caddyfile Syntax](https://caddyserver.com/docs/caddyfile)
- [Reverse Proxy Guide](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy)
- [Automatic HTTPS](https://caddyserver.com/docs/automatic-https)
- [JSON Config](https://caddyserver.com/docs/json/) (alternativa ao Caddyfile)

---

## 🐛 Debug Mode

Para mais verbosidade nos logs:

```caddyfile
{
    debug  # Habilita modo debug
}
```

Ou via linha de comando:
```bash
docker exec caddy-proxy caddy run --config /etc/caddy/Caddyfile --adapter caddyfile --debug
```

---

## ✅ Checklist - Deploy Produção

- [ ] Remover `local_certs` do global config
- [ ] Configurar email real para Let's Encrypt
- [ ] Descomentar seção de produção no Caddyfile
- [ ] Configurar DNS records apontando para load balancer
- [ ] Aguardar Let's Encrypt emitir certificados (automático)
- [ ] Testar HTTPS com [SSL Labs](https://www.ssllabs.com/ssltest/)
- [ ] Configurar CORS restrito (remover `*`)
- [ ] Adicionar rate limiting
- [ ] Configurar alertas para logs de erro
- [ ] Habilitar métricas Prometheus
- [ ] Testar failover de backends
- [ ] Documentar runbook de incidentes

---

**Pronto! Caddy está configurado e pronto para uso.** 🚀
