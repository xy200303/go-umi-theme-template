package apigatewaysvc

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"backend/internal/models/dto/response"
	"backend/internal/models/entities"
	"backend/internal/pkg/utils"
	"gorm.io/gorm"
)

func (s *APIGatewayService) ForwardRequest(writer http.ResponseWriter, req *http.Request, expectedInterfaceType string) error {
	startedAt := time.Now()
	gatewayKey := extractGatewayKey(req)
	if strings.TrimSpace(gatewayKey) == "" {
		return utils.NewAppError(http.StatusUnauthorized, utils.ErrCodeInvalidRequest, "gateway key is required")
	}

	target, err := s.resolveGatewayTarget(gatewayKey, expectedInterfaceType, req.URL.Path, req.URL.RawQuery)
	if err != nil {
		return err
	}

	bodyReader, contentLength, err := prepareForwardBody(req, target.InterfaceType, target.DefaultModel)
	if err != nil {
		return err
	}
	requestInsight, _ := inspectPreparedRequestBody(req, target.InterfaceType, target.DefaultModel)

	upstreamReq, err := http.NewRequestWithContext(req.Context(), req.Method, target.TargetURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create upstream request: %w", err)
	}
	upstreamReq.ContentLength = contentLength

	copyForwardHeaders(upstreamReq.Header, req.Header)
	applyTargetAuthorization(upstreamReq, target.InterfaceType, target.TargetAPIKey)

	resp, err := s.httpClient.Do(upstreamReq)
	if err != nil {
		s.logGatewayRequest(target, req.Method, req.URL.Path, 0, time.Since(startedAt), err.Error(), requestInsight)
		s.logChatRecord(target, req.Method, req.URL.Path, 0, time.Since(startedAt), err.Error(), requestInsight, ChatResponseInsight{})
		return fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logGatewayRequest(target, req.Method, req.URL.Path, resp.StatusCode, time.Since(startedAt), err.Error(), requestInsight)
		s.logChatRecord(target, req.Method, req.URL.Path, resp.StatusCode, time.Since(startedAt), err.Error(), requestInsight, ChatResponseInsight{})
		return fmt.Errorf("failed to read upstream response: %w", err)
	}
	responseInsight := inspectResponseBody(rawResponse, resp.Header.Get("Content-Type"), target.InterfaceType)

	copyResponseHeaders(writer.Header(), resp.Header)
	writer.WriteHeader(resp.StatusCode)
	if _, err := writer.Write(rawResponse); err != nil {
		s.logGatewayRequest(target, req.Method, req.URL.Path, resp.StatusCode, time.Since(startedAt), err.Error(), requestInsight)
		s.logChatRecord(target, req.Method, req.URL.Path, resp.StatusCode, time.Since(startedAt), err.Error(), requestInsight, responseInsight)
		return fmt.Errorf("failed to stream upstream response: %w", err)
	}
	s.logGatewayRequest(target, req.Method, req.URL.Path, resp.StatusCode, time.Since(startedAt), "", requestInsight)
	s.logChatRecord(target, req.Method, req.URL.Path, resp.StatusCode, time.Since(startedAt), "", requestInsight, responseInsight)
	return nil
}

