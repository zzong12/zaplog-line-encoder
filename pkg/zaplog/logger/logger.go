package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/config"
	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/context"
	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/encoder"
	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger defines the logging interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger

	Sync() error
}

// Field represents a key-value pair for log fields
type Field struct {
	Key   string
	Value interface{}
}

// zapLogger implements the Logger interface
type zapLogger struct {
	zapLogger *zap.Logger
	cfg       *config.Config
}

// New creates a new Logger with the given configuration
func New(cfg *config.Config) (Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Apply defaults
	cfg = cfg.WithDefaults()

	// Create encoder
	enc := encoder.NewLineEncoder(zapcore.EncoderConfig{
		TimeKey:          zapcore.OmitKey,
		LevelKey:         "|",
		NameKey:          zapcore.OmitKey,
		CallerKey:        zapcore.OmitKey,
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       zapcore.OmitKey,
		StacktraceKey:    zapcore.OmitKey,
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout(cfg.TimeFormat),
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "|",
	}, true)

	// Create level enablers
	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev < zapcore.ErrorLevel && lev >= zapcore.DebugLevel
	})

	// Create write syncers
	var cores []zapcore.Core

	if cfg.EnableFile {
		infoFileWriteSyncer := createWriteSyncer(cfg, "proc", cfg.EnableBuffer)
		errorFileWriteSyncer := createWriteSyncer(cfg, "err", cfg.EnableBuffer)

		infoFileCore := zapcore.NewCore(enc, zapcore.NewMultiWriteSyncer(infoFileWriteSyncer), lowPriority)
		errorFileCore := zapcore.NewCore(enc, zapcore.NewMultiWriteSyncer(errorFileWriteSyncer), highPriority)

		cores = append(cores, infoFileCore, errorFileCore)
	}

	if cfg.EnableConsole {
		consoleCore := zapcore.NewCore(enc, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), lowPriority)
		cores = append(cores, consoleCore)
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no output enabled (enable file or console)")
	}

	// Create tee core
	core := zapcore.NewTee(cores...)

	// Create logger with caller
	zapLogger := zap.New(core, zap.AddCaller())

	return &zapLogger{
		zapLogger: zapLogger,
		cfg:       cfg,
	}, nil
}

// NewFromEnv creates a new Logger from environment variables
func NewFromEnv() (Logger, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}
	return New(cfg)
}

// NewDefault creates a new Logger with default configuration
// Must set APPNAME and HOSTNAME environment variables
func NewDefault() (Logger, error) {
	return NewFromEnv()
}

// createWriteSyncer creates a write syncer for the given log type
func createWriteSyncer(cfg *config.Config, logType string, buffered bool) zapcore.WriteSyncer {
	fileName := cfg.GetLogFileName(logType)
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.LogDirectory + "/" + fileName,
		MaxSize:    cfg.MaxSize,    // megabytes
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,     // days
		Compress:   cfg.Compress,
		LocalTime:  true,
	}

	ws := zapcore.AddSync(lumberjackLogger)

	if buffered {
		return &zapcore.BufferedWriteSyncer{WS: ws}
	}

	return ws
}

// Logger interface methods

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.zapLogger.Debug(msg, fieldsToZap(fields)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.zapLogger.Info(msg, fieldsToZap(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.zapLogger.Warn(msg, fieldsToZap(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.zapLogger.Error(msg, fieldsToZap(fields)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.zapLogger.Fatal(msg, fieldsToZap(fields)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		zapLogger: l.zapLogger.With(fieldsToZap(fields)...),
		cfg:       l.cfg,
	}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	traceInfo := context.FromContext(ctx)
	if traceInfo == nil {
		// No trace info, return logger as is
		return l
	}

	// Create encoder with trace info
	enc := encoder.NewLineEncoder(zapcore.EncoderConfig{
		TimeKey:          zapcore.OmitKey,
		LevelKey:         "|",
		NameKey:          zapcore.OmitKey,
		CallerKey:        zapcore.OmitKey,
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       zapcore.OmitKey,
		StacktraceKey:    zapcore.OmitKey,
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout(l.cfg.TimeFormat),
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "|",
	}, true)

	enc.SetTraceInfo(
		traceInfo.AppName,
		traceInfo.TraceID,
		traceInfo.SpanID,
		traceInfo.ParentSpanID,
		traceInfo.ProcessID,
		traceInfo.ThreadID,
	)

	return &zapLogger{
		zapLogger: l.zapLogger,
		cfg:       l.cfg,
	}
}

func (l *zapLogger) Sync() error {
	return l.zapLogger.Sync()
}

// Helper functions

// fieldsToZap converts Field slices to zap.Field slices
func fieldsToZap(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}

// Package-level convenience functions for backward compatibility

var defaultLogger Logger

// InitLogger initializes the default logger from environment variables
// Deprecated: Use NewFromEnv() instead
func InitLogger() {
	logger, err := NewFromEnv()
	if err != nil {
		panic(err)
	}
	defaultLogger = logger
}

// GetLogger returns the default logger (must call InitLogger first)
// Deprecated: Use NewFromEnv() instead
func GetLogger() *zap.Logger {
	if defaultLogger == nil {
		panic("logger not initialized, call InitLogger first")
	}
	if zl, ok := defaultLogger.(*zapLogger); ok {
		return zl.zapLogger
	}
	panic("logger type error")
}

// GetZapLogger returns a zap.Logger with trace context
// Deprecated: Use NewFromEnv().WithContext() instead
func GetZapLogger(ctx context.Context) *zap.Logger {
	if defaultLogger == nil {
		panic("logger not initialized, call InitLogger first")
	}

	traceInfo := context.FromContext(ctx)
	if traceInfo == nil {
		if zl, ok := defaultLogger.(*zapLogger); ok {
			return zl.zapLogger
		}
		panic("logger type error")
	}

	if zl, ok := defaultLogger.(*zapLogger); ok {
		return zl.zapLogger.With(
			zap.String("|", traceInfo.AppName),
			zap.String("|", traceInfo.SpanID),
			zap.String("|", traceInfo.TraceID),
			zap.Int("|", traceInfo.ProcessID),
			zap.String("|", ""),
		)
	}

	panic("logger type error")
}

// String creates a string field
func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

// Int creates an int field
func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

// Int64 creates an int64 field
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any creates a field with any value
func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}
