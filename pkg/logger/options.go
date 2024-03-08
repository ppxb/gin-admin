package logger

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	l Options
}

type Options struct {
	Debug      bool
	Level      string
	CallerSkip int
	File       LogFile
	Hooks      []*HookOptions
}

type LogFile struct {
	Enable     bool
	Path       string
	MaxSize    int
	MaxBackups int
}

type HookOptions struct {
	Enable    bool
	Level     string
	Type      string
	MaxBuffer int
	MaxThread int
	Options   map[string]string
	Extra     map[string]string
}

type HookHandlerFunc func(ctx context.Context, hookOptions *HookOptions) (*Hook, error)

func WithConfig(ctx context.Context, options *Options, handlers ...HookHandlerFunc) (func(), error) {
	var zapConfig zap.Config

	if options.Debug {
		options.Level = "debug"
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	level, err := zapcore.ParseLevel(options.Level)
	if err != nil {
		return nil, err
	}
	zapConfig.Level.SetLevel(level)

	var (
		logger   *zap.Logger
		cleanFns []func()
	)

	if options.File.Enable {
		filename := options.File.Path
		_ = os.Mkdir(filepath.Dir(filename), 0777)
		writer := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    options.File.MaxSize,
			MaxBackups: options.File.MaxBackups,
			Compress:   false,
			LocalTime:  true,
		}
		cleanFns = append(cleanFns, func() {
			_ = writer.Close()
		})
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zapConfig.EncoderConfig),
			zapcore.AddSync(writer),
			level,
		)
		logger = zap.New(core)
	} else {
		logger, err = zapConfig.Build()
		if err != nil {
			return nil, err
		}
	}

	skip := options.CallerSkip
	if skip <= 0 {
		skip = 2
	}

	logger = logger.WithOptions(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(skip),
	)

	for _, h := range options.Hooks {
		if !h.Enable || len(handlers) == 0 {
			continue
		}

		writer, err := handlers[0](ctx, h)
		if err != nil {
			return nil, err
		} else if writer == nil {
			continue
		}

		cleanFns = append(cleanFns, func() {
			writer.Flush()
		})
		hookLevel := zap.NewAtomicLevel()
		if level, err := zapcore.ParseLevel(h.Level); err == nil {
			hookLevel.SetLevel(level)
		} else {
			hookLevel.SetLevel(zap.InfoLevel)
		}

		hookEncoder := zap.NewProductionEncoderConfig()
		hookEncoder.EncodeTime = zapcore.EpochMillisTimeEncoder
		hookEncoder.EncodeDuration = zapcore.MillisDurationEncoder
		hookCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(hookEncoder),
			zapcore.AddSync(writer),
			hookLevel,
		)

		logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, hookCore)
		}))
	}

	zap.ReplaceGlobals(logger)
	return func() {
		for _, f := range cleanFns {
			f()
		}
	}, nil
}
