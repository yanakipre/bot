# Internal Shared Packages

This directory contains shared packages used across various applications within the `yanakipre/bot` project (e.g., `app/telegramsearch`, `app/cyprusapis`). The goal is to provide consistent, reusable infrastructure for common concerns.

## Key Areas:

*   **Application Lifecycle (`application`, `readiness`, `closer`):** Provides a standard structure for bootstrapping services, managing component start/stop, and performing readiness checks.
*   **Configuration (`config`, `secret`, `encodingtooling`):** Handles loading configuration from files and environment variables (`confita`), manages sensitive data securely (`secret`), and provides custom types for easier config definition (`encodingtooling`).
*   **Logging (`logger`, `logfmt`, `clouderr`):** Centralized logging based on `zap`, featuring context propagation, structured logging via `clouderr`, configurable filtering, and a tool (`logfmt`) to enforce key naming conventions.
*   **Error Handling (`semerr`, `status`, `codeerr`, `clouderr`, `recoverytooling`):** A layered system for handling errors:
    *   `semerr`: Defines semantic error types (e.g., NotFound, Internal) with stack traces.
    *   `status`: Provides canonical status codes and details for API responses, bridging internal errors.
    *   `codeerr`: Application-specific error codes for APIs.
    *   `clouderr`: Attaches structured `zap.Field` data to errors.
    *   `recoverytooling`: Utilities for panic recovery in goroutines.
*   **Observability (`metrics`, `promtooling`, `client/otlp`, `sentrytooling`):**
    *   `metrics`, `promtooling`: Defines and registers Prometheus metrics.
    *   `client/otlp`: Configures OpenTelemetry tracing exporters.
    *   `sentrytooling`: Integrates with Sentry for intelligent error reporting (filters noise).
*   **HTTP/REST (`resttooling`, `openapiapp`, `httpserver`):** Extensive toolkit for building HTTP clients and servers. Features composable roundtrippers/middleware for retries, timeouts, logging, metrics, tracing, auth, rate limiting. `openapiapp` provides helpers for OpenAPI-based services.
*   **Database (`rdb`, `chdb`, `pgtooling`, `sqltooling`, `redis`):** Instrumented wrappers for PostgreSQL (`rdb`, `r` stands for "relational"), ClickHouse (`chdb`), and Redis (`redis`), including connection management, query execution, retries, transactions (`rdb`), migrations (`pgtooling`), and SQL generation helpers (`sqltooling`).
*   **Background Jobs (`scheduletooling`):** Framework for scheduling and running background tasks using `go-quartz`, with built-in middleware (`scheduletooling/worker`) for instrumentation and reliability.
*   **Concurrency (`concurrent`):** Thread-safe utilities like maps, caches, and atomic-like values.
*   **Rate Limiting (`rate`, `ratetooling`):** Primitives and strategies for implementing rate limits (semaphore, fixed window). `resttooling` uses these.
*   **Retries (`retrytooling`):** Helpers for jitter strategies used by `resttooling` and `rdb`.
*   **Testing (`testtooling`):** Helpers for integration testing using Docker containers (`testcontainers-go`, `gnomock`) for databases (Postgres, ClickHouse, Redis) and other dependencies, plus general assertion utilities.
*   **Build/CLI (`buildtooling`, `clitooling`):** Utilities related to build information and command-line interface creation using Cobra.
*   **Domain Specific (`buses`):** Internal models and clients related to specific domains, like Cyprus buses.
*   **Go Build (`tools`):** Manages Go build tool dependencies.
*   **Utilities:** Various helpers for tasks like generating Haiku names (`haikutooling`), slice manipulation (`slicetooling`), unit conversion (`unittooling`), secret handling (`secret`), JSON/YAML handling (`jsontooling`, `yamlfromstruct`), project path detection (`projectpath`), timer management (`timer`), etc.

## Design Philosophy:

*   **Modularity:** Packages are designed to be reusable across different services.
*   **Consistency:** Provides standard ways to handle configuration, logging, errors, etc.
*   **Observability:** Built-in support for metrics, tracing, and structured logging.
*   **Resilience:** Includes mechanisms for retries, timeouts, and panic recovery.
*   **Testability:** Strong emphasis on integration testing with real dependencies via containers.
