# zaplog-line-encoder

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.18-blue)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/zzong12/zaplog-line-encoder)](https://goreportcard.com/report/github.com/zzong12/zaplog-line-encoder)

A high-performance, production-ready logging library for Go built on Uber Zap with custom line-format encoding. Perfect for microservices and distributed systems requiring structured logging with distributed tracing support.

## Features

- **Custom Line Format**: Clean, pipe-delimited log format optimized for parsing and log aggregation
- **Distributed Tracing**: Built-in support for trace ID and span ID propagation with context-aware logging
- **High Performance**: Buffered writing with object pooling for minimal allocation and maximum throughput
- **Flexible Configuration**: Support for environment variables, programmatic configuration, and sensible defaults
- **Automatic Log Rotation**: Built-in log rotation with compression using lumberjack
- **Production Ready**: Comprehensive error handling, validation, and graceful degradation
- **Context-Aware**: Seamless integration with Go's context package for distributed tracing
- **Dual-Level Output**: Separate log files for error and info/debug levels

## Log Format

Logs are formatted as pipe-delimited fields for easy parsing:

```
[timestamp]|[level]|[app-name]|[trace-id]|[span-id]|[parent-span-id]|[file:line]|[function]|[pid]|[tid]|message|fields
```

### Example

```
[2024-01-13 10:30:45.123]|INFO|myapp|abc123-def456|span-789|parent-span-101|main.go:42|main.processRequest|12345|6|#KEY_PROC:KEYNODE=ProcessOrder, Order processed successfully|user_id=12345 order_id=67890
```

### Field Descriptions

| Field | Description | Example |
|-------|-------------|---------|
| timestamp | Log timestamp | `[2024-01-13 10:30:45.123]` |
| level | Log level | `INFO`, `DEBUG`, `WARN`, `ERROR` |
| app-name | Application name | `myapp` |
| trace-id | Distributed trace ID | `abc123-def456` |
| span-id | Current span ID | `span-789` |
| parent-span-id | Parent span ID | `parent-span-101` |
| file:line | Source location | `main.go:42` |
| function | Calling function | `main.processRequest` |
| pid | Process ID | `12345` |
| tid | Thread ID (goroutine count) | `6` |
| message | Log message | `#KEY_PROC:KEYNODE=...` |
| fields | Additional key-value pairs | `user_id=12345` |

## Installation

```bash
go get github.com/zzong12/zaplog-line-encoder
```

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/zzong12/zaplog-line-encoder/pkg/zaplog/logger"
)

func main() {
    // Set required environment variables
    os.Setenv("APPNAME", "myapp")
    os.Setenv("HOSTNAME", "server-01")

    // Initialize logger
    log, err := logger.NewFromEnv()
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    // Basic logging
    log.Info("Application started")
    log.Warn("High memory usage detected")
    log.Error("Database connection failed")
}
```

### With Custom Configuration

```go
package main

import (
    "github.com/zzong12/zaplog-line-encoder/pkg/zaplog/config"
    "github.com/zzong12/zaplog-line-encoder/pkg/zaplog/logger"
)

func main() {
    cfg := &config.Config{
        AppName:        "myapp",
        Hostname:       "server-01",
        LogDirectory:   "/var/log/myapp",
        EnableConsole:  true,
        EnableFile:     true,
        EnableBuffer:   true,
        MaxSize:        100, // 100 MB
        MaxBackups:     10,
        MaxAge:         30, // days
        Compress:       true,
        MinLevel:       "info",
        EnableCaller:   true,
        EnableTracing:  true,
    }

    log, err := logger.New(cfg)
    if err != nil {
        panic(err)
    }
    defer log.Sync()

    log.Info("Application configured and ready")
}
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `APPNAME` | Yes | - | Application name (used in log filenames) |
| `HOSTNAME` | Yes | - | Server/hostname identifier |
| `ENV` | No | `production` | Environment (development, production) |
| `LOG_DIR` | No | `log` | Log directory path |
| `LOG_LEVEL` | No | `info` | Minimum log level (debug, info, warn, error) |
| `ERROR_LEVEL` | No | `error` | Error log level |
| `LOG_MAX_SIZE` | No | `100` | Maximum log file size in MB |
| `LOG_MAX_BACKUPS` | No | `10` | Maximum number of backup files |
| `LOG_MAX_AGE` | No | `30` | Maximum age of log files in days |
| `LOG_COMPRESS` | No | `true` | Compress rotated log files |
| `LOG_ENABLE_CONSOLE` | No | `false` | Enable console output |
| `LOG_ENABLE_BUFFER` | No | `true` | Enable buffered writing |
| `LOG_BUFFER_SIZE` | No | `262144` | Buffer size in bytes (256 KB) |
| `LOG_ENABLE_CALLER` | No | `true` | Include caller information |
| `LOG_ENABLE_TRACE` | No | `true` | Enable distributed tracing |
| `LOG_FLUSH_INTERVAL` | No | `5s` | Buffer flush interval |
| `LOG_TIME_FORMAT` | No | `[2006-01-02 15:04:05.000]` | Time format |

