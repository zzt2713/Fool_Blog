package model

import "time"

// VisitorStats 全站访问统计（单行记录）
type VisitorStats struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Date       string    `gorm:"size:10;uniqueIndex" json:"date"`
	TodayCount int64     `json:"today_count"`
	TotalCount int64     `json:"total_count"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (VisitorStats) TableName() string { return "visitor_stats" }

// VisitorLog 每日 IP 去重日志
type VisitorLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Date      string    `gorm:"size:10;index:idx_ip_date,unique" json:"date"`
	IP        string    `gorm:"size:45;index:idx_ip_date,unique" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

func (VisitorLog) TableName() string { return "visitor_logs" }
