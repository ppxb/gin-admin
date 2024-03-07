package bootstrap

import (
	"context"
	"os"

	"gin-admin/internal/config"
	"gin-admin/pkg/logger"
	"gin-admin/pkg/utils"
	"go.uber.org/zap"
)

type Options struct {
	Configs string
}

func Run(ctx context.Context, options Options) error {
	config.MustLoad(options.Configs)

	cleanLoggerFn, err := logger.WithConfig(ctx, &config.C.Logger)
	if err != nil {
		return err
	}
	ctx = logger.NewTag(ctx, logger.TagKeyMain)
	logger.Context(ctx).Info(
		"starting service...",
		zap.String("version", config.C.Global.Version),
		zap.Int("pid", os.Getpid()),
	)

	return utils.Run(ctx, func(ctx context.Context) (func(), error) {
		cleanHttpServerFn, err := httpServerListener(ctx)
		if err != nil {
			return nil, err
		}

		return func() {
			if cleanHttpServerFn != nil {
				cleanHttpServerFn()
			}
			if cleanLoggerFn != nil {
				cleanLoggerFn()
			}
		}, nil
	})
}
