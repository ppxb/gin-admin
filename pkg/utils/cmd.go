package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(ctx context.Context, handler func(ctx context.Context) (func(), error)) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	clean, err := handler(ctx)
	if err != nil {
		return err
	}

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clean()
	return nil
}
