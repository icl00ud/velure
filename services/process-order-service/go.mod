module github.com/icl00ud/velure/services/process-order-service

go 1.25.5

require (
	github.com/icl00ud/velure/shared v0.0.0
	github.com/joho/godotenv v1.5.1
	github.com/prometheus/client_golang v1.19.1
	github.com/rabbitmq/amqp091-go v1.10.0
	golang.org/x/sync v0.17.0
)

replace github.com/icl00ud/velure/shared => ../../shared

require (
	github.com/alicebob/miniredis/v2 v2.38.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/redis/go-redis/v9 v9.19.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
