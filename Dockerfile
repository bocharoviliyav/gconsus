# # Multi-stage Dockerfile for Git Analytics Backend

# FROM golang:1.24-bookworm AS build_base

# WORKDIR /build/src

# # Install build dependencies
# RUN apt-get update && apt-get install -y --no-install-recommends \
#     ca-certificates \
#     git \
#     && rm -rf /var/lib/apt/lists/*

# # Change to backend directory
# WORKDIR /build/src/backend

# # Cache dependencies
# COPY backend/go.mod backend/go.sum ./
# RUN go mod download

# # Copy backend source code
# COPY backend/ .

# # Build binary with version info
# ARG VERSION=dev
# ARG BUILD_TIME
# ARG GIT_COMMIT
# RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
#     -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
#     -o app .

# # Production stage
# FROM alpine:3.19 AS production

# # Install runtime dependencies
# RUN apk add --no-cache \
#     ca-certificates \
#     tzdata \
#     wget \
#     curl \
#     && addgroup -g 1000 appuser \
#     && adduser -D -u 1000 -G appuser appuser

# # Copy binary
# COPY --from=build_base /build/src/backend/app /bin/app

# # Copy migrations (if backend has its own migrations)
# COPY --chown=appuser:appuser migrations /migrations

# WORKDIR /app

# USER appuser

# # Health check
# HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# EXPOSE 8000

# ENTRYPOINT ["/bin/app"]
