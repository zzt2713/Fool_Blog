package model

import "time"

const (
	ArticleStatusDraft     = 0
	ArticleStatusPublished = 1

	ArticleNotTop = 0
	ArticleTop    = 1
)

type Article struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Title        string     `gorm:"size:255;not null;index" json:"title"`
	Slug         string     `gorm:"size:255;index" json:"slug"`
	Summary      string     `gorm:"size:512" json:"summary"`
	ContentMD    string     `gorm:"type:text" json:"content_md"`
	ContentHTML  string     `gorm:"type:text" json:"content_html"`
	Cover        string     `gorm:"size:255" json:"cover"`
	Status       int        `gorm:"default:0;index" json:"status"`
	IsTop        int        `gorm:"default:0;index" json:"is_top"`
	ViewCount    int        `gorm:"default:0" json:"view_count"`
	LikeCount    int        `gorm:"default:0" json:"like_count"`
	CommentCount int        `gorm:"default:0" json:"comment_count"`
	AIReview     string     `gorm:"type:text" json:"ai_review"`
	AuthorID     uint       `gorm:"index" json:"author_id"`
	Author       *User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Tags         []Tag      `gorm:"many2many:article_tags;" json:"tags,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PublishedAt  *time.Time `json:"published_at"`
}

func (Article) TableName() string { return "articles" }

type ArticleLike struct {
	ID        uint `gorm:"primaryKey"`
	ArticleID uint `gorm:"index;uniqueIndex:idx_article_user"`
	UserID    uint `gorm:"index;uniqueIndex:idx_article_user"`
}

func (ArticleLike) TableName() string { return "article_likes" }
