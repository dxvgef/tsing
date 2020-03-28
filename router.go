package tsing

import (
	"math"
	"net/http"
	"os"
	"path"
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
func (router *Router) Group(urlPath string, handlers ...Handler) *Router {
	return &Router{
		Handlers: router.combineHandlers(handlers),
		basePath: router.calculateAbsolutePath(urlPath),
		engine:   router.engine,
	}
}

// 处理路由
func (router *Router) handle(method, urlPath string, handlers HandlersChain) RouterInterface {
	absolutePath := router.calculateAbsolutePath(urlPath)  // 计算绝对路径
	handlers = router.combineHandlers(handlers)            // 合并处理器
	router.engine.addRoute(method, absolutePath, handlers) // 添加路由
	return router.getRouter()
}

// 注册自定义HTTP方法的路由
func (router *Router) Handle(method, urlPath string, handlers ...Handler) RouterInterface {
	if matches, err := regexp.MatchString("^[A-Z]+$", method); !matches || err != nil {
		panic("The HTTP method [" + method + "] is not valid")
	}
	return router.handle(method, urlPath, handlers)
}

// 注册POST路由
func (router *Router) POST(urlPath string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodPost, urlPath, handlers)
}

// 注册GET路由
func (router *Router) GET(urlPath string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodGet, urlPath, handlers)
}

// 注册DELETE路由
func (router *Router) DELETE(urlPath string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodDelete, urlPath, handlers)
}

// 注册PATCH路由
func (router *Router) PATCH(urlPath string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodPatch, urlPath, handlers)
}

// 注册PUT路由
func (router *Router) PUT(urlPath string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodPut, urlPath, handlers)
}

// 注册OPTIONS路由
func (router *Router) OPTIONS(urlPath string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodOptions, urlPath, handlers)
}

// 注册HEAD路由
func (router *Router) HEAD(urlPath string, handlers ...Handler) RouterInterface {
	return router.handle(http.MethodHead, urlPath, handlers)
}

// 注册所有路由
func (router *Router) Any(relativePath string, handlers ...Handler) RouterInterface {
	router.handle(http.MethodGet, relativePath, handlers)
	router.handle(http.MethodPost, relativePath, handlers)
	router.handle(http.MethodPut, relativePath, handlers)
	router.handle(http.MethodPatch, relativePath, handlers)
	router.handle(http.MethodHead, relativePath, handlers)
	router.handle(http.MethodOptions, relativePath, handlers)
	router.handle(http.MethodDelete, relativePath, handlers)
	router.handle(http.MethodConnect, relativePath, handlers)
	router.handle(http.MethodTrace, relativePath, handlers)
	return router.getRouter()
}

// 注册一个响应服务端文件的路由
func (router *Router) File(urlPath, absPath string) RouterInterface {
	absPath = filepath.Clean(absPath)
	if strings.Contains(urlPath, ":") || strings.Contains(urlPath, "*") {
		panic("urlPath for this route cannot use ':' and '*'")
	}
	handler := func(ctx *Context) error {
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			panic("Unable to find file '" + absPath + "'")
		}
		if fileInfo.IsDir() {
			panic("This route cannot set a directory")
		}
		http.ServeFile(ctx.ResponseWriter, ctx.Request, absPath)
		// ctx.Abort()
		return nil
	}
	router.GET(urlPath, handler)
	router.HEAD(urlPath, handler)
	return router.getRouter()
}

// 注册一个响应服务端目录的路由
func (router *Router) Dir(urlPath, absPath string) RouterInterface {
	absPath = filepath.Clean(absPath)
	if strings.Contains(urlPath, ":") || strings.Contains(urlPath, "*") {
		panic("urlPath for this route cannot use ':' and '*'")
	}
	if urlPath[len(urlPath)-1] != 47 {
		panic("urlPath must end with '/'")
	}
	if absPath == "" {
		panic("absPath cannot be empty")
	}

	handler := func(ctx *Context) error {
		relPath := ctx.PathParams.Value("filepath")
		finalAbsPath := path.Join(absPath, relPath)
		_, err := os.Stat(finalAbsPath)
		if err != nil {
			panic("Unable to find directory '" + finalAbsPath + "'")
		}
		http.ServeFile(ctx.ResponseWriter, ctx.Request, finalAbsPath)
		// ctx.Abort()
		return nil
	}

	finalUrlPath := path.Join(urlPath, "/*filepath")
	router.GET(finalUrlPath, handler)
	router.HEAD(finalUrlPath, handler)

	return router.getRouter()
}
