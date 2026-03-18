FROM golang:1.25.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY shared /shared

COPY services/auth-service/go.mod services/auth-service/go.sum ./

RUN go mod edit -replace github.com/icl00ud/velure-shared=/shared

RUN go mod download

COPY services/auth-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .

COPY services/auth-service/wait-for-it.sh .
RUN chmod +x wait-for-it.sh

RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 3020

CMD ["./main"]
