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
	maxParams   int
	maxSections int
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

	// 设置默认配置或使用提供的配置
	if len(config) > 0 {
		engine.config = config[0]
	} else {
		engine.config = Config{
			MaxMultipartMemory:     32 << 20, // 32 MB
			HandleMethodNotAllowed: false,
		}
	}

	engine.contextPool.New = func() any {
		return engine.allocateContext(engine.maxParams)
	}

	return engine
}

func (engine *Engine) allocateContext(maxParams int) *Context {
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
		log.Fatalln("方法不能为空")
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
	ctx, ok := engine.contextPool.Get().(*Context)
	if !ok {
		panic("context pool is not set")
	}
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
					// _, _ = ctx.ResponseWriter.Write(strToBytes(ctx.Error.Error())) //nolint:errcheck
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
	t := engine.trees

	// 在指定方法树中查找路径
	for i := 0; i < len(t); i++ {
		if t[i].method != method {
			continue
		}

		root := t[i].root
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
		if engine.config.AfterHandler != nil {
			defer engine.config.AfterHandler(ctx)
		}

		for _, handler := range node.handlers {
			if ctx.broke {
				break
			}
			if err = handler(ctx); err != nil {
				handleError(ctx, engine, err, http.StatusInternalServerError)
				return
			}
		}
		return
	}

	// 处理 405 错误
	if engine.config.HandleMethodNotAllowed {
		for _, tree := range engine.trees {
			if tree.method == method {
				continue
			}
			node = tree.root.getValue(url, nil, ctx.skippedNodes)
			if node.handlers != nil {
				handleError(ctx, engine, errors.New(http.StatusText(http.StatusMethodNotAllowed)), http.StatusMethodNotAllowed)
				return
			}
		}
	}

	// 404 错误
	handleError(ctx, engine, errors.New(http.StatusText(http.StatusNotFound)), http.StatusNotFound)
}

// 处理错误并执行错误处理器
func handleError(ctx *Context, engine *Engine, err error, status int) {
	ctx.broke = true
	ctx.Status = status
	ctx.Error = err

	if engine.config.ErrorHandler != nil {
		engine.config.ErrorHandler(ctx)
		return
	}
	ctx.ResponseWriter.WriteHeader(ctx.Status)
	// if _, err = ctx.ResponseWriter.Write(strToBytes(ctx.Error.Error())); err != nil {
	// }
}
