package usersvc

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"backend/internal/models/dto/requests"
	"backend/internal/models/dto/response"
	"backend/internal/models/mapper"
	"backend/internal/pkg/config"
	"backend/internal/pkg/utils"
	userrepo "backend/internal/repository/user"
	accesssvc "backend/internal/service/access"
	authsvc "backend/internal/service/auth"
	filesvc "backend/internal/service/file"

	"gorm.io/gorm"
)

type UserService struct {
	cfg            *config.Config
	userRepo       *userrepo.UserRepository
	smsService     *authsvc.SMSService
	casbin         *accesssvc.CasbinService
	fileService    *filesvc.FileService
}

func NewUserService(
	cfg *config.Config,
	userRepo *userrepo.UserRepository,
	smsService *authsvc.SMSService,
	casbin *accesssvc.CasbinService,
	fileService *filesvc.FileService,
) *UserService {
	return &UserService{
		cfg:            cfg,
		userRepo:       userRepo,
		smsService:     smsService,
		casbin:         casbin,
		fileService:    fileService,
	}
}

func (s *UserService) GetProfile(userID uint) (*response.UserResp, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	resp := mapper.ToUserResp(*user)
	resp.AvatarURL = mapper.ResolveStoredFileURL(resp.AvatarURL, s.fileService.BuildDownloadURL)
	resp.OperationIDs = accesssvc.BuildUserOperationIDs(user.Roles, s.casbin)
	return &resp, nil
}

func (s *UserService) UpdateProfile(userID uint, req requests.UpdateProfileReq) (*response.UserResp, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	email := strings.TrimSpace(req.Email)
	if email != "" {
		existingByEmail, err := s.userRepo.FindByEmail(email)
		if err == nil && existingByEmail.ID != userID {
			return nil, fmt.Errorf("email already exists")
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	user.Email = email
	if avatarFileID := strings.TrimSpace(req.AvatarFileID); avatarFileID != "" {
		if s.fileService == nil {
			return nil, fmt.Errorf("file service unavailable")
		}
		if _, err := s.fileService.RequireFileOwnedBy(avatarFileID, userID); err != nil {
			return nil, err
		}
		if err := s.fileService.BindFile(avatarFileID); err != nil {
			return nil, err
		}
		user.AvatarURL = avatarFileID
	}
	user.Signature = strings.TrimSpace(req.Signature)
	user.Gender = strings.TrimSpace(req.Gender)
	user.Age = req.Age

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	resp := mapper.ToUserResp(*user)
	resp.AvatarURL = mapper.ResolveStoredFileURL(resp.AvatarURL, s.fileService.BuildDownloadURL)
	resp.OperationIDs = accesssvc.BuildUserOperationIDs(user.Roles, s.casbin)
	return &resp, nil
}

func (s *UserService) ResetPassword(userID uint, req requests.ResetPasswordReq) error {
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if !utils.ComparePassword(user.PasswordHash, req.OldPassword) {
		return fmt.Errorf("old password mismatch")
	}
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(userID, newHash)
}

func (s *UserService) ChangePhone(ctx context.Context, userID uint, req requests.ChangePhoneReq) error {
	newPhone := strings.TrimSpace(req.NewPhone)
	oldPhoneCode := strings.TrimSpace(req.OldPhoneCode)
	newPhoneCode := strings.TrimSpace(req.NewPhoneCode)

	if err := utils.ValidatePhone(newPhone); err != nil {
		return err
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if s.cfg.SMSVerifyEnabled {
		if oldPhoneCode == "" {
			return fmt.Errorf("old phone code is required")
		}
		if newPhoneCode == "" {
			return fmt.Errorf("new phone code is required")
		}
		if err := s.smsService.VerifyCode(ctx, user.Phone, "change_phone_old", oldPhoneCode); err != nil {
			return fmt.Errorf("old phone verification failed: %w", err)
		}
		if err := s.smsService.VerifyCode(ctx, newPhone, "change_phone_new", newPhoneCode); err != nil {
			return fmt.Errorf("new phone verification failed: %w", err)
		}
	}

	existingByPhone, err := s.userRepo.FindByPhone(newPhone)
	if err == nil && existingByPhone.ID != userID {
		return fmt.Errorf("phone already exists")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	user.Phone = newPhone
	return s.userRepo.Update(user)
}

func (s *UserService) UploadAvatar(userID uint, fileHeader *multipart.FileHeader) (string, error) {
	if s.fileService == nil {
		return "", fmt.Errorf("file service unavailable")
	}

	item, err := s.fileService.UploadMultipartFile(fileHeader, filesvc.UploadFileOptions{
		UserID:   userID,
	})
	if err != nil {
		return "", err
	}
	if err := s.fileService.BindFile(item.ID); err != nil {
		return "", err
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}
	user.AvatarURL = item.ID
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}
	return mapper.ResolveStoredFileURL(item.ID, s.fileService.BuildDownloadURL), nil
}
