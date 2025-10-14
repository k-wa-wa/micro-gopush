# Stage 1: Builder
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY main.go .

# Build the application
# CGO_ENABLED=0 is important for static linking, making the binary self-contained
# -ldflags -s -w reduces the binary size by omitting debug information
RUN CGO_ENABLED=0 go build -o /micro-gopush main.go

# Stage 2: Runner
FROM scratch

# Set the working directory
WORKDIR /root/

# Copy the CA certificates from the builder stage (alpine)
# This is crucial for TLS verification in a scratch image
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary from the builder stage
COPY --from=builder /micro-gopush .

# Expose the port the application listens on
EXPOSE 8080

# Command to run the executable
CMD ["./micro-gopush"]
