package entities

import "time"

type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Username       string    `gorm:"size:32;uniqueIndex;not null" json:"username"`
	Phone          string    `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	PasswordHash   string    `gorm:"size:255;not null" json:"-"`
	SessionVersion int       `gorm:"not null;default:0" json:"-"`
	Email          string    `gorm:"size:120;uniqueIndex:idx_users_email,where:email <> ''" json:"email"`
	AvatarURL      string    `gorm:"size:512" json:"avatar_url"`
	Signature      string    `gorm:"size:255" json:"signature"`
	Gender         string    `gorm:"size:10" json:"gender"`
	Age            int       `json:"age"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Roles          []Role    `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	DisplayName string    `gorm:"size:100;not null" json:"display_name"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserRole struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	RoleID    uint      `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}
