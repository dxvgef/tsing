package tsing

import (
	"net/url"
	"strings"
	"unicode"
)

// 节点类型
type nodeType uint8

const (
	static   nodeType = iota // 静态节点
	root                     // 根节点
	param                    // 参数
	catchAll                 // 所有
)

// 节点结构
type node struct {
	path      string        // 路径
	indices   string        // 指数
	children  []*node       // 子节点
	handlers  HandlersChain // 处理器集合
	priority  uint32        // 优先级
	nType     nodeType      // 节点类型
	maxParams uint8         // 最大参数
	wildChild bool          // 通配符
	fullPath  string        // 完整路径
}

// nodeValue holds return values of (*Node).getValue method
type nodeValue struct {
	handlers HandlersChain
	params   Params
	tsr      bool
	fullPath string
}

// 添加一个路径及处理器集合到一个方法合集中
// 非并发安全！
func (n *node) addRoute(path string, handlers HandlersChain) {
	fullPath := path
	n.priority++                   // 递增当前节点的优先级
	numParams := countParams(path) // 路由参数的数量

	parentFullPathIndex := 0 // 上级完整路径的索引位置

	// 非空的树(当前节点已经有路径，或者有子路径)
	if len(n.path) > 0 || len(n.children) > 0 {
	walk:
		for {
			// 更新当前节点的参数数量
			if numParams > n.maxParams {
				n.maxParams = numParams
			}

			// 找到最长的公共前缀(不包含'：'或'*'字符的路径字符串)的字符截止位置
			i := 0
			max := min(len(path), len(n.path))
			for i < max && path[i] == n.path[i] {
				i++
			}

			// 根据当前节点的路径，切割出子路径
			if i < len(n.path) {
				child := node{
					path:      n.path[i:],
					wildChild: n.wildChild,
					indices:   n.indices,
					children:  n.children,
					handlers:  n.handlers,
					priority:  n.priority - 1,
					fullPath:  n.fullPath,
				}

				// 更新子节点的最大参数数量
				for i := range child.children {
					if child.children[i].maxParams > child.maxParams {
						child.maxParams = child.children[i].maxParams
					}
				}

				n.children = []*node{&child}
				// []byte类型用于unicode字符转换
				n.indices = string([]byte{n.path[i]})
				n.path = path[:i]
				n.handlers = nil
				n.wildChild = false
				n.fullPath = fullPath[:parentFullPathIndex+i]
			}

			// 使新节点成为当前节点的子节点
			if i < len(path) {
				path = path[i:]

				if n.wildChild {
					parentFullPathIndex += len(n.path)
					n = n.children[0]
					n.priority++

					// 更新这个新节点的最大参数数量
					if numParams > n.maxParams {
						n.maxParams = numParams
					}
					numParams--

					// 检查通配符是否匹配
					if len(path) >= len(n.path) && n.path == path[:len(n.path)] {
						// 检查更长的通配符，例如 :name 和 :names
						if len(n.path) >= len(path) || path[len(n.path)] == '/' {
							continue walk
						}
					}

					pathSeg := path
					if n.nType != catchAll {
						pathSeg = strings.SplitN(path, "/", 2)[0]
					}
					prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
					panic(pathSeg + " 在新路径 " + fullPath + " 与现有的通配符冲突 " + n.path + " 在已存在的前缀 " + prefix)
				}

				c := path[0]

				// 处理参数后面的斜线
				if n.nType == param && c == '/' && len(n.children) == 1 {
					parentFullPathIndex += len(n.path)
					n = n.children[0]
					n.priority++
					continue walk
				}

				// Check if a child with the next path byte exists
				// 检查是否存在下一个子路径字节
				for i := 0; i < len(n.indices); i++ {
					if c == n.indices[i] {
						parentFullPathIndex += len(n.path)
						i = n.incrementChildPrio(i)
						n = n.children[i]
						continue walk
					}
				}

				// 否则插入
				if c != ':' && c != '*' {
					// []byte类型用于unicode字符转换
					n.indices += string([]byte{c})
					child := &node{
						maxParams: numParams,
						fullPath:  fullPath,
					}
					n.children = append(n.children, child)
					n.incrementChildPrio(len(n.indices) - 1)
					n = child
				}
				n.insertChild(numParams, path, fullPath, handlers)
				return

			} else if i == len(path) { // Make node a (in-path) leaf
				if n.handlers != nil {
					panic("handlers are already registered for path '" + fullPath + "'")
				}
				n.handlers = handlers
			}
			return
		}
	} else { // Empty tree
		n.insertChild(numParams, path, fullPath, handlers)
		n.nType = root
	}
}

