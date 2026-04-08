package fileservice

import (
	"backend/internal/models/entities"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"backend/internal/models/dto/requests"
	"backend/internal/models/dto/response"
	"backend/internal/pkg/config"
	"backend/internal/pkg/utils"
	filerepo "backend/internal/repository/file"
	systemrepo "backend/internal/repository/system"

	"gorm.io/gorm"
)

type UploadFileOptions struct {
	UserID   uint
}

const (
	fileUploadModeDirect = "direct"
	fileUploadModeProxy  = "proxy"
)

var fileIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

type FileService struct {
	fileRepo       *filerepo.FileRepository
	storageService *StorageService
	cfgRepo        *systemrepo.SystemConfigRepository
	cfg            *config.Config
}

func NewFileService(cfg *config.Config, fileRepo *filerepo.FileRepository, storageService *StorageService, cfgRepo *systemrepo.SystemConfigRepository) *FileService {
	return &FileService{
		fileRepo:       fileRepo,
		storageService: storageService,
		cfgRepo:        cfgRepo,
		cfg:            cfg,
	}
}

func (s *FileService) UploadMultipartFile(fileHeader *multipart.FileHeader, options UploadFileOptions) (*entities.File, error) {
	if fileHeader == nil {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "请选择要上传的文件")
	}

	uploadOptions, err := s.resolveUploadOptions(fileHeader.Filename, fileHeader.Size)
	if err != nil {
		return nil, err
	}

	fileID, err := calculateMultipartMD5(fileHeader)
	if err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "计算文件摘要失败", err)
	}

	if existing, err := s.fileRepo.FindByID(fileID); err == nil {
		if existing.UploadStatus == fileStatusDeleted {
			return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "该文件已失效，请重新上传")
		}
		if err := s.syncExistingFileMetadata(existing, strings.TrimSpace(fileHeader.Filename), normalizeFileMimeType(fileHeader), fileHeader.Size); err != nil {
			return nil, err
		}
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	uploadOptions.ObjectName = fileID + strings.ToLower(filepath.Ext(fileHeader.Filename))
	storedFile, err := s.storageService.StoreWithOptions(fileHeader, *uploadOptions)
	if err != nil {
		return nil, err
	}

	item := &entities.File{
		ID:            fileID,
		StorageDriver: storedFile.StorageDriver,
		StoragePath:   storedFile.StoragePath,
		OriginalName:  strings.TrimSpace(fileHeader.Filename),
		Ext:           strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), "."),
		MimeType:      normalizeFileMimeType(fileHeader),
		Size:          fileHeader.Size,
		UploadStatus:  fileStatusUploaded,
		UploadedBy:    options.UserID,
	}

	if err := s.fileRepo.Create(item); err != nil {
		if existing, findErr := s.fileRepo.FindByID(fileID); findErr == nil {
			return existing, nil
		}
		return nil, err
	}

	return item, nil
}

