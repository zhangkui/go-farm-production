# syntax=docker/dockerfile:1
# ---- Build stage ----
FROM golang:1.22-bookworm AS builder
WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static-ish binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Runtime stage ----
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/server /app/server

# Migrations are embedded in the binary; no extra copy needed.
ENV APP_ENV=prod \
    HTTP_ADDR=:8080
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -fsS http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/server"]
