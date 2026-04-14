package apigatewaysvc

import (
	"fmt"
	"net/url"
	"strings"
)

func normalizeInterfaceType(raw string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "openai_api":
		return "openai_api", nil
	case "openai_response":
		return "openai_response", nil
	case "claude":
		return "claude", nil
	case "gemini":
		return "gemini", nil
	default:
		return "", fmt.Errorf("unsupported interface type")
	}
}

func normalizeTargetBaseURL(raw string, interfaceType string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("target base url is required")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid target base url")
	}
	parsed.Path = normalizeTargetFullPath(parsed.Path, interfaceType)
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func normalizeTargetFullPath(path string, interfaceType string) string {
	cleaned := "/" + strings.Trim(strings.TrimSpace(path), "/")
	if cleaned == "/" {
		return defaultTargetPath(interfaceType)
	}

	switch interfaceType {
	case "openai_api":
		return appendMissingPathSegments(cleaned, []string{"v1", "chat", "completions"})
	case "openai_response":
		return appendMissingPathSegments(cleaned, []string{"v1", "responses"})
	case "claude":
		return appendMissingPathSegments(cleaned, []string{"v1", "messages"})
	case "gemini":
		return normalizeGeminiPath(cleaned)
	default:
		return cleaned
	}
}

func defaultTargetPath(interfaceType string) string {
	switch interfaceType {
	case "openai_response":
		return "/v1/responses"
	case "claude":
		return "/v1/messages"
	case "gemini":
		return "/v1beta/models/{model}:generateContent"
	default:
		return "/v1/chat/completions"
	}
}

func appendMissingPathSegments(cleaned string, requiredSegments []string) string {
	currentSegments := splitPathSegments(cleaned)
	if len(currentSegments) == 0 {
		return "/" + strings.Join(requiredSegments, "/")
	}

	overlap := longestSuffixPrefixOverlap(currentSegments, requiredSegments)
	fullSegments := append(append([]string{}, currentSegments...), requiredSegments[overlap:]...)
	return "/" + strings.Join(fullSegments, "/")
}

func normalizeGeminiPath(cleaned string) string {
	lower := strings.ToLower(cleaned)
	if strings.Contains(lower, "/models/") && strings.Contains(lower, ":generatecontent") {
		return cleaned
	}
	return appendMissingPathSegments(cleaned, []string{"v1beta", "models", "{model}:generateContent"})
}

func splitPathSegments(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func longestSuffixPrefixOverlap(currentSegments []string, requiredSegments []string) int {
	maxOverlap := len(requiredSegments)
	if len(currentSegments) < maxOverlap {
		maxOverlap = len(currentSegments)
	}

	for overlap := maxOverlap; overlap >= 1; overlap-- {
		start := len(currentSegments) - overlap
		matched := true
		for index := 0; index < overlap; index++ {
			if currentSegments[start+index] != requiredSegments[index] {
				matched = false
				break
			}
		}
		if matched {
			return overlap
		}
	}
	return 0
}
