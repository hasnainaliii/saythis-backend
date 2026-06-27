FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /bin/saythis ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="saythis-backend"
LABEL org.opencontainers.image.description="SayThis Go API server"
LABEL org.opencontainers.image.source="https://github.com/your-org/saythis-backend"

COPY --from=builder /bin/saythis /saythis

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/saythis"]
