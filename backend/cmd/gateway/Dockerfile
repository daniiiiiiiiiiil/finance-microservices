FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o gateway ./cmd/gateway/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gateway .
COPY --from=builder /app/docs ./docs

EXPOSE 8081

CMD ["./gateway"]