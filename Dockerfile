FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o planguard ./cmd/planguard

# Create final lightweight image
FROM alpine:latest

RUN apk --no-cache add ca-certificates git

WORKDIR /planguard

# Copy binary from builder
COPY --from=builder /build/planguard /usr/local/bin/planguard

# Copy default rules
COPY rules /planguard/rules

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/planguard"]
