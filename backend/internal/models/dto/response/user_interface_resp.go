package response

import "time"

type UserInterfaceResp struct {
	ID               uint       `json:"id"`
	Name             string     `json:"name"`
	InterfaceType    string     `json:"interface_type"`
	TargetBaseURL    string     `json:"target_base_url"`
	TargetAPIKeyMask string     `json:"target_api_key_mask"`
	DefaultModel     string     `json:"default_model"`
	Enabled          bool       `json:"enabled"`
	GatewayKey       string     `json:"gateway_key"`
	GatewayKeyPrefix string     `json:"gateway_key_prefix"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateUserInterfaceResp struct {
	Interface  UserInterfaceResp `json:"interface"`
	GatewayKey string            `json:"gateway_key"`
}

type RegenerateUserInterfaceGatewayKeyResp struct {
	Interface  UserInterfaceResp `json:"interface"`
	GatewayKey string            `json:"gateway_key"`
}

type TestUserInterfaceResp struct {
	OK         bool   `json:"ok"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

type GatewayDisplayConfigResp struct {
	GatewayBaseURL string `json:"gateway_base_url"`
}
