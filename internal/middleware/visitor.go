package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"fool_blog_go/internal/model"
)

// VisitorStatsData 传递给模板的访问统计
type VisitorStatsData struct {
	TodayCount int64
	TotalCount int64
}

// TrackVisitor 中间件：按 IP+日期 去重统计访问量
func TrackVisitor(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		today := time.Now().Format("2006-01-02")

		// 检查今日该 IP 是否已记录
		var existing model.VisitorLog
		if db.Where("date = ? AND ip = ?", today, ip).First(&existing).Error != nil {
			// 新 IP → 记录日志
			db.Create(&model.VisitorLog{Date: today, IP: ip, CreatedAt: time.Now()})

			// 更新统计：尝试找到今日记录
			var stats model.VisitorStats
			if db.Where("date = ?", today).First(&stats).Error != nil {
				// 今日首条记录
				db.Create(&model.VisitorStats{
					Date:       today,
					TodayCount: 1,
					TotalCount: currentTotalCount(db) + 1,
					UpdatedAt:  time.Now(),
				})
			} else {
				// 今日已有记录，累加
				db.Model(&stats).Updates(map[string]interface{}{
					"today_count": stats.TodayCount + 1,
					"total_count": currentTotalCount(db) + 1,
					"updated_at":  time.Now(),
				})
			}
		}

		// 读取统计放入 context
		var stats model.VisitorStats
		if err := db.Where("date = ?", today).First(&stats).Error; err != nil {
			stats = model.VisitorStats{Date: today, TodayCount: 0, TotalCount: 0}
		}
		c.Set("visitorStats", VisitorStatsData{
			TodayCount: stats.TodayCount,
			TotalCount: currentTotalCount(db),
		})

		c.Next()
	}
}

func currentTotalCount(db *gorm.DB) int64 {
	var maxTotal struct{ V int64 }
	db.Raw("SELECT COALESCE(MAX(total_count), 0) AS v FROM visitor_stats").Scan(&maxTotal)
	return maxTotal.V
}

// GetVisitorStats 从 context 取访问统计
func GetVisitorStats(c *gin.Context) VisitorStatsData {
	if v, ok := c.Get("visitorStats"); ok {
		return v.(VisitorStatsData)
	}
	return VisitorStatsData{}
}
