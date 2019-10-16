package tsing

import "math"

const abortIndex int8 = math.MaxInt8 / 2

// 处理器
type HandlerFunc func(*Context) error

// 处理器数组
type HandlersChain []HandlerFunc

// 设置404处理器
// func (app *App) NotFoundHandlers(handlers ...HandlerFunc) {
// app.handers.notFound = handlers
// app.handers.allNotFound = app.Router.mergeHandlers(app.handers.notFound)
// }

// 重新构建404处理器集合
// func (app *App) rebuild404Handlers() {
// 	app.allNotFoundHandlers = app.mergeHandlers(app.notFoundHandlers)
// }

// 设置405处理器
// func (app *App) MethodNotAllowedHandlers(handlers ...HandlerFunc) {
// app.handers.methodNotAllowed = handlers
// app.handers.allNotMethod = app.Router.mergeHandlers(app.handers.methodNotAllowed)
// }

// func (app *App) rebuild405Handlers() {
// 	app.allNotMethodHandlers = app.mergeHandlers(app.methodNotAllowedHandlers)
// }

// 返回链中的最后一个处理程序
func (handlers HandlersChain) last() HandlerFunc {
	if length := len(handlers); length > 0 {
		return handlers[length-1]
	}
	return nil
}
