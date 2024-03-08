package bootstrap

import (
	"context"

	"github.com/spf13/cast"

	"gin-admin/internal/config"
	"gin-admin/pkg/gormx"
	"gin-admin/pkg/logger"
)

func initLoggerHook(_ context.Context, cfg *logger.HookOptions) (*logger.Hook, error) {
	extra := cfg.Extra
	if extra == nil {
		extra = make(map[string]string)
	}
	extra["appname"] = config.C.Global.AppName

	switch cfg.Type {
	case "gorm":
		db, err := gormx.New(gormx.Config{
			Debug:        cast.ToBool(cfg.Options["Debug"]),
			DBType:       cast.ToString(cfg.Options["DBType"]),
			DSN:          cast.ToString(cfg.Options["DSN"]),
			MaxLifeTime:  cast.ToInt(cfg.Options["MaxLifeTime"]),
			MaxIdleTime:  cast.ToInt(cfg.Options["MaxIdleTime"]),
			MaxOpenConns: cast.ToInt(cfg.Options["MaxOpenConns"]),
			MaxIdleConns: cast.ToInt(cfg.Options["MaxIdleConns"]),
			TablePrefix:  config.C.DataSource.DB.TablePrefix,
		})
		if err != nil {
			return nil, err
		}

		return logger.NewHook(logger.NewGormHook(db),
			logger.WithExtra(cfg.Extra),
			logger.WithMaxJobs(cfg.MaxBuffer),
			logger.WithMaxWorkers(cfg.MaxThread),
		), nil
	default:
		return nil, nil
	}
}
