# Sky RPC Core

Sky RPC Core is a compact Go service-runtime product for internal RPC workloads. It provides a real `google.golang.org/grpc` server, standard gRPC health semantics, reflection for development tooling, a small HTTP operations surface, optional bearer authentication, bounded concurrency, graceful shutdown, and hardened container packaging.

> **Status:** productization branch under exact-head CI verification. It is a service foundation, not a fabricated business-domain API.

## Runtime surfaces

| Surface | Default address | Purpose |
|---|---:|---|
| gRPC | `:9090` | native service-to-service transport, health, reflection |
| HTTP | `:8080` | health/readiness, metrics, compatibility processing |
| HTTP health | `GET /healthz` | liveness |
| HTTP readiness | `GET /readyz` | readiness |
| HTTP metrics | `GET /metrics` | processed/rejected/uptime counters |
| HTTP process | `POST /v1/process` | bounded example work-acceptance surface |

The gRPC runtime currently exposes the standard health service and reflection. Real application RPCs should be defined as versioned `.proto` contracts rather than pretending the example HTTP handler is a business gRPC method.

## Security controls

- Go 1.25.13 baseline;
- `google.golang.org/grpc` 1.82.1;
- 1 MiB maximum gRPC send/receive size;
- 1 MiB HTTP process-payload limit;
- configurable maximum concurrent unary RPCs;
- optional constant-time bearer authentication through `RPC_AUTH_TOKEN`;
- duplicate/oversized `x-request-id` rejection;
- HTTP read/write/header/idle timeouts;
- health/readiness remain available for orchestration without credentials;
- graceful HTTP and gRPC shutdown;
- non-root distroless runtime image.

See `SECURITY.md` for the exact threat model and limitations.

## Quick start

```bash
go mod download
go test -race ./...
go vet ./...
export RPC_AUTH_TOKEN="$(openssl rand -hex 32)"
go run .
```

Check health:

```bash
curl http://127.0.0.1:8080/healthz
```

Authenticated HTTP operation:

```bash
curl -i \
  -H "Authorization: Bearer $RPC_AUTH_TOKEN" \
  -X POST \
  http://127.0.0.1:8080/v1/process
```

The service also includes a built-in container health check:

```bash
./sky-rpc-core --healthcheck
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `GRPC_ADDR` | `:9090` | gRPC listen address |
| `HTTP_ADDR` | `:8080` | HTTP operations listen address |
| `MAX_CONCURRENT_RPCS` | `256` | unary-RPC ceiling, range `1..100000` |
| `RPC_AUTH_TOKEN` | unset | bearer token for non-health HTTP/gRPC surfaces |

The process warns if the auth token is unset. The packaged Compose deployment requires it.

## Container

```bash
export RPC_AUTH_TOKEN="$(openssl rand -hex 32)"
docker compose up --build
```

The Compose package binds both ports to localhost by default, runs a read-only root filesystem, drops Linux capabilities, enables `no-new-privileges`, and uses the built-in `/healthz` probe.

## Architecture

```text
                         clients/services
                               |
               +---------------+---------------+
               |                               |
            gRPC :9090                       HTTP :8080
               |                               |
      auth + concurrency guard          auth + payload guard
               |                               |
      gRPC health/reflection       health/readiness/metrics/process
               |                               |
               +---------------+---------------+
                               |
                         ServiceState
                               |
                         graceful stop
```

## Verification

GitHub Actions is configured to run:

- module resolution and verification;
- `gofmt` enforcement;
- `go vet`;
- unit/behavior tests;
- race-detector tests;
- `govulncheck`;
- binary build;
- Docker image build.

A branch is not treated as complete until the exact PR head passes those gates.

## Product boundary

Sky RPC Core does **not** currently claim mTLS/SPIFFE identity, OIDC/JWT verification, tenant RBAC/ABAC, persistent queues, distributed tracing, exactly-once execution, production capacity figures, or an SLA. See `PRODUCT.md` for commercial extension paths.
