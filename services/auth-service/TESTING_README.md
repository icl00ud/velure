# Auth-Service - Testes Unitários

## Status Atual

Infraestrutura de testes unitários estabelecida com sucesso usando **gomock** e **table-driven tests** (padrão idiomático em Go).

### ✅ Implementado

1. **Setup de Infraestrutura**
   - ✅ Instalado `go.uber.org/mock` para mocking
   - ✅ Interfaces dos repositories criadas (`internal/repositories/interfaces.go`)
   - ✅ Mocks gerados com `mockgen` (`internal/mocks/mock_repositories.go`)

2. **Test Helpers** (`internal/testutil/`)
   - ✅ `fixtures.go` - Dados de teste reutilizáveis
     - `SetupTestDB()` - Banco SQLite in-memory para testes
     - `CreateTestUser()` - User factory com overrides opcionais
     - `CreateTestUsers()` - Multiple users factory
     - `CreateTestSession()` - Session factory
     - `CreateTestPasswordReset()` - PasswordReset factory
     - `CreateTestConfig()` - Config mock para JWT
     - Helpers para hashing/comparação de senhas

3. **Testes Implementados**
   - ✅ **UserRepository** (`internal/repositories/user_repository_test.go`)
     - 8 testes completos cobrindo todos os métodos
     - Table-driven tests para múltiplos cenários
     - ✅ TestUserRepository_Create (happy path + custom data)
     - ✅ TestUserRepository_Create_DuplicateEmail
     - ✅ TestUserRepository_GetByEmail (found + not found)
     - ✅ TestUserRepository_GetByID (found + not found)
     - ✅ TestUserRepository_GetAll
     - ✅ TestUserRepository_Update
     - ✅ TestUserRepository_Delete
     - ✅ TestUserRepository_GetByPage (4 cenários de paginação)

4. **Cobertura Atual**
   - Repository layer: **48.9%**
   - Total: **4.7%** (esperado, apenas 1 arquivo testado)
   - ✅ Todos os testes passando

5. **Arquivos Removidos**
   - ✅ `main_test.go` (testes de integração antigos substituídos)

---

## 📂 Estrutura Criada

```
services/auth-service/
├── internal/
│   ├── mocks/
│   │   └── mock_repositories.go         [NOVO - Gerado por mockgen]
│   ├── testutil/
│   │   └── fixtures.go                  [NOVO - Helpers reutilizáveis]
│   ├── repositories/
│   │   ├── interfaces.go                [NOVO - Interfaces para mocking]
│   │   └── user_repository_test.go      [NOVO - 8 testes, 100% pass]
│   ├── services/                        [TODO - Precisa testes]
│   ├── handlers/                        [TODO - Precisa testes]
│   ├── models/                          [TODO - Precisa testes]
│   ├── config/                          [TODO - Precisa testes]
│   └── middleware/                      [TODO - Precisa testes]
├── coverage.out                         [GERADO]
└── TESTING_README.md                    [NOVO - Este arquivo]
```

---

## 🚀 Como Rodar os Testes

### Todos os testes
```bash
cd services/auth-service
go test -v ./internal/...
```

### Com cobertura
```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Apenas UserRepository
```bash
go test -v ./internal/repositories/ -run TestUserRepository
```

### Com race detector
```bash
go test -race ./internal/...
```

---

## 📝 Próximos Passos para Completar

### 1. SessionRepository Tests (`internal/repositories/session_repository_test.go`)

Seguir o mesmo padrão de `user_repository_test.go`:

```go
func TestSessionRepository_Create(t *testing.T) { /* ... */ }
func TestSessionRepository_GetByUserID(t *testing.T) { /* ... */ }
func TestSessionRepository_GetByRefreshToken(t *testing.T) { /* ... */ }
func TestSessionRepository_Update(t *testing.T) { /* ... */ }
func TestSessionRepository_InvalidateByRefreshToken(t *testing.T) { /* ... */ }
func TestSessionRepository_Delete(t *testing.T) { /* ... */ }
```

### 2. PasswordResetRepository Tests (`internal/repositories/password_reset_repository_test.go`)

```go
func TestPasswordResetRepository_Create(t *testing.T) { /* ... */ }
func TestPasswordResetRepository_GetByToken(t *testing.T) { /* ... */ }
func TestPasswordResetRepository_Delete(t *testing.T) { /* ... */ }
```

### 3. AuthService Tests (CORE - Mais importante!)

**Arquivo:** `internal/services/auth_service_test.go`

**Padrão:** Usar **gomock** para mockar repositories

```go
package services

import (
	"testing"
	"velure-auth-service/internal/mocks"
	"velure-auth-service/internal/models"
	"velure-auth-service/internal/testutil"

	"go.uber.org/mock/gomock"
)

