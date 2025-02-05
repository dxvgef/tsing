package tsing

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Context is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Status         int   // 处理器执行结果的状态码(HTTP)
	Error          error // 处理器执行错误时的消息

	broke        bool
	index        int8
	fullPath     string
	engine       *Engine
	params       *Params
	skippedNodes *[]skippedNode
	queryCache   url.Values
	formCache    url.Values
}

func (ctx *Context) reset() {
	ctx.Status = 200
	ctx.Error = nil
	ctx.index = -1
	ctx.broke = false
	ctx.fullPath = ""
	ctx.queryCache = nil
	ctx.formCache = nil
	*ctx.params = (*ctx.params)[:0]
	*ctx.skippedNodes = (*ctx.skippedNodes)[:0]
}

// EngineConfig 获取引擎配置
func (ctx *Context) EngineConfig() Config {
	return ctx.engine.config
}

// FullPath 返回路由注册时的路径
func (ctx *Context) FullPath() string {
	return ctx.fullPath
}

// Abort 停止执行该路由注册的其它处理器
func (ctx *Context) Abort() *Context {
	ctx.broke = true
	return ctx
}

// IsAborted 判断是否已停止执行其它处理器
func (ctx *Context) IsAborted() bool {
	return ctx.broke
}

// SetValue 在Context中写入键值，可用于在本次会话的处理器链中传递
func (ctx *Context) SetValue(key, value any) {
	if key == nil {
		return
	}
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), key, value))
}

// GetValue 从Context中读取键值
func (ctx *Context) GetValue(key any) any {
	if key == nil {
		return nil
	}
	return ctx.Request.Context().Value(key)
}

// PathValue 获取路径参数值
func (ctx *Context) PathValue(key string) string {
	return ctx.params.ByName(key)
}

// PathParam 获取路径参数，并判断参数是否存在
func (ctx *Context) PathParam(key string) (string, bool) {
	return ctx.params.Get(key)
}

// AllPathValues 获取所有路径参数值
func (ctx *Context) AllPathValue() []Param {
	return *ctx.params
}

// 初始化查询参数缓存
func (ctx *Context) initQueryCache() {
	if ctx.queryCache == nil {
		ctx.queryCache = url.Values{}
		if ctx.Request != nil {
			ctx.queryCache = ctx.Request.URL.Query()
		}
	}
}

// InitFormCache 初始化表单参数缓存
func (ctx *Context) InitFormCache() error {
	if ctx.formCache == nil {
		ctx.formCache = make(url.Values)

		if err := ctx.Request.ParseMultipartForm(ctx.engine.config.MaxMultipartMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
			return err
		}

		ctx.formCache = ctx.Request.PostForm
	}
	return nil
}

// QueryValue 获取某个GET参数的第一个值
func (ctx *Context) QueryValue(key string) string {
	ctx.initQueryCache()
	return ctx.queryCache.Get(key)
}

// QueryParam 获取某个GET参数的第一个值，并判断参数是否存在
func (ctx *Context) QueryParam(key string) (string, bool) {
	ctx.initQueryCache()
	values, ok := ctx.queryCache[key]
	if !ok {
		return "", false
	}
	return values[0], true
}

// QueryValues 获取某个GET参数的所有值
func (ctx *Context) QueryValues(key string) []string {
	ctx.initQueryCache()
	return ctx.queryCache[key]
}

// QueryParams 获取某个GET参数的所有值，并判断参数是否存在
func (ctx *Context) QueryParams(key string) ([]string, bool) {
	ctx.initQueryCache()
	values, ok := ctx.queryCache[key]
	return values, ok
}

// AllQueryValue 获取所有GET参数
func (ctx *Context) AllQueryValue() url.Values {
	ctx.initQueryCache()
	return ctx.queryCache
}

// FormValue 获取某个Form参数值的string类型
func (ctx *Context) FormValue(key string) string {
	if err := ctx.InitFormCache(); err != nil {
		return ""
	}
	return ctx.formCache.Get(key)
}

// FormParam 获取某个Form参数的第一个值，并判断参数是否存在
func (ctx *Context) FormParam(key string) (string, bool) {
	if err := ctx.InitFormCache(); err != nil {
		return "", false
	}
	values, exist := ctx.formCache[key]
	if !exist {
		return "", false
	}
	return values[0], true
}

