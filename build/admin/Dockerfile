FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY admin-service/go.mod admin-service/go.sum ./

RUN go mod download

COPY admin-service/ ./admin-service/

WORKDIR /app/admin-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o admin ./cmd/admin/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/admin-service/admin .

EXPOSE 50054 9094

CMD ["./admin"]