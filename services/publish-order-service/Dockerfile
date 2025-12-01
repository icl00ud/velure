FROM golang:1.24.6-alpine AS build

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

# Copy shared module first
COPY shared /shared

# Copy go.mod and go.sum
COPY services/publish-order-service/go.mod services/publish-order-service/go.sum ./

# Update replace directive to use absolute path in container
RUN go mod edit -replace github.com/icl00ud/velure-shared=/shared

RUN go mod download

# Copy source code
COPY services/publish-order-service/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o publish-order-service ./main.go

FROM scratch

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/publish-order-service ./publish-order-service
COPY --from=build /build/internal/migrations ./internal/migrations

USER 1000:1000
EXPOSE 8080

ENTRYPOINT ["/app/publish-order-service"]
