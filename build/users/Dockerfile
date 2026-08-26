FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY users-service/go.mod users-service/go.sum ./

RUN go mod download

COPY users-service/ ./users-service/

WORKDIR /app/users-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o users ./cmd/users/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/users-service/users .

EXPOSE 50052 9092

CMD ["./users"]