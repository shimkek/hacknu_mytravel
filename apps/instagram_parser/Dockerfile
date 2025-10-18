# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY ../../go.mod ../../go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project (for shared internal packages)
COPY ../.. .

# Build the application
WORKDIR /app/apps/instagram_parser
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/apps/instagram_parser/main .

# Make sure the binary is executable
RUN chmod +x ./main

CMD ["./main"]