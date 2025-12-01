# Dockerfile para Go
FROM golang:1.24.6-alpine AS builder

WORKDIR /app

# Instalar dependências do sistema
RUN apk add --no-cache git

# Copy shared module first
COPY shared /shared

# Copiar go.mod e go.sum
COPY services/product-service/go.mod services/product-service/go.sum ./

# Update replace directive to use absolute path in container
RUN go mod edit -replace github.com/icl00ud/velure-shared=/shared

# Download das dependências
RUN go mod download

# Copiar código fonte
COPY services/product-service/ .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Imagem final
FROM alpine:latest

# Instalar ca-certificates para conexões HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copiar o binário da etapa de build
COPY --from=builder /app/main .

# Expor a porta
RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 3010

# Comando para executar a aplicação
CMD ["./main"]
