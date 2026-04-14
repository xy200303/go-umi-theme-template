package apigatewaysvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"backend/internal/pkg/utils"
)

type GatewayTarget struct {
	InterfaceType string
	TargetURL     string
	TargetAPIKey  string
	DefaultModel  string
	UserID        uint
	InterfaceID   uint
}

func extractGatewayKey(req *http.Request) string {
	if token := strings.TrimSpace(extractBearerToken(req.Header.Get("Authorization"))); token != "" {
		return token
	}
	if token := strings.TrimSpace(req.Header.Get("x-api-key")); token != "" {
		return token
	}
	if token := strings.TrimSpace(req.Header.Get("x-goog-api-key")); token != "" {
		return token
	}
	if token := strings.TrimSpace(req.URL.Query().Get("key")); token != "" {
		return token
	}
	return ""
}

func extractBearerToken(headerValue string) string {
	parts := strings.Fields(strings.TrimSpace(headerValue))
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return parts[1]
	}
	return ""
}

func buildProxyTargetURL(storedTargetBaseURL string, interfaceType string, requestPath string, rawQuery string) (string, error) {
	parsed, err := url.Parse(storedTargetBaseURL)
	if err != nil {
		return "", utils.NewAppError(http.StatusBadGateway, utils.ErrCodeInternal, "invalid stored target base url")
	}

	switch interfaceType {
	case "gemini":
		geminiPath := strings.TrimPrefix(requestPath, "/gemini")
		if !strings.HasPrefix(geminiPath, "/v1beta/models/") && !strings.HasPrefix(geminiPath, "/v1/models/") {
			return "", utils.NewAppError(http.StatusBadRequest, utils.ErrCodeInvalidRequest, "invalid gemini request path")
		}
		parsed.Path = geminiPath
	default:
	}

	query := parsed.Query()
	incomingQuery, err := url.ParseQuery(rawQuery)
	if err == nil {
		for key, values := range incomingQuery {
			if interfaceType == "gemini" && strings.EqualFold(key, "key") {
				continue
			}
			query.Del(key)
			for _, value := range values {
				query.Add(key, value)
			}
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func copyForwardHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		if shouldSkipRequestHeader(key) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func shouldSkipRequestHeader(key string) bool {
	switch strings.ToLower(key) {
	case "authorization", "x-api-key", "x-goog-api-key", "host", "content-length", "connection", "proxy-connection", "keep-alive", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func applyTargetAuthorization(req *http.Request, interfaceType string, targetAPIKey string) {
	switch interfaceType {
	case "claude":
		req.Header.Set("x-api-key", targetAPIKey)
	case "gemini":
		req.Header.Set("x-goog-api-key", targetAPIKey)
	default:
		req.Header.Set("Authorization", "Bearer "+targetAPIKey)
	}
}

func copyResponseHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		if shouldSkipResponseHeader(key) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func shouldSkipResponseHeader(key string) bool {
	switch strings.ToLower(key) {
	case "connection", "proxy-connection", "keep-alive", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func prepareForwardBody(req *http.Request, interfaceType string, defaultModel string) (io.Reader, int64, error) {
	if req.Body == nil {
		return nil, 0, nil
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read request body: %w", err)
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(rawBody))

	if strings.TrimSpace(defaultModel) == "" || len(rawBody) == 0 || !strings.Contains(strings.ToLower(req.Header.Get("Content-Type")), "application/json") {
		return bytes.NewReader(rawBody), int64(len(rawBody)), nil
	}

	nextBody, changed, err := injectDefaultModel(rawBody, interfaceType, defaultModel)
	if err != nil {
		return nil, 0, err
	}
	if !changed {
		return bytes.NewReader(rawBody), int64(len(rawBody)), nil
	}
	return bytes.NewReader(nextBody), int64(len(nextBody)), nil
}

func injectDefaultModel(rawBody []byte, interfaceType string, defaultModel string) ([]byte, bool, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return rawBody, false, nil
	}

	switch interfaceType {
	case "openai_api", "openai_response", "claude":
	default:
		return rawBody, false, nil
	}

	if current, ok := payload["model"]; ok {
		if text, ok := current.(string); ok && strings.TrimSpace(text) != "" {
			return rawBody, false, nil
		}
	}

	payload["model"] = defaultModel
	nextBody, err := json.Marshal(payload)
	if err != nil {
		return nil, false, fmt.Errorf("failed to encode request body: %w", err)
	}
	return nextBody, true, nil
}

func buildConnectivityTestPayload(interfaceType string, defaultModel string) ([]byte, string, error) {
	var payload map[string]interface{}

	switch interfaceType {
	case "openai_api":
		payload = map[string]interface{}{
			"model":      defaultModel,
			"messages":   []map[string]string{{"role": "user", "content": "ping"}},
			"max_tokens": 1,
		}
	case "openai_response":
		payload = map[string]interface{}{
			"model":             defaultModel,
			"input":             "ping",
			"max_output_tokens": 1,
		}
	case "claude":
		payload = map[string]interface{}{
			"model":      defaultModel,
			"messages":   []map[string]string{{"role": "user", "content": "ping"}},
			"max_tokens": 1,
		}
	case "gemini":
		payload = map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]string{{"text": "ping"}},
				},
			},
			"generationConfig": map[string]int{
				"maxOutputTokens": 1,
			},
		}
	default:
		return nil, "", fmt.Errorf("unsupported interface type")
	}

	rawBody, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encode test payload: %w", err)
	}
	return rawBody, "application/json", nil
}

func truncateText(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	return value[:maxLen]
}

type GatewayRequestInsight struct {
	Model  string
	Inputs []string
}

type ChatResponseInsight struct {
	Model  string
	Output string
}

func extractGatewayRequestInsight(rawBody []byte, interfaceType string, fallbackModel string) GatewayRequestInsight {
	insight := GatewayRequestInsight{
		Model: strings.TrimSpace(fallbackModel),
	}
	if len(rawBody) == 0 {
		return insight
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return insight
	}

	if model, ok := payload["model"].(string); ok && strings.TrimSpace(model) != "" {
		insight.Model = strings.TrimSpace(model)
	}

	insight.Inputs = collectPayloadSnippets(payload, interfaceType)
	return insight
}

func collectPayloadSnippets(payload map[string]interface{}, interfaceType string) []string {
	var snippets []string

	appendSnippet := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			snippets = append(snippets, trimmed)
		}
	}

	switch interfaceType {
	case "openai_api", "claude":
		if messages, ok := payload["messages"].([]interface{}); ok {
			for _, message := range messages {
				if msgMap, ok := message.(map[string]interface{}); ok {
					appendSnippet(extractTextFromContent(msgMap["content"]))
				}
			}
		}
	case "openai_response":
		appendSnippet(extractTextFromContent(payload["input"]))
	case "gemini":
		if contents, ok := payload["contents"].([]interface{}); ok {
			for _, content := range contents {
				if contentMap, ok := content.(map[string]interface{}); ok {
					appendSnippet(extractTextFromContent(contentMap["parts"]))
				}
			}
		}
	}

	return snippets
}

