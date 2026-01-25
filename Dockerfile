# Start from golang base image
FROM golang:1.24-alpine

# Install git and libc-compat
RUN apk add --no-cache git libc6-compat

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install Air for hot reloading
RUN go install github.com/air-verse/air@v1.52.2

# Copy the source code
COPY . .

# Expose port
EXPOSE 8000

# Default command for development
CMD ["air", "-c", ".air.toml"]
