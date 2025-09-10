# Portabilidade de APIs do ui-service para ui-service-2

Este documento descreve a portabilidade das APIs do frontend Angular (`ui-service`) para o frontend React/Vite (`ui-service-2`).

## Estrutura de Arquivos Criados

### Serviços (`/src/services/`)
- `config.service.ts` - Configuração de URLs das APIs
- `authentication.service.ts` - Serviços de autenticação (login, logout, registro)
- `product.service.ts` - Serviços de produtos (CRUD, paginação, busca)
- `cart.service.ts` - Gerenciamento do carrinho de compras
- `blob.service.ts` - Manipulação de blobs e downloads
- `index.ts` - Exportações centralizadas

### Types (`/src/types/`)
- `user.types.ts` - Interfaces para usuários e autenticação
- `product.types.ts` - Interfaces para produtos e carrinho
- `index.ts` - Exportações centralizadas

### Hooks React (`/src/hooks/`)
- `use-auth.ts` - Hook para gerenciamento de autenticação
- `use-cart.ts` - Hook para gerenciamento do carrinho
- `use-products.ts` - Hooks para busca de produtos (simples e paginados)

### Configuração
- `config/environment.ts` - Configuração de ambiente baseada em variáveis Vite

## APIs Portadas

### 1. Authentication Service
- ✅ Login
- ✅ Logout  
- ✅ Registro
- ✅ Validação de token
- ✅ Verificação de autenticação

### 2. Product Service
- ✅ Buscar todos os produtos
- ✅ Buscar produto por ID
- ✅ Buscar produtos com paginação
- ✅ Buscar produtos por categoria e paginação
- ✅ Contagem de produtos
- ✅ CRUD de produtos (criar, atualizar, deletar)

### 3. Cart Service
- ✅ Adicionar produto ao carrinho
- ✅ Remover produto do carrinho
- ✅ Atualizar quantidade
- ✅ Calcular preço total
- ✅ Limpar carrinho
- ✅ Persistência no localStorage
- ✅ Sistema de observadores (reactive)

### 4. Blob Service
- ✅ Converter URL para base64
- ✅ Download de arquivos

## Principais Diferenças

### Angular → React
- **Observable → Promise/async-await**: Migrado de RxJS Observable para Promise nativo
- **HttpClient → fetch**: Substituído HttpClient do Angular por fetch API nativo
- **BehaviorSubject → Custom Observers**: Sistema próprio de observadores para reatividade
- **Dependency Injection → Singleton Services**: Classes singleton ao invés de DI do Angular
- **Guards → Hooks**: Guards do Angular convertidos em hooks React customizados

### Melhorias Implementadas
- **TypeScript mais rigoroso**: Tipos mais específicos e seguros
- **Error Handling**: Tratamento de erro melhorado
- **Loading States**: Estados de carregamento integrados nos hooks
- **Reactive Updates**: Sistema de observers para updates em tempo real
- **Environment Variables**: Suporte a variáveis de ambiente do Vite

## Como Usar

### 1. Autenticação
```typescript
import { useAuth } from '@/hooks/use-auth';

const { isAuthenticated, login, logout, register, isLoading } = useAuth();

// Login
await login({ email: 'user@example.com', password: 'password' });

// Registro
await register({ name: 'User', email: 'user@example.com', password: 'password' });

// Logout
await logout();
```

### 2. Produtos
```typescript
import { useProducts, useProductsPaginated } from '@/hooks/use-products';

// Todos os produtos
const { products, loading, error } = useProducts();

// Produtos paginados
const { products, loading, totalPages } = useProductsPaginated(1, 10, 'dogs');
```

### 3. Carrinho
```typescript
import { useCart } from '@/hooks/use-cart';

const { cartItems, totalPrice, addToCart, removeFromCart } = useCart();

// Adicionar ao carrinho
addToCart(product, 2);

// Remover do carrinho
removeFromCart(productId);
```

## Configuração das Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
VITE_PRODUCT_SERVICE_URL=http://localhost:3010
VITE_AUTHENTICATION_SERVICE_URL=http://localhost:3020
```

## Próximos Passos

1. **Testes**: Implementar testes unitários para os serviços
2. **Error Boundaries**: Adicionar Error Boundaries para tratamento de erros
3. **Interceptors**: Implementar interceptors para tokens JWT
4. **Offline Support**: Adicionar suporte offline com Service Workers
5. **Cache**: Implementar cache inteligente para produtos
6. **Otimização**: Lazy loading e code splitting

## Compatibilidade

- ✅ Mantém compatibilidade total com as APIs backend existentes
- ✅ Preserva todas as funcionalidades do frontend Angular
- ✅ Melhora a experiência do desenvolvedor com hooks React
- ✅ Performance otimizada com bundle menor (Vite vs Angular CLI)
