package fileservice

import (
	"fmt"
	"strings"

	systemrepo "backend/internal/repository/system"
	sharedservice "backend/internal/service/shared"
)

const (
	fileAccessSignedURLExpireMinutesConfigKey            = "upload.file_access_signed_url_expire_minutes"
	cosDirectUploadSignedURLExpireMinutesConfigKey       = "upload.cos_direct_upload_signed_url_expire_minutes"
	cosDownloadSignedURLExpireMinutesConfigKey           = "upload.cos_download_signed_url_expire_minutes"
	cosCORSAutoSyncEnabledConfigKey                      = "upload.cos_cors_auto_sync_enabled"
	cosCORSAllowedOriginsConfigKey                       = "upload.cos_cors_allowed_origins"
	cosCORSAllowedMethodsConfigKey                       = "upload.cos_cors_allowed_methods"
	cosCORSAllowedHeadersConfigKey                       = "upload.cos_cors_allowed_headers"
	cosCORSExposeHeadersConfigKey                        = "upload.cos_cors_expose_headers"
	cosCORSMaxAgeSecondsConfigKey                        = "upload.cos_cors_max_age_seconds"
	unboundFileTTLHoursConfigKey                         = "upload.unbound_file_ttl_hours"
	unboundFileCleanupIntervalMinutesConfigKey           = "upload.unbound_file_cleanup_interval_minutes"
	unboundFileCleanupBatchSizeConfigKey                 = "upload.unbound_file_cleanup_batch_size"
	defaultFileAccessSignedURLExpireMinutes        int64 = 60
	defaultCOSDirectUploadSignedURLExpireMinutes   int64 = 30
	defaultCOSDownloadSignedURLExpireMinutes       int64 = 60
	defaultCOSCORSAutoSyncEnabled                        = true
	defaultCOSCORSMaxAgeSeconds                    int64 = 600
	defaultUnboundFileTTLHours                     int64 = 24
	defaultUnboundFileCleanupIntervalMin           int64 = 30
	defaultUnboundFileCleanupBatchSize             int64 = 100
)

const FileAccessSignedURLExpireMinutesConfigKey = fileAccessSignedURLExpireMinutesConfigKey
const COSDirectUploadSignedURLExpireMinutesConfigKey = cosDirectUploadSignedURLExpireMinutesConfigKey
const COSDownloadSignedURLExpireMinutesConfigKey = cosDownloadSignedURLExpireMinutesConfigKey
const COSCORSAutoSyncEnabledConfigKey = cosCORSAutoSyncEnabledConfigKey
const COSCORSAllowedOriginsConfigKey = cosCORSAllowedOriginsConfigKey
const COSCORSAllowedMethodsConfigKey = cosCORSAllowedMethodsConfigKey
const COSCORSAllowedHeadersConfigKey = cosCORSAllowedHeadersConfigKey
const COSCORSExposeHeadersConfigKey = cosCORSExposeHeadersConfigKey
const COSCORSMaxAgeSecondsConfigKey = cosCORSMaxAgeSecondsConfigKey
const UnboundFileTTLHoursConfigKey = unboundFileTTLHoursConfigKey
const UnboundFileCleanupIntervalMinutesConfigKey = unboundFileCleanupIntervalMinutesConfigKey
const UnboundFileCleanupBatchSizeConfigKey = unboundFileCleanupBatchSizeConfigKey

func resolveFileAccessSignedURLExpireMinutes(cfgRepo *systemrepo.SystemConfigRepository) int64 {
	if cfgRepo != nil {
		if item, err := cfgRepo.FindByKey(fileAccessSignedURLExpireMinutesConfigKey); err == nil && item != nil {
			if parsed, ok := sharedservice.ParsePositiveInt64(item.ConfigVal); ok {
				return parsed
			}
		}
		if item, err := cfgRepo.FindByKey(cosDownloadSignedURLExpireMinutesConfigKey); err == nil && item != nil {
			if parsed, ok := sharedservice.ParsePositiveInt64(item.ConfigVal); ok {
				return parsed
			}
		}
	}

	return defaultFileAccessSignedURLExpireMinutes
}

