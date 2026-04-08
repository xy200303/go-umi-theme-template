package adminsvc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	policytemplate "backend/generate"
	"backend/internal/models/dto/requests"
	"backend/internal/models/dto/response"
	"backend/internal/models/entities"
	"backend/internal/models/mapper"
	"backend/internal/pkg/utils"
	filerepo "backend/internal/repository/file"
	rolerepo "backend/internal/repository/role"
	systemrepo "backend/internal/repository/system"
	userrepo "backend/internal/repository/user"
	accesssvc "backend/internal/service/access"
	filesvc "backend/internal/service/file"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AdminService struct {
	userRepo    *userrepo.UserRepository
	fileRepo    *filerepo.FileRepository
	roleRepo    *rolerepo.RoleRepository
	cfgRepo     *systemrepo.SystemConfigRepository
	auditRepo   *systemrepo.AuditLogRepository
	casbin      *accesssvc.CasbinService
	redisClient *redis.Client
	fileService *filesvc.FileService
}

const reservedAdminUsername = "admin"
const reservedAdminRoleName = "admin"
const defaultUserRoleName = "user"
const auditMaxRecordsConfigKey = "audit.max_records"
const auditMaxRecordsDefault = 10000

func NewAdminService(
	userRepo *userrepo.UserRepository,
	fileRepo *filerepo.FileRepository,
	roleRepo *rolerepo.RoleRepository,
	cfgRepo *systemrepo.SystemConfigRepository,
	auditRepo *systemrepo.AuditLogRepository,
	casbin *accesssvc.CasbinService,
	redisClient *redis.Client,
	fileService *filesvc.FileService,
) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		fileRepo:    fileRepo,
		roleRepo:    roleRepo,
		cfgRepo:     cfgRepo,
		auditRepo:   auditRepo,
		casbin:      casbin,
		redisClient: redisClient,
		fileService: fileService,
	}
}

func (s *AdminService) Stats(ctx context.Context) (*response.SystemStatsResp, error) {
	userCount, err := s.userRepo.Count()
	if err != nil {
		return nil, err
	}
	roleCount, err := s.roleRepo.Count()
	if err != nil {
		return nil, err
	}
	configCount, err := s.cfgRepo.Count()
	if err != nil {
		return nil, err
	}
	redisOnline := s.redisClient.Ping(ctx).Err() == nil

	return &response.SystemStatsResp{
		UserCount:         userCount,
		RoleCount:         roleCount,
		SystemConfigCount: configCount,
		RedisOnline:       redisOnline,
	}, nil
}

func (s *AdminService) ListPolicyTemplates() []response.PolicyTemplateResp {
	return policytemplate.List()
}

func (s *AdminService) ListUsers(keyword string) ([]response.UserResp, error) {
	users, err := s.userRepo.ListUsers(keyword)
	if err != nil {
		return nil, err
	}
	resp := make([]response.UserResp, 0, len(users))
	for _, u := range users {
		item := mapper.ToUserResp(u)
		item.AvatarURL = mapper.ResolveStoredFileURL(item.AvatarURL, s.fileService.BuildDownloadURL)
		item.OperationIDs = accesssvc.BuildUserOperationIDs(u.Roles, s.casbin)
		resp = append(resp, item)
	}
	return resp, nil
}

