package response

type ChatRecordResp struct {
	ID              uint   `json:"id"`
	UserID          uint   `json:"user_id"`
	UserInterfaceID uint   `json:"user_interface_id"`
	InterfaceType   string `json:"interface_type"`
	RequestMethod   string `json:"request_method"`
	RequestPath     string `json:"request_path"`
	Model           string `json:"model"`
	UserInput       string `json:"user_input"`
	ModelOutput     string `json:"model_output"`
	StatusCode      int    `json:"status_code"`
	DurationMS      int64  `json:"duration_ms"`
	ErrorMessage    string `json:"error_message"`
	CreatedAt       string `json:"created_at"`
}

type ChatRecordListResp struct {
	List     []ChatRecordResp `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}
