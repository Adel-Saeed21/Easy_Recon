package utils

import (
	"os"
	"strings"
)

func SaveToFile(path string, data []string) error {
	content := strings.Join(data, "\n")
	return os.WriteFile(path, []byte(content), 0644)
}