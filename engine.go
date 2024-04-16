package tsing

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Config 引擎参数配置
type Config struct {
	MaxMultipartMemory     int64           // 允许的请求Body大小(默认32 << 20 = 32MB)
	Recovery               bool            // 自动恢复panic，防止进程退出
	HandleMethodNotAllowed bool            // 不处理 405 错误（可以减少路由匹配时间），以 404 错误返回
	ErrorHandler           CallbackHandler // 错误回调处理器
	AfterHandler           CallbackHandler // 后置回调处理器，总是会在其它处理器全部执行完之后执行
}

// Engine 引擎
type Engine struct {
	RouterGroup
	config      Config
	maxParams   uint16
	maxSections uint16
	contextPool sync.Pool
	trees       methodTrees
}

// Handler 路由处理器
type Handler func(*Context) error

// CallbackHandler 回调处理器
type CallbackHandler func(*Context)

// HandlersChain 处理器链
type HandlersChain []Handler

// New 新建引擎实例
func New(config ...Config) *Engine {
	engine := &Engine{
		RouterGroup: RouterGroup{
			handlers: nil,
			basePath: "/",
			root:     true,
		},
		trees: make(methodTrees, 0, 9),
	}
	engine.RouterGroup.engine = engine

	if len(config) == 0 {
		engine.config = Config{
			MaxMultipartMemory:     32 << 20, // 32 MB,
			HandleMethodNotAllowed: false,
		}
	} else {
		engine.config = config[0]
	}

	engine.contextPool.New = func() any {
		return engine.allocateContext(engine.maxParams)
	}

	return engine
}

func (engine *Engine) allocateContext(maxParams uint16) *Context {
	v := make(Params, 0, maxParams)
	skippedNodes := make([]skippedNode, 0, engine.maxSections)
	return &Context{
		engine:       engine,
		params:       &v,
		skippedNodes: &skippedNodes,
	}
}

func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
	if path[0] != '/' {
		log.Fatalln("路径必须以'/'开头")
	}
	if method == "" {
		log.Fatalln("HTTP方法不能为空")
	}
	if len(handlers) == 0 {
		log.Fatalln("必须有至少一个处理器")
	}

	root := engine.trees.get(method)
	if root == nil {
		root = new(Node)
		root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)

	// 更新 maxParams
	if paramsCount := countParams(path); paramsCount > engine.maxParams {
		engine.maxParams = paramsCount
	}

	if sectionsCount := countSections(path); sectionsCount > engine.maxSections {
		engine.maxSections = sectionsCount
	}
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := engine.contextPool.Get().(*Context)
	ctx.Request = req
	ctx.ResponseWriter = w
	ctx.reset()

	// 处理panic
	if engine.config.Recovery {
		defer func() {
			if err := recover(); err != nil {
				ctx.Status = http.StatusInternalServerError
				ctx.Error = fmt.Errorf("%v", err)
				if engine.config.ErrorHandler != nil {
					engine.config.ErrorHandler(ctx)
				} else {
					ctx.ResponseWriter.WriteHeader(ctx.Status)
					_, _ = ctx.ResponseWriter.Write(strToBytes(ctx.Error.Error())) //nolint:errcheck
				}
			}
		}()
	}

	engine.handleRequest(ctx)

	engine.contextPool.Put(ctx)
}

func (engine *Engine) handleRequest(ctx *Context) {
	var (
		err  error
		node nodeValue
	)
	method := ctx.Request.Method
	url := ctx.Request.URL.Path

	// 在指定方法树中查找路径
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		// 只在方法匹配的树中查找
		if t[i].method != method {
			continue
		}
		root := t[i].root
		// 查找路由节点
		node = root.getValue(url, ctx.params, ctx.skippedNodes)
		// 如果存在路由参数
		if node.params != nil {
			ctx.params = node.params
		}
		break
	}

	// 如果找到了处理器
	if node.handlers != nil {
		ctx.fullPath = node.fullPath
		// 函数退出时执行后置处理器
		if engine.config.AfterHandler != nil {
			defer engine.config.AfterHandler(ctx)
		}
		// 执行处理器
		for k := range node.handlers {
			if ctx.broke {
				break
			}
			if err = node.handlers[k](ctx); err != nil {
				ctx.broke = true
				ctx.Status = http.StatusInternalServerError
				ctx.Error = err
				if engine.config.ErrorHandler != nil {
					engine.config.ErrorHandler(ctx)
				} else {
					ctx.ResponseWriter.WriteHeader(ctx.Status)
					_, _ = ctx.ResponseWriter.Write(strToBytes(ctx.Error.Error())) //nolint:errcheck
				}
				break
			}
		}
		return
	}
	// 处理 405 错误
	if engine.config.HandleMethodNotAllowed {
		for _, tree := range engine.trees {
			// 只在方法不匹配的树中查找
			if tree.method == method {
				continue
			}
			// 找到了其它方法的路径
			if node = tree.root.getValue(url, nil, ctx.skippedNodes); node.handlers != nil {
				// 405 错误
				ctx.broke = true
				ctx.Status = http.StatusMethodNotAllowed
				ctx.Error = errors.New(http.StatusText(http.StatusMethodNotAllowed))
				if engine.config.ErrorHandler != nil {
					engine.config.ErrorHandler(ctx)
				} else {
					ctx.ResponseWriter.WriteHeader(ctx.Status)
					_, _ = ctx.ResponseWriter.Write(strToBytes(ctx.Error.Error())) //nolint:errcheck
				}
				return
			}
		}
	}
	// 404 错误
	ctx.broke = true
	ctx.Status = http.StatusNotFound
	ctx.Error = errors.New(http.StatusText(http.StatusNotFound))
	if engine.config.ErrorHandler != nil {
		engine.config.ErrorHandler(ctx)
	} else {
		ctx.ResponseWriter.WriteHeader(ctx.Status)
		_, _ = ctx.ResponseWriter.Write(strToBytes(ctx.Error.Error())) //nolint:errcheck
	}
}