func (s *AdminService) CreateUser(req requests.CreateUserReq) (*response.UserResp, error) {
	if err := utils.ValidateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := utils.ValidatePhone(req.Phone); err != nil {
		return nil, err
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	username := strings.TrimSpace(req.Username)
	phone := strings.TrimSpace(req.Phone)
	email := strings.TrimSpace(req.Email)

	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, fmt.Errorf("username already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if _, err := s.userRepo.FindByPhone(phone); err == nil {
		return nil, fmt.Errorf("phone already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if email != "" {
		if _, err := s.userRepo.FindByEmail(email); err == nil {
			return nil, fmt.Errorf("email already exists")
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		Username:     username,
		Phone:        phone,
		PasswordHash: passwordHash,
		Email:        email,
		Signature:    strings.TrimSpace(req.Signature),
		Gender:       strings.TrimSpace(req.Gender),
		Age:          req.Age,
		IsActive:     req.IsActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	if avatarFileID := strings.TrimSpace(req.AvatarFileID); avatarFileID != "" {
		if err := s.bindUserAvatar(user, avatarFileID); err != nil {
			return nil, err
		}
	}

	roleNames := req.RoleNames
	if len(roleNames) == 0 {
		roleNames = []string{defaultUserRoleName}
	}

	if err := s.syncUserRoles(user.ID, roleNames); err != nil {
		return nil, err
	}

	freshUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return nil, err
	}
	resp := mapper.ToUserResp(*freshUser)
	resp.AvatarURL = mapper.ResolveStoredFileURL(resp.AvatarURL, s.fileService.BuildDownloadURL)
	resp.OperationIDs = accesssvc.BuildUserOperationIDs(freshUser.Roles, s.casbin)
	return &resp, nil
}

func (s *AdminService) UpdateUser(userID uint, actorUserID uint, req requests.UpdateUserReq) (*response.UserResp, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if err := utils.ValidateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := utils.ValidatePhone(req.Phone); err != nil {
		return nil, err
	}

	username := strings.TrimSpace(req.Username)
	phone := strings.TrimSpace(req.Phone)
	email := strings.TrimSpace(req.Email)

	if strings.EqualFold(user.Username, reservedAdminUsername) && !strings.EqualFold(username, reservedAdminUsername) {
		return nil, fmt.Errorf("reserved username admin cannot be changed")
	}
	if userID == actorUserID && !req.IsActive {
		return nil, fmt.Errorf("cannot deactivate current user")
	}

	existingByUsername, err := s.userRepo.FindByUsername(username)
	if err == nil && existingByUsername.ID != userID {
		return nil, fmt.Errorf("username already exists")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	existingByPhone, err := s.userRepo.FindByPhone(phone)
	if err == nil && existingByPhone.ID != userID {
		return nil, fmt.Errorf("phone already exists")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if email != "" {
		existingByEmail, err := s.userRepo.FindByEmail(email)
		if err == nil && existingByEmail.ID != userID {
			return nil, fmt.Errorf("email already exists")
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	user.Username = username
	user.Phone = phone
	user.Email = email
	user.Signature = strings.TrimSpace(req.Signature)
	user.Gender = strings.TrimSpace(req.Gender)
	user.Age = req.Age
	user.IsActive = req.IsActive

	if avatarFileID := strings.TrimSpace(req.AvatarFileID); avatarFileID != "" {
		if _, err := s.fileService.RequireFile(avatarFileID); err != nil {
			return nil, err
		}
		user.AvatarURL = avatarFileID
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	if avatarFileID := strings.TrimSpace(req.AvatarFileID); avatarFileID != "" {
		if err := s.fileService.BindFile(avatarFileID); err != nil {
			return nil, err
		}
	}

	freshUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return nil, err
	}
	resp := mapper.ToUserResp(*freshUser)
	resp.AvatarURL = mapper.ResolveStoredFileURL(resp.AvatarURL, s.fileService.BuildDownloadURL)
	resp.OperationIDs = accesssvc.BuildUserOperationIDs(freshUser.Roles, s.casbin)
	return &resp, nil
}

func (s *AdminService) bindUserAvatar(user *entities.User, avatarFileID string) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	if strings.TrimSpace(avatarFileID) == "" {
		return nil
	}
	if s.fileService == nil {
		return fmt.Errorf("file service unavailable")
	}
	if _, err := s.fileService.RequireFile(avatarFileID); err != nil {
		return err
	}
	if err := s.fileService.BindFile(avatarFileID); err != nil {
		return err
	}
	user.AvatarURL = avatarFileID
	return s.userRepo.Update(user)
}

func (s *AdminService) ResetUserPassword(userID uint, req requests.ResetUserPasswordReq) error {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		return err
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return err
	}
	passwordHash, err := utils.HashPassword(strings.TrimSpace(req.Password))
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(userID, passwordHash)
}

func (s *AdminService) DeleteUser(userID uint, actorUserID uint) error {
	if userID == actorUserID {
		return fmt.Errorf("cannot delete current user")
	}

	targetUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	actorUser, err := s.userRepo.FindByID(actorUserID)
	if err != nil {
		return err
	}

	targetIsAdmin := false
	for _, role := range targetUser.Roles {
		if strings.EqualFold(role.Name, reservedAdminRoleName) {
			targetIsAdmin = true
			break
		}
	}

	if targetIsAdmin && !strings.EqualFold(actorUser.Username, reservedAdminUsername) {
		return fmt.Errorf("only super admin can delete admin users")
	}

	return s.userRepo.Delete(userID)
}

func (s *AdminService) UpdateUserRoles(userID uint, req requests.UpdateUserRolesReq) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if strings.EqualFold(user.Username, reservedAdminUsername) {
		return fmt.Errorf("reserved admin user roles cannot be changed")
	}
	if err := s.syncUserRoles(userID, req.RoleNames); err != nil {
		return err
	}
	return nil
}

func (s *AdminService) syncUserRoles(userID uint, roleNames []string) error {
	if len(roleNames) == 0 {
		return s.userRepo.SetRoles(userID, []entities.Role{})
	}

	roles, err := s.roleRepo.FindByNames(roleNames)
	if err != nil {
		return err
	}
	if len(roles) != len(roleNames) {
		return fmt.Errorf("some roles are undefined")
	}
	return s.userRepo.SetRoles(userID, roles)
}

func (s *AdminService) ListRoles() ([]entities.Role, error) {
	return s.roleRepo.List()
}

func (s *AdminService) CreateRole(req requests.CreateRoleReq) (*entities.Role, error) {
	role := &entities.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
	}
	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *AdminService) UpdateRole(roleID uint, req requests.UpdateRoleReq) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if strings.EqualFold(role.Name, reservedAdminRoleName) {
		return fmt.Errorf("reserved admin role cannot be modified")
	}
	role.DisplayName = req.DisplayName
	role.Description = req.Description
	return s.roleRepo.Update(role)
}

func (s *AdminService) DeleteRole(roleID uint) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if strings.EqualFold(role.Name, reservedAdminRoleName) {
		return fmt.Errorf("reserved admin role cannot be deleted")
	}
	return s.roleRepo.Delete(roleID)
}

func (s *AdminService) SetRolePolicies(roleID uint, req requests.SetRolePoliciesReq) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if strings.EqualFold(role.Name, reservedAdminRoleName) {
		return fmt.Errorf("reserved admin role cannot be modified")
	}
	policies := make([]accesssvc.Policy, 0, len(req.Policies))
	for _, p := range req.Policies {
		policies = append(policies, accesssvc.Policy{Path: p.Path, Method: p.Method})
	}
	return s.casbin.SetRolePolicies(role.Name, policies)
}

