package registry

import (
	"fmt"
)

type Container struct {
	*LoggerService
	*ConfigService
	*DbService
	*ResponseBuilder
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

	responseBuilderService := NewResponseBuilderService(loggerService)

	container := &Container{
		LoggerService:   loggerService,
		ConfigService:   configService,
		DbService:       dbService,
		ResponseBuilder: responseBuilderService,
	}

	return container, nil
}
