package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	ErrMissingAppName    = errors.New("APPNAME is required")
	ErrMissingHostname   = errors.New("HOSTNAME is required")
	ErrInvalidLogLevel   = errors.New("invalid log level")
	ErrInvalidMaxSize    = errors.New("invalid max size (must be > 0)")
	ErrInvalidMaxBackups = errors.New("invalid max backups (must be >= 0)")
	ErrInvalidMaxAge     = errors.New("invalid max age (must be >= 0)")
)

// Config holds all configuration for the logger
type Config struct {
	// Application identity
	AppName  string
	Hostname string
	Env      string // environment: development, production, etc.

	// Log output
	LogDirectory  string
	EnableConsole bool
	EnableFile    bool
	EnableBuffer  bool

	// Log rotation
	MaxSize    int  // megabytes
	MaxBackups int
	MaxAge     int // days
	Compress   bool

	// Log format
	TimeFormat       string
	EnableCaller     bool
	EnableStacktrace bool

	// Log levels
	MinLevel   string // debug, info, warn, error
	ErrorLevel string // logs at this level and above go to error file

	// Thread/Process info
	ProcessID int
	ThreadID  string

	// Performance
	FlushInterval time.Duration
	BufferSize    int // bytes

	// Tracing
	EnableTracing bool
}

// LoadFromEnv creates a Config from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := Default()

	// Required fields
	cfg.AppName = os.Getenv("APPNAME")
	cfg.Hostname = os.Getenv("HOSTNAME")

	// Optional fields
	if env := os.Getenv("ENV"); env != "" {
		cfg.Env = env
	}
	if env := os.Getenv("LOG_DIR"); env != "" {
		cfg.LogDirectory = env
	}
	if env := os.Getenv("LOG_LEVEL"); env != "" {
		cfg.MinLevel = env
	}
	if env := os.Getenv("ERROR_LEVEL"); env != "" {
		cfg.ErrorLevel = env
	}
	if env := os.Getenv("LOG_TIME_FORMAT"); env != "" {
		cfg.TimeFormat = env
	}
	if env := os.Getenv("LOG_MAX_SIZE"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			cfg.MaxSize = val
		}
	}
	if env := os.Getenv("LOG_MAX_BACKUPS"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			cfg.MaxBackups = val
		}
	}
	if env := os.Getenv("LOG_MAX_AGE"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			cfg.MaxAge = val
		}
	}
	if env := os.Getenv("LOG_COMPRESS"); env != "" {
		cfg.Compress = env == "true" || env == "1"
	}
	if env := os.Getenv("LOG_ENABLE_CONSOLE"); env != "" {
		cfg.EnableConsole = env == "true" || env == "1"
	}
	if env := os.Getenv("LOG_ENABLE_BUFFER"); env != "" {
		cfg.EnableBuffer = env == "true" || env == "1"
	}
	if env := os.Getenv("LOG_BUFFER_SIZE"); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			cfg.BufferSize = val
		}
	}
	if env := os.Getenv("LOG_ENABLE_CALLER"); env != "" {
		cfg.EnableCaller = env == "true" || env == "1"
	}
	if env := os.Getenv("LOG_ENABLE_TRACE"); env != "" {
		cfg.EnableTracing = env == "true" || env == "1"
	}
	if env := os.Getenv("LOG_FLUSH_INTERVAL"); env != "" {
		if val, err := time.ParseDuration(env); err == nil {
			cfg.FlushInterval = val
		}
	}

	// Process ID
	cfg.ProcessID = os.Getpid()

	return cfg, cfg.Validate()
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.AppName == "" {
		return ErrMissingAppName
	}
	if c.Hostname == "" {
		return ErrMissingHostname
	}
	if !isValidLogLevel(c.MinLevel) {
		return fmt.Errorf("%w: %s", ErrInvalidLogLevel, c.MinLevel)
	}
	if !isValidLogLevel(c.ErrorLevel) {
		return fmt.Errorf("%w: %s", ErrInvalidLogLevel, c.ErrorLevel)
	}
	if c.MaxSize <= 0 {
		return fmt.Errorf("%w: %d", ErrInvalidMaxSize, c.MaxSize)
	}
	if c.MaxBackups < 0 {
		return fmt.Errorf("%w: %d", ErrInvalidMaxBackups, c.MaxBackups)
	}
	if c.MaxAge < 0 {
		return fmt.Errorf("%w: %d", ErrInvalidMaxAge, c.MaxAge)
	}
	return nil
}

func isValidLogLevel(level string) bool {
	switch level {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

// WithDefaults sets default values for any unset fields
func (c *Config) WithDefaults() *Config {
	if c.LogDirectory == "" {
		c.LogDirectory = "log"
	}
	if c.TimeFormat == "" {
		c.TimeFormat = "[2006-01-02 15:04:05.000]"
	}
	if c.MinLevel == "" {
		c.MinLevel = "info"
	}
	if c.ErrorLevel == "" {
		c.ErrorLevel = "error"
	}
	if c.MaxSize == 0 {
		c.MaxSize = 100 // 100 MB
	}
	if c.MaxBackups == 0 {
		c.MaxBackups = 10
	}
	if c.MaxAge == 0 {
		c.MaxAge = 30 // days
	}
	if c.FlushInterval == 0 {
		c.FlushInterval = 5 * time.Second
	}
	if c.BufferSize == 0 {
		c.BufferSize = 256 * 1024 // 256 KB
	}
	if c.Env == "" {
		c.Env = "production"
	}
	return c
}

// GetLogFileName returns the log file name with the given prefix
func (c *Config) GetLogFileName(prefix string) string {
	return fmt.Sprintf("%s-v01-%s-%s.log", prefix, c.AppName, c.Hostname)
}
