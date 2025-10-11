# --- Stage 1: Build ---
# Use the official Golang image as the builder.
# Using a specific version is good practice for reproducibility.
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies.
COPY go.mod go.sum ./
# Download dependencies. This is done as a separate step to leverage Docker's layer caching.
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build the Go application.
# -o /bin/server: specifies the output file name and location.
# CGO_ENABLED=0: disables CGO to create a statically linked binary.
# GOOS=linux: ensures the binary is built for a Linux environment (for the final image).
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server .

# --- Stage 2: Final Image ---
# Use a minimal, non-root base image for the final container.
# Alpine is a good choice for its small size.
FROM alpine:latest

# Copy the compiled binary from the builder stage.
COPY --from=builder /bin/server /bin/server

# Expose port 8080 to the outside world.
EXPOSE 8080

# Define the command to run when the container starts.
CMD ["/bin/server"]
