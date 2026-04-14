package apigatewaysvc

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend/internal/models/dto/requests"
	"backend/internal/models/dto/response"
	"backend/internal/models/entities"
	"backend/internal/pkg/config"
	"backend/internal/pkg/utils"
	apigatewayrepo "backend/internal/repository/api_gateway"
)

type APIGatewayService struct {
	cfg               *config.Config
	userInterfaceRepo *apigatewayrepo.UserInterfaceRepository
	logRepo           *apigatewayrepo.GatewayRequestLogRepository
	chatRepo          *apigatewayrepo.ChatRecordRepository
	httpClient        *http.Client
}

func NewAPIGatewayService(
	cfg *config.Config,
	userInterfaceRepo *apigatewayrepo.UserInterfaceRepository,
	logRepo *apigatewayrepo.GatewayRequestLogRepository,
	chatRepo *apigatewayrepo.ChatRecordRepository,
) *APIGatewayService {
	return &APIGatewayService{
		cfg:               cfg,
		userInterfaceRepo: userInterfaceRepo,
		logRepo:           logRepo,
		chatRepo:          chatRepo,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.MindGateUpstreamTimeoutSec) * time.Second,
		},
	}
}

func (s *APIGatewayService) ListInterfaces(userID uint) ([]response.UserInterfaceResp, error) {
	items, err := s.userInterfaceRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}

	resp := make([]response.UserInterfaceResp, 0, len(items))
	for _, item := range items {
		mapped, err := s.toUserInterfaceResp(item)
		if err != nil {
			return nil, err
		}
		resp = append(resp, mapped)
	}
	return resp, nil
}

func (s *APIGatewayService) CreateInterface(userID uint, req requests.CreateUserInterfaceReq) (*response.CreateUserInterfaceResp, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	interfaceType, err := normalizeInterfaceType(req.InterfaceType)
	if err != nil {
		return nil, err
	}
	targetBaseURL, err := normalizeTargetBaseURL(req.TargetBaseURL, interfaceType)
	if err != nil {
		return nil, err
	}
	targetAPIKey := strings.TrimSpace(req.TargetAPIKey)
	if targetAPIKey == "" {
		return nil, fmt.Errorf("target api key is required")
	}
	encryptedKey, err := utils.EncryptString(s.cfg.MindGateKeyEncryptionSecret, targetAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt target api key: %w", err)
	}
	gatewayKey, gatewayKeyHash, gatewayKeyPrefix, err := utils.GenerateGatewayKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate gateway key: %w", err)
	}
	gatewayKeyEncrypted, err := utils.EncryptString(s.cfg.MindGateKeyEncryptionSecret, gatewayKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt gateway key: %w", err)
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	item := &entities.UserInterface{
		UserID:                userID,
		Name:                  name,
		InterfaceType:         interfaceType,
		TargetBaseURL:         targetBaseURL,
		TargetAPIKeyEncrypted: encryptedKey,
		TargetAPIKeyMask:      utils.MaskSecret(targetAPIKey),
		DefaultModel:          strings.TrimSpace(req.DefaultModel),
		Enabled:               enabled,
		GatewayKeyEncrypted:   gatewayKeyEncrypted,
		GatewayKeyHash:        gatewayKeyHash,
		GatewayKeyPrefix:      gatewayKeyPrefix,
	}
	if err := s.userInterfaceRepo.Create(item); err != nil {
		return nil, err
	}

	interfaceResp, err := s.toUserInterfaceResp(*item)
	if err != nil {
		return nil, err
	}

	return &response.CreateUserInterfaceResp{
		Interface:  interfaceResp,
		GatewayKey: gatewayKey,
	}, nil
}