func (s *AdminService) GetRolePolicies(roleID uint) ([]accesssvc.Policy, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, err
	}
	return s.casbin.GetRolePolicies(role.Name), nil
}

func (s *AdminService) ListSystemConfigs() ([]entities.SystemConfig, error) {
	return s.cfgRepo.List()
}

func (s *AdminService) ListAdminFiles(req requests.ListAdminFilesReq) (*response.AdminFileListResp, error) {
	items, total, page, pageSize, err := s.fileRepo.ListAdminPage(filerepo.AdminFileFilter{
		Keyword:      req.Keyword,
		UploadStatus: req.UploadStatus,
	}, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	respList := make([]response.AdminFileResp, 0, len(items))
	for _, item := range items {
		respList = append(respList, response.AdminFileResp{
			ID:            item.ID,
			StorageDriver: item.StorageDriver,
			StoragePath:   item.StoragePath,
			OriginalName:  item.OriginalName,
			Ext:           item.Ext,
			MimeType:      item.MimeType,
			Size:          item.Size,
			UploadStatus:  item.UploadStatus,
			UploadedBy:    item.UploadedBy,
			Remark:        item.Remark,
			FileURL:       s.fileService.BuildFileURL(&item),
			CreatedAt:     item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     item.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &response.AdminFileListResp{
		List:     respList,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AdminService) GetAdminFileStats(req requests.ListAdminFilesReq) (*response.AdminFileStatsResp, error) {
	stats, err := s.fileRepo.GetAdminStats(filerepo.AdminFileFilter{
		Keyword:      req.Keyword,
		UploadStatus: req.UploadStatus,
	})
	if err != nil {
		return nil, err
	}

	return &response.AdminFileStatsResp{
		TotalCount:    stats.TotalCount,
		UploadedCount: stats.UploadedCount,
		BoundCount:    stats.BoundCount,
		DeletedCount:  stats.DeletedCount,
	}, nil
}

func (s *AdminService) UpsertSystemConfig(req requests.SystemConfigReq) error {
	if strings.TrimSpace(req.ConfigKey) == auditMaxRecordsConfigKey {
		value, err := strconv.Atoi(strings.TrimSpace(req.ConfigVal))
		if err != nil || value <= 0 {
			return fmt.Errorf("audit.max_records must be a positive integer")
		}
	}

	item := &entities.SystemConfig{
		ConfigGroup: req.ConfigGroup,
		ConfigKey:   req.ConfigKey,
		ConfigVal:   req.ConfigVal,
		Remark:      req.Remark,
	}
	return s.cfgRepo.Upsert(item)
}

func (s *AdminService) CreateAuditLog(item *entities.AuditLog) error {
	if item == nil {
		return fmt.Errorf("audit log is nil")
	}
	if err := s.auditRepo.Create(item); err != nil {
		return err
	}

	limit := auditMaxRecordsDefault
	configItem, err := s.cfgRepo.GetByKey(auditMaxRecordsConfigKey)
	if err == nil {
		if parsed, parseErr := strconv.Atoi(strings.TrimSpace(configItem.ConfigVal)); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	return s.auditRepo.PruneExcess(limit)
}

func (s *AdminService) ListAuditLogs(req requests.ListAuditLogsReq) (*response.AuditLogListResp, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	switch {
	case pageSize <= 0:
		pageSize = 20
	case pageSize > 100:
		pageSize = 100
	}

	items, total, err := s.auditRepo.List(req.Keyword, req.MenuKey, req.StatusCode, page, pageSize)
	if err != nil {
		return nil, err
	}

	respList := make([]response.AuditLogResp, 0, len(items))
	for _, item := range items {
		respList = append(respList, response.AuditLogResp{
			ID:            item.ID,
			UserID:        item.UserID,
			Username:      item.Username,
			Method:        item.Method,
			RoutePath:     item.RoutePath,
			RequestPath:   item.RequestPath,
			OperationID:   item.OperationID,
			OperationName: item.OperationName,
			MenuKey:       item.MenuKey,
			MenuLabel:     item.MenuLabel,
			StatusCode:    item.StatusCode,
			ClientIP:      item.ClientIP,
			UserAgent:     item.UserAgent,
			DurationMS:    item.DurationMS,
			CreatedAt:     item.CreatedAt.Format(time.RFC3339),
		})
	}

	return &response.AuditLogListResp{
		List:     respList,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
