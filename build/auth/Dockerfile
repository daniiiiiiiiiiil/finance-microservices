FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY auth-service/go.mod auth-service/go.sum ./

RUN go mod download

COPY auth-service/ ./auth-service/

WORKDIR /app/auth-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth ./cmd/auth/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/auth-service/auth .

EXPOSE 50051 9091

CMD ["./auth"]