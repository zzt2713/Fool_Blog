package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"fool_blog_go/internal/util"
)

// UploadArticleImage 富文本/编辑器图片上传
func (h *Handler) UploadArticleImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "缺少 file"})
			return
		}
		max := int64(h.Cfg.Upload.MaxSizeMB) * 1024 * 1024
		url, err := util.SaveUploadedImage(file, h.Cfg.Upload.ArticleDir, max)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": gin.H{"url": url}})
	}
}

// UploadCover 上传封面
func (h *Handler) UploadCover() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "缺少 file"})
			return
		}
		max := int64(h.Cfg.Upload.MaxSizeMB) * 1024 * 1024
		url, err := util.SaveUploadedImage(file, h.Cfg.Upload.CoverDir, max)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": gin.H{"url": url}})
	}
}

func mkdirAll(p string) error {
	return os.MkdirAll(p, 0o755)
}
