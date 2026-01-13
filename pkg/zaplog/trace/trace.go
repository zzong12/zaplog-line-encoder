package trace

import (
	"fmt"
	"os"
	"runtime"

	"github.com/google/uuid"
)

// TraceInfo holds distributed tracing information
type TraceInfo struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
	ProcessID    int
	ThreadID     int
	AppName      string
}

// NewTraceInfo creates a new TraceInfo with a new span ID
func NewTraceInfo(appName string) *TraceInfo {
	return &TraceInfo{
		SpanID:    uuid.New().String(),
		ProcessID: os.Getpid(),
		ThreadID:  getGoroutineID(),
		AppName:   appName,
	}
}

// NewTraceInfoWithTrace creates a new TraceInfo with the given trace ID
func NewTraceInfoWithTrace(appName, traceID string) *TraceInfo {
	info := NewTraceInfo(appName)
	info.TraceID = traceID
	return info
}

// WithSpan creates a new TraceInfo with a new span ID but same trace ID
func (t *TraceInfo) WithSpan() *TraceInfo {
	newInfo := &TraceInfo{
		TraceID:      t.TraceID,
		SpanID:       uuid.New().String(),
		ParentSpanID: t.SpanID,
		ProcessID:    t.ProcessID,
		ThreadID:     getGoroutineID(),
		AppName:      t.AppName,
	}
	return newInfo
}

// WithParentSpan sets the parent span ID
func (t *TraceInfo) WithParentSpan(parentSpanID string) *TraceInfo {
	t.ParentSpanID = parentSpanID
	return t
}

// String returns a string representation of the trace info
func (t *TraceInfo) String() string {
	return fmt.Sprintf("trace=%s span=%s parent=%s app=%s pid=%d",
		t.TraceID, t.SpanID, t.ParentSpanID, t.AppName, t.ProcessID)
}

// getGoroutineID returns the current goroutine ID
// Note: This is a rough approximation using runtime.NumGoroutine
// For accurate goroutine IDs, we would need to use runtime.Stack
func getGoroutineID() int {
	// For now, use the number of goroutines as a proxy
	// In a real implementation, you might want to extract the actual goroutine ID
	// from runtime.Stack(), but that's more expensive
	return runtime.NumGoroutine()
}

// GetActualGoroutineID returns the actual goroutine ID (more expensive)
// Use this for debugging or when you need the exact goroutine ID
func GetActualGoroutineID() int64 {
	// This is a more expensive operation that gives us the actual goroutine ID
	// We use runtime.Stack to get the stack trace and parse the goroutine ID
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := buf[:n]
	// Parse "goroutine 123 [running]:" format
	var id int64
	fmt.Sscanf(string(idField), "goroutine %d", &id)
	return id
}
