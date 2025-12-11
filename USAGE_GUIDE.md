# A2A Registry - Usage & Testing Guide

This guide provides step-by-step instructions on how to run, use, and test the A2A Registry Service (Go implementation).

## 1. Prerequisites
- **Go**: Version 1.21 or higher.
- **Terminal**: Any standard terminal (bash, zsh).
- **Curl**: For manual API testing.

## 2. Running the Server
Start the server using the following command:

```bash
# Default port is 3000
go run cmd/server/main.go
```

*You should see logs indicating the server is running on port 3000.*

## 3. Use Cases & API Examples

### Use Case 1: Registering an Agent
Register a new agent with a full A2A-compliant Agent Card.

**Endpoint**: `POST /api/v1/agents/`

**Request**:
```bash
curl -X POST http://localhost:3000/api/v1/agents/ \
  -H "Content-Type: application/json" \
  -d '{
    "agentCard": {
      "did": "did:peer:123456789",
      "name": "Weather Agent",
      "description": "Provides weather updates.",
      "protocolVersion": "1.0",
      "supportedInterfaces": [
        {
          "protocolBinding": "HTTP+JSON",
          "url": "https://weather-agent.example.com/api/v1"
        }
      ],
      "provider": {
        "organization": "Weather Corp",
        "url": "https://weather.example.com"
      },
      "capabilities": {
        "streaming": true,
        "pushNotifications": false
      },
      "skills": [
        {
          "name": "GetForecast",
          "description": "Returns weather forecast for a location.",
          "tags": ["weather", "forecast"]
        }
      ]
    },
    "tags": ["weather", "public"],
    "metadata": {
      "region": "us-east"
    }
  }'
```

**Response (201 Created)**:
Returns the created `RegistryEntry` with the `agentId` (which matches the DID).

---

### Use Case 2: Retrieving an Agent
Get the details of a registered agent using its DID.

**Endpoint**: `GET /api/v1/agents/:agentId`

**Request**:
```bash
curl http://localhost:3000/api/v1/agents/did:peer:123456789
```

---

### Use Case 3: Listing Agents
List all agents, optionally filtering by tags or skills.

**Endpoint**: `GET /api/v1/agents/`

**Examples**:
```bash
# List all
curl http://localhost:3000/api/v1/agents/

# Filter by tag
curl "http://localhost:3000/api/v1/agents/?tags=weather"

# Filter by skill name
curl "http://localhost:3000/api/v1/agents/?skill=GetForecast"
```

---

### Use Case 4: Sending a Heartbeat
Agents must send heartbeats to indicate they are active.

**Endpoint**: `POST /api/v1/agents/:agentId/heartbeat`

**Request**:
```bash
curl -X POST http://localhost:3000/api/v1/agents/did:peer:123456789/heartbeat
```

---

### Use Case 5: Deleting an Agent
Remove an agent from the registry.

**Endpoint**: `DELETE /api/v1/agents/:agentId`

**Request**:
```bash
curl -X DELETE http://localhost:3000/api/v1/agents/did:peer:123456789
```

## 4. Automated Testing
The project includes integration tests that verify the entire flow.

**Run All Tests**:
```bash
go test -v ./tests/...
```

**Expected Output**:
```text
=== RUN   TestRegisterAgent
--- PASS: TestRegisterAgent (0.00s)
=== RUN   TestRegisterAgentValidationFailure
--- PASS: TestRegisterAgentValidationFailure (0.00s)
...
PASS
ok      github.com/jenish2917/a2a-registry-go/tests     0.515s
```

## 5. gRPC Usage & Testing
The server also listens on port `50051` for gRPC requests.

### Option A: Using `grpcurl`
If you have `grpcurl` installed, you can interact with the API directly.

**List Agents**:
```bash
grpcurl -plaintext localhost:50051 a2a.registry.v1.RegistryService/ListAgents
```

**Get Agent**:
```bash
grpcurl -plaintext -d '{"agent_id": "did:peer:123456789"}' localhost:50051 a2a.registry.v1.RegistryService/GetAgent
```

### Option B: Using the Custom Go Client
We have provided a simple Go client to verify connectivity.

**Run Client**:
```bash
go run tools/grpc_client.go
```
*This will connect to localhost:50051 and attempt to list agents.*

### Option C: Using Postman
Postman supports gRPC requests. Follow these steps:

1.  **Create Request**: Click "New" -> "gRPC Request".
2.  **Enter URL**: Set the server URL to `localhost:50051`.
3.  **Import Proto**:
    *   Go to the "Service Definition" tab.
    *   Select "Import .proto file".
    *   Choose `api/proto/v1/registry.proto` from your project.
    *   Postman will load the `RegistryService` and its methods.

#### Example Payloads (Message JSON)

**1. RegisterAgent**
Select `RegisterAgent` method and use this JSON in the "Message" tab:
```json
{
  "agent_card": {
    "did": "did:peer:postman1",
    "name": "Postman Agent",
    "protocol_version": "1.0",
    "supported_interfaces": [
      {
        "protocol_binding": "HTTP+JSON",
        "url": "http://localhost:4000"
      }
    ]
  },
  "tags": ["postman", "test"]
}
```

**2. GetAgent**
Select `GetAgent` method:
```json
{
  "agent_id": "did:peer:postman1"
}
```

**3. ListAgents**
Select `ListAgents` method:
```json
{
  "limit": 10,
  "tags": ["postman"]
}
```

**4. Heartbeat**
Select `Heartbeat` method:
```json
{
  "agent_id": "did:peer:postman1"
}
```

**5. DeleteAgent**
Select `DeleteAgent` method:
```json
{
  "agent_id": "did:peer:postman1"
}
```
