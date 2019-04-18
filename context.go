package tsing

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"

	"github.com/dxvgef/filter"
	"github.com/julienschmidt/httprouter"
)

// Context 上下文
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	routerParams   httprouter.Params // 路由参数
	Next           bool              // 继续执行下一个中间件或处理器
	app            *App
	parsed         bool // 是否已解析body
}

// ReqValue 请求参数值
type ReqValue struct {
	Key   string // 参数名
	Value string // 参数值
	Error error  // 错误
}

// Continue 继续执行下一个中间件或处理器
func (ctx Context) Continue() (Context, error) {
	ctx.Next = true
	return ctx, nil
}

// Break 中断，不继续执行下一个中间件或处理器，如果err不为nil，则同时抛出500事件
func (ctx Context) Break(err error) (Context, error) {
	ctx.Next = false
	return ctx, err
}

// Event 触发500事件，使用此方法是为了精准记录触发事件的源码文件及行号
func (ctx Context) Event(err error) error {
	if err != nil && ctx.app.Event.Handler != nil {
		event := Event{
			Status:         500,
			Message:        err,
			ResponseWriter: ctx.ResponseWriter,
			Request:        ctx.Request,
		}
		if ctx.app.Event.EnableTrace == true {
			_, file, line, _ := runtime.Caller(1)
			l := strconv.Itoa(line)
			if ctx.app.Event.ShortCaller == true {
				short := file
				fileLen := len(file)
				for i := fileLen - 1; i > 0; i-- {
					if file[i] == '/' {
						short = file[i+1:]
						break
					}
				}
				file = short
			}
			event.Trace = append(event.Trace, file+":"+l)
		}
		ctx.app.Event.Handler(event)
	}
	// 不再将传入的error返回，避免再触发handle500函数
	return nil
}

// SetContextValue 在ctx里存储值，如果key存在则替换值
func (ctx *Context) SetContextValue(key string, value interface{}) {
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), key, value))
}

// ContextValue 获取ctx里的值，取出后根据写入的类型自行断言
func (ctx Context) ContextValue(key string) interface{} {
	return ctx.Request.Context().Value(key)
}

// Redirect 重定向
func (ctx Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return errors.New("状态码只能是300-308之间的值")
	}
	ctx.ResponseWriter.Header().Set("Location", url)
	ctx.ResponseWriter.WriteHeader(code)
	return nil
}

// RealIP 获得客户端真实IP
func (ctx Context) RealIP() string {
	ra := ctx.Request.RemoteAddr
	if ip := ctx.Request.Header.Get("X-Forwarded-For"); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := ctx.Request.Header.Get("X-Real-IP"); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// 解析body数据
func (ctx Context) parseBody() error {
	// 判断是否已经解析过body
	if ctx.parsed == true {
		return nil
	}
	if strings.HasPrefix(ctx.Request.Header.Get("Content-Type"), "multipart/form-data") {
		if err := ctx.Request.ParseMultipartForm(http.DefaultMaxHeaderBytes); err != nil {
			return err
		}
	} else {
		if err := ctx.Request.ParseForm(); err != nil {
			return err
		}
	}
	// 标记该context中的body已经解析过
	ctx.parsed = true
	return nil
}

// RouteValue 获取路由参数值
func (ctx Context) RouteValue(key string) ReqValue {
	return ReqValue{
		Key:   key,
		Value: ctx.routerParams.ByName(key),
	}
}

// RouteValues 获取所有路由参数值
func (ctx Context) RouteValues() []httprouter.Param {
	return ctx.routerParams
}

// RawRouteValue 获取某个路由参数值的string类型
func (ctx Context) RawRouteValue(key string) string {
	return ctx.routerParams.ByName(key)
}

// QueryValues 获取所有GET参数值
func (ctx Context) QueryValues() url.Values {
	return ctx.Request.URL.Query()
}

// QueryValue 获取某个GET参数值
func (ctx Context) QueryValue(key string) ReqValue {
	if len(ctx.Request.URL.Query()[key]) == 0 {
		return ReqValue{
			Key:   key,
			Error: errors.New("GET参数" + key + "不存在"),
		}
	}
	return ReqValue{
		Key:   key,
		Value: ctx.Request.URL.Query()[key][0],
	}
}

// RawQueryValue 获取某个GET参数值的string类型
func (ctx Context) RawQueryValue(key string) string {
	if len(ctx.Request.URL.Query()[key]) == 0 {
		return ""
	}
	return ctx.Request.URL.Query()[key][0]
}

// FormValue 获取某个POST参数值
func (ctx Context) FormValue(key string) ReqValue {
	err := ctx.parseBody()
	if err != nil {
		return ReqValue{
			Key:   key,
			Error: err,
		}
	}
	vs := ctx.Request.Form[key]
	if len(vs) == 0 {
		return ReqValue{
			Key:   key,
			Error: errors.New(ctx.Request.Method + "参数" + key + "不存在"),
		}
	}
	return ReqValue{
		Key:   key,
		Value: ctx.Request.Form[key][0],
	}
}

// PostValues 获取所有POST参数值
func (ctx Context) PostValues() url.Values {
	err := ctx.parseBody()
	if err != nil {
		return url.Values{}
	}
	return ctx.Request.PostForm
}

// FormValues 获取所有GET/POST参数值
func (ctx Context) FormValues() url.Values {
	err := ctx.parseBody()
	if err != nil {
		return url.Values{}
	}
	return ctx.Request.Form
}

// RawFormValue 获取某个POST参数值的string类型
func (ctx Context) RawFormValue(key string) string {
	err := ctx.parseBody()
	if err != nil {
		return ""
	}
	vs := ctx.Request.Form[key]
	if len(vs) == 0 {
		return ""
	}
	return ctx.Request.Form[key][0]
}

// String将参数值转为string
func (bv ReqValue) String(rules ...filter.Rule) (string, error) {
	if bv.Error != nil {
		return "", bv.Error
	}
	bv.Value, bv.Error = filter.String(bv.Value, rules...)
	return bv.Value, bv.Error
}

// MustString将参数值转为string，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustString(def string, rules ...filter.Rule) string {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	return bv.Value
}