// FormValues 获取某个Form参数的所有值
func (ctx *Context) FormValues(key string) []string {
	if err := ctx.InitFormCache(); err != nil {
		return nil
	}
	return ctx.formCache[key]
}

// FormParams 获取某个Form参数的所有值，并判断参数是否存在
func (ctx *Context) FormParams(key string) ([]string, bool) {
	if err := ctx.InitFormCache(); err != nil {
		return nil, false
	}
	values, exist := ctx.formCache[key]
	if !exist {
		return nil, false
	}
	return values, true
}

// AllFormValue 获取所有Form参数
func (ctx *Context) AllFormValue() url.Values {
	if err := ctx.InitFormCache(); err != nil {
		return nil
	}
	return ctx.formCache
}

// FormFile 根据参数名获取上传的第一个文件
func (ctx *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	if ctx.Request.MultipartForm == nil {
		if err := ctx.Request.ParseMultipartForm(ctx.engine.config.MaxMultipartMemory); err != nil {
			return nil, nil, err
		}
	}
	return ctx.Request.FormFile(name)
}

// FormFiles 根据参数名获取上传的所有文件
func (ctx *Context) FormFiles(name string) ([]*multipart.FileHeader, error) {
	if ctx.Request.MultipartForm == nil {
		if err := ctx.Request.ParseMultipartForm(ctx.engine.config.MaxMultipartMemory); err != nil {
			return nil, err
		}
	}
	fileHeaders, exist := ctx.Request.MultipartForm.File[name]
	if !exist {
		return nil, nil
	}
	return fileHeaders, nil
}

// SaveFile 保存上传的文件到本地路径
func (ctx *Context) SaveFile(fileHeader *multipart.FileHeader, savePath string, perm os.FileMode) error {
	f, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	if err = os.MkdirAll(filepath.Dir(savePath), perm); err != nil {
		return err
	}

	var out *os.File
	out, err = os.Create(savePath) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		err = out.Close()
	}()

	_, err = io.Copy(out, f)
	return err
}

// ServeFile 将服务端指定路径的文件写入到客户端流
func (ctx *Context) ServeFile(filePath string) {
	http.ServeFile(ctx.ResponseWriter, ctx.Request, filePath)
}

// FileFromFS 发布本地目录为静态文件目录
func (ctx *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		ctx.Request.URL.Path = old
	}(ctx.Request.URL.Path)

	ctx.Request.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(ctx.ResponseWriter, ctx.Request)
}

// Redirect 向客户端发送重定向响应
func (ctx *Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return errors.New("状态码只能是30x")
	}
	ctx.Status = code
	ctx.ResponseWriter.Header().Set("Location", url)
	ctx.ResponseWriter.WriteHeader(code)
	return nil
}

// String 输出字符串
func (ctx *Context) String(status int, data string, charset ...string) (err error) {
	if len(charset) == 0 {
		ctx.ResponseWriter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		ctx.ResponseWriter.Header().Set("Content-Type", "text/plain; charset="+charset[0])
	}
	ctx.ResponseWriter.WriteHeader(status)
	_, err = ctx.ResponseWriter.Write(strToBytes(data))
	return
}

// JSON 输出JSON
func (ctx *Context) JSON(status int, data any, charset ...string) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	contentType := "application/json; charset=utf-8"
	if len(charset) > 0 {
		contentType = "application/json; charset=" + charset[0]
	}
	ctx.ResponseWriter.Header().Set("Content-Type", contentType)
	ctx.ResponseWriter.WriteHeader(status)
	_, err = ctx.ResponseWriter.Write(buf)
	return err //nolint:wrapcheck
}

// NoContent 输出204状态码
func (ctx *Context) NoContent() error {
	ctx.Status = http.StatusNoContent
	ctx.ResponseWriter.WriteHeader(http.StatusNoContent)
	return nil
}

// StatusCode 输出状态码
func (ctx *Context) StatusCode(code int) (err error) {
	ctx.Status = code
	ctx.ResponseWriter.WriteHeader(code)
	return
}

// ParseJSON 将json格式的body数据反序列化到传入的对象
func (ctx *Context) ParseJSON(obj any) error {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, obj)
}
