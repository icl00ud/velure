FROM golang:1.24.6-alpine AS build

WORKDIR /app

COPY internal/migrations ./internal/migrations

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/internal/migrations ./migrations

COPY --from=build /app/main .

RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 3030

CMD ["./main"]
