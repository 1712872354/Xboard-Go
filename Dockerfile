# ── Build Stage ──────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source (frontend dist should be pre-built and placed in internal/static/dist/)
COPY . .

# Build binary
ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o xboard-go ./cmd/server/

# ── Runtime Stage ────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /build/xboard-go /usr/local/bin/xboard-go

# Create data directory
RUN mkdir -p /data

EXPOSE 8080 50051

VOLUME ["/data"]

ENTRYPOINT ["xboard-go"]
CMD ["-config", "/data/config.yaml"]
