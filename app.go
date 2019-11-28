package tsing

import (
	"net/http"
	"os"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// 框架配置
type Config struct {
	RootPath              string // 项目根路径，用于缩短路径
	RedirectTrailingSlash bool   // 自动重定向到结尾带有或不带有斜杠的URL
	HandleOPTIONS         bool   // 自定响应OPTIONS请求
	FixPath               bool   // 修复路径
	Recover               bool   // 自动恢复panic

	EventHandler          EventHandler // 事件处理器，如果未传值则不触发任何事件
	ErrorEvent            bool         // 启用处理器返回的错误事件
	NotFoundEvent         bool         // 启用404错误事件
	MethodNotAllowedEvent bool         // 启用405错误事件
	Trigger               bool         // 启用事件的触发信息(影响性能)
	Trace                 bool         // 启用事件的跟踪信息(影响性能)
	ShortPath             bool         // 启用事件的短文件路径，自动去除{RootPath}
}

// 框架实例
type App struct {
	Config      *Config            // 配置
	Router      *RouterGroup       // 路由器
	httpRouter  *httprouter.Router // httprouter 实例
	contextPool sync.Pool          // context池
}

// 返回一个指定配置的框架实例
func New(config *Config) *App {
	var app App
	app.Config = config
	// 创建httprouter实例并设置参数
	app.httpRouter = httprouter.New()
	app.httpRouter.HandleOPTIONS = config.HandleOPTIONS
	app.httpRouter.RedirectTrailingSlash = config.RedirectTrailingSlash
	app.httpRouter.RedirectFixedPath = config.FixPath

	// 如果注册了事件处理器
	if config.EventHandler != nil {
		// 启用panic恢复处理，默认未启用
		// 注意，启用此功能会使性能明显下降
		if config.Recover {
			app.httpRouter.PanicHandler = app.panicHandler
		}

		// 启用404事件
		if config.NotFoundEvent {
			app.httpRouter.NotFound = http.HandlerFunc(app.notFoundHandler)
		}

		// 启用405事件
		if config.MethodNotAllowedEvent {
			app.httpRouter.HandleMethodNotAllowed = true
			app.httpRouter.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedHandler)
		}
	}

	// 如果启用了事件短路径
	if config.ShortPath && config.RootPath == "" {
		// 使用当前路径
		path, err := os.Getwd()
		if err == nil {
			config.RootPath = path
		}
	}

	// 创建默认的根路由组
	app.Router = &RouterGroup{
		app: &app,
	}

	// 创建context池
	app.contextPool.New = func() interface{} {
		return &Context{app: &app}
	}
	return &app
}

// 实现http.Handler接口
func (app *App) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// 使用httprouter处理请求
	app.httpRouter.ServeHTTP(resp, req)
}
