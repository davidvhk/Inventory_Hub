# Dockerfile for inv-consumer
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /inv-consumer ./cmd/inv-consumer

FROM alpine:latest
WORKDIR /
COPY --from=builder /inv-consumer /inv-consumer

CMD ["/inv-consumer"]
