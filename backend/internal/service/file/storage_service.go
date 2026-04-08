package fileservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"backend/internal/pkg/config"
	"backend/internal/pkg/utils"
	systemrepo "backend/internal/repository/system"
	sharedservice "backend/internal/service/shared"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
)

const cosCORSManagedRuleID = "mercall-direct-upload"

type StorageService struct {
	cfg     *config.Config
	cfgRepo *systemrepo.SystemConfigRepository
}

type StoredFile struct {
	StorageDriver string
	StoragePath   string
	URL           string
}

type DirectUploadTicket struct {
	StorageDriver string
	Method        string
	UploadURL     string
	Headers       map[string]string
	ExpiresAt     time.Time
}

type COSCORSSyncResult struct {
	COSConfigured         bool
	DirectUploadSupported bool
	AutoSyncEnabled       bool
	Changed               bool
	RuleID                string
	AllowedOrigins        []string
	AllowedMethods        []string
	AllowedHeaders        []string
	ExposeHeaders         []string
	MaxAgeSeconds         int64
	SkippedReason         string
}

type UploadOptions struct {
	AllowedSuffixes  []string
	SubDir           string
	MaxSizeBytes     int64
	SizeLimitMessage string
	ObjectName       string
}

func NewStorageService(cfg *config.Config, cfgRepo *systemrepo.SystemConfigRepository) *StorageService {
	return &StorageService{cfg: cfg, cfgRepo: cfgRepo}
}

func (s *StorageService) Upload(fileHeader *multipart.FileHeader) (string, error) {
	return s.UploadWithOptions(fileHeader, UploadOptions{})
}

func (s *StorageService) UploadWithOptions(fileHeader *multipart.FileHeader, options UploadOptions) (string, error) {
	storedFile, err := s.StoreWithOptions(fileHeader, options)
	if err != nil {
		return "", err
	}
	return storedFile.URL, nil
}

func (s *StorageService) StoreWithOptions(fileHeader *multipart.FileHeader, options UploadOptions) (*StoredFile, error) {
	if err := s.validate(fileHeader, options); err != nil {
		return nil, err
	}

	driver := strings.ToLower(s.cfg.UploadDriver)
	if driver == "cos" || driver == "auto" || driver == "" {
		if s.canUseCOS() {
			storedFile, err := s.uploadToCOS(fileHeader, options)
			if err == nil {
				return storedFile, nil
			}
			if driver == "cos" {
				return nil, err
			}
		}
	}

	return s.uploadToLocal(fileHeader, options)
}

func (s *StorageService) StoreBytesWithOptions(fileName string, data []byte, options UploadOptions) (*StoredFile, error) {
	if err := s.validateMeta(fileName, int64(len(data)), options); err != nil {
		return nil, err
	}

	driver := strings.ToLower(s.cfg.UploadDriver)
	if driver == "cos" || driver == "auto" || driver == "" {
		if s.canUseCOS() {
			storedFile, err := s.uploadBytesToCOS(fileName, data, options)
			if err == nil {
				return storedFile, nil
			}
			if driver == "cos" {
				return nil, err
			}
		}
	}

	return s.uploadBytesToLocal(fileName, data, options)
}

func (s *StorageService) validate(fileHeader *multipart.FileHeader, options UploadOptions) error {
	if fileHeader == nil {
		return utils.NewAppError(400, utils.ErrCodeInvalidRequest, "请选择要上传的文件")
	}
	return s.validateMeta(fileHeader.Filename, fileHeader.Size, options)
}

