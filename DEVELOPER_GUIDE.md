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

## 6. ADK & Mesh Integration (Remote Agents)

This section explains the execution flow when an ADK agent calls another agent over the AgentMesh.

### 6.1. Component Overview
- **Client (`cmd/adk_example/client/main.go`)**: The consumer application. It runs a `root_agent` that wants to delegate a task (e.g., summarization) to another agent. It uses `RemoteTool` to create a proxy for the remote agent.
- **Sidecar (`cmd/sidecar/main.go`)**: The mesh proxy. It runs alongside the application and handles all network communication, routing, and protocol translation between agents.
- **Agent Server (`cmd/adk_example/server/main.go`)**: The provider application. It exposes a specific ADK agent ("summary_agent") to the network using `mesh_adk.ServeAgent`.
- **Registry (`cmd/server/main.go`)**: Stores agent discovery information (address, supported skills), allowing sidecars to find each other.

### 6.2. Detailed Execution Flow

1.  **Client Initialization**:
    *   In `client/main.go`, the application initializes.
    *   It creates a `root_agent` (Gemini-powered).
    *   It creates a proxy tool using `mesh_adk.RemoteTool("summary_agent", ...)` defined in `pkg/adk/tool.go`.
    *   The `agenttool` logic automatically defines a schema expecting a `"request"` string.

2.  **Tool Invocation (Client-Side)**:
    *   During execution, the `root_agent` generates a call to `summary_agent` with a specific prompt.
    *   The handler in `pkg/adk/tool.go` is invoked:
        *   It extracts the `"request"` string from the arguments.
        *   It connects via gRPC to the **Local Sidecar** (default port `50052`).
        *   It sends a `TaskSendRequest` (Protobuf) containing the prompt to the sidecar via a specialized stream.

3.  **Mesh Routing (Sidecar)**:
    *   The Sidecar receiving the request checks the `TargetAgentId` ("summary_agent").
    *   *(In a distributed setup, it would query the Registry/A2A to find the remote IP. In this local example, it routes locally).*
    *   It forwards the gRPC stream to the **Agent Server**.

4.  **Agent Execution (Server-Side)**:
    *   The **Agent Server** (`pkg/adk/server.go`) receives the `TaskStart` event.
    *   It creates a transient **ADK Runner** instance specifically for this task.
    *   It initializes an **In-Memory Session** to track the conversation state.
    *   It unwraps the text prompt and calls `runner.Run()`.
    *   This executes the real `summary_agent` logic (another Gemini call) locally on the server.

5.  **Streaming Response**:
    *   As the `summary_agent` generates tokens (thinking or final answer), the `server.go` handler captures these events.
    *   It wraps the text output into `TaskStatusUpdate` messages (status: `WORKING` or `COMPLETED`).
    *   These messages are streamed back through the gRPC connection: `Server` -> `Sidecar` -> `Client`.

6.  **Completion**:
    *   The Client's `RemoteTool` handler aggregates the streamed text chunks.
    *   Upon completion, it returns the final summary string to the `root_agent`.
    *   The `root_agent` uses this summary to formulate its final answer to the user.

### 6.3. How to Start the Application

To run the complete system locally, open three separate terminal windows:

**1. Start the Sidecar**
Acts as the communication hub.
```bash
cd cmd/sidecar
go run main.go
# Output:
# Local listener started on localhost:50052
# External listener started on 0.0.0.0:50053
```

**2. Start the Agent Server (Provider)**
Hosts the "summary_agent".
```bash
cd cmd/adk_example/server
export GOOGLE_API_KEY="your_api_key_here"
go run main.go
# Output:
# Agent server listening on localhost:50054
```

**3. Run the Client (Consumer)**
Runs the "root_agent" which calls the summary agent.
```bash
cd cmd/adk_example/client
export GOOGLE_API_KEY="your_api_key_here"
go run main.go "Summarize: A magical poetic retelling of the fairy tale The Snow Queen..."
# Output:
# Agent -> The text describes a poetic retelling of...
```
