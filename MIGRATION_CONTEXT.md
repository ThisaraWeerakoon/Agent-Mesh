# Migration Context: A2A Registry
- **Source:** TypeScript (Node.js) in folder `/a2a-registry`
- **Destination:** Golang in folder `/go-a2a-registry`
- **Goal:** Port the registry functionality to Go.
- **Constraint:** NO database persistence. Use In-Memory (RAM) storage with Mutex.
- **Framework:** Gin (HTTP), Zap (Logging), Viper (Config).
- **Architecture:** Clean Architecture (Domain -> Adapter -> Handler).