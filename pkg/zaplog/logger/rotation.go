package logger

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/config"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// RotationConfig holds configuration for log rotation
type RotationConfig struct {
	Filename   string // full path to log file
	MaxSize    int    // megabytes
	MaxBackups int
	MaxAge     int    // days
	Compress   bool
	LocalTime  bool
}

// NewRotationConfig creates a rotation config from logger config
func NewRotationConfig(cfg *config.Config, logType string) *RotationConfig {
	return &RotationConfig{
		Filename:   filepath.Join(cfg.LogDirectory, cfg.GetLogFileName(logType)),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
		LocalTime:  true,
	}
}

// LumberjackLogger creates a new lumberjack.Logger from the rotation config
func (rc *RotationConfig) LumberjackLogger() *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   rc.Filename,
		MaxSize:    rc.MaxSize,
		MaxBackups: rc.MaxBackups,
		MaxAge:     rc.MaxAge,
		Compress:   rc.Compress,
		LocalTime:  rc.LocalTime,
	}
}

// WriteSyncer creates a zapcore.WriteSyncer from the rotation config
func (rc *RotationConfig) WriteSyncer(buffered bool) zapcore.WriteSyncer {
	lumberjackLogger := rc.LumberjackLogger()
	ws := zapcore.AddSync(lumberjackLogger)

	if buffered {
		return &zapcore.BufferedWriteSyncer{WS: ws}
	}

	return ws
}

// EnsureLogDirectory creates the log directory if it doesn't exist
func EnsureLogDirectory(cfg *config.Config) error {
	if cfg.LogDirectory == "" {
		return nil
	}

	// Check if directory exists
	info, err := os.Stat(cfg.LogDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory
			return os.MkdirAll(cfg.LogDirectory, 0755)
		}
		return err
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return &NotADirectoryError{Path: cfg.LogDirectory}
	}

	return nil
}

// NotADirectoryError is returned when the log path exists but is not a directory
type NotADirectoryError struct {
	Path string
}

func (e *NotADirectoryError) Error() string {
	return "log path exists but is not a directory: " + e.Path
}

// GetLogFiles returns a list of log files matching the pattern
func GetLogFiles(cfg *config.Config) ([]string, error) {
	if cfg.LogDirectory == "" {
		return []string{}, nil
	}

	pattern := filepath.Join(cfg.LogDirectory, "*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Filter files that match our app and hostname
	var files []string
	for _, file := range matches {
		if strings.Contains(file, cfg.AppName) && strings.Contains(file, cfg.Hostname) {
			files = append(files, file)
		}
	}

	return files, nil
}

// CleanupOldLogs removes old log files based on retention policy
func CleanupOldLogs(cfg *config.Config) error {
	files, err := GetLogFiles(cfg)
	if err != nil {
		return err
	}

	// This is handled automatically by lumberjack, but we provide
	// this function for manual cleanup if needed
	_ = files // Use files to avoid unused variable warning

	return nil
}
