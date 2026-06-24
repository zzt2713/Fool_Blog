package model

import "time"

type OperationLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OperatorID uint      `gorm:"index" json:"operator_id"`
	Operator   string    `gorm:"size:64" json:"operator"`
	Module     string    `gorm:"size:64;index" json:"module"`
	Action     string    `gorm:"size:64;index" json:"action"`
	TargetType string    `gorm:"size:64" json:"target_type"`
	TargetID   string    `gorm:"size:64" json:"target_id"`
	Summary    string    `gorm:"size:512" json:"summary"`
	IP         string    `gorm:"size:64" json:"ip"`
	UserAgent  string    `gorm:"size:255" json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

func (OperationLog) TableName() string { return "operation_logs" }
