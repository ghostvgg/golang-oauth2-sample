# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o oauth-demo main.go

# Runtime stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/oauth-demo /usr/local/bin/oauth-demo

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/oauth-demo"]
