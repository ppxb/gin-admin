package prom

import (
	"github.com/gin-gonic/gin"

	"gin-admin/internal/config"
	"gin-admin/pkg/promx"
	"gin-admin/pkg/utils"
)

var (
	Ins           *promx.PrometheusWrapper
	GinMiddleware gin.HandlerFunc
)

func Init() {
	logMethod := make(map[string]struct{})
	logApi := make(map[string]struct{})
	for _, m := range config.C.Util.Prometheus.LogMethods {
		logMethod[m] = struct{}{}
	}
	for _, a := range config.C.Util.Prometheus.LogApis {
		logApi[a] = struct{}{}
	}
	Ins = promx.NewPrometheusWrapper(&promx.Config{
		Enable:         config.C.Util.Prometheus.Enable,
		App:            config.C.Global.AppName,
		ListenPort:     config.C.Util.Prometheus.Port,
		BasicUsername:  config.C.Util.Prometheus.BasicUsername,
		BasicPassword:  config.C.Util.Prometheus.BasicPassword,
		LogApi:         logApi,
		LogMethod:      logMethod,
		Objectives:     map[float64]float64{0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
		DefaultCollect: config.C.Util.Prometheus.DefaultCollect,
	})
	GinMiddleware = promx.NewGinAdapter(Ins).Middleware(config.C.Util.Prometheus.Enable, utils.ReqBodyKey)
}
