package pkg

import (
	"os"
	"strconv"
	"time"
)

func GetEnvInt(env string, fallback int) int {
	if value, ok := os.LookupEnv(env); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func GetEnvDuration(env string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(env); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue) * time.Millisecond
		}
	}
	return fallback
}