func (s *StorageService) validateMeta(fileName string, fileSize int64, options UploadOptions) error {
	if options.MaxSizeBytes > 0 && fileSize > options.MaxSizeBytes {
		message := strings.TrimSpace(options.SizeLimitMessage)
		if message == "" {
			message = "文件过大，请压缩后再上传"
		}
		return utils.NewAppError(400, utils.ErrCodeFileTooLarge, message)
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	if len(options.AllowedSuffixes) > 0 {
		allowed := make(map[string]struct{}, len(options.AllowedSuffixes))
		for _, item := range options.AllowedSuffixes {
			normalized := strings.TrimSpace(strings.ToLower(item))
			if normalized != "" {
				allowed[normalized] = struct{}{}
			}
		}
		if _, ok := allowed[ext]; !ok {
			return utils.NewAppError(400, utils.ErrCodeFileExtensionUnsupported, "不支持的文件格式")
		}
	}

	return nil
}

func (s *StorageService) canUseCOS() bool {
	return s.cfg.COSSecretID != "" && s.cfg.COSSecretKey != "" && s.cfg.COSBucketURL != ""
}

func (s *StorageService) prefersCOS() bool {
	driver := strings.ToLower(strings.TrimSpace(s.cfg.UploadDriver))
	return driver == "" || driver == "auto" || driver == "cos"
}

func (s *StorageService) SupportsDirectUpload() bool {
	return s.prefersCOS() && s.canUseCOS()
}

func (s *StorageService) newCOSClient() (*cos.Client, error) {
	bucketURL, err := url.Parse(s.cfg.COSBucketURL)
	if err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "COS Bucket 地址配置不正确", err)
	}

	return cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  s.cfg.COSSecretID,
			SecretKey: s.cfg.COSSecretKey,
		},
	}), nil
}

func (s *StorageService) uploadToCOS(fileHeader *multipart.FileHeader, options UploadOptions) (*StoredFile, error) {
	client, err := s.newCOSClient()
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "读取上传文件失败", err)
	}
	defer file.Close()

	keyPrefix := "uploads"
	if strings.TrimSpace(options.SubDir) != "" {
		keyPrefix += "/" + strings.Trim(strings.ReplaceAll(options.SubDir, "\\", "/"), "/")
	}
	key := s.buildStorageObjectPath(keyPrefix, fileHeader.Filename, options)
	_, err = client.Object.Put(context.Background(), key, file, nil)
	if err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "上传到 COS 失败", err)
	}

	return &StoredFile{
		StorageDriver: "cos",
		StoragePath:   key,
		URL:           s.BuildURL("cos", key),
	}, nil
}

func (s *StorageService) uploadToLocal(fileHeader *multipart.FileHeader, options UploadOptions) (*StoredFile, error) {
	basePath := s.cfg.UploadLocalPath
	if strings.TrimSpace(options.SubDir) != "" {
		basePath = filepath.Join(basePath, options.SubDir)
	}
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "创建本地上传目录失败", err)
	}

	name := s.buildObjectName(fileHeader.Filename, options)
	fullPath := filepath.Join(basePath, name)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "读取上传文件失败", err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "创建本地文件失败", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "保存上传文件失败", err)
	}

	urlPath := name
	if strings.TrimSpace(options.SubDir) != "" {
		urlPath = strings.Trim(strings.ReplaceAll(options.SubDir, "\\", "/"), "/") + "/" + name
	}
	return &StoredFile{
		StorageDriver: "local",
		StoragePath:   urlPath,
		URL:           s.BuildURL("local", urlPath),
	}, nil
}

func (s *StorageService) uploadBytesToCOS(fileName string, data []byte, options UploadOptions) (*StoredFile, error) {
	client, err := s.newCOSClient()
	if err != nil {
		return nil, err
	}

	keyPrefix := "uploads"
	if strings.TrimSpace(options.SubDir) != "" {
		keyPrefix += "/" + strings.Trim(strings.ReplaceAll(options.SubDir, "\\", "/"), "/")
	}
	key := s.buildStorageObjectPath(keyPrefix, fileName, options)
	_, err = client.Object.Put(context.Background(), key, bytes.NewReader(data), nil)
	if err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "上传到 COS 失败", err)
	}

	return &StoredFile{
		StorageDriver: "cos",
		StoragePath:   key,
		URL:           s.BuildURL("cos", key),
	}, nil
}

func (s *StorageService) uploadBytesToLocal(fileName string, data []byte, options UploadOptions) (*StoredFile, error) {
	basePath := s.cfg.UploadLocalPath
	if strings.TrimSpace(options.SubDir) != "" {
		basePath = filepath.Join(basePath, options.SubDir)
	}
	if err := os.MkdirAll(basePath, 0o755); err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "创建本地上传目录失败", err)
	}

	name := s.buildObjectName(fileName, options)
	fullPath := filepath.Join(basePath, name)
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "保存上传文件失败", err)
	}

	urlPath := name
	if strings.TrimSpace(options.SubDir) != "" {
		urlPath = strings.Trim(strings.ReplaceAll(options.SubDir, "\\", "/"), "/") + "/" + name
	}
	return &StoredFile{
		StorageDriver: "local",
		StoragePath:   urlPath,
		URL:           s.BuildURL("local", urlPath),
	}, nil
}

