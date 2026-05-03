# Dockerfile for inv-collector
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /inv-collector ./cmd/inv-collector

FROM alpine:latest
WORKDIR /
COPY --from=builder /inv-collector /inv-collector
COPY config.yaml /config.yaml

EXPOSE 8088
CMD ["/inv-collector", "-config", "/config.yaml"]
