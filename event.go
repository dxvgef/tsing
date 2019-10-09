package tsing

import (
	"errors"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

// Event 事件结构
type Event struct {
	Status int // HTTP状态码
	// Caller struct { // 错误调用信息
	// 	File string // 源码文件
	// 	Line int    // 行号
	// }
	Trace          []string // 跟踪信息
	Message        error    // 消息文本
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

// EventHandler 事件处理器类型
type EventHandler func(Event)

// 500(panic)事件
func (d *App) event500(resp http.ResponseWriter, req *http.Request, err interface{}) {
	if d.Event.Handler == nil {
		return
	}
	event := Event{
		Request:        req,
		ResponseWriter: resp,
		Status:         500,
	}

	if errStr, ok := err.(string); ok == true {
		event.Message = errors.New(errStr)
	} else if errErr, ok := err.(error); ok == true {
		event.Message = errErr
	} else {
		event.Message = errors.New("未知错误")
	}
	if d.Event.EnableTrace == true {
		goRoot := runtime.GOROOT()
		for skip := 0; ; skip++ {
			_, file, line, ok := runtime.Caller(skip)
			// 排除trace中的标准包信息
			if strings.HasPrefix(file, goRoot) == false {
				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line))
			}
			if ok == false {
				break
			}
		}
	}

	d.Event.Handler(event)
}

// 404事件处理
func (d *App) event404(resp http.ResponseWriter, req *http.Request) {
	if d.Event.Handler == nil {
		return
	}
	d.Event.Handler(Event{
		Status:         404,
		Message:        errors.New(http.StatusText(404)),
		Request:        req,
		ResponseWriter: resp,
	})
}

// 405事件处理
func (d *App) event405(resp http.ResponseWriter, req *http.Request) {
	if d.Event.Handler == nil {
		return
	}
	d.Event.Handler(Event{
		Status:         405,
		Message:        errors.New(http.StatusText(405)),
		Request:        req,
		ResponseWriter: resp,
	})
}
