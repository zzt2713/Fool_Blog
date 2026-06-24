package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"fool_blog_go/internal/ai"
	"fool_blog_go/internal/model"
	"fool_blog_go/internal/service"
)

// AIReview generates (or returns cached) AI review for an article.
func (h *Handler) AIReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var article model.Article
		if err := h.DB.First(&article, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "文章不存在"})
			return
		}

		// 已有点评，渲染 markdown 返回
		if article.AIReview != "" {
			rendered, _ := service.RenderMarkdown(article.AIReview)
			c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"review": article.AIReview, "rendered": rendered}})
			return
		}

		// 异步生成
		review, err := ai.Review(article.Title, article.ContentMD)
		if err != nil {
			log.Printf("[AI] 生成点评失败 article=%d: %v", article.ID, err)
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "AI 点评生成失败"})
			return
		}
		h.DB.Model(&article).UpdateColumn("ai_review", review)
		log.Printf("[AI] 文章点评已保存 article=%d", article.ID)
		rendered, _ := service.RenderMarkdown(review)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"review": review, "rendered": rendered}})
	}
}
