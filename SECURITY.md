# Sky RPC Core Security Model

Sky RPC Core is a reusable internal-service boundary for gRPC and HTTP operations. It is designed to run behind a trusted network or service mesh and supports an optional shared bearer token for application-layer authentication.

## Implemented controls

- Go 1.25.13 security baseline.
- `google.golang.org/grpc` 1.82.1.
- 1 MiB maximum gRPC send/receive message size.
- 1 MiB HTTP process-payload limit.
- bounded concurrent unary RPC execution.
- constant-time bearer-token comparison.
- duplicate/oversized request-id rejection.
- HTTP read/write/header/idle timeouts.
- graceful gRPC and HTTP shutdown.
- health/readiness endpoints that remain available without credentials.
- non-root distroless runtime image.
- CI gates for formatting, vet, tests, race detection, `govulncheck`, binary build, and container build.

## Trust boundaries and limitations

`RPC_AUTH_TOKEN` is a shared secret, not a substitute for workload identity. This release does **not** claim:

- mTLS or SPIFFE/SPIRE identity;
- OIDC/JWT validation;
- per-tenant RBAC/ABAC;
- distributed rate limiting;
- persistent job execution;
- queue semantics or exactly-once processing;
- TLS termination inside the process;
- production SLA/capacity figures;
- independent penetration testing.

For production service-to-service deployment, terminate or originate TLS with a trusted ingress/service mesh and prefer workload identity over long-lived bearer credentials.

## Secret handling

Never commit `RPC_AUTH_TOKEN`. Inject it through the deployment platform's secret mechanism. Rotate it when exposure is suspected and avoid logging authorization headers.
