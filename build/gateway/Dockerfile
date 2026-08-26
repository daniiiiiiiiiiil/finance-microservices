FROM golang:1.26.4-alpine AS builder

ENV GOPROXY=https://goproxy.io,https://proxy.golang.org,direct
ENV GO111MODULE=on

WORKDIR /app

COPY gateway-service/go.mod gateway-service/go.sum ./
RUN go mod download

COPY gateway-service/ ./gateway-service/
COPY docs/ ./docs/

WORKDIR /app/gateway-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./cmd/gateway/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root

COPY --from=builder /app/gateway-service/gateway .
COPY --from=builder /app/docs ./docs

EXPOSE 8081

CMD ["./gateway"]