# Use Go 1.23 bookworm as base image
FROM golang:1.23-bookworm AS base

# Move to working directory /app
WORKDIR /app

# Copy the go.mod and go.sum files to the /build directory
COPY . .

# Install dependencies
RUN go mod download

# Build the application
RUN go mod tidy

RUN go build -o binary ./cmd/see-containers

# Set the entry point of the container to the binary
ENTRYPOINT ["/app/binary"]
