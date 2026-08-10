FROM golang:1.26.4-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/gateway ./cmd/gateway

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/gateway /app/gateway
CMD ["/app/gateway"]