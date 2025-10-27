# 🚨 IMPORTANTE: Como acessar a aplicação Velure

## ❌ URL Errada (acesso direto ao container)
```
https://ui-service.local.orb.local
```
**Problema:** Acessa o nginx do UI service diretamente, sem passar pelo Caddy proxy. As chamadas `/api/*` não são roteadas corretamente.

## ✅ URL Correta (através do Caddy proxy)
```
https://velure.local
```
**Benefícios:** 
- Todas as rotas `/api/auth/*` são proxy para auth-service
- Todas as rotas `/api/product/*` são proxy para product-service  
- Todas as rotas `/api/order/*` são proxy para publish-order-service
- CORS configurado corretamente
- Headers de segurança aplicados

---

## 🔧 Configuração Necessária

### 1. Adicionar ao /etc/hosts

```bash
sudo nano /etc/hosts
```

Adicione:
```
127.0.0.1 velure.local
```

### 2. Aceitar o Certificado SSL Local

Quando acessar `https://velure.local` pela primeira vez:
1. Chrome/Safari mostrará aviso de certificado
2. Clique em "Avançado" → "Continuar para velure.local"
3. O certificado é auto-assinado para desenvolvimento local

---

## 📡 Endpoints Disponíveis

### Frontend
- **URL:** https://velure.local
- **Descrição:** Interface React (Vite + TypeScript)

### Auth API
- **URL:** https://velure.local/api/auth/*
- **Proxy para:** auth-service:3020
- **Exemplos:**
  - POST /api/auth/register
  - POST /api/auth/login
  - POST /api/auth/validateToken
  - GET /api/auth/users

### Product API
- **URL:** https://velure.local/api/product/*
- **Proxy para:** product-service:3010
- **Exemplos:**
  - GET /api/product/categories
  - GET /api/product/products
  - GET /api/product/product/:id

### Order API
- **URL:** https://velure.local/api/order/*
- **Proxy para:** publish-order-service:8080
- **Exemplos:**
  - POST /api/order/orders
  - GET /api/order/orders
  - GET /api/order/stream (SSE)

---

## 🧪 Testar Registro de Usuário

### Via Browser Console (https://velure.local)
```javascript
fetch("/api/auth/register", {
  method: "POST",
  headers: {
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    name: "Israel Schroeder",
    email: "israelschroederm@gmail.com",
    password: "Mano@11sou"
  })
})
.then(r => r.json())
.then(console.log);
```

### Via cURL
```bash
curl -X POST https://velure.local/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Israel Schroeder",
    "email": "israelschroederm@gmail.com",
    "password": "Mano@11sou"
  }' \
  --insecure
```

---

## 🐛 Troubleshooting

### Erro: "ERR_NAME_NOT_RESOLVED"
```bash
# Adicionar ao /etc/hosts
echo "127.0.0.1 velure.local" | sudo tee -a /etc/hosts
```

### Erro: "405 Method Not Allowed" em /api/auth/*
**Causa:** Está acessando via `ui-service.local.orb.local`  
**Solução:** Acessar via `https://velure.local`

### Verificar se Caddy está funcionando
```bash
# Verificar se está ouvindo nas portas
docker ps | grep caddy

# Ver logs do Caddy
docker logs caddy-proxy --tail=50

# Testar endpoint direto
curl http://localhost/health
```

### Verificar roteamento
```bash
# Testar auth service através do Caddy
curl http://localhost/api/auth/users

# Testar product service através do Caddy  
curl http://localhost/api/product/categories
```

---

## 🏗️ Arquitetura do Roteamento

```
Browser (https://velure.local)
         │
         ▼
   Caddy Proxy (Port 443)
         │
         ├─── /api/auth/* ──────► auth-service:3020
         │
         ├─── /api/product/* ───► product-service:3010
         │
         ├─── /api/order/* ─────► publish-order-service:8080
         │
         └─── /* ───────────────► ui-service:8080 (React SPA)
```

**IMPORTANTE:** Sempre acesse via `https://velure.local` para que as rotas funcionem corretamente!