func TestAuthService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryInterface(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepositoryInterface(ctrl)
	mockPasswordResetRepo := mocks.NewMockPasswordResetRepositoryInterface(ctrl)

	cfg := testutil.CreateTestConfig()
	service := NewAuthService(mockUserRepo, mockSessionRepo, mockPasswordResetRepo, cfg)

	tests := []struct {
		name      string
		req       models.CreateUserRequest
		setupMock func()
		wantErr   bool
	}{
		{
			name: "successful user creation",
			req: testutil.CreateTestCreateUserRequest(),
			setupMock: func() {
				// Email não existe (GetByEmail retorna erro)
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound)

				// Create é chamado com sucesso
				mockUserRepo.EXPECT().
					Create(gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "duplicate email error",
			req: testutil.CreateTestCreateUserRequest(),
			setupMock: func() {
				// Email já existe
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any()).
					Return(testutil.CreateTestUser(), nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := service.CreateUser(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Testes necessários:**
- `TestAuthService_CreateUser` (✅ Exemplo acima)
- `TestAuthService_Login`
- `TestAuthService_ValidateAccessToken`
- `TestAuthService_Logout`
- `TestAuthService_GetUserByEmail`
- `TestAuthService_GetUserByID`
- `TestAuthService_GetUsers`
- `TestAuthService_generateAccessToken`
- `TestAuthService_generateRefreshToken`

### 4. AuthHandler Tests

**Arquivo:** `internal/handlers/auth_handler_test.go`

**Padrão:** Usar **httptest** e mockar AuthService

```go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"velure-auth-service/internal/mocks"
	"velure-auth-service/internal/models"
	"velure-auth-service/internal/testutil"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// TODO: Criar mock do AuthService
	// mockService := mocks.NewMockAuthServiceInterface(ctrl)
	// handler := NewAuthHandler(mockService)

	tests := []struct {
		name         string
		body         interface{}
		setupMock    func()
		expectedCode int
	}{
		{
			name: "successful registration",
			body: testutil.CreateTestCreateUserRequest(),
			setupMock: func() {
				// mockService.EXPECT().CreateUser(gomock.Any()).Return(nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "invalid JSON",
			body: "invalid json",
			setupMock: func() {},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			bodyBytes, _ := json.Marshal(tt.body)
			c.Request = httptest.NewRequest("POST", "/authentication/register", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			// handler.Register(c)

			if w.Code != tt.expectedCode {
				t.Errorf("Register() status = %d, want %d", w.Code, tt.expectedCode)
			}
		})
	}
}
```

**Testes necessários:**
- `TestAuthHandler_Register`
- `TestAuthHandler_Login`
- `TestAuthHandler_ValidateToken`
- `TestAuthHandler_GetUsers`
- `TestAuthHandler_GetUserByID`
- `TestAuthHandler_GetUserByEmail`
- `TestAuthHandler_Logout`

### 5. Models Tests

**Arquivo:** `internal/models/models_test.go`

```go
func TestUser_ToResponse(t *testing.T) {
	user := testutil.CreateTestUser()
	resp := user.ToResponse()

	// Password não deve estar presente
	if resp.ID != user.ID {
		t.Error("ToResponse() ID mismatch")
	}
	// ...
}

func TestUser_BeforeCreate(t *testing.T) { /* ... */ }
func TestUser_BeforeUpdate(t *testing.T) { /* ... */ }
```

**Arquivo:** `internal/models/pagination_test.go`

```go
func TestNewPaginatedUsersResponse(t *testing.T) {
	// Testar cálculo de totalPages, etc.
}
```

### 6. Config Tests

**Arquivo:** `internal/config/config_test.go`

```go
func TestLoad(t *testing.T) {
	// Usar t.Setenv() para mockar env vars
	t.Setenv("JWT_SECRET", "test-secret")

	cfg := Load()

	if cfg.JWT.Secret != "test-secret" {
		t.Error("Load() failed to read JWT_SECRET")
	}
}
```

### 7. Middleware Tests

**Arquivo:** `internal/middleware/middleware_test.go`

```go
func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.Status(200)
	})

	c.Request = httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, c.Request)

	// Verificar headers CORS
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS() missing Access-Control-Allow-Origin header")
	}
}
```

---

## 🎯 Meta de Cobertura

- **Repositories:** 80%+ (CRUD completo)
- **Services:** 80%+ (lógica de negócio crítica)
- **Handlers:** 70%+ (HTTP layer)
- **Models:** 60%+ (métodos e hooks)
- **Config/Middleware:** 50%+
- **Total:** **70-90%** ✅

---

## 💡 Dicas

### 1. Regenerar Mocks após Mudanças

Se modificar as interfaces:
```bash
cd services/auth-service
mockgen -source=internal/repositories/interfaces.go \
        -destination=internal/mocks/mock_repositories.go \
        -package=mocks
```

### 2. Criar Interface do AuthService

Para mockar no Handler, crie `internal/services/interfaces.go`:
```go
type AuthServiceInterface interface {
	CreateUser(req CreateUserRequest) error
	Login(req LoginRequest) (*LoginResponse, error)
	ValidateAccessToken(token string) (*ValidateTokenResponse, error)
	// ...todos os métodos públicos
}
```

E gere o mock:
```bash
mockgen -source=internal/services/interfaces.go \
        -destination=internal/mocks/mock_auth_service.go \
        -package=mocks
```

### 3. Table-Driven Tests

Sempre use table-driven tests para múltiplos cenários:
```go
tests := []struct {
	name    string
	input   X
	want    Y
	wantErr bool
}{
	{name: "case 1", input: ..., want: ..., wantErr: false},
	{name: "case 2", input: ..., want: ..., wantErr: true},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		// test logic
	})
}
```

### 4. Fixtures com Overrides

Use o padrão de overrides para customizar fixtures:
```go
user := testutil.CreateTestUser(func(u *models.User) {
	u.Email = "custom@example.com"
	u.Name = "Custom Name"
})
```

---

## 📚 Referências

- [Go Testing Best Practices](https://go.dev/blog/table-driven-tests)
- [gomock Documentation](https://github.com/uber-go/mock)
- [Gin Testing Guide](https://gin-gonic.com/docs/testing/)
- [GORM Testing](https://gorm.io/docs/testing.html)

---

## ✅ Checklist de Progresso

- [x] Setup de infraestrutura (gomock, mocks, fixtures)
- [x] Testes do UserRepository (100% completo)
- [ ] Testes do SessionRepository
- [ ] Testes do PasswordResetRepository
- [ ] Testes do AuthService (CORE)
- [ ] Testes do AuthHandler
- [ ] Testes de Models
- [ ] Testes de Config
- [ ] Testes de Middleware
- [ ] Cobertura total: 70-90%