// getValue returns the handle registered with the given path (key). The values of
// wildcards are saved to a map.
// If no handle can be found, a TSR (trailing slash redirect) recommendation is
// made if a handle exists with an extra (without the) trailing slash for the
// given path.
func (n *node) getValue(path string, po Params, unescape bool) (value nodeValue) {
	value.params = po
walk: // Outer loop for walking the tree
	for {
		if len(path) > len(n.path) {
			if path[:len(n.path)] == n.path {
				path = path[len(n.path):]
				// If this node does not have a wildcard (param or catchAll)
				// child,  we can just look up the next child node and continue
				// to walk down the tree
				if !n.wildChild {
					c := path[0]
					for i := 0; i < len(n.indices); i++ {
						if c == n.indices[i] {
							n = n.children[i]
							continue walk
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					value.tsr = path == "/" && n.handlers != nil
					return
				}

				// handle wildcard child
				n = n.children[0]
				switch n.nType {
				case param:
					// find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// save param value
					if cap(value.params) < int(n.maxParams) {
						value.params = make(Params, 0, n.maxParams)
					}
					i := len(value.params)
					value.params = value.params[:i+1] // expand slice within preallocated capacity
					value.params[i].Key = n.path[1:]
					val := path[:end]
					if unescape {
						var err error
						if value.params[i].Value, err = url.QueryUnescape(val); err != nil {
							value.params[i].Value = val // fallback, in case of error
						}
					} else {
						value.params[i].Value = val
					}

					// we need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						value.tsr = len(path) == end+1
						return
					}

					if value.handlers = n.handlers; value.handlers != nil {
						value.fullPath = n.fullPath
						return
					}
					if len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
						value.tsr = n.path == "/" && n.handlers != nil
					}

					return

				case catchAll:
					// save param value
					if cap(value.params) < int(n.maxParams) {
						value.params = make(Params, 0, n.maxParams)
					}
					i := len(value.params)
					value.params = value.params[:i+1] // expand slice within preallocated capacity
					value.params[i].Key = n.path[2:]
					if unescape {
						var err error
						if value.params[i].Value, err = url.QueryUnescape(path); err != nil {
							value.params[i].Value = path // fallback, in case of error
						}
					} else {
						value.params[i].Value = path
					}

					value.handlers = n.handlers
					value.fullPath = n.fullPath
					return

				default:
					panic("invalid node type")
				}
			}
		} else if path == n.path {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if value.handlers = n.handlers; value.handlers != nil {
				value.fullPath = n.fullPath
				return
			}

			if path == "/" && n.wildChild && n.nType != root {
				value.tsr = true
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i := 0; i < len(n.indices); i++ {
				if n.indices[i] == '/' {
					n = n.children[i]
					value.tsr = (len(n.path) == 1 && n.handlers != nil) ||
						(n.nType == catchAll && n.children[0].handlers != nil)
					return
				}
			}

			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		value.tsr = (path == "/") ||
			(len(n.path) == len(path)+1 && n.path[len(path)] == '/' &&
				path == n.path[:len(n.path)-1] && n.handlers != nil)
		return
	}
}

// findCaseInsensitivePath makes a case-insensitive lookup of the given path and tries to find a handler.
// It can optionally also fix trailing slashes.
// It returns the case-corrected path and a bool indicating whether the lookup
// was successful.
func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (ciPath []byte, found bool) {
	ciPath = make([]byte, 0, len(path)+1) // preallocate enough memory

	// Outer loop for walking the tree
	for len(path) >= len(n.path) && strings.EqualFold(path[:len(n.path)], n.path) {
		path = path[len(n.path):]
		ciPath = append(ciPath, n.path...)

		if len(path) > 0 {
			// If this node does not have a wildcard (param or catchAll) child,
			// we can just look up the next child node and continue to walk down
			// the tree
			if !n.wildChild {
				r := unicode.ToLower(rune(path[0]))
				for i, index := range n.indices {
					// must use recursive approach since both nextHandlerIndex and
					// ToLower(nextHandlerIndex) could exist. We must check both.
					if r == unicode.ToLower(index) {
						out, found := n.children[i].findCaseInsensitivePath(path, fixTrailingSlash)
						if found {
							return append(ciPath, out...), true
						}
					}
				}

				// Nothing found. We can recommend to redirect to the same URL
				// without a trailing slash if a leaf exists for that path
				found = fixTrailingSlash && path == "/" && n.handlers != nil
				return
			}

			n = n.children[0]
			switch n.nType {
			case param:
				// find param end (either '/' or path end)
				k := 0
				for k < len(path) && path[k] != '/' {
					k++
				}

				// add param value to case insensitive path
				ciPath = append(ciPath, path[:k]...)

				// we need to go deeper!
				if k < len(path) {
					if len(n.children) > 0 {
						path = path[k:]
						n = n.children[0]
						continue
					}

					// ... but we can't
					if fixTrailingSlash && len(path) == k+1 {
						return ciPath, true
					}
					return
				}

				if n.handlers != nil {
					return ciPath, true
				} else if fixTrailingSlash && len(n.children) == 1 {
					// No handle found. Check if a handle for this path + a
					// trailing slash exists
					n = n.children[0]
					if n.path == "/" && n.handlers != nil {
						return append(ciPath, '/'), true
					}
				}
				return

			case catchAll:
				return append(ciPath, path...), true

			default:
				panic("invalid node type")
			}
		} else {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if n.handlers != nil {
				return ciPath, true
			}

			// No handle found.
			// Try to fix the path by adding a trailing slash
			if fixTrailingSlash {
				for i := 0; i < len(n.indices); i++ {
					if n.indices[i] == '/' {
						n = n.children[i]
						if (len(n.path) == 1 && n.handlers != nil) ||
							(n.nType == catchAll && n.children[0].handlers != nil) {
							return append(ciPath, '/'), true
						}
						return
					}
				}
			}
			return
		}
	}

	// Nothing found.
	// Try to fix the path by adding / removing a trailing slash
	if fixTrailingSlash {
		if path == "/" {
			return ciPath, true
		}
		if len(path)+1 == len(n.path) && n.path[len(path)] == '/' &&
			strings.EqualFold(path, n.path[:len(path)]) &&
			n.handlers != nil {
			return append(ciPath, n.path...), true
		}
	}
	return
}
