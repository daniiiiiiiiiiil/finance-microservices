FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache git ca-certificates
RUN git config --global http.lowSpeedLimit 0 && \
    git config --global http.lowSpeedTime 999999 && \
    git config --global http.postBuffer 524288000

ENV GOPROXY=https://proxy.golang.org|https://goproxy.io|direct
ENV GOSUMDB=off
ENV GO111MODULE=on

WORKDIR /app

COPY admin-service/go.mod admin-service/go.sum ./

RUN for i in 1 2 3 4 5; do \
        echo "Attempt $i to download dependencies..."; \
        go mod download && break || sleep 15; \
    done

COPY admin-service/ ./admin-service/

WORKDIR /app/admin-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o admin ./cmd/admin/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/admin-service/admin .

EXPOSE 50054 9094

CMD ["./admin"]