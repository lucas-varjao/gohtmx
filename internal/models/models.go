// Package models defines the data models for the application.
package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	gorm.Model
	// Identity information
	Username     string `json:"username"     gorm:"unique;not null;index"`
	Email        string `json:"email"        gorm:"unique;not null;index"`
	DisplayName  string `json:"display_name" gorm:"not null"`
	PasswordHash string `json:"-"            gorm:"not null"`

	// Profile information
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`

	// Account status
	Active        bool      `json:"active"         gorm:"default:true"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	LastLogin     time.Time `json:"last_login"`
	LastActive    time.Time `json:"last_active"`

	// Access control
	Role        string `json:"role"                  gorm:"default:user"`
	Permissions string `json:"permissions,omitempty" gorm:"type:text"` // JSON string of permissions

	// Password reset (kept separate from session management)
	ResetToken       string    `json:"-"`
	ResetTokenExpiry time.Time `json:"-"`
}