// Int 将参数值转为int类型
func (bv ReqValue) Int(rules ...filter.Rule) (int, error) {
	if bv.Error != nil {
		return 0, bv.Error
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return 0, bv.Error
	}
	value, err := strconv.Atoi(bv.Value)
	if err != nil {
		bv.Error = err
		return 0, err
	}
	return value, err
}

// MustInt 将参数值转为int类型，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustInt(def int, rules ...filter.Rule) int {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	value, err := strconv.Atoi(bv.Value)
	if err != nil {
		bv.Error = err
		return def
	}
	return value
}

// Int32 将参数值转为int32类型
func (bv ReqValue) Int32(rules ...filter.Rule) (int32, error) {
	if bv.Error != nil {
		return 0, bv.Error
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return 0, bv.Error
	}
	value, err := strconv.ParseInt(bv.Value, 10, 32)
	if err != nil {
		bv.Error = err
		return 0, err
	}
	return int32(value), err
}

// Int32 将参数值转为int32类型，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustInt32(def int32, rules ...filter.Rule) int32 {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	value, err := strconv.ParseInt(bv.Value, 10, 32)
	if err != nil {
		bv.Error = err
		return def
	}
	return int32(value)
}

// Int64 将参数值转为int64类型
func (bv ReqValue) Int64(rules ...filter.Rule) (int64, error) {
	if bv.Error != nil {
		return 0, bv.Error
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return 0, bv.Error
	}
	value, err := strconv.ParseInt(bv.Value, 10, 64)
	if err != nil {
		bv.Error = err
		return 0, err
	}
	return value, err
}

// MustInt64 将参数值转为int64类型，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustInt64(def int64, rules ...filter.Rule) int64 {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	value, err := strconv.ParseInt(bv.Value, 10, 64)
	if err != nil {
		bv.Error = err
		return def
	}
	return value
}

// Uint32 将参数值转为uint32类型
func (bv ReqValue) Uint32(rules ...filter.Rule) (uint32, error) {
	if bv.Error != nil {
		return 0, bv.Error
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return 0, bv.Error
	}
	value, err := strconv.ParseUint(bv.Value, 10, 32)
	if err != nil {
		bv.Error = err
		return 0, err
	}
	return uint32(value), err
}

// MustUint32 将参数值转为uint32类型，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustUint32(def uint32, rules ...filter.Rule) uint32 {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	value, err := strconv.ParseUint(bv.Value, 10, 32)
	if err != nil {
		bv.Error = err
		return def
	}
	return uint32(value)
}

// Uint64 将参数值转为uint64类型
func (bv ReqValue) Uint64(rules ...filter.Rule) (uint64, error) {
	if bv.Error != nil {
		return 0, bv.Error
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return 0, bv.Error
	}
	value, err := strconv.ParseUint(bv.Value, 10, 64)
	if err != nil {
		bv.Error = err
		return 0, err
	}
	return value, err
}

// MustUint64 将参数值转为uint64类型，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustUint64(def uint64, rules ...filter.Rule) uint64 {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	value, err := strconv.ParseUint(bv.Value, 10, 64)
	if err != nil {
		return def
	}
	return value
}

// Float32 将参数值转为float32类型
func (bv ReqValue) Float32(rules ...filter.Rule) (float32, error) {
	if bv.Error != nil {
		return 0, bv.Error
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return 0, bv.Error
	}
	value, err := strconv.ParseFloat(bv.Value, 32)
	if err != nil {
		bv.Error = err
		return 0, err
	}
	return float32(value), err
}

// MustFloat32 将参数值转为float32类型，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustFloat32(def float32, rules ...filter.Rule) float32 {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	value, err := strconv.ParseFloat(bv.Value, 32)
	if err != nil {
		return def
	}
	return float32(value)
}

// Float64 将参数值转为float64类型
func (bv ReqValue) Float64(rules ...filter.Rule) (float64, error) {
	if bv.Error != nil {
		return 0, bv.Error
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return 0, bv.Error
	}
	value, err := strconv.ParseFloat(bv.Value, 64)
	if err != nil {
		bv.Error = err
		return 0, err
	}
	return value, err
}

// MustFloat64 将参数值转为float64类型，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustFloat64(def float64, rules ...filter.Rule) float64 {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	value, err := strconv.ParseFloat(bv.Value, 64)
	if err != nil {
		bv.Error = err
		return def
	}
	return value
}

// Bool 将参数值转为bool类型
func (bv ReqValue) Bool(rules ...filter.Rule) (bool, error) {
	if bv.Error != nil {
		return false, bv.Error
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return false, bv.Error
	}
	value, err := strconv.ParseBool(bv.Value)
	if err != nil {
		bv.Error = err
		return false, err
	}
	return value, err
}

// MustBool 将参数值转为bool类型，如果出错或者校验失败则返回默认值
func (bv ReqValue) MustBool(def bool, rules ...filter.Rule) bool {
	if bv.Error != nil {
		return def
	}
	if bv.Value, bv.Error = filter.String(bv.Value, rules...); bv.Error != nil {
		return def
	}
	value, err := strconv.ParseBool(bv.Value)
	if err != nil {
		return def
	}
	return value
}
