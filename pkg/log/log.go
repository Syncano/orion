package log

import (
	"os"

	"github.com/TheZeroSlave/zapsentry"
	ltsv "github.com/hnakamur/zap-ltsv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc/grpclog"
)

var (
	logger, sentryLogger *zap.Logger
)

// Init sets up a logger.
func Init(dsn string, debug bool) error {
	var config zap.Config

	if os.Getenv("FORCE_TERM") == "1" || terminal.IsTerminal(int(os.Stdout.Fd())) {
		config = zap.NewDevelopmentConfig()
		if debug {
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
	} else {
		if err := ltsv.RegisterLTSVEncoder(); err != nil {
			return err
		}
		config = ltsv.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		if debug {
			config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		}
	}

	var err error
	logger, err = config.Build()
	if err != nil {
		return err
	}
	sentryLogger, err = addSentryLogger(logger, dsn)

	// Set grpc logger.
	var zapgrpcOpts []zapgrpc.Option
	if debug {
		zapgrpcOpts = append(zapgrpcOpts, zapgrpc.WithDebug())
	}
	grpclog.SetLogger(zapgrpc.NewLogger(logger, zapgrpcOpts...))
	return err
}

func addSentryLogger(log *zap.Logger, dsn string) (*zap.Logger, error) {
	cfg := zapsentry.Configuration{
		Level: zapcore.ErrorLevel,
	}
	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromDSN(dsn))
	return zapsentry.AttachCoreToLogger(core, log), err
}

// RawLogger returns zap logger without sentry.
func RawLogger() *zap.Logger {
	return logger
}

// Logger returns zap logger with sentry support on error.
func Logger() *zap.Logger {
	return sentryLogger
}

// Sync syncs both loggers.
func Sync() {
	logger.Sync()       // nolint: errcheck
	sentryLogger.Sync() // nolint: errcheck
}
