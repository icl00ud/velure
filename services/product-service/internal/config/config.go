package config

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoURI      string
	DatabaseName  string
	RedisAddr     string
	RedisPassword string
	Port          string
}

func New() *Config {
	// Priority 1: Check if full MONGODB_URI is provided (for MongoDB Atlas with mongodb+srv://)
	mongoURI := getEnv("MONGODB_URI", "")

	// Priority 2: If not provided, construct from individual components (for local/self-hosted MongoDB)
	if mongoURI == "" {
		mongoHost := getEnv("MONGODB_HOST", "localhost")
		mongoPort := getEnv("MONGODB_PORT", "27017")
		mongoUser := getEnv("MONGODB_NORMAL_USER", "")
		mongoPassword := getEnv("MONGODB_NORMAL_PASSWORD", "")
		mongoAuthDB := getEnv("MONGODB_AUTH_DATABASE", "admin")

		if mongoUser != "" && mongoPassword != "" {
			mongoURI = "mongodb://" + mongoUser + ":" + mongoPassword + "@" + mongoHost + ":" + mongoPort + "/?authSource=" + mongoAuthDB
		} else {
			mongoURI = "mongodb://" + mongoHost + ":" + mongoPort
		}
	}

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisAddr := redisHost + ":" + redisPort

	return &Config{
		MongoURI:      mongoURI,
		DatabaseName:  getEnv("MONGODB_DBNAME", "product_service"),
		RedisAddr:     redisAddr,
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		Port:          getEnv("PRODUCT_SERVICE_APP_PORT", "3010"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func NewMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Configure client options
	clientOptions := options.Client().ApplyURI(uri)

	// Only set ServerAPI for mongodb+srv:// URIs (MongoDB Atlas)
	// Let the driver handle TLS automatically
	if len(uri) > 13 && uri[:13] == "mongodb+srv://" {
		clientOptions.SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

func NewRedis(addr, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
}
