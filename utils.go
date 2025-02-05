package tsing

import (
	"bytes"
	"path"
	"strings"
)

// strToBytes 将字符串转换为字节数组
func strToBytes(s string) []byte {
	// 使用标准库的 []byte 转换
	return []byte(s)
}

// bytesToStr 将字节数组转换为字符串
func bytesToStr(b []byte) string {
	// 使用标准库的 string 转换
	return string(b)
}

// joinPaths 合并绝对路径和相对路径，生成最终路径
func joinPaths(absolutePath, relativePath string) string {
	// 如果相对路径为空，直接返回绝对路径
	if relativePath == "" {
		return absolutePath
	}

	// 安全地合并路径
	finalPath := path.Join(absolutePath, relativePath)

	// 如果相对路径以 '/' 结尾，确保最终路径也以 '/' 结尾
	if strings.HasSuffix(relativePath, "/") {
		return finalPath + "/"
	}
	return finalPath
}

// minNumber 返回两个整数中的较小值
func minNumber(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// longestCommonPrefix 计算两个字符串的最长公共前缀长度
func longestCommonPrefix(a, b string) int {
	maxNumber := minNumber(len(a), len(b))
	for i := 0; i < maxNumber; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return maxNumber
}

// countParams 统计路径中参数的数量
func countParams(path string) int {
	return bytes.Count(strToBytes(path), strColon) + bytes.Count(strToBytes(path), strStar)
}

// countSections 统计路径中的段数
func countSections(path string) int {
	return bytes.Count(strToBytes(path), strSlash)
}
