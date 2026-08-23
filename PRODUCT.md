# Sky RPC Core

Sky RPC Core is an independently deployable Go service foundation for internal RPC workloads. It provides a hardened gRPC runtime plus a small HTTP operations surface for health, readiness, counters, and compatibility processing.

## Implemented

- gRPC health checking and server reflection
- bounded 1 MiB gRPC message sizes
- configurable concurrent unary-RPC ceiling
- request-metadata validation
- HTTP health, readiness, metrics, and process endpoints
- 1 MiB HTTP payload guard
- server read/write/header/idle timeouts
- graceful gRPC and HTTP shutdown with forced-stop fallback
- configurable listen addresses
- non-root distroless container
- CI gates for formatting, vet, tests, race detector, vulnerability scanning, binary build, and container build

## Product boundary

This repository is a service foundation, not a business-domain API. It does not claim authentication, authorization, TLS termination, durable storage, queues, distributed tracing, billing, tenant management, or an SLA until those capabilities are separately implemented and verified.

## Commercial use

Use it as a hardened starting runtime for independently deployed Go microservices, internal gRPC services, or SKYCOIN4444 backend components. Domain-specific protobuf APIs can be layered on top without coupling the runtime to the larger ecosystem.
