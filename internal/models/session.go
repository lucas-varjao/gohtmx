package models

import (
	"time"
)

// Session represents an authentication session stored in the database
type Session struct {
	ID        string    `json:"id"                   gorm:"primaryKey;type:varchar(64)"`
	UserID    uint      `json:"user_id"              gorm:"index;not null"`
	ExpiresAt time.Time `json:"expires_at"           gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at"`
	UserAgent string    `json:"user_agent,omitempty" gorm:"type:varchar(500)"`
	IP        string    `json:"ip,omitempty"         gorm:"type:varchar(45)"` // Supports IPv6
}

// TableName specifies the table name for GORM
func (Session) TableName() string {
	return "sessions"
}
