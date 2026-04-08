package utils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func FallbackDownloadFilename(storagePath string) string {
	base := filepath.Base(strings.TrimSpace(storagePath))
	if base == "" || base == "." || base == "/" {
		return "download"
	}
	return base
}

func SanitizeDownloadFilename(fileName string, fallback string) string {
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = strings.TrimSpace(fallback)
	}
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	if name == "" {
		name = "download"
	}
	return name
}

func ContentDispositionAttachmentValue(fileName string) string {
	safe := SanitizeDownloadFilename(fileName, "download")
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, safe, url.QueryEscape(safe))
}
