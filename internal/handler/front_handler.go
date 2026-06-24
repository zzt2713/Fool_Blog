package handler

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"fool_blog_go/internal/middleware"
	"fool_blog_go/internal/model"
	"fool_blog_go/internal/service"
)

// Index 首页：置顶 + 最新文章列表
func (h *Handler) Index() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}
		const pageSize = 10

		var topArticles []model.Article
		h.DB.Preload("Author").Preload("Tags").
			Where("status = ? AND is_top = ?", model.ArticleStatusPublished, model.ArticleTop).
			Order("COALESCE(published_at, created_at) DESC").Limit(5).Find(&topArticles)

		var articles []model.Article
		var total int64
		h.DB.Model(&model.Article{}).Where("status = ?", model.ArticleStatusPublished).Count(&total)
		h.DB.Preload("Author").Preload("Tags").
			Where("status = ?", model.ArticleStatusPublished).
			Order("COALESCE(published_at, created_at) DESC").
			Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles)

		var tags []model.Tag
		h.DB.Order("name").Find(&tags)

		site := h.loadSite()
		var announcementArticle *model.Article
		if site.AnnouncementArticleID != nil {
			h.DB.Where("id = ? AND status = ?", *site.AnnouncementArticleID, model.ArticleStatusPublished).First(&announcementArticle)
		}

		data := h.viewData(c, gin.H{
			"Title":       "首页",
			"TopArticles": topArticles,
			"Articles":    articles,
			"Tags":        tags,
			"Page":        page,
			"PageSize":    pageSize,
			"Total":       total,
			"HasNext":     int64(page*pageSize) < total,
			"HasPrev":     page > 1,
		})
		data["AnnouncementArticle"] = announcementArticle
		c.HTML(http.StatusOK, "index", data)
	}
}

// ArticleDetail 文章详情
func (h *Handler) ArticleDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		var article model.Article
		// 先尝试按 slug 查，找不到再按 ID 查
		q := h.DB.Preload("Author").Preload("Tags").Where("status = ?", model.ArticleStatusPublished)
		if err := q.Where("slug = ?", slug).First(&article).Error; err != nil {
			log.Printf("[ArticleDetail] slug=%q not found, trying ID lookup", slug)
			if err := h.DB.Preload("Author").Preload("Tags").Where("status = ? AND id = ?", model.ArticleStatusPublished, slug).First(&article).Error; err != nil {
				log.Printf("[ArticleDetail] id=%q not found: %v", slug, err)
				c.String(http.StatusNotFound, "文章不存在")
				return
			}
		}
		// 浏览量 + 1
		h.DB.Model(&article).UpdateColumn("view_count", article.ViewCount+1)
		article.ViewCount++

		// 从点赞表重新计算，确保一致
		var realLikeCount int64
		h.DB.Model(&model.ArticleLike{}).Where("article_id = ?", article.ID).Count(&realLikeCount)
		article.LikeCount = int(realLikeCount)
		h.DB.Model(&article).UpdateColumn("like_count", article.LikeCount)

		toc := service.ExtractTOC(article.ContentHTML)

		var comments []model.Comment
		h.DB.Preload("User").Where("article_id = ? AND status = ?", article.ID, model.CommentStatusNormal).
			Order("created_at DESC").Find(&comments)

		var liked bool
		if u := middleware.CurrentUser(c); u != nil {
			var count int64
			h.DB.Model(&model.ArticleLike{}).Where("article_id = ? AND user_id = ?", article.ID, u.ID).Count(&count)
			liked = count > 0
		}

		var aiReviewHTML template.HTML
		if article.AIReview != "" {
			if rendered, err := service.RenderMarkdown(article.AIReview); err == nil {
				aiReviewHTML = template.HTML(rendered)
			}
		}

		c.HTML(http.StatusOK, "article", h.viewData(c, gin.H{
			"Title":       article.Title,
			"Article":     article,
			"TOC":         toc,
			"Comments":    comments,
			"Liked":       liked,
			"AIReviewHTML": aiReviewHTML,
		}))
	}
}

