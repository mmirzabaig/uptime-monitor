# Build stage
FROM golang:1.25 AS builder

WORKDIR /app

# Copy dependency files first
COPY go.mod go.sum ./

RUN go mod download

# Copy application source
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o uptime-monitor ./cmd/server


# Runtime stage
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/uptime-monitor .

EXPOSE 8080

CMD ["./uptime-monitor"]