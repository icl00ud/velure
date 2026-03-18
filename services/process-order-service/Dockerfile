FROM golang:1.25.5-alpine AS build

WORKDIR /app

COPY shared /shared

COPY services/process-order-service/go.mod services/process-order-service/go.sum ./

RUN go mod edit -replace github.com/icl00ud/velure-shared=/shared

RUN go mod download

COPY services/process-order-service/ .

RUN go build -o main ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main .

RUN addgroup -S app && adduser -S app -G app
USER app
EXPOSE 3040

CMD ["./main"]