// Search 文章搜索（标题/摘要/正文/标签）
func (h *Handler) Search() gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}
		const pageSize = 20

		var articles []model.Article
		var total int64
		if q != "" {
			like := "%" + q + "%"
			subTag := h.DB.Table("article_tags at").
				Select("at.article_id").
				Joins("JOIN tags t ON t.id = at.tag_id").
				Where("t.name LIKE ?", like)

			countQ := h.DB.Model(&model.Article{}).
				Where("status = ?", model.ArticleStatusPublished).
				Where("title LIKE ? OR summary LIKE ? OR content_md LIKE ? OR id IN (?)",
					like, like, like, subTag)
			countQ.Count(&total)

			h.DB.Preload("Author").Preload("Tags").
				Where("status = ?", model.ArticleStatusPublished).
				Where("title LIKE ? OR summary LIKE ? OR content_md LIKE ? OR id IN (?)",
					like, like, like, subTag).
				Order("published_at DESC, created_at DESC").
				Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles)
		}
		c.HTML(http.StatusOK, "search", h.viewData(c, gin.H{
			"Title":    "搜索: " + q,
			"Q":        q,
			"Articles": articles,
			"Page":     page,
			"PageSize": pageSize,
			"Total":    total,
			"HasNext":  int64(page*pageSize) < total,
			"HasPrev":  page > 1,
		}))
	}
}

// TagList 所有标签
func (h *Handler) TagList() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tags []model.Tag
		h.DB.Order("name").Find(&tags)
		c.HTML(http.StatusOK, "tags", h.viewData(c, gin.H{
			"Title": "标签",
			"Tags":  tags,
		}))
	}
}

// TagDetail 标签下的文章
func (h *Handler) TagDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}
		const pageSize = 20

		var tag model.Tag
		if err := h.DB.First(&tag, id).Error; err != nil {
			c.String(http.StatusNotFound, "标签不存在")
			return
		}

		var total int64
		h.DB.Model(&model.Article{}).
			Joins("JOIN article_tags at ON at.article_id = articles.id").
			Where("at.tag_id = ? AND articles.status = ?", tag.ID, model.ArticleStatusPublished).
			Count(&total)

		var articles []model.Article
		h.DB.Preload("Author").Preload("Tags").
			Joins("JOIN article_tags at ON at.article_id = articles.id").
			Where("at.tag_id = ? AND articles.status = ?", tag.ID, model.ArticleStatusPublished).
			Order("articles.published_at DESC").
			Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles)

		c.HTML(http.StatusOK, "tag", h.viewData(c, gin.H{
			"Title":    "标签: " + tag.Name,
			"Tag":      tag,
			"Articles": articles,
			"Page":     page,
			"PageSize": pageSize,
			"Total":    total,
			"HasNext":  int64(page*pageSize) < total,
			"HasPrev":  page > 1,
		}))
	}
}

// Archive 归档
type archiveYear struct {
	Year   int
	Months []archiveMonth
}

type archiveMonth struct {
	Month    int
	Articles []model.Article
}

func (h *Handler) Archive() gin.HandlerFunc {
	return func(c *gin.Context) {
		var articles []model.Article
		h.DB.Where("status = ?", model.ArticleStatusPublished).
			Order("published_at DESC").Find(&articles)

		groups := map[int]map[int][]model.Article{}
		for _, a := range articles {
			t := a.CreatedAt
			if a.PublishedAt != nil {
				t = *a.PublishedAt
			}
			y, m := t.Year(), int(t.Month())
			if _, ok := groups[y]; !ok {
				groups[y] = map[int][]model.Article{}
			}
			groups[y][m] = append(groups[y][m], a)
		}

		var years []archiveYear
		var ys []int
		for y := range groups {
			ys = append(ys, y)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(ys)))
		for _, y := range ys {
			months := groups[y]
			var ms []int
			for m := range months {
				ms = append(ms, m)
			}
			sort.Sort(sort.Reverse(sort.IntSlice(ms)))
			var mList []archiveMonth
			for _, m := range ms {
				mList = append(mList, archiveMonth{Month: m, Articles: months[m]})
			}
			years = append(years, archiveYear{Year: y, Months: mList})
		}

		c.HTML(http.StatusOK, "archive", h.viewData(c, gin.H{
			"Title": "归档",
			"Years": years,
		}))
	}
}

