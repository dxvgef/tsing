package tsing

import (
	"bytes"
	"path"
	"unsafe"
)

func strToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

func bytesToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)

	// 如果 relativePath 以 '/' 结尾，确保 finalPath 也以 '/' 结尾
	if lastChar(relativePath) == '/' {
		return finalPath + "/"
	}
	return finalPath
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func minNumber(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	maxNumber := minNumber(len(a), len(b))
	for i := 0; i < maxNumber; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return maxNumber
}

func countParams(path string) int {
	var n int
	s := strToBytes(path)
	n += bytes.Count(s, strColon)
	n += bytes.Count(s, strStar)
	return n
}

func countSections(path string) int {
	return bytes.Count(strToBytes(path), strSlash)
}
