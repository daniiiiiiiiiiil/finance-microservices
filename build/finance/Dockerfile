FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache git ca-certificates
RUN git config --global http.lowSpeedLimit 0 && \
    git config --global http.lowSpeedTime 999999 && \
    git config --global http.postBuffer 524288000

ENV GOPROXY=https://proxy.golang.org|https://goproxy.io|direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

COPY finance-service/go.mod finance-service/go.sum ./

RUN for i in 1 2 3 4 5; do \
        echo "Attempt $i to download dependencies..."; \
        go mod download && break || sleep 15; \
    done

COPY finance-service/ ./finance-service/

WORKDIR /app/finance-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o finance ./cmd/finance/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/finance-service/finance .

EXPOSE 50053 9093

CMD ["./finance"]