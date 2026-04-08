package requests

type UpdateUserRolesReq struct {
	RoleNames []string `json:"role_names" binding:"required"`
}

type CreateUserReq struct {
	Username     string   `json:"username" binding:"required"`
	Phone        string   `json:"phone" binding:"required"`
	Password     string   `json:"password" binding:"required"`
	Email        string   `json:"email"`
	AvatarFileID string   `json:"avatar_file_id"`
	Signature    string   `json:"signature"`
	Gender       string   `json:"gender"`
	Age          int      `json:"age"`
	IsActive     bool     `json:"is_active"`
	RoleNames    []string `json:"role_names"`
}

type UpdateUserReq struct {
	Username     string `json:"username" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Email        string `json:"email"`
	AvatarFileID string `json:"avatar_file_id"`
	Signature    string `json:"signature"`
	Gender       string `json:"gender"`
	Age          int    `json:"age"`
	IsActive     bool   `json:"is_active"`
}

type ResetUserPasswordReq struct {
	Password string `json:"password" binding:"required"`
}

type CreateRoleReq struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
}

type UpdateRoleReq struct {
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
}

type CasbinPolicyReq struct {
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}

type SetRolePoliciesReq struct {
	Policies []CasbinPolicyReq `json:"policies" binding:"required"`
}

type SystemConfigReq struct {
	ConfigGroup string `json:"config_group" binding:"required"`
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigVal   string `json:"config_val" binding:"required"`
	Remark      string `json:"remark"`
}

type ListAuditLogsReq struct {
	Keyword    string `form:"keyword"`
	MenuKey    string `form:"menu_key"`
	StatusCode int    `form:"status_code"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type ListAdminFilesReq struct {
	Keyword      string `form:"keyword"`
	UploadStatus string `form:"upload_status"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}
