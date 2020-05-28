package tsing

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// 默认body限制
const MaxMultipartMemory = 1 << 20

// 上下文
type Context struct {
	PathParams     PathParams
	handlers       HandlersChain
	ResponseWriter http.ResponseWriter
	fullPath       string
	engine         *Engine
	Request        *http.Request
	index          int8
	parsed         bool // 是否已解析body
}

var emptyValues url.Values

// 重置Context
func (ctx *Context) reset(req *http.Request, resp http.ResponseWriter) {
	ctx.Request = req
	ctx.ResponseWriter = resp
	ctx.PathParams = ctx.PathParams[0:0]
	ctx.handlers = nil
	ctx.index = -1
	ctx.fullPath = ""
	ctx.parsed = false
}

// 解析form数据
func (ctx *Context) parseForm() error {
	if ctx.parsed {
		return nil
	}
	if strings.HasPrefix(ctx.Request.Header.Get("Content-Type"), "multipart/form-data") {
		if err := ctx.Request.ParseMultipartForm(ctx.engine.Config.MaxMultipartMemory); err != nil {
			return err
		}
	} else {
		if err := ctx.Request.ParseForm(); err != nil {
			return err
		}
	}
	ctx.parsed = true
	return nil
}

// 继续执行下一个处理器
func (ctx *Context) next() {
	var err error
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		// 执行处理器
		err = ctx.handlers[ctx.index](ctx)
		if err == nil {
			ctx.index++
			continue
		}
		// 500事件
		if ctx.engine.Config.EventHandler == nil || !ctx.engine.Config.EventHandlerError {
			break
		}
		if !ctx.engine.Config.EventSource {
			ctx.engine.handlerErrorEvent(ctx.ResponseWriter, ctx.Request, nil, err)
			break
		}
		source := getFuncInfo(ctx.handlers[ctx.index])
		if ctx.engine.Config.EventShortPath {
			source.File = strings.TrimPrefix(source.File, ctx.engine.Config.RootPath)
		}
		ctx.engine.handlerErrorEvent(ctx.ResponseWriter, ctx.Request, source, err)
		break
	}
}

// 在处理器函数内return一个error时，用Caller可以记录下处理器内的行号详细信息
func (ctx *Context) Caller(err error) error {
	if err == nil {
		return nil
	}
	// 使用contextSourceHandler来触发事件
	ctx.engine.contextSourceHandler(ctx.ResponseWriter, ctx.Request, err)
	ctx.Abort()
	// 清空错误，防止引擎再使用handlerErrorEvent()触发一次重复事件
	return nil
}

// 中止执行
func (ctx *Context) Abort() {
	ctx.index = abortIndex
}

// 是否已中止
func (ctx *Context) IsAborted() bool {
	return ctx.index >= abortIndex
}

// 在Context中写值
func (ctx *Context) SetValue(key, value interface{}) {
	if key == nil {
		return
	}
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), key, value))
}

// 从Context中取值
func (ctx *Context) GetValue(key interface{}) interface{} {
	if key == nil {
		return nil
	}
	return ctx.Request.Context().Value(key)
}

// 向客户端发送重定向响应
func (ctx *Context) Redirect(code int, url string) {
	if code < 300 || code > 308 {
		ctx.engine.panicEvent(ctx.ResponseWriter, ctx.Request, "The status code can only be 30x")
		return
	}
	ctx.ResponseWriter.Header().Set("Location", url)
	ctx.ResponseWriter.WriteHeader(code)
}

// 获得客户端真实IP
func (ctx *Context) RemoteIP() string {
	ra := ctx.Request.RemoteAddr
	if ip := ctx.Request.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := ctx.Request.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		var err error
		ra, _, err = net.SplitHostPort(ra)
		if err != nil {
			return ""
		}
	}
	return ra
}

// 获取所有GET参数值
func (ctx *Context) QueryParams() url.Values {
	return ctx.Request.URL.Query()
}

// 获取某个GET参数值的string类型
func (ctx *Context) Query(key string) string {
	if len(ctx.Request.URL.Query()[key]) == 0 {
		return ""
	}
	return ctx.Request.URL.Query()[key][0]
}

// 获取某个GET参数
func (ctx *Context) QueryParam(key string) (string, bool) {
	if len(ctx.Request.URL.Query()[key]) == 0 {
		return "", false
	}
	return ctx.Request.URL.Query()[key][0], true
}

// 获取所有POST/PATCH/PUT参数值
func (ctx *Context) PostParams() url.Values {
	if err := ctx.parseForm(); err != nil {
		return emptyValues
	}
	return ctx.Request.PostForm
}

// 获取某个POST/PATCH/PUT参数值的string类型
func (ctx *Context) Post(key string) string {
	if err := ctx.parseForm(); err != nil {
		return ""
	}
	vs := ctx.Request.PostForm[key]
	if len(vs) == 0 {
		return ""
	}
	return ctx.Request.PostForm[key][0]
}

// 获取某个POST/PATCH/PUT参数
func (ctx *Context) PostParam(key string) (string, bool) {
	if err := ctx.parseForm(); err != nil {
		return "", false
	}
	vs := ctx.Request.PostForm[key]
	if len(vs) == 0 {
		return "", false
	}
	return ctx.Request.PostForm[key][0], true
}

// 获取所有GET/POST/PUT参数值
func (ctx *Context) FormParams() url.Values {
	if err := ctx.parseForm(); err != nil {
		return emptyValues
	}
	return ctx.Request.Form
}

// 获取某个GET/POST/PUT参数值的string类型
func (ctx *Context) Form(key string) string {
	if err := ctx.parseForm(); err != nil {
		return ""
	}
	vs := ctx.Request.Form[key]
	if len(vs) == 0 {
		return ""
	}
	return ctx.Request.Form[key][0]
}

// 获取单个GET/POST/PUT参数
func (ctx *Context) FormParam(key string) (string, bool) {
	if err := ctx.parseForm(); err != nil {
		return "", false
	}
	vs := ctx.Request.Form[key]
	if len(vs) == 0 {
		return "", false
	}
	return ctx.Request.Form[key][0], true
}

// 将body里的json数据反序列化到传入的对象
func (ctx *Context) UnmarshalJSON(obj interface{}) error {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, obj)
}
