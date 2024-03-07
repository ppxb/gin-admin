package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

func Stop() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "stop server",
		Action: func(c *cli.Context) error {
			appName := c.App.Name
			lockFile := fmt.Sprintf("%s.lock", appName)
			pid, err := os.ReadFile(lockFile)
			if err != nil {
				return err
			}

			command := exec.Command("kill", string(pid))
			err = command.Start()
			if err != nil {
				return err
			}

			err = os.Remove(lockFile)
			if err != nil {
				return fmt.Errorf("can't remove %s.lock: %s", appName, err)
			}

			fmt.Printf("server %s stopped\n", appName)
			return nil
		},
	}
}