func (s *APIGatewayService) UpdateInterface(userID uint, interfaceID uint, req requests.UpdateUserInterfaceReq) (*response.UserInterfaceResp, error) {
	item, err := s.userInterfaceRepo.FindByIDForUser(interfaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("interface not found")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	interfaceType, err := normalizeInterfaceType(req.InterfaceType)
	if err != nil {
		return nil, err
	}
	targetBaseURL, err := normalizeTargetBaseURL(req.TargetBaseURL, interfaceType)
	if err != nil {
		return nil, err
	}

	item.Name = name
	item.InterfaceType = interfaceType
	item.TargetBaseURL = targetBaseURL
	item.DefaultModel = strings.TrimSpace(req.DefaultModel)
	if req.Enabled != nil {
		item.Enabled = *req.Enabled
	}

	if targetAPIKey := strings.TrimSpace(req.TargetAPIKey); targetAPIKey != "" {
		encryptedKey, err := utils.EncryptString(s.cfg.MindGateKeyEncryptionSecret, targetAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt target api key: %w", err)
		}
		item.TargetAPIKeyEncrypted = encryptedKey
		item.TargetAPIKeyMask = utils.MaskSecret(targetAPIKey)
	}

	if err := s.userInterfaceRepo.Update(item); err != nil {
		return nil, err
	}

	resp, err := s.toUserInterfaceResp(*item)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *APIGatewayService) DeleteInterface(userID uint, interfaceID uint) error {
	if _, err := s.userInterfaceRepo.FindByIDForUser(interfaceID, userID); err != nil {
		return fmt.Errorf("interface not found")
	}
	return s.userInterfaceRepo.Delete(interfaceID, userID)
}

func (s *APIGatewayService) RegenerateInterfaceGatewayKey(userID uint, interfaceID uint) (*response.RegenerateUserInterfaceGatewayKeyResp, error) {
	item, err := s.userInterfaceRepo.FindByIDForUser(interfaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("interface not found")
	}

	gatewayKey, gatewayKeyHash, gatewayKeyPrefix, err := utils.GenerateGatewayKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate gateway key: %w", err)
	}
	gatewayKeyEncrypted, err := utils.EncryptString(s.cfg.MindGateKeyEncryptionSecret, gatewayKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt gateway key: %w", err)
	}

	item.GatewayKeyEncrypted = gatewayKeyEncrypted
	item.GatewayKeyHash = gatewayKeyHash
	item.GatewayKeyPrefix = gatewayKeyPrefix
	if err := s.userInterfaceRepo.Update(item); err != nil {
		return nil, err
	}

	resp, err := s.toUserInterfaceResp(*item)
	if err != nil {
		return nil, err
	}

	return &response.RegenerateUserInterfaceGatewayKeyResp{
		Interface:  resp,
		GatewayKey: gatewayKey,
	}, nil
}

func (s *APIGatewayService) toUserInterfaceResp(item entities.UserInterface) (response.UserInterfaceResp, error) {
	gatewayKey := ""
	if strings.TrimSpace(item.GatewayKeyEncrypted) != "" {
		decrypted, err := utils.DecryptString(s.cfg.MindGateKeyEncryptionSecret, item.GatewayKeyEncrypted)
		if err != nil {
			return response.UserInterfaceResp{}, fmt.Errorf("failed to decrypt gateway key: %w", err)
		}
		gatewayKey = decrypted
	}

	return response.UserInterfaceResp{
		ID:               item.ID,
		Name:             item.Name,
		InterfaceType:    item.InterfaceType,
		TargetBaseURL:    item.TargetBaseURL,
		TargetAPIKeyMask: item.TargetAPIKeyMask,
		DefaultModel:     item.DefaultModel,
		Enabled:          item.Enabled,
		GatewayKey:       gatewayKey,
		GatewayKeyPrefix: item.GatewayKeyPrefix,
		LastUsedAt:       item.LastUsedAt,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}, nil
}

func (s *APIGatewayService) ListGatewayRequestLogs(userID uint, req requests.ListGatewayRequestLogsReq) (*response.GatewayRequestLogListResp, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	items, total, err := s.logRepo.ListByUserID(userID, req.InterfaceID, req.Keyword, req.StatusCode, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]response.GatewayRequestLogResp, 0, len(items))
	for _, item := range items {
		list = append(list, response.GatewayRequestLogResp{
			ID:              item.ID,
			UserID:          item.UserID,
			UserInterfaceID: item.UserInterfaceID,
			InterfaceType:   item.InterfaceType,
			RequestMethod:   item.RequestMethod,
			RequestPath:     item.RequestPath,
			UpstreamURL:     item.UpstreamURL,
			Model:           item.Model,
			StatusCode:      item.StatusCode,
			DurationMS:      item.DurationMS,
			ErrorMessage:    item.ErrorMessage,
			CreatedAt:       item.CreatedAt.Format(time.RFC3339),
		})
	}

	return &response.GatewayRequestLogListResp{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *APIGatewayService) ListChatRecords(userID uint, req requests.ListChatRecordsReq) (*response.ChatRecordListResp, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	items, total, err := s.chatRepo.ListByUserID(userID, req.InterfaceID, req.Keyword, req.StatusCode, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]response.ChatRecordResp, 0, len(items))
	for _, item := range items {
		list = append(list, response.ChatRecordResp{
			ID:              item.ID,
			UserID:          item.UserID,
			UserInterfaceID: item.UserInterfaceID,
			InterfaceType:   item.InterfaceType,
			RequestMethod:   item.RequestMethod,
			RequestPath:     item.RequestPath,
			Model:           item.Model,
			UserInput:       item.UserInput,
			ModelOutput:     item.ModelOutput,
			StatusCode:      item.StatusCode,
			DurationMS:      item.DurationMS,
			ErrorMessage:    item.ErrorMessage,
			CreatedAt:       item.CreatedAt.Format(time.RFC3339),
		})
	}

	return &response.ChatRecordListResp{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *APIGatewayService) GetGatewayDisplayConfig() *response.GatewayDisplayConfigResp {
	return &response.GatewayDisplayConfigResp{
		GatewayBaseURL: strings.TrimRight(strings.TrimSpace(s.cfg.PublicBackendBaseURL), "/"),
	}
}
