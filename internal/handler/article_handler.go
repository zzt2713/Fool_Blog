package handler

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"fool_blog_go/internal/ai"
	"fool_blog_go/internal/middleware"
	"fool_blog_go/internal/model"
	"fool_blog_go/internal/service"
)

// AdminArticles 后台文章列表
func (h *Handler) AdminArticles() gin.HandlerFunc {
	return func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		statusStr := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page < 1 {
			page = 1
		}
		const pageSize = 20

		countDB := h.DB.Model(&model.Article{})
		listDB := h.DB.Preload("Author").Preload("Tags").Order("is_top DESC, created_at DESC")
		if q != "" {
			like := "%" + q + "%"
			countDB = countDB.Where("title LIKE ? OR summary LIKE ?", like, like)
			listDB = listDB.Where("title LIKE ? OR summary LIKE ?", like, like)
		}
		if statusStr != "" {
			countDB = countDB.Where("status = ?", statusStr)
			listDB = listDB.Where("status = ?", statusStr)
		}

		var total int64
		countDB.Count(&total)

		var articles []model.Article
		listDB.Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles)

		c.HTML(http.StatusOK, "admin_articles", h.adminView(c, gin.H{
			"Title":    "文章管理",
			"Articles": articles,
			"Q":        q,
			"Status":   statusStr,
			"Page":     page,
			"PageSize": pageSize,
			"Total":    total,
			"HasNext":  int64(page*pageSize) < total,
			"HasPrev":  page > 1,
		}))
	}
}

// AdminArticleNew 新增文章页
func (h *Handler) AdminArticleNew() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tags []model.Tag
		h.DB.Order("name").Find(&tags)
		c.HTML(http.StatusOK, "admin_article_form", h.adminView(c, gin.H{
			"Title":       "新增文章",
			"Article":     &model.Article{},
			"Tags":        tags,
			"IsNew":       true,
			"SelectedTag": map[uint]bool{},
		}))
	}
}

// AdminArticleCreate 创建文章
func (h *Handler) AdminArticleCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		article, tagIDs, err := h.bindArticleForm(c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		u := middleware.CurrentUser(c)
		if u != nil {
			article.AuthorID = u.ID
		}
		article.CreatedAt = time.Now()
		article.UpdatedAt = time.Now()
		if article.Status == model.ArticleStatusPublished && article.PublishedAt == nil {
			now := time.Now()
			article.PublishedAt = &now
		}
		if err := h.DB.Create(article).Error; err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		h.attachTags(article, tagIDs)
		if article.Status == model.ArticleStatusPublished {
			go h.generateAIReview(article)
		}
		h.logOp(c, "article", "create", "article", strconv.Itoa(int(article.ID)), "新增文章: "+article.Title)
		c.Redirect(http.StatusFound, "/admin/articles")
	}
}

// AdminArticleEdit 编辑页
func (h *Handler) AdminArticleEdit() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var article model.Article
		if err := h.DB.Preload("Tags").First(&article, id).Error; err != nil {
			c.String(http.StatusNotFound, "文章不存在")
			return
		}
		var tags []model.Tag
		h.DB.Order("name").Find(&tags)

		selected := map[uint]bool{}
		for _, t := range article.Tags {
			selected[t.ID] = true
		}

		c.HTML(http.StatusOK, "admin_article_form", h.adminView(c, gin.H{
			"Title":       "编辑文章",
			"Article":     &article,
			"Tags":        tags,
			"IsNew":       false,
			"SelectedTag": selected,
		}))
	}
}

// AdminArticleUpdate 更新
func (h *Handler) AdminArticleUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var article model.Article
		if err := h.DB.First(&article, id).Error; err != nil {
			c.String(http.StatusNotFound, "文章不存在")
			return
		}
		parsed, tagIDs, err := h.bindArticleForm(c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		article.Title = parsed.Title
		article.Slug = parsed.Slug
		article.Summary = parsed.Summary
		article.ContentMD = parsed.ContentMD
		article.ContentHTML = parsed.ContentHTML
		article.Cover = parsed.Cover
		article.Status = parsed.Status
		article.IsTop = parsed.IsTop
		article.UpdatedAt = time.Now()
		if article.Status == model.ArticleStatusPublished && article.PublishedAt == nil {
			now := time.Now()
			article.PublishedAt = &now
		}
		h.DB.Save(&article)
		h.attachTags(&article, tagIDs)
		if article.Status == model.ArticleStatusPublished && article.AIReview == "" {
			go h.generateAIReview(&article)
		}
		h.logOp(c, "article", "update", "article", id, "更新文章: "+article.Title)
		c.Redirect(http.StatusFound, "/admin/articles")
	}
}

// AdminArticleDelete 删除
func (h *Handler) AdminArticleDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var a model.Article
		if err := h.DB.First(&a, id).Error; err == nil {
			h.DB.Exec("DELETE FROM article_tags WHERE article_id = ?", a.ID)
			h.DB.Delete(&a)
		}
		h.logOp(c, "article", "delete", "article", id, "删除文章")
		c.Redirect(http.StatusFound, "/admin/articles")
	}
}

