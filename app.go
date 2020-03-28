package tsing

import (
	"net/http"
	"sync"
)

// 路由处理器
type Handler func(*Context) error

// 处理器链
type HandlersChain []Handler

// 引擎配置
type Config struct {
	UseRawPath         bool         // 使用url.RawPath查找参数
	UnescapePathValues bool         // 反转义路由参数
	MaxMultipartMemory int64        // 允许的请求Body大小(默认1 << 20 = 1MB)
	EventHandler       EventHandler // 事件-处理器函数，如果不赋值，则不启用事件
	EventTrace         bool         // 事件-启用跟踪信息
	EventShortPath     bool         // 事件-启用短文件路径
	RootPath           string       // 应用的根路径
	EventHandlerError  bool         // 事件-启用处理器返回的错误
	EventSource        bool         // 事件-启用来源
	Recover            bool         // 自动恢复处理器的panic
}

// 引擎
type Engine struct {
	*Router                 // 路由器
	Config      *Config     // 配置
	contextPool sync.Pool   // context池
	eventPool   sync.Pool   // event池
	trees       methodTrees // 路由树
}

// 添加路由
func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
	if path[0] != '/' {
		panic("The path must begin with '/'")
	}
	if method == "" {
		panic("HTTP method can not be empty")
	}
	if len(handlers) == 0 {
		panic("[" + method + "]" + path + " must be at least one handler")
	}

	// 查找方法是否存在
	root := engine.trees.get(method)
	// 如果方法不存在
	if root == nil {
		// 创建一个新的根节点
		root = new(node)
		root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{
			method: method,
			root:   root,
		})
	}
	root.addRoute(path, handlers)
}

// 创建一个新引擎
func New(config *Config) *Engine {
	if config.MaxMultipartMemory == 0 {
		config.MaxMultipartMemory = MaxMultipartMemory
	}
	// 初始化一个引擎
	engine := &Engine{
		// 初始化根路由组
		Router: &Router{
			Handlers: nil,  // 空处理器链
			basePath: "/",  // 设置基本路径
			root:     true, // 标记为根路由组
		},
		Config: config,
		// UseRawPath:         false,
		// UnescapePathValues: true,
		// MaxMultipartMemory: defaultMultipartMemory,

		// 初始化一个路由树，递增值是
		trees: make(methodTrees, 0, 7),
	}

	// 将引擎对象传入路由组中，便于访问引擎对象
	engine.engine = engine

	// 设置ctx池
	engine.contextPool.New = func() interface{} {
		return &Context{engine: engine}
	}

	// 设置event池
	if config.EventHandler != nil {
		engine.eventPool.New = func() interface{} {
			return &Event{
				Status:  0,
				Message: nil,
				Source:  nil,
				Trace:   nil,
			}
		}
	}

	return engine
}

// 实现http.Handler接口，并且是连接调度的入口
func (engine *Engine) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if engine.Config.Recover {
		defer func() {
			err := recover()
			if err != nil && engine.Config.EventHandler != nil {
				// 触发panic事件
				engine.panicEvent(resp, req, err)
			}
		}()
	}

	// 从池中取出一个ctx
	ctx := engine.contextPool.Get().(*Context)
	// 重置取出的ctx
	ctx.reset(req, resp)

	// 处理请求
	engine.handleRequest(ctx)

	// 将ctx放回池中
	engine.contextPool.Put(ctx)
}

// 处理连接请求
func (engine *Engine) handleRequest(ctx *Context) {
	httpMethod := ctx.Request.Method
	rPath := ctx.Request.URL.Path
	unescape := false

	if engine.Config.UseRawPath && len(ctx.Request.URL.RawPath) > 0 {
		rPath = ctx.Request.URL.RawPath
		unescape = engine.Config.UnescapePathValues
	}

	for k := range engine.trees {
		if engine.trees[k].method != httpMethod {
			continue
		}

		root := engine.trees[k].root
		value := root.getValue(rPath, ctx.PathParams, unescape)
		if value.handlers != nil {
			// 为ctx属性赋值
			ctx.handlers = value.handlers
			ctx.PathParams = value.params
			ctx.fullPath = value.fullPath
			// 执行ctx中的处理器
			ctx.next()
			return
		}
		break
	}

	for k := range engine.trees {
		if engine.trees[k].method == httpMethod {
			continue
		}
		if value := engine.trees[k].root.getValue(rPath, nil, unescape); value.handlers != nil {
			ctx.handlers = nil
			// 触发405事件
			engine.methodNotAllowedEvent(ctx.ResponseWriter, ctx.Request)
			return
		}
	}

	// 触发404事件
	ctx.handlers = nil
	engine.notFoundEvent(ctx.ResponseWriter, ctx.Request)
}