### Configuration Presets

```go
// Development configuration
cfg := config.Development()
cfg.AppName = "myapp"
cfg.Hostname = "localhost"

// Production configuration
cfg := config.Production()
cfg.AppName = "myapp"
cfg.Hostname = "server-01"

// Default configuration
cfg := config.Default()
cfg.AppName = "myapp"
cfg.Hostname = "server-01"
```

## Distributed Tracing

### Creating a Trace Context

```go
import (
    "context"
    "github.com/zzong12/zaplog-line-encoder/pkg/zaplog/context"
)

func handleRequest(ctx context.Context) {
    // Create a new trace context
    ctx = context.NewContext(ctx, "trace-123-456")
    ctx = context.WithAppName(ctx, "myapp")

    // Use logger with trace context
    log := log.WithContext(ctx)
    log.Info("Handling request")
}
```

### Creating Child Spans

```go
func processOrder(ctx context.Context) {
    // Create child span
    ctx = context.WithSpan(ctx)
    log := log.WithContext(ctx)

    log.Info("Processing order")

    // Create another child span
    ctx = context.WithSpan(ctx)
    log = log.WithContext(ctx)
    log.Info("Validating order")
}
```

### Adding Parent Span

```go
func withParentSpan(ctx context.Context, parentSpanID string) {
    ctx = context.WithParentSpan(ctx, parentSpanID)
    log := log.WithContext(ctx)
    log.Info("Processing with parent span")
}
```

### Extracting Trace Information

```go
ctx := context.NewContext(context.Background(), "trace-abc-123")

traceID := context.GetTraceID(ctx)
spanID := context.GetSpanID(ctx)
parentSpanID := context.GetParentSpanID(ctx)
appName := context.GetAppName(ctx)

fmt.Printf("Trace: %s, Span: %s, Parent: %s, App: %s\n",
    traceID, spanID, parentSpanID, appName)
```

## Message Formatting

### Using Formatters

```go
import "github.com/zzong12/zaplog-line-encoder/pkg/zaplog/formatter"

// Key-process message
log.Info(formatter.FormatKeyProc("ProcessOrder", "Order processed successfully"))
// Output: #KEY_PROC:KEYNODE=ProcessOrder, Order processed successfully

// Error message
log.Error(formatter.FormatError("DB001", "Failed to connect to database"))
// Output: #EX_ERR:ERR=DB001,EMSG=Failed to connect to database

// Warning message
log.Warn(formatter.FormatWarning("High memory usage detected"))
// Output: #WARN:DESC=High memory usage detected

// Info message
log.Info(formatter.FormatInfo("Processing started"))
// Output: #INFO:Processing started

// Debug message
log.Debug(formatter.FormatDebug("Variable value: x=10"))
// Output: #DEBUG:Variable value: x=10
```

### Custom Formatters

```go
f := formatter.NewLineFormatter()

// Register custom template
f.RegisterTemplate("custom_template", "#CUSTOM:ACTION=%s,STATUS=%s")

// Use custom template
msg := f.UseTemplate("custom_template", "Login", "Success")
log.Info(msg)
```

## Logging with Fields

```go
log.Info("User login",
    logger.String("user_id", "12345"),
    logger.String("ip", "192.168.1.1"),
    logger.Int("port", 8080),
)

log.Error("Database connection failed",
    logger.String("host", "localhost:5432"),
    logger.Error(err),
)
```

## Child Loggers

```go
// Create a child logger with preset fields
requestLogger := log.With(
    logger.String("request_id", "req-123"),
    logger.String("user_id", "user-456"),
)

// All log entries from requestLogger will include these fields
requestLogger.Info("Processing request")
requestLogger.Warn("Slow query detected")
requestLogger.Error("Validation failed")
```

## Architecture

### Design Principles

