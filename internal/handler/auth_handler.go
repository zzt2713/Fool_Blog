package handler

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"fool_blog_go/internal/email"
	"fool_blog_go/internal/middleware"
	"fool_blog_go/internal/service"
)

func (h *Handler) LoginPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		next := c.DefaultQuery("next", "/")
		c.HTML(http.StatusOK, "login", h.viewData(c, gin.H{
			"Title": "登录",
			"Next":  next,
		}))
	}
}

func (h *Handler) LoginSubmit() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := strings.TrimSpace(c.PostForm("username"))
		password := c.PostForm("password")
		next := c.DefaultPostForm("next", "/")
		u, err := service.Authenticate(h.DB, username, password)
		if err != nil {
			c.HTML(http.StatusOK, "login", h.viewData(c, gin.H{
				"Title":    "登录",
				"Error":    err.Error(),
				"Username": username,
				"Next":     next,
			}))
			return
		}
		s := sessions.Default(c)
		s.Set(middleware.SessionUserID, u.ID)
		_ = s.Save()
		c.Redirect(http.StatusFound, next)
	}
}

func (h *Handler) RegisterPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "register", h.viewData(c, gin.H{"Title": "注册"}))
	}
}

func (h *Handler) RegisterSubmit() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := strings.TrimSpace(c.PostForm("username"))
		password := c.PostForm("password")
		nickname := strings.TrimSpace(c.PostForm("nickname"))
		emailAddr := strings.TrimSpace(c.PostForm("email"))
		code := strings.TrimSpace(c.PostForm("code"))

		if emailAddr == "" {
			c.HTML(http.StatusOK, "register", h.viewData(c, gin.H{
				"Title":    "注册",
				"Error":    "邮箱不能为空",
				"Username": username,
				"Nickname": nickname,
				"Email":    emailAddr,
			}))
			return
		}
		if !email.VerifyCode(emailAddr, code) {
			c.HTML(http.StatusOK, "register", h.viewData(c, gin.H{
				"Title":    "注册",
				"Error":    "验证码错误或已过期",
				"Username": username,
				"Nickname": nickname,
				"Email":    emailAddr,
			}))
			return
		}

		u, err := service.Register(h.DB, username, password, nickname, emailAddr)
		if err != nil {
			c.HTML(http.StatusOK, "register", h.viewData(c, gin.H{
				"Title":    "注册",
				"Error":    err.Error(),
				"Username": username,
				"Nickname": nickname,
				"Email":    emailAddr,
			}))
			return
		}
		s := sessions.Default(c)
		s.Set(middleware.SessionUserID, u.ID)
		_ = s.Save()
		c.Redirect(http.StatusFound, "/")
	}
}

func (h *Handler) SendVerifyCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		emailAddr := strings.TrimSpace(c.PostForm("email"))
		if emailAddr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱不能为空"})
			return
		}
		_, err := email.SendVerifyCode(emailAddr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
	}
}

func (h *Handler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		s.Clear()
		_ = s.Save()
		c.Redirect(http.StatusFound, "/")
	}
}
