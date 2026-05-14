# =============================================================================
# Stage 1 — Build
# Build a fully static, stripped binary. No CGO so the final image needs
# no libc at all, which lets us use distroless/static as the runtime base.
# =============================================================================
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Resolve dependencies before copying source to maximise Docker layer-cache hits.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the entire source tree.
COPY . .

# -ldflags "-s -w"  → strip debug symbols and DWARF info (shrinks binary ~30%)
# -trimpath         → remove local file-system paths from the binary
# CGO_ENABLED=0     → statically link everything; no shared-library dependency
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /bin/saythis ./cmd/api

# =============================================================================
# Stage 2 — Runtime
# distroless/static contains only ca-certificates and tzdata — nothing else.
# No shell, no package manager, minimal CVE surface.
# nonroot tag enforces a non-root UID (65532) at the image level.
# =============================================================================
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="saythis-backend"
LABEL org.opencontainers.image.description="SayThis Go API server"
LABEL org.opencontainers.image.source="https://github.com/your-org/saythis-backend"

# Pull in the compiled binary from the builder stage.
COPY --from=builder /bin/saythis /saythis

# distroless/nonroot already sets USER nonroot:nonroot (UID/GID 65532).
# Declaring it here makes the intent explicit and visible in `docker inspect`.
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/saythis"]
