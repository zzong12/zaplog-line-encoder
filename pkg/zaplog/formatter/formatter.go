package formatter

import "fmt"

const (
	// Log message format templates
	_logKeyProcMsg   = "#KEY_PROC:KEYNODE=%s, %s"
	_logErrorMsg     = "#EX_ERR:ERR=%s,EMSG=%s"
	_logWarnMsg      = "#WARN:DESC=%s"
	_logInfoMsg      = "#INFO:%s"
	_logDebugMsg     = "#DEBUG:%s"
	_logCustomPrefix = "#CUSTOM:%s"
)

// Formatter defines the interface for message formatting
type Formatter interface {
	FormatKeyProc(keyNode, message string) string
	FormatError(errCode, errMsg string) string
	FormatWarning(desc string) string
	FormatInfo(message string) string
	FormatDebug(message string) string
	Custom(template string, args ...interface{}) string
}

// LineFormatter implements the Formatter interface with predefined formats
type LineFormatter struct {
	templates map[string]string
}

// NewLineFormatter creates a new LineFormatter with default templates
func NewLineFormatter() *LineFormatter {
	return &LineFormatter{
		templates: make(map[string]string),
	}
}

// FormatKeyProc formats a key-process message
// Format: #KEY_PROC:KEYNODE={keyNode}, {message}
func (f *LineFormatter) FormatKeyProc(keyNode, message string) string {
	return fmt.Sprintf(_logKeyProcMsg, keyNode, message)
}

// FormatError formats an error message
// Format: #EX_ERR:ERR={errCode},EMSG={errMsg}
func (f *LineFormatter) FormatError(errCode, errMsg string) string {
	return fmt.Sprintf(_logErrorMsg, errCode, errMsg)
}

// FormatWarning formats a warning message
// Format: #WARN:DESC={desc}
func (f *LineFormatter) FormatWarning(desc string) string {
	return fmt.Sprintf(_logWarnMsg, desc)
}

// FormatInfo formats an info message
// Format: #INFO:{message}
func (f *LineFormatter) FormatInfo(message string) string {
	return fmt.Sprintf(_logInfoMsg, message)
}

// FormatDebug formats a debug message
// Format: #DEBUG:{message}
func (f *LineFormatter) FormatDebug(message string) string {
	return fmt.Sprintf(_logDebugMsg, message)
}

// Custom formats a message with a custom template
func (f *LineFormatter) Custom(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}

// RegisterTemplate registers a custom template for reuse
func (f *LineFormatter) RegisterTemplate(name, template string) {
	if f.templates == nil {
		f.templates = make(map[string]string)
	}
	f.templates[name] = template
}

// UseTemplate formats a message using a registered template
func (f *LineFormatter) UseTemplate(name string, args ...interface{}) string {
	if template, ok := f.templates[name]; ok {
		return fmt.Sprintf(template, args...)
	}
	return fmt.Sprintf("%s %v", name, args)
}

// GetKeyProcMsg is a convenience function for formatting key-process messages
// Deprecated: Use formatter.FormatKeyProc() instead
func GetKeyProcMsg(keyNode, message string) string {
	return fmt.Sprintf(_logKeyProcMsg, keyNode, message)
}

// GetErrorMsg is a convenience function for formatting error messages
// Deprecated: Use formatter.FormatError() instead
func GetErrorMsg(errCode, errMsg string) string {
	return fmt.Sprintf(_logErrorMsg, errCode, errMsg)
}

// GetWarnMsg is a convenience function for formatting warning messages
// Deprecated: Use formatter.FormatWarning() instead
func GetWarnMsg(warnDesc string) string {
	return fmt.Sprintf(_logWarnMsg, warnDesc)
}

// GetInfoMsg is a convenience function for formatting info messages
func GetInfoMsg(message string) string {
	return fmt.Sprintf(_logInfoMsg, message)
}

// GetDebugMsg is a convenience function for formatting debug messages
func GetDebugMsg(message string) string {
	return fmt.Sprintf(_logDebugMsg, message)
}

// DefaultFormatter returns the default formatter instance
var DefaultFormatter = NewLineFormatter()

// Package-level convenience functions that use the default formatter

// FormatKeyProc formats a key-process message using the default formatter
func FormatKeyProc(keyNode, message string) string {
	return DefaultFormatter.FormatKeyProc(keyNode, message)
}

// FormatError formats an error message using the default formatter
func FormatError(errCode, errMsg string) string {
	return DefaultFormatter.FormatError(errCode, errMsg)
}

// FormatWarning formats a warning message using the default formatter
func FormatWarning(desc string) string {
	return DefaultFormatter.FormatWarning(desc)
}

// FormatInfo formats an info message using the default formatter
func FormatInfo(message string) string {
	return DefaultFormatter.FormatInfo(message)
}

// FormatDebug formats a debug message using the default formatter
func FormatDebug(message string) string {
	return DefaultFormatter.FormatDebug(message)
}
