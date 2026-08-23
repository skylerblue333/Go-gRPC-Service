# Product Definition — Sky RPC Core

Sky RPC Core is an independently deployable Go service foundation for internal RPC workloads. It provides a hardened gRPC runtime plus a small HTTP operations surface for health, readiness, counters, and compatibility processing.

## Implemented

- Go 1.25.13 security baseline and maintained `google.golang.org/grpc` runtime;
- standard gRPC health checking and server reflection;
- bounded 1 MiB gRPC send/receive message sizes;
- configurable concurrent unary-RPC ceiling;
- optional shared bearer authentication with constant-time comparison;
- request-metadata validation and rejected-request accounting;
- HTTP health, readiness, metrics, and process endpoints;
- 1 MiB HTTP process-payload guard;
- server read/write/header/idle timeouts;
- graceful gRPC and HTTP shutdown with forced-stop fallback;
- configurable listen addresses;
- non-root distroless container;
- CI gates for formatting, vet, tests, race detector, vulnerability scanning, binary build, and container build.

## Product boundary

This repository is a **service-runtime foundation**, not a fabricated business-domain API. Teams should add explicit versioned `.proto` contracts and generated clients for their own workloads.

It does not claim built-in mTLS, SPIFFE/SPIRE workload identity, OIDC/JWT validation, per-tenant RBAC, durable queues, persistent job execution, distributed tracing, billing, tenant management, production capacity figures, or an SLA until those capabilities are separately implemented and verified.

## Commercial packaging

Use it as:

1. a standalone private RPC service runtime;
2. a distroless OCI container;
3. a base for generated protobuf APIs and SDKs;
4. a reusable SKYCOIN4444 service boundary;
5. a foundation for paid integration work around mTLS, telemetry, service discovery, and deployment automation.
