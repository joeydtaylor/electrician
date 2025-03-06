package types

// LogLevel represents the severity levels for logging within the system, facilitating differentiated
// logging behavior and output customization.
type LogLevel int

// SinkType defines the type of logger sink.
type SinkType string

// Define constants for SinkType
const (
	FileSink    SinkType = "file"
	StdoutSink  SinkType = "stdout"
	NetworkSink SinkType = "network"
)

const (
	DebugLevel  LogLevel = iota // DebugLevel indicates debug messages.
	InfoLevel                   // InfoLevel indicates informational messages.
	WarnLevel                   // WarnLevel indicates warning messages.
	ErrorLevel                  // ErrorLevel indicates error messages.
	DPanicLevel                 // DPanicLevel indicates panic in development, error in production.
	PanicLevel                  // PanicLevel indicates panic messages.
	FatalLevel                  // FatalLevel indicates fatal error messages.
)

// SinkConfig defines the configuration for a logging sink.
type SinkConfig struct {
	Type   string                 // Type of sink, e.g., "file", "stdout", "network"
	Config map[string]interface{} // Detailed configuration specific to the sink type
}

// Logger defines the interface for logging across the library. It provides methods to log messages
// at various levels, which is fundamental for diagnostics and system monitoring.
type Logger interface {
	GetLevel() LogLevel                              // GetLevel returns the current logging level of the logger.
	SetLevel(LogLevel)                               // SetLevel sets the logging level of the logger.
	Debug(msg string, keysAndValues ...interface{})  // Debug logs a debug message.
	Info(msg string, keysAndValues ...interface{})   // Info logs an informational message.
	Warn(msg string, keysAndValues ...interface{})   // Warn logs a warning message.
	Error(msg string, keysAndValues ...interface{})  // Error logs an error message.
	DPanic(msg string, keysAndValues ...interface{}) // DPanic logs a critical message; panics in development.
	Panic(msg string, keysAndValues ...interface{})  // Panic logs a message and panics.
	Fatal(msg string, keysAndValues ...interface{})  // Fatal logs a fatal error and causes the application to exit.
	Flush() error                                    // Add this line to include a Flush method
	AddSink(identifier string, config SinkConfig) error
	RemoveSink(identifier string) error
	ListSinks() ([]string, error)
}
