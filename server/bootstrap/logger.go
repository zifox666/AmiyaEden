package bootstrap

import (
	"amiya-eden/global"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 初始化 zap 日志 + lumberjack 日志切割
func InitLogger() {
	cfg := global.Config.Log

	// 日志级别
	level := parseLogLevel(cfg.Level)

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 文件 Core（JSON 格式，带切割）
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		level,
	)

	// 控制台 Core（彩色文本格式）
	consoleEncoder := encoderConfig
	consoleEncoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoder),
		zapcore.AddSync(os.Stdout),
		level,
	)

	core := zapcore.NewTee(fileCore, consoleCore)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	global.Logger = logger

	zap.ReplaceGlobals(logger)
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
