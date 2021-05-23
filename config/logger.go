package config

import (
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/TheZeroSlave/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//NewLogger creates a new *zap.Logger
func NewLogger(appConfig *App) *zap.Logger {
	if !appConfig.IsProduction {
		return developmentLogger()
	}

	return productionLogger(appConfig)
}

func developmentLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := cfg.Build(zap.AddCaller())
	if err != nil {
		panic(err)
	}

	return logger
}

func productionLogger(appConfig *App) *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	cfg := zapsentry.Configuration{
		Tags:         map[string]string{},
		Level:        zapcore.InfoLevel,
		FlushTimeout: 2 * time.Second,
		Hub:          nil,
	}

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:         appConfig.SentryDSN,
		Release:     "",
		Environment: appConfig.Environment.String(),
	})
	if err != nil {
		panic(err)
	}

	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(client))
	if err != nil {
		panic(err)
	}

	return zapsentry.AttachCoreToLogger(core, logger)
}
