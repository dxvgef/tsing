package tsing

import (
	"errors"
	"net/http"
	"sync"
)

const defaultMultipartMemory = 32 << 20 // 32 MB

type AppConfig struct {
	RedirectTrailingSlash  bool         // 重定向到类似的路由
	FixPath                bool         // 修复URL中的路径，并重定向
	HandleMethodNotAllowed bool         // 处理405事件
	ProxyRemoteIP          bool         // 识别nginx/haproxy反向代理后的客户端IP
	UseRawPath             bool         // 使用url.RawPath查找参数
	UnescapePathValues     bool         // 解码路径中的参数
	MaxMultipartMemory     int64        // http.Request的ParseMultipartForm属性值
	Recovery               bool         // 处理器自动recover，并抛出500事件(影响性能)
	EventHandler           EventHandler // 事件处理器
	EventFuncName          bool         // 事件中启用函数名
	EventTrace             bool         // 事件中启用跟踪
	EventFile              bool         // 事件中启用源码文件
	EventLine              bool         // 事件中启用源码行号
}

// 应用
type App struct {
	Router      RouterGroup  // 路由器
	Config      AppConfig    // 配置
	contextPool sync.Pool    // context池
	methodTrees _MethodTrees // 路由树
}

// 返回一个空实例
func New() *App {
	var config AppConfig
	config.HandleMethodNotAllowed = true
	config.Recovery = false
	config.MaxMultipartMemory = defaultMultipartMemory
	app := &App{
		Router: RouterGroup{
			handlers: nil,
			basePath: "/",
			root:     true,
		},
		Config:      config,
		methodTrees: make(_MethodTrees, 0, 9), // 创建一个方法集合，数量为0，预留9个位置
	}
	app.Router.app = app

	// 定义context池
	app.contextPool.New = func() interface{} {
		return &Context{app: app}
	}

	if app.Config.Recovery {
		app.Router.Use(func(ctx *Context) error {
			defer func() {
				if err := recover(); err != nil {
					// 抛出500事件
					if app.Config.EventHandler != nil {
						app.Config.EventHandler(ctx, Event{
							Status:  http.StatusInternalServerError,
							Message: errors.New(http.StatusText(http.StatusInternalServerError)),
						})
					}
				}
			}()
			ctx.Next()
			return nil
		})
	}
	return app
}

// 实现http.Handler接口
func (app *App) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// 当有一个新连接时，从context池里取出一个对象
	ctx := app.contextPool.Get().(*Context)
	ctx.responseWriter.reset(resp)
	ctx.Request = req
	ctx.reset()

	app.handleHTTPRequest(ctx)

	app.contextPool.Put(ctx)
}

func (app *App) handleHTTPRequest(ctx *Context) {

	httpMethod := ctx.Request.Method
	rPath := ctx.Request.URL.Path
	unescape := false
	if app.Config.UseRawPath && len(ctx.Request.URL.RawPath) > 0 {
		rPath = ctx.Request.URL.RawPath
		unescape = app.Config.UnescapePathValues
	}
	rPath = cleanPath(rPath)

	// Find root of the tree for the given HTTP method
	t := app.methodTrees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root
		// Find route in tree
		value := root.getValue(rPath, ctx.params, unescape)
		if value.handlers != nil {
			ctx.handlers = value.handlers
			ctx.params = value.params
			ctx.fullPath = value.fullPath
			ctx.Next()
			ctx.responseWriter.OverWriteHeader()
			return
		}
		if httpMethod != "CONNECT" && rPath != "/" {
			if value.tsr && app.Config.RedirectTrailingSlash {
				redirectTrailingSlash(ctx)
				return
			}
			if app.Config.FixPath && fixPath(ctx, root, app.Config.FixPath) {
				return
			}
		}
		break
	}

	if app.Config.HandleMethodNotAllowed {
		for _, tree := range app.methodTrees {
			if tree.method == httpMethod {
				continue
			}
			if value := tree.root.getValue(rPath, nil, unescape); value.handlers != nil {
				// 抛出405事件
				if app.Config.EventHandler != nil {
					app.Config.EventHandler(ctx, Event{
						Status:  http.StatusMethodNotAllowed,
						Message: errors.New(http.StatusText(http.StatusMethodNotAllowed)),
					})
				}
				return
			}
		}
	}
	// 抛出404事件
	if app.Config.EventHandler != nil {
		app.Config.EventHandler(ctx, Event{
			Status:  http.StatusNotFound,
			Message: errors.New(http.StatusText(http.StatusNotFound)),
		})
	}
}
