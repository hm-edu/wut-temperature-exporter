# Start from the official Golang image for building
FROM golang:1.27-alpine@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o wut-temperature-exporter

# Use a minimal image for running
FROM alpine:latest@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/wut-temperature-exporter .

# Expose port (change if your app uses a different port)
EXPOSE 9191

VOLUME [ "/etc/wut-temperature-exporter" ]

# Run the app
ENTRYPOINT ["./wut-temperature-exporter"]