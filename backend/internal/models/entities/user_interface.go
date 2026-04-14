package entities

import "time"

type UserInterface struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	UserID                uint       `gorm:"index;not null" json:"user_id"`
	Name                  string     `gorm:"size:100;not null" json:"name"`
	InterfaceType         string     `gorm:"size:32;index;not null" json:"interface_type"`
	TargetBaseURL         string     `gorm:"size:512;not null" json:"target_base_url"`
	TargetAPIKeyEncrypted string     `gorm:"type:text;not null" json:"-"`
	TargetAPIKeyMask      string     `gorm:"size:64;not null" json:"target_api_key_mask"`
	DefaultModel          string     `gorm:"size:120" json:"default_model"`
	Enabled               bool       `gorm:"default:true" json:"enabled"`
	GatewayKeyEncrypted   string     `gorm:"type:text" json:"-"`
	GatewayKeyHash        string     `gorm:"size:128;uniqueIndex;not null" json:"-"`
	GatewayKeyPrefix      string     `gorm:"size:32;index;not null" json:"gateway_key_prefix"`
	LastUsedAt            *time.Time `json:"last_used_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
