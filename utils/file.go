package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// IsFileEmpty 检查指定文件是否为空
func IsFileEmpty(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Size() == 0
}

func GetFileNameWithoutExt(path string) string {
	filename := filepath.Base(path)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	return nameWithoutExt
}