func resolveCOSDirectUploadSignedURLExpireMinutes(cfgRepo *systemrepo.SystemConfigRepository) int64 {
	return sharedservice.ResolvePositiveInt64Config(cfgRepo, cosDirectUploadSignedURLExpireMinutesConfigKey, defaultCOSDirectUploadSignedURLExpireMinutes)
}

func resolveCOSDownloadSignedURLExpireMinutes(cfgRepo *systemrepo.SystemConfigRepository) int64 {
	return sharedservice.ResolvePositiveInt64Config(cfgRepo, cosDownloadSignedURLExpireMinutesConfigKey, defaultCOSDownloadSignedURLExpireMinutes)
}

func resolveCOSCORSAutoSyncEnabled(cfgRepo *systemrepo.SystemConfigRepository) bool {
	return sharedservice.ResolveBoolConfig(cfgRepo, cosCORSAutoSyncEnabledConfigKey, defaultCOSCORSAutoSyncEnabled)
}

func resolveCOSCORSAllowedOrigins(cfgRepo *systemrepo.SystemConfigRepository) []string {
	return sharedservice.ResolveCommaSeparatedConfig(cfgRepo, cosCORSAllowedOriginsConfigKey, []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"https://mercall.leapinfra.cn",
		"http://mercall.leapinfra.cn",
		"http://dev-mercall.leapinfra.cn",
		"https://dev-mercall.leapinfra.cn",
	}, func(value string) string {
		return strings.TrimSpace(value)
	})
}

func resolveCOSCORSAllowedMethods(cfgRepo *systemrepo.SystemConfigRepository) []string {
	return sharedservice.ResolveCommaSeparatedConfig(cfgRepo, cosCORSAllowedMethodsConfigKey, []string{"GET", "HEAD", "PUT"}, func(value string) string {
		return strings.ToUpper(strings.TrimSpace(value))
	})
}

func resolveCOSCORSAllowedHeaders(cfgRepo *systemrepo.SystemConfigRepository) []string {
	return sharedservice.ResolveCommaSeparatedConfig(cfgRepo, cosCORSAllowedHeadersConfigKey, []string{"*"}, func(value string) string {
		return strings.TrimSpace(value)
	})
}

func resolveCOSCORSExposeHeaders(cfgRepo *systemrepo.SystemConfigRepository) []string {
	return sharedservice.ResolveCommaSeparatedConfig(cfgRepo, cosCORSExposeHeadersConfigKey, []string{"ETag", "x-cos-request-id"}, func(value string) string {
		return strings.TrimSpace(value)
	})
}

func resolveCOSCORSMaxAgeSeconds(cfgRepo *systemrepo.SystemConfigRepository) int64 {
	return sharedservice.ResolvePositiveInt64Config(cfgRepo, cosCORSMaxAgeSecondsConfigKey, defaultCOSCORSMaxAgeSeconds)
}

func buildGenericSizeLimitMessage(limitMB int64) string {
	return fmt.Sprintf("文件大小不能超过 %d MB", limitMB)
}

func resolveUnboundFileTTLHours(cfgRepo *systemrepo.SystemConfigRepository) int64 {
	return sharedservice.ResolvePositiveInt64Config(cfgRepo, unboundFileTTLHoursConfigKey, defaultUnboundFileTTLHours)
}

func resolveUnboundFileCleanupIntervalMinutes(cfgRepo *systemrepo.SystemConfigRepository) int64 {
	return sharedservice.ResolvePositiveInt64Config(cfgRepo, unboundFileCleanupIntervalMinutesConfigKey, defaultUnboundFileCleanupIntervalMin)
}

func resolveUnboundFileCleanupBatchSize(cfgRepo *systemrepo.SystemConfigRepository) int64 {
	return sharedservice.ResolvePositiveInt64Config(cfgRepo, unboundFileCleanupBatchSizeConfigKey, defaultUnboundFileCleanupBatchSize)
}
