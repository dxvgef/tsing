package tsing

import (
	"os"
	"path"
	"reflect"
	"runtime"
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

// 获得函数信息
func getFuncInfo(obj interface{}) *_Source {
	ptr := reflect.ValueOf(obj).Pointer()
	file, line := runtime.FuncForPC(ptr).FileLine(ptr)
	return &_Source{
		Func: runtime.FuncForPC(ptr).Name(),
		File: file,
		Line: line,
	}
}

// 获得最小值
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// 计算最长公用前缀
func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] == ':' || path[i] == '*' {
			n++
		}
	}
	if n >= 255 {
		return 255
	}
	return uint8(n)
}
