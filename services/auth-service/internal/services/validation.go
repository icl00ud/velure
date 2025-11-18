package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Pool de validadores para processar validações em paralelo
	validationPool = sync.Pool{
		New: func() interface{} {
			return &validator{}
		},
	}
)

type validator struct{}

type ValidationResult struct {
	IsValid bool
	Error   error
	Field   string
}

// ValidateRegistrationAsync executa validações em paralelo
func ValidateRegistrationAsync(ctx context.Context, name, email, password string) []ValidationResult {
	results := make([]ValidationResult, 3)
	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine 1: Validar nome
	go func() {
		defer wg.Done()
		results[0] = validateName(name)
	}()

	// Goroutine 2: Validar email
	go func() {
		defer wg.Done()
		results[1] = validateEmail(email)
	}()

	// Goroutine 3: Validar password
	go func() {
		defer wg.Done()
		results[2] = validatePassword(password)
	}()

	// Aguardar com timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return results
	case <-ctx.Done():
		return []ValidationResult{{IsValid: false, Error: ctx.Err(), Field: "timeout"}}
	}
}

func validateName(name string) ValidationResult {
	name = strings.TrimSpace(name)
	if name == "" {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("name is required"), Field: "name"}
	}
	if len(name) < 2 {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("name must be at least 2 characters"), Field: "name"}
	}
	if len(name) > 100 {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("name must be less than 100 characters"), Field: "name"}
	}
	return ValidationResult{IsValid: true, Field: "name"}
}

func validateEmail(email string) ValidationResult {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("email is required"), Field: "email"}
	}
	if !emailRegex.MatchString(email) {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("invalid email format"), Field: "email"}
	}
	if len(email) > 255 {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("email must be less than 255 characters"), Field: "email"}
	}
	return ValidationResult{IsValid: true, Field: "email"}
}

func validatePassword(password string) ValidationResult {
	if password == "" {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("password is required"), Field: "password"}
	}
	if len(password) < 6 {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("password must be at least 6 characters"), Field: "password"}
	}
	if len(password) > 128 {
		return ValidationResult{IsValid: false, Error: fmt.Errorf("password must be less than 128 characters"), Field: "password"}
	}
	return ValidationResult{IsValid: true, Field: "password"}
}

// NormalizeEmail normaliza o email (lowercase, trim)
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

// BatchValidateEmails valida múltiplos emails em paralelo
func BatchValidateEmails(ctx context.Context, emails []string) map[string]ValidationResult {
	results := make(map[string]ValidationResult, len(emails))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limitar concorrência
	semaphore := make(chan struct{}, 10)

	for _, email := range emails {
		wg.Add(1)
		go func(e string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := validateEmail(e)
			mu.Lock()
			results[e] = result
			mu.Unlock()
		}(email)
	}

	// Aguardar com timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return results
	case <-ctx.Done():
		return results
	case <-time.After(5 * time.Second):
		return results
	}
}