func (s *StorageService) BuildURL(storageDriver, storagePath string) string {
	path := strings.Trim(strings.ReplaceAll(storagePath, "\\", "/"), "/")
	if path == "" {
		return ""
	}

	switch strings.ToLower(strings.TrimSpace(storageDriver)) {
	case "local":
		return "/static/uploads/" + path
	case "cos":
		return s.buildCOSDownloadURL(path)
	default:
		return path
	}
}

func (s *StorageService) buildCOSDownloadURL(path string) string {
	if !s.canUseCOS() {
		return s.buildCOSObjectURL(path)
	}

	expireMinutes := resolveCOSDownloadSignedURLExpireMinutes(s.cfgRepo)
	expire := time.Duration(expireMinutes) * time.Minute
	if expire <= 0 {
		expire = time.Duration(defaultCOSDownloadSignedURLExpireMinutes) * time.Minute
	}

	client, err := s.newCOSClient()
	if err == nil {
		signedURL, signErr := client.Object.GetPresignedURL3(
			context.Background(),
			http.MethodGet,
			path,
			expire,
			nil,
			false,
		)
		if signErr == nil && signedURL != nil {
			return signedURL.String()
		}
	}

	return s.buildCOSObjectURL(path)
}

// BuildCOSDownloadRedirectURL returns a presigned GET URL for COS objects, with response-content-disposition
// so browsers suggest SanitizeDownloadFilename(originalName, fallback) as the saved name. Used by FileController download only.
func (s *StorageService) BuildCOSDownloadRedirectURL(storagePath, originalName string) string {
	path := strings.Trim(strings.ReplaceAll(storagePath, "\\", "/"), "/")
	if path == "" {
		return ""
	}
	fallback := utils.FallbackDownloadFilename(storagePath)
	safe := utils.SanitizeDownloadFilename(originalName, fallback)
	disp := utils.ContentDispositionAttachmentValue(safe)

	if !s.canUseCOS() {
		return ""
	}

	q := url.Values{}
	q.Set("response-content-disposition", disp)

	expireMinutes := resolveCOSDownloadSignedURLExpireMinutes(s.cfgRepo)
	expire := time.Duration(expireMinutes) * time.Minute
	if expire <= 0 {
		expire = time.Duration(defaultCOSDownloadSignedURLExpireMinutes) * time.Minute
	}

	client, err := s.newCOSClient()
	if err != nil {
		return ""
	}
	signedURL, signErr := client.Object.GetPresignedURL3(
		context.Background(),
		http.MethodGet,
		path,
		expire,
		&cos.PresignedURLOptions{Query: &q},
		false,
	)
	if signErr != nil || signedURL == nil {
		return ""
	}
	return signedURL.String()
}

func (s *StorageService) buildCOSObjectURL(path string) string {
	return s.buildCOSObjectURLWithQuery(path, nil)
}

func (s *StorageService) buildCOSObjectURLWithQuery(path string, values url.Values) string {
	base := strings.TrimSuffix(s.cfg.COSBaseURL, "/")
	if base != "" {
		return buildStorageObjectURL(base, path, values)
	}
	return buildStorageObjectURL(strings.TrimSuffix(s.cfg.COSBucketURL, "/"), path, values)
}

func buildStorageObjectURL(base, path string, values url.Values) string {
	target := strings.TrimRight(strings.TrimSpace(base), "/")
	if target == "" {
		target = "/"
	}
	target = target + "/" + strings.TrimLeft(path, "/")
	if len(values) == 0 {
		return target
	}
	encoded := values.Encode()
	if encoded == "" {
		return target
	}
	return target + "?" + encoded
}

