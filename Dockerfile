# Use the official Golang image as a builder
FROM golang:1.19 AS builder

# Set the working directory inside the container
WORKDIR ./

# Copy the Go files to the container
COPY . .

# Download dependencies (if any)
RUN go mod init golang-ports-services 2>/dev/null || true
RUN go mod tidy

# Build the Go application
RUN go build -o golang-ports-services

# Use a minimal base image for the final container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /golang-ports-services .

# Set executable permissions
RUN chmod +x ./golang-ports-services

EXPOSE 8080

# Command to run the application
CMD ["./golang-ports-services"]