# Use a multi-stage build for a lean final image.

# 1. Build Stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
# Copy all files and download dependencies
COPY . .
RUN go mod download

# Build the application, statically linking it and disabling CGO
# This ensures the binary is self-contained and works in a minimal image.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app-server .

# 2. Final Stage
# Use a minimal base image that doesn't contain source code or build tools.
FROM alpine:latest
WORKDIR /app
# Copy only the compiled binary from the builder stage.
COPY --from=builder /app/app-server .

# Expose the port the server runs on.
EXPOSE 8080

# Define the command to run the application.
CMD ["./app-server"]