func (s *StorageService) Delete(storageDriver, storagePath string) error {
	path := strings.Trim(strings.ReplaceAll(storagePath, "\\", "/"), "/")
	switch strings.ToLower(strings.TrimSpace(storageDriver)) {
	case "local":
		fullPath := filepath.Join(s.cfg.UploadLocalPath, filepath.FromSlash(path))
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return utils.WrapAppError(500, utils.ErrCodeInternal, "删除本地文件失败", err)
		}
		return nil
	case "cos":
		if !s.canUseCOS() {
			return utils.NewAppError(500, utils.ErrCodeInternal, "COS 未正确配置，无法删除文件")
		}

		client, err := s.newCOSClient()
		if err != nil {
			return err
		}

		if _, err := client.Object.Delete(context.Background(), path); err != nil {
			return utils.WrapAppError(500, utils.ErrCodeInternal, "删除 COS 文件失败", err)
		}
		return nil
	default:
		return utils.NewAppError(400, utils.ErrCodeInvalidRequest, "不支持的存储驱动")
	}
}

func (s *StorageService) ResolveLocalFilePath(storagePath string) (string, error) {
	path := strings.Trim(strings.ReplaceAll(storagePath, "\\", "/"), "/")
	if path == "" {
		return "", utils.NewAppError(400, utils.ErrCodeInvalidRequest, "文件路径不能为空")
	}

	baseAbs, err := filepath.Abs(s.cfg.UploadLocalPath)
	if err != nil {
		return "", utils.WrapAppError(500, utils.ErrCodeInternal, "解析本地上传目录失败", err)
	}

	fullPath := filepath.Join(baseAbs, filepath.FromSlash(path))
	fullAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return "", utils.WrapAppError(500, utils.ErrCodeInternal, "解析本地文件路径失败", err)
	}

	prefix := baseAbs + string(os.PathSeparator)
	if fullAbs != baseAbs && !strings.HasPrefix(fullAbs, prefix) {
		return "", utils.NewAppError(400, utils.ErrCodeInvalidRequest, "文件路径不合法")
	}

	return fullAbs, nil
}

func (s *StorageService) InitDirectUpload(storagePath, mimeType string) (*DirectUploadTicket, error) {
	path := strings.Trim(strings.ReplaceAll(storagePath, "\\", "/"), "/")
	if path == "" {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "上传路径不能为空")
	}
	if !s.SupportsDirectUpload() {
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "当前存储驱动不支持直传")
	}

	client, err := s.newCOSClient()
	if err != nil {
		return nil, err
	}

	expireMinutes := resolveCOSDirectUploadSignedURLExpireMinutes(s.cfgRepo)
	expire := time.Duration(expireMinutes) * time.Minute
	if expire <= 0 {
		expire = time.Duration(defaultCOSDirectUploadSignedURLExpireMinutes) * time.Minute
	}

	headers := http.Header{}
	headerMap := make(map[string]string)
	if normalizedMime := strings.TrimSpace(mimeType); normalizedMime != "" {
		headers.Set("Content-Type", normalizedMime)
		headerMap["Content-Type"] = normalizedMime
	}

	signedURL, err := client.Object.GetPresignedURL3(context.Background(), http.MethodPut, path, expire, &cos.PresignedURLOptions{
		Header: &headers,
	}, false)
	if err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "生成直传签名地址失败", err)
	}

	return &DirectUploadTicket{
		StorageDriver: "cos",
		Method:        http.MethodPut,
		UploadURL:     signedURL.String(),
		Headers:       headerMap,
		ExpiresAt:     time.Now().Add(expire),
	}, nil
}

func (s *StorageService) SyncCOSCORS(ctx context.Context) (*COSCORSSyncResult, error) {
	return s.syncCOSCORS(ctx, false)
}

func (s *StorageService) ForceSyncCOSCORS(ctx context.Context) (*COSCORSSyncResult, error) {
	return s.syncCOSCORS(ctx, true)
}

