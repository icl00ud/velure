# Build stage
FROM golang:1.24.6-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared module first
COPY shared /shared

# Copy service files
COPY services/auth-service/go.mod services/auth-service/go.sum ./

# Update replace directive to use absolute path in container
RUN go mod edit -replace github.com/icl00ud/velure-shared=/shared

# Download dependencies
RUN go mod download

# Copy source code
COPY services/auth-service/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy wait-for-it script if it exists
COPY services/auth-service/wait-for-it.sh .
RUN chmod +x wait-for-it.sh

# Expose port
RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 3020

# Command to run
CMD ["./main"]
