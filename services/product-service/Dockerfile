FROM golang:1.25.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY shared /shared

COPY services/product-service/go.mod services/product-service/go.sum ./

RUN go mod edit -replace github.com/icl00ud/velure-shared=/shared

RUN go mod download

COPY services/product-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .

RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 3010

CMD ["./main"]
