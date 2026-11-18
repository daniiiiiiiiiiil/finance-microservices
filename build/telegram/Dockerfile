FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache git ca-certificates
RUN git config --global http.lowSpeedLimit 0 && \
    git config --global http.lowSpeedTime 999999 && \
    git config --global http.postBuffer 524288000

ENV GOPROXY=https://proxy.golang.org|https://goproxy.io|direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

COPY telegram-bot-service/go.mod telegram-bot-service/go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    for i in 1 2 3 4 5; do \
        echo "Attempt $i to download dependencies..."; \
        go mod download && break || sleep 15; \
    done

COPY telegram-bot-service/ ./telegram-bot-service/

WORKDIR /app/telegram-bot-service

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o telegram-bot ./cmd/telegram/main.go

FROM alpine:3.20

RUN apk add --no-cache tzdata ca-certificates wget

WORKDIR /root/

COPY --from=builder /app/telegram-bot-service/telegram-bot .

EXPOSE 50062 8082 9100

HEALTHCHECK --interval=15s --timeout=5s --retries=5 CMD wget -qO- http://127.0.0.1:8082/healthz || exit 1

CMD ["./telegram-bot"]
