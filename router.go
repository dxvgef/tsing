package tsing

import (
	"errors"
	"net/http"
)

// 方法
type _MethodTree struct {
	method string // 方法
	root   *node  // 根节点
}

// 方法集合
type _MethodTrees []_MethodTree

// 路由信息数组
type RoutesInfo []RouteInfo

// 路由信息
type RouteInfo struct {
	Method      string
	Path        string
	Handler     string // 处理器函数名称
	HandlerFunc HandlerFunc
}

// 路由器接口
type Router interface {
	Routes
	Group(string, ...HandlerFunc) *RouterGroup
}

// 路由组
type RouterGroup struct {
	handlers HandlersChain // 处理器集合
	basePath string        // 基本路径
	app      *App          // 引擎
	root     bool          // 是否根路由组
}

// 合并处理器
func (group *RouterGroup) mergeHandlers(handlers HandlersChain) HandlersChain {
	// 计算所有处理器的大小
	finalSize := len(group.handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("注册了太多处理器")
	}
	// 合并处理器
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.handlers)
	copy(mergedHandlers[len(group.handlers):], handlers)
	return mergedHandlers
}

func (app *App) addRoute(method, path string, handlers HandlersChain) {
	if path[0] != '/' {
		panic("路径必须以'/'开头")
	}
	if method == "" {
		panic("HTTP方法不能为空")
	}

	if len(handlers) == 0 {
		panic("至少要有一个路由处理器")
	}

	// 查找该方法的根节点
	root := app.methodTrees.get(method)
	// 如果根节点不存在，则创建一个根节点，并添加到方法树
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		app.methodTrees = append(app.methodTrees, _MethodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)
}

// 根据方法查找节点
func (trees _MethodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) Routes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.mergeHandlers(handlers)
	group.app.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}

func (n *node) insertChild(numParams uint8, path string, fullPath string, handlers HandlersChain) {
	var offset int // 已处理路径的字节

	// 查找直到第一个通配符的前缀（以“：”或“ *”开头）
	for i, max := 0, len(path); numParams > 0; i++ {
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}

		// 查找通配符结尾（“/”或路径结尾）
		end := i + 1
		for end < max && path[end] != '/' {
			switch path[end] {
			// 通配符名称不得包含“：”和“ *”
			case ':', '*':
				panic("每个路由只允许使用一个通配符, 存在:" + path[i:] + " 路径 " + fullPath)
			default:
				end++
			}
		}

		// 检查该节点是否存在子节点
		// 如果在此处使用通配符，则无法访问
		if len(n.children) > 0 {
			panic("通配符路由" + path[i:end] + "冲突" + fullPath)
		}

		// 检查通配符是否有名称
		if end-i < 2 {
			panic("通配符必须有名称" + fullPath)
		}

		if c == ':' { // 路由参数
			// 在通配符开头分割路径
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}

			child := &node{
				nType:     param,
				maxParams: numParams,
				fullPath:  fullPath,
			}
			n.children = []*node{child}
			n.wildChild = true
			n = child
			n.priority++
			numParams--

			// 如果路径不是以通配符结尾，则存在
			// 将是另一个以'/'开头的非通配符子路径
			if end < max {
				n.path = path[offset:end]
				offset = end

				child := &node{
					maxParams: numParams,
					priority:  1,
					fullPath:  fullPath,
				}
				n.children = []*node{child}
				n = child
			}

		} else { // catchAll
			if end != max || numParams > 1 {
				// panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
				panic("捕获的所有路由仅允许在路径" + fullPath + "的末尾")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				// panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
				panic("与路径" + fullPath + "中路径段根的现有句柄的全部冲突")
			}

			// currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}

			n.path = path[offset:i]

			// first node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     catchAll,
				maxParams: 1,
				fullPath:  fullPath,
			}
			n.children = []*node{child}
			n.indices = string(path[i])
			n = child
			n.priority++

			// second node: node holding the variable
			child = &node{
				path:      path[i:],
				nType:     catchAll,
				maxParams: 1,
				handlers:  handlers,
				priority:  1,
				fullPath:  fullPath,
			}
			n.children = []*node{child}

			return
		}
	}

	// insert remaining path part and handle to the leaf
	n.path = path[offset:]
	n.handlers = handlers
	n.fullPath = fullPath
}

// 增加给定子节点的优先级，并在必要时重新排序
func (n *node) incrementChildPrio(pos int) int {
	n.children[pos].priority++
	prio := n.children[pos].priority

	// 移动到前面
	newPos := pos
	for newPos > 0 && n.children[newPos-1].priority < prio {
		// 交换节点位置
		n.children[newPos-1], n.children[newPos] = n.children[newPos], n.children[newPos-1]

		newPos--
	}

	// 建立新的字节符索引
	if newPos != pos {
		n.indices = n.indices[:newPos] + // 前缀不变，可能为空
			n.indices[pos:pos+1] + // 我们移动的字符索引
			n.indices[newPos:pos] + n.indices[pos+1:] // rest without char at 'pos'
	}

	return newPos
}

// 路由返回一部分已注册的路由，其中包括一些有用的信息，例如：http方法，路径和处理器名称
func (app *App) Routes() (routes RoutesInfo) {
	for _, tree := range app.methodTrees {
		routes = iterate("", tree.method, routes, tree.root)
	}
	return routes
}

// 迭代
func iterate(path, method string, routes RoutesInfo, root *node) RoutesInfo {
	path += root.path
	if len(root.handlers) > 0 {
		handlerFunc := root.handlers.last()
		routes = append(routes, RouteInfo{
			Method:      method,
			Path:        path,
			Handler:     getHandlerName(handlerFunc),
			HandlerFunc: handlerFunc,
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	return routes
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup) returnObj() Routes {
	if group.root {
		return &group.app.Router
	}
	return group
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	return func(ctx *Context) error {
		if _, nolisting := fs.(*onlyfilesFS); nolisting {
			ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
			// 抛出404事件
			if group.app.Config.EventHandler != nil {
				group.app.Config.EventHandler(ctx, Event{
					Status:  http.StatusNotFound,
					Message: errors.New(http.StatusText(http.StatusNotFound)),
				})
			}
		}

		file := ctx.ParamValue("filepath")
		// 检查文件是否能访问(可能是文件不存在或者权限不足)
		if _, err := fs.Open(file); err != nil {
			ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
			// 抛出404事件
			if group.app.Config.EventHandler != nil {
				group.app.Config.EventHandler(ctx, Event{
					Status:  http.StatusNotFound,
					Message: errors.New(http.StatusText(http.StatusNotFound)),
				})
			}

			// 重置index
			ctx.nextHandlerIndex = -1
			return nil
		}

		fileServer.ServeHTTP(ctx.ResponseWriter, ctx.Request)
		return nil
	}
}
