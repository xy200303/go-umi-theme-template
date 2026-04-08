package response

type PolicyTemplateResp struct {
	Key         string `json:"key"`
	MenuKey     string `json:"menu_key"`
	MenuLabel   string `json:"menu_label"`
	ActionLabel string `json:"action_label"`
	Description string `json:"description,omitempty"`
	Method      string `json:"method"`
	Path        string `json:"path"`
}

type AuditLogResp struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	Username      string `json:"username"`
	Method        string `json:"method"`
	RoutePath     string `json:"route_path"`
	RequestPath   string `json:"request_path"`
	OperationID   string `json:"operation_id"`
	OperationName string `json:"operation_name"`
	MenuKey       string `json:"menu_key"`
	MenuLabel     string `json:"menu_label"`
	StatusCode    int    `json:"status_code"`
	ClientIP      string `json:"client_ip"`
	UserAgent     string `json:"user_agent"`
	DurationMS    int64  `json:"duration_ms"`
	CreatedAt     string `json:"created_at"`
}

type AuditLogListResp struct {
	List     []AuditLogResp `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type AdminFileResp struct {
	ID            string `json:"id"`
	StorageDriver string `json:"storage_driver"`
	StoragePath   string `json:"storage_path"`
	OriginalName  string `json:"original_name"`
	Ext           string `json:"ext"`
	MimeType      string `json:"mime_type"`
	Size          int64  `json:"size"`
	UploadStatus  string `json:"upload_status"`
	UploadedBy    uint   `json:"uploaded_by"`
	Remark        string `json:"remark"`
	FileURL       string `json:"file_url"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type AdminFileListResp struct {
	List     []AdminFileResp `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type AdminFileStatsResp struct {
	TotalCount    int64 `json:"total_count"`
	UploadedCount int64 `json:"uploaded_count"`
	BoundCount    int64 `json:"bound_count"`
	DeletedCount  int64 `json:"deleted_count"`
}
