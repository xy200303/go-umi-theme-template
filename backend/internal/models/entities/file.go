package entities

import "time"

type File struct {
	ID            string    `gorm:"primaryKey;size:32" json:"id"`
	StorageDriver string    `gorm:"size:32;not null;index" json:"storage_driver"`
	StoragePath   string    `gorm:"type:text;not null" json:"storage_path"`
	OriginalName  string    `gorm:"size:255;not null" json:"original_name"`
	Ext           string    `gorm:"size:32;not null;default:''" json:"ext"`
	MimeType      string    `gorm:"size:128;not null;default:''" json:"mime_type"`
	Size          int64     `gorm:"not null;default:0" json:"size"`
	UploadStatus  string    `gorm:"size:16;not null;default:'uploaded';index" json:"upload_status"`
	UploadedBy    uint      `gorm:"not null;default:0;index" json:"uploaded_by"`
	Remark        string    `gorm:"size:255;not null;default:''" json:"remark"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
