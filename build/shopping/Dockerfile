FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache git ca-certificates
RUN git config --global http.lowSpeedLimit 0 && \
    git config --global http.lowSpeedTime 999999 && \
    git config --global http.postBuffer 524288000

ENV GOPROXY=https://proxy.golang.org|https://goproxy.io|direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

COPY shopping-list-service/go.mod shopping-list-service/go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    for i in 1 2 3 4 5; do \
        echo "Attempt $i to download dependencies..."; \
        go mod download && break || sleep 15; \
    done

COPY shopping-list-service/ ./shopping-list-service/

WORKDIR /app/shopping-list-service

RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shopping-list ./cmd/shopping/main.go

FROM alpine:3.18

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /root/

COPY --from=builder /app/shopping-list-service/shopping-list .

EXPOSE 50056 9096

CMD ["./shopping-list"]
