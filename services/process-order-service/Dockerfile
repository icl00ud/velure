FROM golang:1.24.6-alpine AS build

WORKDIR /app

# Copy shared module first
COPY shared /shared

# Copy go.mod and go.sum
COPY services/process-order-service/go.mod services/process-order-service/go.sum ./

# Update replace directive to use absolute path in container
RUN go mod edit -replace github.com/icl00ud/velure-shared=/shared

RUN go mod download

# Copy source code
COPY services/process-order-service/ .

RUN go build -o main ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main .

RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 3040

CMD ["./main"]
