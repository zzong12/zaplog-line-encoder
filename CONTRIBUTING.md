# Contributing to zaplog-line-encoder

Thank you for your interest in contributing to zaplog-line-encoder! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow:
- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the community
- Show empathy towards other community members

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**: Summarize the issue
- **Steps to reproduce**: Detailed steps to reproduce the behavior
- **Expected behavior**: What you expected to happen
- **Actual behavior**: What actually happened
- **Environment**: Go version, OS, and relevant configuration
- **Logs/error messages**: Relevant log output or error messages
- **Reproduction code**: Minimal code example demonstrating the issue

### Suggesting Enhancements

Enhancement suggestions are welcome! Please include:

- **Clear title and description**: Summarize the enhancement
- **Use cases**: Describe specific use cases for the enhancement
- **Proposed solution**: How you envision the enhancement working
- **Alternatives considered**: Other approaches you've considered
- **Additional context**: Any other relevant information

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Install dependencies**: `go mod download`
3. **Make your changes** following our coding standards
4. **Add tests** for new functionality
5. **Ensure all tests pass**: `go test ./...`
6. **Update documentation** as needed
7. **Commit your changes** with clear commit messages
8. **Push to your fork** and submit a pull request

### Commit Message Format

Follow these guidelines for commit messages:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions or changes
- `chore`: Maintenance tasks

**Example:**
```
feat(logger): add context-aware logging

Implement WithContext method to support distributed tracing
with Go's context package.

Closes #123
```

## Development Setup

### Prerequisites

- Go 1.18 or higher
- Git

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/zzong12/zaplog-line-encoder.git
   cd zaplog-line-encoder
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Run tests**:
   ```bash
   go test ./...
   ```

4. **Run example**:
   ```bash
   cd cmd/example
   go run main.go
   ```

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting
- Follow standard Go project layout
- Use meaningful names for variables, functions, and packages

### Documentation

- Add godoc comments for all exported types, functions, and constants
- Include usage examples in godoc comments
- Update README.md for user-facing changes
- Update CHANGELOG.md for notable changes

### Testing

- Write unit tests for all new functionality
- Aim for high test coverage (target: >80%)
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error cases and edge cases

### Example Test Structure

```go
func TestFunctionName(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:    "success case",
			input:   "valid input",
			want:    "expected output",
			wantErr: false,
		},
		{
			name:    "error case",
			input:   "invalid input",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FunctionName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FunctionName() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

## Running Tests

### All Tests

```bash
go test ./...
```

### Specific Package

```bash
go test ./pkg/zaplog/logger
```

### With Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Benchmarks

```bash
go test -bench=. -benchmem ./test/benchmark/...
```

### Race Detection

```bash
go test -race ./...
```

## Project Structure

```
zaplog-line-encoder/
├── cmd/
│   └── example/          # Example application
├── pkg/
│   └── zaplog/
│       ├── config/       # Configuration management
│       ├── encoder/      # Line encoding logic
│       ├── formatter/    # Message formatting
│       ├── context/      # Distributed tracing context
│       ├── logger/       # Logger implementation
│       └── trace/        # Trace information management
├── test/
│   ├── benchmark/        # Performance benchmarks
│   └── integration/      # Integration tests
├── .github/
│   └── workflows/        # CI/CD workflows
├── go.mod
├── go.sum
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
└── LICENSE
```

## Pull Request Process

1. **Ensure all tests pass**: `go test ./...`
2. **Update documentation**: README, godoc, CHANGELOG
3. **Format code**: `gofmt -w .`
4. **Run linters**: `go vet ./...`
5. ** squash commits** into a single coherent commit
6. **Push to your fork** and create a pull request

### PR Description Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] All tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] No new warnings generated
```

## Release Process

Maintainers follow this process for releases:

1. Update version in `go.mod`
2. Update CHANGELOG.md
3. Create git tag
4. Push tag to trigger GitHub Actions
5. Create GitHub release

## Questions?

If you have questions, please:
- Open an issue for bugs or feature requests
- Start a discussion for general questions
- Check existing issues and discussions first

Thank you for contributing to zaplog-line-encoder!
