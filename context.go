package tsing

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Context 上下文
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	routerParams   httprouter.Params // 路由参数
	Next           bool              // 继续执行下一个中间件或处理器
	app            *App
	parsed         bool // 是否已解析body
}

// Continue 继续执行下一个中间件或处理器
func (ctx Context) Continue() (Context, error) {
	ctx.Next = true
	return ctx, nil
}

// Break 中断，不继续执行下一个中间件或处理器，如果err不为nil，则同时抛出500事件
func (ctx Context) Break(err error) (Context, error) {
	ctx.Next = false
	return ctx, err
}

// Event 触发500事件，使用此方法是为了精准记录触发事件的源码文件及行号
func (ctx Context) Event(err error) error {
	if err != nil && ctx.app.Event.Handler != nil {
		event := Event{
			Status:         500,
			Message:        err,
			ResponseWriter: ctx.ResponseWriter,
			Request:        ctx.Request,
		}
		if ctx.app.Event.EnableTrace == true {
			_, file, line, _ := runtime.Caller(1)
			l := strconv.Itoa(line)
			if ctx.app.Event.ShortCaller == true {
				short := file
				fileLen := len(file)
				for i := fileLen - 1; i > 0; i-- {
					if file[i] == '/' {
						short = file[i+1:]
						break
					}
				}
				file = short
			}
			event.Trace = append(event.Trace, file+":"+l)
		}
		ctx.app.Event.Handler(event)
	}
	// 不再将传入的error返回，避免再触发handle500函数
	return nil
}

// SetContextValue 在ctx里存储值，如果key存在则替换值
func (ctx *Context) SetContextValue(key string, value interface{}) {
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), key, value))
}

// ContextValue 获取ctx里的值，取出后根据写入的类型自行断言
func (ctx Context) ContextValue(key string) interface{} {
	return ctx.Request.Context().Value(key)
}

// Redirect 重定向
func (ctx Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return errors.New("状态码只能是300-308之间的值")
	}
	ctx.ResponseWriter.Header().Set("Location", url)
	ctx.ResponseWriter.WriteHeader(code)
	return nil
}

// RealIP 获得客户端真实IP
func (ctx Context) RealIP() string {
	ra := ctx.Request.RemoteAddr
	if ip := ctx.Request.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := ctx.Request.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// 解析body数据
func (ctx Context) parseBody() error {
	// 判断是否已经解析过body
	if ctx.parsed == true {
		return nil
	}
	if strings.HasPrefix(ctx.Request.Header.Get("Content-Type"), "multipart/form-data") {
		if err := ctx.Request.ParseMultipartForm(http.DefaultMaxHeaderBytes); err != nil {
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

// RouteValues 获取所有路由参数值
func (ctx Context) RouteValues() []httprouter.Param {
	return ctx.routerParams
}

// RouteValueStrict 获取路由参数值
func (ctx Context) RouteValueStrict(key string) (string, error) {
	for i := range ctx.routerParams {
		if ctx.routerParams[i].Key == key {
			return ctx.routerParams[i].Value, nil
		}
	}
	return "", errors.New("路由参数" + key + "不存在")
}

// RouteValue 获取某个路由参数值的string类型
func (ctx Context) RouteValue(key string) string {
	return ctx.routerParams.ByName(key)
}

// QueryValues 获取所有GET参数值
func (ctx Context) QueryValues() url.Values {
	return ctx.Request.URL.Query()
}

// QueryValueStrict 获取某个GET参数值
func (ctx Context) QueryValueStrict(key string) (string, error) {
	if len(ctx.Request.URL.Query()[key]) == 0 {
		return "", errors.New("GET参数" + key + "不存在")
	}
	return ctx.Request.URL.Query()[key][0], nil
}

// QueryValue 获取某个GET参数值的string类型
func (ctx Context) QueryValue(key string) string {
	if len(ctx.Request.URL.Query()[key]) == 0 {
		return ""
	}
	return ctx.Request.URL.Query()[key][0]
}

// PostValues 获取所有POST参数值
func (ctx Context) PostValues() url.Values {
	err := ctx.parseBody()
	if err != nil {
		return url.Values{}
	}
	return ctx.Request.PostForm
}

// FormValueStrict 获取某个POST参数值
func (ctx Context) FormValueStrict(key string) (string, error) {
	err := ctx.parseBody()
	if err != nil {
		return "", err
	}
	vs := ctx.Request.Form[key]
	if len(vs) == 0 {
		return "", errors.New(ctx.Request.Method + "参数" + key + "不存在")
	}
	return ctx.Request.Form[key][0], nil
}

// FormValue 获取某个POST参数值的string类型
func (ctx Context) FormValue(key string) string {
	err := ctx.parseBody()
	if err != nil {
		return ""
	}
	vs := ctx.Request.Form[key]
	if len(vs) == 0 {
		return ""
	}
	return ctx.Request.Form[key][0]
}
