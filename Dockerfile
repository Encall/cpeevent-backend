# Use the official Golang image as the base image
FROM golang:1.23 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 go build -o main main.go

# Use a minimal base image to run the Go app
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Ensure the binary has execute permissions
RUN chmod +x ./main

# Expose the port the app runs on
EXPOSE 8080

LABEL name="CPEEVO-GO-Backend"

# Command to run the executable
CMD ["./main"]