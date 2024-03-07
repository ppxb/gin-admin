package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gin-admin/internal/config"
	"gin-admin/pkg/utils"
	"github.com/gin-gonic/gin"
)

type C int

func httpServerListener(ctx context.Context) (func(), error) {
	handler := gin.New()
	handler.GET("/health", func(c *gin.Context) {
		utils.ResOk(c)
	})

	addr := config.C.Global.Http.Addr
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Second * time.Duration(60),
		WriteTimeout: time.Second * time.Duration(60),
		IdleTimeout:  time.Second * time.Duration(60),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Failed to listen http server:\n %s", err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(60))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Printf("Failed to listen http server:\n %s", err)
		}
	}, nil
}
