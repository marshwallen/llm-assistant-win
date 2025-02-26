package utils

import (
	"fmt"
	"os"
)

// EnsureDir 检查并创建目录
func EnsureDir(dirPath string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// 递归创建目录
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	return nil
}