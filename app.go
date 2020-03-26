package tsing

import (
	"net/http"
	"sync"
)

// 32MB
const defaultMultipartMemory = 32 << 20

// 处理器
type HandlerFunc func(*Context)

// 处理器链
type HandlersChain []HandlerFunc

// 引擎
type Engine struct {
	// 路由组
	*Router

	// 使用url.RawPath查找参数
	UseRawPath bool

	// 反转义路由参数
	UnescapePathValues bool

	// 分配给http.Request的值
	MaxMultipartMemory int64

	// ctx池
	pool sync.Pool

	// 路由树
	trees methodTrees
}

// 添加路由
func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}
	if method == "" {
		panic("HTTP method can not be empty")
	}
	if len(handlers) == 0 {
		panic("There must be at least one handler")
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
func New() *Engine {
	// 初始化一个引擎
	engine := &Engine{
		// 初始化根路由组
		Router: &Router{
			Handlers: nil,  // 空处理器链
			basePath: "/",  // 设置基本路径
			root:     true, // 标记为根路由组
		},
		UseRawPath:         false,
		UnescapePathValues: true,
		MaxMultipartMemory: defaultMultipartMemory,

		// 初始化一个路由树，递增值是
		trees: make(methodTrees, 0, 7),
	}

	// 将引擎对象传入路由组中，便于访问引擎对象
	engine.engine = engine

	// 设置ctx池
	engine.pool.New = func() interface{} {
		return &Context{engine: engine}
	}

	return engine
}

// 实现http.Handler接口，并且是连接调度的入口
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		log.Println(err)
	// 	}
	// }()

	// 从池中取出一个ctx
	c := engine.pool.Get().(*Context)
	// 重置取出的ctx
	c.reset(req, w)

	// 处理请求
	engine.handleRequest(c)

	// 将ctx放回池中
	engine.pool.Put(c)
}

// 处理连接请求
func (engine *Engine) handleRequest(c *Context) {
	httpMethod := c.Request.Method
	rPath := c.Request.URL.Path
	unescape := false

	if engine.UseRawPath && len(c.Request.URL.RawPath) > 0 {
		rPath = c.Request.URL.RawPath
		unescape = engine.UnescapePathValues
	}

	// 先根据HTTP方法查找节点
	for k := range engine.trees {
		if engine.trees[k].method != httpMethod {
			continue
		}
		root := engine.trees[k].root
		value := root.getValue(rPath, c.URLParams, unescape)
		if value.handlers != nil {
			// 为ctx属性赋值
			c.handlers = value.handlers
			c.URLParams = value.params
			c.fullPath = value.fullPath
			// 执行ctx中的处理器
			c.Next()
			return
		}
		break
	}

	var err error
	for k := range engine.trees {
		if engine.trees[k].method == httpMethod {
			continue
		}
		// 405错误
		if value := engine.trees[k].root.getValue(rPath, nil, unescape); value.handlers != nil {
			c.handlers = nil
			c.ResponseWriter.WriteHeader(http.StatusMethodNotAllowed)
			_, err = c.ResponseWriter.Write(stringToBytes(http.StatusText(http.StatusMethodNotAllowed)))
			if err != nil {
				panic(err.Error())
			}
			return
		}
	}

	// 404错误
	c.handlers = nil
	c.ResponseWriter.WriteHeader(http.StatusNotFound)
	_, err = c.ResponseWriter.Write(stringToBytes(http.StatusText(http.StatusNotFound)))
	if err != nil {
		panic(err.Error())
	}
}
