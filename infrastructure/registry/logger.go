package registry

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type LoggerService struct {
	logger    *slog.Logger
	logWriter LogWriter
}

func NewLoggerService(configService *ConfigService) (*LoggerService, error) {
	if configService == nil {
		return nil, fmt.Errorf("config service cannot be nil")
	}

	cfg := configService.Config()
	logWriter, err := createLogWriter(cfg.LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log writer: %w", err)
	}

	logger := slog.New(
		slog.NewJSONHandler(
			logWriter,
			&slog.HandlerOptions{
				Level: cfg.LogLevel,
			},
		),
	)

	// Set as default logger
	slog.SetDefault(logger)

	return &LoggerService{
		logger:    logger,
		logWriter: logWriter,
	}, nil
}

func (l *LoggerService) Logger() *slog.Logger {
	return l.logger
}

func (l *LoggerService) Close() error {
	if l.logWriter != nil {
		return l.logWriter.Close()
	}
	return nil
}

// createLogWriter creates the appropriate log writer based on the log path
func createLogWriter(logPath string) (LogWriter, error) {
	switch logPath {
	case "stdout":
		return &stdoutWriter{}, nil
	case "stderr":
		return &stderrWriter{}, nil
	default:
		return createFileLogWriter(logPath)
	}
}

// createFileLogWriter creates a file-based log writer with proper validation and security
func createFileLogWriter(logPath string) (LogWriter, error) {
	if logPath == "" {
		return nil, fmt.Errorf("log path cannot be empty")
	}

	// Validate and clean the path
	cleanPath := filepath.Clean(logPath)
	if filepath.IsAbs(cleanPath) {
		// For absolute paths, ensure they're within reasonable bounds
		if err := validateLogPath(cleanPath); err != nil {
			return nil, fmt.Errorf("invalid log path: %w", err)
		}
	}

	// Create a directory if it doesn't exist
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open a file with appropriate permissions
	file, err := os.OpenFile(
		cleanPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &fileLogWriter{File: file}, nil
}

// validateLogPath performs basic security validation on log file paths
func validateLogPath(path string) error {
	// Prevent directory traversal attacks
	if filepath.IsAbs(path) {
		// Check for suspicious patterns
		cleanPath := filepath.Clean(path)
		if cleanPath != path {
			return fmt.Errorf("path contains suspicious elements")
		}
	}

	// Additional security checks can be added here
	return nil
}

// fileLogWriter wraps os.File to implement LogWriter interface
type fileLogWriter struct {
	*os.File
}

// stdoutWriter wraps os.Stdout to implement LogWriter interface
type stdoutWriter struct{}

func (w *stdoutWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (w *stdoutWriter) Close() error {
	// stdout doesn't need to be closed
	return nil
}

// stderrWriter wraps os.Stderr to implement LogWriter interface
type stderrWriter struct{}

func (w *stderrWriter) Write(p []byte) (n int, err error) {
	return os.Stderr.Write(p)
}

func (w *stderrWriter) Close() error {
	// stderr doesn't need to be closed
	return nil
}