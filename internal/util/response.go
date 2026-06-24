package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func JSONOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": data})
}

func JSONErr(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": code, "msg": msg})
}

func RedirectFlash(c *gin.Context, url, msg string) {
	c.Redirect(http.StatusFound, url)
	_ = msg
}
