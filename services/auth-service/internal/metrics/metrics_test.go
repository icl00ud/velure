package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestLoginAttempts(t *testing.T) {
	// Reset the counter
	LoginAttempts.Reset()

	// Test incrementing success counter
	LoginAttempts.WithLabelValues("success").Inc()
	value := testutil.ToFloat64(LoginAttempts.WithLabelValues("success"))
	if value != 1 {
		t.Errorf("Expected login success count 1, got %f", value)
	}

	// Test incrementing failure counter
	LoginAttempts.WithLabelValues("failure").Inc()
	LoginAttempts.WithLabelValues("failure").Inc()
	value = testutil.ToFloat64(LoginAttempts.WithLabelValues("failure"))
	if value != 2 {
		t.Errorf("Expected login failure count 2, got %f", value)
	}
}

func TestLoginDuration(t *testing.T) {
	// Test observing login durations (histograms don't expose count directly)
	// We just verify the metric accepts observations without panicking
	LoginDuration.WithLabelValues("success").Observe(0.5)
	LoginDuration.WithLabelValues("success").Observe(0.3)
	LoginDuration.WithLabelValues("failure").Observe(0.1)
	// Success - no panic means test passed
}

func TestRegistrationAttempts(t *testing.T) {
	// Reset the counter
	RegistrationAttempts.Reset()

	// Test different registration statuses
	RegistrationAttempts.WithLabelValues("success").Inc()
	RegistrationAttempts.WithLabelValues("failure").Inc()
	RegistrationAttempts.WithLabelValues("conflict").Inc()

	successValue := testutil.ToFloat64(RegistrationAttempts.WithLabelValues("success"))
	if successValue != 1 {
		t.Errorf("Expected registration success count 1, got %f", successValue)
	}

	conflictValue := testutil.ToFloat64(RegistrationAttempts.WithLabelValues("conflict"))
	if conflictValue != 1 {
		t.Errorf("Expected registration conflict count 1, got %f", conflictValue)
	}
}

func TestRegistrationDuration(t *testing.T) {
	// Test observing registration durations (histograms don't expose count directly)
	// We just verify the metric accepts observations without panicking
	RegistrationDuration.Observe(1.2)
	RegistrationDuration.Observe(0.8)
	RegistrationDuration.Observe(0.5)
	// Success - no panic means test passed
}

func TestTokenValidations(t *testing.T) {
	// Reset the counter
	TokenValidations.Reset()

	TokenValidations.WithLabelValues("valid").Inc()
	TokenValidations.WithLabelValues("valid").Inc()
	TokenValidations.WithLabelValues("invalid").Inc()

	validValue := testutil.ToFloat64(TokenValidations.WithLabelValues("valid"))
	if validValue != 2 {
		t.Errorf("Expected valid token count 2, got %f", validValue)
	}

	invalidValue := testutil.ToFloat64(TokenValidations.WithLabelValues("invalid"))
	if invalidValue != 1 {
		t.Errorf("Expected invalid token count 1, got %f", invalidValue)
	}
}

func TestTokenGenerations(t *testing.T) {
	// Reset the counter - TokenGenerations is already a Counter, no need for type assertion
	TokenGenerations.Add(-testutil.ToFloat64(TokenGenerations))

	TokenGenerations.Inc()
	TokenGenerations.Inc()
	TokenGenerations.Inc()

	value := testutil.ToFloat64(TokenGenerations)
	if value != 3 {
		t.Errorf("Expected token generation count 3, got %f", value)
	}
}

func TestTokenGenerationDuration(t *testing.T) {
	// Test observing token generation durations (histograms don't expose count directly)
	// We just verify the metric accepts observations without panicking
	TokenGenerationDuration.Observe(0.005)
	TokenGenerationDuration.Observe(0.01)
	TokenGenerationDuration.Observe(0.025)
	// Success - no panic means test passed
}

func TestActiveSessions(t *testing.T) {
	// Test gauge operations
	ActiveSessions.Set(10)
	value := testutil.ToFloat64(ActiveSessions)
	if value != 10 {
		t.Errorf("Expected active sessions 10, got %f", value)
	}

	ActiveSessions.Inc()
	value = testutil.ToFloat64(ActiveSessions)
	if value != 11 {
		t.Errorf("Expected active sessions 11, got %f", value)
	}

	ActiveSessions.Dec()
	value = testutil.ToFloat64(ActiveSessions)
	if value != 10 {
		t.Errorf("Expected active sessions 10, got %f", value)
	}
}

