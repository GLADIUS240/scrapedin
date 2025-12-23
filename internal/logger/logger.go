package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(level, format string) (*zap.Logger, error) {

	logLevel := zapcore.InfoLevel
	_ = logLevel.Set(level)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var encoder zapcore.Encoder
	if format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	// 🔥 File rotation
	logFile := os.Getenv("SCRAPEDIN_LOGGING_FILE_PATH")
	var fileSync zapcore.WriteSyncer

	if logFile != "" {
		fileSync = zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10, // MB
			MaxBackups: 10,
			MaxAge:     0, // keep forever until backups exceed
			Compress:   false,
		})
	}

	consoleSync := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleSync, logLevel),
		func() zapcore.Core {
			if fileSync != nil {
				return zapcore.NewCore(
					zapcore.NewJSONEncoder(encoderCfg),
					fileSync,
					logLevel,
				)
			}
			return zapcore.NewNopCore()
		}(),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, nil
}
