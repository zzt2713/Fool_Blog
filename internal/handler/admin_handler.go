package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"fool_blog_go/internal/middleware"
	"fool_blog_go/internal/model"
	"fool_blog_go/internal/service"
)

// Dashboard 后台首页
func (h *Handler) Dashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, _ := service.GetDashboard(h.DB)
		c.HTML(http.StatusOK, "admin_dashboard", h.adminView(c, gin.H{
			"Title": "仪表盘",
			"Stats": stats,
		}))
	}
}

// Tags 标签管理
func (h *Handler) AdminTags() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tags []model.Tag
		h.DB.Order("name").Find(&tags)
		c.HTML(http.StatusOK, "admin_tags", h.adminView(c, gin.H{
			"Title": "标签管理",
			"Tags":  tags,
		}))
	}
}

func (h *Handler) AdminTagCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := strings.TrimSpace(c.PostForm("name"))
		slug := strings.TrimSpace(c.PostForm("slug"))
		if name == "" {
			c.Redirect(http.StatusFound, "/admin/tags")
			return
		}
		tag := &model.Tag{Name: name, Slug: slug, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		h.DB.Create(tag)
		h.logOp(c, "tag", "create", "tag", strconv.Itoa(int(tag.ID)), "新增标签: "+name)
		c.Redirect(http.StatusFound, "/admin/tags")
	}
}

func (h *Handler) AdminTagUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		name := strings.TrimSpace(c.PostForm("name"))
		slug := strings.TrimSpace(c.PostForm("slug"))
		h.DB.Model(&model.Tag{}).Where("id = ?", id).Updates(map[string]any{
			"name": name, "slug": slug, "updated_at": time.Now(),
		})
		h.logOp(c, "tag", "update", "tag", id, "编辑标签: "+name)
		c.Redirect(http.StatusFound, "/admin/tags")
	}
}

func (h *Handler) AdminTagDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var count int64
		h.DB.Table("article_tags").Where("tag_id = ?", id).Count(&count)
		if count > 0 {
			c.String(http.StatusBadRequest, "该标签下仍有文章，无法删除")
			return
		}
		h.DB.Delete(&model.Tag{}, id)
		h.logOp(c, "tag", "delete", "tag", id, "删除标签")
		c.Redirect(http.StatusFound, "/admin/tags")
	}
}

// Comments 评论管理
func (h *Handler) AdminComments() gin.HandlerFunc {
	return func(c *gin.Context) {
		articleID := c.Query("article_id")
		userID := c.Query("user_id")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}
		const pageSize = 20

		countDB := h.DB.Model(&model.Comment{})
		if articleID != "" {
			countDB = countDB.Where("article_id = ?", articleID)
		}
		if userID != "" {
			countDB = countDB.Where("user_id = ?", userID)
		}
		var total int64
		countDB.Count(&total)

		listDB := h.DB.Preload("User").Preload("Article").Order("created_at DESC")
		if articleID != "" {
			listDB = listDB.Where("article_id = ?", articleID)
		}
		if userID != "" {
			listDB = listDB.Where("user_id = ?", userID)
		}
		var comments []model.Comment
		listDB.Offset((page - 1) * pageSize).Limit(pageSize).Find(&comments)

		c.HTML(http.StatusOK, "admin_comments", h.adminView(c, gin.H{
			"Title":    "评论管理",
			"Comments": comments,
			"Page":     page,
			"PageSize": pageSize,
			"Total":    total,
			"HasNext":  int64(page*pageSize) < total,
			"HasPrev":  page > 1,
		}))
	}
}

func (h *Handler) AdminCommentDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var cm model.Comment
		if err := h.DB.First(&cm, id).Error; err == nil {
			h.DB.Delete(&cm)
			h.DB.Model(&model.Article{}).Where("id = ?", cm.ArticleID).
				UpdateColumn("comment_count", gorm.Expr("MAX(comment_count - 1, 0)"))
		}
		h.logOp(c, "comment", "delete", "comment", id, "删除评论")
		c.Redirect(http.StatusFound, "/admin/comments")
	}
}

// Users 用户管理
func (h *Handler) AdminUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}
		const pageSize = 20

		countDB := h.DB.Model(&model.User{})
		listDB := h.DB.Order("id ASC")
		if q != "" {
			like := "%" + q + "%"
			countDB = countDB.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
			listDB = listDB.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
		}
		var total int64
		countDB.Count(&total)

		var users []model.User
		listDB.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

		c.HTML(http.StatusOK, "admin_users", h.adminView(c, gin.H{
			"Title":    "用户管理",
			"Users":    users,
			"Q":        q,
			"Page":     page,
			"PageSize": pageSize,
			"Total":    total,
			"HasNext":  int64(page*pageSize) < total,
			"HasPrev":  page > 1,
		}))
	}
}

