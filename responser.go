// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package tsing

import (
	"bufio"
	"net"
	"net/http"
)

// 自定义http.Responser.ResponseWriter
type _CustomResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier // todo 官方不建议使用

	// 获得当前请求的HTTP响应状态码
	// Status() int

	// 获得已经写入响应的字节大小，用于防止出现多次响应的错误
	// 具体实现见Written()
	Size() int

	// 如果响应数据已经有了，则返回true
	// Written() bool

	// 强制写入http头信息
	OverWriteHeader()
}

type customResponseWriter struct {
	http.ResponseWriter
	size   int // 当值为-1时，说明response里没有数据
	status int
}

// 重置responseWriter的状态
func (w *customResponseWriter) reset(resp http.ResponseWriter) {
	w.ResponseWriter = resp
	w.size = -1
	w.status = http.StatusOK
}

func (w *customResponseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.size != -1 {
			// todo 这里要抛出500事件
			// debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
		}
		w.status = code
	}
}

// 覆写头信息(状态码)，会替换之前设置过的状态码
func (w *customResponseWriter) OverWriteHeader() {
	if w.size != -1 {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

// 实现http.ReponseWriter
func (w *customResponseWriter) Write(data []byte) (n int, err error) {
	w.OverWriteHeader()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

// func (w *customResponseWriter) Status() int {
// 	return w.status
// }

func (w *customResponseWriter) Size() int {
	return w.size
}

// func (w *customResponseWriter) Written() bool {
// 	return w.size != -1
// }

// 实现http.Hijacker接口
func (w *customResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// 实现http.CloseNotify接口
func (w *customResponseWriter) CloseNotify() <-chan bool {
	// todo 官方不建议使用
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// 实现http.Flush接口
func (w *customResponseWriter) Flush() {
	w.OverWriteHeader()
	w.ResponseWriter.(http.Flusher).Flush()
}
