package requests

type UpdateProfileReq struct {
	Email        string `json:"email"`
	AvatarFileID string `json:"avatar_file_id"`
	Signature    string `json:"signature"`
	Gender       string `json:"gender"`
	Age          int    `json:"age"`
}

type ResetPasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ChangePhoneReq struct {
	OldPhoneCode string `json:"old_phone_code"`
	NewPhone     string `json:"new_phone" binding:"required"`
	NewPhoneCode string `json:"new_phone_code"`
}

type CreateUserInterfaceReq struct {
	Name          string `json:"name" binding:"required"`
	InterfaceType string `json:"interface_type" binding:"required"`
	TargetBaseURL string `json:"target_base_url" binding:"required"`
	TargetAPIKey  string `json:"target_api_key" binding:"required"`
	DefaultModel  string `json:"default_model"`
	Enabled       *bool  `json:"enabled"`
}

type UpdateUserInterfaceReq struct {
	Name          string `json:"name" binding:"required"`
	InterfaceType string `json:"interface_type" binding:"required"`
	TargetBaseURL string `json:"target_base_url" binding:"required"`
	TargetAPIKey  string `json:"target_api_key"`
	DefaultModel  string `json:"default_model"`
	Enabled       *bool  `json:"enabled"`
}

type ListGatewayRequestLogsReq struct {
	Keyword     string `form:"keyword"`
	InterfaceID uint   `form:"interface_id"`
	StatusCode  int    `form:"status_code"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

type ListChatRecordsReq struct {
	Keyword     string `form:"keyword"`
	InterfaceID uint   `form:"interface_id"`
	StatusCode  int    `form:"status_code"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}
