package tsing

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

// 路由参数
type Param struct {
	Key   string
	Value string
}

// 一个URL中路由参数的集合，参数是有序的，可以安全的通过索引来获取参数
type Params []Param

type Context struct {
	responseWriter customResponseWriter
	Request        *http.Request
	ResponseWriter _CustomResponseWriter

	params             Params
	handlers           HandlersChain
	nextHandlerIndex   int8
	fullPath           string
	currentHandlerName string
	app                *App

	parsed bool // 是否已解析body

	// (暂未用到)定义一个用于内容协商的接受格式的列表
	// Accepted []string
}

func (ctx *Context) reset() {
	ctx.ResponseWriter = &ctx.responseWriter
	ctx.params = ctx.params[0:0]
	ctx.handlers = nil
	ctx.nextHandlerIndex = -1
	ctx.fullPath = ""
	// ctx.Accepted = nil
}

// 执行下一个处理器
func (ctx *Context) Next() {
	defer func() {
		if err := recover(); err != nil {
			if ctx.app.Config.EventHandler != nil {
				var event Event
				event.Status = 500
				switch t := err.(type) {
				case string:
					event.Message = errors.New(t)
				case error:
					event.Message = t
				case *net.OpError:
					event.Message = t.Err
				default:
					event.Message = errors.New("处理器出现panic错误")
				}
				ctx.app.Config.EventHandler(ctx, event)
			}
		}
	}()
	ctx.nextHandlerIndex++
	for ctx.nextHandlerIndex < int8(len(ctx.handlers)) {
		// ctx.currentHandlerName = getHandlerName(ctx.handlers[ctx.nextHandlerIndex])
		if err := ctx.handlers[ctx.nextHandlerIndex](ctx); err != nil {
			// 如果处理器返回值不为nil，则抛出500事件
			if ctx.app.Config.EventHandler != nil {
				ctx.app.Config.EventHandler(ctx, Event{
					Status:  http.StatusInternalServerError,
					Message: errors.New(http.StatusText(http.StatusInternalServerError)),
				})
			}
			break
		}
		ctx.nextHandlerIndex++
	}
}

// 获得当前处理器的名称(实际是获取最后一个处理器的名称)
func (ctx *Context) HandlerName() string {
	return getHandlerName(ctx.handlers.last())
	// return ctx.currentHandlerName
}

// 获得当前路由注册的所有处理器名
func (ctx *Context) HandlerNames() []string {
	hn := make([]string, 0, len(ctx.handlers))
	for _, val := range ctx.handlers {
		hn = append(hn, getHandlerName(val))
	}
	return hn
}

// 获得完整路径
func (ctx *Context) FullPath() string {
	return ctx.fullPath
}

// 判断是否停止下一个处理器执行的流程
func (ctx *Context) IsAborted() bool {
	return ctx.nextHandlerIndex >= abortIndex
}

// 停止执行下一个处理器
func (ctx *Context) Abort() {
	ctx.nextHandlerIndex = abortIndex
}

// // 响应状态码
// func (ctx *Context) Status(code int) {
// 	ctx.responseWriter.WriteHeader(code)
// }
//
// // 停止执行下一个处理器并响应状态码
// func (ctx *Context) AbortWithStatus(code int) {
// 	ctx.Status(code)
// 	ctx.ResponseWriter.OverWriteHeader()
// 	ctx.Abort()
// }

// 解析body数据
func (ctx *Context) parseBody() error {
	// 判断是否已经解析过body
	if ctx.parsed == true {
		return nil
	}
	if strings.HasPrefix(ctx.Request.Header.Get("Content-Type"), "multipart/form-data") {
		if err := ctx.Request.ParseMultipartForm(ctx.app.Config.MaxMultipartMemory); err != nil {
			return err
		}
	} else {
		if err := ctx.Request.ParseForm(); err != nil {
			return err
		}
	}
	// 标记该context中的body已经解析过
	ctx.parsed = true
	return nil
}

// 向客户端发送重定向响应
func (ctx *Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return errors.New("状态码只能是300-308之间的值")
	}
	ctx.ResponseWriter.Header().Set("Location", url)
	ctx.ResponseWriter.WriteHeader(code)
	return nil
}

// 写入ctx参数值
func (ctx *Context) SetValue(key string, value interface{}) {
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), key, value))
}

// 获得ctx参数值
func (ctx *Context) GetValue(key string) interface{} {
	return ctx.Request.Context().Value(key)
}

// 获得路由参数
func (ctx *Context) Param(key string) (string, bool) {
	for _, v := range ctx.params {
		if v.Key == key {
			return v.Value, true
		}
	}
	return "", false
}

// 获取所有路由参数值
func (ctx *Context) Params() Params {
	return ctx.params
}

// 获得路收参数string类型，不判断参数是否存在
func (ctx *Context) ParamValue(name string) (value string) {
	value, _ = ctx.Param(name)
	return
}

// 获得客户端IP，支持nginx/haproxy的反向代理
func (ctx *Context) RemoteIP() string {
	if ctx.app.Config.ProxyRemoteIP {
		clientIP := ctx.Request.Header.Get("X-Forwarded-For")
		clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
		if clientIP == "" {
			clientIP = strings.TrimSpace(ctx.Request.Header.Get("X-Real-Ip"))
		}
		if clientIP != "" {
			return clientIP
		}
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(ctx.Request.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// 是否websocket连接
func (ctx *Context) IsWebsocket() bool {
	if strings.Contains(strings.ToLower(ctx.Request.Header.Get("Connection")), "upgrade") &&
		strings.EqualFold(ctx.Request.Header.Get("Upgrade"), "websocket") {
		return true
	}
	return false
}

// 以流的方式发送文件给客户端
func (ctx *Context) Stream(step func(w io.Writer) bool) bool {
	w := ctx.ResponseWriter
	clientGone := w.CloseNotify()
	for {
		select {
		case <-clientGone:
			return true
		default:
			keepOpen := step(w)
			w.Flush()
			if !keepOpen {
				return false
			}
		}
	}
}

// 返回文件给客户端
func (ctx *Context) File(filepath string) {
	http.ServeFile(ctx.ResponseWriter, ctx.Request, filepath)
}
