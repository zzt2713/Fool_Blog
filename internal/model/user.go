package model

import "time"

const (
	RoleUser  = 0
	RoleAdmin = 1

	UserStatusActive   = 0
	UserStatusDisabled = 1
)

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Nickname     string     `gorm:"size:64" json:"nickname"`
	Email        string     `gorm:"size:128" json:"email"`
	Avatar       string     `gorm:"size:255" json:"avatar"`
	Role         int        `gorm:"default:0;index" json:"role"`
	Status       int        `gorm:"default:0;index" json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at"`
}

func (User) TableName() string { return "users" }

func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }
