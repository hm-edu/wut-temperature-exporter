# Start from the official Golang image for building
FROM golang:1.27-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o wut-temperature-exporter

# Use a minimal image for running
FROM alpine:latest@sha256:3f4d6b32e26703c6082839234f43769d5046b3ac7c910033911d1e71d8cd630a

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/wut-temperature-exporter .

# Expose port (change if your app uses a different port)
EXPOSE 9191

VOLUME [ "/etc/wut-temperature-exporter" ]

# Run the app
ENTRYPOINT ["./wut-temperature-exporter"]