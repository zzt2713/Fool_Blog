package model

import "time"

type SiteSetting struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	SiteName        string     `gorm:"size:128" json:"site_name"`
	SiteSubtitle    string     `gorm:"size:255" json:"site_subtitle"`
	WelcomeText     string     `gorm:"size:255" json:"welcome_text"`
	FooterText      string     `gorm:"size:255" json:"footer_text"`
	Announcement    string     `gorm:"type:text" json:"announcement"`
	AuthorIntro     string     `gorm:"type:text" json:"author_intro"`
	WallpaperAPI    string     `gorm:"size:255" json:"wallpaper_api"`
	CustomWallpaper string     `gorm:"size:255" json:"custom_wallpaper"`
	EnableWallpaper bool       `json:"enable_wallpaper"`
	DefaultCover    string     `gorm:"size:255" json:"default_cover"`
	ICPInfo         string     `gorm:"size:255" json:"icp_info"`
	SiteCreatedAt   *time.Time `gorm:"column:site_created_at" json:"site_created_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (SiteSetting) TableName() string { return "site_settings" }
