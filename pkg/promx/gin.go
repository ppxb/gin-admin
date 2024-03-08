package promx

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type GinAdapter struct {
	prom *PrometheusWrapper
}

func NewGinAdapter(p *PrometheusWrapper) *GinAdapter {
	return &GinAdapter{prom: p}
}

func (a *GinAdapter) Middleware(enable bool, reqKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !enable {
			ctx.Next()
			return
		}

		start := time.Now()
		rcvdBytes := 0
		if v, ok := ctx.Get(reqKey); ok {
			if b, ok := v.([]byte); ok {
				rcvdBytes = len(b)
			}
		}
		ctx.Next()
		latency := float64(time.Since(start).Milliseconds())

		path := ctx.Request.URL.Path
		for _, param := range ctx.Params {
			path = strings.Replace(path, param.Value, ":"+param.Key, -1)
		}
		a.prom.Log(path, ctx.Request.Method, fmt.Sprintf("%d", ctx.Writer.Status()), float64(ctx.Writer.Size()), float64(rcvdBytes), latency)
	}
}
