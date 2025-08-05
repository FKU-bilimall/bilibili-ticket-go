package utils

import (
	"os"
)

// IsFileEmpty 检查指定文件是否为空
func IsFileEmpty(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Size() == 0
}
