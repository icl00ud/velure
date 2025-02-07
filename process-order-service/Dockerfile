FROM golang:1.23.5-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/migrations ./migrations
COPY --from=build /app/main .

EXPOSE ${PROCESS_ORDER_SERVICE_APP_PORT}

CMD ["./main"]
