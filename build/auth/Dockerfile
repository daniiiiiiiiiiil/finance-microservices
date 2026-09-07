FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache git ca-certificates
RUN git config --global http.lowSpeedLimit 0 && \
    git config --global http.lowSpeedTime 999999 && \
    git config --global http.postBuffer 524288000

ENV GOPROXY=https://proxy.golang.org|https://goproxy.io|direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

COPY auth-service/go.mod auth-service/go.sum ./

RUN for i in 1 2 3 4 5; do \
        echo "Attempt $i to download dependencies..."; \
        go mod download && break || sleep 15; \
    done

COPY auth-service/ ./auth-service/

WORKDIR /app/auth-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth ./cmd/auth/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/auth-service/auth .

EXPOSE 50051 9091

CMD ["./auth"]