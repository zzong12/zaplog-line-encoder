package config

import (
	"time"
)

// Default returns a configuration with sensible defaults
func Default() *Config {
	return &Config{
		// Application identity (must be set by user)
		AppName:  "",
		Hostname: "",
		Env:      "production",

		// Log output
		LogDirectory:  "log",
		EnableConsole: false,
		EnableFile:    true,
		EnableBuffer:  true,

		// Log rotation
		MaxSize:    100, // 100 MB
		MaxBackups: 10,
		MaxAge:     30, // days
		Compress:   true,

		// Log format
		TimeFormat:       "[2006-01-02 15:04:05.000]",
		EnableCaller:     true,
		EnableStacktrace: false,

		// Log levels
		MinLevel:   "info",
		ErrorLevel: "error",

		// Thread/Process info
		ProcessID: 0, // will be set to os.Getpid()
		ThreadID:  "",

		// Performance
		FlushInterval: 5 * time.Second,
		BufferSize:    256 * 1024, // 256 KB

		// Tracing
		EnableTracing: true,
	}
}

// Development returns a configuration suitable for development
func Development() *Config {
	cfg := Default()
	cfg.Env = "development"
	cfg.EnableConsole = true
	cfg.EnableCaller = true
	cfg.EnableStacktrace = true
	cfg.MinLevel = "debug"
	cfg.Compress = false
	cfg.MaxSize = 10 // smaller files for dev
	cfg.MaxAge = 7   // keep for shorter time
	return cfg
}

// Production returns a configuration suitable for production
func Production() *Config {
	cfg := Default()
	cfg.Env = "production"
	cfg.EnableConsole = false
	cfg.EnableBuffer = true
	cfg.MinLevel = "info"
	cfg.Compress = true
	cfg.MaxSize = 100
	cfg.MaxBackups = 10
	cfg.MaxAge = 30
	return cfg
}
