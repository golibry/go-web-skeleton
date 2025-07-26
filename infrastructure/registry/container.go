package registry

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/golibry/go-web-skeleton/config"
)

type Container struct {
	*LoggerService
	*ConfigService
	*DbService
}

func NewContainer() (*Container, error) {
	configService, err := NewConfigService()
	if err != nil {
		return nil, err
	}

	dbService, err := NewDbService(configService)
	if err != nil {
		return nil, err
	}

	container := &Container{
		NewLoggerService(configService),
		configService,
		dbService,
	}

	return container, nil
}

type LoggerService struct {
	logger *slog.Logger
}

func NewLoggerService(configService *ConfigService) *LoggerService {
	var logWriter *os.File
	if configService.Config().LogPath == "stdout" {
		logWriter = os.Stdout
	} else if configService.Config().LogPath == "stderr" {
		logWriter = os.Stderr
	} else {
		logWriter, _ = os.OpenFile(
			configService.Config().LogPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
	}

	logger := slog.New(
		slog.NewJSONHandler(
			logWriter,
			&slog.HandlerOptions{
				Level: configService.Config().LogLevel,
			},
		),
	)
	slog.SetDefault(logger)
	return &LoggerService{logger: logger}
}

func (l *LoggerService) Logger() *slog.Logger {
	return l.logger
}

type ConfigService struct {
	config config.Config
}

func NewConfigService() (*ConfigService, error) {
	cfg, err := config.BuildConfig()
	if err != nil {
		return nil, err
	}
	return &ConfigService{config: cfg}, nil
}

func (c *ConfigService) Config() config.Config {
	return c.config
}

type DbService struct {
	db *sql.DB
}

func NewDbService(cfg *ConfigService) (*DbService, error) {
	db, err := newMysqlDbConnectionPool(cfg.Config().Db)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &DbService{db}, nil
}

func (dbService *DbService) Db() *sql.DB {
	return dbService.db
}