func (s *StorageService) syncCOSCORS(ctx context.Context, force bool) (*COSCORSSyncResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	result := &COSCORSSyncResult{
		COSConfigured:         s.canUseCOS(),
		DirectUploadSupported: s.SupportsDirectUpload(),
		AutoSyncEnabled:       resolveCOSCORSAutoSyncEnabled(s.cfgRepo),
		RuleID:                cosCORSManagedRuleID,
	}

	if !result.COSConfigured {
		result.SkippedReason = "COS 未正确配置"
		return result, nil
	}
	if !force && !result.DirectUploadSupported {
		result.SkippedReason = "当前上传驱动未启用 COS 直传"
		return result, nil
	}
	if !force && !result.AutoSyncEnabled {
		result.SkippedReason = "upload.cos_cors_auto_sync_enabled 已关闭"
		return result, nil
	}

	desiredRule, err := s.buildManagedCOSCORSRule()
	if err != nil {
		return nil, err
	}
	result.AllowedOrigins = sharedservice.CloneStringSlice(desiredRule.AllowedOrigins)
	result.AllowedMethods = sharedservice.CloneStringSlice(desiredRule.AllowedMethods)
	result.AllowedHeaders = sharedservice.CloneStringSlice(desiredRule.AllowedHeaders)
	result.ExposeHeaders = sharedservice.CloneStringSlice(desiredRule.ExposeHeaders)
	result.MaxAgeSeconds = int64(desiredRule.MaxAgeSeconds)

	client, err := s.newCOSClient()
	if err != nil {
		return nil, err
	}

	current, _, err := client.Bucket.GetCORS(ctx)
	if err != nil && !cos.IsNotFoundError(err) {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "读取 COS CORS 配置失败", err)
	}

	var currentRules []cos.BucketCORSRule
	if current != nil {
		currentRules = current.Rules
	}

	mergedRules, changed := mergeManagedCOSCORSRules(currentRules, desiredRule)
	if !changed {
		return result, nil
	}

	if _, err := client.Bucket.PutCORS(ctx, &cos.BucketPutCORSOptions{Rules: mergedRules}); err != nil {
		return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "同步 COS CORS 配置失败", err)
	}

	result.Changed = true
	return result, nil
}

func (s *StorageService) ObjectExists(storageDriver, storagePath string) (bool, error) {
	path := strings.Trim(strings.ReplaceAll(storagePath, "\\", "/"), "/")
	switch strings.ToLower(strings.TrimSpace(storageDriver)) {
	case "local":
		fullPath := filepath.Join(s.cfg.UploadLocalPath, filepath.FromSlash(path))
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, utils.WrapAppError(500, utils.ErrCodeInternal, "检查本地文件失败", err)
		}
		return true, nil
	case "cos":
		if !s.canUseCOS() {
			return false, utils.NewAppError(500, utils.ErrCodeInternal, "COS 未正确配置，无法检查文件")
		}

		client, err := s.newCOSClient()
		if err != nil {
			return false, err
		}

		if _, err := client.Object.Head(context.Background(), path, nil); err != nil {
			var cosErr *cos.ErrorResponse
			if errors.As(err, &cosErr) && cosErr != nil && cosErr.Response != nil && cosErr.Response.StatusCode == http.StatusNotFound {
				return false, nil
			}
			if strings.Contains(strings.ToLower(err.Error()), "status code: 404") || strings.Contains(strings.ToLower(err.Error()), "no such object") {
				return false, nil
			}
			return false, utils.WrapAppError(500, utils.ErrCodeInternal, "检查 COS 文件失败", err)
		}
		return true, nil
	default:
		return false, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "不支持的存储驱动")
	}
}

func (s *StorageService) ReadAll(storageDriver, storagePath string) ([]byte, error) {
	path := strings.Trim(strings.ReplaceAll(storagePath, "\\", "/"), "/")
	switch strings.ToLower(strings.TrimSpace(storageDriver)) {
	case "local":
		fullPath, err := s.ResolveLocalFilePath(path)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "读取本地文件失败", err)
		}
		return data, nil
	case "cos":
		if !s.canUseCOS() {
			return nil, utils.NewAppError(500, utils.ErrCodeInternal, "COS 未正确配置，无法读取文件")
		}

		client, err := s.newCOSClient()
		if err != nil {
			return nil, err
		}

		resp, err := client.Object.Get(context.Background(), path, nil)
		if err != nil {
			return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "读取 COS 文件失败", err)
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, utils.WrapAppError(500, utils.ErrCodeInternal, "读取 COS 文件失败", err)
		}
		return data, nil
	default:
		return nil, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "不支持的存储驱动")
	}
}

