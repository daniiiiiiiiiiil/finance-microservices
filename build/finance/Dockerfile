FROM golang:1.26.4-alpine AS builder

RUN apk add --no-cache ttf-dejavu

WORKDIR /app

ENV GOPROXY=direct
ENV GOSUMDB=off

COPY finance-service/go.mod finance-service/go.sum ./
RUN go mod download

COPY finance-service/ ./finance-service/

WORKDIR /app/finance-service

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o finance ./cmd/finance/main.go

FROM alpine:latest

RUN apk add --no-cache ttf-dejavu ca-certificates

WORKDIR /root/

COPY --from=builder /app/finance-service/finance .

EXPOSE 50053 9093

CMD ["./finance"]