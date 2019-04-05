package tsing

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Dispatcher 调度器结构
type Dispatcher struct {
	// 事件配置
	Event struct {
		Handler     EventHandler // 事件处理器，如果未传值则认为不记录事件
		EnableTrace bool         // 启用500事件的跟踪(影响性能)
		ShortCaller bool         // 缩短事件触发的源码文件名(仅记录源码文件名，仅对ctx.Return触发的500事件有效)s
	}
	// 路由器
	Router RouterGroup
	// httprouter
	httpRouter *httprouter.Router
}

// New 返回一个调度器实例
func New() *Dispatcher {
	var dispatcher Dispatcher
	dispatcher.httpRouter = httprouter.New()
	dispatcher.httpRouter.PanicHandler = dispatcher.handle500
	dispatcher.httpRouter.NotFound = http.HandlerFunc(dispatcher.handle404)
	dispatcher.httpRouter.MethodNotAllowed = http.HandlerFunc(dispatcher.handle405)
	dispatcher.Router = RouterGroup{
		dispatcher: &dispatcher,
	}
	return &dispatcher
}

func (d *Dispatcher) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	d.httpRouter.ServeHTTP(resp, req)
}
