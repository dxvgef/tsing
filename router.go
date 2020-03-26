package tsing

import (
	"math"
	"net/http"
	"regexp"
)

// 中止索引
const abortIndex int8 = math.MaxInt8 / 2

// 路由组接口
type RouterInterface interface {
	Append(...HandlerFunc) RouterInterface
	Handle(string, string, ...HandlerFunc) RouterInterface
	Any(string, ...HandlerFunc) RouterInterface
	GET(string, ...HandlerFunc) RouterInterface
	POST(string, ...HandlerFunc) RouterInterface
	DELETE(string, ...HandlerFunc) RouterInterface
	PATCH(string, ...HandlerFunc) RouterInterface
	PUT(string, ...HandlerFunc) RouterInterface
	OPTIONS(string, ...HandlerFunc) RouterInterface
	HEAD(string, ...HandlerFunc) RouterInterface

	// StaticFile(string, string) RouterInterface
	// Static(string, string) RouterInterface
	// StaticFS(string, http.FileSystem) RouterInterface
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

// 获得路由组对象
func (router *Router) getGroup() RouterInterface {
	if router.root {
		return router.engine
	}
	return router
}

// 添加处理器
func (router *Router) Append(handlers ...HandlerFunc) RouterInterface {
	router.Handlers = append(router.Handlers, handlers...)
	return router.getGroup()
}

// 定义路由组
func (router *Router) Group(path string, handlers ...HandlerFunc) *Router {
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
	return router.getGroup()
}

// 注册自定义HTTP方法的路由
func (router *Router) Handle(method, path string, handlers ...HandlerFunc) RouterInterface {
	if matches, err := regexp.MatchString("^[A-Z]+$", method); !matches || err != nil {
		panic("HTTP method " + method + " is not valid")
	}
	return router.handle(method, path, handlers)
}

// 注册POST路由
func (router *Router) POST(path string, handlers ...HandlerFunc) RouterInterface {
	return router.handle(http.MethodPost, path, handlers)
}

// 注册GET路由
func (router *Router) GET(path string, handlers ...HandlerFunc) RouterInterface {
	return router.handle(http.MethodGet, path, handlers)
}

// 注册DELETE路由
func (router *Router) DELETE(path string, handlers ...HandlerFunc) RouterInterface {
	return router.handle(http.MethodDelete, path, handlers)
}

// 注册PATCH路由
func (router *Router) PATCH(path string, handlers ...HandlerFunc) RouterInterface {
	return router.handle(http.MethodPatch, path, handlers)
}

// 注册PUT路由
func (router *Router) PUT(path string, handlers ...HandlerFunc) RouterInterface {
	return router.handle(http.MethodPut, path, handlers)
}

// 注册OPTIONS路由
func (router *Router) OPTIONS(path string, handlers ...HandlerFunc) RouterInterface {
	return router.handle(http.MethodOptions, path, handlers)
}

// 注册HEAD路由
func (router *Router) HEAD(path string, handlers ...HandlerFunc) RouterInterface {
	return router.handle(http.MethodHead, path, handlers)
}

// 注册所有路由
func (router *Router) Any(path string, handlers ...HandlerFunc) RouterInterface {
	router.handle(http.MethodGet, path, handlers)
	router.handle(http.MethodPost, path, handlers)
	router.handle(http.MethodPut, path, handlers)
	router.handle(http.MethodPatch, path, handlers)
	router.handle(http.MethodHead, path, handlers)
	router.handle(http.MethodOptions, path, handlers)
	router.handle(http.MethodDelete, path, handlers)
	router.handle(http.MethodConnect, path, handlers)
	router.handle(http.MethodTrace, path, handlers)
	return router.getGroup()
}