func (s *FileService) InitDirectUpload(req requests.InitDirectUploadReq, _ uint) (*response.FileDirectUploadInitResp, error) {
	fileID := normalizeFileID(req.FileMD5)
	if !isValidFileID(fileID) {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "文件 MD5 不正确")
	}

	uploadOptions, err := s.resolveUploadOptions(req.FileName, req.FileSize)
	if err != nil {
		return nil, err
	}

	if existing, err := s.fileRepo.FindByID(fileID); err == nil {
		if existing.UploadStatus == fileStatusDeleted {
			return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "该文件已失效，请重新上传")
		}
		if err := s.syncExistingFileMetadata(existing, strings.TrimSpace(req.FileName), strings.TrimSpace(req.MimeType), req.FileSize); err != nil {
			return nil, err
		}
		existingResp := s.BuildUploadResp(existing)
		return &response.FileDirectUploadInitResp{
			Mode:          fileUploadModeDirect,
			AlreadyExists: true,
			FileID:        existing.ID,
			StorageDriver: existing.StorageDriver,
			File:          &existingResp,
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if !s.storageService.SupportsDirectUpload() {
		return &response.FileDirectUploadInitResp{
			Mode: fileUploadModeProxy,
		}, nil
	}

	uploadOptions.ObjectName = buildFileObjectName(fileID, req.FileName)
	keyPrefix := "uploads"
	if strings.TrimSpace(uploadOptions.SubDir) != "" {
		keyPrefix += "/" + strings.Trim(strings.ReplaceAll(uploadOptions.SubDir, "\\", "/"), "/")
	}
	storagePath := s.storageService.BuildStorageObjectPath(keyPrefix, req.FileName, *uploadOptions)
	ticket, err := s.storageService.InitDirectUpload(storagePath, req.MimeType)
	if err != nil {
		return nil, err
	}

	return &response.FileDirectUploadInitResp{
		Mode:          fileUploadModeDirect,
		FileID:        fileID,
		StorageDriver: ticket.StorageDriver,
		UploadMethod:  ticket.Method,
		UploadURL:     ticket.UploadURL,
		UploadHeaders: ticket.Headers,
		ExpiresAt:     formatFileTime(ticket.ExpiresAt),
	}, nil
}

func (s *FileService) CompleteDirectUpload(req requests.CompleteDirectUploadReq, userID uint) (*entities.File, error) {
	fileID := normalizeFileID(req.FileID)
	if !isValidFileID(fileID) {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "文件 ID 不正确")
	}

	uploadOptions, err := s.resolveUploadOptions(req.FileName, req.FileSize)
	if err != nil {
		return nil, err
	}

	if existing, err := s.fileRepo.FindByID(fileID); err == nil {
		if existing.UploadStatus == fileStatusDeleted {
			return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "该文件已失效，请重新上传")
		}
		if err := s.syncExistingFileMetadata(existing, strings.TrimSpace(req.FileName), strings.TrimSpace(req.MimeType), req.FileSize); err != nil {
			return nil, err
		}
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if !s.storageService.SupportsDirectUpload() {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "当前存储驱动不支持直传完成")
	}

	uploadOptions.ObjectName = buildFileObjectName(fileID, req.FileName)
	keyPrefix := "uploads"
	if strings.TrimSpace(uploadOptions.SubDir) != "" {
		keyPrefix += "/" + strings.Trim(strings.ReplaceAll(uploadOptions.SubDir, "\\", "/"), "/")
	}
	storagePath := s.storageService.BuildStorageObjectPath(keyPrefix, req.FileName, *uploadOptions)
	exists, err := s.storageService.ObjectExists("cos", storagePath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "文件尚未上传完成，请稍后重试")
	}

	item := &entities.File{
		ID:            fileID,
		StorageDriver: "cos",
		StoragePath:   storagePath,
		OriginalName:  strings.TrimSpace(req.FileName),
		Ext:           strings.TrimPrefix(strings.ToLower(filepath.Ext(req.FileName)), "."),
		MimeType:      strings.TrimSpace(req.MimeType),
		Size:          req.FileSize,
		UploadStatus:  fileStatusUploaded,
		UploadedBy:    userID,
	}

	if err := s.fileRepo.Create(item); err != nil {
		if existing, findErr := s.fileRepo.FindByID(fileID); findErr == nil {
			return existing, nil
		}
		return nil, err
	}

	return item, nil
}

func (s *FileService) GetByID(fileID string) (*entities.File, error) {
	item, err := s.fileRepo.FindByID(strings.ToLower(strings.TrimSpace(fileID)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(404, utils.ErrCodeNotFound, "文件不存在")
		}
		return nil, err
	}
	if item.UploadStatus == fileStatusDeleted {
		return nil, utils.NewAppError(404, utils.ErrCodeNotFound, "文件不存在")
	}
	return item, nil
}

