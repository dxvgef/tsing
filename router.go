package tsing

import (
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 中止索引
const abortIndex int8 = math.MaxInt8 / 2

// 路由组接口
type RouterInterface interface {
	Append(...Handler) RouterInterface
	Handle(string, string, ...Handler) RouterInterface
	Any(string, ...Handler) RouterInterface
	GET(string, ...Handler) RouterInterface
	POST(string, ...Handler) RouterInterface
	DELETE(string, ...Handler) RouterInterface
	PATCH(string, ...Handler) RouterInterface
	PUT(string, ...Handler) RouterInterface
	OPTIONS(string, ...Handler) RouterInterface
	HEAD(string, ...Handler) RouterInterface
	File(string, string) RouterInterface
	Dir(string, string) RouterInterface
}

// 路由组
type Router struct {
	Handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

// 计算绝对路径
func (router *Router) calculateAbsolutePath(path string) string {
	return joinPaths(router.basePath, path)
}

// 合并处理器
func (router *Router) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(router.Handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("Too many handlers")
	}
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, router.Handlers)
	copy(mergedHandlers[len(router.Handlers):], handlers)
	return mergedHandlers
}

// 获得路由器对象
func (router *Router) getRouter() RouterInterface {
	if router.root {
		return router.engine
	}
	return router
}

// 添加处理器
func (router *Router) Append(handlers ...Handler) RouterInterface {
	router.Handlers = append(router.Handlers, handlers...)
	return router.getRouter()
}

// 定义路由组
func (router *Router) Group(path string, handlers ...Handler) *Router {
	return &Router{
		Handlers: router.combineHandlers(handlers),
		basePath: router.calculateAbsolutePath(path),
		engine:   router.engine,
	}
}

// 处理路由
func (router *Router) handle(method, path string, handlers HandlersChain) RouterInterface {
	absolutePath := router.calculateAbsolutePath(path)     // 计算绝对路径
	handlers = router.combineHandlers(handlers)            // 合并处理器
	router.engine.addRoute(method, absolutePath, handlers) // 添加路由
	return router.getRouter()
}

// 注册自定义HTTP方法的路由
func (router *Router) Handle(method, path string, handlers ...Handler) RouterInterface {
	if matches, err := regexp.MatchString("^[A-Z]+$", method); !matches || err != nil {
		panic("The HTTP method [" + method + "] is not valid")
	}
	return router.handle(method, path, handlers)
}

// 注册POST路由
func (router *Router) POST(path string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodPost, path, handlers)
}

// 注册GET路由
func (router *Router) GET(path string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodGet, path, handlers)
}

// 注册DELETE路由
func (router *Router) DELETE(path string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodDelete, path, handlers)
}

// 注册PATCH路由
func (router *Router) PATCH(path string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodPatch, path, handlers)
}

// 注册PUT路由
func (router *Router) PUT(path string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodPut, path, handlers)
}

// 注册OPTIONS路由
func (router *Router) OPTIONS(path string, handlers ...Handler) RouterInterface {
	path = filepath.Clean(path)
	return router.handle(http.MethodOptions, path, handlers)
}

// 注册HEAD路由
func (router *Router) HEAD(path string, handlers ...Handler) RouterInterface {
	path = filepath.Clean(path)
	return router.handle(http.MethodHead, path, handlers)
}

// 注册所有路由
func (router *Router) Any(path string, handlers ...Handler) RouterInterface {
	router.handle(http.MethodGet, path, handlers)
	router.handle(http.MethodPost, path, handlers)
	router.handle(http.MethodPut, path, handlers)
	router.handle(http.MethodPatch, path, handlers)
	router.handle(http.MethodHead, path, handlers)
	router.handle(http.MethodOptions, path, handlers)
	router.handle(http.MethodDelete, path, handlers)
	router.handle(http.MethodConnect, path, handlers)
	router.handle(http.MethodTrace, path, handlers)
	return router.getRouter()
}

// 注册一个响应服务端文件的路由
func (router *Router) File(path, file string) RouterInterface {
	file = filepath.Clean(file)
	if strings.Contains(path, ":") || strings.Contains(path, "*") {
		panic("this route cannot use ':' and '*' parameter")
	}
	handler := func(ctx *Context) error {
		fileInfo, err := os.Stat(file)
		if err != nil {
			panic("Unable to find file '" + file + "'")
		}
		if fileInfo.IsDir() {
			panic("This route cannot set a directory")
		}
		http.ServeFile(ctx.ResponseWriter, ctx.Request, file)
		return nil
	}
	router.GET(path, handler)
	router.HEAD(path, handler)
	return router.getRouter()
}

// 注册一个响应服务端目录的路由
func (router *Router) Dir(path, serverPath string) RouterInterface {
	serverPath = filepath.Clean(serverPath)
	if strings.Contains(path, ":") || strings.Contains(path, "*") {
		panic("this route cannot use ':' and '*' parameter")
	}

	if serverPath == "" {
		panic("serverPath cannot be empty")
	}
	if path[len(path)-1] != 47 {
		panic("path must end with '/'")
	}

	handler := func(ctx *Context) error {
		fileInfo, err := os.Stat(serverPath)
		if err != nil {
			panic("cannot find directory '" + serverPath + "'")
		}
		if !fileInfo.IsDir() {
			panic("This route cannot set a file")
		}
		http.ServeFile(ctx.ResponseWriter, ctx.Request, serverPath)
		return nil
	}
	router.GET(path, handler)
	router.HEAD(path, handler)
	return router.getRouter()
}
