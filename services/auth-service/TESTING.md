# Auth Service - Test Coverage Documentation

## Overview

This document describes the comprehensive test coverage for the Velure Authentication Service, meeting academic requirements for >75% code coverage.

## Coverage Summary

**Total Coverage: 85.1%** (excluding generated code)

### Coverage by Layer

| Layer | Coverage | Tests | Description |
|-------|----------|-------|-------------|
| **Models** | 100% | 8 | User, Session, PasswordReset models |
| **Config** | 100% | 3 | Configuration loading and validation |
| **Middleware** | 100% | 10 | CORS, Logger, Prometheus middleware |
| **Handlers** | 94.2% | 20 | HTTP request handlers |
| **Services** | 89.9% | 26 | Business logic layer |
| **Repositories** | 97.9% | 17 | Data access layer |
| **Database** | 88.9% | 5 | Database connection and migrations |
| **Metrics** | Tested | 16 | Prometheus metrics |
| **Main Router** | 100% | 7 | Router setup and routes |

**Total Tests: 112 unit tests**

## Running Tests

### Quick Test Run

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/handlers -v
```

### Coverage Analysis

We provide a coverage script that automatically excludes generated code (mocks and testutil):

```bash
# Generate coverage report (recommended)
./coverage.sh

# Manual coverage analysis
go test ./... -coverprofile=coverage_full.out -covermode=atomic
grep -v "/internal/mocks/" coverage_full.out | grep -v "/internal/testutil/" > coverage.out
go tool cover -func=coverage.out | tail -1
```

### HTML Coverage Report

```bash
# Generate and open HTML report
./coverage.sh
open coverage.html

# Or manually
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## Test Organization

### Unit Tests by Package

#### 1. Models (`internal/models`)
- `user_test.go`: User model validation (password hashing, validation)
- `session_test.go`: Session model tests
- `password_reset_test.go`: Password reset model tests

#### 2. Configuration (`internal/config`)
- `config_test.go`: Environment variable loading and defaults

#### 3. Handlers (`internal/handlers`)
- `auth_handler_test.go`: HTTP endpoint tests
  - Registration (success, validation, conflicts)
  - Login (success, invalid credentials, validation)
  - Token validation (valid, invalid, expired)
  - User queries (by ID, by email, pagination)
  - Logout functionality

#### 4. Services (`internal/services`)
- `auth_service_test.go`: Business logic tests
  - User creation and validation
  - Authentication flow
  - Token generation and validation
  - Session management
  - Password reset workflows

#### 5. Repositories (`internal/repositories`)
- `user_repository_test.go`: User data access (CRUD operations)
- `session_repository_test.go`: Session data access
- `password_reset_repository_test.go`: Password reset data access

#### 6. Middleware (`internal/middleware`)
- `middleware_test.go`: CORS and Logger middleware
- `prometheus_test.go`: Prometheus metrics middleware

#### 7. Database (`internal/database`)
- `database_test.go`: Connection and migration tests

#### 8. Metrics (`internal/metrics`)
- `metrics_test.go`: Prometheus metrics functionality

#### 9. Main (`main.go`)
- `main_test.go`: Router setup and route registration

## Test Patterns and Best Practices

### Table-Driven Tests

Most tests use Go's table-driven test pattern:

```go
tests := []struct {
    name           string
    input          InputType
    expectedOutput OutputType
    expectError    bool
}{
    {"success case", validInput, expectedOutput, false},
    {"error case", invalidInput, nil, true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test implementation
    })
}
```

### Mock Usage

Tests use `gomock` for mocking dependencies:

```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockRepo := mocks.NewMockUserRepositoryInterface(ctrl)
mockRepo.EXPECT().Create(gomock.Any()).Return(nil)
```

### HTTP Handler Testing

Handler tests use `httptest` package:

```go
router := gin.New()
router.POST("/register", handler.Register)

req := httptest.NewRequest("POST", "/register", body)
w := httptest.NewRecorder()
router.ServeHTTP(w, req)

assert.Equal(t, http.StatusCreated, w.Code)
```

## Coverage Exclusions

The following directories are excluded from coverage calculations as they contain generated code:

- `internal/mocks/`: Auto-generated mock implementations (gomock)
- `internal/testutil/`: Test utilities (no production code)

This exclusion provides a more accurate representation of actual code coverage.

## CI/CD Integration

Tests run automatically on every commit via GitHub Actions:

```yaml
- name: Run tests with coverage
  run: go test ./... -coverprofile=coverage.out -covermode=atomic

- name: Check coverage threshold
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 75" | bc -l) )); then
      echo "Coverage $coverage% is below 75% threshold"
      exit 1
    fi
```

## Test Data and Fixtures

Tests use in-memory SQLite databases for isolation:

```go
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

This ensures:
- Fast test execution
- No external dependencies
- Complete isolation between tests
- Automatic cleanup

## Continuous Improvement

### Current High Coverage Areas
- Models: 100%
- Config: 100%
- Middleware: 100%
- Router setup: 100%

### Areas for Future Enhancement
- Integration tests for complete user flows
- Performance benchmarks
- Stress testing for concurrent operations

## Academic Requirements Met

This test suite meets and exceeds the following academic requirements:

1. **Minimum 75% Coverage**: ✅ 85.1% achieved
2. **Unit Test Coverage**: ✅ All layers tested
3. **Test Documentation**: ✅ This document
4. **CI/CD Integration**: ✅ Automated testing pipeline
5. **Best Practices**: ✅ Table-driven tests, mocking, isolation

## Tools Used

- **Testing Framework**: Go's built-in `testing` package
- **Mocking**: `go.uber.org/mock` (gomock)
- **HTTP Testing**: `net/http/httptest`
- **Coverage**: `go tool cover`
- **Assertions**: Custom assertions and go's testing utilities
- **Database**: SQLite in-memory for test isolation

## References

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [gomock Documentation](https://github.com/golang/mock)
- [GORM Testing](https://gorm.io/docs/testing.html)

---

**Last Updated**: 2025-01-13
**Minimum Coverage Requirement**: 75%
**Current Coverage**: 85.1%
**Status**: ✅ Requirements Exceeded
