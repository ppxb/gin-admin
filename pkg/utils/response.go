package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ResJson(c *gin.Context, status int, v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	c.Set(ResBodyKey, buf)
	c.Data(status, "application/json; charset=utf-8", buf)
	c.Abort()
}

func ResOk(c *gin.Context) {
	ResJson(c, http.StatusOK, ResponseResult{
		Success: true,
	})
}
