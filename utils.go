package tsing

import (
	"os"
	"path"
)

// 取最后一个字符
func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

// 拼接路径
func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}

// 获得程序的根路径
func getRootPath() string {
	// 使用当前路径
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	return path
}
