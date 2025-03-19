package types

// ComponentMetadata defines the essential identifying information for components within the system.
// It includes identifiers and descriptive information to help manage and differentiate components dynamically.
type ComponentMetadata struct {
	ID   string // Unique identifier for the component.
	Type string // Type of the component, used to distinguish between different classes of components.
	Name string // Human-readable name for the component.
}

// TLSConfig holds the configuration necessary for setting up TLS communication within the system.
// This includes paths to certificate files and other TLS-specific settings.
type TLSConfig struct {
	UseTLS                 bool
	CertFile               string
	KeyFile                string
	CAFile                 string
	SubjectAlternativeName string
	MinTLSVersion          uint16 // e.g., tls.VersionTLS12
	MaxTLSVersion          uint16 // e.g., tls.VersionTLS13
}

// ConcurrencyConfig provides configuration details for managing concurrency within the system.
// It is typically used to configure the behavior of components that handle concurrent operations.
type ConcurrencyConfig[T any] struct {
	MaxConcurrency int // Maximum number of concurrent processing units or wires.
	BufferSize     int // Size of the buffer for channels between concurrent units, controlling flow and capacity.
}

// Transformer represents a function that transforms an input of type T into an output of the same type,
// potentially with an error if the transformation fails. It is a foundational concept for data processing
// and manipulation within the system.
type Transformer[T any] func(T) (T, error)

// Option defines a configuration option function applicable to any component T. This generic approach
// allows for flexible configuration mechanisms across different types of components.
type Option[T any] func(T)
