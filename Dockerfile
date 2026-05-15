# Multi-stage build
FROM golang:1.26-alpine AS builder

WORKDIR /app 
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o monitor ./cmd/monitor-service

# Runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/monitor .

EXPOSE 8080

CMD ["./monitor"]
