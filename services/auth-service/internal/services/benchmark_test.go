package services

import (
	"context"
	"testing"
	"time"
	"velure-auth-service/internal/config"
	"velure-auth-service/internal/models"
)

// BenchmarkRegistrationSequential testa registro sequencial (antes)
func BenchmarkRegistrationSequential(b *testing.B) {
	// Setup
	cfg := &config.Config{
		Performance: config.PerformanceConfig{
			BcryptCost:    10,
			BcryptWorkers: 10,
			EnableCache:   false,
		},
	}

	service := setupTestService(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := models.CreateUserRequest{
				Name:     "Test User",
				Email:    generateEmail(i),
				Password: "password123",
			}
			_, _ = service.CreateUser(req)
			i++
		}
	})
}

// BenchmarkRegistrationParallel testa registro paralelo (depois)
func BenchmarkRegistrationParallel(b *testing.B) {
	// Setup com otimizações
	cfg := &config.Config{
		Performance: config.PerformanceConfig{
			BcryptCost:    10,
			BcryptWorkers: 20,
			EnableCache:   true,
			TokenCacheTTL: 300,
		},
	}

	service := setupTestService(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := models.CreateUserRequest{
				Name:     "Test User",
				Email:    generateEmail(i),
				Password: "password123",
			}
			_, _ = service.CreateUser(req)
			i++
		}
	})
}

// BenchmarkLoginWithoutCache testa login sem cache
func BenchmarkLoginWithoutCache(b *testing.B) {
	cfg := &config.Config{
		Performance: config.PerformanceConfig{
			BcryptCost:    10,
			BcryptWorkers: 10,
			EnableCache:   false,
		},
	}

	service := setupTestService(cfg)

	// Criar usuário de teste
	req := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, _ = service.CreateUser(req)

	loginReq := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = service.Login(loginReq)
		}
	})
}

// BenchmarkLoginWithCache testa login com cache
func BenchmarkLoginWithCache(b *testing.B) {
	cfg := &config.Config{
		Performance: config.PerformanceConfig{
			BcryptCost:    10,
			BcryptWorkers: 10,
			EnableCache:   true,
			TokenCacheTTL: 300,
		},
	}

	service := setupTestService(cfg)

	// Criar usuário de teste
	req := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	resp, _ := service.CreateUser(req)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Validar token (cache hit após primeira vez)
			_, _ = service.ValidateAccessToken(resp.AccessToken)
		}
	})
}

// BenchmarkTokenGenerationSequential testa geração sequencial de tokens
func BenchmarkTokenGenerationSequential(b *testing.B) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret",
			RefreshSecret: "test-refresh-secret",
		},
	}

	service := setupTestService(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.generateAccessToken(uint(i))
		_, _ = service.generateRefreshToken(uint(i))
	}
}

// BenchmarkTokenGenerationParallel testa geração paralela de tokens
func BenchmarkTokenGenerationParallel(b *testing.B) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret",
			RefreshSecret: "test-refresh-secret",
		},
	}

	service := setupTestService(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Simular geração paralela
			done := make(chan bool, 2)

			go func() {
				_, _ = service.generateAccessToken(uint(i))
				done <- true
			}()

			go func() {
				_, _ = service.generateRefreshToken(uint(i))
				done <- true
			}()

			<-done
			<-done
			i++
		}
	})
}

// BenchmarkValidationParallel testa validação paralela de campos
func BenchmarkValidationParallel(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			_ = ValidateRegistrationAsync(ctx, "John Doe", "john@example.com", "password123")
		}
	})
}

// BenchmarkCacheGetOrCompute testa cache com lazy loading
func BenchmarkCacheGetOrCompute(b *testing.B) {
	cache := NewDistributedCache(1 * time.Minute)
	defer cache.Stop()

	compute := func() (interface{}, error) {
		time.Sleep(10 * time.Millisecond) // Simular operação custosa
		return "computed value", nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ctx := context.Background()
			_, _ = cache.GetOrCompute(ctx, "key", 5*time.Second, compute)
			i++
		}
	})
}

// BenchmarkBatchCacheOperations testa operações em lote
func BenchmarkBatchCacheOperations(b *testing.B) {
	cache := NewDistributedCache(1 * time.Minute)
	defer cache.Stop()

	items := make(map[string]interface{}, 100)
	keys := make([]string, 100)
	for i := 0; i < 100; i++ {
		key := generateKey(i)
		items[key] = generateValue(i)
		keys[i] = key
	}

	b.ResetTimer()
	b.Run("BatchSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.BatchSet(items, 5*time.Second)
		}
	})

	b.Run("BatchGet", func(b *testing.B) {
		cache.BatchSet(items, 5*time.Second)
		for i := 0; i < b.N; i++ {
			_ = cache.BatchGet(keys)
		}
	})
}

// BenchmarkWorkerPool testa worker pool do bcrypt
func BenchmarkWorkerPool(b *testing.B) {
	workerPool := make(chan struct{}, 10)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			workerPool <- struct{}{}         // acquire
			time.Sleep(1 * time.Millisecond) // simular trabalho
			<-workerPool                     // release
		}
	})
}

// Helper functions
func setupTestService(cfg *config.Config) *AuthService {
	// Mock repositories e setup básico
	return &AuthService{
		config:           cfg,
		bcryptWorkerPool: make(chan struct{}, cfg.Performance.BcryptWorkers),
	}
}

func generateEmail(i int) string {
	return "user" + string(rune(i)) + "@example.com"
}

func generateKey(i int) string {
	return "key:" + string(rune(i))
}

func generateValue(i int) interface{} {
	return map[string]interface{}{
		"id":    i,
		"value": "data" + string(rune(i)),
	}
}
