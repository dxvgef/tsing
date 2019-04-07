package tsing

import (
	"net/http"
	"os"
	"strings"

	"fmt"

	"github.com/julienschmidt/httprouter"
)

// Handler 处理器类型
type Handler func(Context) error

// RouterGroup 路由组
type RouterGroup struct {
	handlers []Handler // 处理器
	basePath string    // 基路径
	app      *App      // 调度器
}

// Handle 路由
func (r *RouterGroup) Handle(method string, path string, handler Handler, handlers ...Handler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.Handle(method, path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, handlers)
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
				r.app.handle404(resp, req)
				return
			}
		}

		// 判断请求的文件是否存在
		file := local + params.ByName("filepath")
		if _, err := os.Stat(file); err != nil {
			r.app.handle404(resp, req)
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
			r.app.handle404(resp, req)
			return
		}
		http.ServeFile(resp, req, local)
	})
}

// GROUP 路由组
func (r *RouterGroup) GROUP(path string, handlers ...Handler) RouterGroup {
	group := RouterGroup{
		basePath: r.basePath + path,  // 继承父组的路径
		app:      r.app,              // 传入调度器
		handlers: append(r.handlers), // 继承父组的钩子
	}
	// 加入当前传入的钩子
	for k := range handlers {
		group.handlers = append(group.handlers, handlers[k])
	}
	return group
}

// GET 路由
func (r *RouterGroup) GET(path string, handler Handler, handlers ...Handler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.GET(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, handlers)
	})
}

// POST 路由
func (r *RouterGroup) POST(path string, handler Handler, handlers ...Handler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.POST(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, handlers)
	})
}

// PUT 路由
func (r *RouterGroup) PUT(path string, handler Handler, handlers ...Handler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.PUT(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, handlers)
	})
}

// HEAD 路由
func (r *RouterGroup) HEAD(path string, handler Handler, handlers ...Handler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.HEAD(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, handlers)
	})
}

//PATCH 路由
func (r *RouterGroup) PATCH(path string, handler Handler, handlers ...Handler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.PATCH(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, handlers)
	})
}

// DELETE 路由
func (r *RouterGroup) DELETE(path string, handler Handler, handlers ...Handler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.DELETE(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, handlers)
	})
}

//OPTIONS 路由
func (r *RouterGroup) OPTIONS(path string, handler Handler, handlers ...Handler) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err.(string))
			os.Exit(1)
		}
	}()
	r.app.httpRouter.OPTIONS(r.basePath+path, func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		r.execute(resp, req, params, handler, handlers)
	})
}

// 执行Handler
// func (r *RouterGroup) execute(resp http.ResponseWriter, req *http.Request, params httprouter.Params, handler Handler, handlers []Handler) {
func (r *RouterGroup) execute(resp http.ResponseWriter, req *http.Request, params httprouter.Params, handler Handler, handlers []Handler) {
	var ctx Context
	ctx.Request = req
	ctx.ResponseWriter = resp
	ctx.app = r.app
	ctx.routerParams = params

	for k := range r.handlers {
		ctx.next = false
		if err := r.handlers[k](ctx); err != nil {
			r.app.handle500(resp, req, err)
			return
		}
		if ctx.next == false {
			return
		}
	}

	for k := range handlers {
		ctx.next = false
		if err := handlers[k](ctx); err != nil {
			r.app.handle500(resp, req, err)
			return
		}
		if ctx.next == false {
			return
		}
	}

	err := handler(ctx)
	if err != nil {
		r.app.handle500(resp, req, err)
	}
}
