package main

import (
	"context"

	"github.com/zzong12/zaplog-line-encoder/zaplog"

	"go.uber.org/zap"
)

func main() {
	zaplog.InitLogger()

	traceId := "1234567890"
	ctx := zaplog.GetLogContext(context.Background(), traceId)

	logger := zaplog.GetZapLogger(ctx)
	sugar := logger.Sugar()

	for i := 0; i < 10; i++ {
		logger.Info("Logger info", zap.Int("number=", i))
		sugar.Warn("SugaredLogger warn", "num=", i)
	}
}
