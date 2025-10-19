# Start from a Go base image
FROM golang:1.24-alpine as builder

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the rabbitmq consumer
RUN CGO_ENABLED=0 GOOS=linux go build -o /rabbitmq_consumer ./rabbitmq_consumer

# Production image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /rabbitmq_consumer .
CMD ["./rabbitmq_consumer"]
