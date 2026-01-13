package context

import (
	"context"
	"fmt"

	"github.com/zzong12/zaplog-line-encoder/pkg/zaplog/trace"
)

// contextKey is the key type for storing trace info in context
type contextKey struct{}

var traceKey = contextKey{}

// NewContext creates a new context with trace information
func NewContext(ctx context.Context, traceID string) context.Context {
	traceInfo := trace.NewTraceInfoWithTrace("", traceID)
	return context.WithValue(ctx, traceKey, traceInfo)
}

// NewContextWithApp creates a new context with trace information and app name
func NewContextWithApp(ctx context.Context, appName, traceID string) context.Context {
	traceInfo := trace.NewTraceInfoWithTrace(appName, traceID)
	return context.WithValue(ctx, traceKey, traceInfo)
}

// NewContextWithTraceInfo creates a new context with existing trace info
func NewContextWithTraceInfo(ctx context.Context, traceInfo *trace.TraceInfo) context.Context {
	return context.WithValue(ctx, traceKey, traceInfo)
}

// FromContext extracts trace information from context
// Returns nil if no trace info is found
func FromContext(ctx context.Context) *trace.TraceInfo {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(traceKey)
	if val == nil {
		return nil
	}
	traceInfo, ok := val.(*trace.TraceInfo)
	if !ok {
		return nil
	}
	return traceInfo
}

// MustFromContext extracts trace information from context
// Panics if no trace info is found
func MustFromContext(ctx context.Context) *trace.TraceInfo {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		panic("trace info not found in context")
	}
	return traceInfo
}

// WithSpan creates a new context with a new span ID
func WithSpan(ctx context.Context) context.Context {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		return ctx
	}
	newTraceInfo := traceInfo.WithSpan()
	return context.WithValue(ctx, traceKey, newTraceInfo)
}

// WithParentSpan sets the parent span ID in the context
func WithParentSpan(ctx context.Context, parentSpanID string) context.Context {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		return ctx
	}
	traceInfo.WithParentSpan(parentSpanID)
	return context.WithValue(ctx, traceKey, traceInfo)
}

// WithAppName sets the app name in the context
func WithAppName(ctx context.Context, appName string) context.Context {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		// Create new trace info if none exists
		traceInfo = trace.NewTraceInfo(appName)
		return context.WithValue(ctx, traceKey, traceInfo)
	}
	traceInfo.AppName = appName
	return context.WithValue(ctx, traceKey, traceInfo)
}

// GetTraceID returns the trace ID from context
// Returns empty string if no trace info is found
func GetTraceID(ctx context.Context) string {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		return ""
	}
	return traceInfo.TraceID
}

// GetSpanID returns the span ID from context
// Returns empty string if no span info is found
func GetSpanID(ctx context.Context) string {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		return ""
	}
	return traceInfo.SpanID
}

// GetParentSpanID returns the parent span ID from context
// Returns empty string if no parent span info is found
func GetParentSpanID(ctx context.Context) string {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		return ""
	}
	return traceInfo.ParentSpanID
}

// GetAppName returns the app name from context
// Returns empty string if no app name is found
func GetAppName(ctx context.Context) string {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		return ""
	}
	return traceInfo.AppName
}

// TraceString returns a formatted string with trace information
func TraceString(ctx context.Context) string {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		return "no-trace"
	}
	return traceInfo.String()
}

// ValidateContext checks if the context has valid trace information
func ValidateContext(ctx context.Context) error {
	traceInfo := FromContext(ctx)
	if traceInfo == nil {
		return fmt.Errorf("no trace information in context")
	}
	if traceInfo.TraceID == "" {
		return fmt.Errorf("trace ID is empty")
	}
	if traceInfo.SpanID == "" {
		return fmt.Errorf("span ID is empty")
	}
	return nil
}
