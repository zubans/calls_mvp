# Build stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o video-call-server .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Create recordings directory
RUN mkdir -p /root/recordings

# Copy the binary from builder stage
COPY --from=builder /app/video-call-server .

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./video-call-server"]