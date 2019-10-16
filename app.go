package tsing

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// 框架实例
type App struct {
	Recovery              bool // 启用自动恢复
	RedirectTrailingSlash bool // 自动重定向到结尾带有或不带有斜杠的URL
	FixPath               bool // 清理URL中的不规范路径分隔符，例如//和../
	HandleOPTIONS         bool // 自定响应OPTIONS请求
	// 事件配置
	Event struct {
		Handler     EventHandler // 事件处理器，如果未传值则认为不记录事件
		EnableTrace bool         // 启用500事件的跟踪(影响性能)
		ShortCaller bool         // 缩短事件触发的源码文件名(仅记录源码文件名，仅对ctx.Event触发的500事件有效)s
	}
	Router     RouterGroup        // tsing 路由器
	httpRouter *httprouter.Router // httprouter 实例
}

// 返回一个基本配置的框架实例
func New() *App {
	var app App
	app.httpRouter = httprouter.New()
	app.RedirectTrailingSlash = false
	app.httpRouter.RedirectTrailingSlash = false
	app.FixPath = app.httpRouter.RedirectFixedPath
	app.HandleOPTIONS = app.httpRouter.HandleOPTIONS
	app.httpRouter.PanicHandler = app.event500
	app.httpRouter.NotFound = http.HandlerFunc(app.event404)
	app.httpRouter.MethodNotAllowed = http.HandlerFunc(app.event405)
	app.Router = RouterGroup{
		app: &app,
	}
	return &app
}

// 为实现http.Handler的方法，也是连接的入口处理方法
func (d *App) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	d.httpRouter.ServeHTTP(resp, req)
}
