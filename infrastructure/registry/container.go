package registry

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golibry/go-web-skeleton/config"
)

type Container struct {
	*LoggerService
	*ConfigService
	*DbService
}

// Close implements graceful shutdown for all services
func (c *Container) Close() error {
	var errors []error

	// Close database service
	if c.DbService != nil {
		if err := c.DbService.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database service: %w", err))
		}
	}

	// Close logger service
	if c.LoggerService != nil {
		if err := c.LoggerService.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close logger service: %w", err))
		}
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("errors during container shutdown: %v", errors)
	}

	return nil
}

func NewContainer() (*Container, error) {
	configService, err := NewConfigService()
	if err != nil {
		return nil, fmt.Errorf("failed to create config service: %w", err)
	}

	loggerService, err := NewLoggerService(configService)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger service: %w", err)
	}

	dbService, err := NewDbService(configService)
	if err != nil {
		// Clean up logger service on failure
		_ = loggerService.Close()
		return nil, fmt.Errorf("failed to create database service: %w", err)
	}

	container := &Container{
		LoggerService: loggerService,
		ConfigService: configService,
		DbService:     dbService,
	}

	return container, nil
}

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

type ConfigService struct {
	config config.Config
}

func NewConfigService() (*ConfigService, error) {
	cfg, err := config.BuildConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build configuration: %w", err)
	}
	return &ConfigService{config: cfg}, nil
}

func (c *ConfigService) Config() config.Config {
	return c.config
}

type DbService struct {
	db     *sql.DB
	config config.DatabaseConfig
}

func NewDbService(cfg *ConfigService) (*DbService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config service cannot be nil")
	}

	dbConfig := cfg.Config().Db
	db, err := createDatabaseConnection(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	service := &DbService{
		db:     db,
		config: dbConfig,
	}

	// Test the connection with context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.Ping(ctx); err != nil {
		_ = service.Close() // Clean up on failure
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return service, nil
}

func (d *DbService) Db() *sql.DB {
	return d.db
}

func (d *DbService) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *DbService) Ping(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return d.db.PingContext(ctx)
}

// createDatabaseConnection creates and configures a database connection
func createDatabaseConnection(config config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", config.Dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetConnMaxIdleTime(config.ConnectionMaxIdleTime)
	db.SetConnMaxLifetime(config.ConnectionMaxLifetime)

	return db, nil
}
