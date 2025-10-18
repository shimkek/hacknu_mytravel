# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install dependencies for Selenium and ChromeDriver
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    chromium \
    chromium-chromedriver \
    curl \
    wget \
    xvfb \
    && rm -rf /var/cache/apk/*

# Set up Chrome environment
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    CHROMEDRIVER_PATH=/usr/bin/chromedriver

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Make sure the binary is executable
RUN chmod +x ./main

# Create a script to run with virtual display
RUN echo '#!/bin/sh' > /root/run.sh && \
    echo 'Xvfb :99 -screen 0 1024x768x24 > /dev/null 2>&1 &' >> /root/run.sh && \
    echo 'export DISPLAY=:99' >> /root/run.sh && \
    echo 'exec ./main "$@"' >> /root/run.sh && \
    chmod +x /root/run.sh

CMD ["/root/run.sh"]