func (s *FileService) RequireFile(fileID string) (*entities.File, error) {
	item, err := s.GetByID(fileID)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *FileService) RequireFileOwnedBy(fileID string, userID uint) (*entities.File, error) {
	item, err := s.RequireFile(fileID)
	if err != nil {
		return nil, err
	}
	if item.UploadedBy != userID {
		return nil, utils.NewAppError(403, utils.ErrCodeForbidden, "无权使用该文件")
	}
	return item, nil
}

func (s *FileService) BindFile(fileID string) error {
	item, err := s.GetByID(fileID)
	if err != nil {
		return err
	}
	if item.UploadStatus == fileStatusBound {
		return nil
	}
	item.UploadStatus = fileStatusBound
	return s.fileRepo.Update(item)
}

func (s *FileService) BindFileTx(tx *gorm.DB, fileID string) error {
	item, err := s.getByIDTx(tx, fileID)
	if err != nil {
		return err
	}
	if item.UploadStatus == fileStatusBound {
		return nil
	}
	item.UploadStatus = fileStatusBound
	return s.fileRepo.UpdateTx(tx, item)
}

func (s *FileService) BuildFileURL(item *entities.File) string {
	if item == nil {
		return ""
	}
	return s.BuildDownloadURL(item.ID)
}

func (s *FileService) BuildDownloadURL(fileID string) string {
	normalizedID := normalizeFileID(fileID)
	if !isValidFileID(normalizedID) {
		return ""
	}

	expireAt := time.Now().Add(time.Duration(resolveFileAccessSignedURLExpireMinutes(s.cfgRepo)) * time.Minute).Unix()
	signature := s.signFileAccess(normalizedID, expireAt)
	if signature == "" {
		return ""
	}

	return fmt.Sprintf("/api/v1/files/%s/download?expires=%d&sig=%s", normalizedID, expireAt, signature)
}

func (s *FileService) ValidateDownloadSignature(fileID, expireAtRaw, signature string) error {
	normalizedID := normalizeFileID(fileID)
	if !isValidFileID(normalizedID) {
		return utils.NewAppError(400, utils.ErrCodeInvalidRequest, "文件 ID 不正确")
	}

	expireAt, err := strconv.ParseInt(strings.TrimSpace(expireAtRaw), 10, 64)
	if err != nil || expireAt <= 0 {
		return utils.NewAppError(400, utils.ErrCodeInvalidRequest, "访问链接已失效")
	}
	if time.Now().Unix() > expireAt {
		return utils.NewAppError(403, utils.ErrCodeForbidden, "访问链接已过期")
	}

	expected := s.signFileAccess(normalizedID, expireAt)
	if expected == "" || !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature))) {
		return utils.NewAppError(403, utils.ErrCodeForbidden, "访问签名无效")
	}

	return nil
}

func (s *FileService) StorageService() *StorageService {
	if s == nil {
		return nil
	}
	return s.storageService
}

func (s *FileService) Transaction(fn func(tx *gorm.DB) error) error {
	if s == nil || s.fileRepo == nil {
		return fmt.Errorf("文件仓库未初始化")
	}
	return s.fileRepo.Transaction(fn)
}

