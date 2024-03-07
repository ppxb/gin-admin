package cmd

import (
	"os"

	"gin-admin/internal/config"
	"github.com/urfave/cli/v2"
)

func Run() {
	app := &cli.App{
		Name:        config.AppName,
		Version:     config.AppVersion,
		Description: config.AppDescription,
		Commands: []*cli.Command{
			Start(),
			Stop(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
