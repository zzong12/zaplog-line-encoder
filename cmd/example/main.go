package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/config"
	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/context"
	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/formatter"
	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/logger"
)

func main() {
	// Set required environment variables
	os.Setenv("APPNAME", "example-app")
	os.Setenv("HOSTNAME", "localhost")

	// Example 1: Initialize logger from environment variables
	fmt.Println("=== Example 1: Logger from environment ===")
	log1, err := logger.NewFromEnv()
	if err != nil {
		panic(err)
	}
	log1.Info("Logger initialized from environment variables")
	log1.Sync()

	// Example 2: Initialize logger with custom configuration
	fmt.Println("\n=== Example 2: Logger with custom config ===")
	cfg := &config.Config{
		AppName:        "example-app",
		Hostname:       "localhost",
		LogDirectory:   "log",
		EnableConsole:  true,
		EnableFile:     true,
		EnableBuffer:   true,
		MaxSize:        10,  // 10 MB
		MaxBackups:     3,
		MaxAge:         7,   // 7 days
		Compress:       false,
		TimeFormat:     "[2006-01-02 15:04:05.000]",
		EnableCaller:   true,
		MinLevel:       "info",
		EnableTracing:  true,
	}

	log2, err := logger.New(cfg)
	if err != nil {
		panic(err)
	}
	log2.Info("Logger initialized with custom configuration")
	log2.Sync()

	// Example 3: Using different log levels
	fmt.Println("\n=== Example 3: Different log levels ===")
	log2.Debug("This is a debug message (may not appear)")
	log2.Info("This is an info message")
	log2.Warn("This is a warning message")
	log2.Error("This is an error message")

	// Example 4: Logging with fields
	fmt.Println("\n=== Example 4: Logging with fields ===")
	log2.Info("User login",
		logger.String("user_id", "12345"),
		logger.String("ip", "192.168.1.1"),
		logger.Int("port", 8080),
	)
	log2.Error("Database connection failed",
		logger.String("host", "localhost:5432"),
		logger.String("error", "connection timeout"),
	)
	log2.Sync()

	// Example 5: Using formatters
	fmt.Println("\n=== Example 5: Using message formatters ===")
	log2.Info(formatter.FormatKeyProc("ProcessOrder", "Order processed successfully"))
	log2.Warn(formatter.FormatWarning("High memory usage detected"))
	log2.Error(formatter.FormatError("DB001", "Failed to connect to database"))
	log2.Sync()

	// Example 6: Context-aware logging with trace information
	fmt.Println("\n=== Example 6: Context-aware logging ===")
	ctx := context.NewContext(context.Background(), "trace-123-456")
	ctx = context.WithAppName(ctx, "example-app")

	logWithContext := log2.WithContext(ctx)
	logWithContext.Info("Request received",
		logger.String("method", "GET"),
		logger.String("path", "/api/users"),
	)

	// Create a child span
	ctx = context.WithSpan(ctx)
	logWithContext = log2.WithContext(ctx)
	logWithContext.Info("Processing request")

	// Add parent span context
	ctx = context.WithParentSpan(ctx, "parent-span-789")
	logWithContext = log2.WithContext(ctx)
	logWithContext.Info("Database query executed")
	log2.Sync()

	// Example 7: Using With() to create child loggers
	fmt.Println("\n=== Example 7: Child loggers ===")
	requestLogger := log2.With(
		logger.String("request_id", "req-123"),
		logger.String("user_id", "user-456"),
	)
	requestLogger.Info("Processing request")
	requestLogger.Warn("Slow query detected")
	requestLogger.Error("Validation failed")
	log2.Sync()

	// Example 8: Performance benchmark
	fmt.Println("\n=== Example 8: Performance benchmark ===")
	start := time.Now()
	for i := 0; i < 10000; i++ {
		log2.Info(formatter.FormatKeyProc("Benchmark", fmt.Sprintf("Message %d", i)))
	}
	duration := time.Since(start)
	fmt.Printf("Logged 10,000 messages in %v\n", duration)
	fmt.Printf("Average: %.2f μs per message\n", float64(duration.Microseconds())/10000)
	log2.Sync()

	// Example 9: Extract trace information from context
	fmt.Println("\n=== Example 9: Trace information ===")
	ctx = context.NewContext(context.Background(), "trace-abc-123")
	ctx = context.WithSpan(ctx)

	traceID := context.GetTraceID(ctx)
	spanID := context.GetSpanID(ctx)
	appName := context.GetAppName(ctx)

	fmt.Printf("Trace ID: %s\n", traceID)
	fmt.Printf("Span ID: %s\n", spanID)
	fmt.Printf("App Name: %s\n", appName)
	fmt.Printf("Trace String: %s\n", context.TraceString(ctx))

	// Example 10: Error handling
	fmt.Println("\n=== Example 10: Error handling ===")
	errCfg := &config.Config{
		AppName: "", // Missing required field
	}
	_, err = logger.New(errCfg)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	fmt.Println("\n=== All examples completed ===")
	fmt.Println("Check the 'log/' directory for generated log files")
}
