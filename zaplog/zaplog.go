package zaplog

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	_zapLogger *zap.Logger
)

var (
	appName    = os.Getenv("APPNAME") // xxxx
	serverName = os.Getenv("HOSTNAME")
)

func InitLogger() {
	encoder := NewLineEncoder(zapcore.EncoderConfig{
		TimeKey:          zapcore.OmitKey,
		LevelKey:         _log_field_split,
		NameKey:          zapcore.OmitKey,
		CallerKey:        _log_field_split,
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       zapcore.OmitKey,
		StacktraceKey:    zapcore.OmitKey,
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout("[2006-01-02 15:04:05.000]"),
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "|",
	}, true)

	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { //error级别
		return lev >= zap.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool { //info和debug级别,debug级别是最低的
		return lev < zap.ErrorLevel && lev >= zap.DebugLevel
	})

	infoFileWriteSyncer := getInfoWriterSyncer()
	errorFileWriteSyncer := getErrorWriterSyncer()

	infoFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(infoFileWriteSyncer), lowPriority)
	errorFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(errorFileWriteSyncer), highPriority)
	stdoutFileCore := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), lowPriority)

	coreArr := []zapcore.Core{
		infoFileCore,
		errorFileCore,
		stdoutFileCore,
	}

	_zapLogger = zap.New(zapcore.NewTee(coreArr...), zap.AddCaller())
}

func getInfoWriterSyncer() zapcore.WriteSyncer {
	fileName := getLogFileName("proc")
	infoLumberIO := &lumberjack.Logger{
		Filename:   "log/" + fileName,
		MaxSize:    2, // megabytes
		MaxBackups: 2,
		MaxAge:     1,     // days
		Compress:   false, //Compress确定是否应该使用gzip压缩已旋转的日志文件。默认值是不执行压缩。
		LocalTime:  true,
	}
	return zapcore.AddSync(infoLumberIO)
}

func getErrorWriterSyncer() zapcore.WriteSyncer {
	fileName := getLogFileName("err")
	lumberWriteSyncer := &lumberjack.Logger{
		Filename:   "log/" + fileName,
		MaxSize:    2, // megabytes
		MaxBackups: 2,
		MaxAge:     1,     // days
		Compress:   false, //Compress确定是否应该使用gzip压缩已旋转的日志文件。默认值是不执行压缩。
	}
	return zapcore.AddSync(lumberWriteSyncer)
}

func getLogFileName(prefix string) string {
	return fmt.Sprintf("%s-v01-%s-%s.log", prefix, appName, serverName)
}
