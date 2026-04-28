## Builder
FROM golang:alpine3.23 AS builder
WORKDIR /src

# Copy source code into container
COPY . .

# Build binary
RUN mkdir -p bin && GOOS=linux GOARCH=amd64 go build -o bin/registry-linux-amd64 cmd/registry/main.go

## App
FROM alpine:3.23
WORKDIR /app

# Copy seed file from repo
COPY cmd/registry/seed.csv .

# Copy build binary from builder container
COPY --from=builder /src/bin/registry-linux-amd64 /app/

EXPOSE 8080

# Set default container start command
ENTRYPOINT ["/app/registry-linux-amd64"]
