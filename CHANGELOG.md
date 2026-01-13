# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Complete project refactoring with modular package structure
- New configuration system with environment variable support
- Distributed tracing with context-aware logging
- Custom line-format encoder with parent-span-id support
- Flexible configuration presets (Default, Development, Production)
- Comprehensive README with examples and API documentation
- MIT License

### Changed

- Replaced global state with interface-based design
- Replaced panic() calls with explicit error returns
- Fixed hardcoded app_name issue (now configurable)
- Fixed empty thread_id issue (now uses goroutine count)
- Improved code organization with separate packages for config, encoder, formatter, context, logger, and trace

### Fixed

- Fixed hardcoded "cbse" app_name in trace.go:23
- Fixed empty thread_id in helper.go:22
- Added parent-span-id field support in encoder
- Improved error handling throughout the codebase

## [0.1.0] - 2024-01-13

### Added

- Initial release
- Basic line-format logging with Uber Zap
- Log rotation with lumberjack
- Simple context management
- Message formatters
- Example usage in main.go

[Unreleased]: https://github.com/zzong12/zaplog-line-encoder/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/zzong12/zaplog-line-encoder/releases/tag/v0.1.0
