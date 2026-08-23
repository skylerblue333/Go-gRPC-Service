# Architecture — Go gRPC Service

## Service boundary

The service provides two transport surfaces:

- gRPC on `:9090` for service-to-service communication.
- HTTP on `:8080` for compatibility and diagnostics.

The gRPC server registers the standard gRPC health service and reflection. Reflection is useful for development and inspection; production deployments should explicitly decide whether reflection remains enabled outside trusted networks.

## Lifecycle

```text
process start
    |
    +--> bind gRPC :9090
    +--> register health service
    +--> register reflection
    +--> bind HTTP :8080
    |
    v
 serving
    |
 SIGINT/SIGTERM
    |
    +--> gRPC GracefulStop
    +--> HTTP Shutdown(timeout)
    +--> close listener
    |
    v
 stopped
```

## State safety

`ServiceState` uses `sync.RWMutex`. Health reads acquire a read lock; process updates acquire a write lock. Tests read state under the same lock rather than racing the service state.

## Network safety

HTTP `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` provide bounded request lifecycles. The gRPC server is intentionally kept behind the repository's deployment boundary so production operators can add TLS/mTLS, authentication, quotas, and network policy without changing the core service lifecycle.

## Protocol evolution

Business RPCs should be introduced through versioned Protocol Buffer definitions. Generated clients and servers should become the compatibility contract between this service and Go/Rust/TypeScript/Python consumers. The current health API is the initial executable proof that the gRPC runtime is real and inspectable.

## Verification

GitHub Actions is the repository verification path and checks formatting, compilation/tests with the race detector, static analysis, and known Go dependency vulnerabilities.
