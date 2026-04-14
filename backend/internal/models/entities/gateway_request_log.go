package entities

import "time"

type GatewayRequestLog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index;not null" json:"user_id"`
	UserInterfaceID uint      `gorm:"index;not null" json:"user_interface_id"`
	InterfaceType   string    `gorm:"size:32;index;not null" json:"interface_type"`
	RequestMethod   string    `gorm:"size:12;index;not null" json:"request_method"`
	RequestPath     string    `gorm:"size:255;index;not null" json:"request_path"`
	UpstreamURL     string    `gorm:"size:512" json:"upstream_url"`
	Model           string    `gorm:"size:120" json:"model"`
	StatusCode      int       `gorm:"index;not null" json:"status_code"`
	DurationMS      int64     `gorm:"not null" json:"duration_ms"`
	ErrorMessage    string    `gorm:"size:512" json:"error_message"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
}
