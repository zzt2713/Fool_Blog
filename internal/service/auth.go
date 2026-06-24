package service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"fool_blog_go/internal/model"
	"fool_blog_go/internal/util"
)

func Register(db *gorm.DB, username, password, nickname, email string) (*model.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	if len(password) < 6 {
		return nil, errors.New("密码至少 6 位")
	}
	var exist int64
	db.Model(&model.User{}).Where("username = ?", username).Count(&exist)
	if exist > 0 {
		return nil, errors.New("用户名已被使用")
	}
	hash, err := util.HashPassword(password)
	if err != nil {
		return nil, err
	}
	if nickname == "" {
		nickname = username
	}
	u := &model.User{
		Username:     username,
		PasswordHash: hash,
		Nickname:     nickname,
		Email:        email,
		Role:         model.RoleUser,
		Status:       model.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func Authenticate(db *gorm.DB, username, password string) (*model.User, error) {
	var u model.User
	if err := db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	if u.Status == model.UserStatusDisabled {
		return nil, errors.New("账号已被禁用")
	}
	if !util.CheckPassword(u.PasswordHash, password) {
		return nil, errors.New("密码错误")
	}
	now := time.Now()
	u.LastLoginAt = &now
	db.Model(&u).Update("last_login_at", now)
	return &u, nil
}
