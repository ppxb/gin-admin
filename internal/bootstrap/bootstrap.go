package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"

	"gin-admin/internal/config"
	"gin-admin/internal/utiliity/prom"
	"gin-admin/pkg/logger"
	"gin-admin/pkg/utils"
)

type Options struct {
	Configs string
}

func Run(ctx context.Context, options Options) error {
	// flush logger
	defer func() {
		if err := zap.L().Sync(); err != nil {
			fmt.Printf("failed to sync zap logger: %s \n", err.Error())
		}
	}()

	// init configuration
	config.MustLoad(options.Configs)

	// init logger
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

	// start pprof server
	if addr := config.C.Global.PprofAddr; addr != "" {
		logger.Context(ctx).Info("pprof server is listening on " + addr)
		go func() {
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				logger.Context(ctx).Error("failed to listen pprof server", zap.Error(err))
			}
		}()
	}

	// init global prometheus metrics
	prom.Init()

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
