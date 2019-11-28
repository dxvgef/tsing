package tsing

import (
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// 路由处理器
type Handler func(*Context) error

// 路由组
type RouterGroup struct {
	handlers []Handler // 处理器
	basePath string    // 基路径
	app      *App      // 调度器
}

// GROUP 路由组
func (r *RouterGroup) GROUP(path string, handlers ...Handler) *RouterGroup {
	// 生成一个新的路由组
	group := RouterGroup{
		handlers: r.handlers,        // 初始处理器为上级路由组的处理器
		basePath: r.basePath + path, // 拼接路由组的基本路径
		app:      r.app,
	}
	// 组合上级路由组处理器和当前传入的处理器
	for k := range handlers {
		group.handlers = append(group.handlers, handlers[k])
	}
	return &group
}

// Handle 路由
func (r *RouterGroup) Handle(method string, path string, handler Handler, middlewareHandlers ...Handler) {
	r.app.httpRouter.Handle(method, r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// PATH 定义路由到目录，不支持路由组和中间件
func (r *RouterGroup) PATH(url string, local string, list bool) {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	url += "*filepath"

	// 使用GET方法模拟httprouter.ServeFiles()，防止其内部直接输出404消息给客户端
	r.app.httpRouter.GET(url, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		// 如果请求的是目录，而判断是否允许列出目录
		if params.ByName("filepath") == "" || params.ByName("filepath")[len(params.ByName("filepath"))-1:] == "/" {
			if !list {
				// 如果不允许列出目录，则触发404事件处理
				r.app.notFoundHandler(resp, req)
				return
			}
		}

		// 判断请求的文件是否存在
		file := local + params.ByName("filepath")
		if _, err := os.Stat(file); err != nil {
			r.app.notFoundHandler(resp, req)
			return
		}
		http.ServeFile(resp, req, file)
	})
}

// FILE 定义路由到文件，不支持路由组和中间件
func (r *RouterGroup) FILE(url string, local string) {
	// 使用GET方法模拟httprouter.ServeFiles()，防止其内部直接输出404消息给客户端
	r.app.httpRouter.GET(url, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		if _, err := os.Stat(local); err != nil {
			r.app.notFoundHandler(resp, req)
			return
		}
		http.ServeFile(resp, req, local)
	})
}

// GET 路由
func (r *RouterGroup) GET(path string, handler Handler, middlewareHandlers ...Handler) {
	r.app.httpRouter.GET(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// POST 路由
func (r *RouterGroup) POST(path string, handler Handler, middlewareHandlers ...Handler) {
	r.app.httpRouter.POST(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// PUT 路由
func (r *RouterGroup) PUT(path string, handler Handler, middlewareHandlers ...Handler) {
	r.app.httpRouter.PUT(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// HEAD 路由
func (r *RouterGroup) HEAD(path string, handler Handler, middlewareHandlers ...Handler) {
	r.app.httpRouter.HEAD(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// PATCH 路由
func (r *RouterGroup) PATCH(path string, handler Handler, middlewareHandlers ...Handler) {
	r.app.httpRouter.PATCH(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// DELETE 路由
func (r *RouterGroup) DELETE(path string, handler Handler, middlewareHandlers ...Handler) {
	r.app.httpRouter.DELETE(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// OPTIONS 路由
func (r *RouterGroup) OPTIONS(path string, handler Handler, middlewareHandlers ...Handler) {
	r.app.httpRouter.OPTIONS(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// 执行处理器函数
func (r *RouterGroup) execute(resp http.ResponseWriter, req *http.Request, params httprouter.Params, handler Handler, middlewares []Handler) {
	// 当有一个新连接时，从context池里取出一个对象
	ctx := r.app.contextPool.Get().(*Context)
	// 重置ctx
	ctx.app = r.app
	ctx.ResponseWriter = resp
	ctx.Request = req
	ctx.next = false
	ctx.parsed = false
	ctx.routeParams = params

	var err error
	var trigger *_Trigger

	// 遍历执行路由组的处理器
	for k := range r.handlers {
		if err = r.handlers[k](ctx); err != nil {
			// 生成事件触发信息
			if r.app.Config.EventHandler != nil && r.app.Config.ErrorEvent && r.app.Config.Trigger {
				trigger = getFuncInfo(r.handlers[k])
				if r.app.Config.ShortPath {
					trigger.File = strings.TrimPrefix(trigger.File, r.app.Config.RootPath)
				}
				r.app.funcErrorHandler(resp, req, trigger, err)
			}
			// 将ctx放回池中
			r.app.contextPool.Put(ctx)
			return
		}
		if !ctx.next {
			// 将ctx放回池中
			r.app.contextPool.Put(ctx)
			return
		}
		ctx.next = false
	}

	// 执行当前路由中间件
	for k := range middlewares {
		if err = middlewares[k](ctx); err != nil {
			// 生成事件触发信息
			if r.app.Config.EventHandler != nil && r.app.Config.ErrorEvent && r.app.Config.Trigger {
				trigger = getFuncInfo(middlewares[k])
				if r.app.Config.ShortPath {
					trigger.File = strings.TrimPrefix(trigger.File, r.app.Config.RootPath)
				}
				r.app.funcErrorHandler(resp, req, trigger, err)
			}
			// 将ctx放回池中
			r.app.contextPool.Put(ctx)
			return
		}
		if !ctx.next {
			// 将ctx放回池中
			r.app.contextPool.Put(ctx)
			return
		}
		ctx.next = false
	}

	// 执行当前路由处理器
	if err = handler(ctx); err != nil {
		if r.app.Config.EventHandler != nil && r.app.Config.ErrorEvent && r.app.Config.Trigger {
			trigger = getFuncInfo(handler)
			if r.app.Config.ShortPath {
				trigger.File = strings.TrimPrefix(trigger.File, r.app.Config.RootPath)
			}
			r.app.funcErrorHandler(resp, req, trigger, err)
		}
	}
	// 将ctx放回池中
	r.app.contextPool.Put(ctx)
}
