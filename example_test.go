package tsing

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// 测试回应
func TestEcho(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	app.GET("/", func(ctx *Context) error {
		ctx.ResponseWriter.WriteHeader(200)
		_, _ = ctx.ResponseWriter.Write([]byte("Hello !"))
		t.Log("Hello !")
		return nil
	})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试 PathParams
func TestURLParams(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	app.GET("/:test/ok", func(ctx *Context) error {
		t.Log(ctx.PathParams.Value("test"))
		return nil
	})
	r, err := http.NewRequest("GET", "/haha/ok", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试Context传值
func TestContext(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	app.GET("/context", func(ctx *Context) error {
		// 在ctx中写入参数
		ctx.SetValue("test", "hehe")
		t.Log(1, ctx.Request.URL.Path, "写值")
		return nil
	}, func(ctx *Context) error {
		// 从ctx中读取参数
		t.Log(2, ctx.Request.URL.Path, "取值：", ctx.GetValue("test"))
		return nil
	})
	r, err := http.NewRequest("GET", "/context", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试路由组
func TestGroup(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	group := app.Group("/group", func(ctx *Context) error {
		ctx.SetValue("test", "haha")
		t.Log(1, ctx.Request.URL.Path, "写值")
		return nil
	})
	group.GET("/object", func(ctx *Context) error {
		t.Log(2, ctx.Request.URL.Path, "取值：", ctx.GetValue("test"))
		return nil
	}, func(ctx *Context) error {
		t.Log(3, ctx.Request.URL.Path, "取值：", ctx.GetValue("test"))
		return nil
	})
	r, err := http.NewRequest("GET", "/group/object", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试中止
func TestAbort(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	group := app.Group("/group")
	group.GET("/object", func(ctx *Context) error {
		t.Log(1, ctx.Request.URL.Path)
		ctx.Abort()
		t.Log(2, ctx.IsAborted())
		return nil
	}, func(ctx *Context) error {
		t.Log(3, ctx.Request.URL.Path)
		return nil
	})
	r, err := http.NewRequest("GET", "/group/object", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试Append
func TestAppend(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	app.Append(func(ctx *Context) error {
		t.Log(1, "append handler 1")
		return nil
	}, func(ctx *Context) error {
		t.Log(2, "append handler 2")
		return nil
	})
	app.GET("/test", func(ctx *Context) error {
		t.Log(3, ctx.Request.URL.Path)
		return nil
	})
	r, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试QueryValues
func TestQueryValues(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	app.GET("/object", func(ctx *Context) error {
		t.Log(ctx.QueryValues())
		return nil
	})
	r, err := http.NewRequest("GET", "/object?a=1&b=2", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试PostValues
func TestPostValues(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	app.POST("/object", func(ctx *Context) error {
		t.Log(ctx.PostValues())
		return nil
	})

	v := url.Values{}
	v.Add("a", "1")
	v.Add("b", "2")
	r, err := http.NewRequest("POST", "/object", strings.NewReader(v.Encode()))
	if err != nil {
		t.Error(err.Error())
		return
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试FormValues
func TestFormValues(t *testing.T) {
	app := New(&Config{
		UnescapePathValues: true,
		MaxMultipartMemory: 20 << 20,
	})
	app.POST("/object", func(ctx *Context) error {
		t.Log(ctx.FormValues())
		return nil
	})

	v := url.Values{}
	v.Add("c", "3")
	v.Add("d", "4")
	r, err := http.NewRequest("POST", "/object?a=1&b=2", strings.NewReader(v.Encode()))
	if err != nil {
		t.Error(err.Error())
		return
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 事件处理器
func eventHandler(e *Event) {
	log.SetFlags(log.Lshortfile)
	log.Println(e.Status)
	log.Println(e.Message)
	log.Println(e.Source)
	for k := range e.Trace {
		log.Println("  ", e.Trace[k])
	}
}

// 测试404事件
func TestNotFoundEvent(t *testing.T) {
	app := New(&Config{
		RootPath:           getRootPath(),
		UnescapePathValues: true,
		EventHandler:       eventHandler,
	})
	r, err := http.NewRequest("GET", "/404", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试405事件
func TestMethodNotAllowedEvent(t *testing.T) {
	app := New(&Config{
		RootPath:           getRootPath(),
		UnescapePathValues: true,
		EventHandler:       eventHandler,
	})
	app.POST("/", func(ctx *Context) error {
		return nil
	})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试处理器返回的error事件
func TestHandlerErrorEvent(t *testing.T) {
	app := New(&Config{
		RootPath:           getRootPath(),
		UnescapePathValues: true,
		MaxMultipartMemory: 2 << 20,
		EventHandler:       eventHandler,
		EventTrace:         false,
		EventHandlerError:  true,
		EventSource:        true,
		EventShortPath:     true,
	})
	app.GET("/", func(ctx *Context) error {
		return errors.New("这是处理器返回的错误")
	}, func(ctx *Context) error {
		t.Error("处理器链的执行逻辑有异常")
		return nil
	})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试处理器中使用Source()包裹error并返回的事件
func TestContextSourceEvent(t *testing.T) {
	app := New(&Config{
		RootPath:           getRootPath(),
		UnescapePathValues: true,
		MaxMultipartMemory: 2 << 20,
		EventHandler:       eventHandler,
		EventTrace:         false,
		EventHandlerError:  true,
		EventSource:        true,
		EventShortPath:     true,
	})
	app.GET("/", func(ctx *Context) error {
		return ctx.Source(errors.New("这是用ctx.Source()返回的错误，能精准定位到事件来源"))
	}, func(ctx *Context) error {
		t.Error("处理器链的执行逻辑有异常")
		return nil
	})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试panic事件
func TestPanicEvent(t *testing.T) {
	app := New(&Config{
		RootPath:           getRootPath(),
		UnescapePathValues: true,
		MaxMultipartMemory: 2 << 20,
		EventHandler:       eventHandler,
		EventTrace:         true,
		EventSource:        true,
		Recover:            true,
		EventShortPath:     true,
	})
	app.GET("/", func(ctx *Context) error {
		panic("这是panic消息")
	})
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}
