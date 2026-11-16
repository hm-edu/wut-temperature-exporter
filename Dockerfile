# Start from the official Golang image for building
FROM golang:1.27-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o wut-temperature-exporter

# Use a minimal image for running
FROM alpine:latest@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/wut-temperature-exporter .

# Expose port (change if your app uses a different port)
EXPOSE 9191

VOLUME [ "/etc/wut-temperature-exporter" ]

# Run the app
ENTRYPOINT ["./wut-temperature-exporter"]