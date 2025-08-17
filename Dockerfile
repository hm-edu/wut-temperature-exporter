# Start from the official Golang image for building
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o wut-temperature-exporter

# Use a minimal image for running
FROM alpine:latest

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/wut-temperature-exporter .

# Expose port (change if your app uses a different port)
EXPOSE 9191

VOLUME [ "/etc/wut-temperature-exporter" ]

# Run the app
ENTRYPOINT ["./wut-temperature-exporter"]