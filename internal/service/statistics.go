package service

import (
	"time"

	"gorm.io/gorm"

	"fool_blog_go/internal/model"
)

type DashboardStats struct {
	ArticleCount   int64
	UserCount      int64
	CommentCount   int64
	TagCount       int64
	TotalViews     int64
	TotalLikes     int64
	TodayArticles  int64
	TodayComments  int64
	TodayViews     int64
	OnlineUsers    int64
	RecentArticles []model.Article
	RecentComments []model.Comment
	SiteCreated    string
}

func GetDashboard(db *gorm.DB) (*DashboardStats, error) {
	s := &DashboardStats{}
	db.Model(&model.Article{}).Count(&s.ArticleCount)
	db.Model(&model.User{}).Count(&s.UserCount)
	db.Model(&model.Comment{}).Count(&s.CommentCount)
	db.Model(&model.Tag{}).Count(&s.TagCount)

	var sumViews struct{ V int64 }
	db.Raw("SELECT COALESCE(SUM(view_count),0) AS v FROM articles").Scan(&sumViews)
	s.TotalViews = sumViews.V

	var sumLikes struct{ V int64 }
	db.Raw("SELECT COALESCE(SUM(like_count),0) AS v FROM articles").Scan(&sumLikes)
	s.TotalLikes = sumLikes.V

	// 在线用户（7天内登录过）
	weekAgo := time.Now().AddDate(0, 0, -7)
	db.Model(&model.User{}).Where("last_login_at > ?", weekAgo).Count(&s.OnlineUsers)

	today := time.Now().Format("2006-01-02")
	db.Model(&model.Article{}).Where("DATE(created_at) = ?", today).Count(&s.TodayArticles)
	db.Model(&model.Comment{}).Where("DATE(created_at) = ?", today).Count(&s.TodayComments)

	var site model.SiteSetting
	if err := db.First(&site).Error; err == nil {
		if site.SiteCreatedAt != nil {
			s.SiteCreated = site.SiteCreatedAt.Format("2006-01-02 15:04")
		} else if !site.CreatedAt.IsZero() {
			s.SiteCreated = site.CreatedAt.Format("2006-01-02 15:04")
		}
	}
	if s.SiteCreated == "" {
		s.SiteCreated = "-"
	}

	db.Order("created_at DESC").Limit(5).Find(&s.RecentArticles)
	db.Preload("User").Preload("Article").Order("created_at DESC").Limit(5).Find(&s.RecentComments)
	return s, nil
}

// WriteLog 写一条操作日志，失败不阻塞业务。
func WriteLog(db *gorm.DB, log *model.OperationLog) {
	_ = db.Create(log).Error
}