func extractTextFromContent(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return value
	case []interface{}:
		parts := make([]string, 0, len(value))
		for _, item := range value {
			switch typed := item.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					parts = append(parts, typed)
				}
			case map[string]interface{}:
				if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
					parts = append(parts, text)
				}
				if inputText, ok := typed["input_text"].(string); ok && strings.TrimSpace(inputText) != "" {
					parts = append(parts, inputText)
				}
				if inner, ok := typed["content"]; ok {
					if extracted := extractTextFromContent(inner); strings.TrimSpace(extracted) != "" {
						parts = append(parts, extracted)
					}
				}
			}
		}
		return strings.Join(parts, " ")
	case map[string]interface{}:
		if text, ok := value["text"].(string); ok {
			return text
		}
		if content, ok := value["content"]; ok {
			return extractTextFromContent(content)
		}
	}
	return ""
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func inspectResponseBody(rawBody []byte, contentType string, interfaceType string) ChatResponseInsight {
	trimmedBody := bytes.TrimSpace(rawBody)
	if len(trimmedBody) == 0 {
		return ChatResponseInsight{}
	}

	if strings.Contains(strings.ToLower(contentType), "text/event-stream") {
		return extractStreamingResponseInsight(trimmedBody, interfaceType)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(trimmedBody, &payload); err != nil {
		return ChatResponseInsight{}
	}

	return extractResponseInsightFromPayload(payload, interfaceType)
}

func extractStreamingResponseInsight(rawBody []byte, interfaceType string) ChatResponseInsight {
	var payloads []map[string]interface{}

	for _, line := range strings.Split(string(rawBody), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			continue
		}
		payloads = append(payloads, payload)
	}

	if len(payloads) == 0 {
		return ChatResponseInsight{}
	}

	combined := ChatResponseInsight{}
	var outputs []string
	for _, payload := range payloads {
		item := extractResponseInsightFromPayload(payload, interfaceType)
		if combined.Model == "" && item.Model != "" {
			combined.Model = item.Model
		}
		if item.Output != "" {
			outputs = append(outputs, item.Output)
		}
	}
	combined.Output = strings.TrimSpace(strings.Join(outputs, " "))
	return combined
}

func extractResponseInsightFromPayload(payload map[string]interface{}, interfaceType string) ChatResponseInsight {
	insight := ChatResponseInsight{}

	if model, ok := payload["model"].(string); ok && strings.TrimSpace(model) != "" {
		insight.Model = strings.TrimSpace(model)
	}

	switch interfaceType {
	case "openai_api":
		insight.Output = extractOpenAIChatOutput(payload)
	case "openai_response":
		insight.Output = extractOpenAIResponseOutput(payload)
	case "claude":
		insight.Output = extractClaudeOutput(payload)
	case "gemini":
		insight.Output = extractGeminiOutput(payload)
	}

	return insight
}

func extractOpenAIChatOutput(payload map[string]interface{}) string {
	choices, ok := payload["choices"].([]interface{})
	if !ok {
		return ""
	}

	var parts []string
	for _, choice := range choices {
		choiceMap, ok := choice.(map[string]interface{})
		if !ok {
			continue
		}
		if message, ok := choiceMap["message"].(map[string]interface{}); ok {
			if text := extractTextFromContent(message["content"]); strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		}
		if delta, ok := choiceMap["delta"].(map[string]interface{}); ok {
			if text := extractTextFromContent(delta["content"]); strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func extractOpenAIResponseOutput(payload map[string]interface{}) string {
	if outputText, ok := payload["output_text"].(string); ok && strings.TrimSpace(outputText) != "" {
		return strings.TrimSpace(outputText)
	}
	if output, ok := payload["output"].([]interface{}); ok {
		var parts []string
		for _, item := range output {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if text := extractTextFromContent(itemMap["content"]); strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, " "))
	}
	return ""
}

func extractClaudeOutput(payload map[string]interface{}) string {
	return extractTextFromContent(payload["content"])
}

func extractGeminiOutput(payload map[string]interface{}) string {
	candidates, ok := payload["candidates"].([]interface{})
	if !ok {
		return ""
	}

	var parts []string
	for _, candidate := range candidates {
		candidateMap, ok := candidate.(map[string]interface{})
		if !ok {
			continue
		}
		if content, ok := candidateMap["content"].(map[string]interface{}); ok {
			if text := extractTextFromContent(content["parts"]); strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
