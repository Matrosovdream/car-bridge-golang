# ---- build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache module downloads separately from source for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static binary (CGO off) so it runs on a minimal base.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/car-bridge ./cmd/web

# ---- runtime stage ----
FROM alpine:3.20
# ca-certificates: required for outbound HTTPS calls to the integration APIs.
RUN apk add --no-cache ca-certificates wget && adduser -D -u 10001 app
WORKDIR /app

COPY --from=build /out/car-bridge /app/car-bridge
COPY config.yaml /app/config.yaml
COPY db/migrations /app/db/migrations

USER app
EXPOSE 3000

# Note: /health returns 503 when DB/Redis are down, so the container is reported
# unhealthy during a dependency outage. start-period covers slow dependency boot.
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/health >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/car-bridge"]

    