func (s *FileService) signFileAccess(fileID string, expireAt int64) string {
	if s == nil || s.cfg == nil {
		return ""
	}

	secret := strings.TrimSpace(s.cfg.JWTAccessSecret)
	if secret == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(normalizeFileID(fileID)))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(strconv.FormatInt(expireAt, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *FileService) BuildUploadResp(item *entities.File) response.FileUploadResp {
	return response.FileUploadResp{
		ID:            item.ID,
		StorageDriver: item.StorageDriver,
		OriginalName:  item.OriginalName,
		Ext:           item.Ext,
		MimeType:      item.MimeType,
		Size:          item.Size,
		UploadStatus:  item.UploadStatus,
		FileURL:       s.BuildFileURL(item),
		CreatedAt:     formatFileTime(item.CreatedAt),
		UpdatedAt:     formatFileTime(item.UpdatedAt),
	}
}

func (s *FileService) syncExistingFileMetadata(item *entities.File, fileName string, mimeType string, size int64) error {
	if item == nil {
		return nil
	}

	normalizedName := strings.TrimSpace(fileName)
	normalizedMimeType := strings.TrimSpace(mimeType)
	normalizedExt := strings.TrimPrefix(strings.ToLower(filepath.Ext(normalizedName)), ".")
	changed := false

	if normalizedName != "" && strings.TrimSpace(item.OriginalName) != normalizedName {
		item.OriginalName = normalizedName
		changed = true
	}
	if normalizedExt != "" && strings.TrimSpace(item.Ext) != normalizedExt {
		item.Ext = normalizedExt
		changed = true
	}
	if normalizedMimeType != "" && strings.TrimSpace(item.MimeType) != normalizedMimeType {
		item.MimeType = normalizedMimeType
		changed = true
	}
	if size > 0 && item.Size != size {
		item.Size = size
		changed = true
	}
	if !changed {
		return nil
	}
	return s.fileRepo.Update(item)
}

func (s *FileService) ReadFileBytes(fileID string) (*entities.File, []byte, error) {
	item, err := s.GetByID(fileID)
	if err != nil {
		return nil, nil, err
	}
	if s.storageService == nil {
		return nil, nil, utils.NewAppError(500, utils.ErrCodeInternal, "文件存储服务未初始化")
	}

	data, err := s.storageService.ReadAll(item.StorageDriver, item.StoragePath)
	if err != nil {
		return nil, nil, err
	}
	return item, data, nil
}

func (s *FileService) CreateFileFromBytes(fileName string, data []byte, options UploadFileOptions) (*entities.File, error) {
	normalizedName := strings.TrimSpace(fileName)
	if normalizedName == "" {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "文件名不能为空")
	}
	if len(data) == 0 {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "文件内容不能为空")
	}

	uploadOptions, err := s.resolveUploadOptions(normalizedName, int64(len(data)))
	if err != nil {
		return nil, err
	}

	fileID := buildGeneratedFileID(options.UserID, normalizedName, data)
	if existing, err := s.fileRepo.FindByID(fileID); err == nil {
		if existing.UploadStatus == fileStatusDeleted {
			return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "该文件已失效，请重新生成")
		}
		if existing.UploadedBy == options.UserID {
			return existing, nil
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	uploadOptions.ObjectName = buildFileObjectName(fileID, normalizedName)
	storedFile, err := s.storageService.StoreBytesWithOptions(normalizedName, data, *uploadOptions)
	if err != nil {
		return nil, err
	}

	item := &entities.File{
		ID:            fileID,
		StorageDriver: storedFile.StorageDriver,
		StoragePath:   storedFile.StoragePath,
		OriginalName:  normalizedName,
		Ext:           strings.TrimPrefix(strings.ToLower(filepath.Ext(normalizedName)), "."),
		MimeType:      strings.TrimSpace(http.DetectContentType(data)),
		Size:          int64(len(data)),
		UploadStatus:  fileStatusUploaded,
		UploadedBy:    options.UserID,
	}

	if err := s.fileRepo.Create(item); err != nil {
		if existing, findErr := s.fileRepo.FindByID(fileID); findErr == nil && existing.UploadedBy == options.UserID {
			return existing, nil
		}
		return nil, err
	}

	return item, nil
}

func (s *FileService) CleanupExpiredUploadedFiles() (int, error) {
	ttlHours := resolveUnboundFileTTLHours(s.cfgRepo)
	batchSize := int(resolveUnboundFileCleanupBatchSize(s.cfgRepo))
	expireBefore := time.Now().Add(-time.Duration(ttlHours) * time.Hour)

	items, err := s.fileRepo.ListStaleUploadedBefore(expireBefore, batchSize)
	if err != nil {
		return 0, err
	}

	cleaned := 0
	for _, item := range items {
		if err := s.storageService.Delete(item.StorageDriver, item.StoragePath); err != nil {
			return cleaned, err
		}
		if err := s.fileRepo.Transaction(func(tx *gorm.DB) error {
			fresh, err := s.getByIDTx(tx, item.ID)
			if err != nil {
				return err
			}
			if fresh.UploadStatus != fileStatusUploaded {
				return nil
			}
			fresh.UploadStatus = fileStatusDeleted
			fresh.Remark = "auto cleaned: unbound file expired"
			return s.fileRepo.UpdateTx(tx, fresh)
		}); err != nil {
			return cleaned, err
		}
		cleaned++
	}

	return cleaned, nil
}

