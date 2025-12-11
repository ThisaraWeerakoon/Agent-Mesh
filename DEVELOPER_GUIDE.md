# A2A Registry - Developer Guide

This guide provides an in-depth look at the architecture, implementation details, and design decisions of the A2A Registry Service. It is intended for developers who want to understand, extend, or maintain the codebase.

## 1. Architecture Overview

The project follows **Clean Architecture** (also known as Hexagonal Architecture or Ports and Adapters). This design separates the core business logic from external concerns like databases, HTTP frameworks, and gRPC interfaces.

### High-Level Diagram

```mermaid
graph TD
    subgraph "Core (Business Logic)"
        Domain[Domain Models]
        Ports[Ports (Interfaces)]
        Service[Service Implementation]
    end

    subgraph "Adapters (Infrastructure)"
        HTTP[HTTP Handler (Gin)]
        gRPC[gRPC Server]
        Repo[Memory Repository]
    end

    HTTP --> Service
    gRPC --> Service
    Service --> Ports
    Repo ..|> Ports
    Service --> Domain
```

### Layers

1.  **Domain Layer** (`internal/core/domain`):
    -   Contains the core entities: `RegistryEntry` and `AgentCard`.
    -   These are pure Go structs with no dependencies on external libraries (except standard ones like `time`).
    -   **Key Decision**: We use strict structs for `AgentCard` to enforce the A2A Protocol JSON Schema, ensuring type safety and validation.

2.  **Ports Layer** (`internal/core/ports`):
    -   Defines interfaces (contracts) for the Service and Repository.
    -   `RegistryService`: Defines business operations (Register, Get, List, etc.).
    -   `RegistryRepository`: Defines persistence operations (Create, Get, Update, etc.).
    -   **Benefit**: Allows us to swap implementations (e.g., switch from Memory to Postgres) without changing the core logic.

3.  **Service Layer** (`internal/core/services`):
    -   Implements the `RegistryService` interface.
    -   Contains the business rules (e.g., generating IDs, validation logic).
    -   Orchestrates data flow between the Adapters and the Repository.

4.  **Adapters Layer** (`internal/adapters`):
    -   **Handlers**:
        -   `http`: Uses the Gin framework to expose RESTful endpoints.
        -   `grpc`: Implements the generated Protobuf server interface.
    -   **Repository**:
        -   `memory`: A thread-safe, in-memory implementation using `sync.RWMutex`.

## 2. Directory Structure

```text
go-a2a-registry/
├── api/
│   └── proto/v1/          # Protobuf definitions
├── cmd/
│   └── server/            # Main application entry point
├── internal/
│   ├── adapters/
│   │   ├── handler/
│   │   │   ├── grpc/      # gRPC server implementation
│   │   │   └── http/      # HTTP (Gin) handlers
│   │   └── repository/
│   │       └── memory/    # In-memory storage implementation
│   └── core/
│       ├── domain/        # Domain models (AgentCard, etc.)
│       ├── ports/         # Interfaces (Service, Repository)
│       └── services/      # Business logic implementation
├── pkg/
│   └── api/v1/            # Generated Go code from Protobuf
├── tests/                 # Integration tests
└── tools/                 # Utility scripts (e.g., gRPC client)
```

## 3. Key Design Decisions

### 3.1. Clean Architecture
**Why?** To ensure the system is testable and maintainable.
-   **Testability**: We can easily mock the Repository to test the Service logic in isolation.
-   **Flexibility**: We can add new transports (like gRPC) without touching the core logic.

### 3.2. In-Memory Persistence (Phase 1)
**Why?** To focus on API contract and protocol compliance first.
-   We used a `map[string]*RegistryEntry` protected by `sync.RWMutex`.
-   **Trade-off**: Data is lost on restart. This is acceptable for Phase 1 but will be replaced by PostgreSQL in Phase 2.

### 3.3. A2A Protocol Compliance
**Why?** The registry must strictly adhere to the A2A JSON Schema.
-   We moved from a generic `map[string]interface{}` to a fully typed `AgentCard` struct.
-   **Validation**: We use Gin's `binding:"required"` tags and custom validation logic to ensure required fields (like `SupportedInterfaces`) are present.

### 3.4. Dual Transport (HTTP & gRPC)
**Why?** To support modern, high-performance clients (gRPC) while maintaining backward compatibility and ease of use (HTTP).
-   **Implementation**: Both servers run in the same process.
-   `main.go` uses a goroutine for the gRPC server so it doesn't block the HTTP server.

## 4. Implementation Details

### gRPC Support
-   **Proto Definition**: Located in `api/proto/v1/registry.proto`.
-   **Code Gen**: We use `protoc` to generate Go stubs in `pkg/api/v1`.
-   **Adapter**: The `RegistryServer` struct in `internal/adapters/handler/grpc` maps Protobuf messages to Domain models and calls the Service.

### Validation
-   HTTP requests are validated using struct tags (`binding:"required"`).
-   Logic errors (e.g., "Agent not found") are mapped to appropriate HTTP status codes (404) or gRPC codes (NotFound).

## 5. Future Roadmap

1.  **Persistence**: Replace `memory` repository with a `postgres` implementation.
2.  **Authentication**: Add an Auth Middleware (JWT/OAuth2) to secure endpoints.
3.  **Events**: Implement an Event Bus to publish registry events (AgentRegistered, AgentOffline).
