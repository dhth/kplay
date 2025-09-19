package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var ErrInvalidEnvValue = errors.New("invalid value provided")

func getBoolEnvVar(envVar string, defaultVal bool) (bool, error) {
	valueStr := os.Getenv(envVar)
	if valueStr == "" {
		return defaultVal, nil
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false, fmt.Errorf("%w for %s: %q; expected a boolean value", ErrInvalidEnvValue, envVar, valueStr)
	}

	return value, nil
}

func getStringEnvVar(envVar string, defaultVal string) string {
	valueStr := os.Getenv(envVar)
	if valueStr == "" {
		return defaultVal
	}
	return valueStr
}

func getConstrainedStringEnvVar(envVar, defaultVal string, minLen, maxLen int) (string, error) {
	valueStr := os.Getenv(envVar)
	if valueStr == "" {
		return defaultVal, nil
	}

	trimmedValue := strings.TrimSpace(valueStr)
	if len(trimmedValue) < minLen {
		return "", fmt.Errorf("%w for %s too short: %q; needs to be at least %d characters", ErrInvalidEnvValue, envVar, valueStr, minLen)
	}
	if len(trimmedValue) > maxLen {
		return "", fmt.Errorf("%w for %s too long: %q; needs to be at most %d characters", ErrInvalidEnvValue, envVar, valueStr, maxLen)
	}

	return trimmedValue, nil
}

func getUintEnvVar[T uint16 | uint32](envVar string, defaultVal, minVal, maxVal T, bitSize int) (T, error) {
	valueStr := os.Getenv(envVar)
	if valueStr == "" {
		return defaultVal, nil
	}

	value64, err := strconv.ParseUint(valueStr, 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("%w for %s: %q; expected a valid integer in the range [%d,%d]", ErrInvalidEnvValue, envVar, valueStr, minVal, maxVal)
	}

	value := T(value64)
	if value < minVal || value > maxVal {
		return 0, fmt.Errorf("%w for %s out of range: %d; expected range: [%d,%d]", ErrInvalidEnvValue, envVar, value, minVal, maxVal)
	}

	return value, nil
}

func getUint16EnvVar(envVar string, defaultVal, minVal, maxVal uint16) (uint16, error) {
	return getUintEnvVar(envVar, defaultVal, minVal, maxVal, 16)
}

func getUint32EnvVar(envVar string, defaultVal, minVal, maxVal uint32) (uint32, error) {
	return getUintEnvVar(envVar, defaultVal, minVal, maxVal, 32)
}
