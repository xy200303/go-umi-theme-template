package requests

type InitDirectUploadReq struct {
	FileMD5  string `json:"file_md5" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required"`
	MimeType string `json:"mime_type"`
}

type CompleteDirectUploadReq struct {
	FileID   string `json:"file_id" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required"`
	MimeType string `json:"mime_type"`
}
