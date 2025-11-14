package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedURI    string
		expectedDBName string
		expectedPort   string
	}{
		{
			name: "with full MONGODB_URI",
			envVars: map[string]string{
				"MONGODB_URI":              "mongodb+srv://user:pass@cluster.mongodb.net/mydb",
				"MONGODB_DBNAME":           "testdb",
				"PRODUCT_SERVICE_APP_PORT": "3000",
				"REDIS_HOST":               "redis-host",
				"REDIS_PORT":               "6380",
			},
			expectedURI:    "mongodb+srv://user:pass@cluster.mongodb.net/mydb",
			expectedDBName: "testdb",
			expectedPort:   "3000",
		},
		{
			name: "with individual MongoDB components - with auth",
			envVars: map[string]string{
				"MONGODB_HOST":             "mongo-host",
				"MONGODB_PORT":             "27018",
				"MONGODB_NORMAL_USER":      "testuser",
				"MONGODB_NORMAL_PASSWORD":  "testpass",
				"MONGODB_AUTH_DATABASE":    "admin",
				"MONGODB_DBNAME":           "products",
				"PRODUCT_SERVICE_APP_PORT": "3010",
			},
			expectedURI:    "mongodb://testuser:testpass@mongo-host:27018/?authSource=admin",
			expectedDBName: "products",
			expectedPort:   "3010",
		},
		{
			name: "with individual MongoDB components - without auth",
			envVars: map[string]string{
				"MONGODB_HOST":             "localhost",
				"MONGODB_PORT":             "27017",
				"MONGODB_DBNAME":           "mydb",
				"PRODUCT_SERVICE_APP_PORT": "8080",
			},
			expectedURI:    "mongodb://localhost:27017",
			expectedDBName: "mydb",
			expectedPort:   "8080",
		},
		{
			name:           "with defaults",
			envVars:        map[string]string{},
			expectedURI:    "mongodb://localhost:27017",
			expectedDBName: "product_service",
			expectedPort:   "3010",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create config
			cfg := New()

			// Assert expectations
			assert.Equal(t, tt.expectedURI, cfg.MongoURI)
			assert.Equal(t, tt.expectedDBName, cfg.DatabaseName)
			assert.Equal(t, tt.expectedPort, cfg.Port)

			// Clean up
			os.Clearenv()
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		envValue      string
		setEnv        bool
		expectedValue string
	}{
		{
			name:          "environment variable set",
			key:           "TEST_VAR",
			defaultValue:  "default",
			envValue:      "custom",
			setEnv:        true,
			expectedValue: "custom",
		},
		{
			name:          "environment variable not set",
			key:           "TEST_VAR",
			defaultValue:  "default",
			envValue:      "",
			setEnv:        false,
			expectedValue: "default",
		},
		{
			name:          "environment variable set to empty string",
			key:           "TEST_VAR",
			defaultValue:  "default",
			envValue:      "",
			setEnv:        true,
			expectedValue: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the specific env var
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expectedValue, result)

			// Clean up
			os.Unsetenv(tt.key)
		})
	}
}

func TestNewRedis(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		password string
	}{
		{
			name:     "with password",
			addr:     "localhost:6379",
			password: "secret",
		},
		{
			name:     "without password",
			addr:     "localhost:6379",
			password: "",
		},
		{
			name:     "custom address",
			addr:     "redis-host:6380",
			password: "mypassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewRedis(tt.addr, tt.password)

			assert.NotNil(t, client)
			assert.Equal(t, tt.addr, client.Options().Addr)
			assert.Equal(t, tt.password, client.Options().Password)
			assert.Equal(t, 0, client.Options().DB)

			// Clean up
			client.Close()
		})
	}
}

func TestNewMongoDB_InvalidURI(t *testing.T) {
	// Test with an invalid URI
	client, err := NewMongoDB("invalid://uri")
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewMongoDB_AtlasURI(t *testing.T) {
	// This test will fail to connect but should validate URI format
	t.Skip("Skipping MongoDB connection test - requires actual MongoDB instance")
}

func TestNewMongoDB_StandardURI(t *testing.T) {
	// This test will fail to connect but should validate URI format
	t.Skip("Skipping MongoDB connection test - requires actual MongoDB instance")
}

func TestConfig_RedisAddr(t *testing.T) {
	os.Clearenv()
	os.Setenv("REDIS_HOST", "custom-redis")
	os.Setenv("REDIS_PORT", "6380")

	cfg := New()

	assert.Equal(t, "custom-redis:6380", cfg.RedisAddr)

	os.Clearenv()
}

func TestConfig_RedisPassword(t *testing.T) {
	os.Clearenv()
	os.Setenv("REDIS_PASSWORD", "mypassword")

	cfg := New()

	assert.Equal(t, "mypassword", cfg.RedisPassword)

	os.Clearenv()
}

func TestConfig_Defaults(t *testing.T) {
	os.Clearenv()

	cfg := New()

	assert.Equal(t, "mongodb://localhost:27017", cfg.MongoURI)
	assert.Equal(t, "product_service", cfg.DatabaseName)
	assert.Equal(t, "localhost:6379", cfg.RedisAddr)
	assert.Equal(t, "", cfg.RedisPassword)
	assert.Equal(t, "3010", cfg.Port)

	os.Clearenv()
}

func TestConfig_MongoURIPriority(t *testing.T) {
	os.Clearenv()

	// Set both MONGODB_URI and individual components
	os.Setenv("MONGODB_URI", "mongodb+srv://priority@cluster.mongodb.net/db")
	os.Setenv("MONGODB_HOST", "localhost")
	os.Setenv("MONGODB_PORT", "27017")

	cfg := New()

	// MONGODB_URI should take priority
	assert.Equal(t, "mongodb+srv://priority@cluster.mongodb.net/db", cfg.MongoURI)

	os.Clearenv()
}
