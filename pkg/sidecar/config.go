package sidecar

// Config holds the configuration for the Sidecar.
type Config struct {
	// AgentID is the identity of the local agent.
	AgentID string
	// RegistryURL is the address of the central Registry.
	RegistryURL string
	// LocalPort is the port for the local agent to connect to (Plaintext gRPC).
	LocalPort int
	// ExternalPort is the port for other sidecars to connect to (mTLS gRPC).
	ExternalPort int
	// CertFile is the path to the certificate file for mTLS.
	CertFile string
	// KeyFile is the path to the key file for mTLS.
	KeyFile string
	// CAFile is the path to the CA certificate file for mTLS.
	CAFile string
	// AppPort is the port where the local agent application is running.
	AppPort int
}
