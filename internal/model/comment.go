package model

import "time"

const (
	CommentStatusNormal = 0
	CommentStatusHidden = 1
)

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ArticleID uint      `gorm:"index" json:"article_id"`
	Article   *Article  `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	UserID    uint      `gorm:"index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Status    int       `gorm:"default:0;index" json:"status"`
	IP        string    `gorm:"size:64" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Comment) TableName() string { return "comments" }
