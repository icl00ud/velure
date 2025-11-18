FROM golang:1.24.6-alpine AS build

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o publish-order-service ./main.go

FROM scratch

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/publish-order-service ./publish-order-service
COPY --from=build /build/internal/migrations ./internal/migrations

USER 1000:1000
EXPOSE 8080

ENTRYPOINT ["/app/publish-order-service"]
