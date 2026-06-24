package model

import "time"

type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Slug      string    `gorm:"size:64;index" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Tag) TableName() string { return "tags" }
