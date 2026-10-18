# Start from the official Golang image for building
FROM golang:1.27-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o wut-temperature-exporter

# Use a minimal image for running
FROM alpine:latest@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/wut-temperature-exporter .

# Expose port (change if your app uses a different port)
EXPOSE 9191

VOLUME [ "/etc/wut-temperature-exporter" ]

# Run the app
ENTRYPOINT ["./wut-temperature-exporter"]