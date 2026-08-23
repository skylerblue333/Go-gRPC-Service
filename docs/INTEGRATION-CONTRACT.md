# Institutional Integration Contract

## Role

`Go-gRPC-Service` is the typed service boundary for internal RPC. Business contracts should be represented by versioned Protocol Buffer definitions and generated clients rather than ad-hoc JSON conventions.

## Integration sequence

`Py-Microservice-Gateway -> Go-Rate-Limiter -> Go-gRPC-Service -> domain services`

The service can consume resilience controls from `Rust-Circuit-Breaker` and publish health/telemetry signals to the platform observability layer.

## Production gates

- versioned `.proto` contracts;
- generated clients and compatibility checks;
- authenticated transport and mTLS where required;
- request deadlines and bounded payloads;
- health-aware discovery;
- structured tracing and metrics;
- integration tests against real service dependencies;
- load and failure testing;
- backward-compatible rollout policy.

The existing README correctly distinguishes the verified gRPC runtime/health/reflection surface from future business RPC contracts. This document preserves that evidence boundary.
