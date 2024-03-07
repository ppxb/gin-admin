package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"gin-admin/internal/bootstrap"
	"github.com/urfave/cli/v2"
)

func Start() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "start server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "runtime configuration files",
				DefaultText: "configs",
				Value:       "configs",
			},
			&cli.BoolFlag{
				Name:    "daemon",
				Aliases: []string{"d"},
				Usage:   "run as a daemon",
			},
		},
		Action: func(c *cli.Context) error {
			configs := c.String("config")
			if c.Bool("daemon") {
				return startDaemon(configs, c.App.Name)
			}

			return bootstrap.Run(context.Background(), bootstrap.Options{Configs: configs})
		},
	}
}

func startDaemon(configs string, appName string) error {
	bin, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{"start", "-c", configs}
	command := exec.Command(bin, args...)
	err = command.Start()
	if err != nil {
		return err
	}

	pid := command.Process.Pid
	err = os.WriteFile(fmt.Sprintf("%s.lock", appName), []byte(fmt.Sprintf("%d", pid)), 0666)
	if err != nil {
		return err
	}

	fmt.Printf("%s daemon thread started with pid %d \n", appName, pid)
	os.Exit(0)
	return nil
}
