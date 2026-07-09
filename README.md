# car-bridge-golang

Go **integration bridge** that sits in front of third-party transportation, telematics,
geo, finance and comms APIs and exposes clean HTTP endpoints for the `dot-portal` product.

Layered (clean-architecture) following
[khannedy/golang-clean-architecture](https://github.com/khannedy/golang-clean-architecture),
adapted to: **Fiber + Postgres (pgx, no ORM) + Viper + Logrus + validator**.

## Layout

```
cmd/web/              entrypoint
internal/
  config/             wiring (viper, logrus, validator, postgres, redis, fiber, Bootstrap)
  delivery/http/      controllers, middleware (incl. rate limiter), routes (Fiber)
  service/            business logic — one service per entity (depends on integration ports + repositories)
  repository/         pgx raw-SQL data access
  entity/             DB row structs
  model/              request/response DTOs (+ converter/)
  integrations/       <-- third-party API clients, split by TYPE, one package per provider
db/migrations/        golang-migrate SQL
```

### integrations

Each *type* folder declares capability **ports** (interfaces); each *provider* is its own
package implementing them. Business logic depends on the port, never the vendor.

| type | ports | providers |
|------|-------|-----------|
| `vehicle` | VINDecoder, CarrierLookup, PlateDecoder | transgov, saferweb, carsxe, vehicledatabases |
| `telematics` | ConnectedVehicle | smartcar |
| `geo` | Geocoder, RouteMatrix, FuelStationFinder | mapbox, googlemaps, nrel |
| `finance` | BankVerifier, IdentityVerifier, LoanCalculator | plaid, kyc, apr |
| `comms` | SMSSender, EmailSender | twilio, postmark, sendgrid |

**Adding a provider:** add a config block, implement `New(cfg, base)` + the port method(s)
in the provider package, wire it in `internal/config/app.go`. No business-logic changes.

## HTTP Endpoints

Inbound routes exposed by this service (registered in
`internal/delivery/http/route/route.go`). Everything under `internal/integrations/`
is an *outbound* third-party client, not an endpoint. All routes pass through the
global middleware chain: `recover` → `requestid` → request logger → per-IP rate limiter.

| Method | Path                 | Handler                      | Description |
|--------|----------------------|------------------------------|-------------|
| `GET`  | `/health`            | `HealthController.Check`     | Liveness/readiness probe; reports `db` and `redis` status. Returns `503` when degraded. |
| `GET`  | `/api/carriers/:dot` | `CarrierController.GetByDOT` | Carrier lookup by USDOT number (via saferweb). Returns `501` until the provider is implemented. |

## Run

```bash
cp .env.example .env          # fill secrets as needed
make db-up                    # start Postgres + Redis (docker compose)
make migrate-up               # apply migrations
make run                      # start the server on :3000

curl localhost:3000/health                # {"db":"ok","redis":"ok","status":"ok"}
curl localhost:3000/api/carriers/123456   # 501 until saferweb is implemented
```

Inbound requests are rate-limited per client IP (`ratelimit.max` / `window_seconds`
in config.yaml), Redis-backed when `redis.url` is set and in-memory otherwise.
