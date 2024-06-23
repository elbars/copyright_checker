# Use the official Golang image as the base image
FROM golang:1.22-alpine

# Set the working directory to /app
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o copyright-checker main.go

# Set the entrypoint to the built binary
ENTRYPOINT ["/app/copyright-checker"]
