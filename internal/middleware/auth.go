package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"fool_blog_go/internal/model"
)

const (
	SessionUserID = "user_id"
)

// LoadCurrentUser 将当前登录用户挂到 context.user 与模板上下文。
func LoadCurrentUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := sessions.Default(c)
		v := s.Get(SessionUserID)
		if v != nil {
			if uid, ok := v.(uint); ok && uid > 0 {
				var u model.User
				if err := db.First(&u, uid).Error; err == nil && u.Status == model.UserStatusActive {
					c.Set("user", &u)
				}
			}
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) *model.User {
	if v, ok := c.Get("user"); ok {
		if u, ok := v.(*model.User); ok {
			return u
		}
	}
	return nil
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUser(c) == nil {
			if isHTMLRequest(c) {
				c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.RequestURI())
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
				return
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u == nil {
			c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.RequestURI())
			c.Abort()
			return
		}
		if !u.IsAdmin() {
			c.String(http.StatusForbidden, "403 - 需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GuestOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentUser(c) != nil {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}

func isHTMLRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	if accept == "" {
		return true
	}
	return strings.Contains(accept, "text/html") || strings.Contains(accept, "application/xhtml")
}
