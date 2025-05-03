# --- Builder stage ---
FROM golang:1.23-bullseye AS builder

WORKDIR /app

# Install build tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends git gcc build-essential && \
    rm -rf /var/lib/apt/lists/*

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o mcp-server-splunk ./cmd/mcp-server-splunk/main.go

# --- Final image ---
FROM debian:bullseye-slim

ENV DEBIAN_FRONTEND=noninteractive

# Install minimal runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends tzdata ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    ln -fs /usr/share/zoneinfo/Etc/UTC /etc/localtime && \
    echo "Etc/UTC" > /etc/timezone && \
    groupadd --system mcp && \
    useradd --system --create-home --home-dir /home/mcp --gid mcp mcp

ENV TZ=Etc/UTC
ENV HOME=/home/mcp

WORKDIR /home/mcp

# Copy the built binary
COPY --from=builder /app/mcp-server-splunk /usr/local/bin/mcp-server-splunk
# Copy the resource file
COPY resources/data-dictionary.csv /home/mcp/resources/data-dictionary.csv

# Set permissions
RUN chown -R mcp:mcp /home/mcp

USER mcp

ENTRYPOINT ["/usr/local/bin/mcp-server-splunk"]