package tsing

import (
	"errors"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// 事件结构
type Event struct {
	Status         int      // HTTP状态码
	Message        error    // 消息文本
	Trigger        _Trigger // 触发信息
	Trace          []string // 跟踪信息
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

// 触发信息
type _Trigger struct {
	Func string // 函数名
	File string // 文件名
	Line int    // 行号
}

// 事件处理器类型
type EventHandler func(*Event)

// 获得函数信息
func getFuncInfo(obj interface{}) _Trigger {
	ptr := reflect.ValueOf(obj).Pointer()
	file, line := runtime.FuncForPC(ptr).FileLine(ptr)
	return _Trigger{
		Func: runtime.FuncForPC(ptr).Name(),
		File: file,
		Line: line,
	}
}

// 触发handler的panic事件
func (d *App) eventHandlerPanic(resp http.ResponseWriter, req *http.Request, err interface{}) {
	if d.Config.EventHandler == nil {
		return
	}
	event := Event{
		Request:        req,
		ResponseWriter: resp,
		Status:         500,
	}

	switch t := err.(type) {
	case string:
		event.Message = errors.New(t)
	case error:
		event.Message = t
	default:
		event.Message = errors.New("未知错误消息类型")
	}

	if d.Config.EventTrace {
		goRoot := runtime.GOROOT()
		for skip := 0; ; skip++ {
			funcPtr, file, line, ok := runtime.Caller(skip)
			// 排除trace中的标准包信息
			if !strings.HasPrefix(file, goRoot) {
				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line))
			}
			if skip == 3 && d.Config.EventTrigger {
				event.Trigger.File = file
				event.Trigger.Line = line
				event.Trigger.Func = runtime.FuncForPC(funcPtr).Name()
			}
			if !ok {
				break
			}
		}
	}

	d.Config.EventHandler(&event)
}

// 触发handler的error事件
func (d *App) eventHandlerError(resp http.ResponseWriter, req *http.Request, trigger _Trigger, err error) {
	if d.Config.EventHandler == nil {
		return
	}
	event := Event{
		Request:        req,
		ResponseWriter: resp,
		Status:         http.StatusInternalServerError,
		Message:        err,
		Trigger:        trigger,
	}

	if d.Config.EventTrace {
		goRoot := runtime.GOROOT()
		for skip := 0; ; skip++ {
			_, file, line, ok := runtime.Caller(skip)
			// 排除trace中的标准包信息
			if !strings.HasPrefix(file, goRoot) {
				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line))
			}
			if !ok {
				break
			}
		}
	}

	d.Config.EventHandler(&event)
}

// 触发404事件
func (d *App) eventNotFound(resp http.ResponseWriter, req *http.Request) {
	if d.Config.EventHandler == nil {
		return
	}
	d.Config.EventHandler(&Event{
		Status:         http.StatusNotFound,
		Message:        errors.New(http.StatusText(http.StatusNotFound)),
		Request:        req,
		ResponseWriter: resp,
	})
}

// 触发405事件
func (d *App) eventMethodNotAllowed(resp http.ResponseWriter, req *http.Request) {
	if d.Config.EventHandler == nil {
		return
	}
	d.Config.EventHandler(&Event{
		Status:         http.StatusMethodNotAllowed,
		Message:        errors.New(http.StatusText(http.StatusMethodNotAllowed)),
		Request:        req,
		ResponseWriter: resp,
	})
}
