FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache git ca-certificates

RUN git config --global http.lowSpeedLimit 0 && \
    git config --global http.lowSpeedTime 999999 && \
    git config --global http.postBuffer 524288000

ENV GOPROXY=https://proxy.golang.org|https://goproxy.io|direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

COPY currency-service/go.mod currency-service/go.sum ./

RUN for i in 1 2 3 4 5; do \
        echo "Attempt $i to download dependencies..."; \
        go mod download && break || sleep 15; \
    done

COPY currency-service/ ./currency-service/

WORKDIR /app/currency-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o currency ./cmd/currency/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/currency-service/currency .

EXPOSE 50052 9092

CMD ["./currency"]