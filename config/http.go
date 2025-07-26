package config

import (
	"time"

	"github.com/golibry/go-params/params"
)

// HttpServerConfig contains HTTP server configuration settings.
// It defines how the HTTP server should bind and handle requests.
type HttpServerConfig struct {
	// BindAddress specifies the IP address the HTTP server should bind to.
	// Must be a valid IPv4 address (e.g., "0.0.0.0", "127.0.0.1").
	BindAddress string `validate:"ipv4"`
	
	// BindPort specifies the port number the HTTP server should listen on.
	// Must be a numeric value (e.g., "8080", "3000").
	BindPort string `validate:"numeric"`
	
	// MaxHeaderBytes controls the maximum number of bytes the server will read
	// parsing the request header's keys and values, including the request line.
	// Must be between 0 and 64000 bytes.
	MaxHeaderBytes int `validate:"number,gte=0,lte=64000"`
	
	// RequestTimeout specifies the maximum duration for reading the entire request,
	// including the body. A zero or negative value means there will be no timeout.
	RequestTimeout time.Duration
}

func newHttpServerConfig() HttpServerConfig {
	bindAddress, _ := params.GetEnvAsString("HTTP_BIND_ADDRESS", "0.0.0.0")
	bindPort, _ := params.GetEnvAsString("HTTP_BIND_PORT", "8080")
	maxHeaderBytes, _ := params.GetEnvAsInt("HTTP_MAX_HEADER_BYTES", 1024*16)
	requestTimeout, _ := params.GetEnvAsDuration("HTTP_REQUEST_TIMEOUT", 15*time.Second)

	return HttpServerConfig{
		BindAddress:    bindAddress,
		BindPort:       bindPort,
		MaxHeaderBytes: maxHeaderBytes,
		RequestTimeout: requestTimeout,
	}
}
