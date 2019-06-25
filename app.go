package tsing

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// App 框架实例
type App struct {
	// 事件配置
	Event struct {
		Handler     EventHandler // 事件处理器，如果未传值则认为不记录事件
		EnableTrace bool         // 启用500事件的跟踪(影响性能)
		ShortCaller bool         // 缩短事件触发的源码文件名(仅记录源码文件名，仅对ctx.Event触发的500事件有效)s
	}
	// 路由器
	Router RouterGroup
	// httprouter
	httpRouter *httprouter.Router
}

// New 返回一个框架实例
func New() *App {
	var app App
	app.httpRouter = httprouter.New()
	app.httpRouter.PanicHandler = app.event500
	app.httpRouter.NotFound = http.HandlerFunc(app.event404)
	app.httpRouter.MethodNotAllowed = http.HandlerFunc(app.event405)
	app.Router = RouterGroup{
		app: &app,
	}
	return &app
}

func (d *App) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	d.httpRouter.ServeHTTP(resp, req)
}
