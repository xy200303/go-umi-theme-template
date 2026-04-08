package sharedservice

import (
	"strconv"
	"strings"

	systemrepo "backend/internal/repository/system"
)

func CloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func ParsePositiveInt64(value string) (int64, bool) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}

func ResolvePositiveInt64Config(cfgRepo *systemrepo.SystemConfigRepository, configKey string, fallback int64) int64 {
	if cfgRepo == nil {
		return fallback
	}

	item, err := cfgRepo.FindByKey(configKey)
	if err != nil || item == nil {
		return fallback
	}

	if parsed, ok := ParsePositiveInt64(item.ConfigVal); ok {
		return parsed
	}

	return fallback
}

func ResolveBoolConfig(cfgRepo *systemrepo.SystemConfigRepository, configKey string, fallback bool) bool {
	if cfgRepo == nil {
		return fallback
	}

	item, err := cfgRepo.FindByKey(configKey)
	if err != nil || item == nil {
		return fallback
	}

	switch strings.ToLower(strings.TrimSpace(item.ConfigVal)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func ResolveCommaSeparatedConfig(
	cfgRepo *systemrepo.SystemConfigRepository,
	configKey string,
	fallback []string,
	normalizer func(string) string,
) []string {
	if cfgRepo == nil {
		return CloneStringSlice(fallback)
	}

	item, err := cfgRepo.FindByKey(configKey)
	if err != nil || item == nil {
		return CloneStringSlice(fallback)
	}

	parts := strings.Split(item.ConfigVal, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if normalizer != nil {
			value = normalizer(value)
		}
		if value == "" {
			continue
		}
		result = append(result, value)
	}

	if len(result) == 0 {
		return CloneStringSlice(fallback)
	}

	return result
}

func ResolveAllowedSuffixesConfig(cfgRepo *systemrepo.SystemConfigRepository, configKey string, fallback []string) []string {
	return ResolveCommaSeparatedConfig(cfgRepo, configKey, fallback, func(value string) string {
		return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), ".")
	})
}