// AdminArticleImport Markdown 导入
func (h *Handler) AdminArticleImport() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/articles/new")
			return
		}
		if !strings.EqualFold(filepath.Ext(file.Filename), ".md") {
			c.Redirect(http.StatusFound, "/admin/articles/new")
			return
		}
		tmpDir := "data/tmp"
		_ = ensureDir(tmpDir)
		dst := filepath.Join(tmpDir, file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.Redirect(http.StatusFound, "/admin/articles/new")
			return
		}
		res, err := service.ImportMarkdown(dst, file.Filename, h.Cfg.Upload.ArticleDir)
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/articles/new")
			return
		}
		var tags []model.Tag
		h.DB.Order("name").Find(&tags)
		summary := service.MakeSummary(res.Content, 200)
		c.HTML(http.StatusOK, "admin_article_form", h.adminView(c, gin.H{
			"Title": "导入文章",
			"Article": &model.Article{
				Title:     res.Title,
				Summary:   summary,
				ContentMD: res.Content,
			},
			"Tags":        tags,
			"IsNew":       true,
			"SelectedTag": map[uint]bool{},
		}))
	}
}

// AdminArticleExportZip 批量导出文章为 zip
func (h *Handler) AdminArticleExportZip() gin.HandlerFunc {
	return func(c *gin.Context) {
		ids := c.PostFormArray("ids")
		if len(ids) == 0 {
			c.String(http.StatusBadRequest, "请选择至少一篇文章")
			return
		}
		var articles []model.Article
		h.DB.Where("id IN ?", ids).Find(&articles)
		if len(articles) == 0 {
			c.String(http.StatusNotFound, "未找到文章")
			return
		}
		items := make([]struct{ Title, Content string }, 0, len(articles))
		titles := make([]string, 0, len(articles))
		for _, a := range articles {
			items = append(items, struct{ Title, Content string }{a.Title, a.ContentMD})
			titles = append(titles, a.Title)
		}
		c.Header("Content-Type", "application/zip")
		c.Header("Content-Disposition", `attachment; filename="articles-`+time.Now().Format("20060102-150405")+`.zip"`)
		_ = service.WriteExportZip(c.Writer, items)
		h.logOp(c, "article", "export_zip", "article", "", "批量导出: "+strings.Join(titles, ", "))
	}
}

// AdminArticleExport 单篇导出
func (h *Handler) AdminArticleExport() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var article model.Article
		if err := h.DB.First(&article, id).Error; err != nil {
			c.String(http.StatusNotFound, "文章不存在")
			return
		}
		_, name, err := service.ExportSingle(article.Title, article.ContentMD, "exports/markdown")
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		h.logOp(c, "article", "export", "article", id, "导出文章: "+article.Title)
		c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		c.String(http.StatusOK, article.ContentMD)
	}
}

// 解析表单
func (h *Handler) bindArticleForm(c *gin.Context) (*model.Article, []uint, error) {
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		return nil, nil, gin.Error{Err: errMissing("title"), Type: gin.ErrorTypePublic}
	}
	contentMD := c.PostForm("content_md")
	summary := strings.TrimSpace(c.PostForm("summary"))
	if summary == "" {
		summary = service.MakeSummary(contentMD, 200)
	}
	htmlStr, _ := service.RenderMarkdown(contentMD)

	status, _ := strconv.Atoi(c.PostForm("status"))
	isTop := 0
	if c.PostForm("is_top") == "on" || c.PostForm("is_top") == "1" {
		isTop = 1
	}

	a := &model.Article{
		Title:       title,
		Slug:        strings.TrimSpace(c.PostForm("slug")),
		Summary:     summary,
		ContentMD:   contentMD,
		ContentHTML: htmlStr,
		Cover:       strings.TrimSpace(c.PostForm("cover")),
		Status:      status,
		IsTop:       isTop,
	}

	var tagIDs []uint
	for _, v := range c.PostFormArray("tag_ids") {
		id, err := strconv.Atoi(v)
		if err == nil && id > 0 {
			tagIDs = append(tagIDs, uint(id))
		}
	}
	return a, tagIDs, nil
}

func (h *Handler) attachTags(article *model.Article, tagIDs []uint) {
	h.DB.Exec("DELETE FROM article_tags WHERE article_id = ?", article.ID)
	for _, id := range tagIDs {
		h.DB.Exec("INSERT INTO article_tags(article_id, tag_id) VALUES (?, ?)", article.ID, id)
	}
}

type stringErr string

func (s stringErr) Error() string { return string(s) }
func errMissing(field string) error {
	return stringErr("缺少必填字段: " + field)
}

func ensureDir(p string) error {
	return mkdirAll(p)
}

func (h *Handler) generateAIReview(article *model.Article) {
	review, err := ai.Review(article.Title, article.ContentMD)
	if err != nil {
		log.Printf("[AI] 生成点评失败 article=%d: %v", article.ID, err)
		return
	}
	h.DB.Model(article).UpdateColumn("ai_review", review)
	log.Printf("[AI] 文章点评已保存 article=%d", article.ID)
}
