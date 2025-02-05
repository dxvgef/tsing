package tsing

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 错误回调处理器
func errorHandler(ctx *Context) {
	// 记录错误日志
	log.Println("记录错误日志", ctx.Status, ctx.Error)
	// 输出错误信息到客户端
	ctx.ResponseWriter.WriteHeader(ctx.Status)
}

// 后置回调处理器
func afterHandler(ctx *Context) {
	log.Println("执行了后置处理器", ctx.IsAborted())
}

// 测试回应
func TestEcho(t *testing.T) {
	app := New()
	app.GET("/", func(ctx *Context) error {
		t.Log(ctx.Request.RequestURI)
		t.Log("Hello Tsing")
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

func TestStatusCode(t *testing.T) {
	app := New()
	app.GET("/", func(ctx *Context) error {
		return ctx.NoContent()
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)
	t.Log(resp.Code)
}

// 测试处理器
func TestHandlers(t *testing.T) {
	app := New(Config{
		Recovery:     true,
		AfterHandler: afterHandler,
	})
	app.Use(func(ctx *Context) error {
		t.Log(ctx.Request.RequestURI, "1 执行了全局中间件")
		return nil
	})
	group := app.Group("/group", func(ctx *Context) error {
		t.Log(ctx.Request.RequestURI, "2 执行了 /group")
		return nil
	})
	group.Use(func(ctx *Context) error {
		t.Log(ctx.Request.RequestURI, "3 执行了路由组 /group 中间件")
		return nil
	})
	group.GET("/object", func(ctx *Context) error {
		t.Log(ctx.Request.RequestURI, "4 执行了 /group/object")
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/group2/object", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试 PathValue
func TestPathValue(t *testing.T) {
	app := New()
	app.GET("/:path/:file", func(ctx *Context) error {
		t.Log("path=", ctx.PathValue("path"))
		t.Log("file=", ctx.PathValue("file"))
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/haha/123", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试Context传值
func TestContextValue(t *testing.T) {
	app := New()
	app.GET("/", func(ctx *Context) error {
		// 在ctx中写入参数
		ctx.SetValue("hello", "tsing")
		return nil
	}, func(ctx *Context) error {
		t.Log("hello=", ctx.GetValue("hello"))
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试中止处理器链
func TestAbort(t *testing.T) {
	app := New()
	group := app.Group("/group")
	group.GET("/object", func(ctx *Context) error {
		t.Log("ok")
		ctx.Abort()
		return nil
	}, func(ctx *Context) error {
		t.Error(ctx.Request.RequestURI, "中止失败")
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/group/object", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试QueryValue
func TestQueryParams(t *testing.T) {
	app := New()
	app.GET("/", func(ctx *Context) error {
		t.Log("id=", ctx.QueryValue("id"))
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/?id=123", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试FormValue
func TestFormValue(t *testing.T) {
	app := New()
	app.POST("/", func(ctx *Context) error {
		t.Log("test=", ctx.FormValue("test"))
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader("test=ok"))
	if err != nil {
		t.Error(err)
		return
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试 404 错误
func TestNotFoundError(t *testing.T) {
	app := New(Config{
		ErrorHandler: errorHandler,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/404", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试 405 事件
func TestMethodNotAllowedError(t *testing.T) {
	app := New(Config{
		HandleMethodNotAllowed: true,
		ErrorHandler:           errorHandler,
	})
	app.POST("/", func(ctx *Context) error {
		t.Log(ctx.Request.RequestURI)
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试panic事件
func TestPanicError(t *testing.T) {
	app := New(Config{
		Recovery:     true,
		ErrorHandler: errorHandler,
	})
	app.GET("/", func(ctx *Context) error {
		t.Log(ctx.Request.RequestURI)
		panic("panic消息")
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}

// 测试CORS
func TestCORS(t *testing.T) {
	app := New(Config{
		ErrorHandler:           errorHandler, // 通过错误处理器来实现自动响应OPTIONS请求
		HandleMethodNotAllowed: true,         // 错误处理器中需要判断 405 状态码
	})
	app.GET("/", func(ctx *Context) error {
		t.Log(ctx.Request.RequestURI)
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, http.MethodOptions, "/", nil)
	if err != nil {
		t.Error(err)
		return
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
}
