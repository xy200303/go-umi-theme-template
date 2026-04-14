package response

type GatewayRequestLogResp struct {
	ID              uint   `json:"id"`
	UserID          uint   `json:"user_id"`
	UserInterfaceID uint   `json:"user_interface_id"`
	InterfaceType   string `json:"interface_type"`
	RequestMethod   string `json:"request_method"`
	RequestPath     string `json:"request_path"`
	UpstreamURL     string `json:"upstream_url"`
	Model           string `json:"model"`
	StatusCode      int    `json:"status_code"`
	DurationMS      int64  `json:"duration_ms"`
	ErrorMessage    string `json:"error_message"`
	CreatedAt       string `json:"created_at"`
}

type GatewayRequestLogListResp struct {
	List     []GatewayRequestLogResp `json:"list"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}
