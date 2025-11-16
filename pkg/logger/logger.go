package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

//go:generate mockgen -source=logger.go -destination=mock/mock_logger.go -package=mock
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Sync()
}

type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

type Option func(cfg *zap.Config)

func WithMode(mode string) Option {
	return func(cfg *zap.Config) {
		switch mode {
		case "dev", "development":
			*cfg = zap.NewDevelopmentConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		case "prod", "production":
			*cfg = zap.NewProductionConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

		default:
			fmt.Fprintf(os.Stderr,
				"[logger] warning: unknown mode %q, defaulting to production\n",
				mode,
			)
		}

		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}
}

func WithLevel(l zapcore.Level) Option {
	return func(cfg *zap.Config) {
		cfg.Level = zap.NewAtomicLevelAt(l)
	}
}

func WithDisableCaller(disable bool) Option {
	return func(cfg *zap.Config) {
		cfg.DisableCaller = disable
		if disable {
			cfg.EncoderConfig.CallerKey = ""
		} else if cfg.EncoderConfig.CallerKey == "" {
			cfg.EncoderConfig.CallerKey = "caller"
		}
	}
}

func WithDisableStacktrace(disable bool) Option {
	return func(cfg *zap.Config) {
		cfg.DisableStacktrace = disable
	}
}

func WithSampling(s *zap.SamplingConfig) Option {
	return func(cfg *zap.Config) {
		cfg.Sampling = s
	}
}

func WithEncoding(enc string) Option {
	return func(cfg *zap.Config) {
		cfg.Encoding = enc
	}
}

func WithEncoderConfig(f func(ec *zapcore.EncoderConfig)) Option {
	return func(cfg *zap.Config) {
		f(&cfg.EncoderConfig)
	}
}

func WithOutputPaths(paths ...string) Option {
	return func(cfg *zap.Config) {
		cfg.OutputPaths = paths
	}
}

func WithErrorOutputPaths(paths ...string) Option {
	return func(cfg *zap.Config) {
		cfg.ErrorOutputPaths = paths
	}
}

func WithInitialFields(fields map[string]interface{}) Option {
	return func(cfg *zap.Config) {
		cfg.InitialFields = fields
	}
}

func NewLogger(opts ...Option) (Logger, error) {
	cfg := zap.NewProductionConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	z, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		logger: z,
		sugar:  z.Sugar(),
	}, nil
}

func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *ZapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

func (l *ZapLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
}

func (l *ZapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, keysAndValues...)
}

func (l *ZapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
}

func (l *ZapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

func (l *ZapLogger) Sync() {
	_ = l.logger.Sync()
}
