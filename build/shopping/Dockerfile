FROM golang:1.26.4-alpine AS builder

RUN set -e; \
    for i in 1 2 3 4 5; do \
        apk add --no-cache git && exit 0; \
        echo "apk add failed, retry $i/5..."; \
        sleep 5; \
    done; \
    exit 1

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct

COPY shopping-list-service/go.mod shopping-list-service/go.sum ./

RUN set -e; \
    for i in 1 2 3 4 5; do \
        go mod download && exit 0; \
        echo "go mod download failed, retry $i/5..."; \
        sleep 5; \
    done; \
    exit 1

COPY shopping-list-service/ ./shopping-list-service/

WORKDIR /app/shopping-list-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shopping-service ./cmd/shopping/main.go

FROM alpine:latest

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /root/

COPY --from=builder /app/shopping-list-service/shopping-service .

EXPOSE 50060 9097

CMD ["./shopping-service"]
