# Go gRPC Service

A production-oriented Go service boundary with a real gRPC runtime, standard gRPC health checking, server reflection for development tooling, graceful shutdown, and an HTTP compatibility surface.

> **SkyCoin4444 / IITR infrastructure component:** intended as a reusable service boundary for typed internal APIs, protocol services, and gateway-to-backend communication.

## Runtime surfaces

| Surface | Address | Purpose |
|---|---|---|
| gRPC | `:9090` | Native service-to-service transport, health, reflection |
| HTTP | `:8080` | Compatibility/diagnostic endpoints |
| HTTP health | `GET /health` | Service status and processed counter |
| HTTP process | `POST /process` | Example accepted-work path |

The gRPC runtime currently exposes the standard gRPC health service and reflection. Business RPC contracts should be added as `.proto` APIs rather than pretending the existing HTTP handlers are gRPC methods.

## Quick start

```bash
go mod download
go test -race ./...
go vet ./...
go run .
```

The service shuts down gracefully on `SIGINT` and `SIGTERM`.

## Architecture

```text
                    +--------------------+
                    |   SkyCoin / Client |
                    +---------+----------+
                              |
                +-------------+-------------+
                |                           |
             gRPC :9090                 HTTP :8080
                |                           |
        +-------v--------+           +------v-------+
        | gRPC health +  |           | compatibility |
        | reflection     |           | handlers      |
        +-------+--------+           +------+--------+
                |                           |
                +-------------+-------------+
                              |
                         ServiceState
                              |
                         graceful stop
```

## Why this is materially stronger

The repository now contains an actual gRPC server rather than only an HTTP server with a gRPC-themed README. It uses the maintained `google.golang.org/grpc` implementation, standard health semantics, reflection, bounded HTTP timeouts, synchronized state access, and graceful shutdown.

## Institutional integration path

The service is suitable as a foundation for:

1. typed internal RPC contracts;
2. gateway-to-service communication;
3. health-aware service discovery;
4. load-balancer backends;
5. SkyCoin protocol adapters;
6. telemetry and SLO instrumentation;
7. tenant-aware service quotas;
8. authenticated/mTLS service transport;
9. SDK-generated client libraries;
10. managed enterprise service deployments;
11. support, maintenance, and integration services.

These are product/value surfaces, not claims of current revenue or customer adoption.

## Verification

GitHub Actions runs formatting checks, `go vet`, race-detector tests, module download, and `govulncheck` on every push and pull request.