func (s *APIGatewayService) TestInterfaceConnection(userID uint, interfaceID uint) (*response.TestUserInterfaceResp, error) {
	target, err := s.resolveInterfaceTargetForUser(userID, interfaceID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(target.DefaultModel) == "" {
		return nil, fmt.Errorf("default model is required for connection test")
	}

	body, contentType, err := buildConnectivityTestPayload(target.InterfaceType, target.DefaultModel)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, target.TargetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create test request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	applyTargetAuthorization(req, target.InterfaceType, target.TargetAPIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &response.TestUserInterfaceResp{
			OK:         false,
			StatusCode: 0,
			Message:    err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	message := "connection succeeded"
	if resp.StatusCode >= http.StatusBadRequest {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		message = strings.TrimSpace(string(payload))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
	}

	return &response.TestUserInterfaceResp{
		OK:         resp.StatusCode >= 200 && resp.StatusCode < 400,
		StatusCode: resp.StatusCode,
		Message:    message,
	}, nil
}

func (s *APIGatewayService) resolveGatewayTarget(gatewayKey string, expectedInterfaceType string, requestPath string, rawQuery string) (*GatewayTarget, error) {
	item, err := s.userInterfaceRepo.FindByGatewayKeyHash(utils.HashToken(gatewayKey))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.NewAppError(http.StatusUnauthorized, utils.ErrCodeForbidden, "invalid gateway key")
		}
		return nil, err
	}

	if !item.Enabled {
		return nil, utils.NewAppError(http.StatusForbidden, utils.ErrCodeForbidden, "interface is disabled")
	}
	if item.InterfaceType != expectedInterfaceType {
		return nil, utils.NewAppError(http.StatusBadRequest, utils.ErrCodeInvalidRequest, "gateway key does not match this interface type")
	}

	targetAPIKey, err := utils.DecryptString(s.cfg.MindGateKeyEncryptionSecret, item.TargetAPIKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt target api key: %w", err)
	}

	targetURL, err := buildProxyTargetURL(item.TargetBaseURL, item.InterfaceType, requestPath, rawQuery)
	if err != nil {
		return nil, err
	}
	if err := s.userInterfaceRepo.TouchLastUsedAt(item.ID, time.Now()); err != nil {
		return nil, err
	}

	return &GatewayTarget{
		InterfaceType: item.InterfaceType,
		TargetURL:     targetURL,
		TargetAPIKey:  targetAPIKey,
		DefaultModel:  strings.TrimSpace(item.DefaultModel),
		UserID:        item.UserID,
		InterfaceID:   item.ID,
	}, nil
}

func (s *APIGatewayService) resolveInterfaceTargetForUser(userID uint, interfaceID uint) (*GatewayTarget, error) {
	item, err := s.userInterfaceRepo.FindByIDForUser(interfaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("interface not found")
	}
	if !item.Enabled {
		return nil, fmt.Errorf("interface is disabled")
	}

	targetAPIKey, err := utils.DecryptString(s.cfg.MindGateKeyEncryptionSecret, item.TargetAPIKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt target api key: %w", err)
	}

	targetURL := item.TargetBaseURL
	if item.InterfaceType == "gemini" {
		model := strings.TrimSpace(item.DefaultModel)
		if model == "" {
			model = "{model}"
		}
		targetURL = strings.ReplaceAll(item.TargetBaseURL, "{model}", model)
	}

	return &GatewayTarget{
		InterfaceType: item.InterfaceType,
		TargetURL:     targetURL,
		TargetAPIKey:  targetAPIKey,
		DefaultModel:  strings.TrimSpace(item.DefaultModel),
		UserID:        item.UserID,
		InterfaceID:   item.ID,
	}, nil
}

func (s *APIGatewayService) logGatewayRequest(target *GatewayTarget, requestMethod string, requestPath string, statusCode int, duration time.Duration, errorMessage string, insight GatewayRequestInsight) {
	if s.logRepo == nil || target == nil {
		return
	}

	_ = s.logRepo.Create(&entities.GatewayRequestLog{
		UserID:          target.UserID,
		UserInterfaceID: target.InterfaceID,
		InterfaceType:   target.InterfaceType,
		RequestMethod:   requestMethod,
		RequestPath:     requestPath,
		UpstreamURL:     target.TargetURL,
		Model:           firstNonEmpty(insight.Model, target.DefaultModel),
		StatusCode:      statusCode,
		DurationMS:      duration.Milliseconds(),
		ErrorMessage:    truncateText(errorMessage, 512),
	})
}

func (s *APIGatewayService) logChatRecord(
	target *GatewayTarget,
	requestMethod string,
	requestPath string,
	statusCode int,
	duration time.Duration,
	errorMessage string,
	requestInsight GatewayRequestInsight,
	responseInsight ChatResponseInsight,
) {
	if s.chatRepo == nil || target == nil {
		return
	}

	userInput := truncateText(normalizeWhitespace(strings.Join(requestInsight.Inputs, "\n")), 10000)
	modelOutput := truncateText(normalizeWhitespace(responseInsight.Output), 20000)
	if userInput == "" && modelOutput == "" && strings.TrimSpace(errorMessage) == "" {
		return
	}

	_ = s.chatRepo.Create(&entities.ChatRecord{
		UserID:          target.UserID,
		UserInterfaceID: target.InterfaceID,
		InterfaceType:   target.InterfaceType,
		RequestMethod:   requestMethod,
		RequestPath:     requestPath,
		Model:           firstNonEmpty(requestInsight.Model, responseInsight.Model, target.DefaultModel),
		UserInput:       userInput,
		ModelOutput:     modelOutput,
		StatusCode:      statusCode,
		DurationMS:      duration.Milliseconds(),
		ErrorMessage:    truncateText(errorMessage, 512),
	})
}

func inspectPreparedRequestBody(req *http.Request, interfaceType string, fallbackModel string) (GatewayRequestInsight, error) {
	if req.Body == nil {
		return GatewayRequestInsight{Model: fallbackModel}, nil
	}
	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		return GatewayRequestInsight{Model: fallbackModel}, err
	}
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(rawBody))
	return extractGatewayRequestInsight(rawBody, interfaceType, fallbackModel), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
