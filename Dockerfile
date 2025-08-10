# ---- Build Stage ----
FROM golang:1.24.2 AS builder

# Set environment variables for cross-platform builds
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# Copy only required files and folders
COPY app/ ./app/
COPY config/ ./config/
COPY common/ ./common/
COPY go.mod .
COPY go.sum .
COPY main.go .

# Download dependencies
RUN go mod download

# Build the application
RUN go build -o server main.go

# ---- Final Stage ----
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy only the compiled binary
COPY --from=builder /app/server .

# Run the binary
CMD ["./server"]

EXPOSE 8080
