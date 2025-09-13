package models

import "time"

type ActivityLog struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Action   string    `gorm:"type:varchar(50);not null" json:"action"`
	PostID   uint      `gorm:"index;not null" json:"post_id"`
	LoggedAt time.Time `gorm:"autoCreateTime" json:"logged_at"`
} 