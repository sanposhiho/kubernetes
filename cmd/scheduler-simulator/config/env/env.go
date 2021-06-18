package env

// Env represents the system environment.
type Env int

const (
	// Development indicates that it is running in a development environment such as a local machine.
	Development Env = iota
	// Production indicates that it is running in production server.
	Production
)