func (s *StorageService) buildManagedCOSCORSRule() (cos.BucketCORSRule, error) {
	origins := normalizeCOSCORSValues(resolveCOSCORSAllowedOrigins(s.cfgRepo), false)
	if len(origins) == 0 {
		return cos.BucketCORSRule{}, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "请先配置 upload.cos_cors_allowed_origins")
	}

	methods := normalizeCOSCORSValues(resolveCOSCORSAllowedMethods(s.cfgRepo), true)
	if len(methods) == 0 {
		return cos.BucketCORSRule{}, utils.NewAppError(400, utils.ErrCodeInvalidRequest, "请先配置 upload.cos_cors_allowed_methods")
	}

	headers := normalizeCOSCORSValues(resolveCOSCORSAllowedHeaders(s.cfgRepo), false)
	if len(headers) == 0 {
		headers = []string{"*"}
	}

	maxAgeSeconds := resolveCOSCORSMaxAgeSeconds(s.cfgRepo)
	if maxAgeSeconds <= 0 {
		maxAgeSeconds = defaultCOSCORSMaxAgeSeconds
	}

	return cos.BucketCORSRule{
		ID:             cosCORSManagedRuleID,
		AllowedOrigins: origins,
		AllowedMethods: methods,
		AllowedHeaders: headers,
		ExposeHeaders:  normalizeCOSCORSValues(resolveCOSCORSExposeHeaders(s.cfgRepo), false),
		MaxAgeSeconds:  int(maxAgeSeconds),
	}, nil
}

func mergeManagedCOSCORSRules(current []cos.BucketCORSRule, desired cos.BucketCORSRule) ([]cos.BucketCORSRule, bool) {
	desired = normalizeManagedCOSCORSRule(desired)

	merged := make([]cos.BucketCORSRule, 0, len(current)+1)
	found := false
	changed := false
	for _, item := range current {
		if strings.TrimSpace(item.ID) != cosCORSManagedRuleID {
			merged = append(merged, item)
			continue
		}

		if found {
			changed = true
			continue
		}

		found = true
		if !equalCOSCORSRules(item, desired) {
			changed = true
		}
		merged = append(merged, desired)
	}

	if !found {
		merged = append(merged, desired)
		changed = true
	}

	return merged, changed
}

func equalCOSCORSRules(left, right cos.BucketCORSRule) bool {
	lhs := normalizeManagedCOSCORSRule(left)
	rhs := normalizeManagedCOSCORSRule(right)
	if lhs.ID != rhs.ID || lhs.MaxAgeSeconds != rhs.MaxAgeSeconds {
		return false
	}
	return equalStringSlices(lhs.AllowedOrigins, rhs.AllowedOrigins) &&
		equalStringSlices(lhs.AllowedMethods, rhs.AllowedMethods) &&
		equalStringSlices(lhs.AllowedHeaders, rhs.AllowedHeaders) &&
		equalStringSlices(lhs.ExposeHeaders, rhs.ExposeHeaders)
}

func normalizeManagedCOSCORSRule(rule cos.BucketCORSRule) cos.BucketCORSRule {
	rule.ID = cosCORSManagedRuleID
	rule.AllowedOrigins = normalizeCOSCORSValues(rule.AllowedOrigins, false)
	rule.AllowedMethods = normalizeCOSCORSValues(rule.AllowedMethods, true)
	rule.AllowedHeaders = normalizeCOSCORSValues(rule.AllowedHeaders, false)
	rule.ExposeHeaders = normalizeCOSCORSValues(rule.ExposeHeaders, false)
	if rule.MaxAgeSeconds < 0 {
		rule.MaxAgeSeconds = 0
	}
	return rule
}

func normalizeCOSCORSValues(values []string, uppercase bool) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		value := strings.TrimSpace(item)
		if uppercase {
			value = strings.ToUpper(value)
		}
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, value)
	}
	sort.Strings(normalized)
	return normalized
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func (s *StorageService) ValidateMeta(fileName string, fileSize int64, options UploadOptions) error {
	return s.validateMeta(fileName, fileSize, options)
}

func (s *StorageService) BuildStorageObjectPath(prefix, fileName string, options UploadOptions) string {
	return s.buildStorageObjectPath(prefix, fileName, options)
}

func (s *StorageService) buildStorageObjectPath(prefix, fileName string, options UploadOptions) string {
	name := s.buildObjectName(fileName, options)
	prefix = strings.Trim(strings.ReplaceAll(prefix, "\\", "/"), "/")
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}

func (s *StorageService) buildObjectName(fileName string, options UploadOptions) string {
	if strings.TrimSpace(options.ObjectName) != "" {
		return strings.Trim(strings.ReplaceAll(options.ObjectName, "\\", "/"), "/")
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	return fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), uuid.NewString(), ext)
}
