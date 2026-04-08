package entities

import "time"

type AuditLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	Username      string    `gorm:"size:64;index;not null" json:"username"`
	Method        string    `gorm:"size:12;index;not null" json:"method"`
	RoutePath     string    `gorm:"size:255;index;not null" json:"route_path"`
	RequestPath   string    `gorm:"size:255;not null" json:"request_path"`
	OperationID   string    `gorm:"size:128;index" json:"operation_id"`
	OperationName string    `gorm:"size:128" json:"operation_name"`
	MenuKey       string    `gorm:"size:64;index" json:"menu_key"`
	MenuLabel     string    `gorm:"size:64" json:"menu_label"`
	StatusCode    int       `gorm:"index;not null" json:"status_code"`
	ClientIP      string    `gorm:"size:64" json:"client_ip"`
	UserAgent     string    `gorm:"size:512" json:"user_agent"`
	DurationMS    int64     `gorm:"not null" json:"duration_ms"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}
