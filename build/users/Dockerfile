FROM golang:1.26.4-alpine AS builder

ENV GOPROXY=https://proxy.golang.org|https://goproxy.io|direct
ENV GOSUMDB=off

WORKDIR /app

COPY users-service/go.mod users-service/go.sum ./

RUN for i in 1 2 3 4 5; do \
        echo "Attempt $i to download dependencies..."; \
        go mod download && break || sleep 15; \
    done

COPY users-service/ ./users-service/

WORKDIR /app/users-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o users ./cmd/users/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/users-service/users .

EXPOSE 50052 9092

CMD ["./users"]