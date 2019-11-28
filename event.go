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
	Status         int       // HTTP状态码
	Message        error     // 消息文本
	Trigger        *_Trigger // 触发信息
	Trace          []string  // 跟踪信息
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
func getFuncInfo(obj interface{}) *_Trigger {
	ptr := reflect.ValueOf(obj).Pointer()
	file, line := runtime.FuncForPC(ptr).FileLine(ptr)
	return &_Trigger{
		Func: runtime.FuncForPC(ptr).Name(),
		File: file,
		Line: line,
	}
}

// 路由自动执行handler函数得到错误的处理器
func (app *App) funcErrorHandler(resp http.ResponseWriter, req *http.Request, trigger *_Trigger, err error) {
	if err == nil {
		return
	}

	event := Event{
		Request:        req,
		ResponseWriter: resp,
		Status:         500,
		Message:        err,
		Trigger:        trigger,
	}

	// 如果开启了trace
	if app.Config.Trace {
		goRoot := runtime.GOROOT()
		for skip := 0; ; skip++ {
			funcPtr, file, line, ok := runtime.Caller(skip)
			// 使用短路径
			if app.Config.ShortPath {
				event.Trigger.File = strings.TrimPrefix(file, app.Config.RootPath)
			} else {
				event.Trigger.File = file
			}
			// 排除trace中的标准包信息
			if !strings.HasPrefix(file, goRoot) {
				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line)+":"+runtime.FuncForPC(funcPtr).Name())
			}

			if !ok {
				break
			}
		}
	}
	app.Config.EventHandler(&event)
}

// context的Error()触发的错误处理器
func (app *App) contextErrorHandler(ctx *Context, err error) {
	if err == nil || app.Config.EventHandler == nil || !app.Config.ErrorEvent {
		return
	}

	event := Event{
		Request:        ctx.Request,
		ResponseWriter: ctx.ResponseWriter,
		Status:         500,
		Message:        err,
		Trigger:        nil,
	}

	// 如果启用了trigger
	if ctx.app.Config.Trigger {
		funcPtr, file, line, ok := runtime.Caller(2)
		if ok {
			var trigger _Trigger
			// 使用短路径
			if ctx.app.Config.ShortPath {
				trigger.File = strings.TrimPrefix(file, ctx.app.Config.RootPath)
			} else {
				trigger.File = file
			}
			trigger.Line = line
			trigger.Func = runtime.FuncForPC(funcPtr).Name()
			event.Trigger = &trigger
		}
	}

	// 如果开启了trace
	if ctx.app.Config.Trace {
		goRoot := runtime.GOROOT()
		for skip := 0; ; skip++ {
			funcPtr, file, line, ok := runtime.Caller(skip)
			// 使用短路径
			if ctx.app.Config.ShortPath {
				event.Trigger.File = strings.TrimPrefix(file, ctx.app.Config.RootPath)
			} else {
				event.Trigger.File = file
			}
			// 排除trace中的标准包信息
			if !strings.HasPrefix(file, goRoot) {
				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line)+":"+runtime.FuncForPC(funcPtr).Name())
			}

			if !ok {
				break
			}
		}
	}
	ctx.app.Config.EventHandler(&event)
}

// handler的panic处理器
func (app *App) panicHandler(resp http.ResponseWriter, req *http.Request, err interface{}) {
	if !app.Config.Recover && app.Config.EventHandler == nil {
		return
	}

	event := Event{
		Request:        req,
		ResponseWriter: resp,
		Status:         500,
		Trigger:        nil,
	}

	switch t := err.(type) {
	case string:
		event.Message = errors.New(t)
	case error:
		event.Message = t
	default:
		event.Message = errors.New("未知错误消息类型")
	}

	// 如果启用事件的触发信息
	if app.Config.Trigger {
		funcPtr, file, line, ok := runtime.Caller(3)
		if ok {
			var trigger _Trigger
			if app.Config.ShortPath {
				file = strings.TrimPrefix(file, app.Config.RootPath)
			}
			trigger.File = file
			trigger.Line = line
			trigger.Func = runtime.FuncForPC(funcPtr).Name()
			event.Trigger = &trigger
		}
	}

	// 如果启用事件的跟踪信息
	if app.Config.Trace {
		goRoot := runtime.GOROOT()
		for skip := 0; ; skip++ {
			funcPtr, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			if app.Config.ShortPath {
				file = strings.TrimPrefix(file, app.Config.RootPath)
			}
			// 排除trace中的标准包信息
			if !strings.HasPrefix(file, goRoot) {
				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line))
			}
			event.Trigger.File = file
			event.Trigger.Line = line
			event.Trigger.Func = runtime.FuncForPC(funcPtr).Name()
		}
	}

	app.Config.EventHandler(&event)
}

// 404事件处理器
func (d *App) notFoundHandler(resp http.ResponseWriter, req *http.Request) {
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

// 405事件处理器
func (d *App) methodNotAllowedHandler(resp http.ResponseWriter, req *http.Request) {
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
