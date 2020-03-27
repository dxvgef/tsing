package tsing

import (
	"errors"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// 事件
type Event struct {
	Status         int      // HTTP状态码
	Message        error    // 消息(error)
	Source         *_Source // 来源
	Trace          []string // 跟踪
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

// 来源信息
type _Source struct {
	Func string // 函数名
	File string // 文件名
	Line int    // 行号
}

// 事件处理器
type EventHandler func(*Event)

func (e *Event) reset(resp http.ResponseWriter, req *http.Request) {
	e.Message = nil
	e.Source = nil
	e.Status = 0
	e.Trace = nil
	e.ResponseWriter = resp
	e.Request = req
}

// 处理器返回参数事件
func (engine *Engine) handlerErrorEvent(resp http.ResponseWriter, req *http.Request, source *_Source, err error) {
	if err == nil {
		return
	}

	// 从池中取出一个ctx
	event := engine.eventPool.Get().(*Event)
	event.reset(resp, req)

	event.Status = 500
	event.Message = err
	event.Source = source

	if engine.Config.EventTrace {
		for skip := 0; ; skip++ {
			funcPtr, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			// 使用短路径
			if engine.Config.EventShortPath {
				file = strings.TrimPrefix(filepath.Clean(file), filepath.Clean(engine.Config.RootPath))
			}
			event.Trace = append(event.Trace, file+":"+strconv.Itoa(line)+":"+runtime.FuncForPC(funcPtr).Name())
		}
	}

	engine.Config.EventHandler(event)
	engine.eventPool.Put(event)
}

// context的Source()触发的的事件处理器
// 能精准记录事件来源信息
func (engine *Engine) contextSourceHandler(resp http.ResponseWriter, req *http.Request, err error) {
	if err == nil || engine.Config.EventHandler == nil || !engine.Config.EventHandlerError {
		return
	}

	// 从池中取出一个ctx
	event := engine.eventPool.Get().(*Event)
	event.reset(resp, req)

	event.Status = 500
	event.Message = err

	// 如果启用了source
	if engine.engine.Config.EventSource {
		if funcPtr, file, line, ok := runtime.Caller(2); ok {
			// 使用短路径
			if engine.Config.EventShortPath {
				file = strings.TrimPrefix(file, engine.Config.RootPath)
			}
			if event.Source != nil {
				event.Source.File = file
				event.Source.Line = line
				event.Source.Func = runtime.FuncForPC(funcPtr).Name()
			} else {
				var source _Source
				source.File = file
				source.Line = line
				source.Func = runtime.FuncForPC(funcPtr).Name()
				event.Source = &source
			}
		}
	}

	// 如果开启了trace
	if engine.Config.EventTrace {
		for skip := 0; ; skip++ {
			funcPtr, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			// 使用短路径
			if engine.Config.EventShortPath {
				file = strings.TrimPrefix(file, engine.Config.RootPath)
			}
			event.Trace = append(event.Trace, file+":"+strconv.Itoa(line)+":"+runtime.FuncForPC(funcPtr).Name())
		}
	}

	engine.Config.EventHandler(event)
	engine.eventPool.Put(event)
}

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
	if engine.Config.EventSource {
		funcPtr, file, line, ok := runtime.Caller(3)
		if ok {
			// 缩短文件路径
			if engine.Config.EventShortPath {
				file = strings.TrimPrefix(file, engine.Config.RootPath)
			}
			if event.Source != nil {
				event.Source.File = file
				event.Source.Line = line
				event.Source.Func = runtime.FuncForPC(funcPtr).Name()
			} else {
				var source _Source
				source.File = file
				source.Line = line
				source.Func = runtime.FuncForPC(funcPtr).Name()
				event.Source = &source
			}
		}
	}

	// 如果启用事件的跟踪信息
	if engine.Config.EventTrace {
		for skip := 0; ; skip++ {
			_, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			// 缩短路径
			if engine.Config.EventShortPath {
				file = strings.TrimPrefix(file, engine.Config.RootPath)
			}
			event.Trace = append(event.Trace, file+":"+strconv.Itoa(line))
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
