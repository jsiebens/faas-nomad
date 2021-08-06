package types

import (
	"github.com/openfaas/faas-provider/types"
	"time"
)

func ParseStringValueFromMap(values *map[string]string, key string, fallback string) string {
	if values == nil {
		return fallback
	}
	m := *values
	return types.ParseString(m[key], fallback)
}

func ParseIntValueFromMap(values *map[string]string, key string, fallback int) int {
	if values == nil {
		return fallback
	}
	m := *values
	return types.ParseIntValue(m[key], fallback)
}

func ParseBoolValueFromMap(values *map[string]string, key string, fallback bool) bool {
	if values == nil {
		return fallback
	}
	m := *values
	return types.ParseBoolValue(m[key], fallback)
}

func ParseIntOrDurationValueFromMap(values *map[string]string, key string, fallback time.Duration) time.Duration {
	if values == nil {
		return fallback
	}
	m := *values
	return types.ParseIntOrDurationValue(m[key], fallback)
}
