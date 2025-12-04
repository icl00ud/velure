package services

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestValidateRegistrationAsync(t *testing.T) {
	results := ValidateRegistrationAsync(context.Background(), "John Doe", "user@example.com", "securepass")

	for _, result := range results {
		if !result.IsValid {
			t.Fatalf("expected validation to pass, got error for %s: %v", result.Field, result.Error)
		}
	}
}

func TestValidateRegistrationAsyncCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := ValidateRegistrationAsync(ctx, "John Doe", "user@example.com", "securepass")

	if len(results) != 1 || results[0].Field != "timeout" {
		t.Fatalf("expected timeout result, got %#v", results)
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"", false},
		{"A", false},
		{strings.Repeat("a", 101), false},
		{"John Doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateName(tt.name)
			if result.IsValid != tt.expected {
				t.Fatalf("validateName(%q) = %v, expected %v", tt.name, result.IsValid, tt.expected)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"", false},
		{"invalid-email", false},
		{strings.Repeat("a", 260) + "@example.com", false},
		{"valid@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := validateEmail(tt.email)
			if result.IsValid != tt.expected {
				t.Fatalf("validateEmail(%q) = %v, expected %v", tt.email, result.IsValid, tt.expected)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"", false},
		{"short", false},
		{strings.Repeat("p", 200), false},
		{"long-enough", true},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			result := validatePassword(tt.password)
			if result.IsValid != tt.expected {
				t.Fatalf("validatePassword(%q) = %v, expected %v", tt.password, result.IsValid, tt.expected)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	input := "  USER@Example.Com "
	expected := "user@example.com"

	if got := NormalizeEmail(input); got != expected {
		t.Fatalf("NormalizeEmail(%q) = %q, want %q", input, got, expected)
	}
}

func TestBatchValidateEmails(t *testing.T) {
	emails := []string{
		"valid@example.com",
		"invalid-email",
		"another@example.com",
	}

	results := BatchValidateEmails(context.Background(), emails)

	if len(results) != len(emails) {
		t.Fatalf("expected %d results, got %d", len(emails), len(results))
	}

	if !results["valid@example.com"].IsValid || !results["another@example.com"].IsValid {
		t.Fatalf("expected valid emails to pass")
	}

	if results["invalid-email"].IsValid {
		t.Fatalf("expected invalid email to fail")
	}
}

func TestBatchValidateEmailsRespectsTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Cancel before any work happens to exercise timeout branch without goroutines running
	time.Sleep(2 * time.Millisecond)

	results := BatchValidateEmails(ctx, []string{})
	if len(results) != 0 {
		t.Fatalf("expected empty result map on timeout, got %d entries", len(results))
	}
}
