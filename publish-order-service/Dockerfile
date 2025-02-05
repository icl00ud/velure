FROM golang:1.23.5-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/publisher/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/migrations ./migrations
COPY --from=build /app/main .

EXPOSE ${PUBLISH_ORDER_SERVICE_PORT}

CMD ["./main"]