func (s *FileService) StartCleanupWorker(ctx context.Context) {
	go func() {
		for {
			waitMinutes := resolveUnboundFileCleanupIntervalMinutes(s.cfgRepo)
			waitDuration := time.Duration(waitMinutes) * time.Minute
			if waitDuration <= 0 {
				waitDuration = time.Duration(defaultUnboundFileCleanupIntervalMin) * time.Minute
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(waitDuration):
			}

			cleaned, err := s.CleanupExpiredUploadedFiles()
			if err != nil {
				log.Printf("[file-cleanup] cleanup failed: %v", err)
				continue
			}
			if cleaned > 0 {
				log.Printf("[file-cleanup] cleaned %d expired uploaded file(s)", cleaned)
			}
		}
	}()
}

func (s *FileService) resolveUploadOptions(fileName string, fileSize int64) (*UploadOptions, error) {
	maxSizeBytes := int64(10 * 1024 * 1024)
	if s.cfg != nil && s.cfg.UploadMaxSizeMB > 0 {
		maxSizeBytes = s.cfg.UploadMaxSizeMB * 1024 * 1024
	}

	options := &UploadOptions{
		AllowedSuffixes:  nil,
		SubDir:           buildFileStorageSubDir("files", fileName),
		MaxSizeBytes:     maxSizeBytes,
		SizeLimitMessage: buildGenericSizeLimitMessage(maxSizeBytes / 1024 / 1024),
	}
	if s.cfg != nil && len(s.cfg.AllowedUploadSuffix) > 0 {
		options.AllowedSuffixes = append([]string{}, s.cfg.AllowedUploadSuffix...)
	}

	if err := s.storageService.ValidateMeta(fileName, fileSize, *options); err != nil {
		return nil, err
	}
	return options, nil
}

func buildFileStorageSubDir(prefix, fileName string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	if ext == "" {
		return prefix
	}
	return filepath.ToSlash(filepath.Join(prefix, ext))
}

func buildFileObjectName(fileID, fileName string) string {
	return normalizeFileID(fileID) + strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
}

func buildGeneratedFileID(userID uint, fileName string, data []byte) string {
	hasher := md5.New()
	_, _ = hasher.Write([]byte(strconv.FormatUint(uint64(userID), 10)))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write([]byte(strings.TrimSpace(strings.ToLower(filepath.Ext(fileName)))))
	_, _ = hasher.Write([]byte(":"))
	_, _ = hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

func calculateMultipartMD5(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func normalizeFileMimeType(fileHeader *multipart.FileHeader) string {
	if fileHeader == nil {
		return ""
	}
	return strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
}

func formatFileTime(value time.Time) string {
	return value.Format(time.RFC3339)
}

func normalizeFileID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isValidFileID(value string) bool {
	return fileIDPattern.MatchString(normalizeFileID(value))
}

func (s *FileService) getByIDTx(tx *gorm.DB, fileID string) (*entities.File, error) {
	var item entities.File
	if err := tx.First(&item, "id = ?", strings.ToLower(strings.TrimSpace(fileID))).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewAppError(404, utils.ErrCodeNotFound, "文件不存在")
		}
		return nil, err
	}
	if item.UploadStatus == fileStatusDeleted {
		return nil, utils.NewAppError(404, utils.ErrCodeNotFound, "文件不存在")
	}
	return &item, nil
}