// LikeArticle 点赞文章（每用户一次）
func (h *Handler) LikeArticle() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := middleware.CurrentUser(c)
		if u == nil {
			c.JSON(http.StatusOK, gin.H{"code": 401, "msg": "请先登录"})
			return
		}
		id := c.Param("id")
		var a model.Article
		if err := h.DB.First(&a, id).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 404, "msg": "文章不存在"})
			return
		}
		var count int64
		h.DB.Model(&model.ArticleLike{}).Where("article_id = ? AND user_id = ?", a.ID, u.ID).Count(&count)
		if count > 0 {
			var realCount int64
			h.DB.Model(&model.ArticleLike{}).Where("article_id = ?", a.ID).Count(&realCount)
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "你已经点过赞了", "data": gin.H{"like_count": int(realCount), "liked": true}})
			return
		}
		h.DB.Create(&model.ArticleLike{ArticleID: a.ID, UserID: u.ID})
		var realCount int64
		h.DB.Model(&model.ArticleLike{}).Where("article_id = ?", a.ID).Count(&realCount)
		h.DB.Model(&a).UpdateColumn("like_count", int(realCount))
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": gin.H{"like_count": int(realCount), "liked": true}})
	}
}

// PostComment 发表评论（要求登录）
func (h *Handler) PostComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := middleware.CurrentUser(c)
		if u == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		id := c.Param("id")
		var a model.Article
		if err := h.DB.First(&a, id).Error; err != nil {
			c.String(http.StatusNotFound, "文章不存在")
			return
		}
		content := strings.TrimSpace(c.PostForm("content"))
		if content == "" {
			c.Redirect(http.StatusFound, "/article/"+strconv.Itoa(int(a.ID)))
			return
		}
		cm := &model.Comment{
			ArticleID: a.ID,
			UserID:    u.ID,
			Content:   content,
			IP:        c.ClientIP(),
			Status:    model.CommentStatusNormal,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		h.DB.Create(cm)
		h.DB.Model(&a).UpdateColumn("comment_count", a.CommentCount+1)
		// 用 slug 优先，没有 slug 就用 ID
		target := strconv.Itoa(int(a.ID))
		if a.Slug != "" {
			target = a.Slug
		}
		c.Redirect(http.StatusFound, "/article/"+target+"#comments")
	}
}

// viewData 注入公共模板上下文（站点配置、当前用户、壁纸 URL 等）
func (h *Handler) viewData(c *gin.Context, data gin.H) gin.H {
	site := h.loadSite()
	if data == nil {
		data = gin.H{}
	}
	data["Site"] = site
	data["CurrentUser"] = middleware.CurrentUser(c)
	data["Wallpaper"] = h.resolveWallpaper(site)
	// 建站运行秒数：当前时间 - 建站时间
	if site.SiteCreatedAt != nil {
		sec := int(time.Since(*site.SiteCreatedAt).Seconds())
		if sec < 0 {
			sec = 0
		}
		data["SiteSeconds"] = sec
	} else {
		data["SiteSeconds"] = 0
	}
	vs := middleware.GetVisitorStats(c)
	data["TodayVisitors"] = vs.TodayCount
	data["TotalVisitors"] = vs.TotalCount
	return data
}

func (h *Handler) resolveWallpaper(site model.SiteSetting) string {
	if !site.EnableWallpaper {
		return ""
	}
	if site.CustomWallpaper != "" {
		return site.CustomWallpaper
	}
	if site.WallpaperAPI != "" {
		return site.WallpaperAPI
	}
	if h.Cfg.Wallpaper.CustomURL != "" {
		return h.Cfg.Wallpaper.CustomURL
	}
	return h.Cfg.Wallpaper.API
}
