package mapper

import (
	"regexp"
	"strings"
)

var fileIDValuePattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func ResolveStoredFileURL(value string, buildDownloadURL func(string) string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return ""
	}

	if buildDownloadURL != nil && fileIDValuePattern.MatchString(strings.ToLower(normalized)) {
		if target := buildDownloadURL(normalized); target != "" {
			return target
		}
	}

	return normalized
}
