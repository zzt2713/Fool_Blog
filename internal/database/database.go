package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"fool_blog_go/internal/config"
	"fool_blog_go/internal/model"
)

var DB *gorm.DB

func Init(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.Database.Driver {
	case "mysql":
		dsn := cfg.Database.DSN
		if dsn == "" && cfg.Database.Host != "" {
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
		}
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger:                          logger.Default.LogMode(logger.Warn),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	default:
		if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0o755); err != nil {
			return nil, err
		}
		db, err = gorm.Open(sqlite.Open(cfg.Database.Path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
	}
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Article{},
		&model.Tag{},
		&model.Comment{},
		&model.SiteSetting{},
		&model.OperationLog{},
		&model.ArticleLike{},
		&model.VisitorStats{},
		&model.VisitorLog{},
	); err != nil {
		return nil, err
	}

	DB = db
	if err := ensureDefaults(db, cfg); err != nil {
		return nil, err
	}
	return db, nil
}

func ensureDefaults(db *gorm.DB, cfg *config.Config) error {
	var count int64
	if err := db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.DefaultPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin := &model.User{
			Username:     cfg.Admin.DefaultUsername,
			PasswordHash: string(hash),
			Nickname:     "管理员",
			Role:         model.RoleAdmin,
			Status:       model.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := db.Create(admin).Error; err != nil {
			return err
		}
		log.Printf("[init] 创建默认管理员: %s / %s", cfg.Admin.DefaultUsername, cfg.Admin.DefaultPassword)
	}

	var settingCount int64
	if err := db.Model(&model.SiteSetting{}).Count(&settingCount).Error; err != nil {
		return err
	}
	if settingCount == 0 {
		now := time.Now()
		s := &model.SiteSetting{
			SiteName:        cfg.App.Name,
			SiteSubtitle:    "记录、思考、分享",
			Announcement:    "欢迎来到 Fool Blog。",
			AuthorIntro:     "一个热爱编码的开发者。",
			WallpaperAPI:    cfg.Wallpaper.API,
			CustomWallpaper: cfg.Wallpaper.CustomURL,
			EnableWallpaper: cfg.Wallpaper.Enabled,
			SiteCreatedAt:   &now,
		}
		if err := db.Create(s).Error; err != nil {
			return err
		}
	} else {
		// 补偿：老记录没有 site_created_at，用 created_at 兜底
		var s model.SiteSetting
		db.First(&s)
		if s.SiteCreatedAt == nil {
			db.Model(&s).Update("site_created_at", s.CreatedAt)
		}
	}
	return nil
}
