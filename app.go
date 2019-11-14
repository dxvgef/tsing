package tsing

import (
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// 框架配置
type AppConfig struct {
	EventHandler EventHandler // 事件处理器，如果未传值则认为不记录事件
	// Recovery              bool         // 启用自动恢复
	RedirectTrailingSlash bool // 自动重定向到结尾带有或不带有斜杠的URL
	FixPath               bool // 清理URL中的不规范路径分隔符，例如//和../
	HandleOPTIONS         bool // 自定响应OPTIONS请求
	EventTrigger          bool // 记录事件的触发信息(影响性能)
	EventTrace            bool // 启用事件的跟踪信息(影响性能)
	EventShortPath        bool // 使用事件触发信息的短文件名
}

// 框架实例
type App struct {
	Config      AppConfig          // 配置
	Router      *RouterGroup       // 路由器
	httpRouter  *httprouter.Router // httprouter 实例
	contextPool sync.Pool          // context池
}

// 返回一个基本配置的框架实例
func New() *App {
	var config AppConfig
	config.FixPath = true // 自动修复URL路径
	// config.Recovery = true
	var app App
	app.httpRouter = httprouter.New() // 创建httprouter实例
	app.httpRouter.RedirectTrailingSlash = config.RedirectTrailingSlash
	app.httpRouter.RedirectFixedPath = config.FixPath
	app.httpRouter.HandleOPTIONS = config.HandleOPTIONS
	app.httpRouter.NotFound = http.HandlerFunc(app.eventNotFound)                 // 404事件处理器
	app.httpRouter.HandleMethodNotAllowed = true                                  // 处理405事件
	app.httpRouter.MethodNotAllowed = http.HandlerFunc(app.eventMethodNotAllowed) // 405事件处理器
	app.Router = &RouterGroup{
		app: &app,
	}
	// 定义context池
	app.contextPool.New = func() interface{} {
		return &Context{app: &app}
	}
	return &app
}

// 启用panic处理器，默认未启用
// panic处理器会使框架抛出panic事件，并自动恢复，防止进程退出
// 注意，启用此功能会使性能明显下降
func (d *App) EnablePanicHandler() {
	d.httpRouter.PanicHandler = d.eventPanic
}

// 禁用panic处理器，默认禁用
func (d *App) DisablePanicHandler() {
	d.httpRouter.PanicHandler = nil
}

// 实现http.Handler接口
func (app *App) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// 使用httprouter处理请求
	app.httpRouter.ServeHTTP(resp, req)
}
