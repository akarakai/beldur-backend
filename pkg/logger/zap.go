package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func Init() {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	cfg.EncoderConfig.EncodeCaller = func(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(c.FullPath())
	}
	cfg.EncoderConfig.CallerKey = "caller"
	l, err := cfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		panic(err)
	}

	log = l.Sugar()
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

func Debug(msg string, kv ...any) {
	log.Debugw(msg, kv...)
}

func Info(msg string, kv ...any) {
	log.Infow(msg, kv...)
}

func Error(msg string, err error, kv ...any) {
	if err != nil {
		kv = append(kv, "error", err)
	}
	log.Errorw(msg, kv...)
}
