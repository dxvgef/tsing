package tsing

import (
	"net/http"
	"strings"
)

// Router 路由器接口，包括单路由和路由组
type Router interface {
	Routes
	Group(string, ...Handler) *RouterGroup
}

// Routes 定义所有路由器接口
type Routes interface {
	Use(...Handler)

	After(...Handler)
	Handle(string, string, ...Handler)
	GET(string, ...Handler)
	POST(string, ...Handler)
	DELETE(string, ...Handler)
	PATCH(string, ...Handler)
	PUT(string, ...Handler)
	OPTIONS(string, ...Handler)
	HEAD(string, ...Handler)
	Match([]string, string, ...Handler)
}

// RouterGroup 路由组
type RouterGroup struct {
	handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

// Use 使用中间件
func (group *RouterGroup) Use(handlers ...Handler) {
	group.handlers = append(group.handlers, handlers...)
}

// Group 注册路由组
func (group *RouterGroup) Group(relativePath string, handlers ...Handler) *RouterGroup {
	return &RouterGroup{
		handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers)
}

// Handle 注册自定义方法的路由
func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...Handler) {
	group.handle(httpMethod, relativePath, handlers)
}

// POST 注册POST方法的路由
func (group *RouterGroup) POST(relativePath string, handlers ...Handler) {
	group.handle(http.MethodPost, relativePath, handlers)
}

// GET 注册GET方法的路由
func (group *RouterGroup) GET(relativePath string, handlers ...Handler) {
	group.handle(http.MethodGet, relativePath, handlers)
}

// DELETE 注册DELETE方法的路由
func (group *RouterGroup) DELETE(relativePath string, handlers ...Handler) {
	group.handle(http.MethodDelete, relativePath, handlers)
}

// PATCH 注册PATCH方法的路由
func (group *RouterGroup) PATCH(relativePath string, handlers ...Handler) {
	group.handle(http.MethodPatch, relativePath, handlers)
}

// PUT 注册PUT方法的路由
func (group *RouterGroup) PUT(relativePath string, handlers ...Handler) {
	group.handle(http.MethodPut, relativePath, handlers)
}

// OPTIONS 注册OPTIONS方法的路由
func (group *RouterGroup) OPTIONS(relativePath string, handlers ...Handler) {
	group.handle(http.MethodOptions, relativePath, handlers)
}

// HEAD 注册HEAD方法的路由
func (group *RouterGroup) HEAD(relativePath string, handlers ...Handler) {
	group.handle(http.MethodHead, relativePath, handlers)
}

// Match 为一个路径同时注册多个方法的路由
func (group *RouterGroup) Match(methods []string, relativePath string, handlers ...Handler) {
	for _, method := range methods {
		group.handle(method, relativePath, handlers)
	}
}

// StaticFile 注册一个指向服务端本地文件的静态路由，例如：
// StaticFile("favicon.ico", "./resources/favicon.ico")
func (group *RouterGroup) StaticFile(relativePath, filepath string) {
	group.staticFileHandler(relativePath, func(c *Context) error {
		c.ServeFile(filepath)
		return nil
	})
}

// StaticFileFS 与StaticFile函数类型，但可以自定义文件系统，例如：
// StaticFileFS("favicon.ico", "./resources/favicon.ico", Dir{".", false})
func (group *RouterGroup) StaticFileFS(relativePath, filepath string, fs http.FileSystem) {
	group.staticFileHandler(relativePath, func(c *Context) error {
		c.FileFromFS(filepath, fs)
		return nil
	})
}

func (group *RouterGroup) staticFileHandler(relativePath string, handler Handler) {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a staticNode file")
	}
	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(group.handlers) + len(handlers)
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.handlers)
	copy(mergedHandlers[len(group.handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}
