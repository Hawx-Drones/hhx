package util

import (
	"fmt"
	"strings"
)

// FormatSize formats a file size in bytes to a human-readable string
func FormatSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	unitIndex := 0
	floatSize := float64(size)

	for floatSize >= 1024 && unitIndex < len(units)-1 {
		floatSize /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", size, units[unitIndex])
	}

	return fmt.Sprintf("%.2f %s", floatSize, units[unitIndex])
}

// IsUUID checks if a string is a valid UUID
func IsUUID(str string) bool {
	// Simple UUID check - this is not a comprehensive validation
	// but will help differentiate between names and UUIDs
	if len(str) != 36 {
		return false
	}

	// Check for UUID format (8-4-4-4-12 pattern with hyphens)
	sections := []int{8, 4, 4, 4, 12}
	parts := strings.Split(str, "-")
	if len(parts) != 5 {
		return false
	}

	for i, length := range sections {
		if len(parts[i]) != length {
			return false
		}
	}

	return true
}
