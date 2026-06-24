package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"fool_blog_go/internal/middleware"
	"fool_blog_go/internal/model"
	"fool_blog_go/internal/util"
)

// ProfilePage 个人中心
func (h *Handler) ProfilePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := middleware.CurrentUser(c)
		c.HTML(http.StatusOK, "profile", h.viewData(c, gin.H{
			"Title": "个人中心",
			"Me":    u,
		}))
	}
}

// ProfileUpdate 更新昵称、邮箱、密码
func (h *Handler) ProfileUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := middleware.CurrentUser(c)
		if u == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		nickname := strings.TrimSpace(c.PostForm("nickname"))
		email := strings.TrimSpace(c.PostForm("email"))
		oldPwd := c.PostForm("old_password")
		newPwd := c.PostForm("new_password")

		var msg, errMsg string
		updates := map[string]any{"updated_at": time.Now()}
		if nickname != "" {
			updates["nickname"] = nickname
		}
		updates["email"] = email
		if newPwd != "" {
			if len(newPwd) < 6 {
				errMsg = "新密码至少 6 位"
			} else if !util.CheckPassword(u.PasswordHash, oldPwd) {
				errMsg = "旧密码错误"
			} else {
				hash, err := util.HashPassword(newPwd)
				if err != nil {
					errMsg = "密码加密失败"
				} else {
					updates["password_hash"] = hash
				}
			}
		}
		if errMsg == "" {
			h.DB.Model(&model.User{}).Where("id = ?", u.ID).Updates(updates)
			msg = "保存成功"
		}
		h.DB.First(u, u.ID)
		c.HTML(http.StatusOK, "profile", h.viewData(c, gin.H{
			"Title":   "个人中心",
			"Me":      u,
			"Message": msg,
			"Error":   errMsg,
		}))
	}
}

// UploadAvatar 头像上传
func (h *Handler) UploadAvatar() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := middleware.CurrentUser(c)
		if u == nil {
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "未登录"})
			return
		}
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "缺少 file"})
			return
		}
		max := int64(h.Cfg.Upload.MaxSizeMB) * 1024 * 1024
		url, err := util.SaveUploadedImage(file, h.Cfg.Upload.AvatarDir, max)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		h.DB.Model(&model.User{}).Where("id = ?", u.ID).Update("avatar", url)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": gin.H{"url": url}})
	}
}

// RandomAvatar 随机头像
func (h *Handler) RandomAvatar() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := middleware.CurrentUser(c)
		if u == nil {
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "未登录"})
			return
		}
		resp, err := http.Get("https://www.loliapi.com/acg/pp/")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "获取随机头像失败"})
			return
		}
		defer resp.Body.Close()

		ext := ".jpg"
		ct := resp.Header.Get("Content-Type")
		switch {
		case strings.Contains(ct, "png"):
			ext = ".png"
		case strings.Contains(ct, "gif"):
			ext = ".gif"
		case strings.Contains(ct, "webp"):
			ext = ".webp"
		}

		dir := h.Cfg.Upload.AvatarDir
		if err := os.MkdirAll(dir, 0o755); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "创建目录失败"})
			return
		}
		name := uuid.NewString() + ext
		dst := filepath.Join(dir, name)

		out, err := os.Create(dst)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "保存文件失败"})
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, resp.Body); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "写入文件失败"})
			return
		}

		url := "/" + filepath.ToSlash(dst)
		h.DB.Model(&model.User{}).Where("id = ?", u.ID).Update("avatar", url)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": gin.H{"url": url}})
	}
}
