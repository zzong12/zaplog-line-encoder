package zaplog

import (
	"context"
	"os"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	serverTransportKey struct{}

	serverTraceIdKey     = "trace_id"
	serverSpanIdKey      = "span_id"
	serverProcesserIdKey = "processer_id"
	serverThreadIdKey    = "thread_id"
	serverAppNameKey     = "app_name"

	serverProcesserIdVal = os.Getpid()
	serverThreadIdVal    = ""
	serverAppNameVal     = "cbse"
)

func GetLogContext(ctx context.Context, traceId string) context.Context {
	serverTransportVal := map[string]string{
		serverTraceIdKey: traceId,
		serverSpanIdKey:  uuid.NewString(),
	}
	return context.WithValue(ctx, serverTransportKey, serverTransportVal)
}

func GetZapLogger(ctx context.Context) *zap.Logger {
	traceId, spanId := getTraceFields(ctx)
	return _zapLogger.With(
		zap.String(zapcore.OmitKey, serverAppNameVal),
		zap.String(_log_field_split, spanId),
		zap.String(_log_field_split, traceId),
		zap.Int(_log_field_split, serverProcesserIdVal),
		zap.String(_log_field_split, serverThreadIdVal),
	)
}

func getTraceFields(ctx context.Context) (traceId string, spanId string) {
	serverTransportVal := ctx.Value(serverTransportKey)
	if serverTransportVal == nil {
		return "", ""
	}
	serverTransport := serverTransportVal.(map[string]string)
	return serverTransport[serverTraceIdKey], serverTransport[serverSpanIdKey]
}
