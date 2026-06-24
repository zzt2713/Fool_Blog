package handler

import (
	"gorm.io/gorm"

	"fool_blog_go/internal/config"
	"fool_blog_go/internal/model"
)

type Handler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func New(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{DB: db, Cfg: cfg}
}

func (h *Handler) loadSite() model.SiteSetting {
	var s model.SiteSetting
	h.DB.First(&s)
	return s
}
