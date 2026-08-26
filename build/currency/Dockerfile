FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY currency-service/go.mod currency-service/go.sum ./

RUN go mod download

COPY currency-service/ ./currency-service/

WORKDIR /app/currency-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o currency ./cmd/currency/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/currency-service/currency .

EXPOSE 50055 9095

CMD ["./currency"]