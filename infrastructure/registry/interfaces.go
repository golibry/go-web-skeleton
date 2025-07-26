package registry

import (
	"io"
)

// Closer defines the interface for services that need cleanup
type Closer interface {
	Close() error
}

// LogWriter defines the interface for log writers that can be closed
type LogWriter interface {
	io.Writer
	io.Closer
}
