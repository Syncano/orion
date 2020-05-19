package log

import (
	"os"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/getsentry/sentry-go"
	ltsv "github.com/hnakamur/zap-ltsv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
)

type Logger struct {
	logger, sentryLogger *zap.Logger
}

// Init sets up a logger.
func New(debug bool, sc *sentry.Client) (*Logger, error) {
	var config zap.Config

	if os.Getenv("FORCE_TERM") == "1" || terminal.IsTerminal(int(os.Stdout.Fd())) {
		config = zap.NewDevelopmentConfig()
		if debug {
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	} else {
		if err := ltsv.RegisterLTSVEncoder(); err != nil {
			return nil, err
		}
		config = ltsv.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		if debug {
			config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		}
	}

	var err error

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	sentryLogger, err := addSentryLogger(logger, sc)

	return &Logger{
		logger:       logger,
		sentryLogger: sentryLogger,
	}, err
}

func addSentryLogger(log *zap.Logger, sc *sentry.Client) (*zap.Logger, error) {
	cfg := zapsentry.Configuration{
		Level: zapcore.ErrorLevel,
	}
	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(sc))

	return zapsentry.AttachCoreToLogger(core, log), err
}

// RawLogger returns zap logger without sentry.
func (l *Logger) RawLogger() *zap.Logger {
	return l.logger
}

// Logger returns zap logger with sentry support on error.
func (l *Logger) Logger() *zap.Logger {
	return l.sentryLogger
}

// Sync syncs both loggers.
func (l *Logger) Sync() {
	l.logger.Sync()       // nolint: errcheck
	l.sentryLogger.Sync() // nolint: errcheck
}