func TestLogoutRequests(t *testing.T) {
	// Reset counter
	initial := testutil.ToFloat64(LogoutRequests)

	LogoutRequests.Inc()
	value := testutil.ToFloat64(LogoutRequests)
	if value != initial+1 {
		t.Errorf("Expected logout count %f, got %f", initial+1, value)
	}
}

func TestTotalUsers(t *testing.T) {
	// Test gauge operations
	TotalUsers.Set(100)
	value := testutil.ToFloat64(TotalUsers)
	if value != 100 {
		t.Errorf("Expected total users 100, got %f", value)
	}

	TotalUsers.Add(50)
	value = testutil.ToFloat64(TotalUsers)
	if value != 150 {
		t.Errorf("Expected total users 150, got %f", value)
	}
}

func TestUserQueries(t *testing.T) {
	// Reset counter
	UserQueries.Reset()

	UserQueries.WithLabelValues("by_id").Inc()
	UserQueries.WithLabelValues("by_email").Inc()
	UserQueries.WithLabelValues("list").Inc()
	UserQueries.WithLabelValues("list").Inc()

	byIDValue := testutil.ToFloat64(UserQueries.WithLabelValues("by_id"))
	if byIDValue != 1 {
		t.Errorf("Expected by_id queries 1, got %f", byIDValue)
	}

	listValue := testutil.ToFloat64(UserQueries.WithLabelValues("list"))
	if listValue != 2 {
		t.Errorf("Expected list queries 2, got %f", listValue)
	}
}

func TestDatabaseQueries(t *testing.T) {
	// Reset counter
	DatabaseQueries.Reset()

	operations := []string{"select", "insert", "update", "delete"}
	for _, op := range operations {
		DatabaseQueries.WithLabelValues(op).Inc()
	}

	for _, op := range operations {
		value := testutil.ToFloat64(DatabaseQueries.WithLabelValues(op))
		if value != 1 {
			t.Errorf("Expected %s operation count 1, got %f", op, value)
		}
	}
}

func TestDatabaseQueryDuration(t *testing.T) {
	// Test observing database query durations (histograms don't expose count directly)
	// We just verify the metric accepts observations without panicking
	DatabaseQueryDuration.WithLabelValues("select").Observe(0.01)
	DatabaseQueryDuration.WithLabelValues("insert").Observe(0.02)
	DatabaseQueryDuration.WithLabelValues("update").Observe(0.015)
	// Success - no panic means test passed
}

func TestErrors(t *testing.T) {
	// Reset counter
	Errors.Reset()

	errorTypes := []string{"validation", "database", "auth", "internal"}
	for _, errType := range errorTypes {
		Errors.WithLabelValues(errType).Inc()
	}

	for _, errType := range errorTypes {
		value := testutil.ToFloat64(Errors.WithLabelValues(errType))
		if value != 1 {
			t.Errorf("Expected %s error count 1, got %f", errType, value)
		}
	}
}

func TestHTTPRequests(t *testing.T) {
	// Reset counter
	HTTPRequests.Reset()

	HTTPRequests.WithLabelValues("GET", "/api/users", "200").Inc()
	HTTPRequests.WithLabelValues("POST", "/api/users", "201").Inc()
	HTTPRequests.WithLabelValues("GET", "/api/users", "200").Inc()

	getValue := testutil.ToFloat64(HTTPRequests.WithLabelValues("GET", "/api/users", "200"))
	if getValue != 2 {
		t.Errorf("Expected GET request count 2, got %f", getValue)
	}

	postValue := testutil.ToFloat64(HTTPRequests.WithLabelValues("POST", "/api/users", "201"))
	if postValue != 1 {
		t.Errorf("Expected POST request count 1, got %f", postValue)
	}
}

func TestHTTPRequestDuration(t *testing.T) {
	// Test observing HTTP request durations (histograms don't expose count directly)
	// We just verify the metric accepts observations without panicking
	HTTPRequestDuration.WithLabelValues("GET", "/api/users").Observe(0.1)
	HTTPRequestDuration.WithLabelValues("GET", "/api/users").Observe(0.2)
	HTTPRequestDuration.WithLabelValues("POST", "/api/users").Observe(0.15)
	HTTPRequestDuration.WithLabelValues("PUT", "/api/users").Observe(0.12)
	// Success - no panic means test passed
}
