package utils

import (
	"fmt"
	"strings"
)

func RightPadTrim(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return s[:length-3] + "..."
		}
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

func TrimLeft(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return "..." + s[len(s)-length+3:]
		}
		return s[:length]
	}
	return s
}

func TrimRight(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return s[:length-3] + "..."
		}
		return s[:length]
	}
	return s
}

func HumanReadableBytes(bytes uint64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d bytes", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	}
}
