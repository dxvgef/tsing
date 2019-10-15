package tsing

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// 路由处理器
type Handler func(Context) error

// 中间件处理器
type MiddlewareHandler func(Context) (Context, error)

// 路由组
type RouterGroup struct {
	middlewareHandlers []MiddlewareHandler // 处理器
	basePath           string              // 基路径
	app                *App                // 调度器
}

// Handle 路由
func (r *RouterGroup) Handle(method string, path string, handler Handler, middlewareHandlers ...MiddlewareHandler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.Handle(method, path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// PATH 定义路由到目录，不支持路由组和中间件
func (r *RouterGroup) PATH(url string, local string, list bool) {
	defer func() {
		if err := recover(); err != nil {
			// 记录panic事件，但不执行 ServerError处理器，而是直接退出进程
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	if strings.HasPrefix(url, "/") == false {
		url = "/" + url
	}
	if strings.HasSuffix(url, "/") == false {
		url += "/"
	}
	url += "*filepath"

	// 使用GET方法模拟httprouter.ServeFiles()，防止其内部直接输出404消息给客户端
	r.app.httpRouter.GET(url, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		// 如果请求的是目录，而判断是否允许列出目录
		if params.ByName("filepath") == "" || params.ByName("filepath")[len(params.ByName("filepath"))-1:] == "/" {
			if list == false {
				// 如果不允许列出目录，则触发404事件处理
				r.app.event404(resp, req)
				return
			}
		}

		// 判断请求的文件是否存在
		file := local + params.ByName("filepath")
		if _, err := os.Stat(file); err != nil {
			r.app.event404(resp, req)
			return
		}
		http.ServeFile(resp, req, file)
	})
}

// FILE 定义路由到文件，不支持路由组和中间件
func (r *RouterGroup) FILE(url string, local string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	// 使用GET方法模拟httprouter.ServeFiles()，防止其内部直接输出404消息给客户端
	r.app.httpRouter.GET(url, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		if _, err := os.Stat(local); err != nil {
			r.app.event404(resp, req)
			return
		}
		http.ServeFile(resp, req, local)
	})
}

// GROUP 路由组
func (r *RouterGroup) GROUP(path string, middlewareHandlers ...MiddlewareHandler) RouterGroup {
	group := RouterGroup{
		basePath:           r.basePath + path,            // 继承父组的路径
		app:                r.app,                        // 传入调度器
		middlewareHandlers: append(r.middlewareHandlers), // 继承父组的中间件
	}
	// 加入当前传入的钩子
	for k := range middlewareHandlers {
		group.middlewareHandlers = append(group.middlewareHandlers, middlewareHandlers[k])
	}
	return group
}

// GET 路由
func (r *RouterGroup) GET(path string, handler Handler, middlewareHandlers ...MiddlewareHandler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.GET(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// POST 路由
func (r *RouterGroup) POST(path string, handler Handler, middlewareHandlers ...MiddlewareHandler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.POST(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// PUT 路由
func (r *RouterGroup) PUT(path string, handler Handler, middlewareHandlers ...MiddlewareHandler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.PUT(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// HEAD 路由
func (r *RouterGroup) HEAD(path string, handler Handler, middlewareHandlers ...MiddlewareHandler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.HEAD(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// PATCH 路由
func (r *RouterGroup) PATCH(path string, handler Handler, middlewareHandlers ...MiddlewareHandler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.PATCH(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// DELETE 路由
func (r *RouterGroup) DELETE(path string, handler Handler, middlewareHandlers ...MiddlewareHandler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.DELETE(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// OPTIONS 路由
func (r *RouterGroup) OPTIONS(path string, handler Handler, middlewareHandlers ...MiddlewareHandler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.OPTIONS(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, middlewareHandlers)
	})
}

// 执行处理器函数
func (r *RouterGroup) execute(resp http.ResponseWriter, req *http.Request, params httprouter.Params, handler Handler, middlewareHandlers []MiddlewareHandler) {
	var ctx Context
	ctx.Request = req
	ctx.ResponseWriter = resp
	ctx.app = r.app
	ctx.routerParams = params

	var err error

	// 执行路由组的中间件处理器
	for k := range r.middlewareHandlers {
		ctx, err = r.middlewareHandlers[k](ctx)
		if err != nil {
			r.app.event500(resp, req, err)
			return
		}
		if ctx.next == false {
			return
		}
	}

	// 执行当前路由中间件处理器
	for k := range middlewareHandlers {
		var err error
		ctx, err = middlewareHandlers[k](ctx)
		if err != nil {
			r.app.event500(resp, req, err)
			return
		}
		if ctx.next == false {
			return
		}
	}

	// 执行当前路由处理器
	err = handler(ctx)
	if err != nil {
		r.app.event500(resp, req, err)
	}
}