func (h *Handler) AdminUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		role, _ := strconv.Atoi(c.PostForm("role"))
		h.DB.Model(&model.User{}).Where("id = ?", id).Update("role", role)
		h.logOp(c, "user", "role", "user", id, "修改角色为 "+strconv.Itoa(role))
		c.Redirect(http.StatusFound, "/admin/users")
	}
}

func (h *Handler) AdminUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		status, _ := strconv.Atoi(c.PostForm("status"))
		h.DB.Model(&model.User{}).Where("id = ?", id).Update("status", status)
		h.logOp(c, "user", "status", "user", id, "修改状态为 "+strconv.Itoa(status))
		c.Redirect(http.StatusFound, "/admin/users")
	}
}

// AdminUserDelete 删除用户（同时会删该用户的评论）
func (h *Handler) AdminUserDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		me := middleware.CurrentUser(c)
		if me != nil && strconv.Itoa(int(me.ID)) == id {
			c.String(http.StatusBadRequest, "不能删除当前登录账号")
			return
		}
		h.DB.Where("user_id = ?", id).Delete(&model.Comment{})
		h.DB.Delete(&model.User{}, id)
		h.logOp(c, "user", "delete", "user", id, "删除用户")
		c.Redirect(http.StatusFound, "/admin/users")
	}
}

// Site 站点设置
func (h *Handler) AdminSite() gin.HandlerFunc {
	return func(c *gin.Context) {
		site := h.loadSite()
		var articles []model.Article
		h.DB.Where("status = ?", model.ArticleStatusPublished).Order("title").Select("id", "title").Find(&articles)
		c.HTML(http.StatusOK, "admin_site", h.adminView(c, gin.H{
			"Title":    "站点设置",
			"Setting":  site,
			"Articles": articles,
			"Saved":    c.Query("ok") == "1",
		}))
	}
}

func (h *Handler) AdminSiteUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var site model.SiteSetting
		h.DB.First(&site)
		site.SiteName = c.PostForm("site_name")
		site.SiteSubtitle = c.PostForm("site_subtitle")
		site.WelcomeText = c.PostForm("welcome_text")
		site.FooterText = c.PostForm("footer_text")
		site.Announcement = c.PostForm("announcement")
		site.AuthorIntro = c.PostForm("author_intro")
		site.WallpaperAPI = c.PostForm("wallpaper_api")
		site.CustomWallpaper = c.PostForm("custom_wallpaper")
		site.EnableWallpaper = c.PostForm("enable_wallpaper") == "on"
		site.DefaultCover = c.PostForm("default_cover")
		site.ICPInfo = c.PostForm("icp_info")
		if articleID := c.PostForm("announcement_article_id"); articleID != "" {
			if id, err := strconv.Atoi(articleID); err == nil && id > 0 {
				uid := uint(id)
				site.AnnouncementArticleID = &uid
			} else {
				site.AnnouncementArticleID = nil
			}
		} else {
			site.AnnouncementArticleID = nil
		}
		if value := strings.TrimSpace(c.PostForm("site_created_at")); value != "" {
			if t, err := time.ParseInLocation("2006-01-02T15:04", value, time.Local); err == nil {
				site.SiteCreatedAt = &t
			}
		}
		site.UpdatedAt = time.Now()
		h.DB.Save(&site)
		h.logOp(c, "site", "update", "site", "1", "更新站点设置")
		c.Redirect(http.StatusFound, "/admin/site?ok=1")
	}
}

// Logs 操作日志
func (h *Handler) AdminLogs() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}
		const pageSize = 20

		var total int64
		h.DB.Model(&model.OperationLog{}).Count(&total)

		var logs []model.OperationLog
		h.DB.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

		c.HTML(http.StatusOK, "admin_logs", h.adminView(c, gin.H{
			"Title":    "操作日志",
			"Logs":     logs,
			"Page":     page,
			"PageSize": pageSize,
			"Total":    total,
			"HasNext":  int64(page*pageSize) < total,
			"HasPrev":  page > 1,
		}))
	}
}

func (h *Handler) adminView(c *gin.Context, data gin.H) gin.H {
	if data == nil {
		data = gin.H{}
	}
	data["Site"] = h.loadSite()
	data["CurrentUser"] = middleware.CurrentUser(c)
	return data
}

func (h *Handler) logOp(c *gin.Context, module, action, targetType, targetID, summary string) {
	u := middleware.CurrentUser(c)
	log := &model.OperationLog{
		Module:     module,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Summary:    summary,
		IP:         c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		CreatedAt:  time.Now(),
	}
	if u != nil {
		log.OperatorID = u.ID
		log.Operator = u.Username
	}
	service.WriteLog(h.DB, log)
}
