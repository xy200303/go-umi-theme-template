package response

type FileUploadResp struct {
	ID            string `json:"id"`
	StorageDriver string `json:"storage_driver"`
	OriginalName  string `json:"original_name"`
	Ext           string `json:"ext"`
	MimeType      string `json:"mime_type"`
	Size          int64  `json:"size"`
	UploadStatus  string `json:"upload_status"`
	FileURL       string `json:"file_url"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type FileDirectUploadInitResp struct {
	Mode          string            `json:"mode"`
	AlreadyExists bool              `json:"already_exists"`
	FileID        string            `json:"file_id"`
	StorageDriver string            `json:"storage_driver"`
	UploadMethod  string            `json:"upload_method,omitempty"`
	UploadURL     string            `json:"upload_url,omitempty"`
	UploadHeaders map[string]string `json:"upload_headers,omitempty"`
	ExpiresAt     string            `json:"expires_at,omitempty"`
	File          *FileUploadResp   `json:"file,omitempty"`
}
