package tsing

import (
	"errors"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

// 事件结构
type Event struct {
	Status         int      // HTTP状态码
	Message        error    // 消息文本
	Source         *_Source // 触发信息
	Trace          []string // 跟踪信息
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

// 触发信息
type _Source struct {
	Func string // 函数名
	File string // 文件名
	Line int    // 行号
}

// 事件处理器类型
type EventHandler func(*Event)

// 获得函数信息
// func getFuncInfo(obj interface{}) *_Source {
// 	ptr := reflect.ValueOf(obj).Pointer()
// 	file, line := runtime.FuncForPC(ptr).FileLine(ptr)
// 	return &_Source{
// 		Func: runtime.FuncForPC(ptr).Name(),
// 		File: file,
// 		Line: line,
// 	}
// }

func (e *Event) reset(resp http.ResponseWriter, req *http.Request) {
	e.Message = nil
	e.Source = nil
	e.Status = 0
	e.Trace = []string{}
	e.ResponseWriter = resp
	e.Request = req
}

// 路由自动执行handler函数得到错误的处理器
// func (engine *Engine) funcErrorHandler(resp http.ResponseWriter, req *http.Request, trigger *_Source, err error) {
// 	if err == nil {
// 		return
// 	}
//
// 	event := Event{
// 		Request:        req,
// 		ResponseWriter: resp,
// 		Status:         500,
// 		Message:        err,
// 		Source:         trigger, // 这里不用判断是否开启triger，如果没开启会传入nil
// 	}
//
// 	// 如果开启了trace
// 	if engine.Config.EventTrace {
// 		goRoot := filepath.Clean(runtime.GOROOT())
// 		for skip := 0; ; skip++ {
// 			funcPtr, file, line, ok := runtime.Caller(skip)
// 			if !ok {
// 				break
// 			}
// 			// 使用短路径
// 			if engine.Config.EventTraceShortPath {
// 				file = strings.TrimPrefix(filepath.Clean(file), filepath.Clean(engine.Config.RootPath))
// 			}
// 			// 排除trace中的标准包信息
// 			if !strings.HasPrefix(file, goRoot) {
// 				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line)+":"+runtime.FuncForPC(funcPtr).Name())
// 			}
// 		}
// 	}
// 	engine.Config.EventHandler(&event)
// }

// context的Error()触发的错误处理器
// func (engine *Engine) contextErrorHandler(ctx *Context, err error) {
// 	if err == nil || engine.Config.EventHandler == nil || !engine.Config.ErrorEvent {
// 		return
// 	}
//
// 	event := Event{
// 		Request:        ctx.Request,
// 		ResponseWriter: ctx.ResponseWriter,
// 		Status:         500,
// 		Message:        err,
// 		Source:         nil,
// 	}
//
// 	// 如果启用了trigger
// 	if ctx.engine.Config.EventTrigger {
// 		if funcPtr, file, line, ok := runtime.Caller(2); ok {
// 			// 使用短路径
// 			if ctx.engine.Config.EventTraceShortPath {
// 				file = strings.TrimPrefix(file, ctx.engine.Config.RootPath)
// 			}
// 			if event.Source != nil {
// 				event.Source.File = file
// 				event.Source.Line = line
// 				event.Source.Func = runtime.FuncForPC(funcPtr).Name()
// 			} else {
// 				var trigger _Source
// 				trigger.File = file
// 				trigger.Line = line
// 				trigger.Func = runtime.FuncForPC(funcPtr).Name()
// 				event.Source = &trigger
// 			}
// 		}
// 	}
//
// 	// 如果开启了trace
// 	if ctx.engine.Config.EventTrace {
// 		goRoot := runtime.GOROOT()
// 		for skip := 0; ; skip++ {
// 			funcPtr, file, line, ok := runtime.Caller(skip)
// 			if !ok {
// 				break
// 			}
// 			// 使用短路径
// 			if ctx.engine.Config.EventTraceShortPath {
// 				file = strings.TrimPrefix(file, ctx.engine.Config.RootPath)
// 			}
// 			// 排除trace中的标准包信息
// 			if !strings.HasPrefix(file, goRoot) {
// 				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line)+":"+runtime.FuncForPC(funcPtr).Name())
// 			}
// 		}
// 	}
//
// 	ctx.engine.Config.EventHandler(&event)
// }

// handler的panic处理器
func (engine *Engine) panicEvent(resp http.ResponseWriter, req *http.Request, err interface{}) {
	if !engine.Config.Recover && engine.Config.EventHandler == nil {
		return
	}

	// 从池中取出一个ctx
	event := engine.eventPool.Get().(*Event)
	event.reset(resp, req)

	event.Status = 500

	switch t := err.(type) {
	case string:
		event.Message = errors.New(t)
	case error:
		event.Message = t
	default:
		event.Message = errors.New("未知错误消息类型")
	}

	// 如果启用事件的触发信息
	if engine.Config.EventTrigger {
		funcPtr, file, line, ok := runtime.Caller(3)
		if ok {
			// 缩短文件路径
			if engine.Config.EventTraceShortPath {
				file = strings.TrimPrefix(file, engine.Config.RootPath)
			}
			if event.Source != nil {
				event.Source.File = file
				event.Source.Line = line
				event.Source.Func = runtime.FuncForPC(funcPtr).Name()
			} else {
				var trigger _Source
				trigger.File = file
				trigger.Line = line
				trigger.Func = runtime.FuncForPC(funcPtr).Name()
				event.Source = &trigger
			}
		}
	}

	// 如果启用事件的跟踪信息
	if engine.Config.EventTrace {
		goRoot := runtime.GOROOT()
		for skip := 0; ; skip++ {
			_, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			// 缩短路径
			if engine.Config.EventTraceShortPath {
				file = strings.TrimPrefix(file, engine.Config.RootPath)
			}
			// 排除trace中的标准包信息
			if !engine.Config.EventTraceOnlyProject && strings.HasPrefix(file, goRoot) {
				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line))
			}
			if !strings.HasPrefix(file, goRoot) {
				event.Trace = append(event.Trace, file+":"+strconv.Itoa(line))
			}
		}
	}

	engine.Config.EventHandler(event)

	// 将event放回池中
	engine.eventPool.Put(event)
}

// 404事件处理器
func (engine *Engine) notFoundEvent(resp http.ResponseWriter, req *http.Request) {
	if engine.Config.EventHandler == nil {
		return
	}

	// 从池中取出一个ctx
	event := engine.eventPool.Get().(*Event)
	event.reset(resp, req)

	event.Status = http.StatusNotFound
	event.Message = errors.New(http.StatusText(http.StatusNotFound))

	engine.Config.EventHandler(event)

	engine.eventPool.Put(event)
}

// 405事件处理器
func (engine *Engine) methodNotAllowedEvent(resp http.ResponseWriter, req *http.Request) {
	if engine.Config.EventHandler == nil {
		return
	}

	// 从池中取出一个ctx
	event := engine.eventPool.Get().(*Event)
	event.reset(resp, req)

	event.Status = http.StatusMethodNotAllowed
	event.Message = errors.New(http.StatusText(http.StatusMethodNotAllowed))

	engine.Config.EventHandler(event)

	engine.eventPool.Put(event)
}