1. **Separation of Concerns**: Clear separation between configuration, encoding, formatting, and logging
2. **Performance First**: Object pooling, buffered writing, and minimal allocations
3. **Interface-Based Design**: Easy to test and extend
4. **Configuration-Driven**: Behavior controlled via configuration, not hard-coded values
5. **Error Handling**: Explicit error returns instead of panics

### Package Structure

```
pkg/zaplog/
├── config/       # Configuration management
├── encoder/      # Line encoding logic
├── formatter/    # Message formatting
├── context/      # Distributed tracing context
├── logger/       # Logger implementation
└── trace/        # Trace information management
```

### Component Overview

- **config**: Manages logger configuration from environment variables or programmatic setup
- **encoder**: Implements custom line-format encoder with trace information support
- **formatter**: Provides message formatting helpers for consistent log messages
- **context**: Manages distributed tracing context with Go's context package
- **logger**: Main logger implementation with interface-based design
- **trace**: Handles trace information including trace ID, span ID, and parent span ID

## Performance

### Benchmarks

Run on: MacBook Pro M1, Go 1.21

| Operation | Ops/sec | Allocation |
|-----------|---------|------------|
| Simple log | 2.5M | 120 B/op |
| With fields | 1.8M | 256 B/op |
| With context | 1.5M | 320 B/op |
| Buffered write | 3.2M | 80 B/op |

### Running Benchmarks

```bash
cd test/benchmark
go test -bench=. -benchmem
```

### Optimization Tips

1. **Use Buffered Writing**: Enabled by default for better performance
2. **Reuse Loggers**: Create loggers once, reuse throughout application
3. **Appropriate Log Levels**: Set minimum level to avoid unnecessary logging
4. **Batch Operations**: Use child loggers with preset fields for related operations

## API Documentation

### Logger Interface

```go
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
```

### Configuration Functions

```go
// Create logger from environment variables
func NewFromEnv() (Logger, error)

// Create logger with custom configuration
func New(cfg *config.Config) (Logger, error)

// Configuration presets
func Default() *Config
func Development() *Config
func Production() *Config
```

### Context Functions

```go
// Create new trace context
func NewContext(ctx context.Context, traceID string) context.Context

// Add span to context
func WithSpan(ctx context.Context) context.Context

// Add parent span to context
func WithParentSpan(ctx context.Context, parentSpanID string) context.Context

// Extract trace information
func GetTraceID(ctx context.Context) string
func GetSpanID(ctx context.Context) string
func GetParentSpanID(ctx context.Context) string
func GetAppName(ctx context.Context) string
```

### Field Constructors

```go
func String(key, val string) Field
func Int(key string, val int) Field
func Int64(key string, val int64) Field
func Error(err error) Field
func Any(key string, val interface{}) Field
```

## Examples

See the `cmd/example/` directory for comprehensive examples:

```bash
cd cmd/example
go run main.go
```

Examples include:
- Logger initialization from environment variables
- Logger with custom configuration
- Different log levels
- Logging with fields
- Using formatters
- Context-aware logging with trace information
- Child loggers
- Performance benchmark
- Trace information extraction
- Error handling

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Specific Package Tests

```bash
go test ./pkg/zaplog/config
go test ./pkg/zaplog/logger
```

### Run Integration Tests

```bash
go test ./test/integration/...
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./test/benchmark/...
```

### Test Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Troubleshooting

### Logs Not Appearing

1. Check log directory exists and is writable
2. Verify `APPNAME` and `HOSTNAME` environment variables are set
3. Ensure log level is configured correctly
4. Check for error messages during logger initialization

### Performance Issues

1. Enable buffered writing (default: enabled)
2. Increase buffer size if needed (default: 256 KB)
3. Adjust log level to reduce verbosity
4. Consider disabling caller information for better performance

### Missing Trace Information

1. Ensure context is created with `context.NewContext()`
2. Pass context to `logger.WithContext()`
3. Verify trace ID is set in context
4. Check that tracing is enabled in configuration

## Migration from Old API

If you're using the old API with global state:

```go
// Old API (deprecated)
zaplog.InitLogger()
logger := zaplog.GetZapLogger(ctx)

// New API (recommended)
logger, err := zaplog.NewFromEnv()
if err != nil {
    panic(err)
}
logWithContext := logger.WithContext(ctx)
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

## Acknowledgments

Built on top of [Uber Zap](https://github.com/uber-go/zap) - a fast, structured, leveled logging library for Go.

Log rotation powered by [Lumberjack](https://github.com/natefinch/lumberjack) - a rolling logger for Go.
